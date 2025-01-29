package app

import (
	"net/http"

	"github.com/platipy-io/d2s/server"
	"github.com/platipy-io/d2s/types"
)

func LoginBypass(ctx *server.Context) error {
	ctx.User = &types.User{Name: "foo", Email: "foo@bar.com"}
	if err := ctx.SetUser(); err != nil {
		return err
	}
	ctx.Redirect("/", http.StatusFound)
	return nil
}

func Logout(ctx *server.Context) error {
	ctx.DeleteUser()
	ctx.Redirect("/", http.StatusFound)
	return nil
}
