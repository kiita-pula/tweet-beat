package main

import (
	"context"
	"log"
	"strconv"
	"time"

	tele "gopkg.in/tucnak/telebot.v2"
)

type tweetSub struct {
	userID  int
	stopCh  chan struct{}
	tweetCh chan string
}

// Recipient returns user ID. Implements Recipient interface for telegram Send method
func (u *tweetSub) Recipient() string {
	return strconv.Itoa(u.userID)
}

type telebot struct {
	db *db

	api           *tele.Bot
	subscriptions map[int]*tweetSub

	subsCh chan<- *tweetSub
}

func newTelebot(c telegramCfg, db *db, subsCh chan<- *tweetSub) (*telebot, error) {
	log.Println("[telegram-bot] init settings")
	settings := tele.Settings{
		Token:  c.ApiToken,
		Poller: &tele.LongPoller{Timeout: time.Duration(c.Timeout) * time.Second},
		URL:    c.ApiUrl,
	}

	log.Println("[telegram-bot] init api")
	api, err := tele.NewBot(settings)
	if err != nil {
		return nil, err
	}

	return &telebot{
		db:            db,
		subscriptions: make(map[int]*tweetSub),
		api:           api,
		subsCh:        subsCh,
	}, nil
}

func (tb *telebot) start(ctx context.Context) {
	log.Println("[telegram-bot] start: init subscribers")
	users, err := tb.db.fetchSubscribers()
	if err != nil {
		log.Fatal(err)
	}

	for _, u := range users {
		tb.api.Send(&tweetSub{userID: u}, "resubscribed")
		tb.subscribe(ctx, u)
	}
	log.Println("[telegram-bot] resubscribed", len(users))

	tb.api.Handle("/sub", func(m *tele.Message) {
		log.Println("[telegram-bot] received command sub, payload", m.Payload)
		if !m.Private() {
			return
		}

		err := tb.db.addSubscriber(m.Sender.ID)
		if err != nil {
			log.Println(err)
			return
		}

		tb.subscribe(ctx, m.Sender.ID)
		tb.api.Send(m.Sender, "subscribed")
	})

	tb.api.Handle("/stop", func(m *tele.Message) {
		log.Println("[telegram-bot] received command stop")
		if !m.Private() {
			return
		}

		tb.api.Send(m.Sender, "stopping all")
		if v, ok := tb.subscriptions[m.Sender.ID]; ok {
			v.stopCh <- struct{}{}
		}
	})

	tb.api.Start()
}

func (tb *telebot) subscribe(ctx context.Context, ID int) {
	sub := &tweetSub{
		stopCh:  make(chan struct{}, 1),
		userID:  ID,
		tweetCh: make(chan string, 10),
	}
	tb.subscriptions[ID] = sub
	tb.subsCh <- sub
	go func(sub *tweetSub) {
		for {
			select {
			case t := <-sub.tweetCh:
				log.Printf("[telegram-bot] %d send tweet", sub.userID)
				tb.api.Send(sub, t)
			case <-sub.stopCh:
				log.Printf("[telegram-bot] %d stop", sub.userID)
				tb.subsCh <- sub
				return
			case <-ctx.Done():
				log.Println("[telegram-bot] shutdown")
				return
			}
		}
	}(sub)
}
