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

package adaptors

import (
	"context"
	"fmt"
	"log/slog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	adaptorinterface "github.com/openshift-kni/oran-hwmgr-plugin/adaptors/adaptor-interface"
	"github.com/openshift-kni/oran-hwmgr-plugin/internal/controller/utils"
	hwmgmtv1alpha1 "github.com/openshift-kni/oran-o2ims/api/hardwaremanagement/v1alpha1"

	// Import the adaptors
	dellhwmgr "github.com/openshift-kni/oran-hwmgr-plugin/adaptors/dell-hwmgr"
	"github.com/openshift-kni/oran-hwmgr-plugin/adaptors/loopback"
)

// TODO: Define supported adaptor IDs in oran-o2ims
const (
	LoopbackAdaptorID  = "loopback"
	DellHwMgrAdaptorID = "dell-hwmgr"
)

// HwMgrAdaptorController
type HwMgrAdaptorController struct {
	client.Client
	Logger    *slog.Logger
	namespace string
	adaptors  map[string]adaptorinterface.HwMgrAdaptorIntf
}

func NewHwMgrAdaptorController(client client.Client, logger *slog.Logger, namespace string) (controller *HwMgrAdaptorController, err error) {
	controller = &HwMgrAdaptorController{
		Client:    client,
		Logger:    logger,
		namespace: namespace,
		adaptors:  make(map[string]adaptorinterface.HwMgrAdaptorIntf),
	}

	// Setup the supported adaptors
	controller.adaptors[LoopbackAdaptorID] = loopback.NewLoopbackAdaptor(client, logger, namespace)
	controller.adaptors[DellHwMgrAdaptorID] = dellhwmgr.NewDellHwMgrAdaptor(client, logger, namespace)

	for id, adaptor := range controller.adaptors {
		if err := adaptor.SetupAdaptor(); err != nil {
			logger.Error("failed to setup adaptor", "id", id, "error", err)
		}
	}
	return
}

// HandleNodePool calls the applicable adaptor handler to process the NodePool CR
func (c *HwMgrAdaptorController) HandleNodePool(ctx context.Context, nodepool *hwmgmtv1alpha1.NodePool) (ctrl.Result, error) {
	adaptorID := utils.GetAdaptorIdFromHwMgrId(nodepool.Spec.HwMgrId)

	// Validate the specified adaptor ID
	adaptor, exists := c.adaptors[adaptorID]
	if !exists {
		c.Logger.Error("unsupported adaptor ID", "adaptorID", adaptorID)

		if err := utils.UpdateNodePoolStatusCondition(ctx, c.Client, nodepool,
			hwmgmtv1alpha1.Provisioned, hwmgmtv1alpha1.Failed, metav1.ConditionFalse,
			"Unsupported adaptor ID specified: "+adaptorID); err != nil {
			return utils.RequeueWithMediumInterval(),
				fmt.Errorf("failed to update status for NodePool %s: %w", nodepool.Name, err)
		}

		return utils.DoNotRequeue(), nil
	}

	result, err := adaptor.HandleNodePool(ctx, nodepool)
	if err != nil {
		return result, fmt.Errorf("failed HandleNodePool for adaptorID %s: %w", adaptorID, err)
	}

	return result, nil
}

// HandleNodePool calls the applicable adaptor handler to process the NodePool CR deletion
func (c *HwMgrAdaptorController) HandleNodePoolDeletion(ctx context.Context, nodepool *hwmgmtv1alpha1.NodePool) error {
	adaptorID := utils.GetAdaptorIdFromHwMgrId(nodepool.Spec.HwMgrId)

	// Validate the specified adaptor ID
	adaptor, exists := c.adaptors[adaptorID]
	if !exists {
		c.Logger.Error("unsupported adaptor ID", "adaptorID", adaptorID)
		return nil
	}

	if err := adaptor.HandleNodePoolDeletion(ctx, nodepool); err != nil {
		return fmt.Errorf("failed HandleNodePoolDeletion for adaptorID %s: %w", adaptorID, err)
	}

	return nil
}
