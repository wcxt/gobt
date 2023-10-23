package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/edwces/gobt"
	"github.com/edwces/gobt/handshake"
	"github.com/edwces/gobt/message"
)

const RequestLength = 16000

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

    
    missing := make(chan int, metainfo.Info.Length)
    downloaded := make(chan []byte, metainfo.Info.Length)
    
    // Connect to peers
    for _, peer := range peers {
        go func(peer gobt.AnnouncePeer, in chan int, out chan []byte) {
            // Establish conn
            conn, err := net.DialTimeout("tcp", peer.Addr(), 3 * time.Second) 
            if err != nil {
                fmt.Println(err)
                return
            }

            defer conn.Close()
            // Handshake
            hs := handshake.New(hash, clientID)
            handshake.Write(conn, hs)

            hs, err = handshake.Read(conn)
            if err != nil {
                fmt.Println(err)
                return
            }

            if hs.InfoHash != hash {
                return
            }
            
            // Message loop
            //choked := true
            interesting := false
            //choking := true
            //interested := false

            downloadable := []int{}
            //requests := []*message.Request{}
            for {
                msg, err := message.Read(conn)
                if err != nil {
                    fmt.Println(err)
                    return
                }

                switch msg.ID {
                case message.IDChoke:
                    //choking = true
                case message.IDUnchoke:
                    //choking = false
                    // dump all pipelined request for piece
                    //for _, request := range requests {
                    //    msg := &message.Message{ID: message.IDRequest, Payload: message.NewRequestPayload(*request)} 
                    //    _, err := message.Write(conn, msg)
                    //    if err != nil {
                    //        fmt.Println(err)
                    //        return
                    //    }
                    //}

                case message.IDInterested:
                    //interested = true
                case message.IDNotInterested:
                    //interested = false
                case message.IDHave:
                    have := int(msg.Payload.Have())
                    
                    // TODO: Change how we handle which pieces we can download
                    for index := range missing {
                        if have == index {
                            downloadable = append(downloadable, index)
                        }
                        missing<-index
                    }

                    if len(downloadable) != 0 && !interesting {
                        msg := &message.Message{ID: message.IDInterested}
                        _, err := message.Write(conn, msg)
                        if err != nil {
                            fmt.Println(err)
                            return
                        }
                    }
                case message.IDBitfield:
                    bitfield := msg.Payload.Bitfield()
                    
                    // TODO: Change how we handle which pieces we can download
                    for index := range missing {
                        if bitfield.Get(index) {
                            downloadable = append(downloadable, index)
                        }
                        missing<-index
                    }

                    if len(downloadable) != 0 && !interesting {
                        msg := &message.Message{ID: message.IDInterested}
                        _, err := message.Write(conn, msg)
                        if err != nil {
                            fmt.Println(err)
                            return
                        }
                    }
                case message.IDPiece:
                }    
            }
        }(peer, missing, downloaded)
    }

    for block := range downloaded {
        fmt.Println(block)
    }
}
