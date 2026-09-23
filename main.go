package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type WebhookPayload struct {
	Answers []Answer `json:"answers"`
}

type Answer struct {
	Question string      `json:"question"`
	Answer   interface{} `json:"answer"`
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "Read error", http.StatusBadRequest)
		return
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &raw); err != nil {
		log.Printf("Parse error: %v", err)
		http.Error(w, "Parse error", http.StatusBadRequest)
		return
	}

	// Собираем все пары ключ=значение в одну строку
	parts := make([]string, 0, len(raw))
	for key, val := range raw {
		parts = append(parts, fmt.Sprintf("%s=%v", key, val))
	}
	log.Printf("Received: %s", strings.Join(parts, "\n"))

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

func main() {
	http.HandleFunc("/", webhookHandler)

	port := "3000"

	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	log.Printf("Starting server on port %s", port)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
