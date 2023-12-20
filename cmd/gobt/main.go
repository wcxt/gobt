package main

import (
	"crypto/sha1"
	"fmt"
	"math"
	"os"

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

    pieceRequests := make([]bool, len(hashes))
    
    conn, err := gobt.DialTimeout(peers[0].Addr())
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


    for {
        msg, err := conn.ReadMsg()
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
                blocksHash := sha1.Sum(blockBuffer)
                if blocksHash == hashes[block.Index] {
                    fmt.Println("GOT PIECE WITH CORRECT HASH")
                }
                // else mark piece as not being downloaded
                blockBuffer = []byte{}
            }

            requestQueue = requestQueue[1:]

            for i := len(requestQueue); i < MaxPipelinedRequests; i++ {
                if len(downloadable) != 0 && interesting {
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
                        downloadable = downloadable[1:]
                        currentPiece = downloadable[0]
                        // mark piece as being downloaded by this peer

                        currentBlock = 0
                    }
                }
            }

            if len(downloadable) == 0 && interesting {
                _, err := conn.WriteNotInterested()
                if err != nil {
                    fmt.Println(err)
                    return
                }
                interesting = false
            }

        case message.IDUnchoke:
            for i := len(requestQueue); i < MaxPipelinedRequests; i++ {
                if len(downloadable) != 0 && interesting {
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
                        downloadable = downloadable[1:]
                        currentPiece = downloadable[0]

                        currentBlock = 0
                    }
                }
            }
        case message.IDHave:
            have := int(msg.Payload.Have())

            // Detect peer downloaded pieces
            for index, processing := range pieceRequests {
                if have == index && !processing {
                    downloadable = append(downloadable, index)
                }
            }

            // Send our new state
            if len(downloadable) != 0 && !interesting {
                conn.WriteInterested()
                if err != nil {
                    fmt.Println(err)
                    return
                }
                interesting = true

                currentPiece = downloadable[0]
            }
        case message.IDBitfield:
            bitfield := msg.Payload.Bitfield()

            // Detect peer downloaded pieces
            for index, processing := range pieceRequests {
                if bitfield.Get(index) && !processing {
                    downloadable = append(downloadable, index)
                }
            }

            // Send our new state
            if len(downloadable) != 0 && !interesting {
                conn.WriteInterested()
                if err != nil {
                    fmt.Println(err)
                    return
                }
                interesting = true

                currentPiece = downloadable[0]
            }
        }
    }
}
