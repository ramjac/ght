package ght

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/dlclark/regexp2"
	"github.com/tealeg/xlsx"
)

// ImportExcel takes an excel of the correct format and returns a slice of HTTPTest.
func ImportExcel(fileName, tabsToTest *string, logger *VerboseLogger, retries, timeElapse, timeOut int) (r []*HTTPTest) {
	xlFile, err := xlsx.OpenFile(*fileName)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	testTabs := strings.Split(*tabsToTest, ",")

TabLoop:
	for _, tab := range xlFile.Sheets {
		// here is where we could check to see that the specified tab is one that was listed.
		if len(*tabsToTest) > 0 {
			for _, testName := range testTabs {
				if !strings.EqualFold(testName, tab.Name) {
					continue TabLoop
				}
			}
		}

		for _, row := range tab.Rows {
			tmpClient := new(HTTPTest)
			// range over the cells
			for k, v := range row.Cells {
				if v == nil || strings.TrimSpace(v.Value) == "" {
					if k == 1 {
						break
					}
					continue
				}

				switch k {
				case 0:
					tmpClient.Request = new(http.Request)

					// need to move this to new columns
					tmpClient.Retries = retries
					tmpClient.TimeElapse = timeElapse

					tmpClient.Label = `"` + v.Value + `"`
				case 1:
					u, err := url.Parse(v.Value)
					if err == nil {
						tmpClient.Request.URL = u
					} else {
						logger.Println(err)
					}
				case 2:
					tmpClient.setExcelHeaders(v.Value)
				case 3:
					tmpClient.Request.Method = v.Value
				case 4:
					tmpClient.Request.Body = ioutil.NopCloser(strings.NewReader(v.Value))
					tmpClient.Request.ContentLength = int64(len(v.Value))
				case 5:
					s, err := strconv.Atoi(v.Value)
					if err == nil {
						tmpClient.ExpectedStatus = s
					} else {
						logger.Printf("Error parsing status code: %s\n", err)
					}
				case 6:
					tmpClient.ExpectedType = strings.TrimSpace(v.Value)
				case 7:
					if len(v.Value) > 0 {
						s, err := regexp2.Compile(v.Value, regexp2.Compiled)
						if err != nil {
							logger.Printf("Error parsing regular expression: %s\n", err)
						} else {
							tmpClient.Regex = s
						}
					}
				case 8:
					if len(v.Value) > 0 {
						s, err := strconv.ParseBool(v.Value)
						if err != nil {
							logger.Printf("Error parsing the boolean for whether the regex should match or not: %s\n", err)
						} else {
							tmpClient.ExpectMatch = s
						}
					}
				case 9:
					s, err := strconv.Atoi(v.Value)
					if err == nil && s > 0 {
						tmpClient.Retries = s
					} else {
						tmpClient.Retries = retries
						logger.Printf("Error parsing retries: %s\n", err)
					}
				case 10:
					s, err := strconv.Atoi(v.Value)
					if err == nil && s > 0 {
						tmpClient.TimeElapse = s
					} else {
						tmpClient.TimeElapse = timeElapse
						logger.Printf("Error parsing time elapse: %s\n", err)
					}
				case 11:
					s, err := strconv.Atoi(v.Value)
					if err == nil && s > 0 {
						tmpClient.TimeOut = s
					} else {
						tmpClient.TimeOut = timeOut
						logger.Printf("Error parsing time elapse: %s\n", err)
					}
				}
			}

			// set defaults if no value is provided
			if tmpClient.Retries < 1 {
				tmpClient.Retries = retries
			}

			if tmpClient.TimeElapse < 1 {
				tmpClient.TimeElapse = timeElapse
			}

			if tmpClient.TimeOut < 1 {
				tmpClient.TimeOut = timeOut
			}

			if tmpClient.Request != nil {
				AddHTTPTest(tmpClient, &r)
			}
			tmpClient = new(HTTPTest)
		}
	}
	return r
}

func (h *HTTPTest) setExcelHeaders(headerString string) {
	headers := strings.Split(headerString, "\n")
	h.Request.Header = make(map[string][]string)
	for _, tmp := range headers {
		kv := strings.SplitN(tmp, ":", 2)

		if len(kv) != 2 {
			continue
		}

		h.Request.Header.Set(kv[0], strings.TrimSpace(kv[1]))
	}
}
