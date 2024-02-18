package gobt

import (
	"crypto/sha1"
	"fmt"
	"math"
)

// NEEDS:

// - Fast writing block to piece
// - easy way to determine if full / fast access
// - way to verify

func PieceSize(tSize, pMaxSize, pIndex int) int {
	return int(math.Min(float64(pMaxSize), float64(tSize)-float64(pMaxSize)*float64(pIndex)))
}

type Storage struct {
	bufs [][]byte

	tSize    int
	pMaxSize int
}

func NewStorage(tSize, pMaxSize int) *Storage {
	size := PieceCount(tSize, pMaxSize)
	return &Storage{bufs: make([][]byte, size), tSize: tSize, pMaxSize: pMaxSize}
}

func (s *Storage) SaveAt(pIndex int, block []byte, offset int) {
	buf := s.GetPieceData(pIndex)
	copy(buf[offset:], block)
}

func (s *Storage) GetPieceData(pIndex int) []byte {
	buf := s.bufs[pIndex]

	if buf == nil {
		buf = make([]byte, PieceSize(s.tSize, s.pMaxSize, pIndex))
		s.bufs[pIndex] = buf
	}

	return buf
}

func (s *Storage) Verify(pIndex int, hash [20]byte) bool {
	buf := s.GetPieceData(pIndex)
	pHash := sha1.Sum(buf)

	fmt.Printf("PHASH: %v \n", pHash)
	fmt.Printf("HASH: %v \n", hash)

	return pHash == hash
}
