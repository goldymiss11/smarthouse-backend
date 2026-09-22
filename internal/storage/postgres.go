package storage

import (
	"database/sql"
	"log"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var DB *sql.DB

func InitDB(dsn string) {
	var err error
	DB, err = sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Ошибка DSN: %v", err)
	}

	if err := DB.Ping(); err != nil {
		log.Fatalf("Supabase недоступен: %v", err)
	}
	log.Println("Успешное подключение к базе данных")
}

func CheckUserDebt(userID int) (bool, error) {
	var unpaidCount int
	err := DB.QueryRow("SELECT COUNT(id) FROM bills WHERE user_id = $1 AND is_paid = false", userID).Scan(&unpaidCount)
	if err != nil {
		return false, err
	}
	return unpaidCount > 0, nil
}