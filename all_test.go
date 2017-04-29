package blockchain

import (
	"fmt"
	"testing"
	"time"
)

func TestHashValidity(t *testing.T) {
	if !IsHashValid("000deadb33f", 3) {
		t.Error("hash validity check failed")
	}
	if IsHashValid("000deadb33f", 5) {
		t.Error("hash validity check failed")
	}
}

func TestBlockGeneration(t *testing.T) {
	numZeroes := 4
	txn := map[string]string{"b": "1"}
	b := NewBlock("deadb33f", txn, 5, 0)
	miner := NewMiner(numZeroes, 5)
	success := make(chan bool)
	cancelled := make(chan bool)
	go miner.GenerateValidBlock(b, success, cancelled)
	select {
	case <-success:
		fmt.Println("finished")
		if !IsHashValid(b.Hash(), numZeroes) {
			t.Error("block generated with invalid hash")
		}
		fmt.Println(b.Hash())
	case <-cancelled:
	}
}

func TestBlockCancellation(t *testing.T) {
	numZeroes := 8
	txn := map[string]string{"b": "1"}
	b := NewBlock("deadb33f", txn, 5, 0)
	miner := NewMiner(numZeroes, 5)
	success := make(chan bool)
	cancelled := make(chan bool)
	go miner.GenerateValidBlock(b, success, cancelled)
	go func() {
		time.Sleep(time.Second * 1)
		miner.CancelBlockGeneration()
	}()
	select {
	case <-success:
		t.Error("block generation should have cancelled.")
	case <-cancelled:
		fmt.Println("successfully cancelled")
	}
}

func TestBlockChainTipGenesis(t *testing.T) {
	blockchain := NewBlockChain("deadb33f")
	tip := blockchain.ChainTip([]string{})
	t.Log(tip)
	if tip != "deadb33f" {
		t.Error("chain tip did not return genesis in empty chain")
	}
}

func TestSingleChainBlockIntegration(t *testing.T) {
	blockchain := NewBlockChain("deadb33f")
	tip := blockchain.ChainTip([]string{})
	t.Log(tip)
	b := NewNoOpBlock(tip, 9)
	b.Nonce = 9000
	hash := b.Hash()
	t.Log(hash)
	blockchain.IntegrateBlock(b)
	t.Log("tips:", blockchain.tips)
	tip = blockchain.ChainTip([]string{})
	t.Log(tip)
	if tip != hash {
		t.Error("tip of blockchain not correct")
	}
	b2 := NewBlock(hash, map[string]string{"b": "1"}, 9, 87)
	h2 := b2.Hash()
	t.Log(h2)
	blockchain.IntegrateBlock(b2)
	t.Log("tips:", blockchain.tips)
	tip = blockchain.ChainTip([]string{})
	t.Log(tip)
	if tip != h2 {
		t.Error("tip of blockchain not correct")
	}
}

func TestChainForking(t *testing.T) {
	blockchain := NewBlockChain("deadb33f")
	b := NewNoOpBlock(blockchain.ChainTip([]string{}), 9)
	blockchain.IntegrateBlock(b)
	tip := blockchain.ChainTip([]string{})
	b1 := NewBlock(tip, map[string]string{"b": "1"}, 9, 87)
	b2 := NewBlock(tip, map[string]string{"b": "2"}, 9, 17)
	blockchain.IntegrateBlock(b1)
	blockchain.IntegrateBlock(b2)
	t.Log(blockchain.tips)
	if len(blockchain.tips) != 2 {
		t.Error("fork did not produce appropriate number of tips")
	}
}

func TestForkedChainLeastConflictingKeys(t *testing.T) {
	blockchain := NewBlockChain("deadb33f")
	b := NewNoOpBlock(blockchain.ChainTip([]string{}), 9)
	blockchain.IntegrateBlock(b)
	tip := blockchain.ChainTip([]string{})
	b1 := NewBlock(tip, map[string]string{"a": "1"}, 9, 1)
	b2 := NewBlock(tip, map[string]string{"b": "2"}, 9, 2)
	b3 := NewBlock(tip, map[string]string{"c": "3"}, 9, 3)
	b4 := NewBlock(tip, map[string]string{"d": "4"}, 9, 4)
	blockchain.IntegrateBlock(b1)
	blockchain.IntegrateBlock(b2)
	blockchain.IntegrateBlock(b3)
	blockchain.IntegrateBlock(b4)
	t.Log(blockchain.tips)
	if len(blockchain.tips) != 4 {
		t.Error("fork did not produce appropriate number of tips")
	}
	t.Log("b3 hash", b3.Hash())
	tip = blockchain.ChainTip([]string{"a", "b", "d"})
	t.Log("tip:", tip)
	if tip != b3.Hash() {
		t.Error("blockchain did not return tip with least conflicting keys")
	}
	b5 := NewNoOpBlock(tip, 9)
	blockchain.IntegrateBlock(b5)
	if len(blockchain.tips) != 1 {
		t.Error("should only have one tip")
	}
	if b5.Hash() != blockchain.ChainTip([]string{}) {
		t.Error("wrong chain tip")
	}
}

