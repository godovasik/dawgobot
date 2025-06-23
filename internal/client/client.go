package client

import (
	"context"
	"time"

	twitch "github.com/gempir/go-twitch-irc/v4" // костыль пиздец
	"github.com/godovasik/dawgobot/internal/ai/deepseek"
	"github.com/godovasik/dawgobot/internal/ai/ollama"
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
	DSClient *deepseek.Client
	DB       *database.DB

	ctx    context.Context
	cancel context.CancelFunc

	Connetced bool // пока не юзаю, хз зачем оно
}

// TODO:graceful shutdow, write "start/end monitoring to logs"
func (c *Client) MonitorChatEvents(channels ...string) error {
	eventCh := make(chan timeline.Event, 100)
	// FIXME: вынести в горутину
	// defer close(eventCh)

	// NOTE: щас тут с картинками
	c.TWClient.TWClient.OnPrivateMessage(c.GetHandleMonitorWithImages(eventCh))

	// NOTE: эта без катинок
	// c.TWClient.TWClient.OnPrivateMessage(c.GetHandleMonitor(eventCh))

	batch := make([]timeline.Event, 0, 100)

	flushBatch := func() {
		logger.Infof("flushing %d events", len(batch))
		timeline.PrintEvents(batch)
		if err := c.DB.AddEvents(batch); err != nil {
			logger.Errorf("db error: %w", err)
		}
		batch = batch[:0]
	}

	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if len(batch) > 0 {
					flushBatch()
				}

			case event, ok := <-eventCh:
				if !ok {
					flushBatch()
					return
				}

				batch = append(batch, event)
				if len(batch) >= 50 {
					flushBatch()
				}

			}
		}
	}()

	c.TWClient.TWClient.Join(channels...)
	return nil
}

func (c *Client) GetHandleMonitor(eventCh chan timeline.Event) func(message twitch.PrivateMessage) {
	return func(message twitch.PrivateMessage) {
		event := messageToEvent(message)
		eventCh <- event
		// и тут дальше анализ картинок, спич ту текст, скриншоты, все в таймлайн.
	}
}

func (c *Client) GetHandleMonitorWithImages(eventCh chan timeline.Event) func(message twitch.PrivateMessage) {
	return func(message twitch.PrivateMessage) {
		event := messageToEvent(message)
		eventCh <- event

		urls := tw.FindURLs(event.Content)
		if len(urls) == 0 {
			return
		}

		for _, u := range urls {
			ok, err := ollama.CheckUrl(u)
			if err != nil {
				logger.Errorf("err: %w, cant check url: %s", err, u)
				continue
			}
			if !ok {
				logger.Infof("not an image: %s", u)
				continue
			}

			logger.Infof("found image: %s", u)
			image, err := ollama.GetImage(u)
			if err != nil {
				logger.Error(err.Error())
				continue
			}

			desc, err := ollama.DescribeImageBytes(image)
			// TODO: сделать gemeni сюда
			if err != nil {
				logger.Error(err.Error())
				continue
			}

			imageEvent := timeline.Event{
				Type:      timeline.EventImage,
				Content:   desc,
				Author:    event.Author,
				Streamer:  event.Streamer,
				Timestamp: event.Timestamp.Add(time.Millisecond), // наносекунду для последовательности
			}
			eventCh <- imageEvent

		}
		// и тут дальше анализ картинок, спич ту текст, скриншоты, все в таймлайн.
	}
}

func messageToEvent(message twitch.PrivateMessage) timeline.Event {
	event := timeline.Event{
		Type:      timeline.EventChat,
		Content:   message.Message,
		Author:    message.User.Name,
		Streamer:  message.Channel,
		Timestamp: time.Now(),
	}
	return event
}
