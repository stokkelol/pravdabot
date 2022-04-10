package cmd

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/dgraph-io/badger/v3"
	tg "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/spf13/viper"
)

const dbPath = "/tmp/badger/"
const link = "https://www.pravda.com.ua"
const lastID = "id"
const selector = "body > div.main_content > div.container_middle.layout_main > div.container_sub_news > div.container_sub_news_wrapper > div:nth-child(1) > div.article_header > a"

// Client struct
type Client struct {
	client            *http.Client
	db                *badger.DB
	tgClient          *tg.BotAPI
	telegramChannelID int64
	lastID            string
}

func New() (*Client, error) {
	var err error
	c := &Client{}
	c.client = &http.Client{}
	c.db, err = badger.Open(badger.DefaultOptions(dbPath))
	if err != nil {
		return nil, err
	}

	c.telegramChannelID = viper.GetInt64("TELEGRAM_CHAT_ID")
	c.tgClient, err = tg.NewBotAPI(viper.GetString("TELEGRAM_BOT_KEY"))
	if err != nil {
		return nil, fmt.Errorf("error creating bot api: %s", err.Error())
	}

	return c, nil
}

func (c *Client) Run() {
	if err := c.compareLastID(); err != nil {
		log.Fatalf("Error getting last id: %s", err.Error())
	}

	defer func() {
		_ = c.db.Close()
	}()

	ticker := time.NewTicker(time.Second * 120)
	defer ticker.Stop()

	c.run()

	for range ticker.C {
		c.run()
	}
}

func (c *Client) compareLastID() error {
	txn := c.db.NewTransaction(true)
	item, err := txn.Get([]byte(lastID))
	if err != nil && err.Error() != "Key not found" {
		return err
	}

	if item == nil {
		return nil
	}

	c.lastID = item.String()

	return nil
}

func (c *Client) run() {
	resp, err := c.client.Get(link)
	if err != nil {
		log.Fatalf("error reading page: %s", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("wrong status code: %d", resp.StatusCode)
	}

	message, err := c.parse(resp.Body)
	if err != nil {
		log.Fatalf("Something went wrong parsing page: %s", err.Error())
	}

	message = link + message

	if c.lastID != message {
		_, err = c.tgClient.Send(tg.MessageConfig{
			BaseChat: tg.BaseChat{
				ChatID: c.telegramChannelID,
			},
			Text: message,
		})

		c.lastID = message
	}
}

func (c *Client) parse(body io.ReadCloser) (string, error) {
	txt, err := ioutil.ReadAll(body)
	if err != nil {
		log.Fatalf(err.Error())
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(txt))
	if err != nil {
		log.Fatal(err.Error())
	}

	href, _ := doc.Find(selector).First().Attr("href")

	return href, nil
}
