package gobt

import (
	"errors"
	"fmt"
	"sync"

	"github.com/edwces/gobt/bitfield"
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
    count := 0
    for _, val := range pt.done {
        if val {
            count++
        }
    }
    fmt.Printf("DONE: %d\n", count)

	pt.Unlock()
}

func (pt *PieceQueue) MarkRequested(index int) {
	pt.Lock()
	pt.requests[index] = true
    count := 0
    for _, val := range pt.requests {
        if val {
            count++
        }
    }
    fmt.Printf("REQUESTS: %d\n", count)


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

func (pt *PieceQueue) Dequeue(bitfield bitfield.Bitfield) (int, error) {
	pt.Lock()
	defer pt.Unlock()

	for i, requested := range pt.requests {
        val, err := bitfield.Get(i)
        
        if err != nil {
            return -1, errors.New("No available pieces to deque from bitfield")
        }

		if val && !requested {
			return i, nil
		}
	}

	return -1, errors.New("No available pieces to deque from bitfield")
}
