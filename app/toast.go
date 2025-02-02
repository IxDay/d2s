package app

import (
	"math/rand"

	"github.com/a-h/templ"
	"github.com/platipy-io/d2s/server"
)

type toast uint8

const (
	ToastSuccess toast = iota
	ToastWarning
	ToastDanger
)

type Toast struct {
	Message string
	Kind    toast
}

func NewToastSuccess(msg string) templ.Component {
	return ToastTplt(Toast{Message: msg, Kind: ToastSuccess})
}

func NewToastWarning(msg string) templ.Component {
	return ToastTplt(Toast{Message: msg, Kind: ToastWarning})
}

func NewToastDanger(msg string) templ.Component {
	return ToastTplt(Toast{Message: msg, Kind: ToastDanger})
}

func NewAlertSuccess(msg string) templ.Component {
	return AlertTplt(Toast{Message: msg, Kind: ToastSuccess})
}

func NewAlertWarning(msg string) templ.Component {
	return AlertTplt(Toast{Message: msg, Kind: ToastWarning})
}

func NewAlertDanger(msg string) templ.Component {
	return AlertTplt(Toast{Message: msg, Kind: ToastDanger})
}

func Alert(ctx *server.Context) error {
	var alert templ.Component
	switch toast(rand.Int63n(3)) {
	case ToastSuccess:
		alert = NewAlertSuccess("Example alert")
	case ToastWarning:
		alert = NewAlertWarning("Example alert")
	case ToastDanger:
		alert = NewAlertDanger("Example alert")
	}
	return ctx.Render(alert)
}
