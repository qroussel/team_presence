package testutil

import (
	"math/rand"
	"os"
)

// RandomString generates a random string of length n for unique test data
func RandomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// GetTestDatabaseURL returns the database URL for testing from environment or default
func GetTestDatabaseURL() string {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://user:password@localhost:5432/presence_db?sslmode=disable"
	}
	return dbURL
}

// FindInSlice finds the first item in a slice that matches the predicate
// Returns the item and true if found, zero value and false if not found
func FindInSlice[T any](slice []T, predicate func(T) bool) (T, bool) {
	for _, item := range slice {
		if predicate(item) {
			return item, true
		}
	}
	var zero T
	return zero, false
}

// ContainsInSlice returns true if any item in the slice matches the predicate
func ContainsInSlice[T any](slice []T, predicate func(T) bool) bool {
	_, found := FindInSlice(slice, predicate)
	return found
}
