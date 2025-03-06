package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"

	_ "github.com/lib/pq"
	"github.com/mauricekoreman/blog-aggregator/internal/config"
	"github.com/mauricekoreman/blog-aggregator/internal/database"
)

type state struct {
	db     *database.Queries
	config *config.Config
}

func main() {
	cfg := config.Read()

	db, err := sql.Open("postgres", cfg.DB_URL)
	if err != nil {
		fmt.Println("error opening db connection")
		os.Exit(1)
	}

	dbQueries := database.New(db)
	st := &state{config: &cfg, db: dbQueries}
	commands := &commands{commandNames: make(map[string]func(*state, command) error)}

	commands.register("login", handlerLogin)
	commands.register("register", handleRegister)
	commands.register("reset", handleResetAllUsers)
	commands.register("users", handleGetAllUsers)
	commands.register("agg", handleAgg)
	commands.register("addfeed", middlewareLoggedIn(handleAddFeed))
	commands.register("feeds", handleGetAllFeeds)
	commands.register("follow", middlewareLoggedIn(handleFollowFeed))
	commands.register("following", middlewareLoggedIn(handleGetFeedFollowsForUser))
	commands.register("unfollow", middlewareLoggedIn(handleFeedUnfollow))

	argsWithoutProg := os.Args[1:]

	if len(argsWithoutProg) < 1 {
		fmt.Println("not enough arguments")
		os.Exit(1)
	}

	cmd := command{
		name: argsWithoutProg[0],
		args: argsWithoutProg[1:],
	}

	err = commands.run(st, cmd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

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

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("User-Agent", "gator")

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}

	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	feed := RSSFeed{}
	xml.Unmarshal(data, &feed)

	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)
	feed.Channel.Link = html.UnescapeString(feed.Channel.Link)
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)

	for i := range feed.Channel.Item {
		feed.Channel.Item[i].Description = html.UnescapeString(feed.Channel.Item[i].Description)
		feed.Channel.Item[i].Link = html.UnescapeString(feed.Channel.Item[i].Link)
		feed.Channel.Item[i].PubDate = html.UnescapeString(feed.Channel.Item[i].PubDate)
		feed.Channel.Item[i].Title = html.UnescapeString(feed.Channel.Item[i].Title)
	}

	return &feed, nil
}
