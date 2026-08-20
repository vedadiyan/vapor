package vapor

import (
	"context"
	"reflect"
	"testing"
)

func TestKeyValueAdd(t *testing.T) {
	kv := KeyValue{}

	kv.Add("Content-Type", "text/plain")
	kv.Add("CONTENT-TYPE", "application/json")

	if got := kv.Get("content-type"); got != "text/plain,application/json" {
		t.Fatalf("Get() = %q", got)
	}
}

func TestKeyValueSet(t *testing.T) {
	kv := KeyValue{}

	kv.Add("X-Test", "one")
	kv.Add("X-Test", "two")
	kv.Set("X-Test", "three")

	if got := kv.Get("X-Test"); got != "three" {
		t.Fatalf("Get() = %q", got)
	}
}

func TestKeyValueGet(t *testing.T) {
	kv := KeyValue{
		"content-type": {"text/plain"},
		"x-test":       {"one", "two"},
	}

	tests := []struct {
		name string
		key  string
		want string
	}{
		{
			name: "case insensitive",
			key:  "Content-Type",
			want: "text/plain",
		},
		{
			name: "multiple values",
			key:  "X-Test",
			want: "one,two",
		},
		{
			name: "missing",
			key:  "Missing",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := kv.Get(tt.key); got != tt.want {
				t.Fatalf("Get() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestKeyValueRemove(t *testing.T) {
	kv := KeyValue{
		"x-test": {"value"},
	}

	kv.Remove("X-Test")

	if got := kv.Get("X-Test"); got != "" {
		t.Fatalf("Get() after Remove() = %q", got)
	}

	if _, ok := kv["x-test"]; ok {
		t.Fatal("key still exists after Remove()")
	}
}

func TestNewResponse(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	headers := KeyValue{
		"content-type": {"text/plain"},
	}

	content := []byte("hello")

	res := NewResponse(
		201,
		WithContent(content),
		WithHeaders(headers),
		WithContext(ctx),
	)

	if res.Status() != 201 {
		t.Fatalf("Status() = %d, want 201", res.Status())
	}

	if string(res.Content()) != "hello" {
		t.Fatalf("Content() = %q, want %q", res.Content(), "hello")
	}

	if res.Headers().Get("Content-Type") != "text/plain" {
		t.Fatalf("Headers().Get() = %q", res.Headers().Get("Content-Type"))
	}

	if res.Context() != ctx {
		t.Fatal("Context() does not return supplied context")
	}

	cancel()

	select {
	case <-res.Context().Done():
	default:
		t.Fatal("response context was not canceled")
	}
}

func TestResponseOptions(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	headers := KeyValue{
		"x-test": {"value"},
	}

	content := []byte("content")

	res := NewResponse(
		204,
		WithContent(content),
		WithHeaders(headers),
		WithContext(ctx),
	)

	if res.Status() != 204 {
		t.Fatalf("Status() = %d", res.Status())
	}

	gotContent := res.Content()
	if string(gotContent) != string(content) {
		t.Fatalf("Content() = %q, want %q", gotContent, content)
	}

	if res.Headers().Get("X-Test") != "value" {
		t.Fatalf("Headers().Get() = %q, want %q",
			res.Headers().Get("X-Test"), "value")
	}

	if res.Context() != ctx {
		t.Fatal("Context() does not return supplied context")
	}
}

func TestPatternSegments(t *testing.T) {
	tests := []struct {
		name    string
		pattern Pattern
		want    []string
	}{
		{
			name:    "root",
			pattern: "/",
			want:    []string{"", ""},
		},
		{
			name:    "path",
			pattern: "users/:id",
			want:    []string{"users", ":id"},
		},
		{
			name:    "leading slash",
			pattern: "/users/:id",
			want:    []string{"", "users", ":id"},
		},
		{
			name:    "trailing slash",
			pattern: "users/:id/",
			want:    []string{"users", ":id", ""},
		},
		{
			name:    "spaces",
			pattern: "  users/:id  ",
			want:    []string{"users", ":id"},
		},
		{
			name:    "multiple spaces",
			pattern: "  users/:id/orders/:order_id  ",
			want:    []string{"users", ":id", "orders", ":order_id"},
		},
		{
			name:    "static segments",
			pattern: "users/orders",
			want:    []string{"users", "orders"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.pattern.Segments()

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Segments() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestPatternTokens(t *testing.T) {
	tests := []struct {
		name    string
		pattern Pattern
		want    map[string]int
	}{
		{
			name:    "no tokens",
			pattern: "users/orders",
			want:    map[string]int{},
		},
		{
			name:    "single token",
			pattern: "users/:id",
			want: map[string]int{
				"id": 1,
			},
		},
		{
			name:    "multiple tokens",
			pattern: "users/:id/orders/:order_id",
			want: map[string]int{
				"id":       1,
				"order_id": 3,
			},
		},
		{
			name:    "token at root",
			pattern: ":id",
			want: map[string]int{
				"id": 0,
			},
		},
		{
			name:    "token with leading slash",
			pattern: "/users/:id",
			want: map[string]int{
				"id": 2,
			},
		},
		{
			name:    "spaces",
			pattern: "  users/:id/orders/:order_id  ",
			want: map[string]int{
				"id":       1,
				"order_id": 3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.pattern.Tokens()

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Tokens() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestPatternTokensIgnoresStaticColonPrefix(t *testing.T) {
	pattern := Pattern("users/id/:name/orders")

	got := pattern.Tokens()

	want := map[string]int{
		"name": 2,
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Tokens() = %#v, want %#v", got, want)
	}
}
