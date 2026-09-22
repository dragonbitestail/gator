package main

import (
		"log"
		"os"
		"database/sql"
		_ "github.com/lib/pq"
		"gator/pkg/database"
		gtc "gator/pkg/config"
		"gator/pkg/logging"
)

type state struct {
	db *database.Queries
	cfg *gtc.Config
}

var cmdEntered string
var argsEntered []string

var logr = ilogger.GetLogger()

func init(){
	logr.Debug("init()", "args",  os.Args)

	if len(os.Args) < 2 {
		log.Fatal("No command given")
	}

	cmdEntered = os.Args[1]
	if len(os.Args) > 2 {
		argsEntered = os.Args[2:]
		logr.Debug("init() saved args", "argsEntered", argsEntered)
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

	appCmds.register(loginC, middlewareLoggedIn(handlerLogin))
	appCmds.register(registerC, handlerRegister)
	appCmds.register(resetC, handlerDeleteUsers)
	appCmds.register(usersC, handlerGetUsers)
	appCmds.register(aggC, handlerAgg)
	appCmds.register(addfeedC, handlerAddFeed)
	appCmds.register(feedsC, handlerGetFeeds)
	appCmds.register(followC, handlerFollow)
	appCmds.register("following", middlewareLoggedIn(handlerGetFeedFollowsForUser))
	appCmds.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	logr.Debug("main()", "appCmds", appCmds)


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
