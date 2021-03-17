package main

import (
	"crypto/sha256"
	"fmt"
	"math"
	"strconv"
)

func main() {
	newProof := float64(1)
	previousProof := float64(2)
	encodedProof := strconv.Itoa(int(math.Pow(newProof, 2)) - int(math.Pow(previousProof, 2)))
	fmt.Println("-----------------------")

	fmt.Println(encodedProof)
	hashOperation := sha256.Sum256([]byte(encodedProof))
	fmt.Println("-----------------------")
	fmt.Println(hashOperation)
}