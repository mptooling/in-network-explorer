package infrastructure

import (
	"os"
	"path/filepath"
	"testing"
)

// requiredEnvVars sets all required environment variables to valid values.
func requiredEnvVars(t *testing.T) {
	t.Helper()
	vars := map[string]string{
		"AWS_REGION":              "eu-central-1",
		"DYNAMO_TABLE":            "prospects",
		"LINKEDIN_COOKIES_SECRET": "arn:aws:secretsmanager:eu-central-1:123456789:secret:li-cookies",
		"CHROME_PROFILE_DIR":      "/tmp/chrome-profile",
	}
	for k, v := range vars {
		t.Setenv(k, v)
	}
}

// withDotEnv creates a .env file in the current directory with the given
// content and removes it when the test finishes.
func withDotEnv(t *testing.T, content string) {
	t.Helper()
	if err := os.WriteFile(".env", []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(".env") })
}

func TestMustLoad_AllPresent(t *testing.T) {
	withDotEnv(t, "")
	requiredEnvVars(t)
	t.Setenv("BEDROCK_MODEL_ID", "anthropic.claude-3-sonnet-20240229-v1:0")
	t.Setenv("BEDROCK_REGION", "us-east-1")
	t.Setenv("PROXY_ADDR", "proxy.example.com:8080")
	t.Setenv("PROXY_USER", "user")
	t.Setenv("PROXY_PASS", "pass")
	t.Setenv("CHROME_BIN", "/usr/bin/chromium")
	t.Setenv("QDRANT_ADDR", "qdrant.local:6334")
	t.Setenv("QDRANT_COLLECTION", "my-prospects")
	t.Setenv("MAX_PROFILE_VIEWS_PER_DAY", "50")
	t.Setenv("MAX_CONNECTION_REQS_PER_DAY", "15")
	t.Setenv("MAX_PROSPECTS_PER_RUN", "30")
	t.Setenv("ANALYZE_CONCURRENCY", "5")
	t.Setenv("PROMPT_CONFIG_PATH", "prompts/custom.json")

	cfg := MustLoad()

	// Required
	assertEqual(t, "AWSRegion", cfg.AWSRegion, "eu-central-1")
	assertEqual(t, "DynamoTableName", cfg.DynamoTableName, "prospects")
	assertEqual(t, "LinkedInCookiesSecret", cfg.LinkedInCookiesSecret, "arn:aws:secretsmanager:eu-central-1:123456789:secret:li-cookies")
	assertEqual(t, "ChromeProfileDir", cfg.ChromeProfileDir, "/tmp/chrome-profile")

	// Explicit overrides
	assertEqual(t, "BedrockModelID", cfg.BedrockModelID, "anthropic.claude-3-sonnet-20240229-v1:0")
	assertEqual(t, "BedrockRegion", cfg.BedrockRegion, "us-east-1")
	assertEqual(t, "ProxyAddr", cfg.ProxyAddr, "proxy.example.com:8080")
	assertEqual(t, "ProxyUser", cfg.ProxyUser, "user")
	assertEqual(t, "ProxyPass", cfg.ProxyPass, "pass")
	assertEqual(t, "ChromeBin", cfg.ChromeBin, "/usr/bin/chromium")
	assertEqual(t, "QdrantAddr", cfg.QdrantAddr, "qdrant.local:6334")
	assertEqual(t, "QdrantCollection", cfg.QdrantCollection, "my-prospects")
	assertEqualInt(t, "MaxProfileViewsPerDay", cfg.MaxProfileViewsPerDay, 50)
	assertEqualInt(t, "MaxConnectionReqsPerDay", cfg.MaxConnectionReqsPerDay, 15)
	assertEqualInt(t, "MaxProspectsPerRun", cfg.MaxProspectsPerRun, 30)
	assertEqualInt(t, "AnalyzeConcurrency", cfg.AnalyzeConcurrency, 5)
	assertEqual(t, "PromptConfigPath", cfg.PromptConfigPath, "prompts/custom.json")
}

