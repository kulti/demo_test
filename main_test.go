package app_test

import (
	"flag"
	"os"
	"testing"
)

var update bool

func TestMain(m *testing.M) {
	flag.BoolVar(&update, "update", false, "update golden test files")
	os.Exit(m.Run())
}
