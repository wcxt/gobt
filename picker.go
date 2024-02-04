package gobt

import (
	"errors"

	"github.com/edwces/gobt/bitfield"
)

type Piece struct {
	block int
}

type Picker struct {
	states  map[int]*Piece
	ordered []int
}

// NewPicker creates picker with pieces to pick from.
func NewPicker(size int) *Picker {
	ordered := make([]int, size)

	for i := 0; i < size; i++ {
		ordered[i] = i
	}

	return &Picker{ordered: ordered, states: map[int]*Piece{}}
}

// Pick gets a new block from pieces that are available in bitfield.
func (p *Picker) Pick(have bitfield.Bitfield) (int, error) {
	pIndex, error := p.pickPiece(have)

	if error != nil {
		return 0, error
	}

	return pIndex, nil
}

// pickPiece returns and removes piece that is available in peer bitfield from picker ordered pieces.
func (p *Picker) pickPiece(have bitfield.Bitfield) (int, error) {
	for i, val := range p.ordered {
		if have, _ := have.Get(val); have {
			p.ordered = append(p.ordered[:i], p.ordered[i+1:]...)
			return val, nil
		}
	}

	return 0, errors.New("No pieces found")
}
