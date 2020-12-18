package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"net/url"
	"os"
	"regexp"

	"github.com/Cgboal/DomainParser"
)

var extractor parser.Parser

func main() {

	var unique bool
	flag.BoolVar(&unique, "u", false, "")
	flag.BoolVar(&unique, "unique", false, "")

	var verbose bool
	flag.BoolVar(&verbose, "v", false, "")
	flag.BoolVar(&verbose, "verbose", false, "")

	flag.Parse()

	mode := flag.Arg(0)
	fmtStr := flag.Arg(1)

	procFn, ok := map[string]urlProc{
		"keys":     keys,
		"values":   values,
		"keypairs": keyPairs,
		"domains":  domains,
		"domain":   domains,
		"paths":    paths,
		"path":     paths,
		"format":   format,
	}[mode]

	if !ok {
		fmt.Fprintf(os.Stderr, "unknown mode: %s\n", mode)
		return
	}

	sc := bufio.NewScanner(os.Stdin)

	seen := make(map[string]bool)

	extractor = parser.NewDomainParser()

	for sc.Scan() {
		u, err := parseURL(sc.Text())
		if err != nil {
			if verbose {
				fmt.Fprintf(os.Stderr, "parse failure: %s\n", err)
			}
			continue
		}

		// some urlProc functions return multiple things,
		// so it's just easier to always get a slice and
		// loop over it instead of having two kinds of
		// urlProc functions.
		for _, val := range procFn(u, fmtStr) {

			// you do see empty values sometimes
			if val == "" {
				continue
			}

			if seen[val] && unique {
				continue
			}

			fmt.Println(val)

			// no point using up memory if we're outputting dupes
			if unique {
				seen[val] = true
			}
		}
	}

	if err := sc.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to read input: %s\n", err)
	}
}

// parseURL parses a string as a URL and returns a *url.URL
// or any error that occured. If the initially parsed URL
// has no scheme, http:// is prepended and the string is
// re-parsed
func parseURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "" {
		return url.Parse("http://" + raw)
	}

	return u, nil
}

// a urlProc is any function that accepts a URL and some
// kind of format string (which may not actually be used
// by some functions), and returns a slice of strings
// derived from that URL. It's not uncommon for a urlProc
// function to return a slice of length 1, but the return
// type remains a slice because *some* functions need to
// return multiple strings; e.g. the keys function.
type urlProc func(*url.URL, string) []string

// keys returns all of the keys used in the query string
// portion of the URL. E.g. for /?one=1&two=2&three=3 it
// will return []string{"one", "two", "three"}
func keys(u *url.URL, _ string) []string {
	out := make([]string, 0)
	for key, _ := range u.Query() {
		out = append(out, key)
	}
	return out
}

// values returns all of the values in the query string
// portion of the URL. E.g. for /?one=1&two=2&three=3 it
// will return []string{"1", "2", "3"}
func values(u *url.URL, _ string) []string {
	out := make([]string, 0)
	for _, vals := range u.Query() {
		for _, val := range vals {
			out = append(out, val)
		}
	}
	return out
}

// keyPairs returns all the key=value pairs in
// the query string portion of the URL. E.g for
// /?one=1&two=2&three=3 it will return
// []string{"one=1", "two=2", "three=3"}
func keyPairs(u *url.URL, _ string) []string {
	out := make([]string, 0)
	for key, vals := range u.Query() {
		for _, val := range vals {
			out = append(out, fmt.Sprintf("%s=%s", key, val))
		}
	}
	return out
}

// domains returns the domain portion of the URL. e.g.
// for http://sub.example.com/path it will return
// []string{"sub.example.com"}
func domains(u *url.URL, f string) []string {
	return format(u, "%d")
}

// domains returns the path portion of the URL. e.g.
// for http://sub.example.com/path it will return
// []string{"/path"}
func paths(u *url.URL, f string) []string {
	return format(u, "%p")
}

