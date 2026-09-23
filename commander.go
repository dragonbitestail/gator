package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
	"net/url"
	"github.com/google/uuid"

	"gator/pkg/database"
)

var helpC = commandItem {
	key: "help",
	help: "Command to display help for each command.",
}
var loginC = commandItem {
	key: "login",
	help: `Set user to named user >
	gator login user1`,
}
var registerC = commandItem {
	key: "register",
	help: `Add new user to Gator >
	gator register user2`,
}
var resetC = commandItem {
	key: "reset",
	help: `WARNING: Wipe all users which results in cascade delete of all
	associated user records (feeds, follows, posts)`,
}
var usersC = commandItem {
	key: "users",
	help: "List all",
}
var aggC = commandItem {
	key: "agg",
	help: `start a feed gathering loop to collect feed posts for followed
	feeds at specified interval >
	gator agg 15m
	m = minutes, 2h = hours
	To stop: ctrl-c`,
}
var addfeedC = commandItem {
	key: "addfeed",
	help: `Add feed to availble feeds to follow. Start following the feed for
	the current user >
	gator addfeed "NASA News" https://cneos.jpl.nasa.gov/feed/news.xml
	* Any other user can elect to follow the feed using the "follow" command.,
	* Feed post will not be available unless you are running the "agg" command to gather feed data.
	  This can be done in separate window since it keeps looping and gathering feed posts.`,
}
var feedsC = commandItem {
	key: "feeds",
	help: `List available feeds >
	gator feeds`,
}
var followC = commandItem {
	key: "follow",
	help: `Follow existing feed which was added previously by any user using
	the "addfeed" command >
	gator follow https://cneos.jpl.nasa.gov/feed/news.xml
	See the "feeds" command to list available feeds.`,
}
var followingC = commandItem {
	key: "following",
	help: `List feeds already followed by currently logged in user >
	gator following`,
}
var unfollowC = commandItem {
	key: "unfollow",
	help: `Unfollow a feed for the current user >
	gator unfollow https://cneos.jpl.nasa.gov/feed/news.xml`,
}
var browseC = commandItem {
	key: "browse",
	help: `Browse most recently published feeds that were gethered by "agg" command >
	gator browse 20`,
}
// END ===== command item vars

type command struct {
	name string
	args []string
}

type commands struct {
	cmdMap map[string]func(*state, command) error
	help map[string]string
	regOrder []string
}

type commandItem struct {
	key string
	help string
}



// commands struct method receivers============================================

func (c *commands) register(cmdI commandItem, f func(*state, command) error) {
	c.cmdMap[cmdI.key] = f
	c.help[cmdI.key] = cmdI.help
	c.regOrder = append(c.regOrder, cmdI.key)
	logr.Debug("register() handler", "name", cmdI.key, "help", cmdI.help)
	return
}

func (c *commands) run(s *state, cmd command) error {
	logr.Info("run() attempting to call func handler", "cmd.name", cmd.name)
	f, ok := c.cmdMap[cmd.name]
	if !ok {
		return fmt.Errorf("Unknown command: %s", cmd.name)
	}

	return f(s, cmd)
}

// Middleware command functions :===============================================
func middlewareHelp( cmds commands, handler func(s *state, cmd command, cmds commands) error) func(*state, command ) error {

	fCmdHandler := func(s *state, cmd command) error {
		logr.Info("middlewareHelp() >> HoF", "cmd.name", cmd.name, "cmd.args", cmd.args)

		return handler(s, cmd, cmds)
	}

	return fCmdHandler
}


