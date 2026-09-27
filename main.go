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

var tgToken = os.Getenv("TG_BOT_TOKEN")
var recipients = parseRecipients(os.Getenv("TG_RECIPIENTS"))

func parseRecipients(s string) map[string]string {
	result := make(map[string]string)
	pairs := strings.Split(s, ",")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			name := strings.TrimSpace(kv[0])
			id := strings.TrimSpace(kv[1])
			if name != "" && id != "" {
				result[name] = id
			}
		}
	}
	return result
}

type tgMessage struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	toParam := r.URL.Query().Get("to")
	var targetIDs []string

	if toParam == "" {
		for _, id := range recipients {
			targetIDs = append(targetIDs, id)
		}
	} else {
		names := strings.Split(toParam, ",")
		for _, name := range names {
			name = strings.TrimSpace(name)
			if id, ok := recipients[name]; ok {
				targetIDs = append(targetIDs, id)
			} else {
				log.Printf("Unknown recipient: %s", name)
			}
		}
	}

	if len(targetIDs) == 0 {
		log.Printf("No recipients found for: %s", toParam)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"no_recipients"}`))
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "Read error", http.StatusBadRequest)
		return
	}

	dec := json.NewDecoder(bytes.NewReader(bodyBytes))
	var parts []string

	t, _ := dec.Token()
	if delim, ok := t.(json.Delim); ok && delim == '{' {
		for dec.More() {
			keyToken, _ := dec.Token()
			key := keyToken.(string)
			var value interface{}
			dec.Decode(&value)
			valStr := fmt.Sprintf("%v", value)
			key = strings.ReplaceAll(key, "\n", "")
			valStr = strings.ReplaceAll(valStr, "\n", "")
			parts = append(parts, fmt.Sprintf("%s = %v", key, valStr))
		}
	}

	text := strings.Join(parts, "\n\n")
	log.Printf("Received (to=%s): %s", toParam, text)

	for _, chatID := range targetIDs {
		if err := sendTelegram(r.Context(), chatID, text); err != nil {
			log.Printf("Telegram send error to %s: %v", chatID, err)
		} else {
			log.Printf("Telegram: sent to %s", chatID)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

func sendTelegram(ctx context.Context, chatID, text string) error {
	msg := tgMessage{
		ChatID: chatID,
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
