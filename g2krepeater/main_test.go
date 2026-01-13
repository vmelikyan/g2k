package main

import (
	"os"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.uber.org/zap"
)

func TestBuildAllowedReposSet(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected map[string]bool
	}{
		{
			name:     "empty string returns nil",
			envValue: "",
			expected: nil,
		},
		{
			name:     "single repo",
			envValue: "myorg/repo1",
			expected: map[string]bool{"myorg/repo1": true},
		},
		{
			name:     "multiple repos",
			envValue: "myorg/repo1,myorg/repo2,myorg/repo3",
			expected: map[string]bool{
				"myorg/repo1": true,
				"myorg/repo2": true,
				"myorg/repo3": true,
			},
		},
		{
			name:     "repos with spaces",
			envValue: "myorg/repo1 , myorg/repo2 , myorg/repo3",
			expected: map[string]bool{
				"myorg/repo1": true,
				"myorg/repo2": true,
				"myorg/repo3": true,
			},
		},
		{
			name:     "empty entries are ignored",
			envValue: "myorg/repo1,,myorg/repo2, ,myorg/repo3",
			expected: map[string]bool{
				"myorg/repo1": true,
				"myorg/repo2": true,
				"myorg/repo3": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("REPO_FILTERS", tt.envValue)
			defer os.Unsetenv("REPO_FILTERS")

			result := buildAllowedReposSet()

			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}
				return
			}

			if result == nil {
				t.Errorf("Expected %v, got nil", tt.expected)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("Expected length %d, got %d", len(tt.expected), len(result))
			}

			for repo := range tt.expected {
				if !result[repo] {
					t.Errorf("Expected repo %s to be in result", repo)
				}
			}
		})
	}
}

func TestBuildExcludedReposSet(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected map[string]bool
	}{
		{
			name:     "empty string returns nil",
			envValue: "",
			expected: nil,
		},
		{
			name:     "single repo",
			envValue: "myorg/test-repo",
			expected: map[string]bool{"myorg/test-repo": true},
		},
		{
			name:     "multiple repos",
			envValue: "myorg/test-repo,myorg/archived-repo,myorg/old-repo",
			expected: map[string]bool{
				"myorg/test-repo":     true,
				"myorg/archived-repo": true,
				"myorg/old-repo":      true,
			},
		},
		{
			name:     "repos with spaces",
			envValue: " myorg/test-repo , myorg/archived-repo ",
			expected: map[string]bool{
				"myorg/test-repo":     true,
				"myorg/archived-repo": true,
			},
		},
		{
			name:     "empty entries are ignored",
			envValue: "myorg/test-repo,,myorg/archived-repo, ,",
			expected: map[string]bool{
				"myorg/test-repo":     true,
				"myorg/archived-repo": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("REPO_EXCLUDE", tt.envValue)
			defer os.Unsetenv("REPO_EXCLUDE")

			result := buildExcludedReposSet()

			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}
				return
			}

			if result == nil {
				t.Errorf("Expected %v, got nil", tt.expected)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("Expected length %d, got %d", len(tt.expected), len(result))
			}

			for repo := range tt.expected {
				if !result[repo] {
					t.Errorf("Expected repo %s to be in result", repo)
				}
			}
		})
	}
}

