package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/johannesalke/gator/internal/database"
	//"time"
	//"github.com/johannesalke/gator/internal/database"
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

func scrapeFeed(s *state) error {

	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("Error retrieving feeds for fetching: %s", err)
	}

	feed_id := feed.ID
	_, err = s.db.MarkFeedFetched(context.Background(), feed_id)
	if err != nil {
		return fmt.Errorf("Error marking feeds as fetched")
	}
	rssObj, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		return err
	}
	//dbPosts := make([]database.Post,len(rssObj.Channel.Item))
	for _, feedEntry := range rssObj.Channel.Item {
		fmt.Printf("%s\n", feedEntry.Title)

		var publishedAt sql.NullTime
		if t, err := time.Parse(time.RFC1123Z, feedEntry.PubDate); err == nil {
			publishedAt = sql.NullTime{
				Time:  t,
				Valid: true,
			}
		}
		description := sql.NullString{String: feedEntry.Description, Valid: true}

		postParams := database.CreatePostParams{
			ID:          uuid.New(),
			Title:       feedEntry.Title,
			Url:         feedEntry.Link,
			Description: description,
			PublishedAt: publishedAt,
			FeedID:      feed_id,
		}
		_, err := s.db.CreatePost(context.Background(), postParams)
		if err != nil {
			return fmt.Errorf("Error inserting post: %s", err)
		}

	}
	return nil
}
