package main

import (
	"fmt"

	"github.com/godovasik/dawgobot/ai/ollama"
	"github.com/godovasik/dawgobot/twitch"
)

func main() {

	// testLlava()
	// testMonitorChat()
	// testCheckUrl()
	// testFindUrl()

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

func testMonitorChat() {
	client, err := twitch.NewClient()
	if err != nil {
		fmt.Println(err)
		return
	}

	twitch.MonitorChannelChat(client)
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

func testLlava() {
	imagePath := "/home/bailey/Downloads/Telegram Desktop/SBKjnxm.jpg"
	response, err := ollama.DescribeImage(imagePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(response.Response, response.TotalDuration)

}
