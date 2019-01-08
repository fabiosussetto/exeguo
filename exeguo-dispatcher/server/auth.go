package server

import (
	"github.com/jinzhu/gorm"

	// "github.com/volatiletech/authboss"
	"context"
	"encoding/base64"

	"github.com/gorilla/sessions"
	"github.com/volatiletech/authboss"
	abclientstate "github.com/volatiletech/authboss-clientstate"
	_ "github.com/volatiletech/authboss/auth"
	"github.com/volatiletech/authboss/defaults"
	_ "github.com/volatiletech/authboss/register"
)

type AuthStorer struct {
	DB *gorm.DB
}

func NewAuthStorer(DB *gorm.DB) *AuthStorer {
	return &AuthStorer{DB: DB}
	// return &AuthStorer{
	// 	Users: map[string]User{
	// 		"rick@councilofricks.com": User{
	// 			ID:        1,
	// 			Name:      "Rick",
	// 			Password:  "$2a$10$XtW/BrS5HeYIuOCXYe8DFuInetDMdaarMUJEOg/VA/JAIDgw3l4aG", // pass = 1234
	// 			Email:     "rick@councilofricks.com",
	// 			Confirmed: true,
	// 		},
	// 	},
	// 	Tokens: make(map[string][]string),
	// }
}

// Save the user
func (a AuthStorer) Save(ctx context.Context, user authboss.User) error {
	u := user.(*User)
	a.DB.Save(u)
	return nil
}

// Load the user
func (a AuthStorer) Load(ctx context.Context, key string) (user authboss.User, err error) {
	var dbUser User

	if a.DB.Where(&User{Email: key}).First(&dbUser).RecordNotFound() {
		return nil, authboss.ErrUserNotFound
	}

	return &dbUser, nil
}

// New user creation
func (a AuthStorer) New(ctx context.Context) authboss.User {
	return &User{}
}

// Create the user
func (a AuthStorer) Create(ctx context.Context, user authboss.User) error {
	var existingUser User
	u := user.(*User)

	if !a.DB.Where(&User{Email: u.Email}).First(&existingUser).RecordNotFound() {
		return authboss.ErrUserFound
	}

	a.DB.Create(u)
	return nil
}

// LoadByConfirmSelector looks a user up by confirmation token
// func (m MemStorer) LoadByConfirmSelector(ctx context.Context, selector string) (user authboss.ConfirmableUser, err error) {
// 	for _, v := range m.Users {
// 		if v.ConfirmSelector == selector {
// 			return &v, nil
// 		}
// 	}

// 	return nil, authboss.ErrUserNotFound
// }

// // LoadByRecoverSelector looks a user up by confirmation selector
// func (m MemStorer) LoadByRecoverSelector(ctx context.Context, selector string) (user authboss.RecoverableUser, err error) {
// 	for _, v := range m.Users {
// 		if v.RecoverSelector == selector {
// 			return &v, nil
// 		}
// 	}

// 	return nil, authboss.ErrUserNotFound
// }

// // AddRememberToken to a user
// func (m MemStorer) AddRememberToken(ctx context.Context, pid, token string) error {
// 	m.Tokens[pid] = append(m.Tokens[pid], token)
// 	return nil
// }

// // DelRememberTokens removes all tokens for the given pid
// func (m MemStorer) DelRememberTokens(ctx context.Context, pid string) error {
// 	delete(m.Tokens, pid)
// 	return nil
// }

// // UseRememberToken finds the pid-token pair and deletes it.
// // If the token could not be found return ErrTokenNotFound
// func (m MemStorer) UseRememberToken(ctx context.Context, pid, token string) error {
// 	tokens, ok := m.Tokens[pid]
// 	if !ok {
// 		return authboss.ErrTokenNotFound
// 	}

// 	for i, tok := range tokens {
// 		if tok == token {
// 			tokens[len(tokens)-1] = tokens[i]
// 			m.Tokens[pid] = tokens[:len(tokens)-1]
// 			return nil
// 		}
// 	}

// 	return authboss.ErrTokenNotFound
// }

var (
	sessionStore abclientstate.SessionStorer
	cookieStore  abclientstate.CookieStorer
)

