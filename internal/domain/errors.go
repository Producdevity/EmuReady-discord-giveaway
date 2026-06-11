package domain

import "errors"

var (
	ErrNotFound            = errors.New("not found")
	ErrStateInvalid        = errors.New("state invalid")
	ErrGitHubAlreadyLinked = errors.New("github account already linked")
)
