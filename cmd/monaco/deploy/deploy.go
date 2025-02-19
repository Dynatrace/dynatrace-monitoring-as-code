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

package deploy

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"

	"github.com/dynatrace/dynatrace-configuration-as-code/v2/cmd/monaco/deploy/internal/logging"
	"github.com/dynatrace/dynatrace-configuration-as-code/v2/cmd/monaco/dynatrace"
	"github.com/dynatrace/dynatrace-configuration-as-code/v2/internal/errutils"
	"github.com/dynatrace/dynatrace-configuration-as-code/v2/internal/log"
	"github.com/dynatrace/dynatrace-configuration-as-code/v2/internal/log/field"
	"github.com/dynatrace/dynatrace-configuration-as-code/v2/pkg/api"
	"github.com/dynatrace/dynatrace-configuration-as-code/v2/pkg/config"
	"github.com/dynatrace/dynatrace-configuration-as-code/v2/pkg/config/coordinate"
	"github.com/dynatrace/dynatrace-configuration-as-code/v2/pkg/deploy"
	"github.com/dynatrace/dynatrace-configuration-as-code/v2/pkg/manifest"
	manifestloader "github.com/dynatrace/dynatrace-configuration-as-code/v2/pkg/manifest/loader"
	project "github.com/dynatrace/dynatrace-configuration-as-code/v2/pkg/project/v2"
	v2 "github.com/dynatrace/dynatrace-configuration-as-code/v2/pkg/project/v2"
	"github.com/dynatrace/dynatrace-configuration-as-code/v2/pkg/report"
)

func deployConfigs(ctx context.Context, fs afero.Fs, manifestPath string, environmentGroups []string, specificEnvironments []string, specificProjects []string, continueOnErr bool, dryRun bool) error {
	absManifestPath, err := absPath(manifestPath)
	if err != nil {
		formattedErr := fmt.Errorf("error while finding absolute path for `%s`: %w", manifestPath, err)
		report.GetReporterFromContextOrDiscard(ctx).ReportLoading(report.StateError, formattedErr, "", nil)
		return formattedErr
	}

	loadedManifest, err := loadManifest(ctx, fs, absManifestPath, environmentGroups, specificEnvironments)
	if err != nil {
		return err
	}

	ok := verifyEnvironmentGen(ctx, loadedManifest.Environments, dryRun)
	if !ok {
		return fmt.Errorf("unable to verify Dynatrace environment generation")
	}

	loadedProjects, err := loadProjects(ctx, fs, absManifestPath, loadedManifest, specificProjects)
	if err != nil {
		return err
	}

	if err := validateProjectsWithEnvironments(ctx, loadedProjects, loadedManifest.Environments); err != nil {
		return err
	}

	logging.LogProjectsInfo(loadedProjects)
	logging.LogEnvironmentsInfo(loadedManifest.Environments)

	err = validateAuthenticationWithProjectConfigs(loadedProjects, loadedManifest.Environments)
	if err != nil {
		formattedErr := fmt.Errorf("manifest auth field misconfigured: %w", err)
		report.GetReporterFromContextOrDiscard(ctx).ReportLoading(report.StateError, formattedErr, "", nil)
		return formattedErr
	}

	clientSets, err := dynatrace.CreateEnvironmentClients(ctx, loadedManifest.Environments, dryRun)
	if err != nil {
		formattedErr := fmt.Errorf("failed to create API clients: %w", err)
		report.GetReporterFromContextOrDiscard(ctx).ReportLoading(report.StateError, formattedErr, "", nil)
		return formattedErr
	}

	err = deploy.Deploy(ctx, loadedProjects, clientSets, deploy.DeployConfigsOptions{ContinueOnErr: continueOnErr, DryRun: dryRun})
	if err != nil {
		return fmt.Errorf("%v failed - check logs for details: %w", logging.GetOperationNounForLogging(dryRun), err)
	}

	log.Info("%s finished without errors", logging.GetOperationNounForLogging(dryRun))
	return nil
}

