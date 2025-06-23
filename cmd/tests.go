package main

import (
	"fmt"
	"time"

	"github.com/godovasik/dawgobot/internal/timeline"
)

func testBasicTimeline() {
	fmt.Println("=== Test Basic Timeline ===")
	tl := timeline.NewTimeline(10)
	mock := timeline.NewEventMock()

	tl.AddEvent(mock())
	tl.AddEvent(mock())
	tl.AddEvent(mock())

	// Даем время горутине обработать события
	time.Sleep(100 * time.Millisecond)

	events := tl.GetAllEvents()
	fmt.Printf("Events count: %d\n", len(events))
	for i, event := range events {
		fmt.Printf("Event %d: %+v\n", i, event)
	}
	fmt.Println()
}

// Тест переполнения буфера
func testBufferOverflow() {
	fmt.Println("=== Test Buffer Overflow ===")
	tl := timeline.NewTimeline(3) // маленький буфер
	mock := timeline.NewEventMock()

	// Добавляем больше событий чем размер буфера
	for i := 0; i < 5; i++ {
		tl.AddEvent(mock())
	}

	time.Sleep(100 * time.Millisecond)

	events := tl.GetAllEvents()
	fmt.Printf("Buffer size: 3, added 5 events, got: %d events\n", len(events))
	for i, event := range events {
		fmt.Printf("Event %d: content=%s, author=%s\n", i, event.Content, event.Author)
	}
	fmt.Println()
}

// Тест GetLastEvents
func testGetLastEvents() {
	fmt.Println("=== Test GetLastEvents ===")
	tl := timeline.NewTimeline(10)
	mock := timeline.NewEventMock()

	// Добавляем 7 событий
	for i := 0; i < 7; i++ {
		tl.AddEvent(mock())
	}

	time.Sleep(100 * time.Millisecond)

	last3 := tl.GetLastEvents(3)
	fmt.Printf("Last 3 events from 7 total:\n")
	for i, event := range last3 {
		fmt.Printf("Event %d: content=%s\n", i, event.Content)
	}

	last10 := tl.GetLastEvents(10) // больше чем есть
	fmt.Printf("Requested 10, got: %d events\n", len(last10))
	fmt.Println()
}

// Тест с разными типами событий
func testDifferentEventTypes() {
	fmt.Println("=== Test Different Event Types ===")
	tl := timeline.NewTimeline(10)

	// Создаем события разных типов
	chatEvent := timeline.Event{
		Type:      timeline.EventChat,
		Content:   "Привет чат!",
		Author:    "streamer123",
		Timestamp: time.Now(),
	}

	speechEvent := timeline.Event{
		Type:      timeline.EventSpeech,
		Content:   "Стример что-то говорит",
		Timestamp: time.Now(),
	}

	screenshotEvent := timeline.Event{
		Type:      timeline.EventScreenshot,
		Content:   "Описание скриншота",
		Timestamp: time.Now(),
	}

	tl.AddEvent(chatEvent)
	tl.AddEvent(speechEvent)
	tl.AddEvent(screenshotEvent)

	time.Sleep(100 * time.Millisecond)

	events := tl.GetAllEvents()
	for i, event := range events {
		fmt.Printf("Event %d: type=%d, content=%s, author=%s\n",
			i, event.Type, event.Content, event.Author)
	}
	fmt.Println()
}

// Тест GetRecentEvents (этот может не работать из-за бага с мьютексом)
func testGetRecentEvents() {
	fmt.Println("=== Test GetRecentEvents ===")
	tl := timeline.NewTimeline(10)

	// Добавляем событие с правильным timestamp
	oldEvent := timeline.Event{
		Type:      timeline.EventChat,
		Content:   "Старое событие",
		Author:    "user1",
		Timestamp: time.Now().Add(-10 * time.Minute), // 10 минут назад
	}

	newEvent := timeline.Event{
		Type:      timeline.EventChat,
		Content:   "Новое событие",
		Author:    "user2",
		Timestamp: time.Now(),
	}

	tl.AddEvent(oldEvent)
	tl.AddEvent(newEvent)

	time.Sleep(100 * time.Millisecond)

	// Этот тест может зависнуть из-за дедлока в GetRecent!
	fmt.Println("Trying GetRecentEvents (may hang due to deadlock)...")
	recent := tl.GetRecentEvents(5 * time.Minute)
	fmt.Printf("Recent events (last 5 min): %d\n", len(recent))
	fmt.Println()
}

