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
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/openshift-kni/oran-hwmgr-plugin/adaptors/dell-hwmgr/hwmgrclient"
	pluginv1alpha1 "github.com/openshift-kni/oran-hwmgr-plugin/api/hwmgr-plugin/v1alpha1"
	"github.com/openshift-kni/oran-hwmgr-plugin/internal/controller/utils"
	"github.com/openshift-kni/oran-hwmgr-plugin/internal/logging"
	hwmgmtv1alpha1 "github.com/openshift-kni/oran-o2ims/api/hardwaremanagement/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	JobIdAnnotation = "hwmgr-plugin.oran.openshift.io/jobId"
)

// ValidateNodePool performs basic validation of the nodepool data
func (a *Adaptor) ValidateNodePool(nodepool *hwmgmtv1alpha1.NodePool) error {
	if len(nodepool.Spec.NodeGroup) != 2 {
		return fmt.Errorf("nodepool %s invalid: Expected 2 entries in .spec.nodeGroup, got %d", nodepool.Name, len(nodepool.Spec.NodeGroup))
	}

	return nil
}

// HandleNodePoolCreate processes a new NodePool CR, creating a resource group on the hardware manager
func (a *Adaptor) HandleNodePoolCreate(
	ctx context.Context,
	hwmgrClient *hwmgrclient.HardwareManagerClient,
	hwmgr *pluginv1alpha1.HardwareManager,
	nodepool *hwmgmtv1alpha1.NodePool) (ctrl.Result, error) {

	conditionType := hwmgmtv1alpha1.Provisioned
	var conditionReason hwmgmtv1alpha1.ConditionReason
	var conditionStatus metav1.ConditionStatus
	var message string

	// Validate the nodepool data
	if validationErr := a.ValidateNodePool(nodepool); validationErr != nil {
		if err := utils.UpdateNodePoolStatusCondition(ctx, a.Client, nodepool,
			hwmgmtv1alpha1.Provisioned, hwmgmtv1alpha1.Failed, metav1.ConditionFalse,
			"NodePool configuration invalid: "+validationErr.Error()); err != nil {
			return utils.RequeueWithMediumInterval(),
				fmt.Errorf("failed to update status for NodePool %s: %w", nodepool.Name, err)
		}

		return utils.DoNotRequeue(), nil
	}

	if err := a.ProcessNewNodePool(ctx, hwmgrClient, hwmgr, nodepool); err != nil {
		a.Logger.Error("failed createNodePool", "err", err)
		conditionReason = hwmgmtv1alpha1.Failed
		conditionStatus = metav1.ConditionFalse
		message = "Creation request failed: " + err.Error()
	} else {
		conditionReason = hwmgmtv1alpha1.InProgress
		conditionStatus = metav1.ConditionFalse
		message = "Handling creation"
	}

	if err := utils.UpdateNodePoolStatusCondition(ctx, a.Client, nodepool,
		conditionType, conditionReason, conditionStatus, message); err != nil {
		return utils.RequeueWithMediumInterval(),
			fmt.Errorf("failed to update status for NodePool %s: %w", nodepool.Name, err)
	}

	return utils.DoNotRequeue(), nil
}

// ProcessNewNodePool sends a request to the hardware manager to create a resource group
func (a *Adaptor) ProcessNewNodePool(ctx context.Context,
	hwmgrClient *hwmgrclient.HardwareManagerClient,
	hwmgr *pluginv1alpha1.HardwareManager,
	nodepool *hwmgmtv1alpha1.NodePool) error {

	a.Logger.InfoContext(ctx, "Processing ProcessNewNodePool request")

	jobId, err := hwmgrClient.CreateResourceGroup(ctx, nodepool)
	if err != nil {
		return fmt.Errorf("failed CreateResourceGroup: %w", err)
	}

	// Add the jobId in an annotation
	annotations := nodepool.GetAnnotations()
	annotations[JobIdAnnotation] = jobId
	nodepool.SetAnnotations(annotations)
	if err := utils.CreateK8sCR(ctx, a.Client, nodepool, nil, utils.PATCH); err != nil {
		return fmt.Errorf("failed to annotate nodepool %s: %w", nodepool.Name, err)
	}

	return nil
}

