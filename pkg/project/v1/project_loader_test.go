//go:build unit

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

package v1

import (
	"github.com/google/go-cmp/cmp/cmpopts"
	"strings"
	"testing"

	"gotest.tools/assert"

	"github.com/dynatrace/dynatrace-configuration-as-code/pkg/api"
	"github.com/dynatrace/dynatrace-configuration-as-code/pkg/util"
)

func TestIfProjectHasSubproject(t *testing.T) {
	mt := util.ReplacePathSeparators("marvin/trillian")
	mth := util.ReplacePathSeparators("marvin/trillian/hacktar")
	rth := util.ReplacePathSeparators("robot/trillian/hacktar")
	projects := []string{"zem", "marvin", mt, mth, rth}
	assert.Equal(t, hasSubprojectFolder("marvin", projects), true, "Check if `marvin` project has subprojects")
	assert.Equal(t, hasSubprojectFolder(mt, projects), true, "Check if `marvin/trillian` project has subprojects")
	assert.Equal(t, hasSubprojectFolder(mth, projects), false, "Check if `marvin/trillian` project has subprojects")
	assert.Equal(t, hasSubprojectFolder(rth, projects), false, "Check if `marvin/trillian` project has subprojects")
	assert.Equal(t, hasSubprojectFolder("zem", projects), false, "Check if `zem` project has subprojects")
	assert.Equal(t, hasSubprojectFolder("unknown", projects), false, "Check if `zem` project has subprojects")
}

func TestFilterProjectsWithSubproject(t *testing.T) {
	ca := util.ReplacePathSeparators("caveman/anjie")
	cag := util.ReplacePathSeparators("caveman/anjie/garkbit")
	mt := util.ReplacePathSeparators("marvin/trillian")
	allProjectFolders := []string{"zem", ca, cag, mt, "trillian"}
	allProjectFolders = filterProjectsWithSubproject(allProjectFolders)

	assert.Equal(t, allProjectFolders[0], "zem", "Check if `zem` folder in list")
	assert.Equal(t, allProjectFolders[1], cag, "Check if `caveman/anjie/garkbit` folder in list")
	assert.Equal(t, allProjectFolders[2], mt, "Check if `marvin/trillian` folder in list")
	assert.Equal(t, allProjectFolders[3], "trillian", "Check if `trillian` folder in list")
	assert.Equal(t, len(allProjectFolders), 4, "Check if only 4 project folders are returned.")
}

func TestGetAllProjectFoldersRecursivelyFailsOnMixedFolder(t *testing.T) {
	path := util.ReplacePathSeparators("test-resources/configs-and-api-mixed-test/project1")
	fs := util.CreateTestFileSystem()
	apis := api.NewApis()
	_, err := getAllProjectFoldersRecursively(fs, apis, path)

	expected := util.ReplacePathSeparators("found folder with projects and configurations in test-resources/configs-and-api-mixed-test/project1")
	assert.Error(t, err, expected)
}

func TestGetAllProjectFoldersRecursivelyFailsOnMixedFolderInSubproject(t *testing.T) {
	path := util.ReplacePathSeparators("test-resources/configs-and-api-mixed-test/project2")
	fs := util.CreateTestFileSystem()
	apis := api.NewApis()
	_, err := getAllProjectFoldersRecursively(fs, apis, path)

	expected := util.ReplacePathSeparators("found folder with projects and configurations in test-resources/configs-and-api-mixed-test/project2/subproject2")
	assert.Error(t, err, expected)
}

func TestGetAllProjectFoldersRecursivelyPassesOnSeparatedFolders(t *testing.T) {
	path := util.ReplacePathSeparators("test-resources/configs-and-api-mixed-test/project3")
	fs := util.CreateTestFileSystem()
	apis := api.NewApis()
	_, err := getAllProjectFoldersRecursively(fs, apis, path)
	assert.NilError(t, err)
}

func TestGetAllProjectsFoldersRecursivelyPassesOnHiddenFolders(t *testing.T) {
	path := util.ReplacePathSeparators("test-resources/hidden-directories/project1")
	fs := util.CreateTestFileSystem()
	_, err := getAllProjectFoldersRecursively(fs, api.NewV1Apis(), path)
	assert.NilError(t, err)
}

func TestGetAllProjectsFoldersRecursivelyPassesOnProjectsWithinHiddenFolders(t *testing.T) {
	path := util.ReplacePathSeparators("test-resources/hidden-directories/project2")
	fs := util.CreateTestFileSystem()
	projects, err := getAllProjectFoldersRecursively(fs, api.NewV1Apis(), path)

	assert.NilError(t, err)

	// NOT test-resources/hidden-directories/project2/.logs
	assert.DeepEqual(t, projects, []string{"test-resources/hidden-directories/project2/subproject"})
}

func TestGetAllProjectsFoldersRecursivelyPassesOnProjects(t *testing.T) {
	path := util.ReplacePathSeparators("test-resources/hidden-directories")
	fs := util.CreateTestFileSystem()
	projects, err := getAllProjectFoldersRecursively(fs, api.NewV1Apis(), path)

	assert.NilError(t, err)

	// NOT test-resources/hidden-directories/.logs
	// NOT test-resources/hidden-directories/project2/.logs
	assert.DeepEqual(t, projects, []string{
		"test-resources/hidden-directories/project1",
		"test-resources/hidden-directories/project2/subproject",
	}, cmpopts.SortSlices(func(a, b string) bool { return strings.Compare(a, b) < 0 }))
}
