package main

import (
	"net/url"
	"runtime/debug"
	"strconv"
	"testing"
)

func TestFormat(t *testing.T) {
	cases := []struct {
		url       string
		format    string
		expected  string
		wantPanic bool
	}{
		{"https://example.com/foo", "%d", "example.com", false},
		{"https://example.com/foo", "%d%p", "example.com/foo", false},
		{"https://example.com/foo", "%s://%d%p", "https://example.com/foo", false},

		{"https://example.com:8080/foo", "%d", "example.com", false},
		{"https://example.com:8080/foo", "%P", "8080", false},

		{"https://example.com/foo?a=b&c=d", "%p", "/foo", false},
		{"https://example.com/foo?a=b&c=d", "%q", "a=b&c=d", false},

		{"https://example.com/foo#bar", "%f", "bar", false},
		{"https://example.com#bar", "%f", "bar", false},

		{"https://example.com#bar", "foo%%bar", "foo%bar", false},
		{"https://example.com#bar", "%s://%%", "https://%", false},

		{"https://example.com:8080/?foo=bar#frag", "%:", ":", false},
		{"https://example.com/", "%:", "", false},

		{"https://example.com:8080/?foo=bar#frag", "%?", "?", false},
		{"https://example.com/", "%?", "", false},

		{"https://example.com:8080/?foo=bar#frag", "%#", "#", false},
		{"https://example.com/", "%#", "", false},

		{"https://user:pass@example.com:8080/?foo=bar#frag", "%u", "user:pass", false},
		{"https://user:pass@example.com:8080/?foo=bar#frag", "%@", "@", false},
		{"https://example.com/", "%@", "", false},
		{"https://example.com/", "%u", "", true},

		{"https://user:pass@example.com:8080/?foo=bar#frag", "%a", "user:pass@example.com:8080", false},
		{"https://example.com:8080/?foo=bar#frag", "%a", "example.com:8080", true},
		{"https://example.com/?foo=bar#frag", "%a", "example.com", true},

		{"https://sub.example.com:8080/foo", "%S", "sub", false},
		{"https://sub.example.com:8080/foo", "%r", "example", false},
		{"https://sub.example.com:8080/foo", "%t", "com", false},
	}

	for i, c := range cases {

		t.Run(strconv.Itoa(i), func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != c.wantPanic {
					t.Errorf("panic recover = %v, wantPanic = %v", r, c.wantPanic)
					debug.PrintStack()
				}
			}()

			u, err := url.Parse(c.url)
			if err != nil {
				t.Fatal(err)
			}

			actual := format(u, c.format)

			if actual[0] != c.expected {
				t.Errorf("want %s for format(%s, %s); have %s", c.expected, c.url, c.format, actual)
			}
		})
	}
}
