package picker

import (
	"math"

	"github.com/edwces/gobt/bitfield"
)

const DefaultMaxBlockLength = 16000

type Picker interface {
    Pick(bitfield.Bitfield) (int, int)
}

type piece struct {
    index int

    bLeft int
    bPicked []int
}

type picker struct {
    pStates map[int]*piece
    pOrdered []int
    pPicked []int
    pMaxLength int

    length int
}

func New(length, pMaxLength int) Picker {
    p := &picker{length: length, pMaxLength: pMaxLength}
    
    p.pOrdered = []int{}
    p.pPicked = []int{}
    p.pStates = map[int]*piece{}

    for i := 0; i < p.pieceCount(); i++ {
        p.pOrdered = append(p.pOrdered, i)
    }

    return p
}

func (p *picker) pieceCount() int {
    return p.length / p.pMaxLength
}

func (p *picker) pieceLength(i int) int {
    return int(math.Min(float64(p.pMaxLength), float64(p.length - i * p.pMaxLength)))
}

func (p *picker) blockCount(i int) int {
    return p.pieceLength(i) / DefaultMaxBlockLength
}

func (p *picker) Pick(has bitfield.Bitfield) (int, int) {
    pi := p.pickPiece(has)
    if pi == -1 {
        return -1, -1
    }
    
    state := p.getPieceState(pi)
    bi := p.pickBlock(state)

    return pi, bi
}

func (p *picker) getPieceState(i int) *piece {
    state, exists := p.pStates[i]

    if !exists {
        bc := p.blockCount(i)
        bPicked := []int{}

        for j := 0; j < bc; j++ {
            bPicked = append(bPicked, j)
        }

        state = &piece{index: i, bLeft: bc, bPicked: bPicked}
        p.pStates[i] = state
    }

    return state
}

func (p *picker) pickBlock(state *piece) int {
    index := state.bPicked[0]
    state.bPicked = state.bPicked[1:]
    state.bLeft--

    // Stop tracking piece if all request have been made
    if state.bLeft == 0 {
         for j, i := range p.pPicked {
            if i == state.index {
                p.pPicked = append(p.pPicked[:j], p.pPicked[j+1:]...) 
                return index
            }
        }
    }

    return index
}

func (p *picker) pickPiece(has bitfield.Bitfield) int {
    // Strict Priority
    for _, index := range p.pPicked {
        if have, _ := has.Get(index); have {
            return index 
        }
    }

    // New Piece
    for i, index := range p.pOrdered {
        if have, _ := has.Get(index); have {
            p.pOrdered = append(p.pOrdered[:i], p.pOrdered[i+1:]...)
            p.pPicked = append([]int{index}, p.pPicked...)

            return index 
        }
    }

    return -1
}
