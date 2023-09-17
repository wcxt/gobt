package gobt

import "testing"

func TestInfoHash(t *testing.T) {
    ti := &TorrentInfo{
       Name: "Sample",
       Length: 684356,
       PieceLength: 25789,
       Pieces: "qwertytuiop1234567890!@#$%^&*()", 
    }

    hash, err := ti.Hash()
    expected := [20]byte{247, 186, 102, 52, 22, 198, 184, 213, 131, 86, 125, 66, 100, 12, 151, 184, 220, 160, 146, 91}   

    if err != nil {
        t.Fatalf("error = %s, want nil", err.Error())
    }
    if hash != expected {
        t.Fatalf("hash = %v, want %v", hash, expected)
    }
}

func TestPeerId(t *testing.T) {
    tr := &Torrent{}

    _, err := tr.PeerId()

    if err != nil {
        t.Fatalf("error = %s, want nil", err.Error())
    }
}
