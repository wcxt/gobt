package main

import (
	"encoding/binary"
	"fmt"
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

	err = conn.Handshake(torrent.Hash, peerId)
	if err != nil {
		fmt.Println(err)
		return
	}

	bitfield, err := conn.RecvBitfield()
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = conn.SendInterested()
	if err != nil {
		fmt.Println(err)
		return
	}

	if !bitfield.Get(0) {
		fmt.Println("Does not have piece: " + strconv.Itoa(0))
		return
	}

	for {
		msg, err := conn.Recv()
		if err != nil {
            continue
		}

		fmt.Printf("%v+\n", msg)

		if msg.KeepAlive {
            continue
		}

		if msg.ID == wire.MessageChoke {
			_, err = conn.SendInterested()
			if err != nil {
				fmt.Println(err)
				return
			}
		} else if msg.ID == wire.MessageUnchoke {
			_, err = conn.SendRequest(0, 0, uint32(2000))
			if err != nil {
				fmt.Println(err)
				return
			}
		} else if msg.ID == wire.MessagePiece {
            index := binary.BigEndian.Uint32(msg.Payload[0:4])
            begin := binary.BigEndian.Uint32(msg.Payload[4:8])

            b := &wire.Block{Index: index, Offset: begin, Bytes: msg.Payload[8:]}
            fmt.Printf("GOT BLOCK: %v+\n", b)
			break
		}

	}

	conn.Close()
}
