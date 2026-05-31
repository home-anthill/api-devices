package initialization

import (
	"strings"
	"testing"

	"go.uber.org/zap"
)

func TestPrintEnvRequiresHashSecret(t *testing.T) {
	t.Setenv("API_TOKEN_HASH_SECRET", "")

	err := printEnv(zap.NewNop().Sugar())
	if err == nil {
		t.Fatal("expected missing hash secret error")
	}
	if !strings.Contains(err.Error(), "API_TOKEN_HASH_SECRET") {
		t.Fatalf("error = %q, want API_TOKEN_HASH_SECRET", err)
	}
}

func TestPrintEnvRequiresLongHashSecret(t *testing.T) {
	t.Setenv("API_TOKEN_HASH_SECRET", "short")

	err := printEnv(zap.NewNop().Sugar())
	if err == nil {
		t.Fatal("expected short hash secret error")
	}
	if !strings.Contains(err.Error(), "at least 32 characters") {
		t.Fatalf("error = %q, want length message", err)
	}
}

func TestPrintEnvRequiresEncryptionKey(t *testing.T) {
	t.Setenv("API_TOKEN_HASH_SECRET", "12345678901234567890123456789012")
	t.Setenv("API_TOKEN_ENCRYPTION_KEY", "")

	err := printEnv(zap.NewNop().Sugar())
	if err == nil {
		t.Fatal("expected missing encryption key error")
	}
	if !strings.Contains(err.Error(), "API_TOKEN_ENCRYPTION_KEY") {
		t.Fatalf("error = %q, want API_TOKEN_ENCRYPTION_KEY", err)
	}
}

func TestPrintEnvAcceptsRequiredSecrets(t *testing.T) {
	t.Setenv("API_TOKEN_HASH_SECRET", "12345678901234567890123456789012")
	t.Setenv("API_TOKEN_ENCRYPTION_KEY", "12345678901234567890123456789012")

	if err := printEnv(zap.NewNop().Sugar()); err != nil {
		t.Fatalf("print env: %v", err)
	}
}

func TestInitLoggerReturnsLogger(t *testing.T) {
	t.Setenv("ENV", "testing")

	logger := InitLogger()
	if logger == nil {
		t.Fatal("expected logger")
	}
}

func TestGetLogFileWriterReturnsWriter(t *testing.T) {
	t.Setenv("LOG_FOLDER", t.TempDir())

	writer := getLogFileWriter()
	if writer == nil {
		t.Fatal("expected log file writer")
	}
}
