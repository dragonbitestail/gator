package main

import (
		"log"
		"os"
		"database/sql"
		_ "github.com/lib/pq"
		"github.com/dragonbitestail/gator/pkg/database"
		gtc "github.com/dragonbitestail/gator/pkg/config"
		"github.com/dragonbitestail/gator/pkg/logging"
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
		log.Fatal("No command given. Try: gator help")
	}

	cmdEntered = os.Args[1]
	if len(os.Args) > 2 {
		argsEntered = os.Args[2:]
		logr.Debug("init() saved args", "argsEntered", argsEntered)
	}

}

func main() {
	var cState state

	cState.cfg = readConfig()
	db, err := sql.Open("postgres", cState.cfg.DbURL)
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)

	cState.db = dbQueries

	var appCmds = commands {
		cmdMap: make(map[string]func(*state, command) error),
		help: make(map[string]string),
	}

	appCmds.register(registerC, handlerRegister)
	appCmds.register(loginC, middlewareLoggedIn(handlerLogin))
	appCmds.register(usersC, handlerGetUsers)
	appCmds.register(aggC, handlerAgg)
	appCmds.register(addfeedC, handlerAddFeed)
	appCmds.register(feedsC, handlerGetFeeds)
	appCmds.register(followC, handlerFollow)
	appCmds.register(followingC, middlewareLoggedIn(handlerGetFeedFollowsForUser))
	appCmds.register(unfollowC, middlewareLoggedIn(handlerUnfollow))
	appCmds.register(browseC, middlewareLoggedIn(handlerBrowse))
	appCmds.register(resetC, handlerDeleteUsers)
	appCmds.register(helpC, middlewareHelp(appCmds, handlerHelp)) // MUST GO LAST TO CONTAIN ALL REGISERED COMMANDS
	logr.Debug("main()", "appCmds", appCmds)

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
