/*
Copyright 2018 The Kubernetes Authors.

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

package plugins

import (
	"fmt"
	"strconv"

	"k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/volume"

	"sigs.k8s.io/gcp-compute-persistent-disk-csi-driver/pkg/common"
)

const (
	GCEPDDriverName = "com.google.csi.gcepd"
)

type GCEPD struct{}

// TranslateToCSI takes a volume.Spec and will translate it to a
// CSIPersistentVolumeSource if the translation logic for that
// specific in-tree volume spec has been implemented
func (g *GCEPD) TranslateToCSI(spec *volume.Spec) (*v1.CSIPersistentVolumeSource, error) {
	// TODO: Check PV Failure domain zone, theres a util function "isregionalPD" or something
	// that can tell me whether this is regional or not
	if spec.PersistentVolume != nil && spec.PersistentVolume.Spec.GCEPersistentDisk != nil {
		pdSource := spec.PersistentVolume.Spec.GCEPersistentDisk
		csiSource := &v1.CSIPersistentVolumeSource{
			Driver:       GCEPDDriverName,
			VolumeHandle: common.GenerateUnderspecifiedVolumeID(pdSource.PDName, true /* isZonal */),
			ReadOnly:     pdSource.ReadOnly,
			FSType:       pdSource.FSType,
			VolumeAttributes: map[string]string{
				"partition": strconv.FormatInt(int64(pdSource.Partition), 10),
			},
		}
		return csiSource, nil
	} else if spec.Volume != nil && spec.Volume.GCEPersistentDisk != nil {
		return nil, fmt.Errorf("In-line volume migration is not yet supported")
	}
	return nil, fmt.Errorf("spec %v does not represent a GCE PD Volume", spec)
}

// TranslateToIntree takes a CSIPersistentVolumeSource and will translate
// it to a volume.Spec for the specific in-tree volume specified by
//`inTreePlugin`, if that translation logic has been implemented
func (g *GCEPD) TranslateToInTree(source *v1.CSIPersistentVolumeSource) (*volume.Spec, error) {
	key, err := common.VolumeIDToKey(source.VolumeHandle)
	if err != nil {
		return nil, fmt.Errorf("failed to translate volume handle %v to key: %v", source.VolumeHandle, err)
	}
	spec := &volume.Spec{
		PersistentVolume: &v1.PersistentVolume{
			Spec: v1.PersistentVolumeSpec{
				PersistentVolumeSource: v1.PersistentVolumeSource{
					GCEPersistentDisk: &v1.GCEPersistentDiskVolumeSource{
						PDName:   key.Name,
						FSType:   source.FSType,
						ReadOnly: source.ReadOnly,
					},
				},
			},
		},
	}
	if partition, ok := source.VolumeAttributes["partition"]; ok {
		partInt, err := strconv.Atoi(partition)
		if err != nil {
			return nil, fmt.Errorf("Failed to convert partition %v to integer: %v", partition, err)
		}
		spec.PersistentVolume.Spec.PersistentVolumeSource.GCEPersistentDisk.Partition = int32(partInt)
	}
	return spec, nil
}

// CanSupport tests whether the plugin supports a given volume
// specification from the API.  The spec pointer should be considered
// const.
func (g *GCEPD) CanSupport(spec *volume.Spec) bool {
	return (spec.PersistentVolume != nil && spec.PersistentVolume.Spec.GCEPersistentDisk != nil) ||
		(spec.Volume != nil && spec.Volume.GCEPersistentDisk != nil)
}
