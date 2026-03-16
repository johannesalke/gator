package main

import (
	"fmt"
	"github.com/johannesalke/gator/internal/config"
	"os"
)

func main() {
	s, _ := os.UserHomeDir()
	fmt.Printf("%s\n", s)
	fmt.Printf("%s\n", config.Read())

	var cfg config.Config
	cfg = config.Read()
	cfg.SetUser("johannes")
	fmt.Printf("%s", cfg.CurrentUsername)

	fmt.Print(config.Read())
}
