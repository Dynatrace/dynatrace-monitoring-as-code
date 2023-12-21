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

package downloader_test

import (
	"context"
	accountmanagement "github.com/dynatrace/dynatrace-configuration-as-code-core/gen/account_management"
	stringutils "github.com/dynatrace/dynatrace-configuration-as-code/v2/internal/strings"
	"github.com/dynatrace/dynatrace-configuration-as-code/v2/pkg/account"
	"github.com/dynatrace/dynatrace-configuration-as-code/v2/pkg/account/downloader"
	"github.com/dynatrace/dynatrace-configuration-as-code/v2/pkg/account/downloader/internal/http"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestDownloader_DownloadConfiguration(t *testing.T) {
	uuidVar := "27dde8b6-2ed3-48f1-90b5-e4c0eae8b9bd"
	tests := []struct {
		name      string
		given     mockData
		expected  account.Resources
		expectErr bool
	}{
		{
			name:  "empty account",
			given: mockData{},
			expected: account.Resources{
				Policies: make(map[account.PolicyId]account.Policy),
				Groups:   make(map[account.GroupId]account.Group),
				Users:    make(map[account.UserId]account.User),
			},
		},
		{
			name: "global policies",
			given: mockData{
				policies: []accountmanagement.PolicyOverview{
					{
						Uuid:        "2ff9314d-3c97-4607-bd49-460a53de1390",
						Name:        "test policy - tenant",
						Description: "some description",
						LevelId:     "",
						LevelType:   "account",
					}},
				policieDef: &accountmanagement.LevelPolicyDto{
					Uuid:           "07beda6d-6a02-4827-9c1c-49037c96f176",
					Name:           "test policy",
					Description:    "user friendly description",
					StatementQuery: "THIS IS statement",
				},
			},
			expected: account.Resources{
				Policies: map[account.PolicyId]account.Policy{
					stringutils.Sanitize("test policy - tenant"): {
						ID:             stringutils.Sanitize("test policy - tenant"),
						Name:           "test policy - tenant",
						Level:          account.PolicyLevelAccount{Type: "account"},
						Description:    "some description",
						Policy:         "THIS IS statement",
						OriginObjectID: "2ff9314d-3c97-4607-bd49-460a53de1390",
					},
				},
				Groups: make(map[account.GroupId]account.Group),
				Users:  make(map[account.UserId]account.User),
			},
		},
		{
			name: "environment policies",
			given: mockData{
				policies: []accountmanagement.PolicyOverview{
					{
						Uuid:        "2ff9314d-3c97-4607-bd49-460a53de1390",
						Name:        "test policy - tenant",
						Description: "some description",
						LevelId:     "abc12345",
						LevelType:   "environment",
					}},
				policieDef: &accountmanagement.LevelPolicyDto{
					Uuid:           "07beda6d-6a02-4827-9c1c-49037c96f176",
					Name:           "test policy",
					Description:    "user friendly description",
					StatementQuery: "THIS IS statement",
				},
			},
			expected: account.Resources{
				Policies: map[account.PolicyId]account.Policy{
					stringutils.Sanitize("test policy - tenant"): {
						ID:   stringutils.Sanitize("test policy - tenant"),
						Name: "test policy - tenant",
						Level: account.PolicyLevelEnvironment{
							Type:        "environment",
							Environment: "abc12345",
						},
						Description:    "some description",
						Policy:         "THIS IS statement",
						OriginObjectID: "2ff9314d-3c97-4607-bd49-460a53de1390",
					},
				},
				Groups: make(map[account.GroupId]account.Group),
				Users:  make(map[account.UserId]account.User),
			},
		},
		{
			name: "global policy",
			given: mockData{
				policies: []accountmanagement.PolicyOverview{{
					Uuid:        "07beda6d-6a02-4827-9c1c-49037c96f176",
					Name:        "test global policy",
					Description: "user friendly description",
					LevelId:     "",
					LevelType:   "global",
				}},
			},
			expected: account.Resources{
				Policies: map[account.PolicyId]account.Policy{},
				Groups:   make(map[account.GroupId]account.Group),
				Users:    make(map[account.UserId]account.User),
			},
		},
		{
			name: "no policy details (GetPolicyDefinition returns nil)",
			given: mockData{
				policies: []accountmanagement.PolicyOverview{{
					Uuid:        uuid.New().String(),
					Name:        "test policy",
					Description: "",
					LevelId:     "",
					LevelType:   "account",
				}},
			},
			expectErr: true,
		},
		{
			name: "only user",
			given: mockData{
				users:      []accountmanagement.UsersDto{{Email: "usert@some.org"}},
				userGroups: &accountmanagement.GroupUserDto{Email: "usert@some.org"},
			},
			expected: account.Resources{
				Policies: make(map[account.PolicyId]account.Policy),
				Groups:   make(map[account.GroupId]account.Group),
				Users: map[account.UserId]account.User{
					"usert@some.org": {Email: "usert@some.org"},
				},
			},
		},
		{
			name: "no requested user details (GetGroupsForUser returns nil) ",
			given: mockData{
				users:      []accountmanagement.UsersDto{{Email: "usert@some.org"}},
				userGroups: nil,
			},
			expectErr: true,
		},
		{
			name: "empty group",
			given: mockData{
				ai: &account.AccountInfo{
					Name:        "test",
					AccountUUID: "0b7259a3-61e6-401d-a2ea-c474a219d24b",
				},
				groups: []accountmanagement.GetGroupDto{{
					Uuid:                     &uuidVar,
					Name:                     "test group",
					Description:              nil,
					FederatedAttributeValues: nil,
					Owner:                    "",
					CreatedAt:                "",
					UpdatedAt:                "",
				}},
				policyGroupBindings: []policyGroupBindings{
					{
						levelType: "account",
						levelId:   "0b7259a3-61e6-401d-a2ea-c474a219d24b",
						bindings: &accountmanagement.LevelPolicyBindingDto{
							PolicyBindings: []accountmanagement.Binding{{
								PolicyUuid: uuidVar,
								Groups:     []string{uuidVar},
							}},
						},
						err: nil,
					},
				},
			},
			expected: account.Resources{
				Policies: map[account.PolicyId]account.Policy{},
				Groups: map[account.GroupId]account.Group{
					stringutils.Sanitize("test group"): {
						ID:             stringutils.Sanitize("test group"),
						Name:           "test group",
						OriginObjectID: uuidVar,
					},
				},
				Users: map[account.UserId]account.User{},
			},
		},
		{
			name: "no group details (GetGroupsForUser returns nil)",
			given: mockData{
				users: []accountmanagement.UsersDto{{Email: "test.user@some.org"}},
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := newMockDownloader(tc.given, t).DownloadResources(context.TODO())

			if !tc.expectErr {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, *actual)
			} else {
				assert.Error(t, err)
			}
		})
	}

	t.Run("http client error handling", func(t *testing.T) {
		givenErr := errors.New("given error")
		tests := []struct {
			name            string
			given           mockData
			expectedMessage string
		}{
			{
				name:            "GetEnvironmentsAndMZones returns error",
				given:           mockData{environmentsAndMZonesError: givenErr},
				expectedMessage: "failed to get a list of environments and management zones for account ",
			},
			{
				name:            "GetPoliciesForAccount returns error",
				given:           mockData{policiesError: givenErr},
				expectedMessage: "failed to get a list of policies for account",
			},
			{
				name:            "GetUsers returns error",
				given:           mockData{usersError: givenErr},
				expectedMessage: "failed to get a list of users for account",
			},
			{
				name:            "GetGroups returns error",
				given:           mockData{groupsError: givenErr},
				expectedMessage: "failed to get a list of groups for account",
			},
			{
				name: "GetGroupsForUser returns error",
				given: mockData{
					users:              []accountmanagement.UsersDto{{Email: "test.user@some.org"}},
					groupsForUserError: givenErr,
				},
				expectedMessage: "failed to get a list of bind groups for user",
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				d := newMockDownloader(tc.given, t)

				_, actualErr := d.DownloadResources(context.TODO())

				assert.Error(t, actualErr)
				assert.ErrorContains(t, actualErr, givenErr.Error(), "Returned error must contain original error")
				assert.ErrorContains(t, actualErr, tc.expectedMessage, "Return error must contain additional information")
			})
		}
	})
}

