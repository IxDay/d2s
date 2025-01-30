package app

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/mdobak/go-xerrors"
	"github.com/platipy-io/d2s/internal/github"
	"github.com/platipy-io/d2s/server"
)

const oauthStateCookieName = "oauthstate"
const userCookieName = "user"

var (
	ErrCookieGenerate  = xerrors.Message("failed to generate cookie: " + oauthStateCookieName)
	ErrCookieRetrieval = xerrors.Message("failed to retrieve cookie: " + oauthStateCookieName)
	ErrCookieUser      = xerrors.Message("failed to generate cookie: " + userCookieName)
	ErrInvalidState    = xerrors.Message("invalid oauth github state")
	ErrInvalidCode     = xerrors.Message("invalid oauth github code")

	durationState = 20 * time.Minute
)

func Logout(ctx *server.Context) error {
	ctx.DeleteUser()
	ctx.Redirect("/", http.StatusTemporaryRedirect)
	return nil
}

func Login(ctx *server.Context) error {
	// Create oauthState cookie
	oauthState, err := generateStateOauthCookie(ctx)
	if err != nil {
		return New500HTTPError(err)
	}
	/*
		AuthCodeURL receive state that is a token to protect the user from CSRF attacks. You must always provide a non-empty string and
		validate that it matches the the state query parameter on your redirect callback.
	*/
	ctx.Redirect(github.AuthCodeURL(oauthState), http.StatusTemporaryRedirect)
	return nil
}

func LoginBypass(ctx *server.Context) error {
	user, err := github.UserBypass(ctx.Context())
	if err != nil {
		return New500HTTPError(err)
	}
	ctx.User = user
	if err := ctx.SetUser(); err != nil {
		return New500HTTPError(err)
	}
	ctx.Redirect("/", http.StatusTemporaryRedirect)
	return nil
}

func Callback(ctx *server.Context) error {
	// Read oauthState from Cookie
	oauthState, err := ctx.Cookie(oauthStateCookieName)
	if err != nil {
		return New400HTTPError(err)
	}

	if ctx.FormValue("state") != oauthState.Value {
		return New400HTTPError(ErrInvalidState)
	}
	token, err := github.Exchange(ctx.Context(), ctx.FormValue("code"))
	if err != nil {
		return New400HTTPError(ErrInvalidCode)
	}
	user, err := github.User(ctx.Context(), token)
	if err != nil {
		return New500HTTPError(err)
	}
	ctx.User = user
	if err := ctx.SetUser(); err != nil {
		return New500HTTPError(err)
	}
	ctx.Logger.Info().Msg("successfully logged client through github oauth")
	ctx.Redirect("/", http.StatusTemporaryRedirect)
	return nil
}

func generateStateOauthCookie(ctx *server.Context) (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", xerrors.WithWrapper(ErrCookieGenerate, err)
	}
	state := base64.URLEncoding.EncodeToString(b)
	ctx.SetCookie(oauthStateCookieName, state, durationState)

	return state, nil
}
