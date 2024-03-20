package gobt

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"io"

	bencode "github.com/jackpal/bencode-go"
)

const HashSize = sha1.Size

type Metainfo struct {
	Announce string `bencode:"announce"`
	Info     struct {
		Name        string `bencode:"name"`
		Length      int    `bencode:"length"`
		PieceLength int    `bencode:"piece length"`
		Pieces      string `bencode:"pieces"`
	} `bencode:"info"`
}

func UnmarshalMetainfo(r io.Reader) (*Metainfo, error) {
	mi := &Metainfo{}
	err := bencode.Unmarshal(r, mi)
	if err != nil {
		return nil, err
	}

	return mi, nil
}

func (m Metainfo) InfoHash() ([HashSize]byte, error) {
	var buf bytes.Buffer

	err := bencode.Marshal(&buf, m.Info)
	if err != nil {
		return [HashSize]byte{}, err
	}

	return sha1.Sum(buf.Bytes()), nil
}

func (m Metainfo) PieceHashes() ([][HashSize]byte, error) {
	hashesBytes := []byte(m.Info.Pieces)

	if len(hashesBytes)%HashSize != 0 {
		return [][HashSize]byte{}, errors.New("piece hashes length not divisable by hash size")
	}

	hashesCount := len(hashesBytes) / HashSize
	hashes := make([][HashSize]byte, hashesCount)

	for i := 0; i < hashesCount; i++ {
		start := i * HashSize
		end := i*HashSize + HashSize
		hashes[i] = [HashSize]byte(hashesBytes[start:end])
	}

	return hashes, nil
}
