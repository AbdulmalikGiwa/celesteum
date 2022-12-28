package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

// Response represents the response from the GET request
type Response struct {
	Data []string `json:"data"`
}

func main() {

	var namespaceId string
	flag.StringVar(&namespaceId, "namespaceId", "0c204d39600fddd3", "Namespace Id to post data to on Celestia blockchain")

	flag.Parse()
	if namespaceId == "" {
		fmt.Println("Error: namespaceId flag is required")
		return
	}

	// Load the environment variables from the .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	rabbitMQ := os.Getenv("RABBITMQ_URL")
	fmt.Println(rabbitMQ)

	conn, err := amqp.Dial(rabbitMQ)
	if err != nil {
		log.Fatal(err)
	}
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	msgs, err := ch.Consume(
		"blockHeight", // queue name
		"",            // consumer tag
		true,          // auto-ack
		false,         // exclusive
		false,         // no-local
		false,         // no-wait
		nil,           // args
	)
	if err != nil {
		log.Fatal(err)
	}

	// Listen for messages on the channel and process them as they arrive.
	for {
		fmt.Println("ENTER")
		for msg := range msgs {
			// Get the value of the CELESTIA_NODE_URL
			celestiaNodeUrl := os.Getenv("CELESTIA_NODE_URL")
			blockHeight, err := strconv.Atoi(string(msg.Body))
			if err != nil {
				log.Fatal(err)
			}
			celestiaApi := fmt.Sprintf("%s/namespaced_data/%s/height/%d", celestiaNodeUrl, namespaceId, blockHeight)
			// Send a GET request to the CELESTIA
			fmt.Println(celestiaApi)
			func() {
				resp, err := http.Get(celestiaApi)
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Println(resp.StatusCode)
				defer resp.Body.Close()
				buf := new(bytes.Buffer)
				_, err = io.Copy(buf, resp.Body)
				if err != nil {
					log.Fatal(err)
				}
				// Unmarshal the response body into the Response struct
				var response Response
				if err := json.Unmarshal(buf.Bytes(), &response); err != nil {
					log.Fatal(err)
				}
				for _, item := range response.Data {
					// Decode the item from base64 to bytes
					// Todo: confirm if multiple items can be in data array
					decodeBlock(item)

				}
			}()

		}
	}

}
func decodeBlock(block string) {
	headerJson, err := hex.DecodeString(block)
	if err != nil {
		log.Fatal(err)
	}
	var data interface{}
	err = json.Unmarshal(headerJson, &data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", data)
}