func TestGetFromBlockChain(t *testing.T) {
	blockchain := NewBlockChain("deadb33f")
	bl := NewNoOpBlock(blockchain.ChainTip([]string{}), 9)
	blockchain.IntegrateBlock(bl)
	tip := blockchain.ChainTip([]string{})
	b1 := NewBlock(tip, map[string]string{"a": "1"}, 9, 1)
	blockchain.IntegrateBlock(b1)
	b2 := NewBlock(blockchain.ChainTip(nil), map[string]string{"b": "2"}, 9, 2)
	blockchain.IntegrateBlock(b2)
	b3 := NewBlock(blockchain.ChainTip(nil), map[string]string{"c": "3"}, 9, 3)
	blockchain.IntegrateBlock(b3)
	b4 := NewBlock(blockchain.ChainTip(nil), map[string]string{"d": "4"}, 9, 4)
	blockchain.IntegrateBlock(b4)
	a := blockchain.Get("a")
	b := blockchain.Get("b")
	c := blockchain.Get("c")
	d := blockchain.Get("d")
	e := blockchain.Get("e")
	t.Log(a, b, c, d, e)
	if a != "1" {
		t.Error("wrong value for 'a' returned")
	}
	if b != "2" {
		t.Error("wrong value for 'b' returned")
	}
	if c != "3" {
		t.Error("wrong value for 'c' returned")
	}
	if d != "4" {
		t.Error("wrong value for 'd' returned")
	}
	if e != "" {
		t.Error("wrong value for 'e' returned")
	}
}

func TestProgenyCount(t *testing.T) {
	blockchain := NewBlockChain("deadb33f")
	bl := NewNoOpBlock(blockchain.ChainTip([]string{}), 9)
	blockchain.IntegrateBlock(bl)
	tip := blockchain.ChainTip([]string{})
	b1 := NewBlock(tip, map[string]string{"a": "1"}, 9, 1)
	blockchain.IntegrateBlock(b1)
	b2 := NewBlock(blockchain.ChainTip(nil), map[string]string{"b": "2"}, 9, 2)
	blockchain.IntegrateBlock(b2)
	b3 := NewBlock(blockchain.ChainTip(nil), map[string]string{"c": "3"}, 9, 3)
	blockchain.IntegrateBlock(b3)
	b4 := NewBlock(blockchain.ChainTip(nil), map[string]string{"d": "4"}, 9, 4)
	blockchain.IntegrateBlock(b4)
	if blockchain.progenyCount(b1.Hash()) != 3 {
		t.Error("wrong progeny count")
	}
	if blockchain.progenyCount(b3.Hash()) != 1 {
		t.Error("wrong progeny count")
	}
	b5 := NewBlock(b2.Hash(), map[string]string{"b": "2"}, 9, 2)
	blockchain.IntegrateBlock(b5)
	b6 := NewBlock(b2.Hash(), map[string]string{"b": "2"}, 9, 2)
	blockchain.IntegrateBlock(b6)
	b7 := NewBlock(b2.Hash(), map[string]string{"b": "2"}, 9, 2)
	blockchain.IntegrateBlock(b7)
	if len(blockchain.nodes[b2.Hash()].children) != 4 {
		t.Error("wrong children count")
	}
	if blockchain.progenyCount(b1.Hash()) != 3 {
		t.Error("wrong progeny count")
	}
	if blockchain.progenyCount(b2.Hash()) != 2 {
		t.Error("wrong progeny count")
	}
	if blockchain.progenyCount(b3.Hash()) != 1 {
		t.Error("wrong progeny count")
	}
	if blockchain.progenyCount(b5.Hash()) != 0 {
		t.Error("wrong progeny count")
	}
}
