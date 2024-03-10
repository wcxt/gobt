package gobt

import (
	"errors"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/edwces/gobt/bitfield"
	"golang.org/x/exp/slices"
)

type PieceStatus int
type BlockStatus int

const (
	PieceInQueue PieceStatus = iota
	PieceInProgress
	PiecePending
	PieceDone

	BlockInQueue BlockStatus = iota
	BlockPending
	BlockDone

	MaxBlockLength        = 16000
	RandomPieceEndCounter = 5
)

func CalcPieceCount(tSize, pMaxSize int) int {
	return int(math.Ceil((float64(tSize) / float64(pMaxSize))))
}

func CalcBlockCount(tSize, pMaxSize, pIndex int) int {
	pSize := math.Min(float64(pMaxSize), float64(tSize)-float64(pMaxSize)*float64(pIndex))
	return int(math.Ceil((float64(pSize) / float64(MaxBlockLength))))
}

type Block struct {
	status BlockStatus

	peers []string
}

type Piece struct {
	blocks []*Block
	status PieceStatus

	availability int
}

type Picker struct {
	counter        int
	length         int
	maxPieceLength int

	pieces  map[int]*Piece
	ordered []int
	rand    *rand.Rand

	sync.Mutex
}

// NewPicker creates picker with pieces to pick from.
func NewPicker(length, maxPieceLength int) *Picker {
	rand := rand.New(rand.NewSource(time.Now().Unix()))

	count := CalcPieceCount(length, maxPieceLength)
	ordered := make([]int, count)

	for i := 0; i < count; i++ {
		ordered[i] = i
	}

	return &Picker{length: length, maxPieceLength: maxPieceLength, ordered: ordered, pieces: map[int]*Piece{}, rand: rand}
}

func (p *Picker) SetRandSeed(seed int64) {
	p.rand.Seed(seed)
}

// Pick gets a new block from pieces that are available in bitfield.
func (p *Picker) Pick(have bitfield.Bitfield, peer string) (int, int, error) {
	p.Lock()
	defer p.Unlock()

	if len(p.ordered) == 0 {
		return p.pickEndgame(have, peer)
	}

	pi, bi, err := p.pickStrict(have, peer)
	if err == nil {
		return pi, bi, nil
	}

	if p.counter < RandomPieceEndCounter {
		return p.pickRandom(have, peer)
	}

	return p.pickRarest(have, peer)
}

func (p *Picker) IncrementPieceAvailability(pi int) {
	p.Lock()
	defer p.Unlock()

	piece := p.getPiece(pi)
	piece.availability++

	p.update()
}

func (p *Picker) DecrementAvailability(have bitfield.Bitfield) {
	p.Lock()
	defer p.Unlock()

	have.Range(func(i int, val bool) bool {
		if !val {
			return true
		}

		piece := p.getPiece(i)
		piece.availability--
		return true
	})

	p.update()
}

func (p *Picker) IncrementAvailability(have bitfield.Bitfield) {
	p.Lock()
	defer p.Unlock()

	have.Range(func(i int, val bool) bool {
		if !val {
			return true
		}

		piece := p.getPiece(i)
		piece.availability++
		return true
	})

	p.update()
}

func (p *Picker) MarkBlockDone(pi int, bi int, peer string) {
	p.Lock()
	defer p.Unlock()

	piece := p.getPiece(pi)
	piece.blocks[bi].status = BlockDone
	piece.blocks[bi].peers = slices.DeleteFunc(piece.blocks[bi].peers, func(e string) bool { return e == peer })

	if p.isPieceDone(piece) {
		piece.status = PieceDone
	}
}

func (p *Picker) isPieceDone(piece *Piece) bool {
	for _, block := range piece.blocks {
		if block.status != BlockDone {
			return false
		}
	}

	return true
}

func (p *Picker) IsPieceDone(pi int) bool {
	p.Lock()
	defer p.Unlock()

	piece := p.getPiece(pi)
	return piece.status == PieceDone
}

func (p *Picker) IsBlockDownloaded(pi int, bi int) bool {
	p.Lock()
	defer p.Unlock()

	piece := p.getPiece(pi)
	return len(piece.blocks[bi].peers) != 0
}

func (p *Picker) FailPendingPiece(pi int) {
	p.Lock()
	defer p.Unlock()

	piece := p.getPiece(pi)
	piece.status = PieceInQueue
	piece.blocks = p.newBlocksForPiece(pi)

	p.ordered = append(p.ordered, pi)
	p.update()
}

