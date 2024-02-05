package gobt

import (
	"errors"
	"math"

	"github.com/edwces/gobt/bitfield"
)

const DefaultBlockSize = 16000

func pieceCount(tSize, pMaxSize int) int {
	return int(math.Ceil((float64(tSize) / float64(pMaxSize))))
}

func blockCount(tSize, pMaxSize int) int {
	pSize := tSize % pMaxSize
	return int(math.Ceil((float64(pSize) / float64(DefaultBlockSize))))
}

type Piece struct {
	counter int
	max     int
}

type Picker struct {
	tSize    int
	pMaxSize int

	states  map[int]*Piece
	ordered []int
}

// NewPicker creates picker with pieces to pick from.
func NewPicker(tSize, pMaxSize int) *Picker {
	count := pieceCount(tSize, pMaxSize)
	ordered := make([]int, count)

	for i := 0; i < count; i++ {
		ordered[i] = i
	}

	return &Picker{tSize: tSize, pMaxSize: pMaxSize, ordered: ordered, states: map[int]*Piece{}}
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

// pickBlock returns block index and removes piece from picker if all blocks have been requested
// func (p *Picker) pickBlock(pIndex int) {}

// Returns piece state or creates one if it doesn't exists
func (p *Picker) getState(pIndex int) *Piece {
	state, exists := p.states[pIndex]

	if !exists {
		state = &Piece{counter: 0, max: blockCount(p.tSize, p.pMaxSize)}
		p.states[pIndex] = state
	}

	return state
}
