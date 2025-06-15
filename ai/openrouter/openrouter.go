package openrouter

import (
	"context"
	"fmt"

	"net/http"
	"os"
	"time"

	"github.com/godovasik/dawgobot/logger"
	"github.com/sashabaranov/go-openai"
	"gopkg.in/yaml.v3"
)

var Characters map[string]string

type CharactersConfig struct {
	Characters map[string]string `yaml:"characters"`
}

func LoadCharacters() error {
	// Читаем YAML файл
	data, err := os.ReadFile("ai/openrouter/prompts.yaml")
	if err != nil {
		return fmt.Errorf("cant read file: %w", err)
	}

	// Парсим YAML
	var config CharactersConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return fmt.Errorf("cant unmarshal charagers: %w", err)
	}

	// Присваиваем загруженных персонажей
	Characters = config.Characters

	logger.Infof("Загружено %d персонажей", len(Characters))
	return nil
}

type customHeadersTransport struct {
	underlying http.RoundTripper
	headers    map[string]string
}

func (c customHeadersTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	return c.underlying.RoundTrip(req)
}

func GetNewClient() (*openai.Client, error) {
	apiKey := os.Getenv("OPENROUTER_TOKEN")

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://openrouter.ai/api/v1"

	// Кастомный HTTP клиент с нужными заголовками
	config.HTTPClient = &http.Client{
		Transport: customHeadersTransport{
			underlying: http.DefaultTransport,
			headers: map[string]string{
				"HTTP-Referer": "https://www.dawgobot.com", // ОБЯЗАТЕЛЬНО!
				"X-Title":      "dawgobot",                 // Желательно
			},
		},
		Timeout: 60 * time.Second,
	}
	client := openai.NewClientWithConfig(config)
	return client, nil
}

func GenerateResponse(ctx context.Context, client *openai.Client, character string, str string) (string, error) {
	req := openai.ChatCompletionRequest{
		Model:  "deepseek/deepseek-chat-v3-0324:free",
		Stream: false,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "system",
				Content: Characters[character],
			},
			{
				Role:    "user",
				Content: str,
			},
		},
	}
	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}
