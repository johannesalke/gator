package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"os"
	"time"

	"github.com/johannesalke/gator/internal/config"
	"github.com/johannesalke/gator/internal/database"
	_ "github.com/lib/pq" //This is one of my least favorite things working with SQL in Go currently. You have to import the driver, but you don't use it directly anywhere in your code. The underscore tells Go that you're importing it for its side effects, not because you need to use it.
)

func main() {
	str, _ := os.UserHomeDir()
	fmt.Printf("%s\n", str)
	fmt.Printf("%s\n", config.Read())

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
