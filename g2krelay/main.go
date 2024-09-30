package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// WebhookPayload represents the structure of the GitHub webhook payload kinda hacky can be improved
type WebhookPayload struct {
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
}

func main() {
	http.HandleFunc("/api/webhooks/github", webhookHandler)

	serverPort := os.Getenv("PORT")
	if serverPort == "" {
		serverPort = "5050"
	}

	log.Printf("Starting g2krelay server on %s...", serverPort)

	if err := http.ListenAndServe(":"+serverPort, nil); err != nil {
		log.Fatalf("Server failed: %s", err)
	}
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	topic := "github.events"

	body, err := io.ReadAll(io.Reader(r.Body))
	if err != nil {
		http.Error(w, "Cannot read request body", http.StatusBadRequest)
		return
	}

	// Should I force webhook secret validation? -> probably
	if !validateSignature(r, body) {
		http.Error(w, "Invalid signature, check webhook secret", http.StatusUnauthorized)
		return
	}

	eventType := r.Header.Get("X-GitHub-Event")
	if eventType == "" {
		http.Error(w, "Event type not found", http.StatusBadRequest)
	}

	// Parsing the payload here to be used for setting message key
	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	repoFullName := strings.ReplaceAll(payload.Repository.FullName, "/", ".")
	if repoFullName == "" {
		http.Error(w, "Repository full name not found", http.StatusBadRequest)
		return
	}
	messageKey := repoFullName + "." + eventType

	headers := getKafkaHeaders(r)
	if err := produceToKafka(topic, body, messageKey, headers); err != nil {
		log.Printf("Failed to produce message to Kafka: %s", err)
		http.Error(w, "Failed to process webhook", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Webhook received and processed by g2krelay")
}

func getKafkaHeaders(r *http.Request) []kafka.Header {
	var headers []kafka.Header
	for name, values := range r.Header {
		for _, value := range values {
			header := kafka.Header{
				Key:   name,
				Value: []byte(value),
			}
			headers = append(headers, header)
		}
	}
	return headers
}

func validateSignature(r *http.Request, body []byte) bool {
	// Maybe some of go github libs support this out of the box?
	// Could update this if there is a need of a go github lib for other things.
	signature := r.Header.Get("X-Hub-Signature-256")
	if signature == "" {
		log.Println("No signature provided")
		return false
	}

	secret := os.Getenv("WEBHOOK_SECRET")
	if secret == "" {
		log.Println("WEBHOOK_SECRET not set")
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedSignature := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
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

func produceToKafka(topic string, message []byte, action string, headers []kafka.Header) error {
	config := LoadConfig()

	p, err := kafka.NewProducer(&config)
	if err != nil {
		return fmt.Errorf("failed to create producer: %w", err)
	}
	defer p.Close()

	if err := ensureTopicExists(topic, config); err != nil {
		return err
	}

	deliveryChan := make(chan kafka.Event, 1)

	err = p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            []byte(action),
		Value:          message,
		Headers:        headers,
	}, deliveryChan)
	if err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	// Wait for delivery report
	e := <-deliveryChan
	m := e.(*kafka.Message)
	if m.TopicPartition.Error != nil {
		return fmt.Errorf("delivery failed: %w", m.TopicPartition.Error)
	}

	log.Printf("Message delivered to topic %s [%d] at offset %v",
		*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
	close(deliveryChan)
	return nil
}

func ensureTopicExists(topic string, config kafka.ConfigMap) error {
	adminClient, err := kafka.NewAdminClient(&config)
	if err != nil {
		return fmt.Errorf("failed to create admin client: %w", err)
	}
	defer adminClient.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	results, err := adminClient.CreateTopics(
		ctx,
		[]kafka.TopicSpecification{{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		}},
	)
	if err != nil {
		return fmt.Errorf("failed to create topic: %w", err)
	}

	for _, result := range results {
		if result.Error.Code() != kafka.ErrNoError && result.Error.Code() != kafka.ErrTopicAlreadyExists {
			return fmt.Errorf("failed to create topic %s: %v", result.Topic, result.Error)
		}
	}
	return nil
}
