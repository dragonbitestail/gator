package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"
	"github.com/google/uuid"
	"gator/pkg/database"
)

type command struct {
	name string
	args []string
}

type commands struct {
	cmdMap map[string]func(*state, command) error
}


// commands struct method receivers============================================

func (c *commands) register(name string, f func(*state, command) error) {
	c.cmdMap[name] = f
	log.Println("register() func handler for", name)
	return
}

func (c *commands) run(s *state, cmd command) error {
	log.Println("run() attempting to call func handler for", cmd.name)
	f, ok := c.cmdMap[cmd.name]
	if !ok {
		return fmt.Errorf("Unknown command: %s", cmd.name)
	}

	return f(s, cmd)
}


// Command Handlers============================================================

func handlerLogin(s *state, cmd command) error {

	if len(cmd.args) < 1 {
		return errors.New("handlerLogin() expects a single argument, the username")
	}

	uName := sql.NullString {
		String: cmd.args[0],
		Valid: true,
	}

	_, err := s.db.GetUser(context.Background(), uName)
	if err != nil {
		fmt.Printf("Could not login user \"%s\". Not registered?\n", uName.String)
		return err
	}

	if err := s.cfg.SetUser(cmd.args[0]); err != nil {
		return err
	}

	return nil
}

func handlerRegister(s *state, cmd command) error {

	if len(cmd.args) < 1 {
		return errors.New("the register handler expects a single argument, the new username")
	}

	if err := s.cfg.SetUser(cmd.args[0]); err != nil {
		return err
	}

	uParams := database.CreateUserParams{
		ID: uuid.New(),
		CreatedAt: sql.NullTime{
			Time: time.Now(), Valid: true,
		},
		UpdatedAt:  sql.NullTime{
			Time: time.Now(), Valid: true,
		},
		Name: sql.NullString {
			String: cmd.args[0],
			Valid: true,
		},
	}

	user, err := s.db.CreateUser(context.Background(), uParams)
	if err != nil {
		return err
	}

	s.cfg.CurrentUserName = user.Name.String
	fmt.Printf("User \"%s\" registered\n", user.Name.String)

	log.Println("handlerRegister(): registered", user.Name.String)
	return nil
}

func handlerDeleteUsers(s *state, cmd command) error {

	err := s.db.DeleteUsers(context.Background())
	if err != nil {
		return err
	}
	fmt.Println("users table reset. All users deleted")
	return nil
}

func handlerGetUsers(s *state, cmd command) error {

	r, err := s.db.GetUsers(context.Background())
	if err != nil {
		fmt.Printf("Could get user list")
		return err
	}

	for _, row := range r {
		fmt.Printf("* %s", row.Name.String)
		if row.Name.String == s.cfg.CurrentUserName {
			fmt.Printf(" (current)")
		}
		fmt.Println()
	}

	return nil
}