// Стресс-тест канала
func testChannelStress() {
	fmt.Println("=== Test Channel Stress ===")
	tl := timeline.NewTimeline(5)
	mock := timeline.NewEventMock()

	// Быстро добавляем много событий
	for i := 0; i < 150; i++ { // больше чем размер канала (100)
		tl.AddEvent(mock())
	}

	time.Sleep(200 * time.Millisecond)

	events := tl.GetAllEvents()
	fmt.Printf("Added 150 events rapidly, buffer has: %d events\n", len(events))
	fmt.Println()
}

// func ReactToImages() {
// 	deepseek.LoadCharacters()
// 	tc, err := twitch.NewClient(nil)
// 	err = tc.ReactToImages("lesnoybol1")
// 	err = tc.TWClient.Connect()
// 	fmt.Println("ХУЙ:", err)
//
// }

// func testSimpleDeep() {
// 	client, err := deepseek.NewClient()
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
//
// 	err = deepseek.LoadCharacters()
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	messages := `
// // [15-06-25 13:11:41] FiSHB0NE__: TriHard
// // [15-06-25 13:11:42] ThePositiveBot: [Minigame] AUTOMATIC UNSCRAMBLE! PogChamp The first person to unscramble geremm wins 1 cookie! OpieOP
// // [15-06-25 13:12:49] zyrwoot: Aware forsen was on epstein island
// // [15-06-25 13:13:13] djfors_: docJAM now playing: Top 10 Best Restaurants to Visit in Limassol | Cyp[...]
// // [15-06-25 13:13:43] TwoLetterName: Aware
// // [15-06-25 13:13:54] THIZZBOX707: Aware
// // `
// 	err = deepseek.GetResponse(client, "dawgobot", messages)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// }

// func testScanForImages() {
// 	client, err := twitch.NewClient()
// 	if err != nil {
// 		fmt.Println("fuck you")
// 	}
// 	client.OnPrivateMessage(twitch.ScanForImagesHandler())
// 	client.Join("lesnoybol1")
// 	err = client.Connect()
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
//
// }

// эта функция мониторит твич чат, и раз в 15 секунд отправляет сообщение
// с учетом предыдущих, за 60 секунд.
// TODO: отслеживать чат дольше, мб отслеживать конкретный диалог с тем, кого тегнули, но это уже потом
// func testMonitorAndTimeline() {
// 	tl := timeline.NewTimeline(100)
// 	defer tl.Stop()
//
// 	tw, err := twitch.NewClient(tl)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
//
// 	ds, err := deepseek.NewClient(tl)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
//
// 	err = deepseek.LoadCharacters()
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
//
// 	tw.MonitorChannelChat("thijs")
//
// 	go func() { //yourself
// 		ticker := time.NewTicker(15 * time.Second)
// 		defer ticker.Stop()
// 		for {
// 			select {
// 			case <-ticker.C:
// 				events := tl.GetRecentEvents(60 * time.Second)
// 				if len(events) == 0 {
// 					logger.Info("no new events, skip")
// 				} else {
// 					logger.Infof("new events:%d", len(events))
// 					logger.Infof("Отправляем:%s", timeline.SprintEvents(events))
// 					logger.Info("ждем ответ дипсика...")
//
// 					resp, err := ds.GetResponse("dawgobot", timeline.SprintEvents(events))
// 					if err != nil {
// 						logger.Info(err.Error())
// 					}
// 					fmt.Println("from deepseek:", resp)
// 				}
// 			}
// 		}
//
// 	}()
//
// 	if err := tw.TWClient.Connect(); err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// }
