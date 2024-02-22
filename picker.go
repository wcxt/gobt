package gobt

import (
	"errors"
	"math"
	"math/rand"
	"sync"

	"github.com/edwces/gobt/bitfield"
	"golang.org/x/exp/slices"
)

type PieceStatus int

const (
	PieceInQueue PieceStatus = iota
	PieceRequesting
	PieceResolving
	PieceDone

	DefaultBlockSize      = 16000
	RandomPieceEndCounter = 5
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

	queue     []int
	resolving []int
	done      []int
	status    PieceStatus

	availability int
}

type Picker struct {
	pCounter int
	tSize    int
	pMaxSize int

	states  map[int]*Piece
	ordered []int

	sync.Mutex
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
	p.Lock()
	defer p.Unlock()

	pIndex, error := p.pickPiece(have)

	if error != nil {
		return 0, 0, error
	}

	bIndex := p.pickBlock(pIndex)

	return pIndex, bIndex, nil
}

func (p *Picker) IncrementPieceAvailability(pIndex int) {
	p.Lock()
	defer p.Unlock()

	state := p.getState(pIndex)
	state.availability++

	p.orderPieces()
}

func (p *Picker) DecrementAvailability(have bitfield.Bitfield) {
	p.Lock()
	defer p.Unlock()

	// TEMP Workaround, Should probably use some built-in func in bitfield
	count := PieceCount(p.tSize, p.pMaxSize)
	temp := make([]int, count)

	for i := range temp {
		if has, _ := have.Get(i); has {
			state := p.getState(i)
			state.availability--
		}
	}

	p.orderPieces()
}

func (p *Picker) IncrementAvailability(have bitfield.Bitfield) {
	p.Lock()
	defer p.Unlock()

	// TEMP Workaround, Should probably use some built-in func in bitfield
	count := PieceCount(p.tSize, p.pMaxSize)
	temp := make([]int, count)

	for i := range temp {
		if has, _ := have.Get(i); has {
			state := p.getState(i)
			state.availability++
		}
	}

	p.orderPieces()
}

func (p *Picker) MarkBlockDone(pIndex, bIndex int) {
	p.Lock()
	defer p.Unlock()

	state := p.getState(pIndex)
	state.done = append(state.done, bIndex)

	for i, b := range state.resolving {
		if b == bIndex {
			state.resolving = append(state.resolving[:i], state.resolving[i+1:]...)
		}
	}

	if len(state.done) == state.blocks {
		state.status = PieceDone
	}
}

func (p *Picker) IsPieceDone(pIndex int) bool {
	p.Lock()
	defer p.Unlock()

	state := p.getState(pIndex)

	return state.status == PieceDone
}

// MarkPieceInQueue clears piece state and readds it to the picker
// NOTE: This method is unoptimized as it may cause loop where
//
//	the same peer/peers is constantly corrupting piece
func (p *Picker) MarkPieceInQueue(pIndex int) {
	p.Lock()
	defer p.Unlock()

	p.states[pIndex] = p.createState(pIndex)

	p.ordered = append(p.ordered, pIndex)
	p.orderPieces()
}

// MarkBlockInQueue adds block to requests and optionally puts incomplete piece onto the top of picker
func (p *Picker) MarkBlockInQueue(pIndex, bIndex int) {
	p.Lock()
	defer p.Unlock()

	p.states[pIndex].queue = append(p.states[pIndex].queue, bIndex)

	for i, b := range p.states[pIndex].resolving {
		if b == bIndex {
			p.states[pIndex].resolving = append(p.states[pIndex].resolving[:i], p.states[pIndex].resolving[i+1:]...)
		}
	}

	if p.states[pIndex].status == PieceResolving {
		p.states[pIndex].status = PieceRequesting
		p.ordered = append(p.ordered, pIndex)
		p.orderPieces()
	}
}

// pickPiece returns and removes piece that is available in peer bitfield from picker ordered pieces.
func (p *Picker) pickPiece(have bitfield.Bitfield) (int, error) {
	reqBoundary := 0

	for i, val := range p.ordered {
		state := p.getState(val)

		if state.status == PieceInQueue {
			reqBoundary = i
			break
		}

		if have, _ := have.Get(val); have {
			return val, nil
		}
	}

	if p.pCounter < RandomPieceEndCounter {
		return p.pickRandomPiece(have, reqBoundary)
	}

	return p.pickRarestPiece(have, reqBoundary)
}

func (p *Picker) pickRandomPiece(have bitfield.Bitfield, reqBoundary int) (int, error) {
	ordCopy := make([]int, len(p.ordered)-reqBoundary)
	copy(ordCopy, p.ordered[reqBoundary:])
	rand.Shuffle(len(ordCopy), func(i, j int) {
		ordCopy[i], ordCopy[j] = ordCopy[j], ordCopy[i]
	})

	for _, val := range ordCopy {
		if have, _ := have.Get(val); have {
			return val, nil
		}
	}

	return 0, errors.New("No piece found")
}

func (p *Picker) pickRarestPiece(have bitfield.Bitfield, reqBoundary int) (int, error) {
	for _, val := range p.ordered[reqBoundary:] {
		if have, _ := have.Get(val); have {
			return val, nil
		}
	}

	return 0, errors.New("No piece found")
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
	bIndex := state.queue[0]
	state.queue = state.queue[1:]
	state.resolving = append(state.resolving, bIndex)

	if len(state.queue) == 0 {
		p.removePiece(pIndex)
		state.status = PieceResolving
		return bIndex
	}

	if state.status == PieceInQueue {
		state.status = PieceRequesting
		p.pCounter++
		p.orderPieces()
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

	return &Piece{blocks: bCount, queue: pending, done: done, status: PieceInQueue}
}

func (p *Picker) orderPieces() {
	slices.SortFunc(p.ordered, func(a, b int) int {
		aState := p.getState(a)
		bState := p.getState(b)

		if aState.status == PieceRequesting && bState.status == PieceInQueue {
			return -1
		} else if bState.status == PieceRequesting && aState.status == PieceInQueue {
			return 1
		} else if aState.status == bState.status {
			// Sort based on availability
			if aState.availability < bState.availability {
				return -1
			} else if bState.availability < aState.availability {
				return 1
			} else {
				return 0
			}
		}

		return 0
	})
}
