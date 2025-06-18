package model

import "context"

type contextKey string

const sessionKey contextKey = "session"

func WithSession(ctx context.Context, session *Session) context.Context {
	return context.WithValue(ctx, sessionKey, session)
}

func GetSession(ctx context.Context) (*Session, bool) {
	session, ok := ctx.Value(sessionKey).(*Session)
	return session, ok
}

const userKey contextKey = "user"

func WithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func GetUser(ctx context.Context) (*User, bool) {
	user, ok := ctx.Value(userKey).(*User)
	return user, ok
}

const environmentKey contextKey = "environment"

func WithEnvironment(ctx context.Context, environment *Environment) context.Context {
	return context.WithValue(ctx, environmentKey, environment)
}

func GetEnvironment(ctx context.Context) (*Environment, bool) {
	environment, ok := ctx.Value(environmentKey).(*Environment)
	return environment, ok
}
