package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/mauricekoreman/blog-aggregator/internal/database"
)

type command struct {
	name string
	args []string
}

type commands struct {
	commandNames map[string]func(*state, command) error
}

func (c *commands) register(name string, f func(*state, command) error) {
	// method registers new handler functions for a command name
	c.commandNames[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	// method runs the handler function for the given command name
	handler, ok := c.commandNames[cmd.name]
	if !ok {
		return fmt.Errorf("unknown command: %s", cmd.name)
	}

	err := handler(s, cmd)
	if err != nil {
		return err
	}

	return nil
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("a username is required")
	}

	username := cmd.args[0]
	user, err := s.db.GetUser(context.Background(), username)
	if err != nil {
		return fmt.Errorf("error getting user: %w", err)
	}

	err = s.config.SetUser(user.Name)
	if err != nil {
		return fmt.Errorf("error setting user: %w", err)
	}

	fmt.Printf("User %s set successfully!", username)

	return nil
}

func handleRegister(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("a username is required")
	}

	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
	})

	if err != nil {
		return fmt.Errorf("error creating user: %w", err)
	}

	err = s.config.SetUser(user.Name)
	if err != nil {
		return fmt.Errorf("error setting user: %w", err)
	}

	fmt.Printf("User created successfully! User details: %v", user)

	return nil
}

func handleResetAllUsers(s *state, cmd command) error {
	err := s.db.DeleteAllUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error deleting all users: %w", err)
	}

	fmt.Println("All users deleted successfully!")

	return nil
}

func handleGetAllUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error getting all users: %w", err)
	}

	currentUser := s.config.Username

	fmt.Println("All users:")
	for _, user := range users {
		if currentUser == user.Name {
			fmt.Printf("\n* %s (current)", user.Name)
			continue
		}
		fmt.Printf("\n* %s", user.Name)
	}

	return nil
}

func handleAgg(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("a time_between_req is required")
	}

	timeBetweenRequests, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return fmt.Errorf("error parsing duration: %w", err)
	}

	fmt.Printf("Collection feeds every %v", timeBetweenRequests)
	ticker := time.NewTicker(timeBetweenRequests)

	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}

func handleAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("a feed name and feed URL are required")
	}

	feed, err := s.db.AddFeed(context.Background(), database.AddFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
		Url:       cmd.args[1],
		UserID:    user.ID,
	})
	if err != nil {
		return fmt.Errorf("error adding feed: %w", err)
	}

	_, err = followFeed(s, user.ID, feed.ID)
	if err != nil {
		return err
	}

	fmt.Printf("Feed added successfully! Feed details: %v", feed)

	return nil
}

func handleGetAllFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetAllFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("error getting all feeds: %w", err)
	}

	for _, feed := range feeds {
		fmt.Printf("\n* %s (%s) - %s", feed.Name, feed.Url, feed.UserName)
	}

	return nil
}

func handleFollowFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("a feed URL is required")
	}

	feedURL := cmd.args[0]
	feed, err := s.db.GetFeedByURL(context.Background(), feedURL)
	if err != nil {
		return fmt.Errorf("error getting feed by URL: %w", err)
	}

	feedFollowDetails, err := followFeed(s, user.ID, feed.ID)
	if err != nil {
		return err
	}

	fmt.Printf("Feed followed successfully! Follow details: %v", feedFollowDetails)

	return nil
}

func handleGetFeedFollowsForUser(s *state, cmd command, user database.User) error {
	feedFollows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("error getting feed follows for user: %w", err)
	}

	for _, feedFollow := range feedFollows {
		fmt.Printf("\n* %s (%s)", feedFollow.FeedName, feedFollow.UserName)
	}

	return nil
}

func handleFeedUnfollow(s *state, cmd command, user database.User) error {
	feedURL := cmd.args[0]
	feed, err := s.db.GetFeedByURL(context.Background(), feedURL)
	if err != nil {
		return fmt.Errorf("error getting feed by URL: %w", err)
	}

	err = s.db.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("error deleteting feed-follow by URL: %w", err)
	}

	return nil
}

func handleBrowse(s *state, cmd command, user database.User) error {
	limit := 2
	if len(cmd.args) > 0 {
		newLimit, err := strconv.Atoi(cmd.args[0])
		if err != nil {
			return fmt.Errorf("not a valid limit")
		}

		limit = newLimit
	}

	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		ID:    user.ID,
		Limit: int32(limit),
	})
	if err != nil {
		return fmt.Errorf("error getting posts for user: %w", err)
	}

	for _, post := range posts {
		fmt.Printf("\n* %s (%s)", post.Title, post.Url)
	}

	return nil
}

// utils
func followFeed(s *state, userId uuid.UUID, feedId uuid.UUID) (database.CreateFeedFollowRow, error) {
	feedFollowDetails, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    userId,
		FeedID:    feedId,
	})
	if err != nil {
		return database.CreateFeedFollowRow{}, fmt.Errorf("error following feed: %w", err)
	}

	return feedFollowDetails, nil
}

func scrapeFeeds(s *state) error {
	nextFeed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get the next feed to fetch: %w", err)
	}

	err = s.db.MarkFeedFetched(context.Background(), database.MarkFeedFetchedParams{
		ID: nextFeed.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to mark the next feed as fetched: %w", err)
	}

	feed, err := fetchFeed(context.Background(), nextFeed.Url)
	if err != nil {
		return fmt.Errorf("error fetching feed: %w", err)
	}

	for _, item := range feed.Channel.Item {
		var parsedTime time.Time
		parsedTime, _ = time.Parse(item.PubDate, "2025-03-06 17:51:17.095801")

		err = s.db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       item.Title,
			Description: sql.NullString{String: item.Description, Valid: item.Description != ""},
			Url:         item.Link,
			PublishedAt: sql.NullTime{Time: parsedTime, Valid: parsedTime != time.Time{}},
			FeedID:      nextFeed.ID,
		})

		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" {
				continue
			} else {
				fmt.Println("Postgres error code: " + pqErr.Code)
			}
		} else {
			fmt.Println("erorr occurred creating post: " + err.Error())
		}
	}

	return nil
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
