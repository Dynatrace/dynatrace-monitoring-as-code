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

package download

import (
	"context"
	"errors"
	"fmt"
	"github.com/dynatrace/dynatrace-configuration-as-code/internal/featureflags"
	"github.com/dynatrace/dynatrace-configuration-as-code/internal/log"
	"github.com/dynatrace/dynatrace-configuration-as-code/internal/version"
	"github.com/dynatrace/dynatrace-configuration-as-code/pkg/client"
	"github.com/dynatrace/dynatrace-configuration-as-code/pkg/manifest"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"net/http"

	"github.com/dynatrace/dynatrace-configuration-as-code/cmd/monaco/runner/completion"
)

func GetDownloadCommand(fs afero.Fs, command Command) (downloadCmd *cobra.Command) {
	var project, outputFolder string
	var forceOverwrite bool
	var specificApis []string
	var specificSettings []string
	var specificEnvironment string
	var onlyAPIs bool
	var onlySettings bool
	var token string
	var oAuthClientID, oAuthClientSecret string
	var manifest string
	var url string

	downloadCmd = &cobra.Command{
		Use:   "download --manifest <manifest.yaml> --token <API_TOKEN_ENV_VAR_NAME>",
		Short: "Download configuration from Dynatrace",
		Long: `Download configuration from Dynatrace

Either downloading based on an existing manifest, or define an URL pointing to an environment to download configuration from.`,
		Example: `- monaco download --manifest manifest.yaml --token API_TOKEN_ENV_VAR_NAME
- monaco download --url https://environment.live.dynatrace.com --token API_TOKEN_ENV_VAR_NAME`,

		RunE: func(cmd *cobra.Command, args []string) error {
			if url != "" {
				if token == "" {
					return errors.New("--token flag missing")
				}

				if oAuthClientID != "" && oAuthClientSecret == "" {
					return errors.New("--oauth-client-secret flag missing")
				}

				if oAuthClientSecret != "" && oAuthClientID == "" {
					return errors.New("--oauth-client-id flag missing")
				}

				options := directDownloadCmdOptions{
					environmentUrl: url,
					auth: auth{
						token:        token,
						clientID:     oAuthClientID,
						clientSecret: oAuthClientSecret,
					},
					downloadCmdOptions: downloadCmdOptions{
						sharedDownloadCmdOptions: sharedDownloadCmdOptions{
							projectName:    project,
							outputFolder:   outputFolder,
							forceOverwrite: forceOverwrite,
						},
						specificAPIs:    specificApis,
						specificSchemas: specificSettings,
						onlyAPIs:        onlyAPIs,
						onlySettings:    onlySettings,
					},
				}

				return command.DownloadConfigs(fs, options)

			} else {
				if specificEnvironment == "" {
					return errors.New("missing --environment/-e flag")
				}
				options := manifestDownloadOptions{
					manifestFile:            manifest,
					specificEnvironmentName: specificEnvironment,
					downloadCmdOptions: downloadCmdOptions{
						sharedDownloadCmdOptions: sharedDownloadCmdOptions{
							projectName:    project,
							outputFolder:   outputFolder,
							forceOverwrite: forceOverwrite,
						},
						specificAPIs:    specificApis,
						specificSchemas: specificSettings,
						onlyAPIs:        onlyAPIs,
						onlySettings:    onlySettings,
					},
				}

				return command.DownloadConfigsBasedOnManifest(fs, options)
			}
		},
	}

	setupSharedConfigsFlags(downloadCmd, &project, &outputFolder, &forceOverwrite, &specificApis, &specificSettings, &onlyAPIs, &onlySettings)
	downloadCmd.Flags().StringVar(&token, "token", "", "Token secret to connect to DT server")
	downloadCmd.Flags().StringVar(&oAuthClientID, "oauth-client-id", "", "OAuth client ID is used to connect to DT server via OAuth (mandatory for OAuth access type)")
	downloadCmd.Flags().StringVar(&oAuthClientSecret, "oauth-client-secret", "", "OAuth client secret is used to connect to DT server via OAuth (mandatory for OAuth access type)")
	downloadCmd.Flags().StringVarP(&manifest, "manifest", "m", "manifest.yaml", "Path to the manifest.yaml file to be read")
	downloadCmd.Flags().StringVarP(&specificEnvironment, "environment", "e", "", "Specify a concrete environment that shall be downloaded (only usable with --manifest)")
	downloadCmd.Flags().StringVarP(&url, "url", "u", "", "URL to the dynatrace environment from which to download configuration from")
	if featureflags.Entities().Enabled() {
		getDownloadEntitiesCommand(fs, command, downloadCmd)
	}

	return downloadCmd
}

