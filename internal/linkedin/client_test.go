package linkedin

import (
	"testing"
)

func TestParseCookies(t *testing.T) {
	raw := "li_at=AQEtoken123; JSESSIONID=ajax:sess456"
	params := parseCookies(raw)

	if len(params) != 2 {
		t.Fatalf("parseCookies returned %d params, want 2", len(params))
	}

	if params[0].Name != "li_at" || params[0].Value != "AQEtoken123" {
		t.Errorf("first cookie = %s=%s, want li_at=AQEtoken123", params[0].Name, params[0].Value)
	}
	if params[1].Name != "JSESSIONID" || params[1].Value != "ajax:sess456" {
		t.Errorf("second cookie = %s=%s, want JSESSIONID=ajax:sess456", params[1].Name, params[1].Value)
	}
	for _, p := range params {
		if p.Domain != ".linkedin.com" {
			t.Errorf("domain = %q, want .linkedin.com", p.Domain)
		}
		if p.Path != "/" {
			t.Errorf("path = %q, want /", p.Path)
		}
	}
}

func TestParseCookies_EmptyAndMalformed(t *testing.T) {
	params := parseCookies("  ; badentry; ; key=val ;")
	if len(params) != 1 {
		t.Fatalf("expected 1 valid cookie, got %d", len(params))
	}
	if params[0].Name != "key" || params[0].Value != "val" {
		t.Errorf("cookie = %s=%s, want key=val", params[0].Name, params[0].Value)
	}
}

func TestSlugFromURL(t *testing.T) {
	cases := []struct {
		url  string
		want string
	}{
		{"https://www.linkedin.com/in/alice-smith", "alice-smith"},
		{"https://linkedin.com/in/bob-jones/", "bob-jones"},
		{"https://www.linkedin.com/in/carol", "carol"},
		{"https://example.com/profile", ""},
		{"", ""},
	}
	for _, tc := range cases {
		got := slugFromURL(tc.url)
		if got != tc.want {
			t.Errorf("slugFromURL(%q) = %q, want %q", tc.url, got, tc.want)
		}
	}
}

func TestBuildSearchURL(t *testing.T) {
	url := buildSearchURL("software engineer", "Berlin")
	if url == "" {
		t.Fatal("buildSearchURL returned empty")
	}
	if !contains(url, "keywords=software") {
		t.Errorf("URL missing keywords: %s", url)
	}
	if !contains(url, "geoUrn") {
		t.Errorf("URL missing geoUrn: %s", url)
	}
	if !contains(url, baseURL+searchPath) {
		t.Errorf("URL missing base path: %s", url)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstr(s, substr)
}

func searchSubstr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
