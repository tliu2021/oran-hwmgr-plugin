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
)

// These functions will be mocked on a test basis
var GetTokenFn http.HandlerFunc

// This struct implements the http interface provided by the server infra
type DellServer struct{}

func (s DellServer) GetToken(w http.ResponseWriter, r *http.Request) {
	GetTokenFn(w, r)
}

func (s DellServer) VerifyRequestStatus(w http.ResponseWriter, r *http.Request, jobid string) {
	// To be implemented
}

func (s DellServer) CreateResourceGroup(w http.ResponseWriter, r *http.Request) {
	// To be implemented
}

func (s DellServer) DeleteResourceGroup(w http.ResponseWriter, r *http.Request, resourceGroupId string) {
	// To be implemented
}

func (s DellServer) GetResourceGroup(w http.ResponseWriter, r *http.Request, resourceGroupId string) {
	// To be implemented
}