func TestShouldProcessMessage(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	tests := []struct {
		name          string
		messageKey    string
		allowedRepos  map[string]bool
		excludedRepos map[string]bool
		expected      bool
		description   string
	}{
		// Test cases with no filters
		{
			name:          "no filters - all repos allowed",
			messageKey:    "myorg.myrepo.push.created",
			allowedRepos:  nil,
			excludedRepos: nil,
			expected:      true,
			description:   "When no filters are set, all repos should be processed",
		},

		// Test cases with REPO_FILTERS (inclusion filter)
		{
			name:         "inclusion filter - repo in allow list",
			messageKey:   "myorg.frontend.push.created",
			allowedRepos: map[string]bool{"myorg/frontend": true, "myorg/backend": true},
			excludedRepos: map[string]bool{"myorg/test-repo": true}, // Should be ignored
			expected:     true,
			description:  "When REPO_FILTERS is set and repo is in the list, process it",
		},
		{
			name:         "inclusion filter - repo not in allow list",
			messageKey:   "myorg.mobile.push.created",
			allowedRepos: map[string]bool{"myorg/frontend": true, "myorg/backend": true},
			excludedRepos: map[string]bool{"myorg/test-repo": true}, // Should be ignored
			expected:     false,
			description:  "When REPO_FILTERS is set and repo is not in the list, skip it",
		},

		// Test cases with REPO_EXCLUDE (exclusion filter)
		{
			name:          "exclusion filter - repo not in exclude list",
			messageKey:    "myorg.frontend.push.created",
			allowedRepos:  nil,
			excludedRepos: map[string]bool{"myorg/test-repo": true, "myorg/archived-repo": true},
			expected:      true,
			description:   "When only REPO_EXCLUDE is set and repo is not excluded, process it",
		},
		{
			name:          "exclusion filter - repo in exclude list",
			messageKey:    "myorg.test-repo.push.created",
			allowedRepos:  nil,
			excludedRepos: map[string]bool{"myorg/test-repo": true, "myorg/archived-repo": true},
			expected:      false,
			description:   "When only REPO_EXCLUDE is set and repo is excluded, skip it",
		},
		{
			name:          "exclusion filter - another excluded repo",
			messageKey:    "myorg.archived-repo.pull_request.opened",
			allowedRepos:  nil,
			excludedRepos: map[string]bool{"myorg/test-repo": true, "myorg/archived-repo": true},
			expected:      false,
			description:   "When only REPO_EXCLUDE is set and repo is excluded, skip it",
		},

		// Edge cases
		{
			name:          "invalid message key - too few parts",
			messageKey:    "myorg",
			allowedRepos:  nil,
			excludedRepos: nil,
			expected:      false,
			description:   "Message key with less than 2 parts should be rejected",
		},
		{
			name:          "invalid message key - single part",
			messageKey:    "invalid",
			allowedRepos:  map[string]bool{"myorg/repo": true},
			excludedRepos: nil,
			expected:      false,
			description:   "Invalid message key format should be rejected",
		},
		{
			name:         "inclusion filter takes precedence - repo in both lists",
			messageKey:   "myorg.frontend.push.created",
			allowedRepos: map[string]bool{"myorg/frontend": true},
			excludedRepos: map[string]bool{"myorg/frontend": true}, // Should be ignored
			expected:     true,
			description:  "REPO_FILTERS takes precedence, repo in allow list should be processed even if in exclude list",
		},
		{
			name:         "inclusion filter takes precedence - repo not in allow list but in exclude list",
			messageKey:   "myorg.backend.push.created",
			allowedRepos: map[string]bool{"myorg/frontend": true},
			excludedRepos: map[string]bool{"myorg/backend": true}, // Should be ignored
			expected:     false,
			description:  "REPO_FILTERS takes precedence, repo not in allow list should be skipped regardless of exclude list",
		},

		// Different event types
		{
			name:          "different event type",
			messageKey:    "myorg.myrepo.pull_request.opened",
			allowedRepos:  nil,
			excludedRepos: nil,
			expected:      true,
			description:   "Different event types should work the same way",
		},
		{
			name:          "complex event type",
			messageKey:    "myorg.myrepo.workflow_run.completed",
			allowedRepos:  nil,
			excludedRepos: map[string]bool{"myorg/other-repo": true},
			expected:      true,
			description:   "Complex event types should be handled correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &kafka.Message{
				Key: []byte(tt.messageKey),
			}

			result := shouldProcessMessage(msg, tt.allowedRepos, tt.excludedRepos, logger)

			if result != tt.expected {
				t.Errorf("%s: Expected %v, got %v", tt.description, tt.expected, result)
			}
		})
	}
}

// Test integration scenarios
func TestFilteringScenarios(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	scenarios := []struct {
		name        string
		description string
		repos       []string
		allowed     map[string]bool
		excluded    map[string]bool
		expected    map[string]bool
	}{
		{
			name:        "Scenario 1: No filters - process all",
			description: "When neither filter is set, all repos should be processed",
			repos:       []string{"myorg/repo1", "myorg/repo2", "myorg/repo3"},
			allowed:     nil,
			excluded:    nil,
			expected: map[string]bool{
				"myorg/repo1": true,
				"myorg/repo2": true,
				"myorg/repo3": true,
			},
		},
		{
			name:        "Scenario 2: Only REPO_FILTERS set",
			description: "Only repos in the allow list should be processed",
			repos:       []string{"myorg/repo1", "myorg/repo2", "myorg/repo3"},
			allowed:     map[string]bool{"myorg/repo1": true, "myorg/repo2": true},
			excluded:    nil,
			expected: map[string]bool{
				"myorg/repo1": true,
				"myorg/repo2": true,
				"myorg/repo3": false,
			},
		},
		{
			name:        "Scenario 3: Only REPO_EXCLUDE set",
			description: "All repos except excluded ones should be processed",
			repos:       []string{"myorg/repo1", "myorg/repo2", "myorg/repo3"},
			allowed:     nil,
			excluded:    map[string]bool{"myorg/repo2": true},
			expected: map[string]bool{
				"myorg/repo1": true,
				"myorg/repo2": false,
				"myorg/repo3": true,
			},
		},
		{
			name:        "Scenario 4: Both filters set - REPO_FILTERS wins",
			description: "When both are set, REPO_FILTERS takes precedence and REPO_EXCLUDE is ignored",
			repos:       []string{"myorg/repo1", "myorg/repo2", "myorg/repo3"},
			allowed:     map[string]bool{"myorg/repo1": true, "myorg/repo2": true},
			excluded:    map[string]bool{"myorg/repo1": true, "myorg/repo3": true},
			expected: map[string]bool{
				"myorg/repo1": true,  // In allow list (exclude ignored)
				"myorg/repo2": true,  // In allow list
				"myorg/repo3": false, // Not in allow list (exclude ignored)
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			for _, repo := range scenario.repos {
				repoSlug := repo[len("myorg/"):]
				messageKey := "myorg." + repoSlug + ".push.created"
				msg := &kafka.Message{
					Key: []byte(messageKey),
				}

				result := shouldProcessMessage(msg, scenario.allowed, scenario.excluded, logger)
				expected := scenario.expected[repo]

				if result != expected {
					t.Errorf("%s - Repo %s: Expected %v, got %v", scenario.description, repo, expected, result)
				}
			}
		})
	}
}
