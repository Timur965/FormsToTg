package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	tgToken = os.Getenv("TG_BOT_TOKEN")
	tgChat  = os.Getenv("TG_CHAT_ID")
)

type tgMessage struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
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

	params, _ := raw["params"].(map[string]interface{})
	if params == nil || len(params) == 0 {
		log.Printf("Empty params, skipping")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","skipped":true}`))
		return
	}

	var lines []string
	for key, val := range params {
		if arr, ok := val.([]interface{}); ok {
			// Массив — перебираем по порядку, выводим как есть
			for _, item := range arr {
				itemJSON, _ := json.Marshal(item)
				lines = append(lines, string(itemJSON))
			}
		} else {
			valJSON, _ := json.Marshal(val)
			lines = append(lines, fmt.Sprintf("%s: %s", key, string(valJSON)))
		}
	}

	text := strings.Join(lines, "\n")
	log.Printf("Received: %s", text)

	if tgToken != "" && tgChat != "" && text != "" {
		if err := sendTelegram(r.Context(), text); err != nil {
			log.Printf("Telegram send error: %v", err)
		} else {
			log.Printf("Telegram: sent")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

func sendTelegram(ctx context.Context, text string) error {
	msg := tgMessage{
		ChatID: tgChat,
		Text:   text,
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost,
		fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", tgToken),
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var tgResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&tgResp)
		return fmt.Errorf("telegram api: %v", tgResp)
	}

	return nil
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
