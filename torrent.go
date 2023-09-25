package gobt

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	bencode "github.com/jackpal/bencode-go"
)

const DefaultPort = 6881

type Torrent struct {
	Announce string      `bencode:"announce"`
	Info     TorrentInfo `bencode:"info"`
}

type TorrentInfo struct {
	Name        string `bencode:"name"`
	Length      int    `bencode:"length"`
	PieceLength int    `bencode:"piece length"`
	Pieces      string `bencode:"pieces"`
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

func RandomPeerId() ([20]byte, error) {
	b := [20]byte{}
	_, err := rand.Read(b[:])

	return b, err
}

func (t *Torrent) GetAvailablePort() int {
	port := DefaultPort

	for {
		conn, err := net.DialTimeout("tcp", ":"+strconv.Itoa(port), time.Second*3)
		if err != nil {
			return port
		}

		conn.Close()
		port++
	}
}

func (t *Torrent) BuildTrackerURL(peerId [20]byte) (*url.URL, error) {
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
	keys.Set("port", strconv.Itoa(t.GetAvailablePort()))
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

func (t *Torrent) GetPeers(peerId [20]byte) (*TrackerResponse, error) {
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

func (t *TorrentInfo) PiecesHashes() ([][20]byte, error) {
    bytePieces := []byte(t.Pieces)
    if len(bytePieces) % 20 != 0 {
        return nil, errors.New("Incorrect length of pieces")
    }

    pHashLen := len(bytePieces) / 20
    pHashes := make([][20]byte, pHashLen)

    for i := range pHashes {
        pHashes[i] = [20]byte(bytePieces[i:i+20])
    }

    return pHashes, nil
}
