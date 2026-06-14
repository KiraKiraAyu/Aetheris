package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadParsesCoreConfiguration(t *testing.T) {
	withCoreWorkingDir(t, "")
	t.Setenv("API_KEYS", "secret-a:tenant-a,secret-b:tenant-b")
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://127.0.0.1:5178,http://localhost:5178")
	t.Setenv("REQUEST_MAX_BYTES", "2048")
	t.Setenv("RATE_LIMIT_ENABLED", "true")
	t.Setenv("RATE_LIMIT_PER_MINUTE", "42")

	cfg := Load()

	if cfg.APIKeys["secret-a"] != "tenant-a" || cfg.APIKeys["secret-b"] != "tenant-b" {
		t.Fatalf("api keys = %#v", cfg.APIKeys)
	}
	if len(cfg.CORSAllowedOrigins) != 2 || cfg.CORSAllowedOrigins[0] != "http://127.0.0.1:5178" {
		t.Fatalf("cors origins = %#v", cfg.CORSAllowedOrigins)
	}
	if cfg.RequestMaxBytes != 2048 || !cfg.RateLimitEnabled || cfg.RateLimitPerMinute != 42 {
		t.Fatalf("request/rate config = max:%d enabled:%v limit:%d", cfg.RequestMaxBytes, cfg.RateLimitEnabled, cfg.RateLimitPerMinute)
	}
}

func TestLoadReadsDotenvFromProjectRootWhenStartedInCore(t *testing.T) {
	clearConfigEnv(t, "HTTP_ADDR")
	withCoreWorkingDir(t, "HTTP_ADDR=:19090\n")

	cfg := Load()

	if cfg.HTTPAddr != ":19090" {
		t.Fatalf("HTTPAddr = %q, want :19090", cfg.HTTPAddr)
	}
}


func withCoreWorkingDir(t *testing.T, dotenv string) {
	t.Helper()
	root, coreDir := makeCoreWorkingDir(t)
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte(dotenv), 0o644); err != nil {
		t.Fatalf("write .env: %v", err)
	}
	chdirForTest(t, coreDir)
}

func withCoreWorkingDirWithoutDotenv(t *testing.T) {
	t.Helper()
	_, coreDir := makeCoreWorkingDir(t)
	chdirForTest(t, coreDir)
}

func makeCoreWorkingDir(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	coreDir := filepath.Join(root, "core")
	if err := os.Mkdir(coreDir, 0o755); err != nil {
		t.Fatalf("mkdir core: %v", err)
	}
	return root, coreDir
}

func chdirForTest(t *testing.T, dir string) {
	t.Helper()
	previousWorkingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previousWorkingDir); err != nil {
			t.Fatalf("restore working dir: %v", err)
		}
	})
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir core: %v", err)
	}
}

func clearConfigEnv(t *testing.T, keys ...string) {
	t.Helper()
	for _, key := range keys {
		if previous, ok := os.LookupEnv(key); ok {
			t.Cleanup(func() {
				if err := os.Setenv(key, previous); err != nil {
					t.Fatalf("restore env %s: %v", key, err)
				}
			})
		} else {
			t.Cleanup(func() {
				if err := os.Unsetenv(key); err != nil {
					t.Fatalf("unset env %s: %v", key, err)
				}
			})
		}
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("unset env %s: %v", key, err)
		}
	}
}
