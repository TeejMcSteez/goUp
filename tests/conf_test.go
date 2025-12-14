package utils_test

import (
	"goUp/utils"
	"testing"
)

func TestBadPath(t *testing.T) {
	_, got := utils.LoadConfig("")

	if got == nil {
		t.Error("Got nil error, should throw as string is invalid path")
	}
}
