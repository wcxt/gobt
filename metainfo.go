package gobt

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"io"
	"os"

	bencode "github.com/jackpal/bencode-go"
)

type Metainfo struct {
	Announce string       `bencode:"announce"`
	Info     MetainfoDict `bencode:"info"`
}

type MetainfoDict struct {
	Name        string `bencode:"name"`
	Length      int    `bencode:"length"`
	PieceLength int    `bencode:"piece length"`
	Pieces      string `bencode:"pieces"`
}

func readMetainfo(r io.Reader) (*Metainfo, error) {
	mi := &Metainfo{}
	err := bencode.Unmarshal(r, mi)
	if err != nil {
		return nil, err
	}

	return mi, nil
}

func Open(path string) (*Metainfo, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }

    return readMetainfo(file)

}

func (mid MetainfoDict) Hash() ([20]byte, error) {
	var buf bytes.Buffer
	err := bencode.Marshal(&buf, mid)
	if err != nil {
		return [20]byte{}, err
	}

	return sha1.Sum(buf.Bytes()), nil
}

func (mid MetainfoDict) Hashes() ([][20]byte, error) {
	byteStr := []byte(mid.Pieces)

	if len(byteStr)%20 != 0 {
		return [][20]byte{}, errors.New("pieces length not divisable by 20")
	}
    
	hashesLen := len(byteStr) / 20
	hashes := make([][20]byte, hashesLen)

	for i := 0; i < hashesLen; i++ {
		hashes[i] = [20]byte(byteStr[i:i+20])
	}

	return hashes, nil
}
