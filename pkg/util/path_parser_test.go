package util

import "testing"

func newTestParser(t *testing.T, patterns ...string) *PathPatternParser {
	t.Helper()
	parser := NewPathPatternParser()
	for _, pattern := range patterns {
		if err := parser.AddPattern(pattern); err != nil {
			t.Fatalf("AddPattern(%s) error = %v", pattern, err)
		}
	}
	return parser
}

func TestPathPatternExactMatch(t *testing.T) {
	parser := newTestParser(t, "/api/v1/account")
	if pattern, _ := parser.Match("/api/v1/account"); pattern == nil {
		t.Fatal("/api/v1/account should match /api/v1/account")
	}
	if pattern, _ := parser.Match("/api/v1/account/1"); pattern != nil {
		t.Fatal("/api/v1/account/1 should not match /api/v1/account")
	}
}

func TestPathPatternVariables(t *testing.T) {
	parser := newTestParser(t, "/users/{id}/posts/{postId}")
	pattern, vars := parser.Match("/users/123/posts/789")
	if pattern == nil {
		t.Fatal("path should match pattern with variables")
	}
	if vars["id"] != "123" || vars["postId"] != "789" {
		t.Fatalf("unexpected variables: %v", vars)
	}
	if _, vars := parser.Match("/users/123"); vars != nil {
		t.Fatal("/users/123 should not match /users/{id}/posts/{postId}")
	}
}

func TestPathPatternWildcards(t *testing.T) {
	parser := newTestParser(t, "/static/*", "/resources/**")

	if pattern, _ := parser.Match("/static/style.css"); pattern == nil {
		t.Fatal("/static/style.css should match /static/*")
	}
	if pattern, _ := parser.Match("/static/css/style.css"); pattern != nil {
		t.Fatal("/static/css/style.css should not match /static/*")
	}
	if pattern, _ := parser.Match("/resources/a/b/c.js"); pattern == nil {
		t.Fatal("/resources/a/b/c.js should match /resources/**")
	}
	if pattern, _ := parser.Match("/resources"); pattern == nil {
		t.Fatal("/resources should match /resources/**")
	}
}

func TestPathPatternPriority(t *testing.T) {
	parser := newTestParser(t, "/users/{id}", "/users/me")
	pattern, _ := parser.Match("/users/me")
	if pattern == nil || pattern.Pattern != "/users/me" {
		t.Fatalf("static pattern should win over variable pattern, got %+v", pattern)
	}
}

func TestPathPatternTrailingSlash(t *testing.T) {
	parser := newTestParser(t, "/users/{id}")
	if pattern, _ := parser.Match("/users/123/"); pattern == nil {
		t.Fatal("trailing slash should be ignored when matching")
	}
}

func TestPathPatternInvalidPattern(t *testing.T) {
	parser := NewPathPatternParser()
	if err := parser.AddPattern("/users/({invalid}"); err == nil {
		t.Fatal("expected error for invalid pattern")
	}
}
