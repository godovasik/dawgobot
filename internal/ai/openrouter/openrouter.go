package openrouter

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
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

type Client struct {
	OpenaiCli *openai.Client
}

func LoadCharacters() error {
	// Читаем YAML файл
	data, err := os.ReadFile("internal/ai/prompts.yaml")
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

	logging bool
}

func (c customHeadersTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	if c.logging {
		reqDump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			return nil, fmt.Errorf("ошибка при дампе запроса: %v", err)
		}
		logger.Infof("HTTP Request:\n%s\n", string(reqDump))
	}

	// Выполняем запрос
	resp, err := c.underlying.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// Логируем ответ, если включено логирование
	if c.logging {
		respDump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return nil, fmt.Errorf("ошибка при дампе ответа: %v", err)
		}
		logger.Infof("HTTP Response:\n%s\n", string(respDump))
	}

	return resp, nil
}

func GetNewClient(logging bool) (*Client, error) {
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
			logging: logging,
		},
		Timeout: 60 * time.Second,
	}
	OpenaiClient := openai.NewClientWithConfig(config)


	Client := Client{OpenaiClient}
	return &Client, nil
}

func (c *Client) GenerateResponseDeepseek(character string, str string) (string, error) {
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
	ctx := context.Background()
	resp, err := c.OpenaiCli.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}

	answer := resp.Choices
	if len(answer) == 0 {
		return "", fmt.Errorf("fuck me")
	}
	return resp.Choices[0].Message.Content, nil
}
