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
	"os"
	"strconv"
	"time"
)

var BC BlockChain
var ROUTER = mux.NewRouter().StrictSlash(true)
var PORT string

func init() {
	BC = BlockChain{
		chain:        []Block{},
		transactions: []Transaction{},
		nodes:        nil,
	}
	fmt.Println("-------- initializing blockchain --------")
	BC.CreateBlock(1, [32]byte{'0'})
	fmt.Println(BC.chain)
	fmt.Println(BC.nodes)
	fmt.Println(BC.transactions)
}

func main() {
	PORT := os.Args[1]
	fmt.Println("Starting SandieCoin with address of... " + PORT)
	handleRequests(PORT, BC)
}

type BlockChain struct {
	chain        []Block
	transactions []Transaction
	nodes        []string
}
type Block struct {
	Index        int           `json:"index"`
	Timestamp    time.Time     `json:"timestamp"`
	Proof        float64       `json:"proof"`
	PreviousHash [32]byte      `json:"previous_hash"`
	Transactions []Transaction `json:"transactions"`
}
type Transaction struct {
	sender   string  `json:"sender"`
	receiver string  `json:"receiver"`
	amount   float64 `json:"amount"`
}

func handleRequests(port string, bc BlockChain) {
	ROUTER.HandleFunc("/mine_block", MineBlock).Methods("GET")
	ROUTER.HandleFunc("/get_chain", GetChain).Methods("GET")
	ROUTER.HandleFunc("/is_valid", IsValid).Methods("GET")
	ROUTER.HandleFunc("/add_transaction", AddTransaction).Methods("POST")
	ROUTER.HandleFunc("/connect_node", ConnectNodes).Methods("POST")
	ROUTER.HandleFunc("/replace_chain", ReplaceChain).Methods("GET")
	log.Fatal(http.ListenAndServe(":"+port, ROUTER))
}

type MineBlockRespones struct {
	Message string
	Block
}
type GetChainResponse struct {
	Chain  []Block `json:"chain"`
	Length int     `json:"length"`
}
type IsValidResponse struct {
	Message string
}
type ConnectNodesRequest struct {
	Nodes []string `json:"nodes"`
}
type ConnectNodesResponse struct {
	Message    string
	TotalNodes []string
}
type ReplaceChainResponse struct {
	Message string
	Chain   []Block
}

func MineBlock(w http.ResponseWriter, r *http.Request) {
	previousBlock := BC.GetPreviousBlock()
	previousProof := previousBlock.Proof
	proof := BC.ProofOfWork(previousProof)
	previousHash := BC.hash(previousBlock)
	BC.AddTransaction(Transaction{
		sender:   "http://0.0.0.0:" + PORT,
		receiver: "Person" + "6001",
		amount:   1,
	})
	blk := BC.CreateBlock(proof, previousHash)
	payload, _ := json.Marshal(MineBlockRespones{
		Message: "congrats! you just mined a block",
		Block:   blk,
	})
	json.NewEncoder(w).Encode(payload)
}
func GetChain(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(GetChainResponse{
		Chain:  BC.chain,
		Length: len(BC.chain),
	})
}
func IsValid(w http.ResponseWriter, r *http.Request) {
	var res IsValidResponse
	if BC.IsChainValid(BC.chain) {
		res = IsValidResponse{
			Message: "change is valid!",
		}
	} else {
		res = IsValidResponse{
			Message: "Houston, we have a problem. The Blockchain is not valid.",
		}
	}
	json.NewEncoder(w).Encode(res)
}
func AddTransaction(w http.ResponseWriter, r *http.Request) {
	// WIP
}
func ConnectNodes(w http.ResponseWriter, r *http.Request) {
	var requstConnectNodes ConnectNodesRequest
	reqBody, _ := ioutil.ReadAll(r.Body)
	_ = json.Unmarshal(reqBody, &requstConnectNodes)
	fmt.Println("-------------------------")
	fmt.Println(requstConnectNodes)
	for _, node := range requstConnectNodes.Nodes {
		if len(node) < 1 {
			continue
		}
		BC.AddNode(node)
	}
	json.NewEncoder(w).Encode(ConnectNodesResponse{
		Message:    "all nodes are connected, Sandie coin now contains the follwing code",
		TotalNodes: BC.nodes,
	})
	w.WriteHeader(http.StatusOK)
}
func ReplaceChain(w http.ResponseWriter, r *http.Request) {
	var res ReplaceChainResponse
	if BC.ReplaceChain() {
		res = ReplaceChainResponse{
			Message: "the nodes had diffent chains so the chain was replaced by the longest chain",
			Chain:   BC.chain,
		}
	} else {
		res = ReplaceChainResponse{
			Message: "all good, the chain is the largest one",
			Chain:   BC.chain,
		}
	}
	json.NewEncoder(w).Encode(res)
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
	fmt.Println("--------------------------")
	fmt.Println(bc.chain)
	if len(bc.chain) > 1 {
		return bc.chain[len(bc.chain)-1]
	} else {
		return bc.chain[0]
	}
}

func (bc *BlockChain) createHash(newProof float64, previousProof float64) [32]byte {
	encodedProof := strconv.Itoa(int(math.Pow(newProof, 2)) - int(math.Pow(previousProof, 2)))
	hashOperation := sha256.Sum256([]byte(encodedProof + "\n"))
	return hashOperation
}

func (bc *BlockChain) ProofOfWork(previousProof float64) float64 {
	newProof := float64(1)
	checkProof := false
	for !checkProof {
		encodedProof := strconv.Itoa(int(math.Pow(newProof, 2)) - int(math.Pow(previousProof, 2)))
		hashOperation := sha256.Sum256([]byte(encodedProof))
		fmt.Println("hash_operation : ", hashOperation)
		checkProof = true
		newProof += 1
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
			result = true
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
	var chain GetChainResponse

	for _, node := range bc.nodes {
		response := getReqJson(node + "/getChain")
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
