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

//nolint:all
package assets

import (
	"embed"
	"fmt"

	hwmgrpluginoranopenshiftiov1alpha1 "github.com/openshift-kni/oran-hwmgr-plugin/api/hwmgr-plugin/v1alpha1"
	imsv1alpha1 "github.com/openshift-kni/oran-o2ims/api/hardwaremanagement/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (

	//go:embed manifests/*
	manifests embed.FS

	sch  = runtime.NewScheme()
	cdcs = serializer.NewCodecFactory(sch)
)

func InitCodecs() error {
	if err := corev1.AddToScheme(sch); err != nil {
		return fmt.Errorf("%s failed with error: (%w)", "corev1", err)
	}
	if err := hwmgrpluginoranopenshiftiov1alpha1.AddToScheme(sch); err != nil {
		return fmt.Errorf("%s failed with error: (%w)", "hwmgrv1alpha1", err)
	}
	if err := imsv1alpha1.AddToScheme(sch); err != nil {
		return fmt.Errorf("%s failed with error: (%w)", "imsv1alpha1", err)
	}
	return nil
}

func GetConfigmapFromFile(name string) (*corev1.ConfigMap, error) {
	configmapBytes, err := manifests.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("%s failed with error: (%w)", "readfile", err)
	}

	configmapObject, err := runtime.Decode(cdcs.UniversalDecoder(corev1.SchemeGroupVersion), configmapBytes)
	if err != nil {
		return nil, fmt.Errorf("%s failed with error: (%w)", "decode", err)
	}

	return configmapObject.(*corev1.ConfigMap), nil
}

func GetHardwareManageFromFile(name string) (*hwmgrpluginoranopenshiftiov1alpha1.HardwareManager, error) {
	hardwaremgrBytes, err := manifests.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("%s failed with error: (%w)", "readfile", err)
	}

	hardwaremgrObject, err := runtime.Decode(cdcs.UniversalDecoder(hwmgrpluginoranopenshiftiov1alpha1.GroupVersion), hardwaremgrBytes)
	if err != nil {
		return nil, fmt.Errorf("%s failed with error: (%w)", "decode", err)
	}

	return hardwaremgrObject.(*hwmgrpluginoranopenshiftiov1alpha1.HardwareManager), nil
}

func GetNodePoolFromFile(name string) (*imsv1alpha1.NodePool, error) {
	nodepoolBytes, err := manifests.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("%s failed with error: (%w)", "readfile", err)
	}

	nodepoolObject, err := runtime.Decode(cdcs.UniversalDecoder(imsv1alpha1.GroupVersion), nodepoolBytes)
	if err != nil {
		return nil, fmt.Errorf("%s failed with error: (%w)", "decode", err)
	}

	return nodepoolObject.(*imsv1alpha1.NodePool), nil
}
