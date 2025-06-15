package main

import (
	"fmt"
	"time"

	"github.com/godovasik/dawgobot/internal/ai/deepseek"
	"github.com/godovasik/dawgobot/internal/ai/ollama"
	"github.com/godovasik/dawgobot/internal/ai/openrouter"
	"github.com/godovasik/dawgobot/internal/timeline"
	"github.com/godovasik/dawgobot/internal/twitch"
	"github.com/godovasik/dawgobot/logger"
)

func main() {

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

	testMonitorAndTimeline()
}

func testMonitorAndTimeline() {
	tl := timeline.NewTimeline(100)
	defer tl.Stop()

	tw, err := twitch.NewClient(tl)
	if err != nil {
		fmt.Println(err)
		return
	}

	ds, err := deepseek.NewClient(tl)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = deepseek.LoadCharacters()
	if err != nil {
		fmt.Println(err)
		return
	}

	tw.MonitorChannelChat("timour_j")

	go func() { //yourself
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				events := tl.GetRecentEvents(60 * time.Second)
				if len(events) == 0 {
					logger.Info("no new events, skip")
				} else {
					logger.Infof("new events:%d", len(events))
					logger.Infof("Отправляем:%s", timeline.SprintEvents(events))
					logger.Info("ждем ответ дипсика...")

					resp, err := ds.GetResponse("dawgobot", timeline.SprintEvents(events))
					if err != nil {
						logger.Info(err.Error())
					}
					fmt.Println("from deepseek:", resp)
				}
			}
		}

	}()

	if err := tw.TwitchClient.Connect(); err != nil {
		fmt.Println(err)
		return
	}
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

// func testMonitorChat() {
// 	client, err := twitch.NewClient()
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
//
// 	username := "forsen"
// 	twitch.MonitorChannelChat(client, username)
// }

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
