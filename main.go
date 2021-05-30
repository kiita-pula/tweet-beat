package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	path := flag.String("c", "config.json", "json config file path")
	flag.Parse()

	cfg := initConfig(path)

	db := initDatabase(cfg.DB)
	defer func() {
		err := db.conn.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		oscall := <-c
		log.Printf("system call:%+v", oscall)
		cancel()
		time.Sleep(100 * time.Millisecond)
	}()

	tweatBeat := newTweetBeat(cfg.Twitter)
	tweetBeatCh := tweatBeat.start(ctx)

	bot, err := newTelebot(cfg.Telegram, db, tweetBeatCh)
	if err != nil {
		log.Fatal(err)
	}

	go bot.start(ctx)

	log.Println("[main] all systems nominal")
	<-ctx.Done()
}
