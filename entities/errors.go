package entities

import (
	"fmt"
)

type Error struct {
	Msg string
}

func (err *Error) Error() string {
	return fmt.Sprintf("%s", err.Msg)
}

type NotFoundError struct {
	Msg string
}

func (err *NotFoundError) Error() string {
	return fmt.Sprintf("Not found error: %s", err.Msg)
}

type DatabaseError struct {
	Msg string
	Err error
}

func (err *DatabaseError) Error() string {
	return fmt.Sprintf("%s: %s", err.Msg, err.Err)
}
