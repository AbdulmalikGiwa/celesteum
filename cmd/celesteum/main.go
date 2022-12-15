package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/ethereum/go-ethereum/ethclient"
)

type Response struct {
	Height int `json:"height"`
}

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
	alchemyUrl := fmt.Sprintf("wss://eth-mainnet.g.alchemy.com/v2/%s", apiKey)
	client, err := ethclient.Dial(alchemyUrl)
	if err != nil {
		log.Fatal(err)
	}

	// Subscribe to new blocks
	headers := make(chan *types.Header)
	sub, err := client.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		log.Fatal(err)
	}

	// Print the block number and hash for new blocks
	for {
		select {
		case err := <-sub.Err():
			log.Fatal(err)
		case header := <-headers:
			fmt.Printf("Block number: %d\n", header.Number)
			fmt.Printf("Block hash: %s\n", header.Hash().Hex())

			// Convert the block header data to JSON
			headerJson, err := json.Marshal(header)
			if err != nil {
				log.Fatal(err)
			}
			blockData["data"] = hex.EncodeToString(headerJson)
			//blockData["data"] = "f1f20ca8007e910a3bf8b2e61da0f26bca07ef78717a6ea54165f5"
			resp, err := postToCelestia(celestiaNodeUrl, blockData)
			if err != nil {
				log.Fatal(err)
			}
			var response Response
			err = json.Unmarshal(resp, &response)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(response.Height)
			publishToQueue(response.Height)
		}
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

	// Read the response body into a buffer.
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, resp.Body); err != nil {
		return nil, err
	}
	// Return the response body as a byte slice.
	return buf.Bytes(), nil
}

func publishToQueue(heightValue int) {
	rabbitMQ := os.Getenv("RABBITMQ_URL")
	conn, err := amqp.Dial(rabbitMQ)
	if err != nil {
		log.Fatal(err)
	}
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	_, err = ch.QueueDeclare(
		"blockHeight", // queue name
		false,         // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		log.Fatal(err)
	}
	err = ch.Publish(
		"",            // exchange
		"blockHeight", // routing key
		false,         // mandatory
		false,         // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(strconv.Itoa(heightValue)),
		})
	if err != nil {
		log.Fatal(err)
	}

}
