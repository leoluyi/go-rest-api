package entity

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

var uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func TestGenerateID(t *testing.T) {
	id := GenerateID()
	assert.NotEmpty(t, id)
	assert.Regexp(t, uuidRegex, id, "ID should be a valid UUID v4")

	// Each call should produce a unique value.
	assert.NotEqual(t, id, GenerateID())
}

func TestUser_GetID(t *testing.T) {
	u := User{ID: "abc-123", Name: "Alice"}
	assert.Equal(t, "abc-123", u.GetID())
}

func TestUser_GetName(t *testing.T) {
	u := User{ID: "abc-123", Name: "Alice"}
	assert.Equal(t, "Alice", u.GetName())
}