// HandleNodePoolProcessing checks the status of an in-progress NodePool, querying the hardware manager
// for the job status. If the job is completed, it queries for the resource group in order to create
// Node CRs corresponding to the allocated nodes.
func (a *Adaptor) HandleNodePoolProcessing(
	ctx context.Context,
	hwmgrClient *hwmgrclient.HardwareManagerClient,
	hwmgr *pluginv1alpha1.HardwareManager,
	nodepool *hwmgmtv1alpha1.NodePool) (ctrl.Result, error) {

	result := ctrl.Result{}
	annotation := nodepool.GetAnnotations()
	if annotation == nil {
		return result, fmt.Errorf("annotations are missing from nodePool %s in namespace %s", nodepool.Name, nodepool.Namespace)
	}

	// Ensure the jobId annotation exists and is not empty
	jobId, exists := annotation[JobIdAnnotation]
	if !exists || jobId == "" {
		return result, fmt.Errorf("%s annotation is missing or empty from nodepool %s", JobIdAnnotation, nodepool.Name)
	}

	ctx = logging.AppendCtx(ctx, slog.String("jobId", jobId))

	// Query the hardware manager for the job status
	status, err := hwmgrClient.CheckResourceGroupRequest(ctx, jobId)
	if err != nil {
		a.Logger.InfoContext(ctx, "Job progress check failed", slog.String("error", err.Error()))
		return result, fmt.Errorf("failed to check job progress, jobId=%s: %w", jobId, err)
	}

	if status == nil || status.Brief == nil || status.Brief.Status == nil {
		a.Logger.InfoContext(ctx, "Job progress check missing data", slog.Any("status", status))
		return result, fmt.Errorf("job progress check missing data, jobId=%s: %w", jobId, err)
	}

	// Process the status response
	switch *status.Brief.Status {
	case "started":
		a.Logger.InfoContext(ctx, "Job has started")
		return utils.RequeueWithShortInterval(), nil
	case "pending":
		a.Logger.InfoContext(ctx, "Job is pending")
		return utils.RequeueWithShortInterval(), nil
	case "completed":
		a.Logger.InfoContext(ctx, "Job has completed")
	default:
		a.Logger.InfoContext(ctx, "Job has failed", slog.Any("status", status))
		msg := fmt.Sprintf("resource group creation has failed: status=%s", *status.Brief.Status)
		if status.Brief.FailReason != nil {
			msg += fmt.Sprintf(", message=%s", *status.Brief.FailReason)
		}
		return result, fmt.Errorf("%s", msg)
	}

	// The job has completed. Get the resource group data from the hardware manager
	rg, err := hwmgrClient.GetResourceGroup(ctx, nodepool)
	if err != nil {
		a.Logger.InfoContext(ctx, "Failed GetResourceGroup", slog.String("error", err.Error()))

		if err := utils.UpdateNodePoolStatusCondition(ctx, a.Client, nodepool,
			hwmgmtv1alpha1.Provisioned, hwmgmtv1alpha1.Failed, metav1.ConditionFalse,
			"Failed to get resource group: "+err.Error()); err != nil {
			return utils.RequeueWithMediumInterval(),
				fmt.Errorf("failed to update status for NodePool %s: %w", nodepool.Name, err)
		}

		return utils.DoNotRequeue(), nil
	}

	// TODO: Need to add validation to ensure the rg satisfies the nodepool

	// Create the Node CRs corresponding to the allocated resources
	for _, resourceSelector := range *rg.ResourceSelectors {
		for _, node := range *resourceSelector.Resources {
			if slices.Contains(nodepool.Status.Properties.NodeNames, *node.Id) {
				a.Logger.InfoContext(ctx, "Node is already added", slog.String("nodename", *node.Id))
				continue
			}
			if nodename, err := a.AllocateNode(ctx, nodepool, node); err != nil {
				a.Logger.InfoContext(ctx, "Failed allocating node", slog.String("err", err.Error()))
				if err := utils.UpdateNodePoolStatusCondition(ctx, a.Client, nodepool,
					hwmgmtv1alpha1.Provisioned, hwmgmtv1alpha1.Failed, metav1.ConditionFalse,
					fmt.Sprintf("Failed to get create node (%s): %s", *node.Name, err.Error())); err != nil {
					return utils.RequeueWithMediumInterval(),
						fmt.Errorf("failed to update status for NodePool %s: %w", nodepool.Name, err)
				}

				return utils.DoNotRequeue(), nil
			} else {
				nodepool.Status.Properties.NodeNames = append(nodepool.Status.Properties.NodeNames, nodename)
			}
		}
	}

	// Update the NodePool CR
	if err := utils.UpdateK8sCRStatus(ctx, a.Client, nodepool); err != nil {
		return utils.RequeueWithMediumInterval(),
			fmt.Errorf("failed to update status for NodePool %s: %w", nodepool.Name, err)
	}

	a.Logger.InfoContext(ctx, "NodePool request is fully allocated")

	if err := utils.UpdateNodePoolStatusCondition(ctx, a.Client, nodepool,
		hwmgmtv1alpha1.Provisioned, hwmgmtv1alpha1.Completed, metav1.ConditionTrue, "Created"); err != nil {
		return utils.RequeueWithMediumInterval(),
			fmt.Errorf("failed to update status for NodePool %s: %w", nodepool.Name, err)
	}

	result = utils.DoNotRequeue()

	return result, nil
}

