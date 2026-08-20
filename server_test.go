package vapor

import (
	"reflect"
	"testing"
)

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
