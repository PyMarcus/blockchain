package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"

	bc "github.com/PyMarcus/blockchain/blockchain"
	ts "github.com/PyMarcus/blockchain/transaction"
)

var b *bc.Blockchain

func GlobalNodeIdentifier() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

// create a new transaction to a block
func handleTransactionsNew(writer http.ResponseWriter, request *http.Request) {
	if request.Method == "POST" {
		var t ts.Transaction

		if err := json.NewDecoder(request.Body).Decode(&t); err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		index := b.NewTransaction(t.Sender, t.Recipient, t.Amount)
		json.NewEncoder(writer).Encode(map[string]int{"Success": index})

	}
}

// mine a new block
func handleMine(writer http.ResponseWriter, request *http.Request) {
	if request.Method == "GET" {
		lastBlock := b.LastBlock()
		lastProof := lastBlock.Proof
		proof := b.ProofOfWork(lastProof)

		b.NewTransaction("0", GlobalNodeIdentifier(), 1)

		previousHash := b.Hash(lastBlock)
		block := b.NewBlock(proof, previousHash)

		var response = make(map[string]interface{})

		response["message"] = "New block forged!"
		response["index"] = block.Index
		response["transactions"] = block.Transactions
		response["proof"] = block.Proof
		response["previous_hash"] = block.PreviousHash

		json.NewEncoder(writer).Encode(response)
	}
}

// return the full Blockchain
func handleChain(writer http.ResponseWriter, request *http.Request) {
	if request.Method == "GET" {
		var response = make(map[string]interface{})

		response["chain"] = b.Chain
		response["length"] = len(b.Chain)

		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(response)
	}
}

// Consensus Algorithm, which resolves any conflicts—to ensure a node has the correct chain.
func handleNodeResolve(writer http.ResponseWriter, request *http.Request) {
	if request.Method == "GET" {
		var response = make(map[string]interface{})

		replaced := b.SolveConflicts()
		if replaced {

			response["chain"] = b.Chain
			response["message"] = "Chain are authoritative"

			writer.Header().Set("Content-Type", "application/json")
			json.NewEncoder(writer).Encode(response)
		} else {
			response["new_chain"] = b.Chain
			response["message"] = "Chain was replaced"

			writer.Header().Set("Content-Type", "application/json")
			json.NewEncoder(writer).Encode(response)
		}
	}
}

// accept a list of new nodes in the form of URLs.
func handleNodeRegister(writer http.ResponseWriter, request *http.Request) {
	if request.Method == "POST" {
		var values map[string][]string
		decoder := json.NewDecoder(request.Body)
		err := decoder.Decode(&values)
		if err != nil {
			http.Error(writer, "fail to decode JSON", http.StatusBadRequest)
			return
		}

		nodes, ok := values["nodes"]
		if !ok {
			http.Error(writer, "put valid node list", http.StatusBadRequest)
			return
		}

		for _, n := range nodes {
			b.RegisterNode(n)
		}

		var response = make(map[string]interface{})

		response["message"] = "New node has been created!"
		response["total"] = len(nodes)
		response["nodes"] = nodes
		writer.Header().Set("Content-Type", "application/json")

		json.NewEncoder(writer).Encode(response)
	}
}

func Start() {

	b = bc.GenerateBlockchain()
	http.HandleFunc("/transactions/new", handleTransactionsNew)
	http.HandleFunc("/mine", handleMine)
	http.HandleFunc("/chain", handleChain)
	http.HandleFunc("/nodes/register", handleNodeRegister)
	http.HandleFunc("/nodes/resolve", handleNodeResolve)

	log.Println("Listening on localhost:8080...")

	http.ListenAndServe(":5001", nil)
}
