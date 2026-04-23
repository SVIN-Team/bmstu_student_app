//go:build unit

package module

import (
	"io"
	"os"
	"testing"

	"stud_hub/util/logger"
)

func TestMain(m *testing.M) {
	logger.InitLogger("error", io.Discard)
	
	code := m.Run()
	
	os.Exit(code)
}