package main

import (
	"context"
	"net/http"

	"github.com/hwanbin/wanpm-api/internal/data"
)

type contextKey string

const userContextKey = contextKey("user")

func (app *application) contextSetUser(r *http.Request, user *data.AppUser) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

func (app *application) contextGetUser(r *http.Request) *data.AppUser {
	user, ok := r.Context().Value(userContextKey).(*data.AppUser)
	if !ok {
		panic("missing user value in request context")
	}

	return user
}
