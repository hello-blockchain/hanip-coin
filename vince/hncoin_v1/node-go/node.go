package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"
)

type BlockChain struct {
	chain        []Block
	transactions []Transaction
	nodes        []string
}

type Transaction struct {
	sender   string  `json:"sender"`
	receiver string  `json:"receiver"`
	amount   float64 `json:"amount"`
}

type Block struct {
	Index        int           `json:"index"`
	Timestamp    time.Time     `json:"timestamp"`
	Proof        float64       `json:"proof"`
	PreviousHash [32]byte      `json:"previous_hash"`
	Transactions []Transaction `json:"transactions"`
}

func (bc *BlockChain) CreateBlock(proof float64, previousHash [32]byte) Block {
	block := Block{
		Index:        0,
		Timestamp:    time.Now(),
		Proof:        proof,
		PreviousHash: previousHash,
		Transactions: bc.transactions,
	}
	bc.transactions = []Transaction{}
	bc.chain = append(bc.chain, block)
	return block
}

func (bc *BlockChain) GetPreviousBlock() Block {
	if len(bc.chain) > 1 {
		return bc.chain[len(bc.chain)-1]
	} else {
		return bc.chain[0]
	}
}

func (bc *BlockChain) createHash(newProof float64, previousProof float64) [32]byte {
	encodedProof := strconv.Itoa(int(math.Pow(newProof, 2)) - int(math.Pow(previousProof, 2)))
	hashOperation := sha256.Sum256([]byte(encodedProof))
	return hashOperation
}

func (bc *BlockChain) ProofOfWork(previousProof float64) float64 {
	newProof := float64(1)
	checkProof := false
	for !checkProof {
		encodedProof := strconv.Itoa(int(math.Pow(newProof, 2)) - int(math.Pow(previousProof, 2)))
		hashOperation := sha256.Sum256([]byte(encodedProof))
		if string(hashOperation[:4]) == "0000" {
			checkProof = true
		} else {
			newProof += 1
		}
	}
	fmt.Println(newProof)
	return newProof
}

func (bc *BlockChain) hash(block Block) [32]byte {
	result, _ := json.Marshal(block)
	return sha256.Sum256(result)
}

func (bc *BlockChain) IsChainValid(chain []Block) bool {
	prevBlock := chain[0]
	blockIndex := 1
	var result bool

	for blockIndex < len(chain) {
		block := chain[blockIndex]

		if block.PreviousHash != bc.hash(prevBlock) {
			break
		}

		encodedProof := strconv.Itoa(int(math.Pow(block.Proof, 2)) - int(math.Pow(prevBlock.Proof, 2)))
		hashOperation := sha256.Sum256([]byte(encodedProof))
		if string(hashOperation[:4]) != "0000" {
			break
		}

		prevBlock = block
		blockIndex += 1
		result = true

	}
	return result
}

func (bc *BlockChain) AddTransaction(transaction Transaction) int {
	bc.transactions = append(bc.transactions, transaction)
	return bc.GetPreviousBlock().Index + 1
}

func (bc *BlockChain) AddNode(addr string) {
	bc.nodes = append(bc.nodes, addr)
}

func (bc *BlockChain) ReplaceChain() bool {
	var longestChain []Block
	maxLen := len(bc.chain)
	var chain ResponseGetChain

	for _, node := range bc.nodes {
		response := getReqJson(node + "getChain")
		if response.StatusCode != 200 {
			return false
		}
		body, _ := ioutil.ReadAll(response.Body)
		_ = json.Unmarshal(body, &chain)
		length := chain.Length
		chain := chain.Chain
		if length > maxLen && bc.IsChainValid(bc.chain) {
			maxLen = length
			longestChain = chain
		}
	}
	if longestChain != nil {
		bc.chain = longestChain
		return true
	} else {
		return false
	}
}

func getReqJson(url string) (resp *http.Response) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalln(err)
	}

	req.Header.Set("Accept", "application/json")
	client := &http.Client{}
	response, err := client.Do(req)

	if err != nil {
		log.Fatalln(err)
	}
	response.Body.Close()
	return response
}

var bc = BlockChain{
	chain:        nil,
	transactions: nil,
	nodes:        nil,
}

var router = mux.NewRouter().StrictSlash(true)

func main() {
	handleRequests("6000", bc)
}

func handleRequests(port string, bc BlockChain) {
	router.HandleFunc("/get-chain", GetChain).Methods("GET")
	router.HandleFunc("/mine-block", MineBlock).Methods("GET")
	router.HandleFunc("/connect-node", ConnectNode).Methods("POST")
	router.HandleFunc("/replace-chain", ReplaceChain).Methods("GET")
	log.Fatal(http.ListenAndServe(":" + port, router))
}

type ResponseGetChain struct {
	Chain  []Block `json:"chain"`
	Length int     `json:"length"`
}

func GetChain(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(ResponseGetChain{
		Chain:  bc.chain,
		Length: len(bc.chain),
	})
}

type ResponesMineBlock struct {
	Message string
	Block
}

func MineBlock(w http.ResponseWriter, r *http.Request) {
	previousBlock := bc.GetPreviousBlock()
	previousProof := previousBlock.Proof
	proof := bc.ProofOfWork(previousProof)
	previousHash := bc.hash(previousBlock)
	bc.AddTransaction(Transaction{
		sender:   "http://0.0.0.0:6001",
		receiver: "Person" + "6001",
		amount:   1,
	})
	blk := bc.CreateBlock(proof, previousHash)
	payload, _ := json.Marshal(ResponesMineBlock{
		Message: "congrats! you just mined a block",
		Block:   blk,
	})
	fmt.Println("-------------------------")
	fmt.Println(payload)
	w.WriteHeader(http.StatusOK)
}

type RequestConnectNodes struct {
	Nodes []string `json:"nodes"`
}
type ResponseConnectNodes struct {
	Message    string
	TotalNodes []string
}
func ConnectNode(w http.ResponseWriter, r *http.Request) {
	var requstConnectNodes RequestConnectNodes
	reqBody, _ := ioutil.ReadAll(r.Body)
	_ = json.Unmarshal(reqBody, &requstConnectNodes)
	fmt.Println("-------------------------")
	fmt.Println(requstConnectNodes)
	for _, node := range requstConnectNodes.Nodes {
		if len(node) < 1 {
			continue
		}
		bc.AddNode(node)
	}
	json.NewEncoder(w).Encode(ResponseConnectNodes{
		Message:    "all nodes are connected, Sandie coin now contains the follwing code",
		TotalNodes: bc.nodes,
	})
	w.WriteHeader(http.StatusOK)
}

type ResponseReplaceChain struct {
	Message string
	Chain []Block
}
func ReplaceChain(w http.ResponseWriter, r *http.Request) {
	var res ResponseReplaceChain
	if bc.ReplaceChain() {
		res = ResponseReplaceChain{
			Message:"the nodes had diffent chains so the chain was replaced by the longest chain",
			Chain : bc.chain,
		}
	} else {
		res = ResponseReplaceChain{
			Message:"all good, the chain is the largest one",
			Chain : bc.chain,
		}
	}
	json.NewEncoder(w).Encode(res)
	w.WriteHeader(http.StatusOK)
}


















