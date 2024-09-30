package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type Webhook struct {
	Action string `json:"action"`
}

func main() {
	config := LoadConfig()

	consumer, err := kafka.NewConsumer(&config)
	if err != nil {
		log.Fatalf("Failed to create consumer: %s", err)
	}
	defer consumer.Close()

	err = consumer.SubscribeTopics([]string{"github.events"}, nil)
	if err != nil {
		log.Fatalf("Failed to subscribe to topics: %s", err)
	}

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Consumer started. Waiting for messages...")

	run := true

	for run {
		select {
		case sig := <-sigchan:
			log.Printf("Caught signal %v: terminating\n", sig)
			run = false
		default:
			// Poll for messages
			msg, err := consumer.ReadMessage(500 * time.Millisecond)
			if err == nil {
				if err := sendMessageAsHTTPPost(msg); err != nil {
					log.Printf("Failed to send message as HTTP POST: %s", err)
				}
				fmt.Printf("Received message from topic %s: %s\n", *msg.TopicPartition.Topic, "Message received")
			} else if err.(kafka.Error).Code() != kafka.ErrTimedOut {
				fmt.Printf("Consumer error: %v (%v)\n", err, msg)
			}
		}
	}

	log.Println("Closing consumer...")
}

func sendMessageAsHTTPPost(message *kafka.Message) error {
	// todo - support sending (repeating) the post request to multiple endpoints
	url := os.Getenv("REPLAY_ENDPOINT")
	if url == "" {
		return fmt.Errorf("REPLAY_ENDPOINT not set")
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(message.Value))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	for _, header := range message.Headers {
		key := string(header.Key)
		value := string(header.Value)
		req.Header.Set(key, value)
	}
	req.Header.Set("g2krepeater", "true")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// This check can be removed if we want to fire and forget
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("received non-2xx response: %d", resp.StatusCode)
	}

	// Log the successful response (optional)
	log.Printf("Successfully sent HTTP POST request. Response: %d", resp.StatusCode)

	return nil
}

func LoadConfig() kafka.ConfigMap {
	m := make(map[string]kafka.ConfigValue)

	envVars := os.Environ()

	prefix := "KAFKA_"

	for _, envVar := range envVars {
		parts := strings.SplitN(envVar, "=", 2)
		key := parts[0]
		value := parts[1]

		if strings.HasPrefix(key, prefix) {
			key := strings.Replace(key, prefix, "", 1)

			kConfig := strings.ReplaceAll(strings.ToLower(key), "_", ".")

			m[kConfig] = value
			log.Printf("Env %s Value %s", kConfig, value)
		}
	}

	return m
}
