package picker

import (
	"errors"
	"math"

	"github.com/edwces/gobt/bitfield"
)

const DefaultMaxBlockLength = 16000

type Picker interface {
    Pick(bitfield.Bitfield) (*Block, error)
    Done(*Block) 
    Return(*Block) 
    Add(i int)
}

type Piece struct {
    Index int
    Done bool

    bOrdered []int
    bStates []*Block
}

type Block struct {
    Piece *Piece
    Index int
    Done bool

    Offset int
    Length int 
}

type picker struct {
    pStates map[int]*Piece
    pOrdered []int
    pPicked []int
    pMaxLength int

    length int
}

func New(length, pMaxLength int) Picker {
    p := &picker{length: length, pMaxLength: pMaxLength}
    
    p.pOrdered = []int{}
    p.pPicked = []int{}
    p.pStates = map[int]*Piece{}

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

func (p *picker) blockLength(pi, bi int) int {
    return int(math.Min(float64(DefaultMaxBlockLength), float64(p.pieceLength(pi) - bi * DefaultMaxBlockLength)))
}

func (p *picker) Return(block *Block) {
    block.Done = false
    block.Piece.bOrdered = append(block.Piece.bOrdered, block.Index)
    if block.Piece.Done {
        block.Piece.Done = false
        p.pPicked = append(p.pPicked, block.Piece.Index)
    }
}

func (p *picker) Add(i int) {
    p.piece(i)
    delete(p.pStates, i)

    p.pOrdered = append(p.pOrdered, i)
}

func (p *picker) Done(block *Block) {
    block.Done = true
    
    for _, b := range block.Piece.bStates {
        if b == nil || b.Done == false {
            return 
        }
    } 
    
    block.Piece.Done = true
    for j, index := range p.pPicked {
        if index == block.Piece.Index {
            p.pPicked = append(p.pPicked[:j], p.pPicked[j+1:]...)
            return
        }
    }
}

func (p *picker) Pick(has bitfield.Bitfield) (*Block, error) {
    pi := p.pickPiece(has)
    if pi == -1 {
        return nil, errors.New("Piece not found")
    }
    
    state := p.piece(pi)
    bi := p.pickBlock(state)

    block := &Block{Piece: state,
                    Index: bi,
                    Offset: bi * DefaultMaxBlockLength,
                    Length: p.blockLength(pi, bi)}

    state.bStates[block.Index] = block

    return block, nil
}

func (p *picker) piece(i int) *Piece {
    state, exists := p.pStates[i]

    if !exists {
        bc := p.blockCount(i)
        bOrdered := []int{}

        for j := 0; j < bc; j++ {
            bOrdered = append(bOrdered, j)
        }

        state = &Piece{Index: i, bOrdered: bOrdered, bStates: make([]*Block, bc)}
        p.pStates[i] = state
    }

    return state
}

func (p *picker) pickBlock(state *Piece) int {
    index := state.bOrdered[0]
    state.bOrdered = state.bOrdered[1:]

    // Stop tracking piece if all request have been made
    if len(state.bOrdered) == 0 {
         for j, i := range p.pPicked {
            if i == state.Index {
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
