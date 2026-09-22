package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

func SendPushNotification(vkIDs []string, message string, requestID string) error {
	token := os.Getenv("MAX_TOKEN")
	if token == "" {
		log.Println("MAX_TOKEN is empty, skipping push notification for request:", requestID)
		return nil
	}

	deepLink := fmt.Sprintf("https://mini-app.ru/?screen=ukModeration&requestId=%s", requestID)

	payload := map[string]interface{}{
		"users":   vkIDs,
		"message": message,
		"button":  map[string]string{"text": "Открыть", "url": deepLink},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "https://api.max.ru/v1/messages/send", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка отправки push: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("ошибка API бота, статус: %d", resp.StatusCode)
	}
	return nil
}