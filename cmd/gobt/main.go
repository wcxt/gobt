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
	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	torrent, err := gobt.Parse(file)
	if err != nil {
		fmt.Println(err)
		return
	}

	peerId, err := gobt.RandomPeerId()
	if err != nil {
		fmt.Println(err)
		return
	}

	tres, err := torrent.GetPeers(peerId)
	if err != nil {
		fmt.Println(err)
		return
	}

	conn, err := wire.Dial(net.JoinHostPort(tres.Peers[0].Ip, strconv.Itoa(tres.Peers[0].Port)))
	if err != nil {
		fmt.Println(err)
		return
	}

	hash, err := torrent.Info.Hash()
	if err != nil {
		fmt.Println(err)
		return
	}
    
    hashes, err := torrent.Info.PiecesHashes()
    if err != nil {
        fmt.Println(err)
        return
    }

	err = conn.Handshake(hash, peerId)
	if err != nil {
		fmt.Println(err)
		return
	}

    bitfield, err := conn.RecvBitfield()
    if err != nil {
		fmt.Println(err)
		return
    }
}
