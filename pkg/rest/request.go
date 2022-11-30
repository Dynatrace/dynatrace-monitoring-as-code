/**
 * @license
 * Copyright 2020 Dynatrace LLC
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

package rest

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"time"

	"github.com/dynatrace-oss/dynatrace-monitoring-as-code/pkg/util"
	"github.com/dynatrace-oss/dynatrace-monitoring-as-code/pkg/util/log"
	"github.com/dynatrace-oss/dynatrace-monitoring-as-code/pkg/version"
	"github.com/google/uuid"
)

type Response struct {
	StatusCode int
	Body       []byte
	Headers    map[string][]string
}

func get(client *http.Client, url string, apiToken string) (Response, error) {
	req, err := request(http.MethodGet, url, apiToken)

	if err != nil {
		return Response{}, err
	}

	return executeRequest(client, req)
}

// getWithRetry works similarly to retry does for PUT and POST
// this method can be used for API calls we know to have occasional timing issues on GET - e.g. paginated queries that are impacted by replication lag, returning unequal amounts of objects/pages per node
func getWithRetry(client *http.Client, url string, apiToken string, maxRetries int, timeout time.Duration) (resp Response, err error) {
	resp, err = get(client, url, apiToken)

	if err == nil && success(resp) {
		return resp, nil
	}

	for i := 0; i < maxRetries; i++ {
		log.Warn("Retrying failed GET request %s after error (HTTP %d): %w", url, resp.StatusCode, err)
		time.Sleep(timeout)
		resp, err = get(client, url, apiToken)
		if err == nil && success(resp) {
			return resp, err
		}
	}

	var retryErr error
	if err != nil {
		retryErr = fmt.Errorf("GET request %s failed after %d retries: %w", url, maxRetries, err)
	} else {
		retryErr = fmt.Errorf("GET request %s failed after %d retries: (HTTP %d)!\n    Response was: %s", url, maxRetries, resp.StatusCode, resp.Body)
	}
	return Response{}, retryErr
}

// the name delete() would collide with the built-in function
func deleteConfig(client *http.Client, url string, apiToken string, id string) error {
	fullPath := url + "/" + id
	req, err := request(http.MethodDelete, fullPath, apiToken)

	if err != nil {
		return err
	}

	resp, err := executeRequest(client, req)

	if err != nil {
		return err
	}

	if resp.StatusCode == 404 {
		log.Debug("No config with id '%s' found to delete (HTTP 404 response)", id)
		return nil
	}

	if !success(resp) {
		return fmt.Errorf("failed call to DELETE %s (HTTP %d)!\n Response was:\n %s", fullPath, resp.StatusCode, string(resp.Body))
	}

	return nil
}

func post(client *http.Client, url string, data []byte, apiToken string) (Response, error) {
	req, err := requestWithBody(http.MethodPost, url, bytes.NewBuffer(data), apiToken)

	if err != nil {
		return Response{}, err
	}

	return executeRequest(client, req)
}

func postMultiPartFile(client *http.Client, url string, data *bytes.Buffer, contentType string, apiToken string) (Response, error) {
	req, err := requestWithBody(http.MethodPost, url, data, apiToken)

	if err != nil {
		return Response{}, err
	}

	req.Header.Set("Content-type", contentType)

	return executeRequest(client, req)
}

func put(client *http.Client, url string, data []byte, apiToken string) (Response, error) {
	req, err := requestWithBody(http.MethodPut, url, bytes.NewBuffer(data), apiToken)

	if err != nil {
		return Response{}, err
	}

	return executeRequest(client, req)
}

// function type of put and post requests
type sendingRequest func(client *http.Client, url string, data []byte, apiToken string) (Response, error)

func sendWithRetry(client *http.Client, restCall sendingRequest, objectName string, path string, body []byte, apiToken string, maxRetries int, timeout time.Duration) (resp Response, err error) {

	for i := 0; i < maxRetries; i++ {
		log.Warn("\t\t\tDependency of config %s was not available. Waiting for %s before retry...", objectName, timeout)
		time.Sleep(timeout)
		resp, err = restCall(client, path, body, apiToken)
		if err == nil && success(resp) {
			return resp, err
		}
	}

	var retryErr error
	if err != nil {
		retryErr = fmt.Errorf("dependency of config %s was not available after %d retries: %w", objectName, maxRetries, err)
	} else {
		retryErr = fmt.Errorf("dependency of config %s was not available after %d retries: (HTTP %d)!\n    Response was: %s", objectName, maxRetries, resp.StatusCode, resp.Body)
	}
	return Response{}, retryErr
}

func request(method string, url string, apiToken string) (*http.Request, error) {
	return requestWithBody(method, url, nil, apiToken)
}

func requestWithBody(method string, url string, body io.Reader, apiToken string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Api-Token "+apiToken)
	req.Header.Set("Content-type", "application/json")
	req.Header.Set("User-Agent", "Dynatrace Monitoring as Code/"+version.MonitoringAsCode+" "+(runtime.GOOS+" "+runtime.GOARCH))
	return req, nil
}

func executeRequest(client *http.Client, request *http.Request) (Response, error) {
	var requestId string
	if log.IsRequestLoggingActive() {
		requestId = uuid.NewString()
		err := log.LogRequest(requestId, request)

		if err != nil {
			log.Warn("error while writing request log for id `%s`: %v", requestId, err)
		}
	}

	rateLimitStrategy := createRateLimitStrategy()

	response, err := rateLimitStrategy.executeRequest(util.NewTimelineProvider(), func() (Response, error) {
		resp, err := client.Do(request)
		if err != nil {
			log.Error("HTTP Request failed with Error: " + err.Error())
			return Response{}, err
		}
		defer func() {
			err = resp.Body.Close()
		}()
		body, err := io.ReadAll(resp.Body)

		if log.IsResponseLoggingActive() {
			err := log.LogResponse(requestId, resp, string(body))

			if err != nil {
				if requestId != "" {
					log.Warn("error while writing response log for id `%s`: %v", requestId, err)
				} else {
					log.Warn("error while writing response log: %v", requestId, err)
				}
			}
		}

		return Response{
			StatusCode: resp.StatusCode,
			Body:       body,
			Headers:    resp.Header,
		}, err
	})

	if err != nil {
		return Response{}, err
	}
	return response, nil
}
