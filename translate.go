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
	"fmt"

	"github.com/davidz627/kubernetes-csi-migration-library/plugins"
	"k8s.io/api/core/v1"
)

var (
	inTreePlugins = map[string]plugins.InTreePlugin{
		plugins.GCEPDDriverName: &plugins.GCEPD{},
	}
)

// TranslateToCSI takes a volume.Spec and will translate it to a
// CSIPersistentVolumeSource if the translation logic for that
// specific in-tree volume spec has been implemented
func TranslateToCSI(source *v1.PersistentVolumeSource) (*v1.CSIPersistentVolumeSource, error) {
	for _, curPlugin := range inTreePlugins {
		if curPlugin.CanSupport(source) {
			return curPlugin.TranslateToCSI(source)
		}
	}
	return nil, fmt.Errorf("could not find in-tree plugin translation logic for %#v", source)
}

// TranslateToIntree takes a CSIPersistentVolumeSource and will translate
// it to a volume.Spec for the specific in-tree volume specified by
//`inTreePlugin`, if that translation logic has been implemented
func TranslateToInTree(source *v1.CSIPersistentVolumeSource) (*v1.PersistentVolumeSource, error) {
	for driverName, curPlugin := range inTreePlugins {
		if source.Driver == driverName {
			return curPlugin.TranslateToInTree(source)
		}
	}
	return nil, fmt.Errorf("could not find in-tree plugin translation logic for %s", source.Driver)
}
