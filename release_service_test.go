package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDecideRelease(t *testing.T) {
	base := ReleaseOperation{Service: "ledger-api", Version: "2026.09.05", Build: BuildEvent{Commit: "abc123", Pipeline: "main", Passed: true}}
	tests := []struct {
		name                  string
		enabled, passed, want bool
		reason                string
	}{
		{"approved rollout", true, true, true, "approved"},
		{"flag gate", false, true, false, "feature_disabled"},
		{"failed build", true, false, false, "build_failed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op := base
			op.Build.Passed = tt.passed
			got := decide(tt.enabled, op)
			if got.Allowed != tt.want || got.Reason != tt.reason {
				t.Fatalf("got %+v", got)
			}
		})
	}
}

func TestFlagNotFoundIsDisabled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"ok":false,"error":{"code":"FLAG_NOT_FOUND"}}`))
	}))
	defer server.Close()

	on, err := (flagClient{base: server.URL, key: "test", client: server.Client()}).enabled("ledger-rollout")
	if err != nil {
		t.Fatalf("enabled returned error: %v", err)
	}
	if on {
		t.Fatal("missing flag must be treated as disabled")
	}
}
