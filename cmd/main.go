package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/godovasik/dawgobot/internal/ai/deepseek"
	"github.com/godovasik/dawgobot/internal/ai/ollama"
	"github.com/godovasik/dawgobot/internal/ai/openrouter"
	"github.com/godovasik/dawgobot/internal/client"
	database "github.com/godovasik/dawgobot/internal/database"
	"github.com/godovasik/dawgobot/internal/timeline"
	"github.com/godovasik/dawgobot/internal/twitch"
	"github.com/godovasik/dawgobot/logger"
)

// можно переписать с "flags", но мне лень.
func main() {
	if len(os.Args) < 2 {
		fmt.Println("no argument here...")
		// testLlava()
		// testMonitorChat()
		// testCheckUrl()
		// testFindUrl()
		// testScanForImages()
		// testGetImageAndDescribe()
		// testLoadCharacters()
		// testMonitorChat()
		// testSimpleDeep()
		// testTimeline()
		// testMockEvent()
		// testBasicTimeline()
		// testBufferOverflow()
		// testGetLastEvents()
		// testDifferentEventTypes()
		// testChannelStress()
		// testGetRecentEvents()
		// testMonitorAndTimeline()
		// testSqlite()
		// testMonitorChatEvents()

		// testGemini()
		// testRouterAgain()

		testTwitchApi()

		return
	}
	switch os.Args[1] {
	case "log":
		fmt.Println("kek")
	case "img":
		// ReactToImages()
	case "last":
		streamer := ""
		if len(os.Args) >= 3 {
			streamer = os.Args[2]
			fmt.Printf("last events for %s:\n", streamer)
			testGetEvents(streamer)
		} else {
			fmt.Println("last events for ALL:")
			testGetAllEvents()
		}
	case "count":
		streamer := ""
		if len(os.Args) < 3 {
			streamer = "dawgonosik"
		} else {
			streamer = os.Args[2]
		}
		testEventsCount(streamer)
	case "monitor":
		boys := []string{}
		if len(os.Args) < 3 {
			boys = []string{
				"dawgonosik",
				"hak3li",
				"mightypoot",
				"ipoch0__0",
				"timour_j",
				"pixel_bot_o_0",
				"lesnoybol1",
			}
		} else {
			boys = append(boys, os.Args[2])
		}
		fmt.Println("monitoring chat for", boys)
		testMonitorChatEvents(boys...)
	case "images":
		boys := []string{}
		if len(os.Args) < 3 {
			boys = []string{
				"dawgonosik",
				"hak3li",
				"mightypoot",
				"ipoch0__0",
				"timour_j",
				"pixel_bot_o_0",
				"lesnoybol1",
			}
		} else {
			boys = append(boys, os.Args[2])
		}
		fmt.Println("monitoring chat with images for", boys)
		testMonitorChatEventsWithImages(boys...)

	case "replyimg":
		boys := []string{}
		if len(os.Args) < 3 {
			boys = []string{
				"dawgonosik",
				"hak3li",
				"mightypoot",
				"ipoch0__0",
				"timour_j",
				"pixel_bot_o_0",
				"lesnoybol1",
			}
		} else {
			boys = append(boys, os.Args[2])
		}
		testReplyToImages(boys...)
		fmt.Println("mok")
	}
}

func testTwitchApi() {
	twcli, err := twitch.NewClient()
	if err != nil {
		fmt.Println(err)
		return
	}

	data, err := twcli.GetStreamerInfo("silvername")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(data)
	fmt.Println("---")
	fmt.Println(twcli.GetViewerCount("SilverName"))
}

func testGemini() {
	err := openrouter.LoadCharacters()
	if err != nil {
		fmt.Println(err)
		return
	}

	client, err := openrouter.GetNewClient(true)
	if err != nil {
		fmt.Println("getnewclient err,", err)
		return
	}
	url := "https://upload.wikimedia.org/wikipedia/commons/thumb/d/dd/Gfp-wisconsin-madison-the-nature-boardwalk.jpg/2560px-Gfp-wisconsin-madison-the-nature-boardwalk.jpg"
	ctx := context.Background()
	resp, err := client.DescribeImageGemeni(ctx, url)
	if err != nil {
		fmt.Println("describeImage error:", err)
		return
	}
	fmt.Println(resp)
}

func testEventsCount(streamer string) {
	db, err := database.New()
	if err != nil {
		fmt.Println(err)
		return
	}
	count, err := db.GetEventsCountByStreamer(streamer)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(count, "events for", streamer)
}

func testGetAllEvents() {
	db, err := database.New()
	if err != nil {
		fmt.Println(err)
		return
	}
	events, err := db.GetAllEventsByCount(15)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(len(events))
	timeline.PrintEvents(events)
}

