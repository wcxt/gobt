package main

import (
	"crypto/sha1"
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/edwces/gobt"
	"github.com/edwces/gobt/message"
)

const (
	MaxBlockLength       = 16000
	MaxPipelinedRequests = 5
	MaxHashFails         = 5
	MaxPeerTimeout       = 2*time.Minute + 10*time.Second
)

func main() {
	path := os.Args[1]

	// Open the metainfo file
	metainfo, err := gobt.Open(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	clientID, err := gobt.GenRandPeerID()
	if err != nil {
		fmt.Println(err)
		return
	}

	hash, err := metainfo.Info.Hash()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Receive the peers from tracker
	peers, err := gobt.GetAvailablePeers(metainfo.Announce, hash, clientID, metainfo.Info.Length)
	if err != nil {
		fmt.Println(err)
		return
	}

	hashes, err := metainfo.Info.Hashes()
	if err != nil {
		fmt.Println(err)
		return
	}

	pieceRequests := make([]bool, len(hashes))
	pieceDone := make([]bool, len(hashes))
    pieceCounter := len(hashes)

	for _, peer := range peers {
		go func(peer gobt.AnnouncePeer) {
			conn, err := gobt.DialTimeout(peer.Addr())
			if err != nil {
				fmt.Printf("connection error: %v\n", err)
				return
			}
			defer conn.Close()

			err = conn.Handshake(hash, clientID)
			if err != nil {
				fmt.Printf("handshake error: %v\n", err)
				return
			}

			hashFails := 0

			// Message loop
			//choked := true
			interesting := false
			//choking := true
			//interested := false
			blocksPerPiece := int(math.Ceil(float64(metainfo.Info.PieceLength) / float64(MaxBlockLength)))

			downloadable := []int{}
			blockBuffer := []byte{}

			requestQueue := []message.Request{}

			// TODO: Pipeline requests
			// Request vars
			currentPiece := 0
			currentBlock := 0

			timer := time.NewTimer(MaxPeerTimeout)
			defer timer.Stop()

			go func() {
				<-timer.C
				conn.Close()
			}()

			for {
				msg, err := conn.ReadMsg()
				if err != nil {
					fmt.Println(err)
					return
				}

				if msg.KeepAlive {
					if !timer.Stop() {
						return
					}
					timer.Reset(MaxPeerTimeout)
                    continue
				}

				switch msg.ID {
				case message.IDChoke:
					for _, req := range requestQueue {
						currentBlock = int(req.Offset) / MaxBlockLength
						if int(req.Index) != currentPiece {
							downloadable = append([]int{currentPiece}, downloadable...)
                            pieceRequests[currentPiece] = false
							currentPiece = int(req.Index)
						}
					}

                    requestQueue = []message.Request{}
				//case message.IDInterested:
				//case message.IDNotInterested:
				case message.IDPiece:
					block := msg.Payload.Block()
					//fmt.Println(block.Index, block.Offset)

					// Save to downloaded blocks in a piece
					blockBuffer = append(blockBuffer, block.Block...)

					if len(blockBuffer) == metainfo.Info.PieceLength {
						blocksHash := sha1.Sum(blockBuffer)
						if blocksHash == hashes[block.Index] {
							pieceDone[block.Index] = true
                            pieceCounter--

                            fmt.Printf("%s GOT: %d; PIECES LEFT: %d\n", peer.Addr(), block.Index, pieceCounter)

                            _, err = conn.WriteHave(int(block.Index))
                            if err != nil {
                                return
                            }
						} else {
							hashFails += 1
							pieceRequests[block.Index] = false
						}
						// else mark piece as not being downloaded
						blockBuffer = []byte{}
					}

					requestQueue = requestQueue[1:]

					if hashFails >= MaxHashFails {
						fmt.Println("Excedded Maximum hash fails: 5")
						return
					}

					for i := len(requestQueue); i < MaxPipelinedRequests; i++ {
						if interesting {
							offset := currentBlock * MaxBlockLength
							length := math.Min(float64(metainfo.Info.PieceLength-offset), float64(MaxBlockLength))

							_, err := conn.WriteRequest(currentPiece, offset, int(length))
							if err != nil {
								fmt.Println(err)
								return
							}
							currentBlock += 1

							// Put request in pipeline
							requestQueue = append(requestQueue, message.Request{Index: uint32(currentPiece), Offset: uint32(offset), Length: uint32(length)})

							if currentBlock == blocksPerPiece {
                                if len(downloadable) == 0 {
                                    break
                                }

								currentPiece = downloadable[0]
								downloadable = downloadable[1:]

								// TEMP: skip over pieces that are being downloaded by other peers
								for pieceRequests[currentPiece] && len(downloadable) > 0 {
									currentPiece = downloadable[0]
									downloadable = downloadable[1:]
								}

                                if len(downloadable) == 0 {
                                    break
                                }

								// mark piece as being downloaded by this peer
								pieceRequests[currentPiece] = true

								currentBlock = 0
							}
						}
					}

					if len(downloadable) == 0 && interesting && len(requestQueue) == 0 {
						_, err := conn.WriteNotInterested()
						if err != nil {
							fmt.Println(err)
							return
						}
						interesting = false
					}

				case message.IDUnchoke:
					for i := len(requestQueue); i < MaxPipelinedRequests; i++ {
						if interesting {
							offset := currentBlock * MaxBlockLength
							length := math.Min(float64(metainfo.Info.PieceLength-offset), float64(MaxBlockLength))

							_, err := conn.WriteRequest(currentPiece, offset, int(length))
							if err != nil {
								fmt.Println(err)
								return
							}
							currentBlock += 1

							// Put request in pipeline
							requestQueue = append(requestQueue, message.Request{Index: uint32(currentPiece), Offset: uint32(offset), Length: uint32(length)})

							if currentBlock == blocksPerPiece {
                                if len(downloadable) == 0 {
                                    break
                                } 

								currentPiece = downloadable[0]
								downloadable = downloadable[1:]
                        
                                // TEMP: skip over pieces that are being downloaded by other peers
								for pieceRequests[currentPiece] && len(downloadable) > 0 {
									currentPiece = downloadable[0]
									downloadable = downloadable[1:]
								}

                                if len(downloadable) == 0 {
                                    break
                                }


								// mark piece as being downloaded by this peer
								pieceRequests[currentPiece] = true

								currentBlock = 0
							}
						}
					}
				case message.IDHave:
					have := int(msg.Payload.Have())
                    old := len(downloadable)

					// Detect peer downloaded pieces
					for index, processing := range pieceDone {
						if have == index && !processing {
							downloadable = append(downloadable, index)
						}
					}
                    
                    // Select new piece if it's the only one
                    if len(downloadable) == 1 && old == 0 {
                        currentPiece = downloadable[0]
                        downloadable = []int{}

                        if !pieceRequests[currentPiece] {
                            // mark piece as being downloaded by this peer
                            pieceRequests[currentPiece] = true
                            
                            // send new state
                            conn.WriteInterested()
                            if err != nil {
                                fmt.Println(err)
                                return
                            }
                            interesting = true
                        }
                    }
			case message.IDBitfield:
					bitfield := msg.Payload.Bitfield()

					// Detect peer downloaded pieces
					for index, processing := range pieceDone {
						if bitfield.Get(index) && !processing {
							downloadable = append(downloadable, index)
						}
					}

                    fmt.Printf("%s HAS %d PIECES\n", peer.Addr(), len(downloadable))
                    


                    // Select first piece
                    for len(downloadable) != 0 {
                        index := rand.Intn(len(downloadable))
						currentPiece = downloadable[index]
                        downloadable = append(downloadable[:index], downloadable[index+1:]...)

                        if pieceRequests[currentPiece] {
                            continue
                        }

						// mark piece as being downloaded by this peer
						pieceRequests[currentPiece] = true

                        // send new state
                        _, err = conn.WriteInterested()
                        if err != nil {
                            fmt.Println(err)
							return
                        }
						interesting = true
                        break
                    }
				}
			}
		}(peer)
	}

	select {}
}
