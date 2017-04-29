package blockchain

import (
	"crypto/sha256"
	"fmt"
	"strconv"
	"sync"
)

// Miner is responsible for (no-op) block generation
type Miner struct {
	sync.Mutex
	shouldCancel bool
	numZeroes    int
	parentNodeID int
}

func NewMiner(numZeroes, parentNodeID int) *Miner {
	return &Miner{
		numZeroes:    numZeroes,
		parentNodeID: parentNodeID,
	}
}

func (m *Miner) CancelBlockGeneration() {
	m.Lock()
	defer m.Unlock()
	m.shouldCancel = true
}

// GenerateValidBlock generates a block given the hash of the data.
// ** modifies the block that is passed as input
func (m *Miner) GenerateValidBlock(b *Block, success, cancelled chan bool) {
	var nonce uint32
	data := b.String(false)
	m.Lock()
	m.shouldCancel = false
	m.Unlock()

	for {
		m.Lock()
		if m.shouldCancel {
			m.Unlock()
			cancelled <- true
			return
		}
		m.Unlock()
		inputstr := data + strconv.Itoa(int(nonce))
		hash := hash(inputstr)
		if IsHashValid(hash, m.numZeroes) {
			b.Nonce = nonce
			success <- true
			return
		}
		nonce++
	}
}

func IsHashValid(hash string, numZeroes int) (isValid bool) {
	isValid = true
	for _, char := range hash[:numZeroes] {
		if string(char) != "0" {
			isValid = false
			break
		}
	}
	return
}

func hash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}
