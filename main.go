package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	tb "gopkg.in/tucnak/telebot.v2"
)

var msgCounter *prometheus.CounterVec

func main() {
	b, err := tb.NewBot(tb.Settings{
		Token:  os.Getenv("TELEGRAM_BOT_TOKEN"),
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatal(err)
		return
	}
	setupMsgCounter()
	b.Handle(tb.OnText, logMessage)
	b.Handle(tb.OnSticker, logMessage)
	b.Handle(tb.OnPhoto, logMessage)
	b.Handle(tb.OnDocument, logMessage)
	b.Handle(tb.OnVoice, logMessage)
	go b.Start()
	startListener()
}

func logMessage(m *tb.Message) {
	msgType := ""
	senderUsername := m.Sender.Username
	senderID := m.Sender.ID
	groupTitle := m.Chat.Username
	if m.FromGroup() {
		groupTitle = m.Chat.Title
	}
	if m.Text != "" {
		msgType = "TEXT"
	}
	if m.Sticker != nil {
		msgType = "STICKER"
	}
	if m.Photo != nil {
		msgType = "PHOTO"
	}
	if m.Document != nil {
		msgType = "DOCUMENT"
	}
	if m.Voice != nil {
		msgType = "VOICE"
	}
	groupID := m.Chat.ID
	log.Printf("%s(%d)@%s(%d) %s", senderUsername, senderID, groupTitle, groupID, msgType)
	ctr := msgCounter.WithLabelValues(fmt.Sprint(senderID), senderUsername, fmt.Sprint(groupID), groupTitle, msgType)
	ctr.Inc()
}

func setupMsgCounter() {
	chatmsgs := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "telegram_text_messages",
			Help: "standard telegram (text) messages tagged by user and group (or target)",
		},
		[]string{"sender_id", "sender_name", "target_id", "target_name", "message_type"},
	)
	prometheus.MustRegister(chatmsgs)
	msgCounter = chatmsgs
}

func startListener() {
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe("0.0.0.0:3000", nil))
}
