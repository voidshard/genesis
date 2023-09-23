package database

import (
	"github.com/voidshard/genesis/internal/dbutils"
)

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func validNames(in ...string) bool {
	for _, name := range in {
		if dbutils.IsValidName(name) {
			continue
		}
		return false
	}
	return true
}
