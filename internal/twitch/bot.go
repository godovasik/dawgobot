package twitch

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	tw "github.com/gempir/go-twitch-irc/v4"
	"github.com/godovasik/dawgobot/internal/ai/ollama"
	"github.com/godovasik/dawgobot/internal/timeline"
	"github.com/godovasik/dawgobot/logger"
)

type Client struct {
	TwitchClient *tw.Client
	Timeline     *timeline.Timeline
	Connetced    bool // пока не юзаю, хз зачем оно
}

func NewClient(tl *timeline.Timeline) (*Client, error) {
	access_token := os.Getenv("ACCESS_TOKEN")
	if access_token == "" {
		return nil, fmt.Errorf("variable ACCESS_TOKEN is not set")
	}
	twClient := tw.NewClient("dawgobot", fmt.Sprintf("oauth:%s", access_token))
	return &Client{twClient, tl, false}, nil
}

func NewTwitchClient() (*tw.Client, error) {
	access_token := os.Getenv("ACCESS_TOKEN")
	if access_token == "" {
		return nil, fmt.Errorf("variable ACCESS_TOKEN is not set")
	}
	client := tw.NewClient("dawgobot", fmt.Sprintf("oauth:%s", access_token))
	return client, nil
}

// пишет в таймлайн содержимое чата
// надо будет убрать отсюда запись в файл, это будет делать метод таймлайна.
func (c *Client) MonitorChannelChat(username string) error {
	filename := fmt.Sprintf("logger/chatLogs/%s.txt", username)
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	timenow := time.Now().Format("02-01-06 15:04:05")
	file.WriteString(fmt.Sprintf("[%v] Starting monitoring channel \"%s\"\n", timenow, username))

	defer file.WriteString(fmt.Sprintf(
		"[%v] Stopping monitoring channel \"%s\"\n\n", time.Now().Format("02-01-06 15:04:05"), username),
	)

	event := timeline.Event{
		Type:      timeline.EventBot,
		Content:   fmt.Sprintf("Starting monitoring channel \"%s\"", username),
		Author:    "[botInfo]", // TODO: придумать более красивое решение
		Timestamp: time.Now(),
	}
	c.Timeline.AddEvent(event)

	// Обработка Ctrl+C
	// exitCh := make(chan os.Signal, 1)
	// signal.Notify(exitCh, os.Interrupt, syscall.SIGTERM)
	//
	// go func() {
	// 	<-exitCh
	// 	file.WriteString(fmt.Sprintf("[%v] Stopping monitoring channel \"%s\"\n\n",
	// 		time.Now().Format("02-01-06 15:04:05"), username))
	// 	file.Close()
	//
	// 	event := timeline.Event{
	// 		Type:      timeline.EventBot,
	// 		Content:   fmt.Sprintf("Stopping monitoring channel \"%s\"", username),
	// 		Author:    "[botInfo]", // TODO: придумать более красивое решение
	// 		Timestamp: time.Now(),
	// 	}
	// 	c.timeline.AddEvent(event)
	// 	os.Exit(0)
	// }()

	c.TwitchClient.OnPrivateMessage(func(message tw.PrivateMessage) {
		timenow := message.Time.Format("02-01-06 15:04:05")
		// fmt.Printf("[%v] %s: %s\n", timenow, message.User.DisplayName, message.Message)
		file.WriteString(fmt.Sprintf("[%v] %s: %s\n", timenow, message.User.DisplayName, message.Message))
		event := timeline.Event{
			Type:      timeline.EventChat,
			Content:   message.Message,
			Author:    message.User.Name,
			Timestamp: time.Now(),
		}
		c.Timeline.AddEvent(event)
	})

	c.TwitchClient.Join(username)

	return nil
}

func ScanForImagesHandler() func(message tw.PrivateMessage) {
	return func(message tw.PrivateMessage) {
		logger.Info(fmt.Sprintf("%s: %s\n", message.User.DisplayName, message.Message))

		urls := FindURLs(message.Message)
		for _, u := range urls {
			ok, err := ollama.CheckUrl(u)
			if err != nil {
				logger.Info("cant get url " + u) // this is ugly
				continue
			}
			if !ok {
				logger.Info(u + " is not an image")
				continue
			}

			data, err := ollama.GetImage(u)
			if err != nil {
				logger.Info("error getting image:" + err.Error())
				continue
			}
			resp, err := ollama.DescribeImageBytes(data)
			if err != nil {
				logger.Info("ollama error:" + err.Error())
				continue
			}
			fmt.Printf("image url:%s\ndescription: %s\n", u, resp)

		}
	}
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
