package config

import (
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"time"

	"example.com/gator/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func LoginHandler(s *State, cmd Command) error {

	if len(cmd.Args) == 1 {
		return errors.New("username is required")
	}
	name := cmd.Args[1]

	if name == "unknown" {
		return errors.New("unknown found")
	}

	if err := s.Config.SetUser(name); err != nil {
		return err
	}
	fmt.Println("User has been set")

	return nil
}

func RegisterHandler(s *State, cmd Command) error {
	if len(cmd.Args) < 2 {
		return errors.New("Not enough parameters")
	}

	name := cmd.Args[1]

	if name == "unknown" {
		return errors.New("unknown found")
	}

	user, err := s.DB.GetUser(context.Background(), name)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	if user.ID.String() != "" {
		fmt.Println("User already exists...")
		return errors.New("User already exists ...")
	}

	arg := database.CreateUserParams{
		ID: pgtype.UUID{
			Bytes: uuid.New(),
			Valid: true,
		},
		CreatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
		Name: name,
	}
	tx, err := s.Pool.Begin(context.Background())

	if err != nil {
		return err
	}

	//automaticall rollback in case of an error

	defer tx.Rollback(context.Background())

	user, err = s.DB.CreateUser(context.Background(), arg)
	if err != nil {
		return err
	}
	err = s.Config.SetUser(name)

	if err != nil {
		return err
	}
	fmt.Printf("%v", user)

	tx.Commit(context.Background())

	return nil
}

func ResetHandler(s *State, cmd Command) error {

	if err := s.DB.DeleteUser(context.Background()); err != nil {
		return err
	}
	fmt.Println("Database has been reset")

	return nil
}

func UsersHandler(s *State, cmd Command) error {

	userList, err := s.DB.GetUsers(context.Background())

	if err != nil {
		return err
	}

	for _, user := range userList {

		if user == s.Config.Current_User_Name {
			fmt.Printf("* %s (current)\n", user)
		} else {
			fmt.Printf("* %s \n", user)
		}

	}

	return nil

}

func FetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {

	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)

	if err != nil {
		return &RSSFeed{}, err
	}

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	response, err := client.Do(req)

	if err != nil {
		return &RSSFeed{}, err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return &RSSFeed{}, fmt.Errorf("unexpected status: %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)

	if err != nil {
		return &RSSFeed{}, err
	}

	rssFeed := &RSSFeed{}
	if err := xml.Unmarshal(body, rssFeed); err != nil {

		return &RSSFeed{}, err
	}

	rssFeed.Channel.Title = html.UnescapeString(rssFeed.Channel.Title)
	rssFeed.Channel.Description = html.UnescapeString(rssFeed.Channel.Description)

	for _, item := range rssFeed.Channel.Item {
		item.Title = html.UnescapeString(item.Title)
		item.Description = html.UnescapeString(item.Description)
	}

	fmt.Printf("%+v", rssFeed)

	return rssFeed, nil

}
