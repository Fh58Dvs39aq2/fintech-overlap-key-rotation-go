package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

type service struct {
	client    apiClient
	directory string
}

func (s service) rotate(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	var input RotationRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, req.Body, 4096)).Decode(&input); err != nil {
		http.Error(w, "invalid request", 400)
		return
	}
	if err := decide(input); err != nil {
		http.Error(w, err.Error(), 422)
		return
	}
	eventHash := sha256.Sum256([]byte(input.Event.ID))
	identifier := hex.EncodeToString(eventHash[:])
	// One credential and base URL authorize both account changes and deployment-log inspection.
	rotated, err := s.client.call(req.Context(), http.MethodPost, "/v1/account/keys/rotate/"+url.PathEscape(input.TemporaryKeyID), map[string]any{
		"grace_hours": input.GraceHours, "idempotency_key": "rotate-" + identifier,
	})
	if err != nil {
		http.Error(w, err.Error(), 502)
		return
	}
	_, nextSecret, err := credential(rotated)
	if err != nil {
		http.Error(w, err.Error(), 502)
		return
	}
	if err := writeCredential(filepath.Join(s.directory, identifier+".rotated.key"), nextSecret); err != nil {
		http.Error(w, "credential storage failed", 500)
		return
	}
	logs, err := s.client.call(req.Context(), http.MethodGet, "/v1/logs/search", nil)
	if err != nil {
		http.Error(w, err.Error(), 502)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		EventID        string          `json:"event_id"`
		TemporaryKeyID string          `json:"temporary_key_id"`
		OverlapEnds    string          `json:"overlap_ends"`
		DeploymentLogs json.RawMessage `json:"deployment_logs"`
	}{input.Event.ID, input.TemporaryKeyID, time.Now().UTC().Add(time.Duration(input.GraceHours) * time.Hour).Format(time.RFC3339), logs})
}

func main() {
	key := os.Getenv("INFRAI_API_KEY")
	directory := os.Getenv("KEY_DIRECTORY")
	if key == "" || directory == "" {
		log.Fatal("INFRAI_API_KEY and KEY_DIRECTORY are required")
	}
	if err := os.MkdirAll(directory, 0700); err != nil {
		log.Fatal(err)
	}
	address := os.Getenv("LISTEN_ADDR")
	if address == "" {
		address = "127.0.0.1:8080"
	}
	s := service{apiClient{key, &http.Client{Timeout: 20 * time.Second}}, directory}
	mux := http.NewServeMux()
	mux.HandleFunc("/rotation", s.rotate)
	log.Fatal(http.ListenAndServe(address, mux))
}
