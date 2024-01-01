package main

import (
	"crypto/sha1"
	"fmt"
	"os"
	"time"

	"github.com/edwces/gobt"
	"github.com/edwces/gobt/bitfield"
	"github.com/edwces/gobt/message"
	"github.com/edwces/gobt/picker"
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

	pp := picker.New(metainfo.Info.Length, metainfo.Info.PieceLength)
	pieceCounter := len(hashes)
	maxBlocks := metainfo.Info.PieceLength / 16000
	filepieces := make([][][]byte, len(hashes))

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
			bf := bitfield.New(len(hashes))
			reqBacklog := 0
			picked := []*picker.Block{}

			// TODO: Pipeline requests
			// Request vars
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
					for _, block := range picked {
						pp.Return(block)
					}
					picked = []*picker.Block{}
				case message.IDPiece:
					block := msg.Payload.Block()
                    
					// Remove cb from array of picked
					cb := picked[len(picked)-1]
					picked := picked[:len(picked)-1]
					reqBacklog--

					pp.Done(cb)

					// Add piece content to buffer
					if filepieces[cb.Piece.Index] == nil {
						filepieces[cb.Piece.Index] = make([][]byte, maxBlocks)
					}
					filepieces[cb.Piece.Index][cb.Index] = block.Block

					if cb.Piece.Done {
						buffer := []byte{}
						for _, b := range filepieces[cb.Piece.Index] {
							buffer = append(buffer, b...)
						}

						blocksHash := sha1.Sum(buffer)
						if blocksHash == hashes[block.Index] {
							pieceCounter--
							fmt.Printf("%s GOT: %d; PIECES LEFT: %d\n", peer.Addr(), block.Index, pieceCounter)

							_, err = conn.WriteHave(int(block.Index))
							if err != nil {
								return
							}
						} else {
							hashFails += 1
							fmt.Printf("%s GOT FAILED: %d; PIECES LEFT: %d\n", peer.Addr(), block.Index, pieceCounter)
							pp.Add(cb.Piece.Index)
						}
					}

					if hashFails >= MaxHashFails {
						for _, block := range picked {
							pp.Return(block)
						}
						fmt.Println("Excedded Maximum hash fails: 5")
						return
					}

					for i := reqBacklog; i < MaxPipelinedRequests && interesting; i++ {
						// Send request
						req := picked[0]
						_, err := conn.WriteRequest(req.Piece.Index, req.Offset, req.Length)
						if err != nil {
							fmt.Println(err)
							return
						}
						reqBacklog++

						// Choose new piece
						cb, err := pp.Pick(bf)
                        picked = append([]*picker.Block{cb}, picked...)
						if err != nil {
							_, err := conn.WriteNotInterested()
							if err != nil {
								fmt.Println(err)
								return
							}
							interesting = false
						}
					}

				case message.IDUnchoke:
					for i := reqBacklog; i < MaxPipelinedRequests && interesting; i++ {
						// Send request
						req := picked[0]
						_, err := conn.WriteRequest(req.Piece.Index, req.Offset, req.Length)
						if err != nil {
							fmt.Println(err)
							return
						}
						reqBacklog++

						// Choose new piece
						cb, err := pp.Pick(bf)
                        picked = append([]*picker.Block{cb}, picked...)
						if err != nil {
							_, err := conn.WriteNotInterested()
							if err != nil {
								fmt.Println(err)
								return
							}
							interesting = false
						}
					}
				case message.IDHave:
					have := int(msg.Payload.Have())
					err := bf.Set(have)

					if err != nil {
						fmt.Printf("Bitfield: %v\n", err)
						return
					}

				case message.IDBitfield:
					err := bf.Replace(msg.Payload.Bitfield())
					if err != nil {
						fmt.Printf("Bitfield: %v\n", err)
						return
					}

					cb, err := pp.Pick(bf)
					if err == nil {
						picked = append([]*picker.Block{cb}, picked...)
						_, err := conn.WriteInterested()
						if err != nil {
							fmt.Println(err)
							return
						}
						interesting = true
					}

					conn.WriteUnchoke()
				}
			}
		}(peer)
	}

	select {}
}