func getDownloadEntitiesCommand(fs afero.Fs, command Command, downloadCmd *cobra.Command) {
	var project, outputFolder string
	var forceOverwrite bool
	var specificEntitiesTypes []string

	downloadEntitiesCmd := &cobra.Command{
		Use:   "entities",
		Short: "Download entities configuration from Dynatrace",
		Long: `Download entities configuration from Dynatrace

Either downloading based on an existing manifest, or by defining environment URL and API token via the 'direct' sub-command.`,
		Example: `- monaco download entities manifest manifest.yaml some_environment_from_manifest
- monaco download entities direct https://environment.live.dynatrace.com API_TOKEN_ENV_VAR_NAME`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("'direct' or 'manifest' sub-command is required")
		},
	}

	manifestDownloadCmd := &cobra.Command{
		Use:     "manifest [manifest file] [environment to download]",
		Aliases: []string{"m"},
		Short:   "Download configuration from Dynatrace via a manifest file",
		Example: `monaco download entities manifest.yaml some_environment_from_manifest`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 || args[0] == "" || args[1] == "" {
				return fmt.Errorf(`manifest and environment name have to be provided as positional arguments`)
			}
			return nil
		},
		ValidArgsFunction: completion.DownloadManifestCompletion,
		RunE: func(cmd *cobra.Command, args []string) error {
			m := args[0]
			specificEnvironment := args[1]
			options := entitiesManifestDownloadOptions{
				manifestFile:            m,
				specificEnvironmentName: specificEnvironment,
				entitiesDownloadCommandOptions: entitiesDownloadCommandOptions{
					sharedDownloadCmdOptions: sharedDownloadCmdOptions{
						projectName:    project,
						outputFolder:   outputFolder,
						forceOverwrite: forceOverwrite,
					},
					specificEntitiesTypes: specificEntitiesTypes,
				},
			}
			return command.DownloadEntitiesBasedOnManifest(fs, options)
		},
	}

	directDownloadCmd := &cobra.Command{
		Use:     "direct [URL] [TOKEN_NAME]",
		Aliases: []string{"d"},
		Short:   "Download configuration from a Dynatrace environment specified on the command line",
		Example: `monaco download entities direct https://environment.live.dynatrace.com API_TOKEN_ENV_VAR_NAME`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 || args[0] == "" || args[1] == "" {
				return fmt.Errorf(`url and token have to be provided as positional argument`)
			}
			return nil
		},
		ValidArgsFunction: completion.DownloadDirectCompletion,
		RunE: func(cmd *cobra.Command, args []string) error {
			url := args[0]
			tokenEnvVar := args[1]
			options := entitiesDirectDownloadOptions{
				environmentUrl: url,
				envVarName:     tokenEnvVar,
				entitiesDownloadCommandOptions: entitiesDownloadCommandOptions{
					sharedDownloadCmdOptions: sharedDownloadCmdOptions{
						projectName:    project,
						outputFolder:   outputFolder,
						forceOverwrite: forceOverwrite,
					},
					specificEntitiesTypes: specificEntitiesTypes,
				},
			}
			return command.DownloadEntities(fs, options)

		},
	}

	setupSharedEntitiesFlags(manifestDownloadCmd, &project, &outputFolder, &forceOverwrite, &specificEntitiesTypes)
	setupSharedEntitiesFlags(directDownloadCmd, &project, &outputFolder, &forceOverwrite, &specificEntitiesTypes)

	downloadEntitiesCmd.AddCommand(manifestDownloadCmd)
	downloadEntitiesCmd.AddCommand(directDownloadCmd)

	downloadCmd.AddCommand(downloadEntitiesCmd)
}

