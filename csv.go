package ght

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/dlclark/regexp2"
)

// ParseCSV takes a csv of the correct format and returns a slice of HTTPTest.
func ParseCSV(rawCSV *string, logger *VerboseLogger, retries, timeElapse, timeOut int) (r []*HTTPTest) {
	tmpClient := new(HTTPTest)

	colCount := 0

	for _, v := range strings.Split(*rawCSV, ",") {
		v = strings.TrimSpace(v)

		switch colCount {
		case 0:
			tmpClient.Request = new(http.Request)
			tmpClient.Request.Method = http.MethodGet
			tmpClient.Retries = retries
			tmpClient.TimeElapse = timeElapse
			tmpClient.TimeOut = timeOut

			u, err := url.Parse(v)
			if err == nil {
				tmpClient.Request.URL = u
			} else {
				logger.Println(err)
			}
		case 1:
			tmpClient.setCSVHeaders(v)
		case 2:
			s, err := strconv.Atoi(v)
			if err == nil {
				tmpClient.ExpectedStatus = s
			} else {
				logger.Printf("Error parsing status code: %s\n", err)
			}
		case 3:
			tmpClient.ExpectedType = v
		case 4:
			if len(v) > 0 {
				s, err := regexp2.Compile(v, regexp2.Compiled)
				if err != nil {
					logger.Printf("Error parsing regular expression: %s\n", err)
				} else {
					tmpClient.Regex = s
				}
			}
		case 5:
			if len(v) > 0 {
				s, err := strconv.ParseBool(v)
				if err != nil {
					logger.Printf("Error parsing the boolean for whether the regex should match or not: %s\n", err)
				} else {
					tmpClient.ExpectMatch = s
				}
			}

			AddHTTPTest(tmpClient, &r)

			tmpClient = new(HTTPTest)

			colCount = 0
			continue
		}
		colCount++
	}

	// We'll check to see if there is an unadded tmpClient so that trailing commas aren't required.
	if tmpClient.Request != nil {
		AddHTTPTest(tmpClient, &r)
	}

	return r
}

func (h *HTTPTest) setCSVHeaders(headerString string) {
	headers := strings.Split(headerString, "&")
	h.Request.Header = make(map[string][]string)
	for _, tmp := range headers {
		kv := strings.SplitN(tmp, ":", 2)

		if len(kv) != 2 {
			continue
		}

		h.Request.Header.Set(kv[0], kv[1])
	}
}
