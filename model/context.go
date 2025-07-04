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

type HasSession interface {
	GetSession() *Session
}

const userKey contextKey = "user"

func WithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// this one only makes sense as a request-scoped thing
func GetUser(ctx context.Context) (*User, bool) {
	user, ok := ctx.Value(userKey).(*User)
	return user, ok
}

type HasUser interface {
	GetUser() *User
}

type HasEnvironment interface {
	GetEnvironment() *Environment
}
