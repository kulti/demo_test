package app

import (
	"bytes"
	"fmt"
	"log"
	"text/template"
	"time"

	"github.com/cabify/timex"
)

type UsersDB interface {
	AddUser(user User) error
	FindUser(id string) (User, error)
}

type App struct {
	db     UsersDB
	bcTmpl *template.Template
}

func New(db UsersDB) *App {
	var businessCardTemplate = `Name: {{.Name}}
Phone: {{.Phone}}`
	bcTmpl, err := template.New("business card").Parse(businessCardTemplate)
	if err != nil {
		panic("parse business card template: " + err.Error())
	}

	return &App{
		db:     db,
		bcTmpl: bcTmpl,
	}
}

func (a *App) CreateUser(user User) {
	go func() {
		for {
			if err := a.db.AddUser(user); err != nil {
				const retryInterval = 5 * time.Second
				log.Printf("failed to store user into DB (retry in %s): %s\n", retryInterval, err)
				timex.Sleep(retryInterval)
			} else {
				break
			}
		}
	}()
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

func (a *App) MakeBusinessCard(userID string) (string, error) {
	user, err := a.db.FindUser(userID)
	if err != nil {
		return "", fmt.Errorf("find user: %w", err)
	}

	data := struct {
		Name  string
		Phone string
	}{
		Name:  user.Name,
		Phone: user.Phone,
	}

	var buf bytes.Buffer
	if err := a.bcTmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}
