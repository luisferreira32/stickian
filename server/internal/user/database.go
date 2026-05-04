package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

var (
	errUserNotFound = fmt.Errorf("user not found")
)

type UserDatabase interface {
	WriteUser(ctx context.Context, u *User) error
	GetUser(ctx context.Context, email string) (*User, error)
}

type PostgresDatabase struct {
	DB *pgx.Conn
}

const writeUserQuery = "INSERT INTO users (id, email, validated_email, username, hashed_password) VALUES ($1, $2, $3, $4, $5)"

func (db *PostgresDatabase) WriteUser(ctx context.Context, u *User) error {
	_, err := db.DB.Exec(ctx, writeUserQuery, u.ID, u.Email, u.ValidatedEmail, u.Username, u.HashedPassword)
	return err
}

const getUserQuery = "SELECT id, email, validated_email, username, hashed_password FROM users WHERE email = $1"

func (db *PostgresDatabase) GetUser(ctx context.Context, email string) (*User, error) {
	row := db.DB.QueryRow(ctx, getUserQuery, email)
	var u User
	err := row.Scan(&u.ID, &u.Email, &u.ValidatedEmail, &u.Username, &u.HashedPassword)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errUserNotFound
	} else if err != nil {
		return nil, err
	}
	return &u, nil
}
