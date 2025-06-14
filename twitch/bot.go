package twitch

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	tw "github.com/gempir/go-twitch-irc/v4"
)

func NewClient() (*tw.Client, error) {
	access_token := os.Getenv("ACCESS_TOKEN")
	if access_token == "" {
		return nil, fmt.Errorf("variable ACCESS_TOKEN is not set")
	}
	client := tw.NewClient("dawgobot", fmt.Sprintf("oauth:%s", access_token))
	return client, nil
}

// TODO: logger
func MonitorChannelChat(client *tw.Client) error {
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: go run main.go <channel>")
	}

	username := os.Args[1]
	client.OnPrivateMessage(func(message tw.PrivateMessage) {
		timenow := message.Time.Format("02-01-06 15:04:05")
		fmt.Printf("[%v] %s: %s\n", timenow, message.User.DisplayName, message.Message)
	})

	client.Join(username)
	if err := client.Connect(); err != nil {
		return err
	}
	return nil
}

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
