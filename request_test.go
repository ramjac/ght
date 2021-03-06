package ght_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/ramjac/ght"
)

func TestTryRequest(t *testing.T) {
	if os.Getenv("Test_Environment") != "local" {
		t.Skip("Skipping request testing in non-local environments.")
	}

	// table of tests
	// Assumes godoc is running on http://localhost:6060 and go tool tour is also running
	requestTests := []requestCheck{
		{
			input: &ght.HTTPTest{
				Request: &http.Request{
					Method: http.MethodGet,
					URL:    MustParseUrl("http://localhost:6060"),
					Header: http.Header{
						"accepts": {"text/html; charset=utf-8"},
					},
				},
				ExpectedStatus: 200,
				ExpectedType:   "text/html; charset=utf-8",
				Regex:          MustCompileRegex("(?i)(download go)"),
				ExpectMatch:    true,
				Retries:        2,
				TimeElapse:     2,
				TimeOut:        750,
			},
			output: true,
		},
		{
			input: &ght.HTTPTest{
				Request: &http.Request{
					Method: http.MethodPost,
					URL:    MustParseUrl("http://127.0.0.1:3999/fmt"),
					Header: http.Header{
						"Host":             {"127.0.0.1:3999"},
						"User-Agent":       {"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:52.0) Gecko/20100101 Firefox/52.0"},
						"Accept":           {"application/json, text/plain, */*"},
						"Accept-Language":  {"en-US,en;q=0.5"},
						"Accept-Encoding":  {"gzip, deflate"},
						"Referer":          {"http://127.0.0.1:3999/welcome/1"},
						"x-requested-with": {"XMLHttpRequest"},
						"Content-Type":     {"application/x-www-form-urlencoded"},
					},
					Body:          ioutil.NopCloser(strings.NewReader("body=package+main%0A%0Aimport+%22fmt%22%0A%0Afunc+main()+%7B%0A%09fmt.Println(%22Hello%2C+%E4%B8%96%E7%95%8C%22)%0A%7D%0A&imports=false")),
					ContentLength: 135,
				},
				ExpectedStatus: 200,
				ExpectedType:   "text/plain; charset=utf-8",
				Regex:          MustCompileRegex("(fmt.Println)"),
				ExpectMatch:    true,
				Retries:        2,
				TimeElapse:     2,
				TimeOut:        750,
			},
			output: true,
		},
	}

	// setup
	var logger *ght.VerboseLogger
	b := true
	logger.New(&b)
	var wg sync.WaitGroup
	// Handle cancellation
	ctx := context.Background()
	// trap Ctrl+C and call cancel on the context
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// run tests
	for _, rt := range requestTests {
		wg.Add(1)
		result := rt.input.TryRequest(ctx, cancel, logger, &wg)

		if result != rt.output {
			t.Errorf(
				"Request test failed for %s %s\nExpected: %v Actual: %v",
				rt.input.Request.Method,
				rt.input.Request.URL,
				rt.output,
				result,
			)
		}
	}
}

type requestCheck struct {
	input  *ght.HTTPTest
	output bool
}
