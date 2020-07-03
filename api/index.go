package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
)

type GetChatResponse struct {
	Ok   bool `json:"ok"`
	Chat Chat `json:"result"`
}

type SendMessageResponse struct {
	Ok     bool              `json:"ok"`
	Result SendMessageResult `json:"result"`
}

type SendMessageResult struct {
	MessageId int `json:"message_id"`
}

type Chat struct {
	PinnedMessage PinnedMessage `json:"pinned_message"`
}

type PinnedMessage struct {
	Date int64  `json:"date"`
	Text string `json:"text"`
}

var (
	telegramBotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	chatId           = os.Getenv("CHAT_ID")
	botURL           = fmt.Sprintf("https://api.telegram.org/bot%s", telegramBotToken)
	client           = http.Client{}
)

func Handler(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	if from == "" {
		from = "3600"
	}

	fromSeconds, err := strconv.Atoi(from)
	if err != nil {
		log.Fatal(err)
	}
	channels := r.URL.Query()["channels"]
	rssURL := fmt.Sprintf("https://lobste.rs/t/%s.rss", strings.Join(channels, ","))

	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL(rssURL)
	items := feed.Items

	for _, item := range items {
		t, _ := time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", item.Published)

		if time.Now().Unix()-int64(fromSeconds) >= t.Unix() {
			continue
		}
		categories := []string{}
		for _, category := range item.Categories {
			categories = append(categories, "%23"+category)
		}
		message := fmt.Sprintf("%s %s", strings.Join(categories, " "), item.Link)
		sendMessageResult, err := sendMessage(message)
		if err != nil {
			log.Fatal(err)
			break
		}
		if sendMessageResult.MessageId == 0 {
			log.Fatal("Failed to send message")
		}
	}
}

func getPinnedMessage() (PinnedMessage, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/getChat?chat_id=@%s", botURL, chatId),
		nil,
	)
	if err != nil {
		return PinnedMessage{}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return PinnedMessage{}, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	getChatResponse := GetChatResponse{}

	json.Unmarshal(body, &getChatResponse)

	return getChatResponse.Chat.PinnedMessage, nil
}

func sendMessage(text string) (SendMessageResult, error) {
	request, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/sendMessage?chat_id=@%s&text=%s", botURL, chatId, text),
		nil,
	)
	if err != nil {
		return SendMessageResult{}, err
	}

	resp, err := client.Do(request)
	if err != nil {
		return SendMessageResult{}, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return SendMessageResult{}, err
	}

	sendMessageResponse := SendMessageResponse{}
	json.Unmarshal(body, &sendMessageResponse)

	return sendMessageResponse.Result, nil
}

func pinMessage(messageId int) error {
	request, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/pinChatMessage?chat_id=@%s&message_id=%d", botURL, chatId, messageId),
		nil,
	)
	if err != nil {
		return err
	}

	resp, err := client.Do(request)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Println(string(body))

	return nil
}
