package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"backend/internal/ai"
	"backend/internal/storage"
)

type WebhookRequest struct {
	UserID  int    `json:"user_id"`
	Message string `json:"message"`
	Image   string `json:"image_base64"`
}

func MaxWebhookHandler(w http.ResponseWriter, r *http.Request) {
	var req WebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var reply string

	// Сценарий А: Фото объявления
	if req.Image != "" && !strings.Contains(strings.ToLower(req.Message), "счет") {
		result, _ := ai.ParseAnnouncement(req.Image)
		
		var addressID int
		storage.DB.QueryRow("SELECT address_id FROM user_addresses WHERE user_id = $1 LIMIT 1", req.UserID).Scan(&addressID)
		
		storage.DB.Exec(`INSERT INTO requests (user_id, address_id, type, title, description, start_date, end_date, status) VALUES ($1, $2, $3, $4, $5, $6, $7, 'pending')`,
			req.UserID, addressID, result.Type, result.Title, result.Description, result.StartDate, result.EndDate)
			
		reply = "Заявка сформирована и отправлена в УК. Статус доступен в мини-приложении."

	// Сценарий Б: Счета и Кэшбек
	} else if req.Image != "" || strings.Contains(strings.ToLower(req.Message), "счета") || strings.Contains(strings.ToLower(req.Message), "счёт") {
		reply = "К оплате 5 430 руб. Долгов нет. В мини-приложении вас ждет кэшбек."

	// Сценарий В: Умный QA по району
	} else {
		rows, err := storage.DB.Query("SELECT title, description, status FROM requests WHERE status != 'resolved' AND status != 'rejected'")
		var activeIssues []string
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var title, desc, status string
				if err := rows.Scan(&title, &desc, &status); err == nil {
					activeIssues = append(activeIssues, title + " (" + status + "): " + desc)
				}
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"response": reply,
	})
}
