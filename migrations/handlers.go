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

func FetchAggHandler(s *State, cmd Command) error {

	if len(cmd.Args) < 2 {
		return errors.New("Not enough arguments...")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	time_between_reqs := cmd.Args[1]

	duration, err := time.ParseDuration(time_between_reqs)

	if err != nil {
		return err
	}

	ticker := time.NewTicker(duration)

	defer ticker.Stop()

	fmt.Printf("Collecting feeds every %s", time_between_reqs)
	for t := range ticker.C {
		fmt.Println(t)
		scrapFeeds(s)
	}

	rssFeed, err := fetchFeed(ctx, "https://www.wagslane.dev/index.xml")

	if err == nil {
		fmt.Printf("%+v", rssFeed)
		return nil
	}

	return err

}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {

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

	for i := range rssFeed.Channel.Item {
		rssFeed.Channel.Item[i].Title = html.UnescapeString(rssFeed.Channel.Item[i].Title)
		rssFeed.Channel.Item[i].Description =
			html.UnescapeString(rssFeed.Channel.Item[i].Description)
	}

	return rssFeed, nil

}

func AddFeedHandler(s *State, cmd Command, user database.User) error {

	if len(cmd.Args) <= 2 {
		return errors.New("Not enough argument ...")
	}

	name := cmd.Args[1]
	url := cmd.Args[2]
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

	defer cancel()

	//Get current user

	arg := database.CreateFeedParams{
		ID: pgtype.UUID{
			Bytes: uuid.New(),
			Valid: true,
		},
		CreatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
		Name:   name,
		Url:    url,
		UserID: user.ID,
	}

	feed, err := s.DB.CreateFeed(ctx, arg)

	if err != nil {
		return err
	}

	feedFolllowArgs := database.CreateFeedFollowParams{
		ID: pgtype.UUID{
			Bytes: uuid.New(),
			Valid: true,
		},
		CreatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},

		FeedID: feed.ID,
		UserID: user.ID,
	}

	_, err = s.DB.CreateFeedFollow(ctx, feedFolllowArgs)

	if err != nil {
		return err
	}

	fmt.Printf("%+v \n", feed)
	return nil
}

func ListFeedsHandler(s *State, cmd Command) error {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

	defer cancel()

	feeds, err := s.DB.ListFeeds(ctx)

	if err != nil {
		return err
	}

	for _, feed := range feeds {

		fmt.Printf("%s %s %s \n", feed.Name, feed.Url, feed.Name_2)

	}
	return nil
}

func FollowFeedsHandler(s *State, cmd Command, user database.User) error {

	if len(cmd.Args) < 2 {

		return errors.New("Not enough arguments...")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

	defer cancel()

	url := cmd.Args[1]

	//get Feed by URL

	feed, err := s.DB.GetFeedByURL(ctx, url)

	if err != nil {
		return err
	}

	arg := database.CreateFeedFollowParams{
		ID: pgtype.UUID{
			Bytes: uuid.New(),
			Valid: true,
		},
		CreatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},

		FeedID: feed.ID,
		UserID: user.ID,
	}

	feedFollows, err := s.DB.CreateFeedFollow(ctx, arg)

	for _, feedfeedFollow := range feedFollows {

		fmt.Printf("%s %s", feedfeedFollow.FeedName, feedfeedFollow.UserName)
	}

	return nil
}

func FollowingFeedsHandler(s *State, cmd Command, user database.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

	defer cancel()

	feedFollowsForUser, err := s.DB.GetFeedFollowsForUser(ctx, user.ID)

	if err != nil {
		return err
	}

	for _, feedFollow := range feedFollowsForUser {

		fmt.Printf("%s %s \n", feedFollow.FeedName, feedFollow.UserName)
	}
	return nil
}

func UnFollowHandler(s *State, cmd Command, user database.User) error {

	if len(cmd.Args) < 2 {
		return errors.New("Not enough arguments...")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

	defer cancel()

	feed, err := s.DB.GetFeedByURL(ctx, cmd.Args[1])

	if err != nil {
		return err
	}

	args := database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}
	err = s.DB.DeleteFeedFollow(ctx, args)

	if err != nil {
		return err
	}

	return nil
}

func scrapFeeds(s *State) error {

	// 	Get the next feed to fetch from the DB.
	// Mark it as fetched.
	// Fetch the feed using the URL (we already wrote this function)
	// Iterate over the items in the feed and print their titles to the console.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)

	defer cancel()

	feed, err := s.DB.GetNextFeedToFetch(ctx)

	if err != nil {
		return err
	}

	s.DB.MarkFeedFetched(ctx, feed.ID)

	rssfeed, err := fetchFeed(ctx, feed.Url)

	if err != nil {
		return err
	}

	for _, feed := range rssfeed.Channel.Item {
		fmt.Println(feed.Title)
	}
	return nil
}
