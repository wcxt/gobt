package main

import (
	"crypto/sha1"
	"fmt"
	"math"
	"net"
	"os"
	"time"

	"github.com/edwces/gobt"
	"github.com/edwces/gobt/message"
)

const (
	MaxBlockLength       = 16000
	MaxPipelinedRequests = 5
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

	missing := make([]int, len(hashes))
	downloaded := make(chan []byte, metainfo.Info.Length)

	// Connect to peers
	go func(peer gobt.AnnouncePeer, in []int, out chan []byte) {
		// Establish conn
		conn, err := net.DialTimeout("tcp", peer.Addr(), 3*time.Second)
		if err != nil {
			fmt.Printf("connection error: %v\n", err)
			return
		}
		defer conn.Close()

		// Establish handshake
		err = gobt.EstablishHandshake(conn, hash, clientID)
		if err != nil {
			fmt.Printf("handshake error: %v\n", err)
			return
		}

		// Message loop
		//choked := true
		interesting := false
		//choking := true
		//interested := false
		blocksPerPiece := int(math.Ceil(float64(metainfo.Info.PieceLength) / float64(MaxBlockLength)))

		downloadable := []int{}
		blockBuffer := []byte{}

		// TODO: Pipeline requests
		// Request vars
		currentPiece := 0
		currentBlock := 0

		for {
			msg, err := message.Read(conn)
			if err != nil {
				fmt.Println(err)
				return
			}

			switch msg.ID {
			//case message.IDChoke:
			//case message.IDInterested:
			//case message.IDNotInterested:
			case message.IDPiece:
				block := msg.Payload.Block()
				fmt.Println(block.Index, block.Offset)

				// Save to downloaded blocks in a piece
				blockBuffer = append(blockBuffer, block.Block...)
                
				if len(blockBuffer) == metainfo.Info.PieceLength {
					// TODO: Make it no longer downloadable
                    blocksHash := sha1.Sum(blockBuffer)
                    if blocksHash == hashes[block.Index] {
                        fmt.Println("GOT PIECE WITH CORRECT HASH")
                    }
					blockBuffer = []byte{}
				}

				if len(downloadable) > currentPiece && interesting {
					index := downloadable[currentPiece]
					offset := currentBlock * MaxBlockLength
					length := math.Min(float64(metainfo.Info.PieceLength-offset), float64(MaxBlockLength))

					req := message.Request{Index: uint32(index), Offset: uint32(offset), Length: uint32(length)}
					nmsg := &message.Message{ID: message.IDRequest, Payload: message.NewRequestPayload(req)}
					_, err := message.Write(conn, nmsg)
					if err != nil {
						fmt.Println(err)
						return
					}
					currentBlock += 1

					if currentBlock == blocksPerPiece {
						currentPiece += 1
						currentBlock = 0
					}
				}

			case message.IDUnchoke:
				if len(downloadable) > currentPiece && interesting {
					index := downloadable[currentPiece]
					offset := currentBlock * MaxBlockLength
					length := math.Min(float64(metainfo.Info.PieceLength-offset), float64(MaxBlockLength))

					req := message.Request{Index: uint32(index), Offset: uint32(offset), Length: uint32(length)}
					nmsg := &message.Message{ID: message.IDRequest, Payload: message.NewRequestPayload(req)}
					_, err := message.Write(conn, nmsg)
					if err != nil {
						fmt.Println(err)
						return
					}
					currentBlock += 1

					if currentBlock == blocksPerPiece {
						currentPiece += 1
						currentBlock = 0
					}
				}
			case message.IDHave:
				have := int(msg.Payload.Have())

				// Detect peer downloaded pieces
				for index := range missing {
					if have == index {
						downloadable = append(downloadable, index)
					}
				}

				// Send our new state
				if len(downloadable) != 0 && !interesting {
					nmsg := &message.Message{ID: message.IDInterested}
					_, err := message.Write(conn, nmsg)
					if err != nil {
						fmt.Println(err)
						return
					}
					interesting = true
				}
			case message.IDBitfield:
				bitfield := msg.Payload.Bitfield()

				// Detect peer downloaded pieces
				for index := range missing {
					if bitfield.Get(index) {
						downloadable = append(downloadable, index)
					}
				}

				// Send our new state
				if len(downloadable) != 0 && !interesting {
					nmsg := &message.Message{ID: message.IDInterested}
					_, err := message.Write(conn, nmsg)
					if err != nil {
						fmt.Println(err)
						return
					}
					interesting = true
				}
			}
		}
	}(peers[0], missing, downloaded)

	for block := range downloaded {
		fmt.Println(block)
	}
}
