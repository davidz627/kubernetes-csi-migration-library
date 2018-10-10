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
)

const (
	GCEPDDriverName = "com.google.csi.gcepd"

	// Volume ID Expected Format
	// "projects/{projectName}/zones/{zoneName}/disks/{diskName}"
	volIDZonalFmt = "projects/%s/zones/%s/disks/%s"
	// "projects/{projectName}/regions/{regionName}/disks/{diskName}"
	volIDRegionalFmt = "projects/%s/regions/%s/disks/%s"

	UnspecifiedValue = "UNSPECIFIED"
)

type GCEPD struct{}

// TranslateToCSI takes a volume.Spec and will translate it to a
// CSIPersistentVolumeSource if the translation logic for that
// specific in-tree volume spec has been implemented
func (g *GCEPD) TranslatePVSourceToCSI(pvSource *v1.PersistentVolumeSource) (*v1.CSIPersistentVolumeSource, error) {
	// TODO: Check PV Failure domain zone, theres a util function "isregionalPD" or something
	// that can tell me whether this is regional or not
	if pvSource != nil && pvSource.GCEPersistentDisk != nil {
		pdSource := pvSource.GCEPersistentDisk
		csiSource := &v1.CSIPersistentVolumeSource{
			Driver:       GCEPDDriverName,
			VolumeHandle: generateUnderspecifiedVolumeID(pdSource.PDName, true /* isZonal */),
			ReadOnly:     pdSource.ReadOnly,
			FSType:       pdSource.FSType,
			VolumeAttributes: map[string]string{
				"partition": strconv.FormatInt(int64(pdSource.Partition), 10),
			},
		}
		return csiSource, nil
	}
	return nil, fmt.Errorf("spec %v does not represent a GCE PD Persistent Volume", pvSource)
}

// TranslateToIntree takes a CSIPersistentVolumeSource and will translate
// it to a volume.Spec for the specific in-tree volume specified by
//`inTreePlugin`, if that translation logic has been implemented
func (g *GCEPD) TranslatePVSourceToInTree(source *v1.CSIPersistentVolumeSource) (*v1.PersistentVolumeSource, error) {
	key, err := common.VolumeIDToKey(source.VolumeHandle)
	if err != nil {
		return nil, fmt.Errorf("failed to translate volume handle %v to key: %v", source.VolumeHandle, err)
	}
	pvSource := &v1.PersistentVolumeSource{
		GCEPersistentDisk: &v1.GCEPersistentDiskVolumeSource{
			PDName:   key.Name,
			FSType:   source.FSType,
			ReadOnly: source.ReadOnly,
		},
	}

	if partition, ok := source.VolumeAttributes["partition"]; ok {
		partInt, err := strconv.Atoi(partition)
		if err != nil {
			return nil, fmt.Errorf("Failed to convert partition %v to integer: %v", partition, err)
		}
		pvSource.GCEPersistentDisk.Partition = int32(partInt)
	}
	return pvSource, nil
}

// CanSupport tests whether the plugin supports a given volume
// specification from the API.  The spec pointer should be considered
// const.
func (g *GCEPD) CanSupport(source *v1.PersistentVolumeSource) bool {
	return (source != nil && source.GCEPersistentDisk != nil)
}

// TODO: Replace this with the imported one from GCE PD CSI Driver when
// the driver removes all k8s/k8s dependencies
func generateUnderspecifiedVolumeID(diskName string, isZonal bool) string {
	if isZonal {
		return fmt.Sprintf(volIDZonalFmt, UnspecifiedValue, UnspecifiedValue, diskName)
	}
	return fmt.Sprintf(volIDRegionalFmt, UnspecifiedValue, UnspecifiedValue, diskName)
}
