package util

import (
	"github.com/alexedwards/argon2id"
)

func HashFromPassword(input string) (string, error) {
	t := Start("HashFromPasword()")
	defer t.Stop()

	return argon2id.CreateHash(input, argon2id.DefaultParams)
}

func ComparePassword(input, expected string) bool {
	t := Start("ComparePassword()")
	defer t.Stop()

	matches, _ := argon2id.ComparePasswordAndHash(input, expected)

	return matches
}
