package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/johannesalke/gator/internal/config"
	"github.com/johannesalke/gator/internal/database"
	_ "github.com/lib/pq" //This is one of my least favorite things working with SQL in Go currently. You have to import the driver, but you don't use it directly anywhere in your code. The underscore tells Go that you're importing it for its side effects, not because you need to use it.
)

func main() {
	str, _ := os.UserHomeDir()
	fmt.Printf("%s\n", str)
	//fmt.Printf("%s\n", config.Read())

	s := &state{cfg: config.Read()}

	db, err := sql.Open("postgres", s.cfg.DBUrl)
	if err != nil {
		fmt.Println("Error opening database connection", err)
		os.Exit(1)
	}
	dbQueries := database.New(db)
	s.db = dbQueries

	c := commands{make(map[string]func(*state, command) error)}
	c.register("login", handlerLogin)
	c.register("register", handlerRegistration)
	c.register("reset", handlerReset)
	c.register("users", handlerGetUsers)
	c.register("agg", handlerAggregateFeeds)
	c.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	c.register("feeds", handlerGetFeeds)
	c.register("follow", middlewareLoggedIn(handlerFeedFollow))
	c.register("following", middlewareLoggedIn(handlerGetFollowing))
	c.register("unfollow", middlewareLoggedIn(handlerUnfollowFeed))
	c.register("browse", middlewareLoggedIn(handlerBrowse))

	args := os.Args
	if len(args) < 2 {
		fmt.Print("Error: No arguments given.\n")
		os.Exit(1)
	}
	cmd := command{Name: args[1], Args: args[2:]}

	err = c.run(s, cmd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}

/////////////////////////////////////////////////////////////////////////

type state struct {
	cfg *config.Config
	db  *database.Queries
}

type command struct {
	Name string
	Args []string
}

type commands struct {
	commands map[string]func(*state, command) error
}

/////////////|Handlers|////////////////////

func (c *commands) run(s *state, cmd command) error {
	if cmdFunc, ok := c.commands[cmd.Name]; ok {
		return cmdFunc(s, cmd)
	}
	return fmt.Errorf("Error: Command used not registered. ")
}
func (c *commands) register(name string, f func(*state, command) error) {
	c.commands[name] = f
}

/////////////|Middleware|/////////////////////////

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {

		targetUser, err := s.db.GetUser(context.Background(), s.cfg.CurrentUsername)
		if err != nil {
			return fmt.Errorf("Error retrieving user by name:%s", err)
		}

		return handler(s, cmd, targetUser)
	}

}

/////////////|Command Functions|////////////////////

func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("Error: Login expects a single argument, the username")
	}
	if _, err := s.db.GetUser(context.Background(), cmd.Args[0]); err != nil {
		fmt.Printf("User %s does not exist. Did you mean to register instead?", cmd.Args[0])
		os.Exit(1)
	}

	err := s.cfg.SetUser(cmd.Args[0])
	if err != nil {
		return err
	}
	fmt.Printf("Logged in as %s\n", cmd.Args[0])
	return nil
}

func handlerRegistration(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("Error: Login expects a single argument, the username")
	}
	params := database.CreateUserParams{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: cmd.Args[0]}
	usrData, err := s.db.CreateUser(context.Background(), params)
	if err != nil {
		fmt.Print("This user already exists.")
		os.Exit(1)
	}
	err = s.cfg.SetUser(cmd.Args[0])
	if err != nil {
		return err
	}
	fmt.Println("User was created. User data:\n", usrData)
	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.Reset(context.Background())
	if err != nil {
		fmt.Printf("An error occured during reset.")
		return err
	}
	fmt.Print("Database of users successfully reset.")
	return nil
}

func handlerGetUsers(s *state, cmd command) error {
	var users []database.User
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("Error retrieving users from db: %s", err)
	}
	for _, user := range users {
		if user.Name == s.cfg.CurrentUsername {
			fmt.Printf("%s (current)\n", user.Name)
		} else {
			fmt.Printf("%s\n", user.Name)
		}

	}
	return nil
}

func handlerAggregateFeeds(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("This command takes exactly 1 input.")
	}
	timeBetweenRequests, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return err
	}

	ticker := time.NewTicker(timeBetweenRequests)
	fmt.Printf("Collecting feed every %s", timeBetweenRequests)
	for ; ; <-ticker.C {
		err := scrapeFeed(s)
		if err != nil {
			return err
		}
	}

}