func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {

	fCmdHandler := func(s *state, cmd command) error {
		logr.Info("middlewareLoggedIn() >> HoF", "cmd.name", cmd.name, "cmd.args", cmd.args)
		user := s.cfg.CurrentUserName
		if cmd.name == loginC.key && len(cmd.args) == 1 {
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

// Command Handlers============================================================
func handlerHelp(s *state, cmd command, cmds commands) error {
	for _, key := range cmds.regOrder {
		fmt.Printf("command: %s :: %s\n", key, cmds.help[key])
	}
	return nil
}

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

	logr.Debug("handlerRegister(): registered", "user.Name.String", user.Name.String)
	return nil
}

func handlerDeleteUsers(s *state, cmd command) error {
	logr.Debug("handlerDeleteUsers(): deleting users and all associated cascade records.")

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

// aggC
func handlerAgg(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return errors.New("handlerAgg() expects one argument; time_between_reqs (e.g., 1m, 2h, etc...")
	}

	tInterval, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return err
	}

	fmt.Printf("Collecting feeds every %v, one at a time unfeteched first, then oldest last updated.\n", tInterval)

	ticker := time.NewTicker(tInterval)
	for ;; <-ticker.C {
		err := scrapeFeeds(s)
		if err != nil {
			logr.Error("handlerAgg() scrapeFeeds() returned error", "err", err)
		}
	}

	// return nil
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
	logr.Debug("handlerFollow() arg", "url", cmd.args[0])

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
	logr.Debug("insertFeedFollows() IN", "url", url, "usrIdStr", usrIdStr)
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

// browseC
//func handlerLogin(s *state, cmd command, user database.User) error {
func handlerBrowse(s *state, cmd command, user database.User) error {
	var err error
	limitPosts := 2
	if len(cmd.args) < 1 {
		logr.Info("handlerBrowse() no limit provided using default limit")
	} else {
		limitPosts, err = strconv.Atoi(cmd.args[0])
		if err != nil {
			return err
		}
	}

	logr.Debug("handlerBrowse() building params", "user.ID", user.ID, "limitPosts", limitPosts)
	pParams := database.GetPostsForUserParams {
		UserID: uuid.NullUUID {
			UUID: user.ID, Valid: true,
		},
		Limit: int32(limitPosts),
	}

	posts, errP := s.db.GetPostsForUser(context.Background(), pParams)
	if errP != nil {
		return err
	}

	logr.Debug("handlerBrowse() printing posts")
	for _, post := range posts {
		fmt.Printf("%s\t%s\t%s\n", post.Title.String, post.Url, post.PublishedAt.Time)
	}
	return nil
}

/*
Iterate over the items in the feed and print their titles to the console.
*/
func scrapeFeeds(s *state) error {
	// Get the next feed to fetch from the DB and mark it as fetched.
	f, errF := s.db.GetNextFeedToFetch(context.Background())
	if errF != nil {
		return errF
	}

	errU := s.db.MarkFeedFetched(context.Background(), f.ID)
	if errU != nil {
		return errU
	}

	// Fetch the feed using the URL
	rssF, err := fetchFeed(context.Background(), f.Url.String)
	if err != nil {
		return err
	}
	fmt.Printf("Fetched Feed Items for Channel:: %s (%s)\n", rssF.Channel.Title, f.Url.String)
	for _, fItem := range rssF.Channel.Item {
		fmt.Printf("Feed Item Title: %s, PubDate: %s\n", fItem.Title, fItem.PubDate)
		var validDate bool = true
		pDate, err := getDateStrAsTime(fItem.PubDate)
		if err != nil {
			logr.Warn("scrapeFeeds() published date issue", "fItem.Title", fItem.Title, "err", err)
			validDate = false
		}
		// CreatePost :: posts(id, created_at, updated_at, title, url, description, published_at, feed_id)
		pParams := database.CreatePostParams {
			ID: uuid.New(),
			CreatedAt: sql.NullTime{
				Time: time.Now(), Valid: true,
			},
			UpdatedAt:  sql.NullTime{
				Time: time.Now(), Valid: true,
			},
			Title: sql.NullString {
				String: fItem.Title, Valid: true,
			},
			Url: fItem.Link,
			Description: sql.NullString {
				String: fItem.Description, Valid: true,
			},
			PublishedAt:  sql.NullTime{
				Time: pDate, Valid: validDate,
			},
			FeedID: uuid.NullUUID {
				UUID: f.ID, Valid: true,
			},
		}

		pRec, errP := s.db.CreatePost(context.Background(), pParams)
		if errP != nil {
			if strings.Contains(errP.Error(), "unique constraint") && strings.Contains(errP.Error(), "posts_url_key") {
				logr.Info("scrapeFeeds() ignore insert of already inserted post", "fItem.Title", fItem.Title)
			} else {
				logr.Warn("scrapeFeeds() unknown error on insert of post", "fItem.Title", fItem.Title)
			}
		} else {
			logr.Info("scrapeFeeds() post added", "pRec.Title", pRec.Title)
		}
	}

	return nil
}


func getDateStrAsTime(dateStr string) (time.Time, error) {
	logr.Debug("getDateStrAsTime()", "dateStr", dateStr)
	layouts := []string{
		time.RFC1123, time.RFC1123Z,
		time.Layout, time.ANSIC, time.UnixDate, time.RubyDate, time.RFC822,
		time.RFC822Z, time.RFC850, time.RFC3339,
		time.RFC3339Nano,
	}

	var errFinal error
	for _, tLayout := range layouts {
		normalizedDate, err := time.Parse(tLayout, dateStr)
		if err == nil {
			logr.Info("getDateStrAsTime() date successfully parsed", "tLayout", tLayout)
			return normalizedDate, nil
		}
		logr.Warn("getDateStrAsTime() error parsing date", "err", err)
		//logr.Error("getDateStrAsTime() could not parse date", "dateStr", dateStr)
		//return time.Now(), err
		errFinal = err
	}

	// Caller determine use of time.Now() based on error
	logr.Warn("getDateStrAsTime() could not parse date using all known layouts. Using time.Now() fallback.")
	return time.Now(), errFinal
}
