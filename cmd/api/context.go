package main

import (
	"context"
	"net/http"

	"github.com/navarroslab/whisperer/internal/data"
)

// Define a custom contextKey type, with underlying type string
type contextKey string

// Convert the string "user" to a contextKey type and assign it to the userContextKey
// constant. We'll use this constant as the key for getting and setting user information
// in the request.
const userContextKey = contextKey("user")

// The contextSetUser() method returns a new copy of the request with the provided User struct added to the context.
// Note that we use our userContextKey constant as the key.
func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

// The contextGetUser() retrieves the user from the request. We will only use this helper when we logically expect
// that there will be a user in the context. If it doesn't, it will firmly be an 'unexpected' error.
// In that case, it's ok to panic.
func (app *application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}

	return user
}
