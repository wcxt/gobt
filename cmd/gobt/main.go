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

	conn, err := wire.Dial(net.JoinHostPort(tres.Peers[0].IP, strconv.Itoa(tres.Peers[0].Port)))
	if err != nil {
		fmt.Println(err)
		return
	}
    defer conn.Close()

	err = conn.Handshake(torrent.Hash, peerId)
	if err != nil {
		fmt.Println(err)
		return
	}

    bitfield := make(message.Bitfield, len(torrent.Hashes))
    piece := []*message.Block{}
    pieceSize := 0
    requestLength := 16000
    nextOffset := uint32(0)
    pending := make(chan *message.Request, 5)
    
    go func(){
        for {
            msg, err := conn.Recv()
            if err != nil {
                continue
            }

            if msg.KeepAlive {
                continue
            }

            if msg.ID == message.IDBitfield {
                bitfield = msg.Payload.Bitfield()
            } else if msg.ID == message.IDChoke {
                conn.PeerChoking = true
            } else if msg.ID == message.IDUnchoke {
                conn.PeerChoking = false	
            } else if msg.ID == message.IDPiece {
                b := msg.Payload.Block() 
                fmt.Printf("BLOCK Received at: {Index: %d, Offset: %d, Length: %d}\n", b.Index, b.Offset, len(b.Block))
                piece = append(piece, &b)
                pieceSize += int(requestLength)
                <-pending
            } else if msg.ID == message.IDHave {
                index := msg.Payload.Have() 
                bitfield.Set(int(index))
            }
        }
    }()

	for {
        if !conn.ClientInterested && bitfield.Get(0) {
            _, err = conn.SendInterested()
			if err != nil {
				fmt.Println(err)
				return
			}
            conn.ClientInterested = true
        }
        
        if !conn.PeerChoking && conn.ClientInterested && pieceSize < torrent.PieceLength && len(pending) < 5 {
            request := message.Request{
                Index: 0,
                Offset: nextOffset,
                Length: uint32(math.Min(float64(torrent.PieceLength - pieceSize), float64(requestLength))),
            }
            _, err = conn.SendRequest(request)
			if err != nil {
				fmt.Println(err)
				return
			}
            nextOffset += uint32(requestLength)
            pending <- &request
        }

	}
}
