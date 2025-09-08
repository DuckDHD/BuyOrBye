package repositories

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsDuplicateKeyError(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "Nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "SQLite UNIQUE constraint error",
			err:      errors.New("UNIQUE constraint failed: users.email"),
			expected: true,
		},
		{
			name:     "MySQL duplicate entry error",
			err:      errors.New("Duplicate entry 'test@example.com' for key 'email'"),
			expected: true,
		},
		{
			name:     "PostgreSQL duplicate key error",
			err:      errors.New("duplicate key value violates unique constraint"),
			expected: true,
		},
		{
			name:     "Generic duplicate value error",
			err:      errors.New("duplicate value"),
			expected: true,
		},
		{
			name:     "Non-duplicate error",
			err:      errors.New("connection timeout"),
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isDuplicateKeyError(tc.err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestContains(t *testing.T) {
	testCases := []struct {
		name     string
		str      string
		substr   string
		expected bool
	}{
		{
			name:     "Exact match",
			str:      "duplicate key",
			substr:   "duplicate key",
			expected: true,
		},
		{
			name:     "Substring at beginning",
			str:      "duplicate key error",
			substr:   "duplicate",
			expected: true,
		},
		{
			name:     "Substring at end",
			str:      "error duplicate key",
			substr:   "key",
			expected: true,
		},
		{
			name:     "Substring in middle",
			str:      "error duplicate key error",
			substr:   "duplicate",
			expected: true,
		},
		{
			name:     "Case insensitive uppercase match",
			str:      "DUPLICATE KEY",
			substr:   "DUPLICATE KEY",
			expected: true,
		},
		{
			name:     "Mixed case match",
			str:      "Duplicate Key Error",
			substr:   "duplicate key",
			expected: true,
		},
		{
			name:     "Not found",
			str:      "connection timeout",
			substr:   "duplicate",
			expected: false,
		},
		{
			name:     "Empty substring",
			str:      "test string",
			substr:   "",
			expected: true,
		},
		{
			name:     "Substring longer than string",
			str:      "short",
			substr:   "very long substring",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := contains(tc.str, tc.substr)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestToLower(t *testing.T) {
	testCases := []struct {
		name     string
		input    byte
		expected byte
	}{
		{
			name:     "Uppercase A",
			input:    'A',
			expected: 'a',
		},
		{
			name:     "Uppercase Z",
			input:    'Z',
			expected: 'z',
		},
		{
			name:     "Already lowercase",
			input:    'a',
			expected: 'a',
		},
		{
			name:     "Number",
			input:    '5',
			expected: '5',
		},
		{
			name:     "Special character",
			input:    '@',
			expected: '@',
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := toLower(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}