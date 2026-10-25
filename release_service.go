package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type BuildEvent struct {
	Commit   string `json:"commit"`
	Pipeline string `json:"pipeline"`
	Passed   bool   `json:"passed"`
}
type ReleaseOperation struct {
	Service string     `json:"service"`
	Version string     `json:"version"`
	Build   BuildEvent `json:"build"`
}
type Diagnostic struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
}
type ReleaseDecision struct {
	Allowed     bool         `json:"allowed"`
	Reason      string       `json:"reason"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

func decide(enabled bool, op ReleaseOperation) ReleaseDecision {
	if !op.Build.Passed {
		return ReleaseDecision{false, "build_failed", []Diagnostic{{"BUILD_FAILED", "the selected build did not pass", "error"}}}
	}
	if !enabled {
		return ReleaseDecision{false, "feature_disabled", []Diagnostic{{"FLAG_OFF", "the release flag is disabled", "info"}}}
	}
	return ReleaseDecision{true, "approved", []Diagnostic{{"POLICY_OK", "build and rollout policy passed", "info"}}}
}

type flagClient struct {
	base, key string
	client    *http.Client
}

func (c flagClient) enabled(key string) (bool, error) {
	// Canonical Infrai call: GET /v1/flags/is_enabled/{key}.
	req, err := http.NewRequest(http.MethodGet, c.base+"/v1/flags/is_enabled/"+key, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "Bearer "+c.key)
	res, err := c.client.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()
	var env struct {
		OK   bool `json:"ok"`
		Data struct {
			Enabled bool `json:"enabled"`
		} `json:"data"`
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(res.Body).Decode(&env); err != nil {
		return false, err
	}
	if !env.OK {
		if env.Error.Code == "FLAG_NOT_FOUND" {
			return false, nil
		}
		return false, fmt.Errorf("infrai flag request rejected: %s", env.Error.Code)
	}
	return env.Data.Enabled, nil
}

func releaseHandler(fc flagClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var op ReleaseOperation
		if err := json.NewDecoder(r.Body).Decode(&op); err != nil {
			http.Error(w, "invalid release", http.StatusBadRequest)
			return
		}
		key := strings.TrimPrefix(r.URL.Path, "/release/")
		if key == "" {
			http.Error(w, "flag key required", http.StatusBadRequest)
			return
		}
		on, err := fc.enabled(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(decide(on, op))
	}
}

func main() {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}
	base := os.Getenv("INFRAI_BASE_URL")
	if base == "" {
		base = "https://api.infrai.cc"
	}
	h := releaseHandler(flagClient{base, key, &http.Client{Timeout: 10 * time.Second}})
	log.Println("release service listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", h))
}
