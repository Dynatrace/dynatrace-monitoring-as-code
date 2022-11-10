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

package topologysort

import (
	"fmt"
	"strings"

	s "sort"

	config "github.com/dynatrace-oss/dynatrace-monitoring-as-code/pkg/config/v2"
	"github.com/dynatrace-oss/dynatrace-monitoring-as-code/pkg/config/v2/coordinate"
	"github.com/dynatrace-oss/dynatrace-monitoring-as-code/pkg/config/v2/errors"
	"github.com/dynatrace-oss/dynatrace-monitoring-as-code/pkg/config/v2/parameter"
	project "github.com/dynatrace-oss/dynatrace-monitoring-as-code/pkg/project/v2"
	"github.com/dynatrace-oss/dynatrace-monitoring-as-code/pkg/util/log"
	"github.com/dynatrace-oss/dynatrace-monitoring-as-code/pkg/util/sort"
)

type ProjectsPerEnvironment map[string][]project.Project

type CircularDependencyParameterSortError struct {
	Config             coordinate.Coordinate
	EnvironmentDetails errors.EnvironmentDetails
	Parameter          string
	DependsOn          []parameter.ParameterReference
}

func (e CircularDependencyParameterSortError) Coordinates() coordinate.Coordinate {
	return e.Config
}

func (e CircularDependencyParameterSortError) LocationDetails() errors.EnvironmentDetails {
	return e.EnvironmentDetails
}

var (
	_ errors.DetailedConfigError = (*CircularDependencyParameterSortError)(nil)
)

func (e CircularDependencyParameterSortError) Error() string {
	return fmt.Sprintf("%s: circular dependency detected. check parameter dependencies: %s",
		e.Parameter, joinParameterReferencesToString(e.DependsOn))
}

func joinParameterReferencesToString(refs []parameter.ParameterReference) string {
	switch len(refs) {
	case 0:
		return ""
	case 1:
		return refs[0].String()
	}

	result := strings.Builder{}

	for _, ref := range refs {
		result.WriteString(ref.String())
		result.WriteString(", ")
	}

	return result.String()
}

type CircualDependencyProjectSortError struct {
	Environment string
	Project     string
	// slice of project ids
	DependsOn []string
}

func (e CircualDependencyProjectSortError) Error() string {
	return fmt.Sprintf("%s:%s: circular dependency detected.\n check project dependencies: %s",
		e.Environment, e.Project, strings.Join(e.DependsOn, ", "))
}

type CircularDependencyConfigSortError struct {
	Config      coordinate.Coordinate
	Environment string
	DependsOn   []coordinate.Coordinate
}

func (e CircularDependencyConfigSortError) Error() string {
	return fmt.Sprintf("%s:%s: is part of circular dependency.\n depends on: %s",
		e.Environment, e.Config, joinCoordinatesToString(e.DependsOn))
}

func joinCoordinatesToString(coordinates []coordinate.Coordinate) string {
	switch len(coordinates) {
	case 0:
		return ""
	case 1:
		return coordinates[0].String()
	}

	result := strings.Builder{}

	for _, coordinate := range coordinates {
		result.WriteString(coordinate.String())
		result.WriteString(", ")
	}

	return result.String()
}

var (
	_ error = (*CircularDependencyConfigSortError)(nil)
	_ error = (*CircualDependencyProjectSortError)(nil)
	_ error = (*CircularDependencyParameterSortError)(nil)
)

type ParameterWithName struct {
	Name      string
	Parameter parameter.Parameter
}

func (p *ParameterWithName) IsReferencing(config coordinate.Coordinate, param ParameterWithName) bool {
	for _, ref := range p.Parameter.GetReferences() {
		if ref.Config == config && ref.Property == param.Name {
			return true
		}
	}

	return false
}

func SortParameters(group string, environment string, conf coordinate.Coordinate, parameters config.Parameters) ([]ParameterWithName, []error) {
	parametersWithName := make([]ParameterWithName, 0, len(parameters))

	for name, param := range parameters {
		parametersWithName = append(parametersWithName, ParameterWithName{
			Name:      name,
			Parameter: param,
		})
	}

	s.SliceStable(parametersWithName, func(i, j int) bool {
		return strings.Compare(parametersWithName[i].Name, parametersWithName[j].Name) < 0
	})

	matrix, inDegrees := parametersToSortData(conf, parametersWithName)
	sorted, sortErrs := sort.TopologySort(matrix, inDegrees)

	if len(sortErrs) > 0 {
		errs := make([]error, len(sortErrs))
		for i, sortErr := range sortErrs {
			param := parametersWithName[sortErr.OnId]

			errs[i] = &CircularDependencyParameterSortError{
				Config: conf,
				EnvironmentDetails: errors.EnvironmentDetails{
					Group:       group,
					Environment: environment,
				},
				Parameter: param.Name,
				DependsOn: param.Parameter.GetReferences(),
			}
		}
		return nil, errs

	}

	result := make([]ParameterWithName, 0, len(parametersWithName))

	for i := len(sorted) - 1; i >= 0; i-- {
		result = append(result, parametersWithName[sorted[i]])
	}

	return result, nil
}

func parametersToSortData(conf coordinate.Coordinate, parameters []ParameterWithName) ([][]bool, []int) {
	numParameters := len(parameters)
	matrix := make([][]bool, numParameters)
	inDegrees := make([]int, numParameters)

	for i, param := range parameters {
		matrix[i] = make([]bool, numParameters)

		for j, p := range parameters {
			if i == j {
				continue
			}

			if p.IsReferencing(conf, param) {
				logDependency("Config Parameter", p.Name, param.Name)
				matrix[i][j] = true
				inDegrees[i]++
			}
		}
	}

	return matrix, inDegrees
}

