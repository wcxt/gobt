package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"os"
	"strconv"

	"github.com/edwces/gobt"
	"github.com/edwces/gobt/wire"
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

    var bitfield wire.Bitfield
    piece := []*wire.Block{}
    pieceSize := 0
    requestLength := 16000
    nextOffset := uint32(0)

	for {
        msg, err := conn.Recv()
		if err != nil {
            continue
		}

		if msg.KeepAlive {
            continue
		}

        if msg.ID == wire.MessageBitfield {
            bitfield = msg.Payload
        } else if msg.ID == wire.MessageChoke {
			conn.PeerChoking = true
		} else if msg.ID == wire.MessageUnchoke {
		    conn.PeerChoking = false	
		} else if msg.ID == wire.MessagePiece {
            index := binary.BigEndian.Uint32(msg.Payload[0:4])
            begin := binary.BigEndian.Uint32(msg.Payload[4:8])

            b := &wire.Block{Index: index, Offset: begin, Bytes: msg.Payload[8:]}
            fmt.Printf("BLOCK Received at: {Index: %d, Offset: %d, Length: %d}\n", b.Index, b.Offset, len(b.Bytes))
            piece = append(piece, b)
            pieceSize += int(requestLength)
		}

        if !conn.ClientInterested && bitfield != nil && bitfield.Get(0) {
            _, err = conn.SendInterested()
			if err != nil {
				fmt.Println(err)
				return
			}
            conn.ClientInterested = true
        }
        
        if !conn.PeerChoking && conn.ClientInterested && pieceSize < torrent.PieceLength {
            requestLength := math.Min(float64(torrent.PieceLength - pieceSize), float64(requestLength))
            _, err = conn.SendRequest(0, uint32(nextOffset), uint32(requestLength))
			if err != nil {
				fmt.Println(err)
				return
			}
            nextOffset += uint32(requestLength)
        }

	}
}
