/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/oapi-codegen/oapi-codegen/v2/pkg/securityprovider"
	hwmgrapi "github.com/openshift-kni/oran-hwmgr-plugin/adaptors/dell-hwmgr/generated"
	pluginv1alpha1 "github.com/openshift-kni/oran-hwmgr-plugin/api/hwmgr-plugin/v1alpha1"
	"github.com/openshift-kni/oran-hwmgr-plugin/internal/controller/utils"
	corev1 "k8s.io/api/core/v1"
)

// Struct definitions for the token response
type TokenResponse struct {
	AccessToken      string `json:"access_token,omitempty"`
	ExpiresIn        int    `json:"expires_in,omitempty"`
	RefreshExpiresIn int    `json:"refresh_expires_in,omitempty"`
	RefreshToken     string `json:"refresh_token,omitempty"`
	TokenType        string `json:"token_type,omitempty"`
	SessionState     string `json:"session_state,omitempty"`
	IdToken          string `json:"id_token,omitempty"`
	Scope            string `json:"scope,omitempty"`
}

// GetToken sends a request to the hardware manager to request an authentication token
func (r *HardwareManagerReconciler) GetToken(ctx context.Context, client *hwmgrapi.ClientWithResponses, hwmgr *pluginv1alpha1.HardwareManager) (string, error) {
	clientSecrets, err := utils.GetSecret(ctx, r.Client, hwmgr.Spec.DellData.AuthSecret, r.Namespace)
	if err != nil {
		return "", fmt.Errorf("failed to get client secret: %w", err)
	}

	clientId, err := utils.GetSecretField(clientSecrets, "client-id")
	if err != nil {
		return "", fmt.Errorf("failed to get client-id from secret: %s, %w", hwmgr.Spec.DellData.AuthSecret, err)
	}

	username, err := utils.GetSecretField(clientSecrets, corev1.BasicAuthUsernameKey)
	if err != nil {
		return "", fmt.Errorf("failed to get %s from secret: %s, %w", corev1.BasicAuthUsernameKey, hwmgr.Spec.DellData.AuthSecret, err)
	}

	password, err := utils.GetSecretField(clientSecrets, corev1.BasicAuthPasswordKey)
	if err != nil {
		return "", fmt.Errorf("failed to get %s from secret: %s, %w", corev1.BasicAuthPasswordKey, hwmgr.Spec.DellData.AuthSecret, err)
	}

	grant_type := string(pluginv1alpha1.OAuthGrantTypes.Password)

	req := hwmgrapi.GetTokenJSONRequestBody{
		ClientId:  &clientId,
		Username:  &username,
		Password:  &password,
		GrantType: &grant_type,
	}

	tokenrsp, err := client.GetTokenWithResponse(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to get token: response: %v, err: %w", tokenrsp, err)
	}

	if tokenrsp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("token request failed with status %s (%d), message=%s",
			tokenrsp.Status(), tokenrsp.StatusCode(), string(tokenrsp.Body))
	}

	var tokenData TokenResponse
	if err := json.Unmarshal(tokenrsp.Body, &tokenData); err != nil {
		return "", fmt.Errorf("failed to parse token: response: %v, err: %w", tokenrsp, err)
	}

	return tokenData.AccessToken, nil
}

// NewClientWithResponses creates an authenticated client connected to the hardware manager
func (r *HardwareManagerReconciler) NewClientWithResponses(ctx context.Context, hwmgr *pluginv1alpha1.HardwareManager) (*hwmgrapi.ClientWithResponses, error) {
	var caBundle string
	if hwmgr.Spec.DellData.CaBundleName != nil {
		cm, err := utils.GetConfigmap(ctx, r.Client, *hwmgr.Spec.DellData.CaBundleName, r.Namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to get configmap: %s", err.Error())
		}

		caBundle, err = utils.GetConfigMapField(cm, "ca-bundle.pem")
		if err != nil {
			return nil, fmt.Errorf("failed to get certificate bundle from configmap: %s", err.Error())
		}
	}

	config := utils.OAuthClientConfig{
		CaBundle: []byte(caBundle),
	}

	tr, err := utils.GetTransportWithCaBundle(config)
	if err != nil {
		return nil, fmt.Errorf("failed to get http transport: %w", err)
	}

	httpClient := &http.Client{Transport: tr}

	client, err := hwmgrapi.NewClientWithResponses(
		hwmgr.Spec.DellData.ApiUrl,
		hwmgrapi.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to setup client to %s: %w", hwmgr.Spec.DellData.ApiUrl, err)
	}

	token, err := r.GetToken(ctx, client, hwmgr)
	if err != nil {
		return nil, fmt.Errorf("failed to get token for %s: %w", hwmgr.Name, err)
	}

	bearerAuth, err := securityprovider.NewSecurityProviderBearerToken(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create security provider for %s: %w", hwmgr.Name, err)
	}

	// Create a new client with an intercept to add the bearer token
	authClient, err := hwmgrapi.NewClientWithResponses(
		hwmgr.Spec.DellData.ApiUrl,
		hwmgrapi.WithHTTPClient(httpClient),
		hwmgrapi.WithRequestEditorFn(bearerAuth.Intercept))
	if err != nil {
		return nil, fmt.Errorf("failed to setup auth client for %s: %w", hwmgr.Name, err)
	}

	return authClient, nil
}

// GetResourceGroup sends a request to the hardware manager
func (r *HardwareManagerReconciler) GetResourceGroup(ctx context.Context, client *hwmgrapi.ClientWithResponses, rgId string) error {
	response, err := client.GetResourceGroupWithResponse(ctx, rgId, nil)
	if err != nil {
		return fmt.Errorf("failed to get resource group %s: response: %v, err: %w", rgId, response, err)
	}

	if response.StatusCode() != http.StatusOK {
		return fmt.Errorf("resource group get failed with status %s (%d), message=%s",
			response.Status(), response.StatusCode(), string(response.Body))
	}

	return nil
}
