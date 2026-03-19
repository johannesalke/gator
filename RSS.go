package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func (obj *RSSItem) EscapeRSSItem() { //This function deals with escaping strings on rss objects.
	obj.Title = html.UnescapeString(obj.Title)
	obj.Description = html.UnescapeString(obj.Description)
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {

	client := http.DefaultClient
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("Error creating feed request: %s", err)
	}
	req.Header.Set("User-Agent", "gator")

	res, err := client.Do(req) //|Send request|//
	if err != nil {
		return nil, fmt.Errorf("Error requesting feed: %s", err)
	}
	var rssBytes []byte //|read response to bytecode|//
	rssBytes, err = io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response body: %s", err)
	}
	var rssFeed RSSFeed //|Unmarshall to struct|//
	err = xml.Unmarshal(rssBytes, &rssFeed)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling xml: %s", err)
	}

	rssFeed.Channel.Title = html.UnescapeString(rssFeed.Channel.Title)
	rssFeed.Channel.Description = html.UnescapeString(rssFeed.Channel.Description)
	for i, _ := range rssFeed.Channel.Item {
		rssFeed.Channel.Item[i].EscapeRSSItem()
	}
	return &rssFeed, err
}
