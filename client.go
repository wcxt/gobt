package gobt

import (
	"io"

	bencode "github.com/jackpal/bencode-go"
)

type Torrent struct {
	Announce string
	Info     TorrentInfo
}

type TorrentInfo struct {
	Name        string
	Length      int
	PieceLength int
	Pieces      string
}

func Parse(r io.Reader) (*Torrent, error) {
    torrent := &Torrent{}  
	err := bencode.Unmarshal(r, torrent)
    if err != nil {
        return nil, err
    }

	return torrent, nil
}
