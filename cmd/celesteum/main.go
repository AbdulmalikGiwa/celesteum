package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {

	var namespaceId string
	flag.StringVar(&namespaceId, "namespaceId", "0c204d39600fddd3", "Namespace Id to post data to on Celestia blockchain")
	blockData := map[string]interface{}{
		"namespace_id": namespaceId,
		"gas_limit":    70000,
	}
	// Parse the command line flags.
	flag.Parse()

	// Load the environment variables from the .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// Get the value of the API_KEY environment variable
	apiKey := os.Getenv("API_KEY")
	// Get the value of the CELESTIA_NODE_URL to post data to
	celestiaNodeUrl := os.Getenv("CELESTIA_NODE_URL")

	// Create an ethclient.Client
	alchemyUrl := fmt.Sprintf("https://eth-goerli.g.alchemy.com/v2/%s", apiKey)
	client, err := ethclient.Dial(alchemyUrl)
	if err != nil {
		log.Fatal(err)
	}

	// Get the latest block
	block, err := client.BlockByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	// Print the block number and hash
	fmt.Printf("Block number: %d\n", block.Number())
	fmt.Printf("Block hash: %s\n", block.Hash().Hex())

	// Iterate over the transactions in the block and post hashes to celestia Node
	for _, tx := range block.Transactions() {
		blockData["data"] = tx.Hash().Hex()
		respBody, err := postToCelestia(celestiaNodeUrl, blockData)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(respBody))
	}
}

func postToCelestia(url string, data map[string]interface{}) ([]byte, error) {

	// Add payfordata endpoint
	url = fmt.Sprintf("%s/submit_pfd", url)

	// Marshal the data into JSON.
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// Create the request.
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	// Send the request.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read and return the response body.
	return ioutil.ReadAll(resp.Body)
}
