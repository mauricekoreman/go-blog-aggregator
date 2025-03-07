package main

import (
	"database/sql"
	"fmt"
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
	defer db.Close()

	dbQueries := database.New(db)
	st := &state{config: &cfg, db: dbQueries}
	cmds := commands{commandNames: make(map[string]func(*state, command) error)}

	cmds.register("login", handlerLogin)
	cmds.register("register", handleRegister)
	cmds.register("reset", handleResetAllUsers)
	cmds.register("users", handleGetAllUsers)
	cmds.register("agg", handleAgg)
	cmds.register("addfeed", middlewareLoggedIn(handleAddFeed))
	cmds.register("feeds", handleGetAllFeeds)
	cmds.register("follow", middlewareLoggedIn(handleFollowFeed))
	cmds.register("following", middlewareLoggedIn(handleGetFeedFollowsForUser))
	cmds.register("unfollow", middlewareLoggedIn(handleFeedUnfollow))
	cmds.register("browse", middlewareLoggedIn(handleBrowse))

	argsWithoutProg := os.Args[1:]

	if len(argsWithoutProg) < 1 {
		fmt.Println("not enough arguments")
		os.Exit(1)
	}

	cmd := command{
		name: argsWithoutProg[0],
		args: argsWithoutProg[1:],
	}

	err = cmds.run(st, cmd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
