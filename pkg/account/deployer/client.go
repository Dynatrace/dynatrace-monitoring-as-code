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

package deployer

import (
	"context"
	"fmt"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/clients/accounts"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	accountmanagement "github.com/dynatrace/dynatrace-configuration-as-code-core/gen/account_management"
	"github.com/dynatrace/dynatrace-configuration-as-code/v2/pkg/account"
	"github.com/go-logr/logr"
	"io"
	"net/http"
	"slices"
)

type (
	Permissions    = accountmanagement.PermissionsDto
	Policy         = accountmanagement.CreateOrUpdateLevelPolicyRequestDto
	Group          = accountmanagement.PutGroupDto
	ManagementZone = accountmanagement.ManagementZoneResourceDto

	accountManagementClient struct {
		accountInfo          account.AccountInfo
		supportedPermissions []remoteId
		client               *accounts.Client
	}
)

func NewClient(info account.AccountInfo, client *accounts.Client, supportedPermissions []remoteId) *accountManagementClient {
	return &accountManagementClient{
		accountInfo:          info,
		client:               client,
		supportedPermissions: supportedPermissions,
	}
}

func (d *accountManagementClient) getAccountInfo() account.AccountInfo {
	return d.accountInfo
}

func (d *accountManagementClient) getGlobalPolicies(ctx context.Context) (map[string]remoteId, error) {
	globalPolicies, resp, err := d.client.PolicyManagementAPI.GetLevelPolicies(ctx, "global", "global").Execute()
	defer closeResponseBody(resp)
	if err = d.handleClientResponseError(resp, err, "unable get global policies"); err != nil {
		return nil, err
	}

	result := make(map[string]remoteId)
	for _, glP := range globalPolicies.GetPolicies() {
		result[glP.Name] = glP.GetUuid()
	}
	return result, nil
}

func (d *accountManagementClient) getAllGroups(ctx context.Context) (map[string]remoteId, error) {
	groups, resp, err := d.client.GroupManagementAPI.GetGroups(ctx, d.accountInfo.AccountUUID).Execute()
	defer closeResponseBody(resp)
	if err = d.handleClientResponseError(resp, err, "unable get all groups for account "+d.accountInfo.AccountUUID); err != nil {
		return nil, err
	}
	result := make(map[string]remoteId)
	for _, glP := range groups.GetItems() {
		result[glP.Name] = glP.GetUuid()
	}
	return result, nil

}

func (d *accountManagementClient) getManagementZones(ctx context.Context) ([]ManagementZone, error) {
	envResources, resp, err := d.client.EnvironmentManagementAPI.GetEnvironmentResources(ctx, d.accountInfo.AccountUUID).Execute()
	defer closeResponseBody(resp)
	if err = d.handleClientResponseError(resp, err, "unable to get environment resources for account "+d.accountInfo.AccountUUID); err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	if envResources == nil {
		return []ManagementZone{}, nil
	}
	return envResources.ManagementZoneResources, nil
}

func (d *accountManagementClient) upsertPolicy(ctx context.Context, policyLevel string, policyLevelId string, policyId string, policy Policy) (remoteId, error) {
	if policyId != "" {

		logr.FromContextOrDiscard(ctx).V(1).Info("Trying to update policy with origin object ID (UUID) " + policyId)
		_, resp, err := d.client.PolicyManagementAPI.UpdateLevelPolicy(ctx, policyLevel, policyLevelId, policyId).CreateOrUpdateLevelPolicyRequestDto(policy).Execute()
		defer closeResponseBody(resp)
		if err = d.handleClientResponseError(resp, err, "unable to update policy with UUID: "+policyId); err != nil {
			return "", err
		}
		return policyId, nil
	}

	logr.FromContextOrDiscard(ctx).V(1).Info("Trying to get policy with name " + policy.Name)
	result, resp, err := d.client.PolicyManagementAPI.GetLevelPolicies(ctx, policyLevel, policyLevelId).Name(policy.Name).Execute()
	defer closeResponseBody(resp)
	if err = d.handleClientResponseError(resp, err, "unable to get policy with name: "+policy.Name); err != nil {
		return "", err
	}

	existingPolicies := result.GetPolicies()

	if len(existingPolicies) == 0 {
		logr.FromContextOrDiscard(ctx).V(1).Info("No policy with name " + policy.Name + " found. Creating a new one")
		var createdPolicy *accountmanagement.LevelPolicyDto
		createdPolicy, resp, err = d.client.PolicyManagementAPI.CreateLevelPolicy(ctx, policyLevel, policyLevelId).CreateOrUpdateLevelPolicyRequestDto(policy).Execute()
		defer closeResponseBody(resp)
		if err = d.handleClientResponseError(resp, err, "unable to create policy with name: "+policy.Name); err != nil {
			return "", err
		}
		return createdPolicy.GetUuid(), nil
	}

	if len(existingPolicies) > 1 { // shouldn't happen
		logr.FromContextOrDiscard(ctx).V(-1).Info("Found multiple policies with name " + policy.Name + ". Updating policy with UUID " + existingPolicies[0].GetUuid())
	}

	logr.FromContextOrDiscard(ctx).V(1).Info("Trying to update existing policy with name " + policy.Name + " and UUID " + existingPolicies[0].GetUuid())
	_, resp, err = d.client.PolicyManagementAPI.UpdateLevelPolicy(ctx, policyLevel, policyLevelId, existingPolicies[0].GetUuid()).CreateOrUpdateLevelPolicyRequestDto(policy).Execute()
	defer closeResponseBody(resp)
	if err = d.handleClientResponseError(resp, err, "unable to update policy with name: "+policy.Name); err != nil {
		return "", err
	}
	return existingPolicies[0].GetUuid(), nil
}

