package main

import (
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

    conn.Close()
}
