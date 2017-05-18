package hab

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

var TEST_HAB_URL = "http://foo.com/v1/depot"

type testData struct {
	packageName   string
	packages      []PackagesInfo
	expected      []string
	statusCode    int
	httpError     error
	expectedError error
}

func makeFakeHTTPClient(t *testing.T, data testData) *http.Client {
	count := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pkgsInfo := data.packages[count]
		JSON, err := json.Marshal(pkgsInfo)

		count += 1

		if err != nil {
			t.Fatalf("Unable to Marshal JSON for PackagesInfo: %v", err)
		}

		w.WriteHeader(data.statusCode)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, string(JSON))
	}))

	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			expectedUrl := fmt.Sprintf("%s/pkgs/%s", TEST_HAB_URL, data.packageName)
			if !strings.HasPrefix(req.URL.String(), expectedUrl) {
				t.Errorf("Requested URL is %s, but it should start with %s", req.URL, expectedUrl)
			}
			return url.Parse(server.URL)
		},
	}

	return &http.Client{Transport: transport}
}

func TestPackagesInfoFromName(t *testing.T) {
	tests := []testData{
		{
			packageName: "foo/test",
			packages: []PackagesInfo{
				PackagesInfo{
					RangeStart: 0,
					RangeEnd:   4,
					TotalCount: 3,
					PackageList: []PackageInfo{
						{
							Origin:  "foo",
							Name:    "test",
							Version: "0.0.1",
						},
						{
							Origin:  "foo",
							Name:    "test",
							Version: "0.0.2",
						},
						{
							Origin:  "foo",
							Name:    "test",
							Version: "0.1.0",
						},
					},
				},
			},
			expected:      []string{"0.0.1", "0.0.2", "0.1.0"},
			statusCode:    200,
			httpError:     nil,
			expectedError: nil,
		},
		{
			packageName: "foo/test",
			packages: []PackagesInfo{
				PackagesInfo{
					RangeStart: 0,
					RangeEnd:   4,
					TotalCount: 5,
					PackageList: []PackageInfo{
						{
							Origin:  "foo",
							Name:    "test",
							Version: "0.0.1",
						},
						{
							Origin:  "foo",
							Name:    "test",
							Version: "0.0.2",
						},
						{
							Origin:  "foo",
							Name:    "test",
							Version: "0.1.0",
						},
						{
							Origin:  "foo",
							Name:    "test",
							Version: "0.1.1",
						},
						{
							Origin:  "foo",
							Name:    "test",
							Version: "1.0.0",
						},
					},
				},
			},
			expected:      []string{"0.0.1", "0.0.2", "0.1.0", "0.1.1", "1.0.0"},
			statusCode:    200,
			httpError:     nil,
			expectedError: nil,
		},
		{
			packageName: "foo/test",
			packages: []PackagesInfo{
				PackagesInfo{
					RangeStart: 0,
					RangeEnd:   4,
					TotalCount: 8,
					PackageList: []PackageInfo{
						{
							Origin:  "foo",
							Name:    "test",
							Version: "0.0.1",
						},
						{
							Origin:  "foo",
							Name:    "test",
							Version: "0.0.2",
						},
						{
							Origin:  "foo",
							Name:    "test",
							Version: "0.1.0",
						},
						{
							Origin:  "foo",
							Name:    "test",
							Version: "0.1.1",
						},
						{
							Origin:  "foo",
							Name:    "test",
							Version: "1.0.0",
						},
					},
				},
				PackagesInfo{
					RangeStart: 5,
					RangeEnd:   9,
					TotalCount: 8,
					PackageList: []PackageInfo{
						{
							Origin:  "foo",
							Name:    "test",
							Version: "1.0.1",
						},
						{
							Origin:  "foo",
							Name:    "test",
							Version: "1.1.0",
						},
						{
							Origin:  "foo",
							Name:    "test",
							Version: "2.0.0",
						},
					},
				},
			},
			expected:      []string{"0.0.1", "0.0.2", "0.1.0", "0.1.1", "1.0.0", "1.0.1", "1.1.0", "2.0.0"},
			statusCode:    200,
			httpError:     nil,
			expectedError: nil,
		},
		{
			packageName: "foo/test",
			packages: []PackagesInfo{
				PackagesInfo{
					RangeStart:  0,
					RangeEnd:    0,
					TotalCount:  0,
					PackageList: []PackageInfo{},
				},
			},
			expected:      nil,
			statusCode:    404,
			httpError:     nil,
			expectedError: errors.New("Package not found"),
		},
		{
			packageName: "foo/test",
			packages: []PackagesInfo{
				PackagesInfo{
					RangeStart:  0,
					RangeEnd:    0,
					TotalCount:  0,
					PackageList: []PackageInfo{},
				},
			},
			expected:      nil,
			statusCode:    500,
			httpError:     nil,
			expectedError: errors.New("Unexpected status code: 500"),
		},
	}

	for _, test := range tests {
		http := makeFakeHTTPClient(t, test)
		testDepot := &depot{TEST_HAB_URL, http}

		results, err := testDepot.PackageVersionsFromName(test.packageName)

		if test.expectedError == nil && err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if test.expectedError != nil {
			if reflect.TypeOf(err) != reflect.TypeOf(test.expectedError) {
				t.Fatalf("Expected error type: %v, actual %v", reflect.TypeOf(test.expectedError), reflect.TypeOf(err))
			} else if err.Error() != test.expectedError.Error() {
				t.Fatalf("Expected error message %v, actual %v", test.expectedError, err)
			}
		} else {
			if !reflect.DeepEqual(results, test.expected) {
				t.Errorf("Get versions %v, but expected %v", results, test.expected)
			}
		}
	}
}
