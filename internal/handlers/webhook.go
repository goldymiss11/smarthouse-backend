package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

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

func MaxWebhookHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewBuffer(body))
	log.Printf("МАХ WEBHOOK RAW: %s", string(body))

	var req WebhookRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.UserID == 0 && req.ChatID != 0 {
		req.UserID = req.ChatID
	}
	if req.UserID == 0 {
		var rawMap map[string]interface{}
		if err := json.Unmarshal(body, &rawMap); err == nil {
			if uid, ok := rawMap["user_id"]; ok {
				req.UserID = toInt(uid)
			} else if cid, ok := rawMap["chat_id"]; ok {
				req.UserID = toInt(cid)
			} else if uObj, ok := rawMap["user"].(map[string]interface{}); ok {
				if id, ok := uObj["id"]; ok {
					req.UserID = toInt(id)
				}
			} else if fObj, ok := rawMap["from"].(map[string]interface{}); ok {
				if id, ok := fObj["id"]; ok {
					req.UserID = toInt(id)
				}
			}
		}
	}
	if req.Message == "" && req.Text != "" {
		req.Message = req.Text
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

	// Отправка исходящего HTTP POST-запроса на API МАХ (https://platform-api2.max.ru/messages)
	if err := bot.SendReplyMessage(req.UserID, reply); err != nil {
		log.Printf("Error sending message to MAX API: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"response": reply,
	})
}