func (d *accountManagementClient) upsertGroup(ctx context.Context, groupId string, group Group) (remoteId, error) {
	if groupId != "" {
		logr.FromContextOrDiscard(ctx).V(1).Info("Trying to update group with origin object ID (UUID) " + groupId)
		resp, err := d.client.GroupManagementAPI.EditGroup(ctx, d.accountInfo.AccountUUID, groupId).PutGroupDto(group).Execute()
		defer closeResponseBody(resp)

		if err = d.handleClientResponseError(resp, err, "unable to update group with UUID: "+groupId); err != nil {
			return "", err
		}
		return groupId, nil
	}

	result, resp, err := d.client.GroupManagementAPI.GetGroups(ctx, d.accountInfo.AccountUUID).Execute()
	defer closeResponseBody(resp)
	if err = d.handleClientResponseError(resp, err, "unable to get group with name: "+group.Name); err != nil {
		return "", err
	}

	// find groups with matching name
	var existingGroups []accountmanagement.GetGroupDto
	for _, g := range result.GetItems() {
		if g.GetName() == group.Name {
			existingGroups = append(existingGroups, g)
		}
	}

	if len(existingGroups) == 0 {
		var createdGroups []accountmanagement.GetGroupDto
		createdGroups, resp, err = d.client.GroupManagementAPI.CreateGroups(ctx, d.accountInfo.AccountUUID).PutGroupDto([]accountmanagement.PutGroupDto{group}).Execute()
		defer closeResponseBody(resp)
		if err = d.handleClientResponseError(resp, err, "unable to create group with name: "+group.Name); err != nil {
			return "", err
		}
		if len(createdGroups) < 1 {
			return "", fmt.Errorf("unable to get UUID of created group with name: %s", group.Name)
		}
		return createdGroups[0].GetUuid(), nil
	}

	if len(existingGroups) > 1 { // shouldn't happen
		logr.FromContextOrDiscard(ctx).V(-1).Info("Updating multiple policies with name " + group.Name + ". Updating group with UUID " + existingGroups[0].GetUuid())
	}

	groupToUpdate := existingGroups[0]

	resp, err = d.client.GroupManagementAPI.EditGroup(ctx, d.accountInfo.AccountUUID, groupToUpdate.GetUuid()).PutGroupDto(group).Execute()
	defer closeResponseBody(resp)
	if err = d.handleClientResponseError(resp, err, "unable to update group with name: "+group.Name); err != nil {
		return "", err
	}
	return groupToUpdate.GetUuid(), nil
}

func (d *accountManagementClient) upsertUser(ctx context.Context, userId string) (remoteId, error) {
	_, resp, err := d.client.UserManagementAPI.GetUserGroups(ctx, d.accountInfo.AccountUUID, userId).Execute()
	defer closeResponseBody(resp)
	if err = d.handleClientResponseError(resp, err, "unable to create user with email: "+userId); err != nil {
		return "", err
	}

	if resp.StatusCode == http.StatusNotFound {
		resp, err = d.client.UserManagementAPI.CreateUserForAccount(ctx, d.accountInfo.AccountUUID).UserEmailDto(accountmanagement.UserEmailDto{Email: userId}).Execute()
		defer closeResponseBody(resp)
		if err = d.handleClientResponseError(resp, err, "unable to create user with email: "+userId); err != nil {
			return "", err
		}

		return userId, nil
	}

	return userId, nil
}

