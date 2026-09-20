package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"
	"net/url"
	"github.com/google/uuid"

	"gator/pkg/database"
)

const (
	loginC = "login"
	registerC = "register"
	resetC = "reset"
	usersC = "users"
	aggC = "agg"
	addfeedC = "addfeed"
	feedsC = "feeds"
	followC = "follow"
	followingC = "following"
	unfollowC = "unfollow"
)

type command struct {
	name string
	args []string
}

type commands struct {
	cmdMap map[string]func(*state, command) error
}


func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {

	fCmdHandler := func(s *state, cmd command) error {
		log.Printf("middlewareLoggedIn() w/ command: %s, args: %v\n", cmd.name, cmd.args)
		user := s.cfg.CurrentUserName
		if cmd.name == loginC && len(cmd.args) == 1 {
			user = cmd.args[0]
		}

		uName := sql.NullString {
			String: user,
			Valid: true,
		}

		u, err := s.db.GetUser(context.Background(), uName)
		if err != nil {
			fmt.Printf("Could not login user \"%s\". Not registered?\n", uName.String)
			return err
		}
		return handler(s, cmd, u)
	}

	return fCmdHandler
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
func handlerUnfollow(s *state, cmd command, user database.User) error {
	log.Printf("handlerUnfollow() cmd: %s, args: %s\n", cmd.name, cmd.args)
	if len(cmd.args) < 1 {
		return errors.New("handlerUnfollow() expects a single argument, the feed url to unfollow.")
	}

	uURL := sql.NullString {
		String: cmd.args[0],
		Valid: true,
	}

	f, err := s.db.GetFeed(context.Background(), uURL)
	if err != nil {
		fmt.Printf("Could not get url \"%s\" to unfollow. Not added with addfeed?\n", uURL.String)
		return err
	}

	uParams := database.DeleteFeedFollowForUserIdParams {
		UserID: uuid.NullUUID {
				UUID: user.ID,
				Valid: true,
		},
		FeedID: uuid.NullUUID {
				UUID: f.ID,
				Valid: true,
		},
	}

	errUnfollow := s.db.DeleteFeedFollowForUserId(context.Background(), uParams)
	if errUnfollow != nil {
		fmt.Printf("Could not unfollow feed for url \"%s\".\n", f.Url.String)
		return errUnfollow
	}
	return nil
}

func handlerLogin(s *state, cmd command, user database.User) error {

	if len(cmd.args) < 1 {
		return errors.New("handlerLogin() expects a single argument, the username")
	}

	if err := s.cfg.SetUser(user.Name.String, user.ID.String()); err != nil {
		return err
	}

	return nil
}

func handlerRegister(s *state, cmd command) error {

	if len(cmd.args) < 1 {
		return errors.New("the register handler expects a single argument, the new username")
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

	if err := s.cfg.SetUser(user.Name.String, user.ID.String()); err != nil {
		return err
	}

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
		fmt.Printf("Could not get user list")
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


func handlerAgg(s *state, cmd command) error {
	rssF, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return err
	}

	fmt.Println(rssF)
	return nil
}

func handlerAddFeed(s *state, cmd command) error {

	if len(cmd.args) < 2 {
		return errors.New("handlerAddFeed() expects two arguments; feed_name & url")
	}

	if _, err := url.ParseRequestURI(cmd.args[1]); err != nil {
		return err
	}

	userUUID, err := uuid.Parse(s.cfg.CurrentUserId)
	if err != nil {
		return err
	}

	uParams := database.CreateFeedParams{
		ID: uuid.New(),
		CreatedAt: sql.NullTime{
			Time: time.Now(), Valid: true,
		},
		UpdatedAt:  sql.NullTime{
			Time: time.Now(), Valid: true,
		},
		Name: sql.NullString {
			String: cmd.args[0], Valid: true,
		},
		Url: sql.NullString {
			String: cmd.args[1], Valid: true,
		},
		UserID: uuid.NullUUID {
			UUID: userUUID, Valid: true,
		},
	}

	f, err := s.db.CreateFeed(context.Background(), uParams)
	if err != nil {
		fmt.Printf("Could not create feed \"%s\".\n", uParams.Name.String)
		return err
	}

	fmt.Printf("Feed \"%s\" from \"%s\" added for \"%s\".\n", f.Name.String, f.Url.String, s.cfg.CurrentUserName)

	ff, errFF := insertFeedFollows(s, uParams.Url.String, s.cfg.CurrentUserId)
	if errFF != nil {
		return errFF
	}
	fmt.Printf("\"%s\" now following feed \"%s\"\n", ff.Username.String, ff.Feedname.String)

	return nil
}
// feedsC
func handlerGetFeeds(s *state, cmd command) error {
	uMap := make(map[uuid.UUID]string)
	uRecords, err := s.db.GetUsers(context.Background())
	if err != nil {
		fmt.Printf("Could not get user list")
		return err
	}

	for _, row := range uRecords {
		uMap[row.ID] = row.Name.String
	}

	fRecords, err := s.db.GetFeeds(context.Background())
	if err != nil {
		fmt.Printf("Could not get feeds list")
		return err
	}

	for _, row := range fRecords {
		fmt.Printf("%s\t%s\t%s\n", row.Name.String, row.Url.String, uMap[row.UserID.UUID])
	}

	return nil
}

// followC
func handlerFollow(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return errors.New("handlerFollow() expects a single argument, a url from existing feed.")
	}
	log.Printf("handlerFollow() url %s\n", cmd.args[0])

  ff, err := insertFeedFollows(s, cmd.args[0], s.cfg.CurrentUserId)
	if err != nil {
		fmt.Printf("Could follow feed for url \"%s\".\n", cmd.args[0])
		return err
	}

	fmt.Printf("\"%s\" now following feed \"%s\"\n", ff.Username.String, ff.Feedname.String)

	return nil
}

// followingC
func handlerGetFeedFollowsForUser(s *state, cmd command, user database.User) error {

	uUserId := uuid.NullUUID {
		UUID: user.ID,
		Valid: true,
	}

	fRecords, err := s.db.GetFeedFollowsForUser(context.Background(), uUserId)
	if err != nil {
		fmt.Printf("Could not get feeds list")
		return err
	}

	for _, row := range fRecords {
		fmt.Printf("%s\n", row.Feedname.String)
	}

	return nil
}

func insertFeedFollows(s *state, url, usrIdStr string) (database.CreateFeedFollowRow, error) {
	log.Printf("insertFeedFollows() url %s, usrIdStr %s\n", url, usrIdStr)
	feedBad := database.CreateFeedFollowRow{} // Satisfy return sig. on various errors.

	uURL := sql.NullString {
		String: url,
		Valid: true,
	}
	// Make sure url is for existing feed
	f, err := s.db.GetFeed(context.Background(), uURL)
	if err != nil {
		fmt.Printf("Could not get url \"%s\" to follow. Not added with addfeed?\n", uURL.String)
		return feedBad, err
	}
	// Prep. params for inserting new follow
	userUUID, err := uuid.Parse(usrIdStr)
	if err != nil {
		return feedBad, err
	}

 	uParams := database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: sql.NullTime{
			Time: time.Now(), Valid: true,
		},
		UpdatedAt:  sql.NullTime{
			Time: time.Now(), Valid: true,
		},
		UserID: uuid.NullUUID {
			UUID: userUUID, Valid: true,
		},
		FeedID: uuid.NullUUID {
			UUID: f.ID, Valid: true,
		},
	}

	ff, err := s.db.CreateFeedFollow(context.Background(), uParams)
	if err != nil {
		fmt.Printf("Could not follow feed for url \"%s\".\n", f.Url.String)
		return ff, err
	}
	return ff, nil
}