func setupSharedConfigsFlags(cmd *cobra.Command, project, outputFolder *string, forceOverwrite *bool, specificApis *[]string, specificSettings *[]string, onlyAPIs, onlySettings *bool) {
	setupSharedFlags(cmd, project, outputFolder, forceOverwrite)
	// flags always available
	cmd.Flags().StringSliceVarP(specificApis, "api", "a", make([]string, 0), "One or more APIs to download (flag can be repeated or value defined as comma-separated list)")
	cmd.Flags().StringSliceVarP(specificSettings, "settings-schema", "s", make([]string, 0), "One or more settings 2.0 schemas to download (flag can be repeated or value defined as comma-separated list)")
	cmd.Flags().BoolVar(onlyAPIs, "only-apis", false, "Only download config APIs, skip downloading settings 2.0 objects")
	cmd.Flags().BoolVar(onlySettings, "only-settings", false, "Only download settings 2.0 objects, skip downloading config APIs")
	cmd.MarkFlagsMutuallyExclusive("settings-schema", "only-apis")
	cmd.MarkFlagsMutuallyExclusive("api", "only-settings")
	cmd.MarkFlagsMutuallyExclusive("only-apis", "only-settings")

	err := cmd.RegisterFlagCompletionFunc("api", completion.AllAvailableApis)
	if err != nil {
		log.Fatal("failed to setup CLI %v", err)
	}
}

func setupSharedEntitiesFlags(cmd *cobra.Command, project, outputFolder *string, forceOverwrite *bool, specificEntitiesTypes *[]string) {
	setupSharedFlags(cmd, project, outputFolder, forceOverwrite)
	cmd.Flags().StringSliceVarP(specificEntitiesTypes, "specific-types", "s", make([]string, 0), "List of entity type IDs specifying which entity types to download")

}
func setupSharedFlags(cmd *cobra.Command, project, outputFolder *string, forceOverwrite *bool) {
	// flags always available
	cmd.Flags().StringVarP(project, "project", "p", "project", "Project to create within the output-folder")
	cmd.Flags().StringVarP(outputFolder, "output-folder", "o", "", "Folder to write downloaded configs to")
	cmd.Flags().BoolVarP(forceOverwrite, "force", "f", false, "Force overwrite any existing manifest.yaml, rather than creating an additional manifest_{timestamp}.yaml. Manifest download: additionally never append source environment name to project folder name")
	err := cmd.MarkFlagDirname("output-folder")
	if err != nil {
		log.Fatal("failed to setup CLI %v", err)
	}
}

// printUploadToSameEnvironmentWarning function may display a warning message on the console,
// notifying the user that downloaded objects cannot be uploaded to the same environment.
// It verifies the version of the tenant and, depending on the result, it may or may not display the warning.
func printUploadToSameEnvironmentWarning(env manifest.EnvironmentDefinition) {
	var serverVersion version.Version
	var err error

	var httpClient *http.Client
	if env.Type == manifest.Classic {
		httpClient = client.NewTokenAuthClient(env.Auth.Token.Value)
	} else {

		credentials := client.OauthCredentials{
			ClientID:     env.Auth.OAuth.ClientID.Value,
			ClientSecret: env.Auth.OAuth.ClientSecret.Value,
			TokenURL:     env.Auth.OAuth.GetTokenEndpointValue(),
		}
		httpClient = client.NewOAuthClient(context.TODO(), credentials)
	}

	serverVersion, err = client.GetDynatraceVersion(httpClient, env.URL.Value)
	if err != nil {
		log.Warn("Unable to determine server version %q: %w", env.URL.Value, err)
		return
	}
	if serverVersion.SmallerThan(version.Version{Major: 1, Minor: 262}) {
		logUploadToSameEnvironmentWarning()
	}
}

func logUploadToSameEnvironmentWarning() {
	log.Warn("Uploading Settings 2.0 objects to the same environment is not possible due to your cluster version " +
		"being below 1.262.0, which Monaco does not support for reliably updating downloaded settings without having " +
		"duplicate configurations. Consider upgrading to 1.262+")
}
