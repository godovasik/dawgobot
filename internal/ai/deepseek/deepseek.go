package deepseek

import (
	"context"
	"fmt"
	"time"

	"os"

	"github.com/godovasik/dawgobot/internal/timeline"
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
	Tl        *timeline.Timeline
}

func LoadCharacters() error {
	// Читаем YAML файл
	data, err := os.ReadFile("internal/ai/openrouter/prompts.yaml")
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

func NewClient(tl *timeline.Timeline) (*Client, error) {
	apiKey := os.Getenv("DEEPSEEK_TOKEN")
	if apiKey == "" {
		return nil, fmt.Errorf("DEEPSEEK_TOKEN not set")
	}
	logger.Info("token set...")

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://api.deepseek.com"
	openaiCli := openai.NewClientWithConfig(config)
	return &Client{OpenaiCli: openaiCli, Tl: tl}, nil
}

func (c *Client) GetResponse(character, message string) (string, error) {
	prompt, ok := Characters[character]
	if !ok {
		return "", fmt.Errorf("no such character")
	}
	req := openai.ChatCompletionRequest{
		Model: "deepseek-chat",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: prompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: message,
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	resp, err := c.OpenaiCli.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("Ошибка при отправке запроса: %v", err)

	}

	return fmt.Sprint(resp.Choices[0].Message.Content), nil
}
