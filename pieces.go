package gobt

import (
	"errors"
	"sync"
)

type PieceQueue struct {
	sync.Mutex

	requests []bool
	done     []bool
}

func NewPieceQueue(length int) *PieceQueue {
	return &PieceQueue{
		requests: make([]bool, length),
		done:     make([]bool, length),
	}
}

func (pt *PieceQueue) MarkDone(index int) {
	pt.Lock()
	pt.done[index] = true
	pt.Unlock()
}

func (pt *PieceQueue) MarkRequested(index int) {
	pt.Lock()
	pt.requests[index] = true
	pt.Unlock()
}

func (pt *PieceQueue) MarkNotDone(index int) {
	pt.Lock()
	pt.done[index] = false
	pt.Unlock()
}

func (pt *PieceQueue) MarkNotRequested(index int) {
	pt.Lock()
	pt.requests[index] = false
	pt.Unlock()
}

func (pt *PieceQueue) Dequeue(bitfield Bitfield) (int, error) {
	pt.Lock()
	defer pt.Unlock()

	for i, requested := range pt.requests {
		if bitfield.Get(i) && !requested {
			return i, nil
		}
	}

    return -1, errors.New("No available pieces to deque from bitfield") 
}
