package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/url"
	"time"

	b "github.com/PyMarcus/blockchain/block"
	ts "github.com/PyMarcus/blockchain/transaction"
)

/*
Blockchain class is responsible for managing the chain.
It will store transactions and have some helper methods
for adding new blocks to the chain.
*/
type Blockchain struct {
	Chain               []b.Block
	CurrentTransactions []ts.Transaction
}

var nodes Set

func GenerateBlockchain() *Blockchain {
	bc := &Blockchain{}
	nodes = Set{}
	bc.NewBlock(int64(100), "1")
	return bc
}

// Creates a new Block and adds it to the chain
func (bc *Blockchain) NewBlock(proof int64, previusHash string) b.Block {

	block := b.Block{
		Index:        len(bc.Chain) + 1,
		Timestamp:    time.Now(),
		Transactions: bc.CurrentTransactions,
		Proof:        proof,
		PreviousHash: previusHash,
	}
	log.Println(block)

	bc.CurrentTransactions = []ts.Transaction{}

	bc.Chain = append(bc.Chain, block)

	return block
}

// Adds a new transaction to the list of transactions
func (bc *Blockchain) NewTransaction(sender string, recipient string, amount int) int {
	bc.CurrentTransactions = append(bc.CurrentTransactions, ts.Transaction{
		Sender:    sender,
		Recipient: recipient,
		Amount:    amount,
	})

	index := bc.LastBlock().Index + 1
	return index
}

// Creates a SHA-256 hash of a Block
func (bc Blockchain) Hash(block b.Block) string {
	hasher := sha256.New()
	hasher.Write([]byte(fmt.Sprintf("%+v", block)))

	sum := hasher.Sum(nil)
	return hex.EncodeToString(sum)
}

// returns the last Block in the chain
func (bc Blockchain) LastBlock() b.Block {
	return bc.Chain[len(bc.Chain)-1]
}

// Simple Proof of Work Algorithm (POW)
func (bc Blockchain) ProofOfWork(lastProof int64) int64 {
	proof := int64(0)
	for {
		if bc.valid(lastProof, proof) {
			break
		}
		proof++
	}

	return proof
}

// Validates the Proof: Does hash(last_proof, proof) contain 4 leading zeroes?
func (bc Blockchain) valid(lastProof, proof int64) bool {
	guess := []byte(fmt.Sprintf("%d%d", lastProof, proof))
	hasher := sha256.New()
	hasher.Write([]byte(guess))
	hash := hasher.Sum(nil)
	return hex.EncodeToString(hash)[:4] == "0000"
}

// Add a new node to the list of nodes
func (bc Blockchain) registerNode(address string) {
	urlParserd, err := url.Parse(address)
	if err != nil {
		log.Println("Fail to parse url")
	}
	nodes.Add(urlParserd.Host)
}
