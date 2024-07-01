package auth

import (
	"context"
	"messages/app/db"
	"messages/app/models"
	"time"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"golang.org/x/crypto/bcrypt"
)

// Event name constants
const (
	UserSignupEvent         = "auth.signup"
	ResendVerificationEvent = "auth.resend.verification"
)

// UserWithVerificationToken is a struct that will be sent over the
// auth.signup event. It holds the User struct and the Verification token string.
type UserWithVerificationToken struct {
	User  User
	Token string
}

type Auth struct {
	UserID   int
	Email    string
	Role     string
	LoggedIn bool
}

func (auth Auth) Check() bool {
	return auth.LoggedIn
}

type User struct {
	ID              int
	Email           string
	FirstName       string
	LastName        string
	PasswordHash    string
	EmailVerifiedAt time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func createUserFromFormValues(values SignupFormValues, role string) (*models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(values.Password), bcrypt.DefaultCost)
	if err != nil {
		return &models.User{}, err
	}

	user := &models.User{
		Email:        values.Email,
		FirstName:    values.FirstName,
		LastName:     values.LastName,
		PasswordHash: string(hash),
		Role:         role,
	}

	err = user.Insert(context.Background(), db.Query, boil.Infer())

	_, err = models.Invitations(
		models.InvitationWhere.Email.EQ(values.Email),
	).DeleteAll(context.Background(), db.Query)
	if err != nil {
		return user, err
	}

	return user, err
}

type Session struct {
	ID          int
	UserID      int
	Token       string
	IPAddress   string
	UserAgent   string
	ExpiresAt   time.Time
	LastLoginAt time.Time
	CreatedAt   time.Time

	User *models.User
}
