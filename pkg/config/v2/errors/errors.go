// @license
// Copyright 2021 Dynatrace LLC
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package errors

import (
	"fmt"

	"github.com/dynatrace/dynatrace-configuration-as-code/internal/environment"
	"github.com/dynatrace/dynatrace-configuration-as-code/pkg/config/v2/coordinate"
)

type ConfigError interface {
	error
	Coordinates() coordinate.Coordinate
}

type EnvironmentDetails struct {
	Group       string
	Environment string
}

type DetailedConfigError interface {
	ConfigError
	LocationDetails() EnvironmentDetails
}

type InvalidJsonError struct {
	Config             coordinate.Coordinate
	EnvironmentDetails EnvironmentDetails
	WrappedError       error
}
type RespError struct {
	WrappedError error
	StatusCode   int
}

func (e InvalidJsonError) Unwrap() error {
	return e.WrappedError
}

var (
	// invalidJsonError must support unwrap function
	_ interface{ Unwrap() error } = (*InvalidJsonError)(nil)
)

func (e InvalidJsonError) Coordinates() coordinate.Coordinate {
	return e.Config
}

func (e InvalidJsonError) LocationDetails() EnvironmentDetails {
	return e.EnvironmentDetails
}

func (e InvalidJsonError) Error() string {
	return e.WrappedError.Error()
}

func (e RespError) Error() string {
	return e.WrappedError.Error()
}

func (e RespError) ConcurrentError() string {
	additionalMessage := ""
	if e.StatusCode == 403 {
		concurrentDownloadLimit := environment.GetEnvValueInt(environment.ConcurrentRequestsEnvKey)
		additionalMessage = fmt.Sprintf("\n\n    A 403 error code probably means too many requests.\n    Reduce your CONCURRENT_REQUESTS environment variable (current value: %d). \n    Then wait a few minutes and retry ", concurrentDownloadLimit)
	}
	retryErr := fmt.Sprintf("%s\n%s", e.WrappedError.Error(), additionalMessage)

	return retryErr
}
