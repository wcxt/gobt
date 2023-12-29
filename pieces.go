package gobt

import (
	"errors"

	"github.com/edwces/gobt/bitfield"
)

type PiecePicker struct {
    pieces []int
    picked []int
}

func NewPiecePicker(size int) *PiecePicker {
    pieces := make([]int, size)
    
    for i := 0; i < size; i++ {
        pieces[i] = i
    }

    return &PiecePicker{pieces: pieces, picked: []int{}}
}

func (pp *PiecePicker) Pick(has bitfield.Bitfield) (int, error) {
    for i, index := range pp.pieces {
        if got, _ := has.Get(index); got {
            pp.pieces = append(pp.pieces[:i], pp.pieces[i+1:]...)
            return index, nil
        }
    }

    return 0, errors.New("No piece found")
}

func (pp *PiecePicker) Add(pi int) {
    for i, index := range pp.picked {
        if index == pi {
            pp.picked = append(pp.picked[:i], pp.picked[i+1:]...)
            pp.pieces = append([]int{pi}, pp.pieces...)
        }
    }

    pp.pieces = append([]int{pi}, pp.pieces...)
}

func (pp *PiecePicker) Remove(pi int) {
     for i, index := range pp.picked {
        if index == pi {
            pp.picked = append(pp.picked[:i], pp.picked[i+1:]...)
            return 
        }
    }

    for i, index := range pp.pieces {
        if index == pi {
            pp.pieces = append(pp.pieces[:i], pp.pieces[i+1:]...)
            return 
        }
    }
}
