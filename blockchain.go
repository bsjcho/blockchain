package blockchain

import "math"

type BlockChain struct {
	ghash         string // genesis hash
	ghashChildren []string
	nodes         map[string]*node // map[blockhash]*Block
	tips          map[string]*node
	purgatory     []*Block
}

func NewBlockChain(ghash string) *BlockChain {
	b := &BlockChain{
		ghash:         ghash,
		tips:          map[string]*node{},
		nodes:         map[string]*node{},
		ghashChildren: []string{},
		purgatory:     []*Block{},
	}
	return b
}

// ChainTip returns the hash (prev-hash for new blocks) of the tip
// of the longest chain in the BlockChain
/*
In the case that there are several (longest) chains, the node should
(1) pick the one that does not cause a transaction abort for the current txn
block it is generating, or if no txn block is being generated or none cause
an abort for the existing transaction, then
(2) pick among the chains uniformly at random.
*/
func (b *BlockChain) ChainTip(currentKeys []string) (hash string) {
	if len(b.tips) == 0 {
		return b.ghash
	}
	conflictMin := math.MaxInt32
	for h, node := range b.tips {
		conflicts := b.countKeyConflicts(currentKeys, node)
		if conflicts < conflictMin {
			conflictMin = conflicts
			hash = h
		}
	}
	return hash
}

// GetChildren - parentHash is either an
// empty string to indicate that the client wants to retrieve the
// SHA 256 hash of the genesis block. Or, parentHash is a string
// identifying the hexadecimal SHA 256 hash of one of the blocks
// in the block-chain. In this case the return value should be the
// string representations of SHA 256 hash values of all of the
// children blocks that have the block identified by parentHash as
// their prev-hash value.
func (b *BlockChain) GetChildren(parentHash string) (children []string) {
	if parentHash == "" {
		return []string{b.ghash}
	}
	if parentHash == b.ghash {
		return b.ghashChildren
	}
	node := b.nodes[parentHash]
	return node.children
}

// Get returns a value associated with a key
// Traverses chain to get last seen key modification.
// TODO - correct, but inefficient.
func (b *BlockChain) Get(key string) (value string) {
	tip := b.ChainTip(nil)
	value = b.getHelper(tip, key)
	return
}

func (b *BlockChain) GetTxID(bhash string) int {
	return b.nodes[bhash].length
}

func (b *BlockChain) IntegrateBlock(block *Block) {
	_, ok := b.nodes[block.PrevHash]
	if !ok && block.PrevHash != b.ghash {
		b.addToPurgatory(block)
		return
	}
	var prevLength int
	if block.PrevHash == b.ghash {
		prevLength = 0
	} else {
		prevLength = b.prevNodeChainLength(block)
	}
	n := newNode(prevLength, block)
	b.nodes[block.Hash()] = n
	b.addAsChild(block)
	b.updateTips(n)
	b.addFromPurgatoryBlocksWithPrevHash(block.Hash())
}

// IsValidated determines whether a block has at least
// x blocks following it in the chain.
func (b *BlockChain) IsValidated(bhash string, validateNum int) bool {
	if b.isInPurgatory(bhash) {
		return false
	}
	return b.progenyCount(bhash) >= validateNum
}

func (b *BlockChain) IsOnLongestChain(bhash string) bool {
	if b.isInPurgatory(bhash) {
		return false
	}
	length := b.nodes[bhash].length
	return length+b.progenyCount(bhash) == b.MaxChainLength()
}

func (b *BlockChain) MaxChainLength() int {
	max := 0
	for _, node := range b.tips {
		if node.length > max {
			max = node.length
		}
	}
	return max
}

/*
Private
*/

type node struct {
	length   int
	block    *Block
	children []string // hash of children nodes
}

func newNode(prevLength int, block *Block) *node {
	return &node{
		length: prevLength + 1,
		block:  block,
	}
}

func (b *BlockChain) addToPurgatory(block *Block) {
	b.purgatory = append(b.purgatory, block)
}

func (b *BlockChain) addFromPurgatoryBlocksWithPrevHash(hash string) {
	children := []*Block{}
	newPurgatory := []*Block{}
	for _, block := range b.purgatory {
		if block.PrevHash == hash {
			children = append(children, block)
		} else {
			newPurgatory = append(newPurgatory, block)
		}
	}
	for _, childBlock := range children {
		b.IntegrateBlock(childBlock)
	}
	b.purgatory = newPurgatory
}

func (b *BlockChain) addAsChild(block *Block) {
	parent := block.PrevHash
	if parent == b.ghash {
		b.ghashChildren = append(b.ghashChildren, block.Hash())
	} else {
		node := b.nodes[parent]
		node.children = append(node.children, block.Hash())
	}
}

func (b *BlockChain) countKeyConflicts(currentKeys []string,
	n *node) (count int) {
	for _, key := range currentKeys {
		_, ok := n.block.Txn[key]
		if ok {
			count++
		}
	}
	return
}

func (b *BlockChain) getHelper(hash, key string) string {
	if hash == b.ghash {
		return ""
	}
	node := b.nodes[hash]
	val, ok := node.block.Txn[key]
	if ok {
		return val
	}
	return b.getHelper(node.block.PrevHash, key)
}

func (b *BlockChain) isInPurgatory(bhash string) bool {
	for _, block := range b.purgatory {
		if bhash == block.Hash() {
			return true
		}
	}
	return false
}

func (b *BlockChain) prevNodeChainLength(block *Block) int {
	n, ok := b.nodes[block.PrevHash]
	if !ok {
		return -1
	}
	return n.length
}

func (b *BlockChain) progenyCount(bhash string) int {
	return b.progenyCountHelper(bhash, 0)
}

func (b *BlockChain) progenyCountHelper(bhash string, current int) int {
	children := b.nodes[bhash].children
	if len(children) == 0 {
		return current
	}
	var max int
	for _, child := range children {
		count := b.progenyCountHelper(child, current)
		if count > max {
			max = count
		}
	}
	return max + 1
}

func (b *BlockChain) pruneTipsLessThan(length int) {
	newTips := map[string]*node{}
	for _, node := range b.tips {
		if node.length == length {
			newTips[node.block.Hash()] = node
		}
	}
	b.tips = newTips
}

func (b *BlockChain) updateTips(n *node) {
	delete(b.tips, n.block.PrevHash)
	b.tips[n.block.Hash()] = n
	length := b.MaxChainLength()
	b.pruneTipsLessThan(length)
}