func (d *accountManagementClient) updatePermissions(ctx context.Context, groupId string, permissions []accountmanagement.PermissionsDto) error {
	if groupId == "" {
		return fmt.Errorf("group id must not be empty")
	}

	if permissions == nil {
		permissions = []accountmanagement.PermissionsDto{}
	}

	for _, p := range permissions {
		if !slices.Contains(d.supportedPermissions, p.PermissionName) {
			return fmt.Errorf("unsupported permission %s. Must be one of: %v", p.PermissionName, d.supportedPermissions)
		}
	}
	resp, err := d.client.PermissionManagementAPI.OverwriteGroupPermissions(ctx, d.accountInfo.AccountUUID, groupId).PermissionsDto(permissions).Execute()
	defer closeResponseBody(resp)
	if err = d.handleClientResponseError(resp, err, "unable to update permissions of group with UUID "+groupId); err != nil {
		return err
	}

	return nil
}

func (d *accountManagementClient) updateAccountPolicyBindings(ctx context.Context, groupId string, policyIds []string) error {
	if groupId == "" {
		return fmt.Errorf("group id must not be empty")
	}
	if policyIds == nil {
		policyIds = []string{}
	}
	data := accountmanagement.PolicyUuidsDto{PolicyUuids: policyIds}

	resp, err := d.client.PolicyManagementAPI.UpdatePolicyBindingsToGroup(ctx, "account", d.accountInfo.AccountUUID, groupId).PolicyUuidsDto(data).Execute()
	defer closeResponseBody(resp)
	if err = d.handleClientResponseError(resp, err, "unable to update policy binding between group with UUID "+groupId+" and policies with UUIDs "+fmt.Sprintf("%v", policyIds)); err != nil {
		return err
	}

	return nil
}

func (d *accountManagementClient) updateEnvironmentPolicyBindings(ctx context.Context, envName string, groupId string, policyIds []string) error {
	if envName == "" {
		return fmt.Errorf("environment name must not be empty")
	}
	if groupId == "" {
		return fmt.Errorf("group id must not be empty")
	}
	if policyIds == nil {
		policyIds = []string{}
	}
	data := accountmanagement.PolicyUuidsDto{PolicyUuids: policyIds}
	resp, err := d.client.PolicyManagementAPI.UpdatePolicyBindingsToGroup(ctx, "environment", envName, groupId).PolicyUuidsDto(data).Execute()
	defer closeResponseBody(resp)
	if err = d.handleClientResponseError(resp, err, "unable to update policy binding between group with UUID "+groupId+" and policies with UUIDs "+fmt.Sprintf("%v", policyIds)); err != nil {
		return err
	}
	return nil
}

func (d *accountManagementClient) deleteAllEnvironmentPolicyBindings(ctx context.Context, groupId string) error {
	environments, resp, err := d.client.EnvironmentManagementAPI.GetEnvironments(ctx, d.accountInfo.AccountUUID).Execute()
	defer closeResponseBody(resp)
	if err = d.handleClientResponseError(resp, err, "unable to get all environments for account with id"+d.accountInfo.AccountUUID); err != nil {
		return err
	}

	for _, e := range environments.Data {
		policies, resp, err := d.client.PolicyManagementAPI.GetPolicyUuidsBindings(ctx, "environment", e.Id, groupId).Execute()
		closeResponseBody(resp)
		if err = d.handleClientResponseError(resp, err, "unable to list all environments policy bindings for account with UUID "+d.accountInfo.AccountUUID+" and group with UUID "+groupId); err != nil {
			return err
		}
		for _, pol := range policies.PolicyUuids {
			resp, err = d.client.PolicyManagementAPI.DeleteLevelPolicyBindingsForPolicyAndGroup(ctx, "environment", e.Id, pol, groupId).ForceMultiple(true).Execute()
			closeResponseBody(resp)
			if err = d.handleClientResponseError(resp, err, "unable to delete all environments policy bindings for account with UUID "+d.accountInfo.AccountUUID+" and group with UUID "+groupId); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *accountManagementClient) updateGroupBindings(ctx context.Context, userId string, groupIds []string) error {
	if userId == "" {
		return fmt.Errorf("user id must not be empty")
	}
	if groupIds == nil {
		groupIds = []string{}
	}
	resp, err := d.client.UserManagementAPI.ReplaceUserGroups(ctx, d.accountInfo.AccountUUID, userId).RequestBody(groupIds).Execute()
	defer closeResponseBody(resp)
	if err = d.handleClientResponseError(resp, err, "unable to add user "+userId+" to groups "+fmt.Sprintf("%v", groupIds)); err != nil {
		return err
	}
	return nil
}

func (d *accountManagementClient) handleClientResponseError(resp *http.Response, clientErr error, errMessage string) error {
	if clientErr != nil && resp == nil {
		return fmt.Errorf(errMessage+": %w", clientErr)
	}

	if !rest.IsSuccess(resp) && resp.StatusCode != http.StatusNotFound {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("unable to read response body %w", err)
		}
		return fmt.Errorf(errMessage+" (HTTP %d): %s", resp.StatusCode, string(body))
	}
	return nil
}

func closeResponseBody(resp *http.Response) {
	if resp != nil {
		_ = resp.Body.Close()
	}
}
