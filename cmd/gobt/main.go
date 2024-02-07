package main

import (
	"crypto/sha1"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/edwces/gobt"
	"github.com/edwces/gobt/bitfield"
	"github.com/edwces/gobt/message"
)

const (
	MaxPipelinedRequests = 5
	// MaxHashFails         = 5
	MaxPeerTimeout = 2*time.Minute + 10*time.Second
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

	pp := gobt.NewPicker(metainfo.Info.Length, metainfo.Info.PieceLength)
	downloaded := make([][][]byte, len(hashes))
	clientBf := bitfield.New(len(hashes))

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

			// Message loop
			interesting := false

			bf := bitfield.New(len(hashes))
			reqQueue := [][]int{}
			// hashFails := 0

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

				if !timer.Stop() {
					return
				}
				timer.Reset(MaxPeerTimeout)

				if msg.KeepAlive {
					continue
				}

				switch msg.ID {
				case message.IDChoke:
					break
				case message.IDPiece:
					reqQueue = reqQueue[1:]

					// Store piece
					block := msg.Payload.Block()
					bCount := gobt.BlockCount(metainfo.Info.Length, metainfo.Info.PieceLength, int(block.Index))

					if downloaded[block.Index] == nil {
						downloaded[block.Index] = make([][]byte, bCount)
					}

					downloaded[block.Index][block.Offset/gobt.DefaultBlockSize] = block.Block

					// Check if piece is full
					fullPiece := true
					for _, val := range downloaded[block.Index] {
						if val == nil {
							fullPiece = false
						}
					}

					if fullPiece {
						buf := []byte{}
						for _, b := range downloaded[block.Index] {
							buf = append(buf, b...)
						}

						pHash := sha1.Sum(buf)
						if pHash == hashes[block.Index] {
							fmt.Printf("-------------------------------------------------- %s GOT: %d; \n", peer.Addr(), block.Index)
							// 		_, err = conn.WriteHave(int(block.Index))
							// 		if err != nil {
							// 			return
							// 		}
						} else {
							fmt.Printf("-------------------------------------------------- %s GOT FAILED: %d; \n", peer.Addr(), block.Index)
							// hashFails += 1
							// if hashFails >= MaxHashFails {
							// 	fmt.Println("Excedded Maximum hash fails: 5")
							// 	return
							// }
						}
					}

					for len(reqQueue) < MaxPipelinedRequests && interesting {
						// Send request
						cp, cb, err := pp.Pick(bf)

						if err != nil {
							_, err := conn.WriteNotInterested()
							if err != nil {
								fmt.Println(err)
								return
							}
							interesting = false
							break
						}

						length := int(math.Min(float64(gobt.DefaultBlockSize), float64(metainfo.Info.PieceLength)-float64(cb*gobt.DefaultBlockSize)))
						_, err = conn.WriteRequest(cp, cb*gobt.DefaultBlockSize, length)
						if err != nil {
							fmt.Println(err)
							return
						}

						reqQueue = append(reqQueue, []int{cp, cb})
					}

				case message.IDUnchoke:
					unresolved := reqQueue
					reqQueue = [][]int{}

					for len(reqQueue) < MaxPipelinedRequests && interesting {
						// Pick block to request
						var cp, cb int

						if len(unresolved) == 0 {
							cp, cb, err = pp.Pick(bf)

							if err != nil {
								_, err := conn.WriteNotInterested()
								if err != nil {
									fmt.Println(err)
									return
								}
								interesting = false
								break
							}
						} else {
							cp = unresolved[0][0]
							cb = unresolved[0][1]
							unresolved = unresolved[1:]
						}

						length := int(math.Min(float64(gobt.DefaultBlockSize), float64(metainfo.Info.Length)-float64(cb*gobt.DefaultBlockSize)))
						_, err = conn.WriteRequest(cp, cb*gobt.DefaultBlockSize, length)
						if err != nil {
							fmt.Println(err)
							return
						}

						reqQueue = append(reqQueue, []int{cp, cb})

					}
				case message.IDHave:
					have := int(msg.Payload.Have())
					err := bf.Set(have)

					if err != nil {
						fmt.Printf("Bitfield: %v\n", err)
						return
					}

					if has, _ := clientBf.Get(have); !interesting && has {
						_, err := conn.WriteInterested()
						if err != nil {
							fmt.Println(err)
							return
						}
						interesting = true
					}
				case message.IDBitfield:
					// Define peer bitfield
					err := bf.Replace(msg.Payload.Bitfield())
					if err != nil {
						fmt.Printf("Bitfield: %v\n", err)
						return
					}

					// Calculate interesting pieces that peer has
					diff, err := bf.Difference(clientBf)
					if err != nil {
						fmt.Printf("Bitfield: %v\n", err)
						return
					}

					// Send interest status if it's not empty
					if !diff.Empty() {
						_, err := conn.WriteInterested()
						if err != nil {
							fmt.Println(err)
							return
						}
						interesting = true
					}
					// conn.WriteUnchoke()
				}
			}
		}(peer)
	}

	select {}
}
