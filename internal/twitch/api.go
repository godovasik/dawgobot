package twitch

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// StreamerInfo содержит всю информацию о стримере и трансляции
type StreamerInfo struct {
	// Основная информация о стриме
	IsLive      bool      `json:"is_live"`
	StreamID    string    `json:"stream_id"`
	Title       string    `json:"title"`
	GameName    string    `json:"game_name"`
	GameID      string    `json:"game_id"`
	ViewerCount int       `json:"viewer_count"`
	StartedAt   time.Time `json:"started_at"`
	Language    string    `json:"language"`
	Tags        []string  `json:"tags"`

	// Информация о канале
	UserID          string `json:"user_id"`
	UserLogin       string `json:"user_login"`
	UserDisplayName string `json:"user_display_name"`

	// Голосования и предсказания
	ActivePolls       []Poll       `json:"active_polls"`
	ActivePredictions []Prediction `json:"active_predictions"`

	// Метаданные
	LastUpdated time.Time `json:"last_updated"`
}

type Poll struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"` // ACTIVE, COMPLETED, TERMINATED
	Duration  int       `json:"duration"`
	StartedAt time.Time `json:"started_at"`
	EndsAt    time.Time `json:"ends_at"`
	Choices   []Choice  `json:"choices"`
}

type Choice struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Votes int    `json:"votes"`
}

type Prediction struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	EndedAt     *time.Time `json:"ended_at,omitempty"`
	LockAt      time.Time  `json:"lock_at"`
	Outcomes    []Outcome  `json:"outcomes"`
	TotalPoints int        `json:"total_points"`
	TotalUsers  int        `json:"total_users"`
}

type Outcome struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Color  string `json:"color"`
	Users  int    `json:"users"`
	Points int    `json:"points"`
}

// Структуры для API ответов
type streamResponse struct {
	Data []struct {
		ID           string    `json:"id"`
		UserID       string    `json:"user_id"`
		UserLogin    string    `json:"user_login"`
		UserName     string    `json:"user_name"`
		GameID       string    `json:"game_id"`
		GameName     string    `json:"game_name"`
		Type         string    `json:"type"`
		Title        string    `json:"title"`
		ViewerCount  int       `json:"viewer_count"`
		StartedAt    time.Time `json:"started_at"`
		Language     string    `json:"language"`
		ThumbnailURL string    `json:"thumbnail_url"`
		TagIDs       []string  `json:"tag_ids"`
		Tags         []string  `json:"tags"`
	} `json:"data"`
}

