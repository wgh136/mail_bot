package server

import (
	"os"
	"strconv"
)

var (
	MaxUserEmailCount = 10
	MaxEmailCount     = 1000
)

func init() {
	maxUserEmailCountEnv := os.Getenv("MAX_USER_EMAIL_COUNT")
	if maxUserEmailCountEnv != "" {
		if maxUserEmailCount, err := strconv.Atoi(maxUserEmailCountEnv); err == nil {
			MaxUserEmailCount = maxUserEmailCount
		}
	}
	maxEmailCountEnv := os.Getenv("MAX_EMAIL_COUNT")
	if maxEmailCountEnv != "" {
		if maxEmailCount, err := strconv.Atoi(maxEmailCountEnv); err == nil {
			MaxEmailCount = maxEmailCount
		}
	}
}
