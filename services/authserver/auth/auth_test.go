package auth_test

import (
	"gotestprj/auth"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestGenerateRefreshToken(t *testing.T) {
	ref, hash, err := auth.GenerateRefreshToken()
	if err != nil {
		t.Error(err)
	}

	if err := bcrypt.CompareHashAndPassword(hash, ref); err != nil {
		t.Error(err)
	}
}