func (p *Picker) FailPendingBlock(pi int, bi int, peer string) {
	p.Lock()
	defer p.Unlock()

	piece := p.getPiece(pi)
	piece.blocks[bi].status = BlockInQueue
	piece.blocks[bi].peers = slices.DeleteFunc(piece.blocks[bi].peers, func(e string) bool { return e == peer })

	if piece.status == PiecePending {
		piece.status = PieceInProgress
		p.ordered = append(p.ordered, pi)
		p.update()
	}
}

func (p *Picker) pickStrict(have bitfield.Bitfield, peer string) (int, int, error) {
	for _, pi := range p.ordered {
		piece := p.getPiece(pi)

		if piece.status != PieceInProgress {
			break
		}

		if has, _ := have.Get(pi); has {
			return pi, p.pickNextBlock(piece, pi, peer), nil
		}
	}

	return 0, 0, errors.New("No pieces found")
}

func (p *Picker) pickRandom(have bitfield.Bitfield, peer string) (int, int, error) {
	shuffled := make([]int, len(p.ordered))
	shuffled = append(shuffled, p.ordered...)

	p.rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	for _, pi := range shuffled {
		piece := p.getPiece(pi)

		if piece.status != PieceInQueue {
			continue
		}

		if has, _ := have.Get(pi); has {
			return pi, p.pickNextBlock(piece, pi, peer), nil
		}
	}

	return 0, 0, errors.New("No piece found")
}

func (p *Picker) pickRarest(have bitfield.Bitfield, peer string) (int, int, error) {
	for _, pi := range p.ordered {
		piece := p.getPiece(pi)

		if piece.status != PieceInQueue {
			continue
		}

		if has, _ := have.Get(pi); has {
			return pi, p.pickNextBlock(piece, pi, peer), nil
		}
	}

	return 0, 0, errors.New("No piece found")
}

func (p *Picker) pickNextBlock(piece *Piece, pi int, peer string) int {
	for bi, block := range piece.blocks {
		if block.status != BlockInQueue {
			continue
		}

		block.status = BlockPending
		block.peers = append(block.peers, peer)

		if piece.status == PieceInQueue {
			piece.status = PieceInProgress
			p.counter++
			p.update()
		} else if p.isPiecePending(piece) {
			piece.status = PiecePending
			p.removePiece(pi)
		}

		return bi
	}

	return -1
}

func (p *Picker) isPiecePending(piece *Piece) bool {
	for _, block := range piece.blocks {
		if block.status == BlockInQueue {
			return false
		}
	}

	return true
}

func (p *Picker) removePiece(pi int) bool {
	for i, val := range p.ordered {
		if pi == val {
			p.ordered = append(p.ordered[:i], p.ordered[i+1:]...)
			return true
		}
	}
	return false
}

func (p *Picker) pickEndgame(have bitfield.Bitfield, peer string) (int, int, error) {
	for pi, piece := range p.pieces {
		if piece.status == PiecePending {
			has, _ := have.Get(pi)
			if !has {
				continue
			}

			bi, err := p.pickBlockEndgame(pi, peer)
			if err != nil {
				continue
			}

			piece.blocks[bi].peers = append(piece.blocks[bi].peers, peer)
			return pi, bi, nil
		}
	}

	return 0, 0, errors.New("No piece found")
}

func (p *Picker) pickBlockEndgame(pi int, peer string) (int, error) {
	piece := p.getPiece(pi)

	for bi, block := range piece.blocks {
		if block.status == BlockPending && !slices.Contains(block.peers, peer) {
			return bi, nil
		}
	}

	return 0, errors.New("No blocks found")
}

// Returns piece state or creates one if it doesn't exists
func (p *Picker) getPiece(pi int) *Piece {
	piece, exists := p.pieces[pi]

	if !exists {
		blocks := p.newBlocksForPiece(pi)
		piece = &Piece{status: PieceInQueue, blocks: blocks}
		p.pieces[pi] = piece
	}

	return piece
}

func (p *Picker) newBlocksForPiece(pi int) []*Block {
	count := CalcBlockCount(p.length, p.maxPieceLength, pi)
	blocks := make([]*Block, count)
	for i := 0; i < count; i++ {
		block := &Block{status: BlockInQueue}
		blocks[i] = block
	}

	return blocks
}

func (p *Picker) update() {
	slices.SortFunc(p.ordered, func(a, b int) int {
		piece1 := p.getPiece(a)
		piece2 := p.getPiece(b)

		if piece1.status == PieceInProgress && piece2.status == PieceInQueue {
			return -1
		} else if piece2.status == PieceInProgress && piece1.status == PieceInQueue {
			return 1
		} else if piece1.status == piece2.status {
			// Sort based on availability
			if piece1.availability < piece2.availability {
				return -1
			} else if piece2.availability < piece1.availability {
				return 1
			} else {
				return 0
			}
		}

		return 0
	})
}
