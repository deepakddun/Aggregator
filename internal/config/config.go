package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"example.com/gator/internal/database"
	"github.com/jackc/pgx/v5/pgxpool"
)

const configFileName string = ".gatorconfig.json"

type Config struct {
	DB_URL            string `json:"db_url"`
	Current_User_Name string `json:"current_user_name"`
}

type State struct {
	Config *Config
	DB     *database.Queries
	Pool   *pgxpool.Pool
}

func NewState(pool *pgxpool.Pool, db *database.Queries, config *Config) *State {
	return &State{
		Pool:   pool,
		DB:     db,
		Config: config,
	}
}

type Command struct {
	Name string
	Args []string
}

type Commands struct {
	CommandHandlerMap map[string]func(*State, Command) error
}

func Read() (Config, error) {

	baseDir, err := getConfigFilePath()
	if err != nil {
		fmt.Println("Error while reading base directory ")
		return Config{}, nil
	}
	file, err := os.ReadFile(fmt.Sprintf("%s/%s", baseDir, configFileName))
	if err != nil {
		fmt.Println("Error when opening a file ")
		return Config{}, nil
	}

	config := Config{}
	json.Unmarshal(file, &config)
	return config, nil

}

func (c *Config) SetUser(currentUser string) error {

	c.Current_User_Name = currentUser
	return write(c)
}

func write(c *Config) error {
	file, err := json.MarshalIndent(c, "", " ")
	if err != nil {
		fmt.Println("Something happened ", err)
		return err
	}
	baseDir, err := getConfigFilePath()
	if err != nil {
		fmt.Println("Error while reading base directory ", err)
		return err
	}
	os.WriteFile(fmt.Sprintf("%s/%s", baseDir, configFileName), file, 0644)
	return nil
}

func getConfigFilePath() (string, error) {

	return os.UserHomeDir()
}

func (c *Commands) Run(s *State, cmd Command) error {

	function, ok := c.CommandHandlerMap[cmd.Name]
	if !ok {
		return errors.New("Command not found")
	}
	return function(s, cmd)

}

func (c *Commands) Register(name string, f func(*State, Command) error) {

	c.CommandHandlerMap[name] = f

}

// func LoginHandler(s *State, cmd Command) error {

// 	if len(cmd.Args) == 1 {
// 		return errors.New("username is required")
// 	}
// 	if err := s.Config.SetUser(cmd.Args[1]); err != nil {
// 		return err
// 	}
// 	fmt.Println("User has been set")

// 	return nil
// }

// func RegisterHandler(s *State, cmd Command) error {
// 	if len(cmd.Args) <= 1 {
// 		return errors.New("name is required")
// 	}

// 	name := cmd.Args[2]

// 	user, err := s.DB.GetUser(context.Background(), name)

// 	if err != nil {
// 		return err
// 	}

// 	if user.ID.String() != "" {
// 		fmt.Println("User already exists...")
// 		return errors.New("User already exists ...")
// 	}

// 	arg := database.CreateUserParams{
// 		ID: pgtype.UUID{
// 			Bytes: uuid.New(),
// 			Valid: true,
// 		},
// 		CreatedAt: pgtype.Timestamptz{
// 			Time:  time.Now(),
// 			Valid: true,
// 		},
// 		Name: name,
// 	}
// 	user, err = s.DB.CreateUser(context.Background(), arg)
// 	if err != nil {
// 		return err
// 	}
// 	fmt.Printf("%v", user)
// 	return nil
// }
