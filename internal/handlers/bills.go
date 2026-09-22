package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"backend/internal/ai"
	"backend/internal/models"
	"backend/internal/storage"
)

func GetBillsHandler(w http.ResponseWriter, r *http.Request) {
	userID := 1 // Заглушка до внедрения JWT
	rows, err := storage.DB.Query("SELECT id, month, amount, is_paid FROM bills WHERE user_id = $1 ORDER BY month DESC", userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var bills []models.Bill
	for rows.Next() {
		var b models.Bill
		if err := rows.Scan(&b.ID, &b.Month, &b.Amount, &b.IsPaid); err == nil {
			bills = append(bills, b)
		}
	}
	
	if bills == nil {
		bills = []models.Bill{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bills)
}

func PayBillHandler(w http.ResponseWriter, r *http.Request) {
	billID := r.PathValue("id")
	storage.DB.Exec("UPDATE bills SET is_paid = true WHERE id = $1", billID)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "paid"})
}

func BillAIAnalysisHandler(w http.ResponseWriter, r *http.Request) {
	billIDStr := r.PathValue("id")
	billID, _ := strconv.Atoi(billIDStr)
	
	analysis, err := ai.AnalyzeBill(billID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analysis)
}