package blockchain

import (
	"fmt"
	"strconv"
)

type Block struct {
	ClientID string
	PrevHash string
	Txn      map[string]string // key value map
	NodeID   int
	Nonce    uint32
}

func NewBlock(prevHash string, txn map[string]string, nodeID int,
	nonce uint32) *Block {
	return &Block{
		PrevHash: prevHash,
		Txn:      txn,
		NodeID:   nodeID,
		Nonce:    nonce,
	}
}

func NewNoOpBlock(prevHash string, nodeID int) *Block {
	return &Block{
		PrevHash: prevHash,
		Txn:      map[string]string{},
		NodeID:   nodeID,
	}
}

func (b *Block) Keys() (keys []string) {
	for k := range b.Txn {
		keys = append(keys, k)
	}
	return
}

func (b *Block) Hash() string {
	return hash(b.String(true))
}

func (b *Block) IsNoOp() bool {
	return len(b.Txn) == 0
}

func (b *Block) String(withNonce bool) string {
	str := b.PrevHash + fmt.Sprint(b.Txn) + strconv.Itoa(b.NodeID)
	if withNonce {
		str += strconv.Itoa(int(b.Nonce))
	}
	return str
}
