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

    client := wire.Client{PeerId: peerId}
    conn, err := client.NewConnection(net.JoinHostPort(tres.Peers[0].Ip, strconv.Itoa(tres.Peers[0].Port)))
    if err != nil {
        fmt.Println(err)
        return
    }
    
    conn.Close()
    fmt.Printf("%v+\n", tres)
} 
