package handlers

import (
	"encoding/json"
	"net/http"

	"backend/internal/bot"
	"backend/internal/storage"
)

func UpdateRequestStatusHandler(w http.ResponseWriter, r *http.Request) {
	requestID := r.PathValue("id")
	
	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := storage.DB.Exec("UPDATE requests SET status = $1 WHERE id = $2", req.Status, requestID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Находим VK ID пользователей для отправки пуша
	query := `
		SELECT u.vk_id 
		FROM users u 
		JOIN user_addresses ua ON u.id = ua.user_id 
		WHERE ua.address_id = (SELECT address_id FROM requests WHERE id = $1 LIMIT 1)
	`
	rows, err := storage.DB.Query(query, requestID)
	if err == nil {
		defer rows.Close()
		var vkIDs []string
		for rows.Next() {
			var vkID string
			if err := rows.Scan(&vkID); err == nil {
				vkIDs = append(vkIDs, vkID)
			}
		}
		if len(vkIDs) > 0 {
			msg := "Статус заявки изменен на: " + req.Status
			bot.SendPushNotification(vkIDs, msg, requestID)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": req.Status})
}
