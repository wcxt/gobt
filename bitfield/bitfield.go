package bitfield

import (
	"fmt"
	"math"
)

type bitfield struct {
    field []byte
    size int
}

type Bitfield interface {
    Replace([]byte) error
    Set(int, bool) error
    Get(int) (bool, error)
}

func New(size int) Bitfield {
    length := int(math.Ceil(float64(size) / 8))

    return &bitfield{
        field: make([]byte, length),
        size: size,
    }
}

func (bf *bitfield) Replace(data []byte) error {
    if len(data) != len(bf.field) {
        return fmt.Errorf("invalid replace data size: %d", len(data)) 
    }

    bf.field = data
    return nil
}

func (bf *bitfield) Set(i int, val bool) error {
    if i >= bf.size || i < 0 {
        return fmt.Errorf("invalid index value: %d", i)
    }

	byteI := i / 8
	bitI := i % 8

    if val {
	    bf.field[byteI] |= 0b10000000 >> bitI
    } else {
        bf.field[byteI] &= 0b01111111 >> bitI
    }

    return nil
}

func (bf bitfield) Get(i int) (bool, error) {
    if i >= bf.size || i < 0 {
        return false, fmt.Errorf("invalid index value: %d", i)
    }

	byteI := i / 8
	bitI := i % 8
	val := int(bf.field[byteI] & (0b10000000 >> bitI))

	return val != 0, nil
}
