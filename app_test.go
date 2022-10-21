package app_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/demo/app"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

//go:generate mockgen -destination=mock_db_test.go -package=app_test . UsersDB

type userMatcher struct {
	app.User
}

func (u userMatcher) Matches(x interface{}) bool {
	u2, ok := x.(app.User)
	if !ok {
		return false
	}
	u2.ID = u.ID
	return u2 == u.User
}

func (u userMatcher) String() string {
	return fmt.Sprintf("is equal to %+v", u.User)
}

func TestDuplicateUser(t *testing.T) {
	mockCtl := gomock.NewController(t)
	mockDB := NewMockUsersDB(mockCtl)
	appInst := app.New(mockDB)

	user := app.User{
		ID:    "test_user_1",
		Name:  "test_name_1",
		Phone: "8-911-234-4567",
	}

	mockDB.EXPECT().FindUser(user.ID).Return(user, nil)
	mockDB.EXPECT().AddUser(userMatcher{user}).Return(nil)
	newID, err := appInst.DuplicateUser(user.ID)

	require.NoError(t, err)
	require.NotEqual(t, newID, user.ID)
}

var errAddUser = errors.New("test error")

func TestDuplicateErr(t *testing.T) {
	mockCtl := gomock.NewController(t)
	mockDB := NewMockUsersDB(mockCtl)
	appInst := app.New(mockDB)

	user := app.User{
		ID:    "test_user_1",
		Name:  "test_name_1",
		Phone: "8-911-234-4567",
	}

	mockDB.EXPECT().FindUser(user.ID).Return(user, nil)
	mockDB.EXPECT().AddUser(gomock.Any()).Return(errAddUser)
	_, err := appInst.DuplicateUser(user.ID)

	require.ErrorIs(t, err, errAddUser)
}
