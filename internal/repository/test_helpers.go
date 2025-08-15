package repository

import (
	"database/sql/driver"
	"time"
)

// stringPtr returns a pointer to the given string
func stringPtr(s string) *string {
	return &s
}

// timePtr returns a pointer to the given time
func timePtr(t time.Time) *time.Time {
	return &t
}

// durationPtr returns a pointer to the given duration
func durationPtr(d time.Duration) *time.Duration {
	return &d
}

// int64Ptr returns a pointer to the given int64
func int64Ptr(i int64) *int64 {
	return &i
}

// float64Ptr returns a pointer to the given float64
func float64Ptr(f float64) *float64 {
	return &f
}

// boolPtr returns a pointer to the given bool
func boolPtr(b bool) *bool {
	return &b
}

// null represents a NULL value for testing
var null = driver.Value(nil)