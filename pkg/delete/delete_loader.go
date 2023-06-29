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

package delete

import (
	"fmt"
	"github.com/dynatrace/dynatrace-configuration-as-code/v2/internal/log"
	"github.com/mitchellh/mapstructure"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

const deleteDelimiter = "/"

type loaderContext struct {
	fs         afero.Fs
	deleteFile string
	knownApis  map[string]struct{}
}

type deleteFileDefinition struct {
	DeleteEntries []interface{} `yaml:"delete"`
}

type deleteEntry struct {
	Project    string `mapstructure:"project"`
	Type       string `mapstructure:"type"`
	ConfigId   string `mapstructure:"id"`
	ConfigName string `mapstructure:"name"`
}

type DeleteEntryParserError struct {
	Value  string `json:"value"`
	Index  int    `json:"index"`
	Reason string `json:"reason"`
}

func newDeleteEntryParserError(value string, index int, reason string) DeleteEntryParserError {
	return DeleteEntryParserError{
		Value:  value,
		Index:  index,
		Reason: reason,
	}
}

func (e DeleteEntryParserError) Error() string {
	return fmt.Sprintf("invalid delete entry `%s` on index `%d`: %s",
		e.Value, e.Index, e.Reason)
}

func LoadEntriesToDelete(fs afero.Fs, knownApis []string, deleteFile string) (map[string][]DeletePointer, []error) {
	context := &loaderContext{
		fs:         fs,
		deleteFile: filepath.Clean(deleteFile),
		knownApis:  toSetMap(knownApis),
	}

	definition, err := readDeleteFile(context)

	if err != nil {
		return nil, []error{err}
	}

	return parseDeleteFileDefinition(context, definition)
}

func toSetMap(strs []string) map[string]struct{} {
	result := make(map[string]struct{})

	for _, s := range strs {
		result[s] = struct{}{}
	}

	return result
}

func readDeleteFile(context *loaderContext) (deleteFileDefinition, error) {
	targetFile, err := filepath.Abs(context.deleteFile)
	if err != nil {
		return deleteFileDefinition{}, fmt.Errorf("could not parse absoulte path to file `%s`: %w", context.deleteFile, err)
	}

	data, err := afero.ReadFile(context.fs, targetFile)

	if err != nil {
		return deleteFileDefinition{}, err
	}

	if len(data) == 0 {
		return deleteFileDefinition{}, fmt.Errorf("file `%s` is empty", targetFile)
	}

	var result deleteFileDefinition

	err = yaml.UnmarshalStrict(data, &result)

	if err != nil {
		return deleteFileDefinition{}, err
	}

	return result, nil
}

func parseDeleteFileDefinition(ctx *loaderContext, definition deleteFileDefinition) (map[string][]DeletePointer, []error) {
	var result = make(map[string][]DeletePointer)
	var errors []error

	for i, e := range definition.DeleteEntries {
		entry, err := parseDeleteEntry(ctx, i, e)

		if err != nil {
			errors = append(errors, err)
			continue
		}

		result[entry.Type] = append(result[entry.Type], entry)
	}

	if errors != nil {
		return nil, errors
	}

	return result, nil
}

func parseDeleteEntry(ctx *loaderContext, index int, entry interface{}) (DeletePointer, error) {

	ptr, err := parseFullEntry(ctx, entry)

	if str, ok := entry.(string); ok && err != nil {
		ptr, err = parseSimpleEntry(str)
	}

	if err != nil {
		return DeletePointer{},
			newDeleteEntryParserError(fmt.Sprintf("%v", entry), index, err.Error())
	}

	return ptr, nil
}

func parseFullEntry(ctx *loaderContext, entry interface{}) (DeletePointer, error) {

	var parsed deleteEntry
	err := mapstructure.Decode(entry, &parsed)
	if err != nil {
		return DeletePointer{}, err
	}

	if _, known := ctx.knownApis[parsed.Type]; known {
		return parseAPIEntry(parsed)
	}

	return parseCoordinateEntry(parsed)
}

func parseAPIEntry(parsed deleteEntry) (DeletePointer, error) {
	if parsed.ConfigName == "" {
		return DeletePointer{}, fmt.Errorf("delete entry of API type requiress config 'name' to be defined")
	}
	if parsed.ConfigId != "" {
		log.Warn("Delete entry %q of API type defines config 'id' - only 'name' will be used.")
	}
	return DeletePointer{
		Type:       parsed.Type,
		Identifier: parsed.ConfigName,
	}, nil
}

func parseCoordinateEntry(parsed deleteEntry) (DeletePointer, error) {
	if parsed.ConfigId == "" {
		return DeletePointer{}, fmt.Errorf("delete entry requires config 'id' to be defined")
	}
	if parsed.Project == "" {
		return DeletePointer{}, fmt.Errorf("delete entry requires 'project' to be defined")
	}
	if parsed.ConfigName != "" {
		log.Warn("Delete entry defines config 'name' - only 'id' will be used.")
	}
	return DeletePointer{
		Project:    parsed.Project,
		Type:       parsed.Type,
		Identifier: parsed.ConfigId,
	}, nil
}

func parseSimpleEntry(entry string) (DeletePointer, error) {
	if !strings.Contains(entry, deleteDelimiter) {
		return DeletePointer{}, fmt.Errorf("invalid format. doesn't contain `%s`", deleteDelimiter)
	}

	parts := strings.SplitN(entry, deleteDelimiter, 2)

	// since the string must contain at least one delimiter and we
	// split the entity by max two, we do not need to test for len of parts
	apiId := parts[0]
	deleteIdentifier := parts[1]

	return DeletePointer{
		Type:       apiId,
		Identifier: deleteIdentifier,
	}, nil
}