func SetupAuth(DB *gorm.DB) *authboss.Authboss {
	ab := authboss.New()

	cookieStoreKey, _ := base64.StdEncoding.DecodeString(`NpEPi8pEjKVjLGJ6kYCS+VTCzi6BUuDzU0wrwXyf5uDPArtlofn2AG6aTMiPmN3C909rsEWMNqJqhIVPGP3Exg==`)
	sessionStoreKey, _ := base64.StdEncoding.DecodeString(`AbfYwmmt8UCwUuhd9qvfNA9UCuN1cVcKJN1ofbiky6xCyyBj20whe40rJa3Su0WOWLWcPpO1taqJdsEI/65+JA==`)
	cookieStore = abclientstate.NewCookieStorer(cookieStoreKey, nil)
	cookieStore.HTTPOnly = false
	cookieStore.Secure = false
	sessionStore = abclientstate.NewSessionStorer("foobar", sessionStoreKey, nil)
	cstore := sessionStore.Store.(*sessions.CookieStore)
	cstore.Options.HttpOnly = false
	cstore.Options.Secure = false

	ab.Config.Storage.Server = NewAuthStorer(DB)
	ab.Config.Storage.SessionState = sessionStore
	ab.Config.Storage.CookieState = cookieStore

	ab.Config.Paths.Mount = "/auth"
	ab.Config.Paths.RootURL = "http://localhost:8080/"

	// This is using the renderer from: github.com/volatiletech/authboss
	// ab.Config.Core.ViewRenderer = abrenderer.New("/auth")
	// ab.Config.Core.ViewRenderer = defaults.JSONRenderer{}
	ab.Config.Core.ViewRenderer = defaults.JSONRenderer{}
	// Probably want a MailRenderer here too.

	// Set up defaults for basically everything besides the ViewRenderer/MailRenderer in the HTTP stack
	defaults.SetCore(&ab.Config, true, false)

	// Initialize authboss (instantiate modules etc.)
	if err := ab.Init(); err != nil {
		panic(err)
	}

	return ab
}

// func xxsetupAuthboss() {
// 	ab := authboss.New()

// 	ab.Config.Paths.RootURL = "http://localhost:3000"

// 	// Set up our server, session and cookie storage mechanisms.
// 	// These are all from this package since the burden is on the
// 	// implementer for these.
// 	ab.Config.Storage.Server = database
// 	ab.Config.Storage.SessionState = sessionStore
// 	ab.Config.Storage.CookieState = cookieStore

// 	// Another piece that we're responsible for: Rendering views.
// 	// Though note that we're using the authboss-renderer package
// 	// that makes the normal thing a bit easier.
// 	ab.Config.Core.ViewRenderer = defaults.JSONRenderer{}

// 	// We render mail with the authboss-renderer but we use a LogMailer
// 	// which simply sends the e-mail to stdout.
// 	ab.Config.Core.MailRenderer = abrenderer.NewEmail("/auth", "ab_views")

// 	// The preserve fields are things we don't want to
// 	// lose when we're doing user registration (prevents having
// 	// to type them again)
// 	ab.Config.Modules.RegisterPreserveFields = []string{"email", "name"}

// 	// This instantiates and uses every default implementation
// 	// in the Config.Core area that exist in the defaults package.
// 	// Just a convenient helper if you don't want to do anything fancy.
// 	defaults.SetCore(&ab.Config, *flagAPI, false)

// 	// Here we initialize the bodyreader as something customized in order to accept a name
// 	// parameter for our user as well as the standard e-mail and password.
// 	//
// 	// We also change the validation for these fields
// 	// to be something less secure so that we can use test data easier.
// 	emailRule := defaults.Rules{
// 		FieldName: "email", Required: true,
// 		MatchError: "Must be a valid e-mail address",
// 		MustMatch:  regexp.MustCompile(`.*@.*\.[a-z]{1,}`),
// 	}
// 	passwordRule := defaults.Rules{
// 		FieldName: "password", Required: true,
// 		MinLength: 4,
// 	}
// 	nameRule := defaults.Rules{
// 		FieldName: "name", Required: true,
// 		MinLength: 2,
// 	}

// 	ab.Config.Core.BodyReader = defaults.HTTPBodyReader{
// 		ReadJSON: *flagAPI,
// 		Rulesets: map[string][]defaults.Rules{
// 			"register":    {emailRule, passwordRule, nameRule},
// 			"recover_end": {passwordRule},
// 		},
// 		Confirms: map[string][]string{
// 			"register":    {"password", authboss.ConfirmPrefix + "password"},
// 			"recover_end": {"password", authboss.ConfirmPrefix + "password"},
// 		},
// 		Whitelist: map[string][]string{
// 			"register": []string{"email", "name", "password"},
// 		},
// 	}

// 	// Initialize authboss (instantiate modules etc.)
// 	if err := ab.Init(); err != nil {
// 		panic(err)
// 	}
// }
