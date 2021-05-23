package main

import (
	"encoding/json"
	"log"
	"os"
)

type dbCfg struct {
	Name string `json:"db_name"`
	Path string `json:"db_path"`
}

func (c dbCfg) DBPath() string {
	return c.Path + c.Name
}

type telegramCfg struct {
	ApiToken string `json:"api_token"`
	ApiUrl   string `json:"api_url"`

	// defines in seconds
	Timeout int `json:"poller_timeout"`
}

type twitterCfg struct {
	ApiToken string `json:"api_token"`
	ApiUrl   string `json:"api_url"`

	// defined interval that will be substracted from time.Now() - TweetsFetchInterval to fetch tweets
	TweetsFetchInterval int `json:"tweet_fetch_interval"`

	// defines delay before perfoming next fetch
	TweetsFetchDelay int `json:"tweet_fetch_delay"`

	ResultsNum int `json:"result_num"`
}

type config struct {
	DB       dbCfg       `json:db"`
	Telegram telegramCfg `json:telegram"`
	Twitter  twitterCfg  `json:twitter"`
}

func initConfig(path *string) config {
	cfgJson, err := os.ReadFile(*path)
	if err != nil {
		log.Fatal(err)
	}

	cfg := config{}
	err = json.Unmarshal(cfgJson, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	return cfg
}
