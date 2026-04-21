package main

import (
	"context"
	"errors"

	"example.com/gator/internal/config"
	"example.com/gator/internal/database"
)

func middlewareLoggedIn(handler func(s *config.State, cmd config.Command, user database.User) error) func(*config.State, config.Command) error {

	return func(s *config.State, cmd config.Command) error {

		user, err := getLoggedInUser(s)

		if err != nil {
			return err
		}

		return handler(s, cmd, user)
	}

	//return nil
}

func getLoggedInUser(s *config.State) (database.User, error) {
	if s.Config.Current_User_Name == "" {
		return database.User{}, errors.New("no user logged in")
	}

	ctx := context.Background()

	user, err := s.DB.GetUser(ctx, s.Config.Current_User_Name)
	if err != nil {
		return database.User{}, err
	}

	return user, nil
}
