package postgresql

import (
	"os"
	"testing"

	"github.com/sayerxofficial/langchaingo/internal/testutil/testctr"
)

func TestMain(m *testing.M) {
	code := testctr.EnsureTestEnv()
	if code == 0 {
		code = m.Run()
	}
	os.Exit(code)
}
