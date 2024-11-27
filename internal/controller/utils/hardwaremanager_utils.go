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

package utils

import (
	"context"
	"fmt"

	pluginv1alpha1 "github.com/openshift-kni/oran-hwmgr-plugin/api/hwmgr-plugin/v1alpha1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetHardwareManagerValidationCondition(hwmgr *pluginv1alpha1.HardwareManager) *metav1.Condition {
	return meta.FindStatusCondition(
		hwmgr.Status.Conditions,
		string(pluginv1alpha1.ConditionTypes.Validation))
}

func IsHardwareManagerValidationCompleted(hwmgr *pluginv1alpha1.HardwareManager) bool {
	validationCondition := GetHardwareManagerValidationCondition(hwmgr)
	if validationCondition != nil && validationCondition.Status == metav1.ConditionTrue {
		return true
	}

	return false
}

func IsHardwareManagerValidationFailed(hwmgr *pluginv1alpha1.HardwareManager) bool {
	validationCondition := GetHardwareManagerValidationCondition(hwmgr)
	if validationCondition != nil && validationCondition.Reason == string(pluginv1alpha1.ConditionReasons.Failed) {
		return true
	}

	return false
}

func UpdateHardwareManagerStatusCondition(
	ctx context.Context,
	c client.Client,
	hwmgr *pluginv1alpha1.HardwareManager,
	conditionType pluginv1alpha1.ConditionType,
	conditionReason pluginv1alpha1.ConditionReason,
	conditionStatus metav1.ConditionStatus,
	message string) error {

	SetStatusCondition(&hwmgr.Status.Conditions,
		string(conditionType),
		string(conditionReason),
		conditionStatus,
		message)

	if err := UpdateK8sCRStatus(ctx, c, hwmgr); err != nil {
		return fmt.Errorf("failed to update hwmgr status %s: %w", hwmgr.Name, err)
	}

	return nil
}
