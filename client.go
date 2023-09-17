package gobt

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
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
	PieceLength int `bencode:"piece length"`
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

func (ti TorrentInfo) Hash() ([20]byte, error) {
    var buf bytes.Buffer
    err := bencode.Marshal(&buf, ti)
    if err != nil {
        return [20]byte{}, err
    }

    return sha1.Sum(buf.Bytes()), nil
}

func (t *Torrent) PeerId() ([20]byte, error) {
    b := [20]byte{}
    _, err := rand.Read(b[:])

    return b, err
}
