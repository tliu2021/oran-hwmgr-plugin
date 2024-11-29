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

package dellserver

import (
	"net/http"

	apiserver "github.com/openshift-kni/oran-hwmgr-plugin/test/adaptors/dell-hwmgr/dell-server/generated"
)

// These functions will be mocked on a test basis
var GetTokenFn http.HandlerFunc

// This struct implements the http interface provided by the server infra
type DellServer struct{}

func (s DellServer) GetToken(w http.ResponseWriter, r *http.Request) {
	GetTokenFn(w, r)
}

func (s DellServer) VerifyRequestStatus(w http.ResponseWriter, r *http.Request, tenant, jobid string) {
	// To be implemented
}

func (s DellServer) CreateResourceGroup(w http.ResponseWriter, r *http.Request, tenant string) {
	// To be implemented
}

func (s DellServer) DeleteResourceGroup(w http.ResponseWriter, r *http.Request, tenant, resourceGroupId string) {
	// To be implemented
}

func (s DellServer) GetResourceGroup(w http.ResponseWriter, r *http.Request, tenant, resourceGroupId string) {
	// To be implemented
}

func (s DellServer) CreateResourcePool(w http.ResponseWriter, r *http.Request, tenant string) {
	// To be implemented
}

func (s DellServer) DeleteResourcePool(w http.ResponseWriter, r *http.Request, tenant, resourcePoolId string, params apiserver.DeleteResourcePoolParams) {
	// To be implemented
}

func (s DellServer) UpdateResource(w http.ResponseWriter, r *http.Request, tenant string) {
	// To be implemented
}

func (s DellServer) CreateResource(w http.ResponseWriter, r *http.Request, tenant string) {
	// To be implemented
}

func (s DellServer) GetResourceDeployments(w http.ResponseWriter, r *http.Request, tenant, id string) {
	// To be implemented
}

func (s DellServer) DeleteResource(w http.ResponseWriter, r *http.Request, tenant, resourceId string, params apiserver.DeleteResourceParams) {
	// To be implemented
}

func (s DellServer) SubscribeResources(w http.ResponseWriter, r *http.Request, tenant string) {
	// To be implemented
}

func (s DellServer) UnsubscribeResources(w http.ResponseWriter, r *http.Request, tenant string) {
	// To be implemented
}

func (s DellServer) GetResourcePools(w http.ResponseWriter, r *http.Request, tenant string) {
	// To be implemented
}

func (s DellServer) GetResourcePool(w http.ResponseWriter, r *http.Request, tenant, id string) {
	// To be implemented
}

func (s DellServer) GetResources(w http.ResponseWriter, r *http.Request, tenant string) {
	// To be implemented
}

func (s DellServer) GetResource(w http.ResponseWriter, r *http.Request, tenant, id string) {
	// To be implemented
}

func (s DellServer) GetResourceSubscriptions(w http.ResponseWriter, r *http.Request, tenant string) {
	// To be implemented
}

func (s DellServer) GetResourceSubscription(w http.ResponseWriter, r *http.Request, tenant, id string) {
	// To be implemented
}

func (s DellServer) GetSecrets(w http.ResponseWriter, r *http.Request, tenant, secretKey string) {
	// To be implemented
}
