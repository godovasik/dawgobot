package hface

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

const (
	baseURL = "https://fancyfeast-joy-caption-alpha-two.hf.space"
	apiPath = "/call/stream_chat"
)

// APIRequest представляет структуру запроса к API
type APIRequest struct {
	Data []interface{} `json:"data"`
}

// APIResponse представляет ответ от API
type APIResponse struct {
	EventID string `json:"event_id"`
}

// StreamResponse представляет потоковый ответ
type StreamResponse struct {
	Msg  string        `json:"msg"`
	Data []interface{} `json:"data"`
}

// ImageCaptionClient клиент для работы с API генерации описаний
type ImageCaptionClient struct {
	httpClient *http.Client
	baseURL    string
}

// NewImageCaptionClient создает новый клиент
func NewImageCaptionClient() *ImageCaptionClient {
	return &ImageCaptionClient{
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		baseURL: baseURL,
	}
}

// uploadImageToTempURL загружает изображение и возвращает временный URL
func (c *ImageCaptionClient) uploadImageToTempURL(imageBytes []byte) (string, error) {
	// Создаем multipart form для загрузки файла
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Добавляем файл в форму
	part, err := writer.CreateFormFile("file", "image.jpg")
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = part.Write(imageBytes)
	if err != nil {
		return "", fmt.Errorf("failed to write image data: %w", err)
	}

	writer.Close()

	// Отправляем запрос на загрузку
	uploadURL := c.baseURL + "/upload"
	req, err := http.NewRequest("POST", uploadURL, &buf)
	if err != nil {
		return "", fmt.Errorf("failed to create upload request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to upload image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("upload failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read upload response: %w", err)
	}

	// Парсим ответ для получения URL
	var uploadResp []string
	if err := json.Unmarshal(body, &uploadResp); err != nil {
		return "", fmt.Errorf("failed to parse upload response: %w", err)
	}

	if len(uploadResp) == 0 {
		return "", fmt.Errorf("empty upload response")
	}

	return uploadResp[0], nil
}

// GenerateCaption генерирует описание изображения
func (c *ImageCaptionClient) GenerateCaption(imageBytes []byte) (string, error) {
	return c.GenerateCaptionWithOptions(imageBytes, CaptionOptions{})
}

// CaptionOptions опции для генерации описания
type CaptionOptions struct {
	CaptionType   string   // "Descriptive", "Training Prompt", "MidJourney", "Booru tag list", "Booru-like tag list", "Art Critic", "Product Listing", "Social Media Post"
	CaptionLength string   // "any", "very short", "short", "medium-length", "long", "very long"
	ExtraOptions  []string // различные дополнительные опции
	PersonName    string   // имя персонажа если применимо
	CustomPrompt  string   // кастомный промпт (переопределяет другие настройки)
}

// GenerateCaptionWithOptions генерирует описание с дополнительными опциями
func (c *ImageCaptionClient) GenerateCaptionWithOptions(imageBytes []byte, options CaptionOptions) (string, error) {
	// Устанавливаем значения по умолчанию
	if options.CaptionType == "" {
		options.CaptionType = "Descriptive"
	}
	if options.CaptionLength == "" {
		options.CaptionLength = "any"
	}
	if options.PersonName == "" {
		options.PersonName = "Hello!!"
	}
	if options.CustomPrompt == "" {
		options.CustomPrompt = "Hello!!"
	}

	// Загружаем изображение и получаем временный URL
	imageURL, err := c.uploadImageToTempURL(imageBytes)
	if err != nil {
		return "", fmt.Errorf("failed to upload image: %w", err)
	}

	// Подготавливаем данные запроса
	requestData := APIRequest{
		Data: []interface{}{
			map[string]string{"path": imageURL},
			options.CaptionType,
			options.CaptionLength,
			options.ExtraOptions,
			options.PersonName,
			options.CustomPrompt,
		},
	}

	// Сериализуем запрос
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Отправляем POST запрос
	postURL := c.baseURL + apiPath
	req, err := http.NewRequest("POST", postURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create POST request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send POST request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("POST request failed with status: %d", resp.StatusCode)
	}

	// Читаем ответ для получения event_id
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read POST response: %w", err)
	}

	// Извлекаем event_id из ответа
	eventID := c.extractEventID(string(body))
	if eventID == "" {
		return "", fmt.Errorf("failed to extract event_id from response")
	}

	// Отправляем GET запрос для получения результата
	return c.getStreamResult(eventID)
}

// extractEventID извлекает event_id из ответа
func (c *ImageCaptionClient) extractEventID(response string) string {
	// Простое извлечение event_id из JSON ответа
	lines := strings.Split(response, "\n")
	for _, line := range lines {
		if strings.Contains(line, `"event_id"`) {
			parts := strings.Split(line, `"`)
			for i, part := range parts {
				if part == "event_id" && i+2 < len(parts) {
					return parts[i+2]
				}
			}
		}
	}
	return ""
}

// getStreamResult получает результат по event_id
func (c *ImageCaptionClient) getStreamResult(eventID string) (string, error) {
	getURL := fmt.Sprintf("%s%s/%s", c.baseURL, apiPath, eventID)

	req, err := http.NewRequest("GET", getURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create GET request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send GET request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GET request failed with status: %d", resp.StatusCode)
	}

	// Читаем потоковый ответ
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read GET response: %w", err)
	}

	// Парсим результат
	return c.parseStreamResponse(string(body))
}

// parseStreamResponse парсит потоковый ответ и извлекает описание
func (c *ImageCaptionClient) parseStreamResponse(response string) (string, error) {
	lines := strings.Split(response, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "data: ") {
			continue
		}

		// Убираем префикс "data: "
		jsonData := strings.TrimPrefix(line, "data: ")

		var streamResp StreamResponse
		if err := json.Unmarshal([]byte(jsonData), &streamResp); err != nil {
			continue
		}

		// Ищем сообщение с результатом
		if streamResp.Msg == "process_completed" && len(streamResp.Data) >= 2 {
			if caption, ok := streamResp.Data[1].(string); ok {
				return caption, nil
			}
		}
	}

	return "", fmt.Errorf("failed to extract caption from stream response")
}

// Пример использования
func main() {
	client := NewImageCaptionClient()

	// Пример чтения изображения из файла
	// imageBytes, err := os.ReadFile("path/to/your/image.jpg")
	// if err != nil {
	//     log.Fatal(err)
	// }

	// Для демонстрации используем пустой слайс
	// В реальном использовании здесь должны быть байты изображения
	var imageBytes []byte

	// Простая генерация описания
	caption, err := client.GenerateCaption(imageBytes)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Generated caption: %s\n", caption)

	// Генерация с дополнительными опциями
	options := CaptionOptions{
		CaptionType:   "Descriptive",
		CaptionLength: "medium-length",
		ExtraOptions:  []string{"If there is a person/character in the image you must refer to them as {name}."},
		PersonName:    "Alice",
		CustomPrompt:  "",
	}

	captionWithOptions, err := client.GenerateCaptionWithOptions(imageBytes, options)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Generated caption with options: %s\n", captionWithOptions)
}
