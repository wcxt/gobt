package gobt

import (
	"net"
	"strconv"
	"testing"
)

func TestInfoHash(t *testing.T) {
    ti := &TorrentInfo{
       Name: "Sample",
       Length: 684356,
       PieceLength: 25789,
       Pieces: "qwertytuiop1234567890!@#$%^&*()", 
    }

    hash, err := ti.Hash()
    want := [20]byte{247, 186, 102, 52, 22, 198, 184, 213, 131, 86, 125, 66, 100, 12, 151, 184, 220, 160, 146, 91}   

    if err != nil {
        t.Fatalf("error = %s, want nil", err.Error())
    }
    if hash != want {
        t.Fatalf("hash = %v, want %v", hash, want)
    }
}

func TestRandomPeerId(t *testing.T) {
    _, err := RandomPeerId()

    if err != nil {
        t.Fatalf("error = %s, want nil", err.Error())
    }
}

func TestGetAvailablePort(t *testing.T) {
    tr := &Torrent{}
    
    t.Run("Open", func(t *testing.T) {
        port := tr.GetAvailablePort()
        want := DefaultPort

        if port != want {
            t.Fatalf("port = %d, want = %d", port, want)
        }
    })

    t.Run("Closed", func(t *testing.T) {
        listener, err := net.Listen("tcp", ":" + strconv.Itoa(DefaultPort))
        if err != nil {
            t.Fail()
        }
        defer listener.Close()

        port := tr.GetAvailablePort()
        want := DefaultPort + 1
        
        if port != want {
            t.Fatalf("port = %d, want = %d", port, want)
        }
    })
}
