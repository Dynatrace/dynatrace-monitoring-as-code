//go:build integration
// +build integration

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

package v2

import (
	"github.com/dynatrace/dynatrace-configuration-as-code/v2/internal/testutils"
	"testing"

	"github.com/dynatrace/dynatrace-configuration-as-code/v2/cmd/monaco/runner"
	"github.com/spf13/afero"
	"gotest.tools/assert"
)

// tests all configs for a single environment
func TestIntegrationAllConfigsClassic(t *testing.T) {
	specificEnvironment := "classic_env"

	runAllConfigsTest(t, specificEnvironment)
}

func TestIntegrationAllConfigsPlatform(t *testing.T) {
	specificEnvironment := "platform_env"

	runAllConfigsTest(t, specificEnvironment)
}

func runAllConfigsTest(t *testing.T, specificEnvironment string) {
	configFolder := "test-resources/integration-all-configs/"
	manifest := configFolder + "manifest.yaml"

	RunIntegrationWithCleanup(t, configFolder, manifest, specificEnvironment, "AllConfigs", func(fs afero.Fs, _ TestContext) {

		// This causes a POST for all configs:

		cmd := runner.BuildCli(fs)
		cmd.SetArgs([]string{"deploy", "--verbose", manifest, "--environment", specificEnvironment})
		err := cmd.Execute()

		assert.NilError(t, err)

		// This causes a PUT for all configs:

		cmd = runner.BuildCli(fs)
		cmd.SetArgs([]string{"deploy", "--verbose", manifest, "--environment", specificEnvironment})
		err = cmd.Execute()
		assert.NilError(t, err)

	})
}

// Tests a dry run (validation)
func TestIntegrationValidationAllConfigs(t *testing.T) {

	t.Setenv("UNIQUE_TEST_SUFFIX", "can-be-nonunique-for-validation")

	configFolder := "test-resources/integration-all-configs/"
	manifest := configFolder + "manifest.yaml"

	cmd := runner.BuildCli(testutils.CreateTestFileSystem())
	cmd.SetArgs([]string{"deploy", "--verbose", "--dry-run", manifest})
	err := cmd.Execute()

	assert.NilError(t, err)
}

func TestCLDDeploy(t *testing.T) {
	configFolder := "/Users/nicola.riedmann/src/monaco/testing/CLD-7817/demodev-2.6.0/"
	//configFolder := "test-resources/integration-all-configs/"
	manifest := configFolder + "manifest.yaml"

	cmd := runner.BuildCli(afero.NewOsFs())
	cmd.SetArgs([]string{"deploy", "--verbose", "--continue-on-error", manifest})
	_ = cmd.Execute()

}

func TestCLDValidate(t *testing.T) {
	configFolder := "/Users/nicola.riedmann/src/monaco/testing/CLD-7817/demodev-2.6.0/"
	manifest := configFolder + "manifest.yaml"

	cmd := runner.BuildCli(afero.NewOsFs())
	cmd.SetArgs([]string{"deploy", "-d", "--verbose", "--continue-on-error", manifest})
	_ = cmd.Execute()

}
