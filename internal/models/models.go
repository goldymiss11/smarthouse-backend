package models

type User struct {
	ID   int    `json:"id"`
	VkID string `json:"vk_id"`
	Role string `json:"role"`
}

type Request struct {
	ID          int    `json:"id"`
	UserID      int    `json:"user_id,omitempty"`
	AddressID   int    `json:"address_id,omitempty"`
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	Status      string `json:"status"`
	PhotoURL    string `json:"photo_url,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
}

type MapObject struct {
	AddressID   int     `json:"address_id"`
	FullAddress string  `json:"full_address"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	Status      string  `json:"status"`
}

type MapHouse struct {
	HouseID           int     `json:"houseId"`
	Lat               float64 `json:"lat"`
	Lng               float64 `json:"lng"`
	WaterStatus       string  `json:"water_status"`
	ElectricityStatus string  `json:"electricity_status"`
}

type Bill struct {
	ID     int     `json:"id"`
	Month  string  `json:"month"`
	Amount float64 `json:"amount"`
	IsPaid bool    `json:"is_paid"`
}

type Offer struct {
	ID          int    `json:"id"`
	PartnerName string `json:"partner_name"`
	BadgeText   string `json:"badge_text"`
	OfferText   string `json:"offer_text"`
	IconURL     string `json:"icon_url"`
	ActionType  string `json:"action_type"`
	ActionData  string `json:"action_data"`
	IsLocked    bool   `json:"is_locked"`
}

type FeedItem struct {
	ID        int    `json:"id"`
	AddressID int    `json:"address_id,omitempty"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	Category  string `json:"category"`
	CreatedAt string `json:"created_at,omitempty"`
}

type MeterReading struct {
	ID          int     `json:"id"`
	Water       float64 `json:"water"`
	Electricity float64 `json:"electricity"`
	SubmittedAt string  `json:"submitted_at,omitempty"`
}

type AIAnalysis struct {
	Summary string `json:"ai_analysis_summary"`
}