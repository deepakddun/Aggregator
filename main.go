package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"example.com/gator/internal/config"
	"example.com/gator/internal/database"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	config.FetchFeed(ctx, "https://www.wagslane.dev/index.xml")
	if len(os.Args) < 2 {
		fmt.Println("not enough arguments were provided")
		os.Exit(1)
	}

	configS, err := config.Read()

	if err != nil {
		fmt.Println("Something happened. Please take a look ", err)
		os.Exit(1)
	}

	//DB connection
	dbPool, err := pgxpool.New(context.Background(), configS.DB_URL)

	if err != nil {
		fmt.Println("Error while connecting to DB ", err)
		os.Exit(1)
	}

	dbQueries := database.New(dbPool)

	state := config.NewState(dbPool, dbQueries, &configS)

	command := config.Command{
		Name: os.Args[1],
		Args: os.Args[1:],
	}

	commands := config.Commands{
		CommandHandlerMap: make(map[string]func(*config.State, config.Command) error),
	}

	commands.Register("login", config.LoginHandler)
	commands.Register("register", config.RegisterHandler)
	commands.Register("reset", config.ResetHandler)
	commands.Register("users", config.UsersHandler)

	err = commands.Run(state, command)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// if err != nil {
	// 	fmt.Println("Something has happened ", err)
	// 	os.Exit(1)
	// }

	// configS.SetUser("lane")

	// configS, err = config.Read()

	// if err != nil {
	// 	fmt.Println("Something has happened ", err)
	// 	os.Exit(1)
	// }

	// fmt.Println(configS)
}
