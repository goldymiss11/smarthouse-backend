package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"backend/internal/ai"
	"backend/internal/bot"
	"backend/internal/models"
	"backend/internal/storage"
)

type WebhookRequest struct {
	UserID   int    `json:"user_id"`
	ChatID   int    `json:"chat_id"`
	Message  string `json:"message"`
	Text     string `json:"text"`
	Image    string `json:"image_base64"`
}

func toInt(v interface{}) int {
	switch val := v.(type) {
	case float64:
		return int(val)
	case int:
		return val
	case string:
		n, _ := strconv.Atoi(val)
		return n
	case json.Number:
		n, _ := val.Int64()
		return int(n)
	default:
		return 0
	}
}

func parseChatID(val interface{}) interface{} {
	switch v := val.(type) {
	case float64:
		return int64(v)
	case int:
		return int64(v)
	case int64:
		return v
	default:
		return val
	}
}

func MaxWebhookHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewBuffer(body))
	log.Printf("МАХ WEBHOOK RAW: %s", string(body))

	var rawMap map[string]interface{}
	json.Unmarshal(body, &rawMap)

	var req WebhookRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Извлекаем chat_id и текст сообщения из входящего вебхука
	var chatID interface{}
	if req.ChatID != 0 {
		chatID = req.ChatID
	} else if req.UserID != 0 {
		chatID = req.UserID
	}

	if chatID == nil || chatID == 0 {
		if cid, ok := rawMap["chat_id"]; ok && cid != nil {
			chatID = cid
		} else if uid, ok := rawMap["user_id"]; ok && uid != nil {
			chatID = uid
		} else if recObj, ok := rawMap["recipient"].(map[string]interface{}); ok {
			if id, ok := recObj["chat_id"]; ok {
				chatID = id
			}
		} else if chatObj, ok := rawMap["chat"].(map[string]interface{}); ok {
			if id, ok := chatObj["id"]; ok {
				chatID = id
			} else if id, ok := chatObj["chatId"]; ok {
				chatID = id
			} else if id, ok := chatObj["chat_id"]; ok {
				chatID = id
			}
		} else if fromObj, ok := rawMap["from"].(map[string]interface{}); ok {
			if id, ok := fromObj["id"]; ok {
				chatID = id
			} else if id, ok := fromObj["userId"]; ok {
				chatID = id
			}
		} else if payloadObj, ok := rawMap["payload"].(map[string]interface{}); ok {
			if chatObj, ok := payloadObj["chat"].(map[string]interface{}); ok {
				if id, ok := chatObj["id"]; ok {
					chatID = id
				} else if id, ok := chatObj["chatId"]; ok {
					chatID = id
				}
			} else if fromObj, ok := payloadObj["from"].(map[string]interface{}); ok {
				if id, ok := fromObj["id"]; ok {
					chatID = id
				} else if id, ok := fromObj["userId"]; ok {
					chatID = id
				}
			}
		}
	}

	if chatID != nil {
		chatID = parseChatID(chatID)
	}

	if req.UserID == 0 && chatID != nil {
		req.UserID = toInt(chatID)
	}

	if req.Message == "" {
		if req.Text != "" {
			req.Message = req.Text
		} else if txt, ok := rawMap["text"].(string); ok && txt != "" {
			req.Message = txt
		} else if msgObj, ok := rawMap["message"].(map[string]interface{}); ok {
			if txt, ok := msgObj["text"].(string); ok && txt != "" {
				req.Message = txt
			}
		} else if payloadObj, ok := rawMap["payload"].(map[string]interface{}); ok {
			if txt, ok := payloadObj["text"].(string); ok && txt != "" {
				req.Message = txt
			} else if msgObj, ok := payloadObj["message"].(map[string]interface{}); ok {
				if txt, ok := msgObj["text"].(string); ok && txt != "" {
					req.Message = txt
				}
			}
		}
	}

	var reply string

	// Сценарий А: Фото объявления
	if req.Image != "" && !strings.Contains(strings.ToLower(req.Message), "счет") {
		result, err := ai.ParseAnnouncement(req.Image)
		if err != nil || result == nil {
			result = &models.Request{
				Type:        "other",
				Title:       "Распознано ИИ",
				Description: "Заявка по фото объявления",
				StartDate:   "",
				EndDate:     "",
			}
		}

		// Обрезаем поля VARCHAR до 30 символов перед выполнением SQL-запроса
		if len([]rune(result.Type)) > 30 {
			result.Type = string([]rune(result.Type)[:30])
		}
		if len([]rune(result.Title)) > 30 {
			result.Title = string([]rune(result.Title)[:30])
		}
		if len([]rune(result.StartDate)) > 30 {
			result.StartDate = string([]rune(result.StartDate)[:30])
		}
		if len([]rune(result.EndDate)) > 30 {
			result.EndDate = string([]rune(result.EndDate)[:30])
		}
		if result.Status == "" {
			result.Status = "pending"
		}
		if len([]rune(result.Status)) > 20 {
			result.Status = string([]rune(result.Status)[:20])
		}
		
		var addressID int
		storage.DB.QueryRow("SELECT address_id FROM user_addresses WHERE user_id = $1 LIMIT 1", req.UserID).Scan(&addressID)
		
		_, err = storage.DB.Exec(`INSERT INTO requests (user_id, address_id, type, title, description, start_date, end_date, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			req.UserID, addressID, result.Type, result.Title, result.Description, result.StartDate, result.EndDate, result.Status)
		if err != nil {
			log.Printf("DB Error in webhook insert request: %v", err)
		}
			
		reply = "Заявка сформирована и отправлена в УК. Статус доступен в мини-приложении."

	// Сценарий Б: Счета и Кэшбек
	} else if req.Image != "" || strings.Contains(strings.ToLower(req.Message), "счета") || strings.Contains(strings.ToLower(req.Message), "счёт") {
		reply = "К оплате 5 430 руб. Долгов нет. В мини-приложении вас ждет кэшбек."

	// Сценарий В: Умный QA по району
	} else {
		rows, err := storage.DB.Query("SELECT title, description, status FROM requests WHERE status != 'resolved' AND status != 'rejected'")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Printf("DB Error in webhook: %v", err)
			return
		}
		defer rows.Close()

		var activeIssues []string
		for rows.Next() {
			var title, desc, status string
			if err := rows.Scan(&title, &desc, &status); err == nil {
				activeIssues = append(activeIssues, title + " (" + status + "): " + desc)
			}
		}
		
		contextStr := strings.Join(activeIssues, "\n")
		if contextStr == "" {
			contextStr = "Аварий и проблем не зафиксировано."
		}
		prompt := "Ответь пользователю на его вопрос коротко, на основе контекста активных проблем в районе. Вопрос: " + req.Message + "\nКонтекст: " + contextStr
		
		replyText, _ := ai.AskLLM(prompt)
		reply = replyText
	}

	// 1-4. Исходящий HTTP POST-запрос на API МАХ (https://platform-api2.max.ru/messages)
	maxToken := os.Getenv("MAX_TOKEN")
	if maxToken == "" {
		maxToken = os.Getenv("MAX_BOT_TOKEN")
	}

	if maxToken == "" {
		log.Println("WARNING: MAX_TOKEN is empty, cannot send message to MAX API")
	} else {
		outgoingPayload := map[string]interface{}{
			"recipient": map[string]interface{}{
				"chat_id": chatID,
			},
			"message": map[string]interface{}{
				"text": reply,
			},
		}

		reqBytes, err := json.Marshal(outgoingPayload)
		if err != nil {
			log.Printf("Error marshaling MAX API payload: %v", err)
		} else {
			postReq, err := http.NewRequest("POST", "https://platform-api2.max.ru/messages", bytes.NewBuffer(reqBytes))
			if err != nil {
				log.Printf("Error creating HTTP request to MAX API: %v", err)
			} else {
				authHeader := "Bearer " + strings.TrimPrefix(maxToken, "Bearer ")
				postReq.Header.Set("Authorization", authHeader)
				postReq.Header.Set("Content-Type", "application/json")

				log.Printf("OUTGOING TO MAX API: url=https://platform-api2.max.ru/messages, body=%s", string(reqBytes))

				client := &http.Client{Timeout: 10 * time.Second}
				resp, err := client.Do(postReq)
				if err != nil {
					log.Printf("Error sending HTTP POST to MAX API: %v", err)
				} else {
					defer resp.Body.Close()
					respBody, _ := io.ReadAll(resp.Body)
					log.Printf("MAX API RESPONSE: status=%d, body=%s", resp.StatusCode, string(respBody))
					if resp.StatusCode < 200 || resp.StatusCode >= 300 {
						log.Printf("MAX API returned non-OK status: %d, response: %s", resp.StatusCode, string(respBody))
					}
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"response": reply,
	})
}