func absPath(manifestPath string) (string, error) {
	manifestPath = filepath.Clean(manifestPath)
	return filepath.Abs(manifestPath)
}

func loadManifest(ctx context.Context, fs afero.Fs, manifestPath string, groups []string, environments []string) (*manifest.Manifest, error) {
	m, errs := manifestloader.Load(&manifestloader.Context{
		Fs:           fs,
		ManifestPath: manifestPath,
		Groups:       groups,
		Environments: environments,
		Opts:         manifestloader.Options{RequireEnvironmentGroups: true},
	})

	if len(errs) > 0 {
		errutils.PrintErrors(errs)
		reporter := report.GetReporterFromContextOrDiscard(ctx)
		for _, err := range errs {
			reporter.ReportLoading(report.StateError, err, "", nil)
		}
		return nil, errors.New("error while loading manifest")
	}

	return &m, nil
}

func verifyEnvironmentGen(ctx context.Context, environments manifest.Environments, dryRun bool) bool {
	if !dryRun {
		return dynatrace.VerifyEnvironmentGeneration(ctx, environments)

	}
	return true
}

func loadProjects(ctx context.Context, fs afero.Fs, manifestPath string, man *manifest.Manifest, specificProjects []string) ([]project.Project, error) {
	projects, errs := project.LoadProjects(ctx, fs, project.ProjectLoaderContext{
		KnownApis:       api.NewAPIs().Filter(api.RemoveDisabled).GetApiNameLookup(),
		WorkingDir:      filepath.Dir(manifestPath),
		Manifest:        *man,
		ParametersSerde: config.DefaultParameterParsers,
	}, specificProjects)

	if errs != nil {
		log.Error("Failed to load projects - %d errors occurred:", len(errs))
		for _, err := range errs {
			log.WithFields(field.Error(err)).Error(err.Error())
		}
		return nil, fmt.Errorf("failed to load projects - %d errors occurred", len(errs))
	}

	return projects, nil
}

type KindCoordinates map[string][]coordinate.Coordinate
type KindCoordinatesPerEnvironment map[string]KindCoordinates
type CoordinatesPerEnvironment map[string][]coordinate.Coordinate

func validateProjectsWithEnvironments(ctx context.Context, projects []project.Project, envs manifest.Environments) error {
	undefinedEnvironments := map[string]struct{}{}
	openPipelineKindCoordinatesPerEnvironment := KindCoordinatesPerEnvironment{}
	platformCoordinatesPerEnvironment := CoordinatesPerEnvironment{}
	for _, p := range projects {
		for envName, cfgPerType := range p.Configs {
			_, found := envs[envName]
			if !found {
				undefinedEnvironments[envName] = struct{}{}
				continue
			}

			openPipelineKindCoordinates, found := openPipelineKindCoordinatesPerEnvironment[envName]
			if !found {
				openPipelineKindCoordinates = KindCoordinates{}
				openPipelineKindCoordinatesPerEnvironment[envName] = openPipelineKindCoordinates
			}
			collectOpenPipelineCoordinatesByKind(cfgPerType, openPipelineKindCoordinates)

			platformCoordinatesPerEnvironment[envName] = append(platformCoordinatesPerEnvironment[envName], collectPlatformCoordinates(cfgPerType)...)
		}
	}

	errs := collectUndefinedEnvironmentErrors(undefinedEnvironments)
	errs = append(errs, collectRequiresPlatformErrors(platformCoordinatesPerEnvironment, envs)...)
	errs = append(errs, collectOpenPipelineCoordinateErrors(openPipelineKindCoordinatesPerEnvironment)...)
	reporter := report.GetReporterFromContextOrDiscard(ctx)

	for _, err := range errs {
		reporter.ReportLoading(report.StateError, err, "", nil)
	}

	return errors.Join(errs...)
}

