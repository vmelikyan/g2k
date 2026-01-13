package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.uber.org/zap"
)

type Webhook struct {
	Action string `json:"action"`
}

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	config := LoadConfig(logger)

	consumer, err := kafka.NewConsumer(&config)
	if err != nil {
		logger.Fatal("Failed to create consumer", zap.Error(err))
	}
	defer consumer.Close()

	err = consumer.SubscribeTopics([]string{"github.events"}, nil)
	if err != nil {
		logger.Fatal("Failed to subscribe to topics", zap.Error(err))
	}

	allowedRepos := buildAllowedReposSet()
	excludedRepos := buildExcludedReposSet()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	logger.Info("Consumer started. Waiting for messages...")

	run := true

	for run {
		select {
		case sig := <-sigchan:
			logger.Info("Caught signal, terminating", zap.String("signal", sig.String()))
			run = false
		default:
			msg, err := consumer.ReadMessage(500 * time.Millisecond)
			if err == nil {
				if !shouldProcessMessage(msg, allowedRepos, excludedRepos, logger) {
					continue
				}

				if err := sendMessageAsHTTPPost(msg, logger); err != nil {
					logger.Error("Failed to send message as HTTP POST", zap.Error(err))
				}

			} else if err.(kafka.Error).Code() != kafka.ErrTimedOut {
				logger.Error("Consumer error",
					zap.Error(err),
					zap.Any("message", msg),
				)
			}
		}
	}

	logger.Info("Closing consumer...")
}

func buildAllowedReposSet() map[string]bool {
	repoFilters := os.Getenv("REPO_FILTERS")
	if repoFilters == "" {
		return nil
	}

	allowed := make(map[string]bool)
	for _, r := range strings.Split(repoFilters, ",") {
		trimmed := strings.TrimSpace(r)
		if trimmed != "" {
			allowed[trimmed] = true
		}
	}
	return allowed
}

func buildExcludedReposSet() map[string]bool {
	repoExclude := os.Getenv("REPO_EXCLUDE")
	if repoExclude == "" {
		return nil
	}

	excluded := make(map[string]bool)
	for _, r := range strings.Split(repoExclude, ",") {
		trimmed := strings.TrimSpace(r)
		if trimmed != "" {
			excluded[trimmed] = true
		}
	}
	return excluded
}

func shouldProcessMessage(msg *kafka.Message, allowedRepos map[string]bool, excludedRepos map[string]bool, logger *zap.Logger) bool {
	// Extract repo name from message key
	keyParts := strings.Split(string(msg.Key), ".")
	if len(keyParts) < 2 {
		return false
	}
	orgRepo := keyParts[0] + "/" + keyParts[1]

	// If inclusion filter is set, use it (ignore exclusion filter)
	if allowedRepos != nil {
		if allowedRepos[orgRepo] {
			return true
		}
		logger.Info("Skipping message",
			zap.String("repo", orgRepo),
			zap.String("reason", "not in inclusion filter"),
		)
		return false
	}

	// If exclusion filter is set, check if repo should be excluded
	if excludedRepos != nil {
		if excludedRepos[orgRepo] {
			logger.Info("Skipping message",
				zap.String("repo", orgRepo),
				zap.String("reason", "in exclusion filter"),
			)
			return false
		}
	}

	// No filters or repo passes all filters
	return true
}

func sendMessageAsHTTPPost(message *kafka.Message, logger *zap.Logger) error {
	endpoints := strings.Split(os.Getenv("REPLAY_ENDPOINTS"), ",")
	if len(endpoints) == 0 || (len(endpoints) == 1 && endpoints[0] == "") {
		return fmt.Errorf("REPLAY_ENDPOINTS not set")
	}

	for _, url := range endpoints {
		url = strings.TrimSpace(url)
		go func(endpoint string) {
			if err := sendToEndpoint(endpoint, message, logger); err != nil {
				logger.Error("Failed to send to endpoint",
					zap.String("endpoint", endpoint),
					zap.Error(err),
				)
			}
		}(url)
	}

	return nil
}

func sendToEndpoint(endpoint string, message *kafka.Message, logger *zap.Logger) error {
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(message.Value))
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

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("received non-2xx response: %d", resp.StatusCode)
	}

	logger.Info("Message sent successfully",
		zap.String("key", string(message.Key)),
		zap.String("endpoint", endpoint),
		zap.Int("status", resp.StatusCode),
	)

	return nil
}

func LoadConfig(logger *zap.Logger) kafka.ConfigMap {
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
			logger.Info("Kafka config loaded",
				zap.String("key", kConfig),
				zap.String("value", value),
			)
		}
	}

	return m
}
