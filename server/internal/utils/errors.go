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

	// ErrUnauthorized can be used by the packages to indicate the user is not authenticated
	//
	// If returned in the utils.WithError function, it will be translated to a 401 response.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden can be used by the packages to indicate the user doesn't have permission
	//
	// If returned in the utils.WithError function, it will be translated to a 403 response.
	ErrForbidden = errors.New("forbidden")
)
