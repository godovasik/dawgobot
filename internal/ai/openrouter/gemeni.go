package openrouter

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

type ContentPart struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageURL *struct {
		URL string `json:"url"`
	} `json:"image_url,omitempty"`
}

func (c *Client) DescribeImageGemeni(ctx context.Context, url string) (string, error) {
	character := "describeImageShort"

	req := openai.ChatCompletionRequest{
		Model: "google/gemini-2.5-flash-lite-preview-06-17",
		Messages: []openai.ChatCompletionMessage{
			{
				Role: "user",
				MultiContent: []openai.ChatMessagePart{
					{
						Type: openai.ChatMessagePartTypeText,
						Text: Characters[character],
					},
					{
						Type: openai.ChatMessagePartTypeImageURL,
						ImageURL: &openai.ChatMessageImageURL{
							URL: url,
							// Detail: openai.ImageURLDetailAuto,
						},
					},
				},
			},
		},
	}

	resp, err := c.OpenaiCli.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}
