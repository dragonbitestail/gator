package main

import (
		"log"
		"os"
		"database/sql"
		_ "github.com/lib/pq"
		"gator/pkg/database"
		gtc "gator/pkg/config"
)

type state struct {
	db *database.Queries
	cfg *gtc.Config
}

var cmdEntered string
var argsEntered []string

func init(){
	log.Printf("%+v\n", os.Args)

	if len(os.Args) < 2 {
		log.Fatal("No command given")
	}

	cmdEntered = os.Args[1]
	if len(os.Args) > 2 {
		argsEntered = os.Args[2:]
		log.Printf("Parsed args into argsEntered slice %s\n", argsEntered)
	}

}

func main() {
	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5432/gator")
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)

	var cState state
	var appCmds = commands {
		cmdMap: make(map[string]func(*state, command) error),
	}

	appCmds.register("login", handlerLogin)
	appCmds.register("register", handlerRegister)
	appCmds.register("reset", handlerDeleteUsers)
	appCmds.register("users", handlerGetUsers)
	appCmds.register("agg", handlerAgg)
	appCmds.register("addfeed", handlerAddFeed)
	appCmds.register("feeds", handlerGetFeeds)
	appCmds.register("follow", handlerFollow)
	appCmds.register("following", handlerGetFeedFollowsForUser)
	log.Printf("%+v\n", appCmds)


	cState.cfg = readConfig()
	cState.db = dbQueries

	cmd := command {
		name: cmdEntered,
		args: argsEntered,
	}

	if err := appCmds.run(&cState, cmd); err != nil {
		log.Fatal(err)
	}

}

func readConfig() *gtc.Config {
	config, err := gtc.Read()
	if err != nil {
		log.Fatal(err)
	}

	return config
}