// ReleaseNodePool frees resources allocated to a NodePool
func (a *Adaptor) ReleaseNodePool(ctx context.Context,
	hwmgrClient *hwmgrclient.HardwareManagerClient,
	hwmgr *pluginv1alpha1.HardwareManager,
	nodepool *hwmgmtv1alpha1.NodePool) error {

	a.Logger.InfoContext(ctx, "Processing ReleaseNodePool request")

	// Issue a resource group deletion request to the hardware manager
	jobId, err := hwmgrClient.DeleteResourceGroup(ctx, nodepool)
	if err != nil {
		return fmt.Errorf("failed CreateResourceGroup: %w", err)
	}

	ctx = logging.AppendCtx(ctx, slog.String("jobId", jobId))

	// TODO: Can we even requeue the finalizer? If so, we should have a separate annotation for this jobID.
	// For now, poll until the job finishes
	finished := false
	for !finished {
		time.Sleep(time.Second * 10)
		a.Logger.InfoContext(ctx, "Checking deletion job progress")

		status, err := hwmgrClient.CheckResourceGroupRequest(ctx, jobId)
		if err != nil {
			a.Logger.InfoContext(ctx, "Job progress check failed", slog.String("error", err.Error()))
			return fmt.Errorf("job progress check failed: %w", err)
		}

		if status == nil || status.Brief == nil || status.Brief.Status == nil {
			a.Logger.InfoContext(ctx, "Job progress check missing data", slog.Any("status", status))
			return fmt.Errorf("job progress check returned with no data: %v", status)
		}

		switch *status.Brief.Status {
		case "started":
			a.Logger.InfoContext(ctx, "Job has started")
			continue
		case "completed":
			a.Logger.InfoContext(ctx, "Job has completed")
			finished = true
		default:
			a.Logger.InfoContext(ctx, "Job has failed", slog.Any("status", status))
			msg := fmt.Sprintf("resource group deletion has failed: status=%s", *status.Brief.Status)
			if status.Brief.FailReason != nil {
				msg += fmt.Sprintf(", message=%s", *status.Brief.FailReason)
			}
			return fmt.Errorf("%s", msg)
		}
	}

	return nil
}