// format is a little bit like a special sprintf for
// URLs; it will return a single formatted string
// based on the URL and the format string. e.g. for
// http://example.com/path and format string "%d%p"
// it will return example.com/path
func format(u *url.URL, f string) []string {
	out := &bytes.Buffer{}

	inFormat := false
	for _, r := range f {

		if r == '%' && !inFormat {
			inFormat = true
			continue
		}

		if !inFormat {
			out.WriteRune(r)
			continue
		}

		switch r {

		// a literal percent rune
		case '%':
			out.WriteRune('%')

		// the scheme; e.g. http
		case 's':
			out.WriteString(u.Scheme)

		// the userinfo; e.g. user:pass
		case 'u':
			if u.User != nil {
				out.WriteString(u.User.String())
			}

		// the domain; e.g. sub.example.com
		case 'd':
			out.WriteString(u.Hostname())

		// the port; e.g. 8080
		case 'P':
			out.WriteString(u.Port())

		// the subdomain; e.g. www
		case 'S':
			out.WriteString(extractFromDomain(u, "subdomain"))

		// the root; e.g. example
		case 'r':
			out.WriteString(extractFromDomain(u, "root"))

		// the tld; e.g. com
		case 't':
			out.WriteString(extractFromDomain(u, "tld"))

		// the path; e.g. /users
		case 'p':
			out.WriteString(u.EscapedPath())

		// the query string; e.g. one=1&two=2
		case 'q':
			out.WriteString(u.RawQuery)

		// the fragment / hash value; e.g. section-1
		case 'f':
			out.WriteString(u.Fragment)

		// an @ if user info is specified
		case '@':
			if u.User != nil {
				out.WriteRune('@')
			}

		// a colon if a port is specified
		case ':':
			if u.Port() != "" {
				out.WriteRune(':')
			}

		// a question mark if there's a query string
		case '?':
			if u.RawQuery != "" {
				out.WriteRune('?')
			}

		// a hash if there is a fragment
		case '#':
			if u.Fragment != "" {
				out.WriteRune('#')
			}

		// the authority; e.g. user:pass@example.com:8080
		case 'a':
			out.WriteString(format(u, "%u%@%d%:%P")[0])

		// default to literal
		default:
			// output untouched
			out.WriteRune('%')
			out.WriteRune(r)
		}

		inFormat = false
	}

	return []string{out.String()}
}

func extractFromDomain(u *url.URL, selection string) string {

	// remove the port before parsing
	portRe := regexp.MustCompile(`(?m):\d+$`)
	
	domain := portRe.ReplaceAllString(u.Host, "")
	
	switch selection {
	case "subdomain":
		return extractor.GetSubdomain(domain)
	case "root":
		return extractor.GetDomain(domain)
	case "tld":
		return extractor.GetTld(domain)
	default:
		return ""
	}
}

func init() {
	flag.Usage = func() {
		h := "Format URLs provided on stdin\n\n"

		h += "Usage:\n"
		h += "  unfurl [OPTIONS] [MODE] [FORMATSTRING]\n\n"

		h += "Options:\n"
		h += "  -u, --unique   Only output unique values\n"
		h += "  -v, --verbose  Verbose mode (output URL parse errors)\n\n"

		h += "Modes:\n"
		h += "  keys     Keys from the query string (one per line)\n"
		h += "  values   Values from the query string (one per line)\n"
		h += "  keypairs Key=value pairs from the query string (one per line)\n"
		h += "  domains  The hostname (e.g. sub.example.com)\n"
		h += "  paths    The request path (e.g. /users)\n"
		h += "  format   Specify a custom format (see below)\n\n"

		h += "Format Directives:\n"
		h += "  %%  A literal percent character\n"
		h += "  %s  The request scheme (e.g. https)\n"
		h += "  %u  The user info (e.g. user:pass)\n"
		h += "  %d  The domain (e.g. sub.example.com)\n"
		h += "  %S  The subdomain (e.g. sub)\n"
		h += "  %r  The root of domain (e.g. example)\n"
		h += "  %t  The TLD (e.g. com)\n"
		h += "  %P  The port (e.g. 8080)\n"
		h += "  %p  The path (e.g. /users)\n"
		h += "  %q  The raw query string (e.g. a=1&b=2)\n"
		h += "  %f  The page fragment (e.g. page-section)\n"
		h += "  %@  Inserts an @ if user info is specified\n"
		h += "  %:  Inserts a colon if a port is specified\n"
		h += "  %?  Inserts a question mark if a query string exists\n"
		h += "  %#  Inserts a hash if a fragment exists\n"
		h += "  %a  Authority (alias for %u%@%d%:%P)\n\n"

		h += "Examples:\n"
		h += "  cat urls.txt | unfurl keys\n"
		h += "  cat urls.txt | unfurl format %s://%d%p?%q\n"

		fmt.Fprint(os.Stderr, h)
	}
}