func collectOpenPipelineCoordinatesByKind(cfgPerType v2.ConfigsPerType, dest KindCoordinates) {
	for cfg := range cfgPerType.AllConfigs {
		if cfg.Skip {
			continue
		}

		if openPipelineType, ok := cfg.Type.(config.OpenPipelineType); ok {
			dest[openPipelineType.Kind] = append(dest[openPipelineType.Kind], cfg.Coordinate)
		}
	}
}

func collectPlatformCoordinates(cfgPerType v2.ConfigsPerType) []coordinate.Coordinate {
	plaformCoordinates := []coordinate.Coordinate{}

	for cfg := range cfgPerType.AllConfigs {
		if cfg.Skip {
			continue
		}

		if configRequiresPlatform(cfg) {
			plaformCoordinates = append(plaformCoordinates, cfg.Coordinate)
		}
	}
	return plaformCoordinates
}

func configRequiresPlatform(c config.Config) bool {
	switch c.Type.(type) {
	case config.AutomationType, config.BucketType, config.DocumentType, config.OpenPipelineType:
		return true
	default:
		return false
	}
}

func collectUndefinedEnvironmentErrors(undefinedEnvironments map[string]struct{}) []error {
	errs := []error{}
	for envName := range undefinedEnvironments {
		errs = append(errs, fmt.Errorf("undefined environment %q", envName))
	}
	return errs
}

func collectOpenPipelineCoordinateErrors(openPipelineKindCoordinatesPerEnvironment KindCoordinatesPerEnvironment) []error {
	errs := []error{}
	for envName, openPipelineKindCoordinates := range openPipelineKindCoordinatesPerEnvironment {

		// check for duplicate configurations for the same kind of openpipeline.
		for kind, coordinates := range openPipelineKindCoordinates {
			if len(coordinates) > 1 {
				errs = append(errs, fmt.Errorf("environment %q has multiple openpipeline configurations of kind %q: %s", envName, kind, coordinateSliceAsString(coordinates)))
			}
		}
	}
	return errs
}

func coordinateSliceAsString(coordinates []coordinate.Coordinate) string {
	coordinateStrings := make([]string, 0, len(coordinates))
	for _, c := range coordinates {
		coordinateStrings = append(coordinateStrings, c.String())
	}
	return strings.Join(coordinateStrings, ", ")
}

func collectRequiresPlatformErrors(platformCoordinatesPerEnvironment CoordinatesPerEnvironment, envs manifest.Environments) []error {
	errs := []error{}
	for envName, coordinates := range platformCoordinatesPerEnvironment {
		env, found := envs[envName]
		if !found || platformEnvironment(env) {
			continue
		}

		if len(coordinates) > 0 {
			exampleCoordinate := coordinates[0]
			errs = append(errs, fmt.Errorf("environment %q is not configured to access platform, but at least one configuration (e.g. %q) requires it", envName, exampleCoordinate))
		}
	}
	return errs
}

func platformEnvironment(e manifest.EnvironmentDefinition) bool {
	return e.Auth.OAuth != nil
}

// validateAuthenticationWithProjectConfigs validates each config entry against the manifest if required credentials are set
// it takes into consideration the project, environments and the skip parameter in each config entry
func validateAuthenticationWithProjectConfigs(projects []project.Project, environments manifest.Environments) error {
	for _, p := range projects {
		for envName, env := range p.Configs {
			for _, file := range env {
				for _, conf := range file {
					if conf.Skip == true {
						continue
					}

					switch conf.Type.(type) {
					case config.ClassicApiType:
						if environments[envName].Auth.Token == nil {
							return fmt.Errorf("API of type '%s' requires a token for environment '%s'", conf.Type, envName)
						}
					case config.SettingsType:
						if environments[envName].Auth.Token == nil && environments[envName].Auth.OAuth == nil {
							return fmt.Errorf("API of type '%s' requires a token or OAuth for environment '%s'", conf.Type, envName)
						}
					default:
						if environments[envName].Auth.OAuth == nil {
							return fmt.Errorf("API of type '%s' requires OAuth for environment '%s'", conf.Type, envName)
						}
					}
				}
			}
		}
	}
	return nil
}
