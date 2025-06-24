package client

import (
	"context"

	"github.com/godovasik/dawgobot/internal/ai/deepseek"
	"github.com/godovasik/dawgobot/internal/ai/openrouter"
	"github.com/godovasik/dawgobot/internal/database"
	"github.com/godovasik/dawgobot/internal/timeline"
	tw "github.com/godovasik/dawgobot/internal/twitch"
)

type ClientBuilder struct {
	Client *Client
}

func NewClientBuilder() *ClientBuilder {
	return &ClientBuilder{Client: &Client{}}
}

func (b *ClientBuilder) Build() *Client {
	return b.Client
}

func (b *ClientBuilder) WithTwitch(tw *tw.Client) *ClientBuilder {
	b.Client.TWClient = tw
	return b
}

func (b *ClientBuilder) WithDeepseek(ds *deepseek.Client) *ClientBuilder {
	b.Client.DSClient = ds
	return b
}

func (b *ClientBuilder) WithDB(db *database.DB) *ClientBuilder {
	b.Client.DB = db
	return b
}

func (b *ClientBuilder) WithTimeline(tl *timeline.Timeline) *ClientBuilder {
	b.Client.Timeline = tl
	return b
}

func (b *ClientBuilder) WithContext(ctx context.Context, cancel context.CancelFunc) *ClientBuilder {
	b.Client.ctx = ctx
	b.Client.cancel = cancel
	return b
}

func (b *ClientBuilder) WithGemeni(gmn *openrouter.Client) *ClientBuilder {
	b.Client.Gemeni = gmn
	return b
}
