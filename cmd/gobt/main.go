package main

import (
	"fmt"
	"os"

	"github.com/edwces/gobt"
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

    uri, err := torrent.BuildTrackerURL()
    if err != nil {
        fmt.Println(err)
        return
    }

    fmt.Println(uri.String())
} 
