package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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

var nodes *Set

func GenerateBlockchain() *Blockchain {
	bc := &Blockchain{}
	nodes = &Set{}
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
func (bc Blockchain) RegisterNode(address string) {
	urlParserd, err := url.Parse(address)
	if err != nil {
		log.Println("Fail to parse url")
	}
	log.Println("Registering node ", urlParserd.Host)
	nodes.Add(urlParserd.Host)
}

// Determine if a given blockchain is valid
func (bc Blockchain) validChain(chain []b.Block) bool {
	lastBlock := chain[0]
	currentIndex := 1
	for i := currentIndex; currentIndex < len(chain); i++ {
		block := chain[currentIndex]
		if block.PreviousHash != bc.Hash(lastBlock) {
			return false
		}

		if !bc.valid(lastBlock.Proof, block.Proof) {
			return false
		}
	}
	return true
}

// it's a Consensus Algorithm, it resolves conflicts
// by replacing the chain with the longest one in the network.
func (bc Blockchain) SolveConflicts() bool {
	neighbours := *nodes
	var newChain []b.Block

	maxLen := len(bc.Chain)

	for node, _ := range neighbours {
		response, err := http.Get(fmt.Sprintf("http://%v/chain", node))


		if err != nil {
			log.Println("Fail to get chain in solveConflicts")
			continue
		}

		if response.StatusCode == http.StatusOK {
			body, err := ioutil.ReadAll(response.Body)
			if err != nil {
				log.Println("Fail to read the body", err)
				continue
			}
			var data map[string]any
			if err := json.Unmarshal(body, &data); err != nil {
				log.Println("fail to decode JSON:", err)
				continue
			}
			
			defer response.Body.Close()

			leng := int(data["length"].(float64))			
			if data["chain"] != nil {
				chain, err := bc.parseBlocks(data["chain"])
				if err != nil {
					log.Printf("Errpr: %v", err)
					continue
				}
				if leng > maxLen && bc.validChain(chain) {
					maxLen = leng
					newChain = chain
				}
			}
		}

	}

	if len(newChain) > 0 {
		bc.Chain = newChain
		return true
	}
	return false
}

func (bc Blockchain) parseBlocks(chainInterface interface{}) ([]b.Block, error) {
	var chain []b.Block

	blocksInterface, ok := chainInterface.([]interface{})
	if !ok {
		return nil, fmt.Errorf("fail []interface{}")
	}

	for _, blockInterface := range blocksInterface {
		blockMap, ok := blockInterface.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf(" map[string]interface{}")
		}

		index := int(blockMap["index"].(float64))
		timestamp, err := time.Parse(time.RFC3339, blockMap["timestamp"].(string))
		if err != nil {
			return nil, fmt.Errorf("timestamp: %v", err)
		}

		transactionsInterface := blockMap["transactions"].([]interface{})
		var transactions []ts.Transaction
		for _, transactionInterface := range transactionsInterface {
			transactionMap, ok := transactionInterface.(map[string]interface{})
			if !ok {
				return nil, nil
			}

			sender := transactionMap["sender"].(string)
			recipient := transactionMap["recipient"].(string)
			amount := int(transactionMap["amount"].(float64))

			currentTransaction := ts.Transaction{
				Sender:    sender,
				Recipient: recipient,
				Amount:    amount,
			}

			transactions = append(transactions, currentTransaction)
		}

		proof := int64(blockMap["proof"].(float64))
		previousHash := blockMap["previous_hash"].(string)

		currentBlock := b.Block{
			Index:        index,
			Timestamp:    timestamp,
			Transactions: transactions,
			Proof:        proof,
			PreviousHash: previousHash,
		}

		chain = append(chain, currentBlock)
	}

	return chain, nil
}
