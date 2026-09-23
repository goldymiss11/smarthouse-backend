package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"backend/internal/ai"
	"backend/internal/models"
	"backend/internal/storage"
)

func GetFeedHandler(w http.ResponseWriter, r *http.Request) {
	userID := 1 // Заглушка, в проде берем из JWT
	
	query := `
		SELECT f.id, f.address_id, f.title, f.body, f.category, f.created_at 
		FROM feed_items f
		JOIN user_addresses ua ON f.address_id = ua.address_id
		WHERE ua.user_id = $1
		ORDER BY f.created_at DESC LIMIT 20
	`
	rows, err := storage.DB.Query(query, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("DB Error in resident feed: %v", err)
		return
	}
	defer rows.Close()

	var feed []models.FeedItem
	for rows.Next() {
		var item models.FeedItem
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.AddressID, &item.Title, &item.Body, &item.Category, &createdAt); err == nil {
			item.CreatedAt = createdAt.Format(time.RFC3339)
			feed = append(feed, item)
		}
	}
	
	if feed == nil {
		feed = []models.FeedItem{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(feed)
}

func GetMetersHandler(w http.ResponseWriter, r *http.Request) {
	userID := 1
	var water, electricity float64
	
	err := storage.DB.QueryRow("SELECT water, electricity FROM meters WHERE user_id = $1 ORDER BY submitted_at DESC LIMIT 1", userID).Scan(&water, &electricity)
	if err != nil {
		// Если показаний нет, возвращаем нули
		water = 0
		electricity = 0
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"water": water, "electricity": electricity})
}

func PostMetersHandler(w http.ResponseWriter, r *http.Request) {
	userID := 1
	var req struct {
		Water       float64 `json:"water"`
		Electricity float64 `json:"electricity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := storage.DB.Exec("INSERT INTO meters (user_id, water, electricity) VALUES ($1, $2, $3)", userID, req.Water, req.Electricity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("DB Error in post meters: %v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func AnalyzePhotoHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ImageBase64 string `json:"image_base64"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := ai.ParseAnnouncement(req.ImageBase64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func CreateRequestHandler(w http.ResponseWriter, r *http.Request) {
	userID := 1
	var req models.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Находим адрес пользователя
	var addressID int
	err := storage.DB.QueryRow("SELECT address_id FROM user_addresses WHERE user_id = $1 LIMIT 1", userID).Scan(&addressID)
	if err != nil {
		http.Error(w, "User address not found: "+err.Error(), http.StatusBadRequest)
		log.Printf("DB Error in user address lookup: %v", err)
		return
	}

	_, err = storage.DB.Exec(`
		INSERT INTO requests (user_id, address_id, type, title, description, start_date, end_date, status) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'pending')`,
		userID, addressID, req.Type, req.Title, req.Description, req.StartDate, req.EndDate)
		
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("DB Error in create request: %v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "created"})
}