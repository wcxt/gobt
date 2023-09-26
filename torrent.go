package gobt

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"

	bencode "github.com/jackpal/bencode-go"
)

const DefaultPort = 6881

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

func UnmarshalMetainfo(r io.Reader) (*Metainfo, error) {
	mi := &Metainfo{}
	err := bencode.Unmarshal(r, mi)
	if err != nil {
		return nil, err
	}

	return mi, nil
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

	for i := 0; i < len(byteStr); i += 20 {
		hashes = append(hashes, [20]byte(byteStr[i:i+20]))
	}

	return hashes, nil
}

type Torrent struct {
	Hash   [20]byte
	Hashes [][20]byte

	Downloaded int
	Uploaded   int
	Size       int
}

func NewTorrentFromMetainfo(mi *Metainfo) (*Torrent, error) {
	hash, err := mi.Info.Hash()
	if err != nil {
		return nil, err
	}

	hashes, err := mi.Info.Hashes()
	if err != nil {
		return nil, err
	}

	return &Torrent{
		Hash:   hash,
		Hashes: hashes,
		Size:   mi.Info.Length,
	}, nil
}

func Open(path string) (*Torrent, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	mi, err := UnmarshalMetainfo(file)
	if err != nil {
        return nil, err
	}

	return NewTorrentFromMetainfo(mi)
}

func RandomPeerId() ([20]byte, error) {
	b := [20]byte{}
	_, err := rand.Read(b[:])

	return b, err
}

func (t *Metainfo) BuildTrackerURL(peerId [20]byte) (*url.URL, error) {
	infoHash, err := t.Info.Hash()
	if err != nil {
		return nil, err
	}

	uri, err := url.Parse(t.Announce)
	if err != nil {
		return nil, err
	}

	keys := url.Values{}
	keys.Set("info_hash", string(infoHash[:]))
	keys.Set("peer_id", string(peerId[:]))
	keys.Set("port", strconv.Itoa(DefaultPort))
	keys.Set("uploaded", "0")
	keys.Set("downloaded", "0")
	keys.Set("left", strconv.Itoa(t.Info.Length))

	uri.RawQuery = keys.Encode()

	return uri, err
}

type TrackerResponse struct {
	Failure  string `bencode:"failure"`
	Interval int    `bencode:"interval"`
	Peers    []Peer `bencode:"peers"`
}

type Peer struct {
	PeerId string `bencode:"peer id"`
	Ip     string `bencode:"ip"`
	Port   int    `bencode:"port"`
}

func (t *Metainfo) GetPeers(peerId [20]byte) (*TrackerResponse, error) {
	uri, err := t.BuildTrackerURL(peerId)
	if err != nil {
		return nil, err
	}

	res, err := http.Get(uri.String())
	if err != nil {
		return nil, err
	}

	tres := &TrackerResponse{}
	err = bencode.Unmarshal(res.Body, tres)
	if err != nil {
		return nil, err
	}

	return tres, nil
}
