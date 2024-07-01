package auth

import (
	"cmp"
	"database/sql"
	"fmt"
	"messages/app/db"
	"messages/app/models"
	"net/http"
	"os"
	"strconv"
	"time"

	errors2 "errors"

	"github.com/anthdm/superkit/kit"
	v "github.com/anthdm/superkit/validate"
	"github.com/google/uuid"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"golang.org/x/crypto/bcrypt"
)

const (
	userSessionName = "user-session"
)

var authSchema = v.Schema{
	"email":    v.Rules(v.Email),
	"password": v.Rules(v.Required),
}

var signupSchema = v.Schema{
	"email": v.Rules(v.Email),
	"password": v.Rules(
		v.ContainsSpecial,
		v.ContainsUpper,
		v.Min(7),
		v.Max(50),
	),
	"firstName": v.Rules(v.Min(2), v.Max(50)),
	"lastName":  v.Rules(v.Min(2), v.Max(50)),
}

func HandleAuthIndex(kit *kit.Kit) error {
	if kit.Auth().Check() {
		redirectURL := cmp.Or(os.Getenv("SUPERKIT_AUTH_REDIRECT_AFTER_LOGIN"), "/")
		return kit.Redirect(http.StatusSeeOther, redirectURL)
	}
	return kit.Render(AuthIndex(AuthIndexPageData{}))
}

func HandleAuthCreate(kit *kit.Kit) error {
	var values LoginFormValues
	errors, ok := v.Request(kit.Request, &values, authSchema)
	if !ok {
		return kit.Render(LoginForm(values, errors))
	}

	user, err := models.Users(
		models.UserWhere.Email.EQ(values.Email),
	).One(kit.Request.Context(), db.Query)
	if err != nil {
		if errors2.Is(err, sql.ErrNoRows) {
			errors.Add("credentials", "unknown user")
			return kit.Render(LoginForm(values, errors))
		}
		errors.Add("credentials", "unknown error: "+err.Error())
		return kit.Render(LoginForm(values, errors))
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(values.Password))
	if err != nil {
		errors.Add("credentials", "invalid credentials")
		return kit.Render(LoginForm(values, errors))
	}

	skipVerify := kit.Getenv("SUPERKIT_AUTH_SKIP_VERIFY", "false")
	fmt.Println(skipVerify)
	if skipVerify != "true" {
		if !user.EmailVerifiedAt.Valid {
			errors.Add("verified", "please verify your email")
			return kit.Render(LoginForm(values, errors))
		}
	}

	sessionExpiryStr := kit.Getenv("SUPERKIT_AUTH_SESSION_EXPIRY_IN_HOURS", "48")
	sessionExpiry, err := strconv.Atoi(sessionExpiryStr)
	if err != nil {
		sessionExpiry = 48
	}

	session := &models.Session{
		UserID:    user.ID,
		Token:     uuid.New().String(),
		ExpiresAt: time.Now().Add(time.Hour * time.Duration(sessionExpiry)),
	}
	session.Insert(kit.Request.Context(), db.Query, boil.Infer())

	// TODO change this with kit.Getenv
	sess := kit.GetSession(userSessionName)
	sess.Values["sessionToken"] = session.Token
	sess.Save(kit.Request, kit.Response)

	redirectURL := kit.Getenv("SUPERKIT_AUTH_REDIRECT_AFTER_LOGIN", "/profile")

	return kit.Redirect(http.StatusSeeOther, redirectURL)
}

func HandleAuthDelete(kit *kit.Kit) error {
	sess := kit.GetSession(userSessionName)
	defer func() {
		sess.Values = map[any]any{}
		sess.Save(kit.Request, kit.Response)
	}()
	_, err := models.Sessions(
		models.SessionWhere.Token.EQ(sess.Values["sessionToken"].(string)),
	).DeleteAll(kit.Request.Context(), db.Query)
	if err != nil {
		return err
	}
	return kit.Redirect(http.StatusSeeOther, "/")
}

func HandleSignupIndex(kit *kit.Kit) error {
	return kit.Render(SignupIndex(SignupIndexPageData{}))
}

func HandleSignupCreate(kit *kit.Kit) error {
	var values SignupFormValues
	errors, ok := v.Request(kit.Request, &values, signupSchema)
	if !ok {
		return kit.Render(SignupForm(values, errors))
	}

	if values.Password != values.PasswordConfirm {
		errors.Add("passwordConfirm", "passwords do not match")
		return kit.Render(SignupForm(values, errors))
	}

	inviteOnly := kit.Getenv("INVITE_ONLY", "true")
	role := "user"

	ok, err := models.Users(
		models.UserWhere.Email.EQ(values.Email),
	).Exists(kit.Request.Context(), db.Query)
	if err != nil {
		errors.Add("form", "internal error")
		return kit.Render(SignupForm(values, errors))
	}
	if ok {
		errors.Add("email", "email already in use")
		return kit.Render(SignupForm(values, errors))
	}

	if inviteOnly == "true" {
		isInvited, err := models.Invitations(
			models.InvitationWhere.Email.EQ(values.Email),
		).Exists(kit.Request.Context(), db.Query)
		if err != nil {
			errors.Add("form", "internal error")
			return kit.Render(SignupForm(values, errors))
		}

		if !isInvited {
			//check if it's the first user
			count, err := models.Users().Count(kit.Request.Context(), db.Query)
			if err != nil {
				errors.Add("form", "internal error")
				return kit.Render(SignupForm(values, errors))
			}
			if count == 0 {
				role = "admin"
			} else {
				errors.Add("email", "you need an invite to sign up")
				return kit.Render(SignupForm(values, errors))
			}
		}
	}

	user, err := createUserFromFormValues(values, role)
	if err != nil {
		return err
	}

	return kit.Render(ConfirmEmail(user.Email))
}

func AuthenticateUser(kit *kit.Kit) (kit.Auth, error) {
	auth := Auth{}
	sess := kit.GetSession(userSessionName)
	token, ok := sess.Values["sessionToken"]
	if !ok {
		return auth, nil
	}

	session, err := models.Sessions(
		models.SessionWhere.Token.EQ(token.(string)),
		models.SessionWhere.ExpiresAt.GT(time.Now()),
		qm.Load(models.SessionRels.User),
	).One(kit.Request.Context(), db.Query)
	if err != nil {
		return auth, nil
	}
	// TODO: do we really need to check if the user is verified
	// even if we check that already in the login process.
	// if session.User.EmailVerifiedAt.Equal(time.Time{}) {
	// 	return Auth{}, nil
	// }
	return Auth{
		LoggedIn: true,
		UserID:   int(session.UserID),
		Email:    session.R.User.Email,
		Role:     session.R.User.Role,
	}, nil
}