type CreateFeedParams struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string
	Url       string
	UserID    uuid.UUID
}

func handlerAddFeed(s *state, cmd command, currentUser database.User) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("This command expects exactly 2 arguments: name and url")
	}
	/*currentUser, err := s.db.GetUser(context.Background(), s.cfg.CurrentUsername)
	if err != nil {
		return fmt.Errorf("Error retrieving user_id")
	}*/
	feedParams := database.CreateFeedParams{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: cmd.Args[0], Url: cmd.Args[1], UserID: currentUser.ID}
	feed, err := s.db.CreateFeed(context.Background(), feedParams)
	if err != nil {
		return fmt.Errorf("Error creating feeds table enrtry")
	}
	cmd.Args = []string{feed.Url}

	handlerFeedFollow(s, cmd, currentUser)
	return nil
}

func handlerGetFeeds(s *state, cmd command) error {

	feeds, err := s.db.GetAllFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("Error retrieving feeds")
	}
	for i, _ := range feeds {
		user, err := s.db.GetUserByID(context.Background(), feeds[i].UserID)
		if err != nil {
			return fmt.Errorf("Error retrieving user of feed")
		}
		fmt.Printf("Name: %s | URL: %s | User: %s\n", feeds[i].Name, feeds[i].Url, user.Name)
	}
	return nil
}

func handlerFeedFollow(s *state, cmd command, targetUser database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("Error: Follow expects a single argument, the url")
	}

	/*targetUser, err := s.db.GetUser(context.Background(), s.cfg.CurrentUsername)
	if err != nil {
		return fmt.Errorf("Error retrieving user by name:%s", err)
	}*/
	targetFeed, err := s.db.GetFeedByURL(context.Background(), cmd.Args[0])
	if err != nil {
		return fmt.Errorf("Error retrieving feed by URL:%s", err)
	}

	feedFollowInput := database.CreateFeedFollowParams{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), UserID: targetUser.ID, FeedID: targetFeed.ID}
	result, err := s.db.CreateFeedFollow(context.Background(), feedFollowInput)
	if err != nil {
		return fmt.Errorf("Error creating feed_follow entry:%s", err)
	}
	fmt.Printf("user '%s' followed feed '%s'", result.UserName, result.FeedName)
	return nil
}

func handlerGetFollowing(s *state, cmd command, targetUser database.User) error {
	if len(cmd.Args) != 0 {
		return fmt.Errorf("Error: Follow expects no arguments")
	}
	/*targetUser, err := s.db.GetUser(context.Background(), s.cfg.CurrentUsername)
	if err != nil {
		return fmt.Errorf("Error retrieving user by name:%s", err)
	}*/

	follows, err := s.db.GetFeedFollowsForUser(context.Background(), targetUser.ID)
	if err != nil {
		return fmt.Errorf("Error retrieving followed feeds:%s", err)
	}
	for _, follow := range follows {
		fmt.Printf("%s\n", follow.FeedName)
	}
	return nil
}

func handlerUnfollowFeed(s *state, cmd command, targetUser database.User) error {

	targetFeed, err := s.db.GetFeedByURL(context.Background(), cmd.Args[0])
	if err != nil {
		return fmt.Errorf("Error retrieving feed by URL:%s", err)
	}

	unfollowInput := database.UnfollowFeedParams{FeedID: targetFeed.ID, UserID: targetUser.ID}
	_, err = s.db.UnfollowFeed(context.Background(), unfollowInput)
	if err != nil {
		return fmt.Errorf("Error removing follow:%s", err)
	}
	fmt.Printf("You are no longer following: %s\n", targetFeed.Name)
	return nil
}

func handlerBrowse(s *state, cmd command, targetUser database.User) error {
	var i int64
	if len(cmd.Args) == 1 {
		i, _ = strconv.ParseInt(cmd.Args[0], 10, 32)
	} else {
		i = 2
	}

	postsParams := database.GetPostsForUserParams{UserID: targetUser.ID, Limit: int32(i)}
	posts, err := s.db.GetPostsForUser(context.Background(), postsParams)
	if err != nil {
		return fmt.Errorf("Error retrieving rss posts: %s", err)
	}
	for _, post := range posts {
		fmt.Printf("Title: %s\n", post.Title)
		fmt.Printf("Url: %s\n", post.Url)
		fmt.Printf("Published: %s\n", post.PublishedAt.Time)
		fmt.Printf("%s\n\n", post.Description.String)

	}
	return nil
}
