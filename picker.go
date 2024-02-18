package gobt

import (
	"errors"
	"math"

	"github.com/edwces/gobt/bitfield"
)

type PieceStatus int

const (
	PiecePending PieceStatus = iota
	PieceIncomplete
	PieceDone

	DefaultBlockSize = 16000
)

func PieceCount(tSize, pMaxSize int) int {
	return int(math.Ceil((float64(tSize) / float64(pMaxSize))))
}

func BlockCount(tSize, pMaxSize, pIndex int) int {
	pSize := math.Min(float64(pMaxSize), float64(tSize)-float64(pMaxSize)*float64(pIndex))
	return int(math.Ceil((float64(pSize) / float64(DefaultBlockSize))))
}

type Piece struct {
	blocks int

	pending []int
	done    []int
	status  PieceStatus
}

type Picker struct {
	tSize    int
	pMaxSize int

	states  map[int]*Piece
	ordered []int
}

// NewPicker creates picker with pieces to pick from.
func NewPicker(tSize, pMaxSize int) *Picker {
	count := PieceCount(tSize, pMaxSize)
	ordered := make([]int, count)

	for i := 0; i < count; i++ {
		ordered[i] = i
	}

	return &Picker{tSize: tSize, pMaxSize: pMaxSize, ordered: ordered, states: map[int]*Piece{}}
}

// Pick gets a new block from pieces that are available in bitfield.
func (p *Picker) Pick(have bitfield.Bitfield) (int, int, error) {
	pIndex, error := p.pickPiece(have)

	if error != nil {
		return 0, 0, error
	}

	bIndex := p.pickBlock(pIndex)

	return pIndex, bIndex, nil
}

func (p *Picker) MarkBlockDone(pIndex, bIndex int) {
	state := p.getState(pIndex)
	state.done = append(state.done, bIndex)

	if len(state.done) == state.blocks {
		state.status = PieceDone
	}
}

func (p *Picker) IsPieceDone(pIndex int) bool {
	state := p.getState(pIndex)

	return state.status == PieceDone
}

// Clear clears piece state and readds it to the picker
// NOTE: This method is unoptimized as it may cause loop where
//
//	the same peer/peers is constantly corrupting piece
func (p *Picker) Clear(pIndex int) {
	p.states[pIndex] = p.createState(pIndex)

	p.ordered = append(p.ordered[:2], p.ordered[1:]...)
	p.ordered[1] = pIndex
}

// Abort adds block to requests and optionally puts incomplete piece onto the top of picker
func (p *Picker) Abort(pIndex, bIndex int) {
	p.states[pIndex].pending = append(p.states[pIndex].pending, bIndex)

	if p.states[pIndex].status == PieceIncomplete {
		p.states[pIndex].status = PiecePending
		p.ordered = append([]int{pIndex}, p.ordered...)
	}
}

// pickPiece returns and removes piece that is available in peer bitfield from picker ordered pieces.
func (p *Picker) pickPiece(have bitfield.Bitfield) (int, error) {
	for _, val := range p.ordered {
		if have, _ := have.Get(val); have {
			return val, nil
		}
	}

	return 0, errors.New("No pieces found")
}

// removePiece returns true if succesfully removes piece from picker
func (p *Picker) removePiece(pIndex int) bool {
	for i, val := range p.ordered {
		if pIndex == val {
			p.ordered = append(p.ordered[:i], p.ordered[i+1:]...)
			return true
		}
	}
	return false
}

// pickBlock returns block index and removes piece from picker if all blocks have been requested
func (p *Picker) pickBlock(pIndex int) int {
	state := p.getState(pIndex)
	bIndex := state.pending[0]
	state.pending = state.pending[1:]

	if len(state.pending) == 0 {
		p.removePiece(pIndex)
		state.status = PieceIncomplete
		return bIndex
	}

	return bIndex
}

// Returns piece state or creates one if it doesn't exists
func (p *Picker) getState(pIndex int) *Piece {
	state, exists := p.states[pIndex]

	if !exists {
		state = p.createState(pIndex)
		p.states[pIndex] = state
	}

	return state
}

func (p *Picker) createState(pIndex int) *Piece {
	bCount := BlockCount(p.tSize, p.pMaxSize, pIndex)
	pending := []int{}
	done := []int{}
	for i := 0; i < bCount; i++ {
		pending = append(pending, i)
	}

	return &Piece{blocks: bCount, pending: pending, done: done, status: PiecePending}
}
