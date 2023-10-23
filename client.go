package gobt

import (
	"crypto/rand"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"

	bencode "github.com/jackpal/bencode-go"
)

const DefaultListenPort = 6881

func GenRandPeerID() ([20]byte, error) {
    b := [20]byte{}
	_, err := rand.Read(b[:])

	return b, err
}

type AnnounceResponse struct {
	Failure  string        `bencode:"failure"`
	Interval int           `bencode:"interval"`
	Peers    []AnnouncePeer `bencode:"peers"`
}

type AnnouncePeer struct {
	ID   string `bencode:"peer id"`
	IP   string `bencode:"ip"`
	Port int    `bencode:"port"`
}

func (ap *AnnouncePeer) Addr() string {
    return net.JoinHostPort(ap.IP, strconv.Itoa(ap.Port))
}

func buildRequestURL(uri string, hash [20]byte, peerID [20]byte, length int, port int) (*url.URL, error) {
    parsed, err := url.Parse(uri)
    if err != nil {
        return nil, err
    }

    query := url.Values{}
    query.Set("info_hash", string(hash[:]))
	query.Set("peer_id", string(peerID[:]))
	query.Set("port", strconv.Itoa(port))
	query.Set("uploaded", strconv.Itoa(0))
	query.Set("downloaded", strconv.Itoa(0))
	query.Set("left", strconv.Itoa(length))

    parsed.RawQuery = query.Encode()

    return parsed, nil
}

func GetAvailablePeers(uri string, hash [20]byte, peerID [20]byte, length int) ([]AnnouncePeer, error) {
    annUri, err := buildRequestURL(uri, hash, peerID, length, DefaultListenPort)
    fmt.Println(annUri.String())
    if err != nil {
        return nil, err
    }

    res, err := http.Get(annUri.String())
    if res != nil {
        defer res.Body.Close()
    }
    if err != nil {
        return nil, err
    }

    ann := &AnnounceResponse{}
	err = bencode.Unmarshal(res.Body, ann)
	if err != nil {
		return nil, err
	}

    return ann.Peers, nil
}


