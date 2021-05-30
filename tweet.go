package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/g8rswimmer/go-twitter"
)

type authorize struct {
	Token string
}

func (a authorize) Add(req *http.Request) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", a.Token))
}

type tweet struct {
	tweetApi      *twitter.Tweet
	fetchInterval int
	fetchDelay    int
	resultsNum    int
	accouts       []string

	subscriptions map[int]*tweetSub
}

func newTweetBeat(c twitterCfg) *tweet {
	t := tweet{
		subscriptions: make(map[int]*tweetSub),
		accouts:       nil,
		fetchInterval: c.TweetsFetchInterval,
		fetchDelay:    c.TweetsFetchDelay,
		resultsNum:    c.ResultsNum,
	}
	tweetApi := &twitter.Tweet{
		Authorizer: authorize{
			Token: c.ApiToken,
		},
		Client: http.DefaultClient,
		Host:   c.ApiUrl,
	}
	t.tweetApi = tweetApi

	return &t
}

func (tw *tweet) start(ctx context.Context) chan *tweetSub {
	query := "(from:elonmusk) -is:retweet"
	fieldOpts := twitter.TweetFieldOptions{
		TweetFields: []twitter.TweetField{twitter.TweetFieldID, twitter.TweetFieldText, twitter.TweetFieldCreatedAt},
	}

	incomingSubs := make(chan *tweetSub, 1)
	go func() {
		log.Println("[tweet] starting subscriptions receiver")
		for {
			select {
			case <-ctx.Done():
				log.Println("[tweet] shutdown subscriptions receiver")
				return
			case s := <-incomingSubs:
				log.Printf("[tweet] received subscripion %d", s.userID)
				if _, ok := tw.subscriptions[s.userID]; ok {
					log.Println("[tweet] subscripion exists: unsubscribe")
					tw.subscriptions[s.userID] = nil
					continue
				}

				tw.subscriptions[s.userID] = s
			}
		}
	}()

	go func() {
		log.Println("[tweet] starting tweet fetcher")
		ticker := time.NewTicker(time.Duration(tw.fetchDelay) * time.Second)
		for {
			select {
			case <-ctx.Done():
				log.Println("[tweet] shutdown tweet fetcher")
				return
			case <-ticker.C:
				if len(tw.subscriptions) == 0 {
					log.Println("[tweet] no subscriptions: skipping")
					continue
				}

				from := time.Now().Add(-time.Second * time.Duration(tw.fetchInterval))
				searchOpts := twitter.TweetRecentSearchOptions{MaxResult: tw.resultsNum, StartTime: from}

				recentSearch, err := tw.tweetApi.RecentSearch(ctx, query, searchOpts, fieldOpts)
				var tweetErr *twitter.TweetErrorResponse
				switch {
				case errors.As(err, &tweetErr):
					log.Println(tweetErr.Detail)
				case err != nil:
					log.Println(err.Error())
				default:
					//log.Printf("[tweet] found %d new tweets", len(recentSearch.LookUps))
					if len(recentSearch.LookUps) == 0 {
						//log.Println("[tweet] no new tweets found")
						continue
					}

					log.Println("[tweet] sending tweets")
					var newTweets string
					for _, lookup := range recentSearch.LookUps {
						newTweets = newTweets + fmt.Sprintf("%s\n\n", lookup.Tweet.Text)

					}

					for _, s := range tw.subscriptions {
						if s == nil {
							continue
						}
						s.tweetCh <- newTweets
					}
				}
			}
		}
	}()

	return incomingSubs
}
