package app_test

import (
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/demo/app"
	"github.com/go-faker/faker/v4"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
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

type AppSuite struct {
	suite.Suite
	mockDB  *MockUsersDB
	appInst *app.App
}

func (s *AppSuite) SetupTest() {
	var seed int64 = time.Now().UnixNano()
	s.T().Logf("rand seed: %d", seed)
	faker.SetRandomSource(rand.NewSource(seed))

	mockCtl := gomock.NewController(s.T())
	s.mockDB = NewMockUsersDB(mockCtl)
	s.appInst = app.New(s.mockDB)
}

func (s *AppSuite) TestDuplicateUser() {
	user := s.genUser()

	s.mockDB.EXPECT().FindUser(user.ID).Return(user, nil)
	s.mockDB.EXPECT().AddUser(userMatcher{user}).Return(nil)
	newID, err := s.appInst.DuplicateUser(user.ID)

	s.Require().NoError(err)
	s.Require().NotEqual(newID, user.ID)
}

var errAddUser = errors.New("test error")

func (s *AppSuite) TestDuplicateErr() {
	user := s.genUser()

	s.mockDB.EXPECT().FindUser(user.ID).Return(user, nil)
	s.mockDB.EXPECT().AddUser(gomock.Any()).Return(errAddUser)
	_, err := s.appInst.DuplicateUser(user.ID)

	s.Require().ErrorIs(err, errAddUser)
}

func (s *AppSuite) genUser() app.User {
	return app.User{
		ID:    faker.Word(),
		Name:  faker.Word(),
		Phone: faker.Phonenumber(),
	}
}

func TestApp(t *testing.T) {
	suite.Run(t, new(AppSuite))
}
