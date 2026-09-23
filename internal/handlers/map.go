package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"backend/internal/models"
	"backend/internal/storage"
)

func GetMapHousesHandler(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT a.id, a.lat, a.lng,
			CASE WHEN COUNT(req.id) FILTER (WHERE req.type = 'water' AND req.status != 'resolved') > 0 THEN 'repair' ELSE 'ok' END as water_status,
			CASE WHEN COUNT(req.id) FILTER (WHERE req.type = 'electricity' AND req.status != 'resolved') > 0 THEN 'repair' ELSE 'ok' END as electricity_status
		FROM addresses a
		LEFT JOIN requests req ON a.id = req.address_id
		GROUP BY a.id, a.lat, a.lng
	`
	rows, err := storage.DB.Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("DB Error in map: %v", err)
		return
	}
	defer rows.Close()

	var houses []models.MapHouse
	for rows.Next() {
		var h models.MapHouse
		if err := rows.Scan(&h.HouseID, &h.Lat, &h.Lng, &h.WaterStatus, &h.ElectricityStatus); err == nil {
			houses = append(houses, h)
		}
	}
	
	if houses == nil {
		houses = []models.MapHouse{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(houses)
}
