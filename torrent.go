package gobt

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"io"
	"net"
	"net/url"
	"strconv"
	"time"

	bencode "github.com/jackpal/bencode-go"
)

const DefaultPort = 6881

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

func (t *Torrent) GeneratePeerId() ([20]byte, error) {
    b := [20]byte{}
    _, err := rand.Read(b[:])

    return b, err
}

func (t *Torrent) GetAvailablePort() int {
    port := DefaultPort

    for {
        conn, err := net.DialTimeout("tcp", ":" + strconv.Itoa(port), time.Second * 3)
        if err != nil {
            return port
        }

        conn.Close()
        port++
    }
}

func (t *Torrent) BuildTrackerURL() (*url.URL, error) {
    infoHash, err := t.Info.Hash()
    if err != nil {
        return nil, err
    }

    peerId, err := t.GeneratePeerId()
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
    keys.Set("left", "0")

    uri.RawQuery = keys.Encode()
    
    return uri, err
} 


