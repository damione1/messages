package auth

import (
	"context"
	"messages/app/db"
	"messages/app/models"
	"time"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	UserID   int
	Email    string
	LoggedIn bool
}

func (auth Auth) Check() bool {
	return auth.LoggedIn
}

type User struct {
	ID              int `bun:"id,pk,autoincrement"`
	Email           string
	FirstName       string
	LastName        string
	PasswordHash    string
	EmailVerifiedAt time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func createUserFromFormValues(values SignupFormValues) (*models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(values.Password), bcrypt.DefaultCost)
	if err != nil {
		return &models.User{}, err
	}

	user := &models.User{
		Email:        values.Email,
		FirstName:    values.FirstName,
		LastName:     values.LastName,
		PasswordHash: string(hash),
	}

	err = user.Insert(context.Background(), db.Query, boil.Infer())

	return user, err
}

type Session struct {
	ID          int `bun:"id,pk,autoincrement"`
	UserID      int
	Token       string
	IPAddress   string
	UserAgent   string
	ExpiresAt   time.Time
	LastLoginAt time.Time
	CreatedAt   time.Time

	User User `bun:"rel:belongs-to,join:user_id=id"`
}