func testReplyToImages(channels ...string) {
	tw, err := twitch.NewClient()
	if err != nil {
		logger.Error(err.Error())
		return
	}

	err = openrouter.LoadCharacters()
	if err != nil {
		logger.Error(err.Error())
		return
	}

	err = deepseek.LoadCharacters() // this is ugly
	if err != nil {
		logger.Error(err.Error())
		return
	}

	gmn, err := openrouter.GetNewClient(false)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	ds, err := deepseek.NewClient()
	if err != nil {
		logger.Error(err.Error())
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	client := client.NewClientBuilder().
		WithTwitch(tw).
		WithContext(ctx, cancel).
		WithGemeni(gmn).
		WithDeepseek(ds).
		Build()

	err = client.ReactToImages(channels...)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	err = client.TWClient.TWClient.Connect()
	if err != nil {
		logger.Error(err.Error())
		return
	}
}

func testGetEvents(streamer string) {
	db, err := database.New()
	if err != nil {
		fmt.Println(err)
		return
	}
	events, err := db.GetEventsByCount(streamer, 15)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(len(events))
	timeline.PrintEvents(events)
}

func testMonitorChatEventsWithImages(channels ...string) {
	err := openrouter.LoadCharacters()
	if err != nil {
		fmt.Println(err)
		return
	}

	tw, err := twitch.NewClient()
	if err != nil {
		fmt.Println(err)
		return
	}

	db, err := database.New()
	if err != nil {
		fmt.Println(err)
		return
	}

	gmn, err := openrouter.GetNewClient(false)
	if err != nil {
		fmt.Println(err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	client := client.NewClientBuilder().
		WithDB(db).
		WithTwitch(tw).
		WithContext(ctx, cancel).
		WithGemeni(gmn).
		Build()

	// эта в горутине, тк она блокирующая
	go func() {
		logger.Info("Connecting to Twitch IRC...")
		if err := client.TWClient.TWClient.Connect(); err != nil {
			logger.Errorf("IRC connection error: %v", err)
		}
	}()

	// этот теперть тоже блокирующий - ждет контекста
	go func() {
		if err := client.MonitorChatEvents(true, channels...); err != nil {
			logger.Errorf("Monitoring error: %v", err)
		}
	}()

	// shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down...")

	// Отключаемся от IRC явно
	client.TWClient.TWClient.Disconnect()

	// Отменяем контекст
	cancel()

	time.Sleep(2 * time.Second)
}

func testMonitorChatEvents(channels ...string) {
	tw, err := twitch.NewClient()
	if err != nil {
		fmt.Println(err)
		return
	}

	db, err := database.New()
	if err != nil {
		fmt.Println(err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	client := client.NewClientBuilder().
		WithDB(db).
		WithTwitch(tw).
		WithContext(ctx, cancel).
		Build()

	err = client.MonitorChatEvents(false, channels...)
	if err != nil {
		fmt.Println(err)
		return
	}

	client.TWClient.TWClient.Connect()
}

func testSqlite() {
	db, err := database.New()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()
	events := []timeline.Event{}
	mock := timeline.NewEventMock()
	events = append(events, mock())
	events = append(events, mock())
	events = append(events, mock())

	// db.AddEvents(events)
	three, err := db.GetEventsByCount("forsen", 3)
	one, err := db.GetEventsByCount("forsen", 1)
	if err != nil {
		fmt.Println(err)
	}

	timeline.PrintEvents(one)
	timeline.PrintEvents(three)

	all, err := db.GetEventsByCount("forsen", 100)
	timeline.PrintEvents(all)
}

func testMockEvent() {
	mock := timeline.NewEventMock()
	fmt.Println(mock())
	fmt.Println(mock())
	fmt.Println(mock())
}

func testTimeline() {
	tl := timeline.NewTimeline(3)
	mock := timeline.NewEventMock()

	tl.AddEvent(mock())
	tl.AddEvent(mock())
	tl.AddEvent(mock())
	time.Sleep(1 * time.Second)

	events := tl.GetAllEvents()
	for _, e := range events {
		fmt.Println(e.Content)
	}
}

func testLoadCharacters() {
	openrouter.LoadCharacters()
}

func testGetImageAndDescribe() {
	// u := "https://sun9-28.userapi.com/impg/GW4o3NxSl2hOWrKy2UjFtcrTbqMgGa9ijf3o1Q/V4oxBB1zvco.jpg?size=551x1178&quality=95&sign=92db4357f91dbd160db4a3b20ec72da7&type=album"
	u := "https://previews.123rf.com/images/moovstock/moovstock1803/moovstock180301567/97690294-casino-roulette-wheel-ball-hits-15-fifteen-black-3d-rendering.jpg"
	ok, err := ollama.CheckUrl(u)
	if err != nil {
		logger.Info("cant get url " + u) // this is ugly
	}
	if !ok {
		logger.Info(u + " is not an image")
	}

	data, err := ollama.GetImage(u)
	if err != nil {
		logger.Info("error getting image:" + err.Error())
	}
	resp, err := ollama.DescribeImageBytes(data)
	if err != nil {
		logger.Info("ollama error:" + err.Error())
	}
	fmt.Printf("image url:%s\ndescription: %s\n", u, resp)
}

func testFindUrl() {
	texts := []string{
		"Посмотри https://example.com/image.jpg",
		"Зайди на google.com",
		"www.github.com очень крутой сайт",
		"Нет ссылок в этом тексте",
		"Много ссылок: https://ya.ru и example.org",
	}
	var urls []string
	for _, s := range texts {
		urls = append(urls, twitch.FindURLs(s)...)
	}
	fmt.Println(urls)
}

func testCheckUrl() {
	// url := "https://sun9-28.userapi.com/impg/GW4o3NxSl2hOWrKy2UjFtcrTbqMgGa9ijf3o1Q/V4oxBB1zvco.jpg?size=551x1178&quality=95&sign=92db4357f91dbd160db4a3b20ec72da7&type=album"
	url2 := "google.com"
	// url3 := "as2.ftcdn.net/jpg/05/25/08/09/1000_F_525080936_JEpnKXh2siYKBPpsqd98pbbcIzy4ySKz.webp"
	ok, err := ollama.CheckUrl(url2)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(ok)
}

// func testLlava() {
// 	imagePath := "/home/bailey/Downloads/Telegram Desktop/SBKjnxm.jpg"
// 	response, err := ollama.DescribeImage(imagePath)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	fmt.Println(response.Response, response.TotalDuration)
//
// }