func GetSortedConfigsForEnvironments(projects []project.Project, environments []string) (map[string][]config.Config, []error) {
	sortedProjectsPerEnvironment, errs := sortProjects(projects, environments)

	if len(errs) > 0 {
		return nil, errs
	}

	result := make(map[string][]config.Config)
	var errors []error

	for env, sortedProject := range sortedProjectsPerEnvironment {
		sortedConfigResult := make([]config.Config, 0)

		for _, project := range sortedProject {
			configs := project.Configs[env]
			sortedConfigs, cfgSortErrs := sortConfigs(getConfigs(configs))

			errors = append(errors, cfgSortErrs...)

			sortedConfigResult = append(sortedConfigResult, sortedConfigs...)
		}

		result[env] = sortedConfigResult
	}

	if errors != nil {
		return nil, errors
	}

	return result, nil
}

func getConfigs(m map[string][]config.Config) []config.Config {
	result := make([]config.Config, 0)

	for _, v := range m {
		result = append(result, v...)
	}

	s.SliceStable(result, func(i, j int) bool {
		return strings.Compare(result[i].Coordinate.Config, result[j].Coordinate.Config) < 0
	})

	return result
}

func sortConfigs(configs []config.Config) ([]config.Config, []error) {
	matrix, inDegrees := configsToSortData(configs)

	sorted, sortErrs := sort.TopologySort(matrix, inDegrees)

	if len(sortErrs) > 0 {
		return nil, parseConfigSortErrors(sortErrs, configs)
	}

	result := make([]config.Config, 0, len(configs))

	for i := len(sorted) - 1; i >= 0; i-- {
		result = append(result, configs[sorted[i]])
	}

	return result, nil
}

func configsToSortData(configs []config.Config) ([][]bool, []int) {
	numConfigs := len(configs)
	matrix := make([][]bool, numConfigs)
	inDegrees := make([]int, len(configs))

	for i, config := range configs {
		matrix[i] = make([]bool, numConfigs)

		for j, c := range configs {
			if i == j {
				continue
			}

			// we do not care about skipped configs
			if c.Skip {
				continue
			}

			if c.HasDependencyOn(config) {
				logDependency("Configuration", c.Coordinate.String(), config.Coordinate.String())
				matrix[i][j] = true
				inDegrees[i]++
			}
		}
	}

	return matrix, inDegrees
}

// parseConfigSortErrors turns [sort.TopologySortError] into [CircularDependencyConfigSortError]
// for each config still has an edge to another after sorting an error will be created by aggregating the sort errors
func parseConfigSortErrors(sortErrs []sort.TopologySortError, configs []config.Config) []error {
	depErrs := make(map[coordinate.Coordinate]CircularDependencyConfigSortError)

	for _, sortErr := range sortErrs {
		conf := configs[sortErr.OnId]

		for _, index := range sortErr.UnresolvedIncomingEdgesFrom {
			dependingConfig := configs[index]

			if err, exists := depErrs[dependingConfig.Coordinate]; exists {
				err.DependsOn = append(err.DependsOn, conf.Coordinate)
				depErrs[dependingConfig.Coordinate] = err
			} else {
				depErrs[dependingConfig.Coordinate] = CircularDependencyConfigSortError{
					Config:      dependingConfig.Coordinate,
					Environment: dependingConfig.Environment,
					DependsOn:   []coordinate.Coordinate{conf.Coordinate},
				}
			}
		}
	}

	errs := make([]error, len(depErrs))
	i := 0
	for _, depErr := range depErrs {
		errs[i] = depErr
		i++
	}

	return errs
}

func sortProjects(projects []project.Project, environments []string) (ProjectsPerEnvironment, []error) {
	var errors []error

	resultByEnvironment := make(ProjectsPerEnvironment)

	for _, env := range environments {
		matrix, inDegrees := projectsToSortData(projects, env)

		sorted, sortErrs := sort.TopologySort(matrix, inDegrees)

		if len(sortErrs) > 0 {
			for _, sortErr := range sortErrs {
				p := projects[sortErr.OnId]

				errors = append(errors, &CircualDependencyProjectSortError{
					Environment: env,
					Project:     p.Id,
					DependsOn:   p.Dependencies[env],
				})
			}
		}

		result := make([]project.Project, 0, len(sorted))

		for i := len(sorted) - 1; i >= 0; i-- {
			result = append(result, projects[sorted[i]])
		}

		resultByEnvironment[env] = result
	}

	if errors != nil {
		return nil, errors
	}

	return resultByEnvironment, nil
}

func projectsToSortData(projects []project.Project, environment string) ([][]bool, []int) {
	numProjects := len(projects)
	matrix := make([][]bool, numProjects)
	inDegrees := make([]int, len(projects))

	for i, project := range projects {
		matrix[i] = make([]bool, numProjects)

		for j, p := range projects {
			if i == j {
				continue
			}

			if p.HasDependencyOn(environment, project) {
				logDependency("Project", p.Id, project.Id)
				matrix[i][j] = true
				inDegrees[i]++
			}
		}
	}

	return matrix, inDegrees
}

func logDependency(prefix string, depending string, dependedOn string) {
	log.Debug("%s: %s has dependency on %s", prefix, depending, dependedOn)
}
