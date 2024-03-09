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

func CalcPieceLength(tSize, pMaxSize int) int {
	return int(math.Ceil((float64(tSize) / float64(pMaxSize))))
}

func CalcBlockLength(tSize, pMaxSize, pIndex int) int {
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

	count := CalcPieceLength(length, maxPieceLength)
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

	pi, err := p.pickPiece(have)

	if err != nil {
		return 0, 0, err
	}

	bi := p.pickBlock(pi, peer)

	return pi, bi, nil
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

// pickPiece returns and removes piece that is available in peer bitfield from picker ordered pieces.
func (p *Picker) pickPiece(have bitfield.Bitfield) (int, error) {
	reqBoundary := 0
	for i, val := range p.ordered {
		piece := p.getPiece(val)

		if piece.status == PieceInQueue {
			reqBoundary = i
			break
		}

		if has, _ := have.Get(val); has {
			return val, nil
		}
	}

	if p.counter < RandomPieceEndCounter {
		return p.pickRandomPiece(have, reqBoundary)
	}

	return p.pickRarestPiece(have, reqBoundary)
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

func (p *Picker) pickRandomPiece(have bitfield.Bitfield, reqBoundary int) (int, error) {
	ordCopy := make([]int, len(p.ordered)-reqBoundary)
	copy(ordCopy, p.ordered[reqBoundary:])
	p.rand.Shuffle(len(ordCopy), func(i, j int) {
		ordCopy[i], ordCopy[j] = ordCopy[j], ordCopy[i]
	})

	for _, val := range ordCopy {
		if has, _ := have.Get(val); has {
			return val, nil
		}
	}

	return 0, errors.New("No piece found")
}

func (p *Picker) pickRarestPiece(have bitfield.Bitfield, reqBoundary int) (int, error) {
	for _, val := range p.ordered[reqBoundary:] {
		if has, _ := have.Get(val); has {
			return val, nil
		}
	}

	return 0, errors.New("No piece found")
}

// removePiece returns true if succesfully removes piece from picker
func (p *Picker) removePiece(pi int) bool {
	for i, val := range p.ordered {
		if pi == val {
			p.ordered = append(p.ordered[:i], p.ordered[i+1:]...)
			return true
		}
	}
	return false
}

// pickBlock returns block index and removes piece from picker if all blocks have been requested
func (p *Picker) pickBlock(pi int, peer string) int {
	piece := p.getPiece(pi)
	var bIndex int

	for bi, block := range piece.blocks {
		if block.status == BlockInQueue {
			bIndex = bi
			block.status = BlockPending
			block.peers = append(block.peers, peer)
			break
		}
	}

	// TEMP:
	isPiecePending := true

	for _, block := range piece.blocks {
		if block.status == BlockInQueue {
			isPiecePending = false
			break
		}
	}

	if isPiecePending {
		p.removePiece(pi)
		piece.status = PiecePending
		return bIndex
	}

	if piece.status == PieceInQueue {
		piece.status = PieceInProgress
		p.counter++
		p.update()
	}

	return bIndex
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
	count := CalcBlockLength(p.length, p.maxPieceLength, pi)
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
