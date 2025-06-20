package client

import (
	"context"
	"time"

	twitch "github.com/gempir/go-twitch-irc/v4" // костыль пиздец
	"github.com/godovasik/dawgobot/internal/ai/deepseek"
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

// TODO:
func (c *Client) MonitorChatEvents(channels ...string) error {
	eventCh := make(chan timeline.Event, 100)
	//FIXME: вынести в горутину
	// defer close(eventCh)

	c.TWClient.TWClient.OnPrivateMessage(c.GetHandleMonitor(eventCh))

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
		event := timeline.Event{
			Type:      timeline.EventChat,
			Content:   message.Message,
			Author:    message.User.Name,
			Streamer:  message.Channel,
			Timestamp: time.Now(),
		}
		eventCh <- event
		// и тут дальше анализ картинок, спич ту текст, скриншоты, все в таймлайн.
	}
}
