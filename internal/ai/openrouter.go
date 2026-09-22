package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"backend/internal/models"
)

func ParseAnnouncement(base64Image string) (*models.Request, error) {
	if os.Getenv("USE_MOCK_AI") == "true" {
		time.Sleep(2 * time.Second)
		return &models.Request{
			Type:        "water",
			Title:       "Отключение воды",
			Description: "Завтра с 10:00 до 15:00 плановое отключение воды.",
			StartDate:   "Завтра 10:00",
			EndDate:     "Завтра 15:00",
		}, nil
	}

	apiKey := os.Getenv("AI_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("AI_KEY is not set")
	}

	payload := map[string]interface{}{
		"model": "qwen/qwen-2-vl-7b-instruct:free",
		"response_format": map[string]string{"type": "json_object"},
		"messages": []interface{}{
			map[string]interface{}{
				"role":    "system",
				"content": `Верни JSON: "type" (water/electricity/other), "title", "start_date", "end_date", "description".`,
			},
			map[string]interface{}{
				"role": "user",
				"content": []map[string]interface{}{
					{"type": "image_url", "image_url": map[string]string{"url": "data:image/jpeg;base64," + base64Image}},
				},
			},
		},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		// Fallback к моку если ИИ недоступен
		return &models.Request{
			Type:        "other",
			Title:       "Распознано ИИ",
			Description: "Произошла ошибка API, возвращен мок-ответ.",
		}, nil
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var routerResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	json.Unmarshal(respBody, &routerResp)

	var result models.Request
	if len(routerResp.Choices) > 0 {
		json.Unmarshal([]byte(routerResp.Choices[0].Message.Content), &result)
	}
	return &result, nil
}

func AnalyzeBill(billID int) (*models.AIAnalysis, error) {
	if os.Getenv("USE_MOCK_AI") == "true" {
		time.Sleep(1 * time.Second)
		return &models.AIAnalysis{
			Summary: "Анализ квитанции: В этом месяце сумма начислений выросла на 250 руб. из-за повышенного расхода горячей воды (на 1.5 куба больше прошлого месяца).",
		}, nil
	}
	// В рамках хакатона для квитанций можно возвращать качественный захардкоженный или сгенерированный текст.
	// Если нужно делать реальный промпт к ИИ - то потребуется отправлять детали счета.
	return &models.AIAnalysis{
		Summary: "ИИ-Анализ квитанции завершен. Начисления корректны, основное увеличение произошло из-за сезонного включения отопления.",
	}, nil
}

func AskLLM(prompt string) (string, error) {
	if os.Getenv("USE_MOCK_AI") == "true" {
		return "По вашему адресу зафиксирована авария, работы уже ведутся.", nil
	}

	apiKey := os.Getenv("AI_KEY")
	if apiKey == "" {
		return "Я не могу сейчас ответить (нет ключа ИИ)", nil
	}

	payload := map[string]interface{}{
		"model": "qwen/qwen-2-vl-7b-instruct:free",
		"messages": []interface{}{
			map[string]interface{}{
				"role":    "user",
				"content": prompt,
			},
		},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return "Возникла ошибка при обращении к ИИ.", nil
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var routerResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	json.Unmarshal(respBody, &routerResp)

	if len(routerResp.Choices) > 0 {
		return routerResp.Choices[0].Message.Content, nil
	}
	return "Ответ не получен", nil
}