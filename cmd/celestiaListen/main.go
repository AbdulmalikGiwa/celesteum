package main

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"io"
	"log"
	"net/http"
	"os"
)

// Response represents the response from the GET request
type Response struct {
	Data   []string `json:"data"`
	Height int      `json:"height"`
}

func main() {

	var namespaceId string
	flag.StringVar(&namespaceId, "namespaceId", "0c204d39600fddd3", "Namespace Id to post data to on Celestia blockchain")

	var blockHeight int
	flag.IntVar(&blockHeight, "blockHeight", 0, "Block height")

	flag.Parse()
	if namespaceId == "" {
		fmt.Println("Error: namespaceId flag is required")
		return
	}
	if blockHeight == 0 {
		fmt.Println("Error: blockHeight flag is required")
		return
	}

	// Load the environment variables from the .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get the value of the CELESTIA_NODE_URL
	celestiaNodeUrl := os.Getenv("CELESTIA_NODE_URL")

	celestiaApi := fmt.Sprintf("%s/namespaced_data/%s/height/%d}", celestiaNodeUrl, namespaceId, blockHeight)
	// Send a GET request to the specified URL
	resp, err := http.Get(celestiaApi)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Unmarshal the response body into the Response struct
	var response Response
	if err := json.Unmarshal(buf.Bytes(), &response); err != nil {
		fmt.Println(err)
		return
	}
	var hexStr string
	// Iterate over the items in the data array
	for _, item := range response.Data {
		// Decode the item from base64 to bytes
		// Todo: confirm if multiple items can be in data array
		responseBytes, err := base64.StdEncoding.DecodeString(item)
		if err != nil {
			fmt.Println(err)
			continue
		}

		// Encode the decoded bytes to hex
		hexStr = hex.EncodeToString(responseBytes)

	}
	blockJsonData, err := json.Marshal(hexStr)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(blockJsonData))
}
