package client

import (
	"context"
	"fmt"
	"strings"
	"time"

	twitch "github.com/gempir/go-twitch-irc/v4" // костыль пиздец
	"github.com/godovasik/dawgobot/internal/ai/deepseek"
	"github.com/godovasik/dawgobot/internal/ai/ollama"
	"github.com/godovasik/dawgobot/internal/ai/openrouter"
	"github.com/godovasik/dawgobot/internal/database"
	"github.com/godovasik/dawgobot/internal/timeline"
	tw "github.com/godovasik/dawgobot/internal/twitch"
	"github.com/godovasik/dawgobot/logger"
)

// возможно стоит добавиь username в twitch client
// но если мне нужно мониторить нескольких стримеров?
type Client struct {
	TWClient *tw.Client
	Timeline *timeline.Timeline
	DB       *database.DB

	DSClient *deepseek.Client
	Gemeni   *openrouter.Client

	ctx    context.Context
	cancel context.CancelFunc

	Connetced bool // пока не юзаю, хз зачем оно
}

// MonitorChatEvents с graceful shutdown
// TODO: Понять
// i am a genius
// это будет основная функция, буду сюда все писать
func (c *Client) MonitorChatEvents(WithImages bool, channels ...string) error {
	eventCh := make(chan timeline.Event, 100)

	// Создаем события начала мониторинга для каждого канала
	for _, channel := range channels {
		startEvent := timeline.Event{
			Type:      timeline.EventGlobal,
			Content:   fmt.Sprintf("Starting monitoring for channel: %s", channel),
			Author:    "system",
			Streamer:  channel,
			Timestamp: time.Now(),
		}
		eventCh <- startEvent
	}

	// Выбираем обработчик в зависимости от WithImages
	c.TWClient.TWClient.OnPrivateMessage(c.GetHandleMonitor(eventCh, WithImages))

	// Запускаем горутину для обработки батчей
	batchDone := make(chan struct{})
	go c.processBatches(eventCh, batchDone)

	// Подключаемся к каналам
	c.TWClient.TWClient.Join(channels...)

	// Ждем сигнала отмены контекста
	<-c.ctx.Done()
	logger.Info("Context cancelled, shutting down...")

	// Создаем события остановки мониторинга для каждого канала
	for _, channel := range channels {
		stopEvent := timeline.Event{
			Type:      timeline.EventGlobal,
			Content:   fmt.Sprintf("Stopping monitoring for channel: %s", channel),
			Author:    "system",
			Streamer:  channel,
			Timestamp: time.Now(),
		}
		eventCh <- stopEvent
	}
	time.Sleep(100 * time.Millisecond)
	close(eventCh)
	logger.Info("Event channel closed")

	// Ждем завершения обработки батчей
	<-batchDone

	return c.ctx.Err()
}

// processBatches - отдельная функция для обработки батчей событий
func (c *Client) processBatches(eventCh <-chan timeline.Event, done chan<- struct{}) {
	defer close(done)

	batch := make([]timeline.Event, 0, 100)
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	flushBatch := func() {
		if len(batch) == 0 {
			return
		}

		logger.Infof("Flushing %d events", len(batch))
		timeline.PrintEvents(batch)

		if err := c.DB.AddEvents(batch); err != nil {
			logger.Errorf("Database error: %v", err)
		}

		batch = batch[:0]
	}

	for {
		select {
		case <-ticker.C:
			flushBatch()

		case event, ok := <-eventCh:
			if !ok {
				// Канал закрыт, обрабатываем последний батч и выходим
				logger.Info("Event channel closed, processing final batch")
				flushBatch()
				return
			}

			batch = append(batch, event)

			// Автоматический flush при достижении лимита
			if len(batch) >= 50 {
				flushBatch()
			}
		}
	}
}

// GetHandleMonitor остается без изменений
func (c *Client) GetHandleMonitor(eventCh chan timeline.Event, withImages bool) func(message twitch.PrivateMessage) {
	return func(message twitch.PrivateMessage) {
		event := messageToEvent(message)

		// Проверяем, не закрыт ли канал
		select {
		case eventCh <- event:
		case <-c.ctx.Done():
			logger.Info("Context cancelled, skipping event")
			return
		default:
			logger.Warn("Event channel full, dropping event")
		}

		if withImages {
			c.processImages(eventCh, event)
		}
	}
}

func (c *Client) processImages(eventCh chan timeline.Event, event timeline.Event) {
	urls := tw.FindURLs(event.Content)
	if len(urls) == 0 {
		return
	}

	for _, u := range urls {
		// Проверяем контекст перед каждой операцией
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		ok, err := ollama.CheckUrl(u)
		if err != nil {
			logger.Errorf("Error checking URL %s: %v", u, err)
			continue
		}
		if !ok {
			logger.Infof("Not an image: %s", u)
			continue
		}

		logger.Infof("Found image: %s", u)
		desc, err := c.Gemeni.DescribeImageGemeni(c.ctx, u)
		if err != nil {
			logger.Errorf("Error describing image %s: %v", u, err)
			continue
		}

		imageEvent := timeline.Event{
			Type:      timeline.EventImage,
			Content:   desc,
			Author:    event.Author,
			Streamer:  event.Streamer,
			Timestamp: event.Timestamp.Add(time.Millisecond), // для правильной последовательности
		}

		// Отправляем событие изображения
		select {
		case eventCh <- imageEvent:
		case <-c.ctx.Done():
			return
		default:
			logger.Warn("Event channel full, dropping image event")
		}
	}
}

func messageToEvent(message twitch.PrivateMessage) timeline.Event {
	return timeline.Event{
		Type:      timeline.EventChat,
		Content:   message.Message,
		Author:    message.User.Name,
		Streamer:  message.Channel,
		Timestamp: time.Now(),
	}
}

func (c *Client) ReactToImages(channels ...string) error {
	c.TWClient.TWClient.OnPrivateMessage(c.GetHandleSimpleImageResponse())
	c.TWClient.TWClient.Join(channels...)
	return nil
}

func (c *Client) GetHandleSimpleImageResponse() func(message twitch.PrivateMessage) {
	return func(message twitch.PrivateMessage) {
		urls := tw.FindURLs(message.Message)
		isReplying := strings.Contains(message.Message, "@dawgobot")
		if len(urls) == 0 {
			if isReplying {
				answer := "чел я пока только умею на картинки отвечать"
				c.TWClient.TWClient.Reply(message.Channel, message.ID, answer)
			}
			return
		}

		u := urls[0] // допустим у нас одна картинка
		ok, err := ollama.CheckUrl(u)
		if err != nil {
			logger.Errorf("Error checking URL %s: %v", u, err)
			return
		}
		if !ok {
			if isReplying {
				answer := "это не картинка это хуй знает что"
				c.TWClient.TWClient.Reply(message.Channel, message.ID, answer)
			}

			logger.Infof("Not an image: %s", u)
			return
		}

		logger.Infof("Found image: %s", u)
		desc, err := c.Gemeni.DescribeImageGemeni(c.ctx, u)
		if err != nil {
			logger.Errorf("Error describing image %s: %v", u, err)
			return
		}

		resp, err := c.DSClient.GetResponse("image", desc)
		if err != nil {
			logger.Errorf("err from deepseekk: %w", err)
			return
		}

		logger.Infof("replying to @%s", message.User.DisplayName)
		c.TWClient.TWClient.Reply(message.Channel, message.ID, resp)
	}
}
