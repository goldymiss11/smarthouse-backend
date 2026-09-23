package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func SendReplyMessage(userID interface{}, message string) error {
	token := os.Getenv("MAX_TOKEN")
	if token == "" {
		token = os.Getenv("MAX_BOT_TOKEN")
	}
	if token == "" {
		log.Println("MAX_TOKEN is empty, skipping reply to user:", userID)
		return nil
	}

	apiURL := "https://platform-api2.max.ru/messages"

	payload := map[string]interface{}{
		"user_id": userID,
		"chat_id": userID,
		"users":   []interface{}{userID},
		"message": message,
		"text":    message,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("ошибка сериализации payload: %v", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("ошибка создания запроса: %v", err)
	}

	authHeader := token
	if !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		authHeader = "Bearer " + authHeader
	}
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")

	log.Printf("Sending MAX reply to %s, user_id=%v, text=%q", apiURL, userID, message)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Network error sending MAX reply to %v: %v", userID, err)
		return fmt.Errorf("ошибка отправки сообщения в MAX: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("MAX API reply response status: %d, body: %s", resp.StatusCode, string(respBody))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("ошибка API МАХ, статус: %d, ответ: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

func SendPushNotification(vkIDs []string, message string, requestID string) error {
	token := os.Getenv("MAX_TOKEN")
	if token == "" {
		token = os.Getenv("MAX_BOT_TOKEN")
	}
	if token == "" {
		log.Println("MAX_TOKEN is empty, skipping push notification for request:", requestID)
		return nil
	}

	deepLink := fmt.Sprintf("https://mini-app.ru/?screen=ukModeration&requestId=%s", requestID)

	payload := map[string]interface{}{
		"users":   vkIDs,
		"message": message,
		"text":    message,
		"button":  map[string]string{"text": "Открыть", "url": deepLink},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "https://platform-api2.max.ru/messages", bytes.NewBuffer(body))
	authHeader := token
	if !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		authHeader = "Bearer " + authHeader
	}
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка отправки push: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("ошибка API бота, статус: %d", resp.StatusCode)
	}
	return nil
}