package twitch

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	tw "github.com/gempir/go-twitch-irc/v4"
	"github.com/godovasik/dawgobot/logger"
)

type Client struct {
	TWClient *tw.Client

	httpClient   *http.Client
	clientID     string
	clientSecret string
	appToken     string
}

func NewClient() (*Client, error) {
	accessToken := os.Getenv("ACCESS_TOKEN")
	if accessToken == "" {
		return nil, fmt.Errorf("variable ACCESS_TOKEN is not set")
	}

	clientID := os.Getenv("TWITCH_CLIENT_ID")
	if clientID == "" {
		return nil, fmt.Errorf("variable TWITCH_CLIENT_ID is not set")
	}

	clientSecret := os.Getenv("TWITCH_CLIENT_SECRET")
	if clientSecret == "" {
		return nil, fmt.Errorf("variable TWITCH_CLIENT_SECRET is not set")
	}

	twClient := tw.NewClient("dawgobot", fmt.Sprintf("oauth:%s", accessToken))

	client := &Client{
		TWClient:     twClient,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
		clientID:     clientID,
		clientSecret: clientSecret,
	}

	// Получаем app access token для API запросов
	if err := client.getAppToken(); err != nil {
		return nil, fmt.Errorf("failed to get app token: %w", err)
	}

	logger.Info("twitch client initialized")
	return client, nil
}

// TODO:
// сейчас я вызываю тут запросы к дипсику и оламе, по хорошему это нужно делать где-то вне.
// как начнут проблемы вылазить - переделаю

func FindURLs(text string) []string {
	var urls []string

	// Ищем полные URL с протоколом
	urlRegex := regexp.MustCompile(`(?i)\b(?:https?://|www\.)[^\s<>"{}|\\^` + "`" + `\[\]]+`)
	urls = append(urls, urlRegex.FindAllString(text, -1)...)

	// Ищем домены без протокола
	domainRegex := regexp.MustCompile(`(?i)\b[a-zA-Z0-9]([a-zA-Z0-9\-]*[a-zA-Z0-9])?\.([a-zA-Z]{2,})\b`)
	domains := domainRegex.FindAllString(text, -1)

	// Фильтруем, чтобы не дублировать уже найденные URL
	for _, domain := range domains {
		found := false
		for _, url := range urls {
			if strings.Contains(url, domain) {
				found = true
				break
			}
		}
		if !found {
			urls = append(urls, domain)
		}
	}

	return urls
}
