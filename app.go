package app

import "fmt"

type UsersDB interface {
	AddUser(user User) error
	FindUser(id string) (User, error)
}

type App struct {
	db UsersDB
}

func New(db UsersDB) *App {
	return &App{
		db: db,
	}
}

func (a *App) DuplicateUser(userID string) (string, error) {
	user, err := a.db.FindUser(userID)
	if err != nil {
		return "", fmt.Errorf("find user: %w", err)
	}

	user.ID = user.ID + "_"
	err = a.db.AddUser(user)
	return user.ID, err
}
