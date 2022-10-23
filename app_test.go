package app_test

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
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

func (s *AppSuite) TestCreateUser() {
	user := s.genUser()

	waitCh := make(chan struct{})
	s.mockDB.EXPECT().AddUser(userMatcher{user}).Do(func(app.User) {
		close(waitCh)
	})
	s.appInst.CreateUser(user)

	s.wait(waitCh)
}

func (s *AppSuite) TestDuplicateUser() {
	user := s.genUser()

	s.mockDB.EXPECT().FindUser(user.ID).Return(user, nil)
	s.mockDB.EXPECT().AddUser(userMatcher{user}).Return(nil)
	newID, err := s.appInst.DuplicateUser(user.ID)

	s.Require().NoError(err)
	s.Require().NotEqual(newID, user.ID)
}

var errAddUser = errors.New("test add error")

func (s *AppSuite) TestDuplicateAddUserErr() {
	user := s.genUser()

	s.mockDB.EXPECT().FindUser(user.ID).Return(user, nil)
	s.mockDB.EXPECT().AddUser(gomock.Any()).Return(errAddUser)
	_, err := s.appInst.DuplicateUser(user.ID)

	s.Require().ErrorIs(err, errAddUser)
}

var errFindUser = errors.New("test find error")

func (s *AppSuite) TestDuplicateFindUserErr() {
	user := s.genUser()

	s.mockDB.EXPECT().FindUser(user.ID).Return(user, errFindUser)
	_, err := s.appInst.DuplicateUser(user.ID)

	s.Require().ErrorIs(err, errFindUser)
}

func (s *AppSuite) TestMakeBusinessCard() {
	user := app.User{
		ID:    faker.Word(),
		Name:  "Mr. Frog",
		Phone: "+0-123-45-67-89",
	}
	s.mockDB.EXPECT().FindUser(user.ID).Return(user, nil)

	card, err := s.appInst.MakeBusinessCard(user.ID)
	s.Require().NoError(err)

	fileName := filepath.Join("testdata", user.Name)

	if update {
		err := os.WriteFile(fileName, []byte(card), 0644)
		s.Require().NoError(err)
	} else {
		expected, err := os.ReadFile(fileName)
		s.Require().NoError(err)
		s.Require().Equal(card, string(expected))
	}
}

func (s *AppSuite) genUser() app.User {
	return app.User{
		ID:    faker.Word(),
		Name:  faker.Word(),
		Phone: faker.Phonenumber(),
	}
}

func (s *AppSuite) wait(ch <-chan struct{}) {
	s.Require().Eventually(func() bool {
		select {
		case <-ch:
			return true
		default:
			return false
		}
	}, time.Second, time.Millisecond)
}

func TestApp(t *testing.T) {
	suite.Run(t, new(AppSuite))
}
