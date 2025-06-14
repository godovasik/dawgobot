package ollama

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type OllamaRequest struct {
	Model  string   `json:"model"`
	Prompt string   `json:"prompt"`
	Images []string `json:"images"`
	Stream bool     `json:"stream"`
}

type OllamaResponse struct {
	Model              string `json:"model"`
	Response           string `json:"response"`
	Done               bool   `json:"done"`
	Context            []int  `json:"context,omitempty"`
	TotalDuration      int64  `json:"total_duration,omitempty"`
	LoadDuration       int64  `json:"load_duration,omitempty"`
	PromptEvalDuration int64  `json:"prompt_eval_duration,omitempty"`
	EvalDuration       int64  `json:"eval_duration,omitempty"`
}

// DescribeImage принимает путь к изображению и возвращает его описание от LLaVA

func DescribeImageBytes(imageBytes []byte) (string, error) {
	imageBytes, err := ResizeImageBytes(imageBytes)
	if err != nil {
		return "", err
	}

	imageData := base64.StdEncoding.EncodeToString(imageBytes)
	// Читаем и кодируем изображение в base64

	// Создаем запрос к Ollama
	request := OllamaRequest{
		Model:  "llava",
		Prompt: "Describe what you see in this image. Focus on the main elements, setting, and any important details. Be clear and concise.",
		Images: []string{imageData},
		Stream: false,
	}

	// Отправляем запрос
	response, err := sendRequest(request)
	if err != nil {
		return "", fmt.Errorf("ошибка при отправке запроса: %w", err)
	}

	return response.Response, nil
}

func sendRequest(req OllamaRequest) (*OllamaResponse, error) {
	// Сериализуем запрос в JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка сериализации JSON: %w", err)
	}

	// Отправляем POST запрос к Ollama API
	resp, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("ошибка HTTP запроса: %w", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP ошибка: %d", resp.StatusCode)
	}

	// Читаем тело ответа
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	// Парсим JSON ответ
	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return nil, fmt.Errorf("ошибка парсинга JSON: %w", err)
	}

	return &ollamaResp, nil
}
