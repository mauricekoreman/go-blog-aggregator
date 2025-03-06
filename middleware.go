package main

import (
	"context"
	"fmt"

	"github.com/mauricekoreman/blog-aggregator/internal/database"
)

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.config.Username)
		if err != nil {
			return fmt.Errorf("error getting user: %w", err)
		}

		return handler(s, cmd, user)
	}
}
