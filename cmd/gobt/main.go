package main

import (
	"fmt"
	"math"
	"net"
	"os"
	"strconv"

	"github.com/edwces/gobt"
	"github.com/edwces/gobt/wire"
	"github.com/edwces/gobt/wire/message"
)

const RequestLength = 16000

func main() {
	path := os.Args[1]

	torrent, err := gobt.Open(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	peerId, err := gobt.RandomPeerId()
	if err != nil {
		fmt.Println(err)
		return
	}

	tres, err := torrent.RequestPeers(peerId)
	if err != nil {
		fmt.Println(err)
		return
	}
    
    addr := net.JoinHostPort(tres.Peers[0].IP, strconv.Itoa(tres.Peers[0].Port))
	peer, err := wire.Dial(addr, torrent.Hash, peerId)
	if err != nil {
		fmt.Println(err)
		return
	}
    defer peer.Close()
    
    // Used for storing pieces that still have to be downloaded
    pieces := torrent.Hashes
    // Used for tracking pieces that are currently are downloading
    downloading := [][20]byte{}

    go func(){
        for {
            msg, err := peer.Recv() 
            if err != nil {
                continue
            }

            if msg.ID == message.IDPiece {
                block := msg.Payload.Block()
                fmt.Printf("PEER: Block{Index: %d, Offset: %d}\n", block.Index, block.Offset)
            }

            peer.Handle(msg) 
        }
    }()

    for {
        // NOTE: Maybe a channel ?
        downloable := []int{}

        for i := range pieces {
            if peer.Has(i) {
                downloable = append(downloable, i)
            }
        }
        
        // should be a separate goroutine
        if len(downloable) != 0 && !peer.Interesting {
            err := peer.Interest()
			if err != nil {
				fmt.Println(err)
				return
			}
            fmt.Println("CLIENT: Interested")
            
            // downloading as long as we are interested
            for len(downloable) != 0 {
                if !peer.Choking {
                    i := downloable[len(downloable)-1]        
                    downloading = append(downloading, pieces[i])
                    pieces = append(pieces[0:i], pieces[i+1:]...)

                    offset := 0
                    size := torrent.PieceLength

                    for offset != size {
                        length := int(math.Min(float64(size - offset), float64(RequestLength)))
                        err := peer.Request(i, offset, length)
                        if err != nil {
                            fmt.Println(err)
                            return
                        }

                        fmt.Printf("CLIENT: Request{Index: %d, Offset: %d, Length: %d}\n", i, offset, length)
                        offset += length
                    }

                    downloable = downloable[:len(downloable)-1]

                    // TODO: Check hash
                }
            }

            err = peer.Uninterest() 
            if err != nil {
				fmt.Println(err)
				return
			}
            fmt.Println("CLIENT: Not Interested")
        }
	}
}
