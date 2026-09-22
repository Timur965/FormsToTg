package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type WebhookPayload struct {
	Answers []Answer `json:"answers"`
}

type Answer struct {
	Question string      `json:"question"`
	Answer   interface{} `json:"answer"`
}


func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload WebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("Ошибка чтения JSON: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	log.Printf("✅ Получена анкета с %d ответами", len(payload.Answers))
	for i, a := range payload.Answers {
		log.Printf("  [%d] %s: %v", i+1, a.Question, a.Answer)
	}

	resp := map[string]interface{}{
		"status":   "ok",
		"received": len(payload.Answers),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
