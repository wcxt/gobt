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

const (
    RequestLength = 16000
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
        
        for {
            msg, err := message.Read(conn)
            if err != nil {
                fmt.Println(err)
                return
            }

            switch msg.ID {
            //case message.IDChoke:
            //case message.IDUnchoke:
            //case message.IDInterested:
            //case message.IDNotInterested:
            //case message.IDPiece:
            case message.IDHave:
                have := int(msg.Payload.Have())
                
                for index := range missing {
                    if have == index {
                        downloadable = append(downloadable, index)
                    }
                }
                
                if len(downloadable) != 0 && !interesting {
                    msg := &message.Message{ID: message.IDInterested}
                    _, err := message.Write(conn, msg)
                    if err != nil {
                        fmt.Println(err)
                        return
                    }
                    interesting = true
                }
            case message.IDBitfield:
                bitfield := msg.Payload.Bitfield()
                
                for index := range missing {
                    if bitfield.Get(index) {
                        downloadable = append(downloadable, index)
                    }
                }
                
                if len(downloadable) != 0 && !interesting {
                    msg := &message.Message{ID: message.IDInterested}
                    _, err := message.Write(conn, msg)
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