type pollResponse struct {
	Data []struct {
		ID                string    `json:"id"`
		BroadcasterID     string    `json:"broadcaster_id"`
		BroadcasterName   string    `json:"broadcaster_name"`
		BroadcasterLogin  string    `json:"broadcaster_login"`
		Title             string    `json:"title"`
		Choices           []Choice  `json:"choices"`
		BitsVotingEnabled bool      `json:"bits_voting_enabled"`
		BitsPerVote       int       `json:"bits_per_vote"`
		Status            string    `json:"status"`
		Duration          int       `json:"duration"`
		StartedAt         time.Time `json:"started_at"`
	} `json:"data"`
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// getAppToken получает App Access Token для Twitch API
func (c *Client) getAppToken() error {
	data := url.Values{
		"client_id":     {c.clientID},
		"client_secret": {c.clientSecret},
		"grant_type":    {"client_credentials"},
	}

	resp, err := c.httpClient.PostForm("https://id.twitch.tv/oauth2/token", data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var tokenResp tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return err
	}

	c.appToken = tokenResp.AccessToken
	return nil
}

// makeAPIRequest выполняет запрос к Twitch API
func (c *Client) makeAPIRequest(endpoint string) (*http.Response, error) {
	req, err := http.NewRequest("GET", "https://api.twitch.tv/helix/"+endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Client-ID", c.clientID)
	req.Header.Set("Authorization", "Bearer "+c.appToken)

	return c.httpClient.Do(req)
}

// GetStreamerInfo получает полную информацию о стримере
func (c *Client) GetStreamerInfo(username string) (*StreamerInfo, error) {
	info := &StreamerInfo{
		LastUpdated: time.Now(),
	}

	// Получаем информацию о стриме
	if err := c.getStreamData(username, info); err != nil {
		return nil, fmt.Errorf("failed to get stream data: %w", err)
	}

	// Если стример онлайн, получаем дополнительную информацию
	if info.IsLive {
		// Получаем голосования
		if err := c.getPolls(info.UserID, info); err != nil {
			// Логируем ошибку, но не прерываем выполнение
			fmt.Printf("Warning: failed to get polls: %v\n", err)
		}

		// Получаем предсказания
		if err := c.getPredictions(info.UserID, info); err != nil {
			// Логируем ошибку, но не прерываем выполнение
			fmt.Printf("Warning: failed to get predictions: %v\n", err)
		}
	}

	return info, nil
}

// getStreamData получает основную информацию о стриме
func (c *Client) getStreamData(username string, info *StreamerInfo) error {
	resp, err := c.makeAPIRequest("streams?user_login=" + username)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var streamResp streamResponse
	if err := json.NewDecoder(resp.Body).Decode(&streamResp); err != nil {
		return err
	}

	if len(streamResp.Data) == 0 {
		// Стример оффлайн
		info.IsLive = false
		info.UserLogin = username
		return nil
	}

	stream := streamResp.Data[0]
	info.IsLive = true
	info.StreamID = stream.ID
	info.UserID = stream.UserID
	info.UserLogin = stream.UserLogin
	info.UserDisplayName = stream.UserName
	info.Title = stream.Title
	info.GameName = stream.GameName
	info.GameID = stream.GameID
	info.ViewerCount = stream.ViewerCount
	info.StartedAt = stream.StartedAt
	info.Language = stream.Language
	info.Tags = stream.Tags

	return nil
}

// getPolls получает активные голосования
func (c *Client) getPolls(broadcasterID string, info *StreamerInfo) error {
	resp, err := c.makeAPIRequest("polls?broadcaster_id=" + broadcasterID)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var pollResp pollResponse
	if err := json.NewDecoder(resp.Body).Decode(&pollResp); err != nil {
		return err
	}

	info.ActivePolls = make([]Poll, 0, len(pollResp.Data))
	for _, p := range pollResp.Data {
		if p.Status == "ACTIVE" {
			poll := Poll{
				ID:        p.ID,
				Title:     p.Title,
				Status:    p.Status,
				Duration:  p.Duration,
				StartedAt: p.StartedAt,
				EndsAt:    p.StartedAt.Add(time.Duration(p.Duration) * time.Second),
				Choices:   p.Choices,
			}
			info.ActivePolls = append(info.ActivePolls, poll)
		}
	}

	return nil
}

// getPredictions получает активные предсказания
func (c *Client) getPredictions(broadcasterID string, info *StreamerInfo) error {
	resp, err := c.makeAPIRequest("predictions?broadcaster_id=" + broadcasterID)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Здесь нужно реализовать парсинг предсказаний
	// Структура аналогична pollResponse

	return nil
}

// Utility методы для удобства

// IsStreaming проверяет, стримит ли пользователь
func (c *Client) IsStreaming(username string) (bool, error) {
	info, err := c.GetStreamerInfo(username)
	if err != nil {
		return false, err
	}
	return info.IsLive, nil
}

// GetViewerCount возвращает количество зрителей
func (c *Client) GetViewerCount(username string) (int, error) {
	info, err := c.GetStreamerInfo(username)
	if err != nil {
		return 0, err
	}
	return info.ViewerCount, nil
}

// HasActivePolls проверяет, есть ли активные голосования
func (c *Client) HasActivePolls(username string) (bool, error) {
	info, err := c.GetStreamerInfo(username)
	if err != nil {
		return false, err
	}
	return len(info.ActivePolls) > 0, nil
}