func TestMustLoad_MissingRequired(t *testing.T) {
	required := []string{
		"AWS_REGION",
		"DYNAMO_TABLE",
		"LINKEDIN_COOKIES_SECRET",
		"CHROME_PROFILE_DIR",
	}

	for _, envVar := range required {
		t.Run(envVar, func(t *testing.T) {
			withDotEnv(t, "")
			requiredEnvVars(t)
			os.Unsetenv(envVar)
			t.Setenv(envVar+"_SENTINEL", "") // force cleanup via t.Setenv side-effect

			defer func() {
				r := recover()
				if r == nil {
					t.Fatalf("expected panic for missing %s", envVar)
				}
				msg, ok := r.(string)
				if !ok {
					t.Fatalf("panic value is not a string: %v", r)
				}
				if !containsStr(msg, envVar) {
					t.Fatalf("panic message %q does not mention %s", msg, envVar)
				}
			}()

			MustLoad()
		})
	}
}

func TestMustLoad_DefaultValues(t *testing.T) {
	withDotEnv(t, "")
	requiredEnvVars(t)

	cfg := MustLoad()

	assertEqual(t, "BedrockModelID", cfg.BedrockModelID, "anthropic.claude-haiku-4-5-20251001-v1:0")
	assertEqual(t, "BedrockRegion", cfg.BedrockRegion, "eu-central-1") // defaults to AWSRegion
	assertEqual(t, "ProxyAddr", cfg.ProxyAddr, "")
	assertEqual(t, "ProxyUser", cfg.ProxyUser, "")
	assertEqual(t, "ProxyPass", cfg.ProxyPass, "")
	assertEqual(t, "QdrantAddr", cfg.QdrantAddr, "localhost:6334")
	assertEqual(t, "QdrantCollection", cfg.QdrantCollection, "prospects")
	assertEqualInt(t, "MaxProfileViewsPerDay", cfg.MaxProfileViewsPerDay, 40)
	assertEqualInt(t, "MaxConnectionReqsPerDay", cfg.MaxConnectionReqsPerDay, 10)
	assertEqualInt(t, "MaxProspectsPerRun", cfg.MaxProspectsPerRun, 20)
	assertEqualInt(t, "AnalyzeConcurrency", cfg.AnalyzeConcurrency, 3)
	assertEqual(t, "PromptConfigPath", cfg.PromptConfigPath, "prompts/scoring.json")
}

func TestConfig_ChromeBinDefault(t *testing.T) {
	withDotEnv(t, "")
	requiredEnvVars(t)
	t.Setenv("CHROME_BIN", "") // explicitly clear — CI runners may have this set

	cfg := MustLoad()

	assertEqual(t, "ChromeBin", cfg.ChromeBin, "")
}

func TestLoadDotEnv_SetsVarsFromFile(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	os.WriteFile(envFile, []byte("FOO=bar\nBAZ=qux\n"), 0o644)

	t.Setenv("FOO", "") // ensure clean
	t.Setenv("BAZ", "")
	os.Unsetenv("FOO")
	os.Unsetenv("BAZ")

	loadDotEnv(envFile)
	defer os.Unsetenv("FOO")
	defer os.Unsetenv("BAZ")

	if got := os.Getenv("FOO"); got != "bar" {
		t.Errorf("FOO = %q, want %q", got, "bar")
	}
	if got := os.Getenv("BAZ"); got != "qux" {
		t.Errorf("BAZ = %q, want %q", got, "qux")
	}
}

func TestLoadDotEnv_DoesNotOverrideExisting(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	os.WriteFile(envFile, []byte("MY_VAR=from-file\n"), 0o644)

	t.Setenv("MY_VAR", "from-env")

	loadDotEnv(envFile)

	if got := os.Getenv("MY_VAR"); got != "from-env" {
		t.Errorf("MY_VAR = %q, want %q (env should win over file)", got, "from-env")
	}
}

func TestLoadDotEnv_SkipsCommentsAndBlanks(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	content := "# this is a comment\n\n  \nVALID_KEY=value\n# another comment\n"
	os.WriteFile(envFile, []byte(content), 0o644)

	os.Unsetenv("VALID_KEY")

	loadDotEnv(envFile)
	defer os.Unsetenv("VALID_KEY")

	if got := os.Getenv("VALID_KEY"); got != "value" {
		t.Errorf("VALID_KEY = %q, want %q", got, "value")
	}
}

func TestLoadDotEnv_MissingFilePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for missing .env file")
		}
	}()
	loadDotEnv("/nonexistent/path/.env")
}

// --- helpers ---

func assertEqual(t *testing.T, field, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %q, want %q", field, got, want)
	}
}

func assertEqualInt(t *testing.T, field string, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %d, want %d", field, got, want)
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && searchStr(s, substr)
}

func searchStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
