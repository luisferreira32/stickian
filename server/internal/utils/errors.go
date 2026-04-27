package utils

import (
	"errors"
)

var (
	// ErrNotFound can be used by the packages to indicate a resource was not found
	//
	// If returned in the utils.WithError function, it will be translated to a 404 response.
	ErrNotFound = errors.New("not found")

	// ErrUserError can be used by the packages to indicate an error caused by user input
	//
	// If returned in the utils.WithError function, it will be translated to a 400 response.
	ErrUserError = errors.New("user error")

	// ErrUserUnauthorized can be used by the packages to indicate an error caused by lacking of permissions.
	//
	// If returned in the utils.WithError function, it will be translated to a 403 response.
	ErrUserUnauthorized = errors.New("unauthorized")
)
