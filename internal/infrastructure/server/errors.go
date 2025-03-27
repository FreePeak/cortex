package server

import "errors"

// Common errors in the server package
var (
	// ErrResponseWriterNotFlusher is returned when the ResponseWriter doesn't support Flusher interface
	ErrResponseWriterNotFlusher = errors.New("response writer does not implement http.Flusher")

	// ErrSessionNotFound is returned when a session cannot be found
	ErrSessionNotFound = errors.New("session not found")

	// ErrSessionClosed is returned when attempting to use a closed session
	ErrSessionClosed = errors.New("session is closed")

	// ErrChannelFull is returned when a notification channel is full
	ErrChannelFull = errors.New("notification channel is full")
)
