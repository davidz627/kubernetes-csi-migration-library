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

package translate

import (
	"reflect"
	"testing"

	"k8s.io/api/core/v1"
)

func TestTranslationStability(t *testing.T) {
	testCases := []struct {
		name string
		pv   *v1.PersistentVolume
	}{

		{
			name: "GCE PD PV Source",
			pv: &v1.PersistentVolume{
				Spec: v1.PersistentVolumeSpec{
					PersistentVolumeSource: v1.PersistentVolumeSource{
						GCEPersistentDisk: &v1.GCEPersistentDiskVolumeSource{
							PDName:    "test-disk",
							FSType:    "ext4",
							Partition: 0,
							ReadOnly:  false,
						},
					},
				},
			},
		},
	}
	for _, test := range testCases {
		t.Logf("Testing %v", test.name)
		csiSource, err := TranslateInTreePVToCSI(test.pv)
		if err != nil {
			t.Errorf("Error when translating to CSI: %v", err)
		}
		newPV, err := TranslateCSIPVToInTree(csiSource)
		if err != nil {
			t.Errorf("Error when translating CSI Source to in tree volume: %v", err)
		}
		if !reflect.DeepEqual(newPV, test.pv) {
			t.Errorf("Volumes after translation and back not equal:\n\nOriginal Volume: %#v\n\nRound-trip Volume: %#v", test.pv, newPV)
		}
	}

}
