package main

import (
	"net/url"
	"strconv"
	"testing"
)

func TestFormat(t *testing.T) {
	cases := []struct {
		url      string
		format   string
		expected string
	}{
		{"https://example.com/foo", "%d", "example.com"},
		{"https://example.com/foo", "%d%p", "example.com/foo"},
		{"https://example.com/foo", "%s://%d%p", "https://example.com/foo"},

		{"https://example.com:8080/foo", "%d", "example.com"},
		{"https://example.com:8080/foo", "%P", "8080"},

		{"https://example.com/foo?a=b&c=d", "%p", "/foo"},
		{"https://example.com/foo?a=b&c=d", "%q", "a=b&c=d"},

		{"https://example.com/foo.jpg?a=b&c=d", "%e", "jpg"},
		{"https://example.com/foo.html?a=b&c=d", "%e", "html"},
		{"https://example.com/foo.tar.gz?a=b&c=d", "%e", "gz"},
		{"https://example.com/foo?a=b&c=d", "%e", ""},
		{"https://example.com/foo.html/test?a=b&c=d", "%e", ""},

		{"https://example.com/foo#bar", "%f", "bar"},
		{"https://example.com#bar", "%f", "bar"},

		{"https://example.com#bar", "foo%%bar", "foo%bar"},
		{"https://example.com#bar", "%s://%%", "https://%"},

		{"https://example.com:8080/?foo=bar#frag", "%:", ":"},
		{"https://example.com/", "%:", ""},

		{"https://example.com:8080/?foo=bar#frag", "%?", "?"},
		{"https://example.com/", "%?", ""},

		{"https://example.com:8080/?foo=bar#frag", "%#", "#"},
		{"https://example.com/", "%#", ""},

		{"https://user:pass@example.com:8080/?foo=bar#frag", "%u", "user:pass"},
		{"https://user:pass@example.com:8080/?foo=bar#frag", "%@", "@"},
		{"https://example.com/", "%@", ""},
		{"https://example.com/", "%u", ""},

		{"https://user:pass@example.com:8080/?foo=bar#frag", "%a", "user:pass@example.com:8080"},
		{"https://example.com:8080/?foo=bar#frag", "%a", "example.com:8080"},
		{"https://example.com/?foo=bar#frag", "%a", "example.com"},

		{"https://sub.example.com:8080/foo", "%S", "sub"},
		{"https://sub.example.com:8080/foo", "%r", "example"},
		{"https://sub.example.com:8080/foo", "%t", "com"},
	}

	for i, c := range cases {

		t.Run(strconv.Itoa(i), func(t *testing.T) {

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
