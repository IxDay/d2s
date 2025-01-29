package server

import (
	"bytes"
	"context"
	"encoding/gob"
	"net/http"
	"strings"

	"github.com/mdobak/go-xerrors"
	"github.com/platipy-io/d2s/types"
)

const (
	cookieName = "session"
)

var (
	ErrEncodeUser   = xerrors.Message("failed encoding user")
	ErrDecodingUser = xerrors.Message("failed decoding user")
)

type userKey struct{}

func newCookieUser() http.Cookie {
	return http.Cookie{Name: cookieName, Path: "/",
		HttpOnly: true, Secure: true, SameSite: http.SameSiteStrictMode,
	}
}

func SetCookieUser(resp http.ResponseWriter, user *types.User) error {
	buf := bytes.Buffer{}

	if err := gob.NewEncoder(&buf).Encode(user); err != nil {
		xerrors.WithWrapper(ErrEncodeUser, err)
	}

	cookie := newCookieUser()
	cookie.Value, cookie.MaxAge = buf.String(), 3600
	return WriteSigned(resp, cookie)
}

func DeleteCookieUser(resp http.ResponseWriter) {
	cookie := newCookieUser()
	cookie.MaxAge = -1
	http.SetCookie(resp, &cookie)
}

func GetCookieUser(req *http.Request) (*types.User, error) {

	encoded, err := ReadSigned(req, cookieName)
	// better handle "invalid cookie", "cookie not found" as bad requests
	if err != nil {
		return nil, xerrors.WithWrapper(ErrDecodingUser, err)
	}

	user := types.User{}
	reader := strings.NewReader(encoded)

	if err := gob.NewDecoder(reader).Decode(&user); err != nil {
		return nil, xerrors.WithWrapper(ErrDecodingUser, err)
	}
	return &user, nil
}

func GetUser(r *http.Request) *types.User {
	user := r.Context().Value(userKey{})
	if user != nil {
		return user.(*types.User)
	}
	return nil
}

func SetUser(r *http.Request, user *types.User) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), userKey{}, user))
}
