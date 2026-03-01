package user

import "fmt"

var (
	errUserNotFound = fmt.Errorf("user not found")
)

type UserDatabase interface {
	WriteUser(u *User) error
	GetUser(id string) (*User, error)
}

// InMemoryDatabase is a placeholder for an actual database implementation.
type InMemoryDatabase struct {
	UserTable map[string]*User
}

func (db *InMemoryDatabase) WriteUser(u *User) error {
	db.UserTable[u.ID] = u
	return nil
}

func (db *InMemoryDatabase) GetUser(id string) (*User, error) {
	u, ok := db.UserTable[id]
	if !ok {
		return nil, errUserNotFound
	}
	return u, nil
}
