/*
 * @license
 * Copyright 2024 Dynatrace LLC
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

package file

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dynatrace/dynatrace-configuration-as-code/v2/pkg/config/parameter"
	envParam "github.com/dynatrace/dynatrace-configuration-as-code/v2/pkg/config/parameter/environment"
)

// TestParseFileValueParameter tests the parsing of file value parameters and that escaping is enabled by default.
func TestParseFileValueParameter(t *testing.T) {
	tests := []struct {
		name                 string
		parameterValue       map[string]any
		expectedEscapedValue bool
	}{
		{
			name:                 "escaped by default",
			parameterValue:       map[string]any{"path": "something.txt"},
			expectedEscapedValue: true,
		},
		{
			name:                 "escaping can be explicitly disabled",
			parameterValue:       map[string]any{"path": "something.txt", "escaped": false},
			expectedEscapedValue: false,
		},
		{
			name:                 "escaping can be explicitly enabled",
			parameterValue:       map[string]any{"path": "something.txt", "escaped": true},
			expectedEscapedValue: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			param, err := parseFileValueParameter(parameter.ParameterParserContext{
				Fs:    afero.NewMemMapFs(),
				Value: tt.parameterValue,
			})

			assert.NoError(t, err)
			fileParam, ok := param.(*FileParameter)
			require.True(t, ok)
			assert.Equal(t, "file", param.GetType())
			assert.Equal(t, "something.txt", fileParam.Path)
			assert.Equal(t, tt.expectedEscapedValue, fileParam.Escaped)
		})
	}
}

// TestParseFileValueParameterEscapedMustBeBoolean tests that setting escaped to a non boolean results in an error.
func TestParseFileValueParameterEscapedMustBeBoolean(t *testing.T) {
	param, err := parseFileValueParameter(parameter.ParameterParserContext{
		Fs:    afero.NewMemMapFs(),
		Value: map[string]any{"path": "something.txt", "escaped": 4},
	})

	assert.Nil(t, param)
	assert.ErrorContains(t, err, "must be a boolean")
}

func TestWriteFileValueParameter(t *testing.T) {
	fileParam := &FileParameter{
		Path:    "myfile",
		Escaped: true,
	}

	context := parameter.ParameterWriterContext{
		Parameter: fileParam,
	}

	result, err := writeFileValueParameter(context)
	require.NoError(t, err)
	assert.Equal(t, "myfile", result["path"])
	assert.Equal(t, true, result["escaped"])
}

func TestWriteFileValueParameter_WrongType(t *testing.T) {
	fileParam := envParam.New("env")

	context := parameter.ParameterWriterContext{
		Parameter: fileParam,
	}

	result, err := writeFileValueParameter(context)
	require.Nil(t, result)
	assert.IsType(t, &parameter.ParameterWriterError{}, err)
}

func TestParseFileValueParameter_MissingPath(t *testing.T) {
	param, err := parseFileValueParameter(parameter.ParameterParserContext{})

	assert.Nil(t, param)
	assert.IsType(t, parameter.ParameterParserError{}, err)
}

// TestResolveValueEscaping tests that escaping of file parameters content can be enabled or disabled.
func TestResolveValueEscaping(t *testing.T) {
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "test-content", []byte(`"test-content"`), 0644)

	tests := []struct {
		name                  string
		fileParam             *FileParameter
		expectedResolvedValue string
	}{
		{
			name: "escaped",
			fileParam: &FileParameter{
				Fs:      fs,
				Path:    "test-content",
				Escaped: true,
			},
			expectedResolvedValue: `\"test-content\"`,
		},
		{
			name: "escaped",
			fileParam: &FileParameter{
				Fs:      fs,
				Path:    "test-content",
				Escaped: false,
			},
			expectedResolvedValue: `"test-content"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.fileParam.ResolveValue(parameter.ResolveContext{})
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedResolvedValue, result)
		})
	}
}

func TestResolveValue(t *testing.T) {
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "test-content", []byte("test-content"), 0644)

	param, err := parseFileValueParameter(parameter.ParameterParserContext{
		Fs:    fs,
		Value: map[string]any{"path": "test-content"},
	})
	assert.NoError(t, err)
	assert.Len(t, param.GetReferences(), 0)

	result, err := param.ResolveValue(parameter.ResolveContext{})
	require.NoError(t, err)
	assert.Equal(t, "test-content", result)
}

func TestResolveValueWithReferences(t *testing.T) {
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "test-content", []byte("test-content {{ .ref1 }} - {{ .ref2 }}"), 0644)

	param, _ := parseFileValueParameter(parameter.ParameterParserContext{
		Fs:    fs,
		Value: map[string]any{"path": "test-content", "references": []any{"ref1", "ref2"}},
	})

	assert.Len(t, param.GetReferences(), 2)

	result, err := param.ResolveValue(parameter.ResolveContext{
		ResolvedParameterValues: map[string]interface{}{
			"ref1": "ref1-resolved",
			"ref2": "ref2-resolved",
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "test-content ref1-resolved - ref2-resolved", result)
}

func TestResolveValueWithRefernces_RefMissing(t *testing.T) {
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "test-content", []byte("test-content {{ .ref1 }} - {{ .ref2 }}"), 0644)

	param, _ := parseFileValueParameter(parameter.ParameterParserContext{
		Fs:    fs,
		Value: map[string]any{"path": "test-content", "references": []any{"ref1", "ref2"}},
	})

	assert.Len(t, param.GetReferences(), 2)

	result, err := param.ResolveValue(parameter.ResolveContext{
		ResolvedParameterValues: map[string]interface{}{
			"ref1": "ref1-resolved",
		},
	})
	assert.Nil(t, result)
	assert.Error(t, err)

}

func TestResolveValue_FileNotFound(t *testing.T) {
	fs := afero.NewMemMapFs()
	param, _ := parseFileValueParameter(parameter.ParameterParserContext{
		Fs:    fs,
		Value: map[string]any{"path": "something"},
	})

	result, err := param.ResolveValue(parameter.ResolveContext{})
	assert.Nil(t, result)
	assert.IsType(t, parameter.ParameterResolveValueError{}, err)
}
