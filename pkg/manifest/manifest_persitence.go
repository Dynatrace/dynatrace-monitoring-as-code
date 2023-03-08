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

package manifest

const simpleProjectType = "simple"
const groupProjectType = "grouping"

type project struct {
	Name string `yaml:"name"`
	Type string `yaml:"type,omitempty"`
	Path string `yaml:"path,omitempty"`
}

type tokenConfig struct {
	Type string `yaml:"type,omitempty"`
	Name string `yaml:"name"`
}

type secretType string

const (
	typeEnvironment secretType = "environment"

	typeValue valueType = "value"
)

// authSecret represents a user-defined client id or client secret. It has a [Type] which is either [typeValue] (default) or [typeEnvironment].
//
// If the type is [typeEnvironment], [Name] contains the environment-variable to resolve the authSecret.
// If the type is [typeValue], [Value] contains the resolved authSecret.
//
// This struct is meant to be reused for fields that require the same behavior.
type authSecret struct {
	Type  secretType `yaml:"type"`
	Name string     `yaml:"name"`Value string    `yaml:"value"`
}

type oAuth struct {
	ClientID     authSecret `yaml:"clientId"`
	ClientSecret authSecret `yaml:"clientSecret"`
}

type auth struct {
	Token tokenConfig `yaml:"token"`
	OAuth oAuth       `yaml:"oAuth"`
}

type environment struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
	Url  url    `yaml:"url"`
	Auth *auth  `yaml:"auth,omitempty"`

	// Token is the user-provided Dynatrace Classic API token
	//
	// Deprecated: Use [Auth.Token] instead. This field is still available to support existing manifests.
	Token *tokenConfig `yaml:"token,omitempty"`
}

type urlType string

const (
	urlTypeEnvironment urlType = "environment"
	urlTypeValue       urlType = "value"
)

type url struct {
	Type  urlType `yaml:"type,omitempty"`
	Value string  `yaml:"value"`
}

type group struct {
	Name         string        `yaml:"name"`
	Environments []environment `yaml:"environments"`
}

type manifest struct {
	ManifestVersion   string    `yaml:"manifestVersion"`
	Projects          []project `yaml:"projects"`
	EnvironmentGroups []group   `yaml:"environmentGroups"`
}
