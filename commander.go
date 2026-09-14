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
	cmd string
	args []string
}

type commands struct {
	cmdMap map[string]func(*state, command) error
}


// commands struct method receivers============================================
func (c *commands) run(s *state, cmd command) error {
	/* your algo here.... */
	return nil
}

func (c *commands) register(name string, f func(*state, command) error) {
	/* your algo here.... */
	return
}

// Command Handlers============================================================
func handlerLogin(s *state, cmd command) error {

	if len(cmd.args) < 1 {
		return errors.New("the login handler expects a single argument, the username")
	}

	if err := s.cfg.SetUser(cmd.args[0]); err != nil {
		return err
	}

	uName := sql.NullString {
		String: cmd.args[0],
		Valid: true,
	}

	_, err := s.db.GetUser(context.Background(), uName)
	if err != nil {
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
	fmt.Println("User", user.Name.String, "registered")

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
