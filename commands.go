package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
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
		return fmt.Errorf("a feed URL is required")
	}

	feedURL := cmd.args[0]
	feed, err := fetchFeed(context.Background(), feedURL)
	if err != nil {
		return fmt.Errorf("error fetching feed: %w", err)
	}

	fmt.Printf("Feed fetched successfully! Feed details: %v", feed)

	return nil
}

func handleAddFeed(s *state, cmd command) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("a feed name and feed URL are required")
	}

	currentUser, err := getCurrentUser(s)
	if err != nil {
		return err
	}

	feed, err := s.db.AddFeed(context.Background(), database.AddFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
		Url:       cmd.args[1],
		UserID:    currentUser.ID,
	})
	if err != nil {
		return fmt.Errorf("error adding feed: %w", err)
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

// util functions
func getCurrentUser(s *state) (database.User, error) {
	user, err := s.db.GetUser(context.Background(), s.config.Username)
	if err != nil {
		return database.User{}, fmt.Errorf("error getting user: %w", err)
	}

	return user, nil
}
