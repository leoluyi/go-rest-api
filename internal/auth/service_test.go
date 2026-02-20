package auth

import (
	"context"
	"testing"

	"github.com/leoluyi/go-api-template/internal/entity"
	"github.com/leoluyi/go-api-template/internal/errors"
	"github.com/leoluyi/go-api-template/pkg/log"
	"github.com/stretchr/testify/assert"
)

func Test_service_Authenticate(t *testing.T) {
	logger, _ := log.NewForTest()
	s := NewService("test_signing_key_must_be_32_chars", 100, "demo", "pass", logger)
	_, err := s.Login(context.Background(), "unknown", "bad")
	assert.Equal(t, errors.Unauthorized(""), err)
	token, err := s.Login(context.Background(), "demo", "pass")
	assert.Nil(t, err)
	assert.NotEmpty(t, token)
}

func Test_service_authenticate(t *testing.T) {
	logger, _ := log.NewForTest()
	s := service{"test_signing_key_must_be_32_chars", 100, "demo", "pass", logger}
	assert.Nil(t, s.authenticate(context.Background(), "unknown", "bad"))
	assert.NotNil(t, s.authenticate(context.Background(), "demo", "pass"))
}

func Test_service_GenerateJWT(t *testing.T) {
	logger, _ := log.NewForTest()
	s := service{"test_signing_key_must_be_32_chars", 100, "demo", "pass", logger}
	token, err := s.generateJWT(entity.User{
		ID:   "100",
		Name: "demo",
	})
	if assert.Nil(t, err) {
		assert.NotEmpty(t, token)
	}
}
