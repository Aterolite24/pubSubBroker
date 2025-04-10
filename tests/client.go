package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"pubSubBroker/internal/models"
	"time"
)

func main() {
	topic := "test-topic"
	subscribeTime, _ := time.Parse(time.RFC3339, "2025-04-09T19:15:01.130+00:00")
	subscribeURL := "http://localhost:9000/subscribe"
	ackURL := "http://localhost:9000/ack"

	subscribeReq := models.SubscribeRequest{
		Topic: topic,
		Time:  subscribeTime,
	}
	subscribeJSON, err := json.Marshal(subscribeReq)
	if err != nil {
		fmt.Println("Error marshaling subscribe request:", err)
		return
	}

	resp, err := http.Post(subscribeURL, "application/json", bytes.NewBuffer(subscribeJSON))
	if err != nil {
		fmt.Println("Error sending subscribe request:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Subscription failed:", resp.Status)
		return
	}

	reader := bufio.NewReader(resp.Body)
	for {

		line, err := reader.ReadBytes('\n')
		if err != nil {
			fmt.Println("Stream ended:", err)
			break
		}

		var envelope models.Envelope
		err = json.Unmarshal(line, &envelope)

		if err != nil {
			fmt.Println("Error marshaling envelope :", err)
			continue
		}

		var payload map[string]int
		err = json.Unmarshal(envelope.Payload, &payload)
		if err != nil {
			fmt.Println("Error marshaling envelope :", err)
			continue
		}
		fmt.Println("Payload value : ", payload["value"])

		ack := models.Ack{
			MessageID: envelope.ID,
			Topic:     topic,
			Timestamp: time.Now(),
		}

		ackJSON, err := json.Marshal(ack)
		if err != nil {
			fmt.Println("Error marshaling ack:", err)
			continue
		}

		ackResp, err := http.Post(ackURL, "application/json", bytes.NewBuffer(ackJSON))
		if err != nil {
			fmt.Println("Error sending ack:", err)
			continue
		}
		defer ackResp.Body.Close()

		if ackResp.StatusCode != http.StatusOK {
			fmt.Println("Ack failed:", ackResp.Status)
		}

		fmt.Println("Ack sent successfully")
	}
}
