package main

import "fmt"

type exitCoder interface {
	error
	ExitCode() int
}

type exitError struct {
	Err  error
	Code int
}

func (e *exitError) Error() string {
	if e == nil {
		return "error: <nil>"
	}

	if e.Err == nil {
		return fmt.Sprintf("error: exit code: %d", e.Code)
	}

	return e.Err.Error()
}

func (e *exitError) ExitCode() int {
	if e == nil {
		return 0
	}
	return e.Code
}

func (e *exitError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func exit(err error, code int) *exitError {
	return &exitError{
		Err:  err,
		Code: code,
	}
}
