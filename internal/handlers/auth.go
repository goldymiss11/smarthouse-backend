package handlers

import (
	"encoding/json"
	"net/http"
	"backend/internal/storage"
	"database/sql"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req struct{ VkID string `json:"user_id"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var userID int
	var role string
	err := storage.DB.QueryRow("SELECT id, role FROM users WHERE vk_id = $1", req.VkID).Scan(&userID, &role)
	
	if err == sql.ErrNoRows {
		err = storage.DB.QueryRow("INSERT INTO users (vk_id, role) VALUES ($1, 'resident') RETURNING id, role", req.VkID).Scan(&userID, &role)
		if err != nil {
			http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"token": req.VkID, "role": role, "user_id": userID})
}

func LinkAddressHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AddressID     int    `json:"address_id"`
		Apartment     string `json:"apartment"`
		AccountNumber string `json:"account_number"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	
	// В реальном проекте берем userID из JWT (тут заглушка userID=1)
	userID := 1 
	
	_, err := storage.DB.Exec(`
		INSERT INTO user_addresses (user_id, address_id, apartment, account_number) 
		VALUES ($1, $2, $3, $4) 
		ON CONFLICT (user_id, address_id) DO NOTHING`,
		userID, req.AddressID, req.Apartment, req.AccountNumber)
		
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}