package gobt_test

import (
	"testing"

	"github.com/edwces/gobt"
	"github.com/edwces/gobt/bitfield"
)

const (
	TestTorrentPieceBlocks = 4
	TestTorrentTotalPieces = 25
	TestTorrentPieceLength = gobt.MaxBlockLength*TestTorrentPieceBlocks - gobt.MaxBlockLength/2
	TestTorrentLength      = TestTorrentPieceLength * TestTorrentTotalPieces
)

func TestPickerPick(t *testing.T) {
	bitfield := bitfield.New(TestTorrentTotalPieces)
	peerID := "1"
	for i := 10; i < TestTorrentTotalPieces; i++ {
		bitfield.Set(i)
	}

	t.Run("strict order", func(t *testing.T) {
		p := gobt.NewPicker(TestTorrentLength, TestTorrentPieceLength)
		want := 0

		for i := 0; i < 10*TestTorrentPieceBlocks; i++ {
			_, bi, err := p.Pick(bitfield, peerID)

			if err != nil {
				t.Fatalf("want nil, got err: %s", err.Error())
			}

			if bi != want {
				t.Fatalf("want %d, got %d", want, bi)
			}

			want = (want + 1) % TestTorrentPieceBlocks
		}
	})

	t.Run("random", func(t *testing.T) {
		p := gobt.NewPicker(TestTorrentLength, TestTorrentPieceLength)
		p.SetRandSeed(0)

		want := []int{19, 13, 20, 15, 18}

		for i := 0; i < 5*TestTorrentPieceBlocks; i++ {
			pi, _, err := p.Pick(bitfield, peerID)

			if err != nil {
				t.Fatalf("want nil, got err: %s", err.Error())
			}

			if pi != want[i/TestTorrentPieceBlocks] {
				t.Fatalf("want %d, got %d", want[i/TestTorrentPieceBlocks], pi)
			}
		}
	})

	t.Run("rarest", func(t *testing.T) {
		p := gobt.NewPicker(TestTorrentLength, TestTorrentPieceLength)
		p.SetRandSeed(0)

		p.IncrementPieceAvailability(15)

		want := 15

		for i := 0; i < 14*TestTorrentPieceBlocks; i++ {
			_, _, err := p.Pick(bitfield, peerID)

			if err != nil {
				t.Fail()
			}
		}

		for i := 0; i < 1*TestTorrentPieceBlocks; i++ {
			pi, _, err := p.Pick(bitfield, peerID)

			if err != nil {
				t.Fatalf("want nil, got err: %s", err.Error())
			}

			if pi != want {
				t.Fatalf("want %d, got %d", want, pi)
			}
		}
	})
}
