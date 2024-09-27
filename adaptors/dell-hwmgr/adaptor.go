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

package dellhwmgr

import (
	"context"
	"log/slog"

	"github.com/openshift-kni/oran-hwmgr-plugin/internal/controller/utils"
	hwmgmtv1alpha1 "github.com/openshift-kni/oran-o2ims/api/hardwaremanagement/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Setup the Loopback Adaptor
type DellHwMgrAdaptor struct {
	client.Client
	logger    *slog.Logger
	namespace string
}

func NewDellHwMgrAdaptor(client client.Client, logger *slog.Logger, namespace string) *DellHwMgrAdaptor {
	return &DellHwMgrAdaptor{
		Client:    client,
		logger:    logger,
		namespace: namespace,
	}
}

func (a *DellHwMgrAdaptor) SetupAdaptor() error {
	a.logger.Info("SetupAdaptor called for DellHwMgr")
	return nil
}

func (a *DellHwMgrAdaptor) HandleNodePool(ctx context.Context, nodepool *hwmgmtv1alpha1.NodePool) (ctrl.Result, error) {
	result := utils.DoNotRequeue()

	a.logger.Error("DellHwMgr is not yet implemented")
	utils.SetStatusCondition(&nodepool.Status.Conditions,
		hwmgmtv1alpha1.Provisioned,
		hwmgmtv1alpha1.Failed,
		metav1.ConditionFalse,
		"Unsupported hwmgr adaptor: dell-hwmgr is not yet implemented")

	return result, nil
}

func (a *DellHwMgrAdaptor) HandleNodePoolDeletion(ctx context.Context, nodepool *hwmgmtv1alpha1.NodePool) error {
	a.logger.InfoContext(ctx, "DellHwMgr HandleNodePoolDeletion", "name", nodepool.Name)

	return nil
}