type (
	policyGroupBindings struct {
		levelType, levelId string
		bindings           *accountmanagement.LevelPolicyBindingDto
		err                error
	}

	mockData struct {
		ai                  *account.AccountInfo
		envs                []accountmanagement.TenantResourceDto
		mzones              []accountmanagement.ManagementZoneResourceDto
		policies            []accountmanagement.PolicyOverview
		policieDef          *accountmanagement.LevelPolicyDto
		policyGroupBindings []policyGroupBindings
		permissions         accountmanagement.PermissionsGroupDto
		groups              []accountmanagement.GetGroupDto
		users               []accountmanagement.UsersDto
		userGroups          *accountmanagement.GroupUserDto

		environmentsAndMZonesError,
		policiesError,
		policyDefinitionError,
		groupsError,
		usersError,
		groupsForUserError,
		permissionError error
	}
)

func newMockDownloader(d mockData, t *testing.T) *downloader.Downloader {
	if d.ai == nil {
		d.ai = &account.AccountInfo{
			Name:        "test",
			AccountUUID: uuid.New().String(),
		}
	}
	client := http.NewMockhttpClient(gomock.NewController(t))

	ctx := gomock.AssignableToTypeOf(context.TODO())

	client.EXPECT().GetEnvironmentsAndMZones(ctx, d.ai.AccountUUID).Return(d.envs, d.mzones, d.environmentsAndMZonesError).MinTimes(0).MaxTimes(1)
	client.EXPECT().GetPolicies(ctx, d.ai.AccountUUID).Return(d.policies, d.policiesError).MinTimes(0).MaxTimes(1)
	client.EXPECT().GetPolicyDefinition(ctx, policy(d.policies)).Return(d.policieDef, d.policyDefinitionError).AnyTimes()
	if len(d.policyGroupBindings) == 0 {
		client.EXPECT().GetPolicyGroupBindings(ctx, gomock.Any(), gomock.Any()).Return(&accountmanagement.LevelPolicyBindingDto{}, nil).AnyTimes()
	} else {
		for _, b := range d.policyGroupBindings {
			client.EXPECT().GetPolicyGroupBindings(ctx, b.levelType, b.levelId).Return(b.bindings, b.err).Times(1)
		}
	}
	client.EXPECT().GetPermissionFor(ctx, gomock.Any(), gomock.Any()).Return(&d.permissions, d.permissionError).AnyTimes()
	client.EXPECT().GetGroups(ctx, d.ai.AccountUUID).Return(d.groups, d.groupsError).MinTimes(0).MaxTimes(1)
	client.EXPECT().GetUsers(ctx, d.ai.AccountUUID).Return(d.users, d.usersError).MinTimes(0).MaxTimes(1)
	client.EXPECT().GetGroupsForUser(ctx, userEmail(d.users), d.ai.AccountUUID).Return(d.userGroups, d.groupsForUserError).AnyTimes()

	return downloader.New4Test(d.ai, client)
}

func userEmail(u []accountmanagement.UsersDto) string {
	if u == nil {
		return ""
	}
	return u[0].Email
}

func policy(ps []accountmanagement.PolicyOverview) accountmanagement.PolicyOverview {
	if len(ps) == 0 {
		return accountmanagement.PolicyOverview{}
	}
	return ps[0]
}
