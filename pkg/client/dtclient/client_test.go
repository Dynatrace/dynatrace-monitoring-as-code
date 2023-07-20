//go:build unit

/*
 * @license
 * Copyright 2023 Dynatrace LLC
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package dtclient

import (
	"context"
	"fmt"
	"github.com/dynatrace/dynatrace-configuration-as-code/v2/internal/cache"
	"github.com/dynatrace/dynatrace-configuration-as-code/v2/internal/concurrency"
	"github.com/dynatrace/dynatrace-configuration-as-code/v2/internal/idutils"
	"github.com/dynatrace/dynatrace-configuration-as-code/v2/internal/trafficlogs"
	"github.com/dynatrace/dynatrace-configuration-as-code/v2/internal/version"
	"github.com/dynatrace/dynatrace-configuration-as-code/v2/pkg/api"
	"github.com/dynatrace/dynatrace-configuration-as-code/v2/pkg/config/coordinate"
	"github.com/dynatrace/dynatrace-configuration-as-code/v2/pkg/rest"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

var mockAPI = api.API{ID: "mock-api", SingleConfiguration: true}
var mockAPINotSingle = api.API{ID: "mock-api", SingleConfiguration: false}

func TestNewClassicClient(t *testing.T) {
	t.Run("Client has correct urls and settings api path", func(t *testing.T) {
		client, err := NewClassicClient("https://some-url.com", nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, "https://some-url.com", client.environmentURL)
		assert.Equal(t, "https://some-url.com", client.environmentURLClassic)
		assert.Equal(t, settingsSchemaAPIPathClassic, client.settingsSchemaAPIPath)
		assert.Equal(t, settingsObjectAPIPathClassic, client.settingsObjectAPIPath)

	})

	t.Run("URL is empty - should throw an error", func(t *testing.T) {
		_, err := NewClassicClient("", nil)
		assert.ErrorContains(t, err, "empty url")

	})

	t.Run("invalid URL - should throw an error", func(t *testing.T) {
		_, err := NewClassicClient("INVALID_URL", nil)
		assert.ErrorContains(t, err, "not valid")

	})

	t.Run("URL suffix is trimmed", func(t *testing.T) {
		client, err := NewClassicClient("http://some-url.com/", nil)
		assert.NoError(t, err)
		assert.Equal(t, "http://some-url.com", client.environmentURL)
		assert.Equal(t, "http://some-url.com", client.environmentURLClassic)
	})

	t.Run("URL with leading space - should return an error", func(t *testing.T) {
		_, err := NewClassicClient(" https://my-environment.live.dynatrace.com/", nil)
		assert.Error(t, err)

	})

	t.Run("URL starts with http", func(t *testing.T) {
		client, err := NewClassicClient("http://some-url.com", nil)
		assert.NoError(t, err)
		assert.Equal(t, "http://some-url.com", client.environmentURL)
		assert.Equal(t, "http://some-url.com", client.environmentURLClassic)

	})

	t.Run("URL is without scheme - should throw an error", func(t *testing.T) {
		_, err := NewClassicClient("some-url.com", nil)
		assert.ErrorContains(t, err, "not valid")

	})

	t.Run("URL is without valid local path - should return an error", func(t *testing.T) {
		_, err := NewClassicClient("/my-environment/live/dynatrace.com/", nil)
		assert.ErrorContains(t, err, "no host specified")

	})

	t.Run("without valid protocol - should return an error", func(t *testing.T) {
		var err error

		_, err = NewClassicClient("https//my-environment.live.dynatrace.com/", nil)
		assert.ErrorContains(t, err, "not valid")
	})
}

func TestNewPlatformClient(t *testing.T) {

	t.Run("Client has correct urls and settings api path", func(t *testing.T) {
		client, err := NewPlatformClient("https://some-url.com", "https://some-url2.com", nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, "https://some-url.com", client.environmentURL)
		assert.Equal(t, "https://some-url2.com", client.environmentURLClassic)
		assert.Equal(t, settingsSchemaAPIPathPlatform, client.settingsSchemaAPIPath)
		assert.Equal(t, settingsObjectAPIPathPlatform, client.settingsObjectAPIPath)

	})

	t.Run("URL is empty - should throw an error", func(t *testing.T) {
		_, err := NewPlatformClient("", "", nil, nil)
		assert.ErrorContains(t, err, "empty url")

		_, err = NewPlatformClient("http://some-url.com", "", nil, nil)
		assert.ErrorContains(t, err, "empty url")
	})

	t.Run("invalid URL - should throw an error", func(t *testing.T) {
		_, err := NewPlatformClient("INVALID_URL", "", nil, nil)
		assert.ErrorContains(t, err, "not valid")

		_, err = NewPlatformClient("http://some-url.com", "INVALID_URL", nil, nil)
		assert.ErrorContains(t, err, "not valid")
	})

	t.Run("URL suffix is trimmed", func(t *testing.T) {
		client, err := NewPlatformClient("http://some-url.com/", "http://some-url2.com/", nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, "http://some-url.com", client.environmentURL)
		assert.Equal(t, "http://some-url2.com", client.environmentURLClassic)
	})

	t.Run("URL with leading space - should return an error", func(t *testing.T) {
		_, err := NewPlatformClient(" https://my-environment.live.dynatrace.com/", "", nil, nil)
		assert.Error(t, err)

		_, err = NewPlatformClient("https://my-environment.live.dynatrace.com/", " https://my-environment.live.dynatrace.com/\"", nil, nil)
		assert.Error(t, err)
	})

	t.Run("URL starts with http", func(t *testing.T) {
		client, err := NewPlatformClient("http://some-url.com", "https://some-url.com", nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, "http://some-url.com", client.environmentURL)

		client, err = NewPlatformClient("https://my-environment.live.dynatrace.com/", "http://some-url.com", nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, "http://some-url.com", client.environmentURLClassic)
	})

	t.Run("URL is without scheme - should throw an error", func(t *testing.T) {
		_, err := NewPlatformClient("some-url.com", "", nil, nil)
		assert.ErrorContains(t, err, "not valid")

		_, err = NewPlatformClient("https://some-url.com", "some-url.com", nil, nil)
		assert.ErrorContains(t, err, "not valid")
	})

	t.Run("URL is without valid local path - should return an error", func(t *testing.T) {
		_, err := NewPlatformClient("/my-environment/live/dynatrace.com/", "https://some-url.com", nil, nil)
		assert.ErrorContains(t, err, "no host specified")

		_, err = NewPlatformClient("https://some-url.com", "/my-environment/live/dynatrace.com/", nil, nil)
		assert.ErrorContains(t, err, "no host specified")
	})

	t.Run("without valid protocol - should return an error", func(t *testing.T) {
		var err error

		_, err = NewPlatformClient("https//my-environment.live.dynatrace.com/", "", nil, nil)
		assert.ErrorContains(t, err, "not valid")

		_, err = NewPlatformClient("http//my-environment.live.dynatrace.com/", "", nil, nil)
		assert.ErrorContains(t, err, "not valid")
	})
}

func TestReadByIdReturnsAnErrorUponEncounteringAnError(t *testing.T) {
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		http.Error(res, "", http.StatusForbidden)
	}))
	defer func() { testServer.Close() }()
	client := DynatraceClient{
		environmentURLClassic: testServer.URL,
		classicClient:         rest.NewRestClient(testServer.Client(), trafficlogs.NewFileBased(), rest.CreateRateLimitStrategy()),
		limiter:               concurrency.NewLimiter(5),
		generateExternalID:    idutils.GenerateExternalID,
	}

	_, err := client.ReadConfigById(mockAPI, "test")
	assert.ErrorContains(t, err, "Response was")
}

func TestReadByIdEscapesTheId(t *testing.T) {
	unescapedID := "ruxit.perfmon.dotnetV4:%TimeInGC:time_in_gc_alert_high_generic"

	testServer := httptest.NewTLSServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {}))
	defer func() { testServer.Close() }()
	client := DynatraceClient{
		environmentURLClassic: testServer.URL,
		classicClient:         rest.NewRestClient(testServer.Client(), nil, rest.CreateRateLimitStrategy()),
		limiter:               concurrency.NewLimiter(5),
		generateExternalID:    idutils.GenerateExternalID,
	}
	_, err := client.ReadConfigById(mockAPINotSingle, unescapedID)
	assert.NoError(t, err)
}

func TestReadByIdReturnsTheResponseGivenNoError(t *testing.T) {
	body := []byte{1, 3, 3, 7}

	testServer := httptest.NewTLSServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		_, _ = res.Write(body)
	}))
	defer func() { testServer.Close() }()

	client := DynatraceClient{
		environmentURLClassic: testServer.URL,
		classicClient:         rest.NewRestClient(testServer.Client(), nil, rest.CreateRateLimitStrategy()),
		limiter:               concurrency.NewLimiter(5),
		generateExternalID:    idutils.GenerateExternalID,
	}

	resp, err := client.ReadConfigById(mockAPI, "test")
	assert.NoError(t, err, "there should not be an error")
	assert.Equal(t, body, resp)
}

func TestListKnownSettings(t *testing.T) {

	tests := []struct {
		name                      string
		givenSchemaID             string
		givenListSettingsOpts     ListSettingsOptions
		givenServerResponses      []testServerResponse
		want                      []DownloadSettingsObject
		wantQueryParamsPerAPICall [][]testQueryParams
		wantNumberOfAPICalls      int
		wantError                 bool
	}{
		{
			name:          "Lists Settings objects as expected",
			givenSchemaID: "builtin:something",
			givenServerResponses: []testServerResponse{
				{200, `{ "items": [ {"objectId": "f5823eca-4838-49d0-81d9-0514dd2c4640", "externalId": "RG9jdG9yIFdobwo="} ] }`},
			},
			want: []DownloadSettingsObject{
				{
					ExternalId: "RG9jdG9yIFdobwo=",
					ObjectId:   "f5823eca-4838-49d0-81d9-0514dd2c4640",
				},
			},
			wantQueryParamsPerAPICall: [][]testQueryParams{
				{
					{"schemaIds", "builtin:something"},
					{"pageSize", "500"},
					{"fields", defaultListSettingsFields},
				},
			},
			wantNumberOfAPICalls: 1,
			wantError:            false,
		},
		{
			name:                  "Lists Settings objects without value field as expected",
			givenSchemaID:         "builtin:something",
			givenListSettingsOpts: ListSettingsOptions{DiscardValue: true},
			givenServerResponses: []testServerResponse{
				{200, `{ "items": [ {"objectId": "f5823eca-4838-49d0-81d9-0514dd2c4640", "externalId": "RG9jdG9yIFdobwo="} ] }`},
			},
			want: []DownloadSettingsObject{
				{
					ExternalId: "RG9jdG9yIFdobwo=",
					ObjectId:   "f5823eca-4838-49d0-81d9-0514dd2c4640",
				},
			},
			wantQueryParamsPerAPICall: [][]testQueryParams{
				{
					{"schemaIds", "builtin:something"},
					{"pageSize", "500"},
					{"fields", reducedListSettingsFields},
				},
			},
			wantNumberOfAPICalls: 1,
			wantError:            false,
		},
		{
			name:          "Lists Settings objects with filter as expected",
			givenSchemaID: "builtin:something",
			givenListSettingsOpts: ListSettingsOptions{Filter: func(o DownloadSettingsObject) bool {
				return o.ExternalId == "RG9jdG9yIFdobwo="
			}},
			givenServerResponses: []testServerResponse{
				{200, `{ "items": [ {"objectId": "f5823eca-4838-49d0-81d9-0514dd2c4640", "externalId": "RG9jdG9yIFdobwo="} ] }`},
				{200, `{ "items": [ {"objectId": "f5823eca-4838-49d0-81d9-0514dd2c4641", "externalId": "RG9jdG9yIabcdef="} ] }`},
			},
			want: []DownloadSettingsObject{
				{
					ExternalId: "RG9jdG9yIFdobwo=",
					ObjectId:   "f5823eca-4838-49d0-81d9-0514dd2c4640",
				},
			},
			wantQueryParamsPerAPICall: [][]testQueryParams{
				{
					{"schemaIds", "builtin:something"},
					{"pageSize", "500"},
					{"fields", defaultListSettingsFields},
				},
			},
			wantNumberOfAPICalls: 1,
			wantError:            false,
		},
		{
			name:          "Handles Pagination when listing settings objects",
			givenSchemaID: "builtin:something",
			givenServerResponses: []testServerResponse{
				{200, `{ "items": [ {"objectId": "f5823eca-4838-49d0-81d9-0514dd2c4640", "externalId": "RG9jdG9yIFdobwo="} ], "nextPageKey": "page42" }`},
				{200, `{ "items": [ {"objectId": "b1d4c623-25e0-4b54-9eb5-6734f1a72041", "externalId": "VGhlIE1hc3Rlcgo="} ] }`},
			},
			want: []DownloadSettingsObject{
				{
					ExternalId: "RG9jdG9yIFdobwo=",
					ObjectId:   "f5823eca-4838-49d0-81d9-0514dd2c4640",
				},
				{
					ExternalId: "VGhlIE1hc3Rlcgo=",
					ObjectId:   "b1d4c623-25e0-4b54-9eb5-6734f1a72041",
				},
			},

			wantQueryParamsPerAPICall: [][]testQueryParams{
				{
					{"schemaIds", "builtin:something"},
					{"pageSize", "500"},
					{"fields", defaultListSettingsFields},
				},
				{
					{"nextPageKey", "page42"},
				},
			},
			wantNumberOfAPICalls: 2,
			wantError:            false,
		},
		{
			name:          "Returns empty if list if no items exist",
			givenSchemaID: "builtin:something",
			givenServerResponses: []testServerResponse{
				{200, `{ "items": [ ] }`},
			},
			want: []DownloadSettingsObject{},
			wantQueryParamsPerAPICall: [][]testQueryParams{
				{
					{"schemaIds", "builtin:something"},
					{"pageSize", "500"},
					{"fields", defaultListSettingsFields},
				},
			},
			wantNumberOfAPICalls: 1,
			wantError:            false,
		},
		{
			name:          "Returns error if HTTP error is encountered - 400",
			givenSchemaID: "builtin:something",
			givenServerResponses: []testServerResponse{
				{400, `epic fail`},
			},
			want: nil,
			wantQueryParamsPerAPICall: [][]testQueryParams{
				{
					{"schemaIds", "builtin:something"},
					{"pageSize", "500"},
					{"fields", defaultListSettingsFields},
				},
			},
			wantNumberOfAPICalls: 1,
			wantError:            true,
		},
		{
			name:          "Returns error if HTTP error is encountered - 403",
			givenSchemaID: "builtin:something",
			givenServerResponses: []testServerResponse{
				{403, `epic fail`},
			},
			want: nil,
			wantQueryParamsPerAPICall: [][]testQueryParams{
				{
					{"schemaIds", "builtin:something"},
					{"pageSize", "500"},
					{"fields", defaultListSettingsFields},
				},
			},
			wantNumberOfAPICalls: 1,
			wantError:            true,
		},
		{
			name:          "Retries on HTTP error on paginated request and returns eventual success",
			givenSchemaID: "builtin:something",
			givenServerResponses: []testServerResponse{
				{200, `{ "items": [ {"objectId": "f5823eca-4838-49d0-81d9-0514dd2c4640", "externalId": "RG9jdG9yIFdobwo="} ], "nextPageKey": "page42" }`},
				{400, `get next page fail`},
				{400, `retry fail`},
				{200, `{ "items": [ {"objectId": "b1d4c623-25e0-4b54-9eb5-6734f1a72041", "externalId": "VGhlIE1hc3Rlcgo="} ] }`},
			},
			want: []DownloadSettingsObject{
				{
					ExternalId: "RG9jdG9yIFdobwo=",
					ObjectId:   "f5823eca-4838-49d0-81d9-0514dd2c4640",
				},
				{
					ExternalId: "VGhlIE1hc3Rlcgo=",
					ObjectId:   "b1d4c623-25e0-4b54-9eb5-6734f1a72041",
				},
			},
			wantQueryParamsPerAPICall: [][]testQueryParams{
				{
					{"schemaIds", "builtin:something"},
					{"pageSize", "500"},
					{"fields", defaultListSettingsFields},
				},
				{
					{"nextPageKey", "page42"},
				},
				{
					{"nextPageKey", "page42"},
				},
				{
					{"nextPageKey", "page42"},
				},
			},
			wantNumberOfAPICalls: 4,
			wantError:            false,
		},
		{
			name:          "Returns error if HTTP error is encountered getting further paginated responses",
			givenSchemaID: "builtin:something",
			givenServerResponses: []testServerResponse{
				{200, `{ "items": [ {"objectId": "f5823eca-4838-49d0-81d9-0514dd2c4640", "externalId": "RG9jdG9yIFdobwo="} ], "nextPageKey": "page42" }`},
				{400, `get next page fail`},
				{400, `retry fail 1`},
				{400, `retry fail 2`},
				{400, `retry fail 3`},
			},
			want: nil,
			wantQueryParamsPerAPICall: [][]testQueryParams{
				{
					{"schemaIds", "builtin:something"},
					{"pageSize", "500"},
					{"fields", defaultListSettingsFields},
				},
				{
					{"nextPageKey", "page42"},
				},
				{
					{"nextPageKey", "page42"},
				},
				{
					{"nextPageKey", "page42"},
				},
				{
					{"nextPageKey", "page42"},
				},
			},
			wantNumberOfAPICalls: 5,
			wantError:            true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiCalls := 0
			server := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				if len(tt.wantQueryParamsPerAPICall) > 0 {
					params := tt.wantQueryParamsPerAPICall[apiCalls]
					for _, param := range params {
						addedQueryParameter := req.URL.Query()[param.key]
						assert.NotNil(t, addedQueryParameter)
						assert.NotEmpty(t, addedQueryParameter)
						assert.Equal(t, addedQueryParameter[0], param.value)
					}
				} else {
					assert.Equal(t, "", req.URL.RawQuery, "expected no query params - but '%s' was sent", req.URL.RawQuery)
				}

				resp := tt.givenServerResponses[apiCalls]
				if resp.statusCode != 200 {
					http.Error(rw, resp.body, resp.statusCode)
				} else {
					_, _ = rw.Write([]byte(resp.body))
				}

				apiCalls++
				assert.LessOrEqualf(t, apiCalls, tt.wantNumberOfAPICalls, "expected at most %d API calls to happen, but encountered call %d", tt.wantNumberOfAPICalls, apiCalls)
			}))
			defer server.Close()

			restClient := rest.NewRestClient(server.Client(), nil, rest.CreateRateLimitStrategy())
			client, _ := NewClassicClient(server.URL, restClient,
				WithRetrySettings(testRetrySettings),
				WithClientRequestLimiter(concurrency.NewLimiter(5)),
				WithExternalIDGenerator(idutils.GenerateExternalID))

			res, err1 := client.ListSettings(context.TODO(), tt.givenSchemaID, tt.givenListSettingsOpts)

			if tt.wantError {
				assert.Error(t, err1)
			} else {
				assert.NoError(t, err1)
			}

			assert.Equal(t, tt.want, res)

			assert.Equal(t, apiCalls, tt.wantNumberOfAPICalls, "expected exactly %d API calls to happen but %d calls where made", tt.wantNumberOfAPICalls, apiCalls)
		})
	}
}

func TestGetSettingById(t *testing.T) {
	type fields struct {
		environmentURL string
		retrySettings  rest.RetrySettings
	}
	type args struct {
		objectID string
	}
	tests := []struct {
		name                string
		fields              fields
		args                args
		givenTestServerResp *testServerResponse
		wantURLPath         string
		wantResult          *DownloadSettingsObject
		wantErr             bool
	}{
		{
			name:   "Get Setting by ID - server response != 2xx",
			fields: fields{},
			args: args{
				objectID: "12345",
			},
			givenTestServerResp: &testServerResponse{
				statusCode: 500,
				body:       "{}",
			},
			wantURLPath: "/api/v2/settings/objects/12345",
			wantResult:  nil,
			wantErr:     true,
		},
		{
			name:   "Get Setting by ID - invalid server response",
			fields: fields{},
			args: args{
				objectID: "12345",
			},
			givenTestServerResp: &testServerResponse{
				statusCode: 200,
				body:       `{bs}`,
			},
			wantURLPath: "/api/v2/settings/objects/12345",
			wantResult:  nil,
			wantErr:     true,
		},
		{
			name:   "Get Setting by ID",
			fields: fields{},
			args: args{
				objectID: "12345",
			},
			givenTestServerResp: &testServerResponse{
				statusCode: 200,
				body:       `{"objectId":"12345","externalId":"54321", "schemaVersion":"1.0","schemaId":"builtin:bla","scope":"tenant"}`,
			},
			wantURLPath: "/api/v2/settings/objects/12345",
			wantResult: &DownloadSettingsObject{
				ExternalId:    "54321",
				SchemaVersion: "1.0",
				SchemaId:      "builtin:bla",
				ObjectId:      "12345",
				Scope:         "tenant",
				Value:         nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			server := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				assert.Equal(t, tt.wantURLPath, req.URL.Path)
				if resp := tt.givenTestServerResp; resp != nil {
					if resp.statusCode != 200 {
						http.Error(rw, resp.body, resp.statusCode)
					} else {
						_, _ = rw.Write([]byte(resp.body))
					}
				}

			}))
			defer server.Close()

			var envURL string
			if tt.fields.environmentURL != "" {
				envURL = tt.fields.environmentURL
			} else {
				envURL = server.URL
			}

			d := DynatraceClient{
				environmentURL:        envURL,
				platformClient:        rest.NewRestClient(server.Client(), nil, rest.CreateRateLimitStrategy()),
				retrySettings:         tt.fields.retrySettings,
				settingsObjectAPIPath: "/api/v2/settings/objects",
				limiter:               concurrency.NewLimiter(5),
				generateExternalID:    idutils.GenerateExternalID,
			}

			settingsObj, err := d.GetSettingById(tt.args.objectID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantResult, settingsObj)

		})
	}

}

func TestDeleteSettings(t *testing.T) {
	type fields struct {
		environmentURL string
		retrySettings  rest.RetrySettings
	}
	type args struct {
		objectID string
	}
	tests := []struct {
		name                string
		fields              fields
		args                args
		givenTestServerResp *testServerResponse
		wantURLPath         string
		wantErr             bool
	}{
		{
			name:   "Delete Settings - server response != 2xx",
			fields: fields{},
			args: args{
				objectID: "12345",
			},
			givenTestServerResp: &testServerResponse{
				statusCode: 500,
				body:       "{}",
			},
			wantURLPath: "/api/v2/settings/objects/12345",
			wantErr:     true,
		},
		{
			name:   "Delete Settings - server response 404 does not result in an err",
			fields: fields{},
			args: args{
				objectID: "12345",
			},
			givenTestServerResp: &testServerResponse{
				statusCode: 404,
				body:       "{}",
			},
			wantURLPath: "/api/v2/settings/objects/12345",
			wantErr:     false,
		},
		{
			name:   "Delete Settings - object ID is passed",
			fields: fields{},
			args: args{
				objectID: "12345",
			},
			wantURLPath: "/api/v2/settings/objects/12345",
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			server := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				assert.Equal(t, tt.wantURLPath, req.URL.Path)
				if resp := tt.givenTestServerResp; resp != nil {
					if resp.statusCode != 200 {
						http.Error(rw, resp.body, resp.statusCode)
					} else {
						_, _ = rw.Write([]byte(resp.body))
					}
				}

			}))
			defer server.Close()

			var envURL string
			if tt.fields.environmentURL != "" {
				envURL = tt.fields.environmentURL
			} else {
				envURL = server.URL
			}

			d := DynatraceClient{
				environmentURL:        envURL,
				platformClient:        rest.NewRestClient(server.Client(), nil, rest.CreateRateLimitStrategy()),
				retrySettings:         tt.fields.retrySettings,
				settingsObjectAPIPath: settingsObjectAPIPathClassic,
				limiter:               concurrency.NewLimiter(5),
				generateExternalID:    idutils.GenerateExternalID,
			}

			if err := d.DeleteSettings(tt.args.objectID); (err != nil) != tt.wantErr {
				t.Errorf("DeleteSettings() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUpsertSettingsRetries(t *testing.T) {
	numAPICalls := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodGet {
			rw.WriteHeader(200)
			_, _ = rw.Write([]byte("{}"))
			return
		}

		numAPICalls++
		if numAPICalls < 3 {
			rw.WriteHeader(409)
			return
		}
		rw.WriteHeader(200)
		_, _ = rw.Write([]byte(`[{"objectId": "abcdefg"}]`))
	}))
	defer server.Close()

	restClient := rest.NewRestClient(server.Client(), nil, rest.CreateRateLimitStrategy())
	client, _ := NewPlatformClient(server.URL, server.URL, restClient, restClient,
		WithRetrySettings(testRetrySettings),
		WithClientRequestLimiter(concurrency.NewLimiter(5)),
		WithExternalIDGenerator(idutils.GenerateExternalID))

	_, err := client.UpsertSettings(context.TODO(), SettingsObject{
		Coordinate: coordinate.Coordinate{Type: "some:schema", ConfigId: "id"},
		SchemaId:   "some:schema",
		Content:    []byte("{}"),
	})

	assert.NoError(t, err)
	assert.Equal(t, numAPICalls, 3)
}

func TestUpsertSettingsFromCache(t *testing.T) {
	numAPIGetCalls := 0
	numAPIPostCalls := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/some:schema" {
			rw.WriteHeader(http.StatusOK)
			rw.Write([]byte("{}"))
			return
		}
		if req.Method == http.MethodGet {
			numAPIGetCalls++
			rw.WriteHeader(200)
			rw.Write([]byte("{}"))
			return
		}

		numAPIPostCalls++
		rw.WriteHeader(200)
		rw.Write([]byte(`[{"objectId": "abcdefg"}]`))
	}))
	defer server.Close()

	restClient := rest.NewRestClient(server.Client(), nil, rest.CreateRateLimitStrategy())
	client, _ := NewPlatformClient(server.URL, server.URL, restClient, restClient,
		WithRetrySettings(testRetrySettings),
		WithClientRequestLimiter(concurrency.NewLimiter(5)),
		WithExternalIDGenerator(idutils.GenerateExternalID))

	_, err := client.UpsertSettings(context.TODO(), SettingsObject{
		Coordinate: coordinate.Coordinate{Type: "some:schema", ConfigId: "id"},
		SchemaId:   "some:schema",
		Content:    []byte("{}"),
	})

	assert.NoError(t, err)
	assert.Equal(t, 1, numAPIGetCalls)
	assert.Equal(t, 1, numAPIPostCalls)

	_, err = client.UpsertSettings(context.TODO(), SettingsObject{
		Coordinate: coordinate.Coordinate{Type: "some:schema", ConfigId: "id"},
		SchemaId:   "some:schema",
		Content:    []byte("{}"),
	})

	assert.NoError(t, err)
	assert.Equal(t, 1, numAPIGetCalls) // still one
	assert.Equal(t, 2, numAPIPostCalls)
}

func TestUpsertSettingsFromCache_CacheInvalidated(t *testing.T) {
	numGetAPICalls := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/some:schema" {
			rw.WriteHeader(http.StatusOK)
			rw.Write([]byte("{}"))
			return
		}
		if req.Method == http.MethodGet {
			numGetAPICalls++
			rw.WriteHeader(200)
			_, _ = rw.Write([]byte("{}"))
			return
		}

		rw.WriteHeader(409)
		rw.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := DynatraceClient{
		environmentURL:         server.URL,
		platformClient:         rest.NewRestClient(server.Client(), nil, rest.CreateRateLimitStrategy()),
		retrySettings:          testRetrySettings,
		limiter:                concurrency.NewLimiter(5),
		generateExternalID:     idutils.GenerateExternalID,
		settingsCache:          &cache.DefaultCache[[]DownloadSettingsObject]{},
		classicConfigsCache:    &cache.DefaultCache[[]Value]{},
		schemaConstraintsCache: &cache.DefaultCache[SchemaConstraints]{},
	}

	client.UpsertSettings(context.TODO(), SettingsObject{
		Coordinate: coordinate.Coordinate{Type: "some:schema", ConfigId: "id"},
		SchemaId:   "some:schema",
		Content:    []byte("{}"),
	})
	assert.Equal(t, 1, numGetAPICalls)

	client.UpsertSettings(context.TODO(), SettingsObject{
		Coordinate: coordinate.Coordinate{Type: "some:schema", ConfigId: "id"},
		SchemaId:   "some:schema",
		Content:    []byte("{}"),
	})
	assert.Equal(t, 2, numGetAPICalls)

	client.UpsertSettings(context.TODO(), SettingsObject{
		Coordinate: coordinate.Coordinate{Type: "some:schema", ConfigId: "id"},
		SchemaId:   "some:schema",
		Content:    []byte("{}"),
	})
	assert.Equal(t, 3, numGetAPICalls)

}

func TestListEntities(t *testing.T) {

	testType := "SOMETHING"

	tests := []struct {
		name                      string
		givenEntitiesType         EntitiesType
		givenServerResponses      []testServerResponse
		want                      []string
		wantQueryParamsPerAPICall [][]testQueryParams
		wantNumberOfAPICalls      int
		wantError                 bool
	}{
		{
			name:              "Lists Entities objects as expected",
			givenEntitiesType: EntitiesType{EntitiesTypeId: testType},
			givenServerResponses: []testServerResponse{
				{200, fmt.Sprintf(`{ "entities": [ {"entityId": "%s-1A28B791C329D741", "type": "%s"} ] }`, testType, testType)},
			},
			want: []string{
				fmt.Sprintf(`{"entityId": "%s-1A28B791C329D741", "type": "%s"}`, testType, testType),
			},
			wantQueryParamsPerAPICall: [][]testQueryParams{
				{
					{"entitySelector", fmt.Sprintf(`type("%s")`, testType)},
					{"pageSize", defaultPageSizeEntities},
					{"fields", defaultListEntitiesFields},
				},
			},
			wantNumberOfAPICalls: 1,
			wantError:            false,
		},
		{
			name:              "Handles Pagination when listing entities objects",
			givenEntitiesType: EntitiesType{EntitiesTypeId: testType},
			givenServerResponses: []testServerResponse{
				{200, fmt.Sprintf(`{ "entities": [ {"entityId": "%s-1A28B791C329D741", "type": "%s"} ], "nextPageKey": "page42"  }`, testType, testType)},
				{200, fmt.Sprintf(`{ "entities": [ {"entityId": "%s-C329D7411A28B791", "type": "%s"} ] }`, testType, testType)},
			},
			want: []string{
				fmt.Sprintf(`{"entityId": "%s-1A28B791C329D741", "type": "%s"}`, testType, testType),
				fmt.Sprintf(`{"entityId": "%s-C329D7411A28B791", "type": "%s"}`, testType, testType),
			},

			wantQueryParamsPerAPICall: [][]testQueryParams{
				{
					{"entitySelector", fmt.Sprintf(`type("%s")`, testType)},
					{"pageSize", defaultPageSizeEntities},
					{"fields", defaultListEntitiesFields},
				},
				{
					{"nextPageKey", "page42"},
				},
			},
			wantNumberOfAPICalls: 2,
			wantError:            false,
		},
		{
			name:              "Returns empty if list if no entities exist",
			givenEntitiesType: EntitiesType{EntitiesTypeId: testType},
			givenServerResponses: []testServerResponse{
				{200, `{ "entities": [ ] }`},
			},
			want: []string{},
			wantQueryParamsPerAPICall: [][]testQueryParams{
				{
					{"entitySelector", fmt.Sprintf(`type("%s")`, testType)},
					{"pageSize", defaultPageSizeEntities},
					{"fields", defaultListEntitiesFields},
				},
			},
			wantNumberOfAPICalls: 1,
			wantError:            false,
		},
		{
			name:              "Returns error if HTTP error is encountered",
			givenEntitiesType: EntitiesType{EntitiesTypeId: testType},
			givenServerResponses: []testServerResponse{
				{400, `epic fail`},
			},
			want: nil,
			wantQueryParamsPerAPICall: [][]testQueryParams{
				{
					{"entitySelector", fmt.Sprintf(`type("%s")`, testType)},
					{"pageSize", defaultPageSizeEntities},
					{"fields", defaultListEntitiesFields},
				},
			},
			wantNumberOfAPICalls: 1,
			wantError:            true,
		},
		{
			name:              "Retries on HTTP error on paginated request and returns eventual success",
			givenEntitiesType: EntitiesType{EntitiesTypeId: testType},
			givenServerResponses: []testServerResponse{
				{200, fmt.Sprintf(`{ "entities": [ {"entityId": "%s-1A28B791C329D741", "type": "%s"} ], "nextPageKey": "page42"  }`, testType, testType)},
				{400, `get next page fail`},
				{400, `retry fail`},
				{200, fmt.Sprintf(`{ "entities": [ {"entityId": "%s-C329D7411A28B791", "type": "%s"} ] }`, testType, testType)},
			},
			want: []string{
				fmt.Sprintf(`{"entityId": "%s-1A28B791C329D741", "type": "%s"}`, testType, testType),
				fmt.Sprintf(`{"entityId": "%s-C329D7411A28B791", "type": "%s"}`, testType, testType),
			},
			wantQueryParamsPerAPICall: [][]testQueryParams{
				{
					{"entitySelector", fmt.Sprintf(`type("%s")`, testType)},
					{"pageSize", defaultPageSizeEntities},
					{"fields", defaultListEntitiesFields},
				},
				{
					{"nextPageKey", "page42"},
				},
				{
					{"nextPageKey", "page42"},
				},
				{
					{"nextPageKey", "page42"},
				},
			},
			wantNumberOfAPICalls: 4,
			wantError:            false,
		},
		{
			name:              "Returns error if HTTP error is encountered getting further paginated responses",
			givenEntitiesType: EntitiesType{EntitiesTypeId: testType},
			givenServerResponses: []testServerResponse{
				{200, fmt.Sprintf(`{ "entities": [ {"entityId": "%s-1A28B791C329D741", "type": "%s"} ], "nextPageKey": "page42"  }`, testType, testType)},
				{400, `get next page fail`},
				{400, `retry fail 1`},
				{400, `retry fail 2`},
				{400, `retry fail 3`},
			},
			want: nil,
			wantQueryParamsPerAPICall: [][]testQueryParams{
				{
					{"entitySelector", fmt.Sprintf(`type("%s")`, testType)},
					{"pageSize", defaultPageSizeEntities},
					{"fields", defaultListEntitiesFields},
				},
				{
					{"nextPageKey", "page42"},
				},
				{
					{"nextPageKey", "page42"},
				},
				{
					{"nextPageKey", "page42"},
				},
				{
					{"nextPageKey", "page42"},
				},
			},
			wantNumberOfAPICalls: 5,
			wantError:            true,
		},
		{
			name:              "Retries on empty paginated response",
			givenEntitiesType: EntitiesType{EntitiesTypeId: testType},
			givenServerResponses: []testServerResponse{
				{200, fmt.Sprintf(`{ "entities": [ {"entityId": "%s-1A28B791C329D741", "type": "%s"} ], "nextPageKey": "page42"  }`, testType, testType)},
				{200, fmt.Sprintf(`{ "entities": [] }`)},
				{200, fmt.Sprintf(`{ "entities": [] }`)},
				{200, fmt.Sprintf(`{ "entities": [ {"entityId": "%s-C329D7411A28B791", "type": "%s"} ] }`, testType, testType)},
			},
			want: []string{
				fmt.Sprintf(`{"entityId": "%s-1A28B791C329D741", "type": "%s"}`, testType, testType),
				fmt.Sprintf(`{"entityId": "%s-C329D7411A28B791", "type": "%s"}`, testType, testType),
			},
			wantQueryParamsPerAPICall: [][]testQueryParams{
				{
					{"entitySelector", fmt.Sprintf(`type("%s")`, testType)},
					{"pageSize", defaultPageSizeEntities},
					{"fields", defaultListEntitiesFields},
				},
				{
					{"nextPageKey", "page42"},
				},
				{
					{"nextPageKey", "page42"},
				},
				{
					{"nextPageKey", "page42"},
				},
			},
			wantNumberOfAPICalls: 4,
			wantError:            false,
		},
		{
			name:              "Retries on wrong field for entity type",
			givenEntitiesType: EntitiesType{EntitiesTypeId: testType},
			givenServerResponses: []testServerResponse{
				{200, fmt.Sprintf(`{ "entities": [ {"entityId": "%s-1A28B791C329D741", "type": "%s"} ], "nextPageKey": "page42"  }`, testType, testType)},
				{400, fmt.Sprintf(`{{
					"error":{
						"code":400,
						"message":"Constraints violated.",
						"constraintViolations":[{
							"path":"fields",
							"message":"'ipAddress' is not a valid property for type '%s'",
							"parameterLocation":"QUERY",
							"location":null
						}]
					}
				}
				}`, testType)},
				{200, fmt.Sprintf(`{ "entities": [ {"entityId": "%s-C329D7411A28B791", "type": "%s"} ] }`, testType, testType)},
			},
			want: []string{
				fmt.Sprintf(`{"entityId": "%s-1A28B791C329D741", "type": "%s"}`, testType, testType),
				fmt.Sprintf(`{"entityId": "%s-C329D7411A28B791", "type": "%s"}`, testType, testType),
			},
			wantQueryParamsPerAPICall: [][]testQueryParams{
				{
					{"entitySelector", fmt.Sprintf(`type("%s")`, testType)},
					{"pageSize", defaultPageSizeEntities},
					{"fields", defaultListEntitiesFields},
				},
				{
					{"nextPageKey", "page42"},
				},
				{
					{"nextPageKey", "page42"},
				},
			},
			wantNumberOfAPICalls: 3,
			wantError:            false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiCalls := 0
			server := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				if len(tt.wantQueryParamsPerAPICall) > 0 {
					params := tt.wantQueryParamsPerAPICall[apiCalls]
					for _, param := range params {
						addedQueryParameter := req.URL.Query()[param.key]
						assert.NotNil(t, addedQueryParameter)
						assert.Greater(t, len(addedQueryParameter), 0)
						assert.Equal(t, addedQueryParameter[0], param.value)
					}
				} else {
					assert.Equal(t, "", req.URL.RawQuery, "expected no query params - but '%s' was sent", req.URL.RawQuery)
				}

				resp := tt.givenServerResponses[apiCalls]
				if resp.statusCode != 200 {
					http.Error(rw, resp.body, resp.statusCode)
				} else {
					_, _ = rw.Write([]byte(resp.body))
				}

				apiCalls++
				assert.LessOrEqualf(t, apiCalls, tt.wantNumberOfAPICalls, "expected at most %d API calls to happen, but encountered call %d", tt.wantNumberOfAPICalls, apiCalls)
			}))
			defer server.Close()

			client := DynatraceClient{
				environmentURL:     server.URL,
				platformClient:     rest.NewRestClient(server.Client(), nil, rest.CreateRateLimitStrategy()),
				retrySettings:      testRetrySettings,
				limiter:            concurrency.NewLimiter(5),
				generateExternalID: idutils.GenerateExternalID,
			}

			res, err1 := client.ListEntities(context.TODO(), tt.givenEntitiesType)

			if tt.wantError {
				assert.Error(t, err1)
			} else {
				assert.NoError(t, err1)
			}

			assert.Equal(t, tt.want, res)

			assert.Equal(t, apiCalls, tt.wantNumberOfAPICalls, "expected exactly %d API calls to happen but %d calls where made", tt.wantNumberOfAPICalls, apiCalls)
		})
	}
}

func TestCreateDynatraceClientWithAutoServerVersion(t *testing.T) {
	t.Run("Server version is correctly set to determined value", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			_, _ = rw.Write([]byte(`{"version" : "1.262.0.20230214-193525"}`))
		}))

		dcl, err := NewClassicClient(server.URL, rest.NewRestClient(server.Client(), nil, rest.CreateRateLimitStrategy()), WithAutoServerVersion())

		server.Close()
		assert.NoError(t, err)
		assert.Equal(t, version.Version{Major: 1, Minor: 262}, dcl.serverVersion)
	})

	t.Run("Server version is correctly set to unknown", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			_, _ = rw.Write([]byte(`{}`))
		}))

		dcl, err := NewClassicClient(server.URL, rest.NewRestClient(server.Client(), nil, rest.CreateRateLimitStrategy()), WithAutoServerVersion())
		server.Close()
		assert.NoError(t, err)
		assert.Equal(t, version.UnknownVersion, dcl.serverVersion)
	})
}
