package main

import (
	"log"
	"net/http"
	"os"

	"backend/internal/handlers"
	"backend/internal/storage"

	"github.com/joho/godotenv"
)

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Разрешаем сервер-сервер запросы (без Origin, например от МАХ) и эндпоинт вебхука напрямую
		if r.URL.Path == "/api/webhook/max" || r.Header.Get("Origin") == "" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Загружаем ENV. В докере переменные будут прокинуты напрямую.
	godotenv.Load("../../.env")

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL environment variable is not set")
	}
	storage.InitDB(dbURL)

	mux := http.NewServeMux()

	// Авторизация
	mux.HandleFunc("POST /api/auth/login", handlers.LoginHandler)
	mux.HandleFunc("POST /api/user/link-address", handlers.LinkAddressHandler)

	// Житель
	mux.HandleFunc("GET /api/feed", handlers.GetFeedHandler)
	mux.HandleFunc("GET /api/bills", handlers.GetBillsHandler)
	mux.HandleFunc("POST /api/bills/{id}/pay", handlers.PayBillHandler)

	// Эндпоинты по DATA-API.yml
	mux.HandleFunc("GET /api/v1/profile/privileges", handlers.GetPrivilegesHandler)
	mux.HandleFunc("GET /api/bills/{id}/ai-analysis", handlers.BillAIAnalysisHandler)
	mux.HandleFunc("POST /api/requests/ai-recognize", handlers.AnalyzePhotoHandler)

	mux.HandleFunc("GET /api/meters", handlers.GetMetersHandler)
	mux.HandleFunc("POST /api/meters", handlers.PostMetersHandler)
	mux.HandleFunc("POST /api/requests", handlers.CreateRequestHandler)

	// УК
	mux.HandleFunc("GET /api/uk/requests", handlers.GetUKRequestsHandler)
	
	// Эндпоинты по DATA-API.yml (УК/Модератор)
	mux.HandleFunc("PATCH /api/requests/{id}/status", handlers.UpdateRequestStatusHandler)
	mux.HandleFunc("GET /api/map/houses", handlers.GetMapHousesHandler)

	// Старые эндпоинты для обратной совместимости, если где-то на фронте они захардкожены
	mux.HandleFunc("POST /api/uk/requests/{id}/approve", handlers.ApproveRequestHandler)
	mux.HandleFunc("POST /api/uk/requests/{id}/reject", handlers.RejectRequestHandler)

	mux.HandleFunc("POST /api/uk/broadcast", handlers.BroadcastHandler)
	mux.HandleFunc("GET /api/uk/objects", handlers.GetUKObjectsHandler)
	mux.HandleFunc("GET /api/uk/analytics", handlers.GetUKAnalyticsHandler)

	// Интеграция с МАХ (Webhook)
	mux.HandleFunc("POST /api/webhook/max", handlers.MaxWebhookHandler)

	log.Println("Сервер запущен на :8080")
	log.Fatal(http.ListenAndServe(":8080", enableCORS(mux)))
}
