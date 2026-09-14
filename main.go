package main

import (
//    "fmt"
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
	var appCmds = commands{
		cmdMap: map[string]func(*state, command) error{
			"login": handlerLogin,
			"register": handlerRegister,
			"reset": handlerDeleteUsers,
		},
	}


	log.Printf("%+v\n", appCmds)


	cState.cfg = readConfig()
	cState.db = dbQueries

	//log.Printf("%+v\n", cState.cfg)

	cmd := command {
		cmd: cmdEntered,
		args: argsEntered,
	}

	f, ok := appCmds.cmdMap[cmdEntered]
	if !ok {
		log.Fatal("given command \"", cmdEntered, "\" not found")
	}
	if err := f(&cState, cmd); err != nil {
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
