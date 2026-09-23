package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
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
		log.Printf("Ошибка чтения тела: %v", err)
		http.Error(w, "Read error", http.StatusBadRequest)
		return
	}

	log.Printf("📨 Длина тела: %d байт", len(bodyBytes))

	var raw map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &raw); err != nil {
		log.Printf("Ошибка парсинга: %v. Сырой текст: %s", err, string(bodyBytes))
		http.Error(w, "Parse error", http.StatusBadRequest)
		return
	}

	for key, val := range raw {
		log.Printf("🔑 Ключ '%s': %v", key, val)
	}

	if params, ok := raw["params"]; ok {
		paramsJSON, _ := json.MarshalIndent(params, "", "  ")
		log.Printf("📦 Содержимое params:\n%s", string(paramsJSON))
	}

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
