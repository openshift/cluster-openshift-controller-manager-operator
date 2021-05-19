// Code generated for package v311_00_assets by go-bindata DO NOT EDIT. (@generated)
// sources:
// bindata/v3.11.0/config/defaultconfig.yaml
// bindata/v3.11.0/openshift-controller-manager/build-config-change-controller-clusterrole.yaml
// bindata/v3.11.0/openshift-controller-manager/build-config-change-controller-clusterrolebinding.yaml
// bindata/v3.11.0/openshift-controller-manager/build-controller-clusterrole.yaml
// bindata/v3.11.0/openshift-controller-manager/build-controller-clusterrolebinding.yaml
// bindata/v3.11.0/openshift-controller-manager/buildconfigstatus-clusterrole.yaml
// bindata/v3.11.0/openshift-controller-manager/buildconfigstatus-clusterrolebinding.yaml
// bindata/v3.11.0/openshift-controller-manager/cm.yaml
// bindata/v3.11.0/openshift-controller-manager/default-rolebindings-controller-clusterrole.yaml
// bindata/v3.11.0/openshift-controller-manager/default-rolebindings-controller-clusterrolebinding-deployer.yaml
// bindata/v3.11.0/openshift-controller-manager/default-rolebindings-controller-clusterrolebinding-image-builder.yaml
// bindata/v3.11.0/openshift-controller-manager/default-rolebindings-controller-clusterrolebinding-image-puller.yaml
// bindata/v3.11.0/openshift-controller-manager/default-rolebindings-controller-clusterrolebinding.yaml
// bindata/v3.11.0/openshift-controller-manager/deployer-controller-clusterrole.yaml
// bindata/v3.11.0/openshift-controller-manager/deployer-controller-clusterrolebinding.yaml
// bindata/v3.11.0/openshift-controller-manager/deploymentconfig-controller-clusterrole.yaml
// bindata/v3.11.0/openshift-controller-manager/deploymentconfig-controller-clusterrolebinding.yaml
// bindata/v3.11.0/openshift-controller-manager/ds.yaml
// bindata/v3.11.0/openshift-controller-manager/image-import-controller-clusterrole.yaml
// bindata/v3.11.0/openshift-controller-manager/image-import-controller-clusterrolebinding.yaml
// bindata/v3.11.0/openshift-controller-manager/image-trigger-controller-clusterrole.yaml
// bindata/v3.11.0/openshift-controller-manager/image-trigger-controller-clusterrolebinding.yaml
// bindata/v3.11.0/openshift-controller-manager/informer-clusterrole.yaml
// bindata/v3.11.0/openshift-controller-manager/informer-clusterrolebinding.yaml
// bindata/v3.11.0/openshift-controller-manager/ingress-to-route-controller-clusterrole.yaml
// bindata/v3.11.0/openshift-controller-manager/ingress-to-route-controller-clusterrolebinding.yaml
// bindata/v3.11.0/openshift-controller-manager/leader-role.yaml
// bindata/v3.11.0/openshift-controller-manager/leader-rolebinding.yaml
// bindata/v3.11.0/openshift-controller-manager/ns.yaml
// bindata/v3.11.0/openshift-controller-manager/old-leader-role.yaml
// bindata/v3.11.0/openshift-controller-manager/old-leader-rolebinding.yaml
// bindata/v3.11.0/openshift-controller-manager/openshift-global-ca-cm.yaml
// bindata/v3.11.0/openshift-controller-manager/openshift-service-ca-cm.yaml
// bindata/v3.11.0/openshift-controller-manager/origin-namespace-controller-clusterrole.yaml
// bindata/v3.11.0/openshift-controller-manager/origin-namespace-controller-clusterrolebinding.yaml
// bindata/v3.11.0/openshift-controller-manager/sa.yaml
// bindata/v3.11.0/openshift-controller-manager/separate-sa-role.yaml
// bindata/v3.11.0/openshift-controller-manager/separate-sa-rolebinding.yaml
// bindata/v3.11.0/openshift-controller-manager/service-ingress-ip-controller-clusterrole.yaml
// bindata/v3.11.0/openshift-controller-manager/service-ingress-ip-controller-clusterrolebinding.yaml
// bindata/v3.11.0/openshift-controller-manager/serviceaccount-controller-clusterrole.yaml
// bindata/v3.11.0/openshift-controller-manager/serviceaccount-controller-clusterrolebinding.yaml
// bindata/v3.11.0/openshift-controller-manager/serviceaccount-pull-secrets-controller-clusterrole.yaml
// bindata/v3.11.0/openshift-controller-manager/serviceaccount-pull-secrets-controller-clusterrolebinding.yaml
// bindata/v3.11.0/openshift-controller-manager/servicemonitor-role.yaml
// bindata/v3.11.0/openshift-controller-manager/servicemonitor-rolebinding.yaml
// bindata/v3.11.0/openshift-controller-manager/svc.yaml
// bindata/v3.11.0/openshift-controller-manager/template-instance-controller-clusterrole.yaml
// bindata/v3.11.0/openshift-controller-manager/template-instance-controller-clusterrolebinding-admin.yaml
// bindata/v3.11.0/openshift-controller-manager/template-instance-controller-clusterrolebinding.yaml
// bindata/v3.11.0/openshift-controller-manager/template-instance-finalizer-controller-clusterrole.yaml
// bindata/v3.11.0/openshift-controller-manager/template-instance-finalizer-controller-clusterrolebinding-admin.yaml
// bindata/v3.11.0/openshift-controller-manager/template-instance-finalizer-controller-clusterrolebinding.yaml
// bindata/v3.11.0/openshift-controller-manager/tokenreview-clusterrole.yaml
// bindata/v3.11.0/openshift-controller-manager/tokenreview-clusterrolebinding.yaml
// bindata/v3.11.0/openshift-controller-manager/unidling-controller-clusterrole.yaml
// bindata/v3.11.0/openshift-controller-manager/unidling-controller-clusterrolebinding.yaml
package v311_00_assets

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

// Name return file name
func (fi bindataFileInfo) Name() string {
	return fi.name
}

// Size return file size
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}

// Mode return file mode
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}

// Mode return file modify time
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir return file whether a directory
func (fi bindataFileInfo) IsDir() bool {
	return fi.mode&os.ModeDir != 0
}

// Sys return file is sys mode
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _v3110ConfigDefaultconfigYaml = []byte(`apiVersion: openshiftcontrolplane.config.openshift.io/v1
kind: OpenShiftControllerManagerConfig
`)

func v3110ConfigDefaultconfigYamlBytes() ([]byte, error) {
	return _v3110ConfigDefaultconfigYaml, nil
}

func v3110ConfigDefaultconfigYaml() (*asset, error) {
	bytes, err := v3110ConfigDefaultconfigYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/config/defaultconfig.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerBuildConfigChangeControllerClusterroleYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: system:openshift:openshift-controller-manager:build-config-change-controller
rules:
- apiGroups:
  - ""
  - build.openshift.io
  resources:
  - buildconfigs
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  - build.openshift.io
  resources:
  - buildconfigs/instantiate
  verbs:
  - create
- apiGroups:
  - ""
  - build.openshift.io
  resources:
  - builds
  verbs:
  - delete
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - patch
  - update
`)

func v3110OpenshiftControllerManagerBuildConfigChangeControllerClusterroleYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerBuildConfigChangeControllerClusterroleYaml, nil
}

func v3110OpenshiftControllerManagerBuildConfigChangeControllerClusterroleYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerBuildConfigChangeControllerClusterroleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/build-config-change-controller-clusterrole.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerBuildConfigChangeControllerClusterrolebindingYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: system:openshift:openshift-controller-manager:build-config-change-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:openshift:openshift-controller-manager:build-config-change-controller
subjects:
- kind: ServiceAccount
  name: build-config-change-controller
  namespace: openshift-infra
`)

func v3110OpenshiftControllerManagerBuildConfigChangeControllerClusterrolebindingYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerBuildConfigChangeControllerClusterrolebindingYaml, nil
}

func v3110OpenshiftControllerManagerBuildConfigChangeControllerClusterrolebindingYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerBuildConfigChangeControllerClusterrolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/build-config-change-controller-clusterrolebinding.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerBuildControllerClusterroleYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: system:openshift:openshift-controller-manager:build-controller
rules:
- apiGroups:
  - ""
  - build.openshift.io
  resources:
  - builds
  verbs:
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - ""
  - build.openshift.io
  resources:
  - builds/finalizers
  verbs:
  - update
- apiGroups:
  - ""
  - build.openshift.io
  resources:
  - buildconfigs
  verbs:
  - get
- apiGroups:
  - ""
  - build.openshift.io
  resources:
  - builds/custom
  - builds/docker
  - builds/jenkinspipeline
  - builds/optimizeddocker
  - builds/source
  verbs:
  - create
- apiGroups:
  - ""
  - image.openshift.io
  resources:
  - imagestreams
  verbs:
  - get
  - list
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - get
  - list
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - create
  - get
  - list
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - create
  - delete
  - get
  - list
- apiGroups:
  - ""
  resources:
  - namespaces
  verbs:
  - get
- apiGroups:
  - ""
  resources:
  - serviceaccounts
  verbs:
  - get
  - list
- apiGroups:
  - ""
  - security.openshift.io
  resources:
  - podsecuritypolicysubjectreviews
  verbs:
  - create
- apiGroups:
  - config.openshift.io
  resources:
  - builds
  verbs:
  - get
  - list
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - patch
  - update
`)

func v3110OpenshiftControllerManagerBuildControllerClusterroleYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerBuildControllerClusterroleYaml, nil
}

func v3110OpenshiftControllerManagerBuildControllerClusterroleYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerBuildControllerClusterroleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/build-controller-clusterrole.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerBuildControllerClusterrolebindingYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: system:openshift:openshift-controller-manager:build-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:openshift:openshift-controller-manager:build-controller
subjects:
- kind: ServiceAccount
  name: build-controller
  namespace: openshift-infra
`)

func v3110OpenshiftControllerManagerBuildControllerClusterrolebindingYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerBuildControllerClusterrolebindingYaml, nil
}

func v3110OpenshiftControllerManagerBuildControllerClusterrolebindingYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerBuildControllerClusterrolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/build-controller-clusterrolebinding.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerBuildconfigstatusClusterroleYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: system:openshift:openshift-controller-manager:update-buildconfig-status
rules:
- apiGroups:
  - build.openshift.io
  resources:
  - buildconfigs/status
  verbs:
  - "*"`)

func v3110OpenshiftControllerManagerBuildconfigstatusClusterroleYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerBuildconfigstatusClusterroleYaml, nil
}

func v3110OpenshiftControllerManagerBuildconfigstatusClusterroleYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerBuildconfigstatusClusterroleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/buildconfigstatus-clusterrole.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerBuildconfigstatusClusterrolebindingYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: system:openshift:openshift-controller-manager:update-buildconfig-status
roleRef:
  kind: ClusterRole
  name: system:openshift:openshift-controller-manager:update-buildconfig-status
subjects:
- kind: ServiceAccount
  namespace: openshift-controller-manager
  name: openshift-controller-manager-sa
- kind: ServiceAccount
  namespace: openshift-infra
  name: build-config-change-controller`)

func v3110OpenshiftControllerManagerBuildconfigstatusClusterrolebindingYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerBuildconfigstatusClusterrolebindingYaml, nil
}

func v3110OpenshiftControllerManagerBuildconfigstatusClusterrolebindingYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerBuildconfigstatusClusterrolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/buildconfigstatus-clusterrolebinding.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerCmYaml = []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  namespace: openshift-controller-manager
  name: config
data:
  config.yaml:
`)

func v3110OpenshiftControllerManagerCmYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerCmYaml, nil
}

func v3110OpenshiftControllerManagerCmYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerCmYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/cm.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerDefaultRolebindingsControllerClusterroleYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: system:openshift:openshift-controller-manager:default-rolebindings-controller
rules:
- apiGroups:
  - rbac.authorization.k8s.io
  resources:
  - rolebindings
  verbs:
  - create
- apiGroups:
  - ""
  resources:
  - namespaces
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - rbac.authorization.k8s.io
  resources:
  - rolebindings
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - patch
  - update
`)

func v3110OpenshiftControllerManagerDefaultRolebindingsControllerClusterroleYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerDefaultRolebindingsControllerClusterroleYaml, nil
}

func v3110OpenshiftControllerManagerDefaultRolebindingsControllerClusterroleYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerDefaultRolebindingsControllerClusterroleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/default-rolebindings-controller-clusterrole.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerDefaultRolebindingsControllerClusterrolebindingDeployerYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: system:openshift:openshift-controller-manager:deployer
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:deployer
subjects:
- kind: ServiceAccount
  name: default-rolebindings-controller
  namespace: openshift-infra
`)

func v3110OpenshiftControllerManagerDefaultRolebindingsControllerClusterrolebindingDeployerYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerDefaultRolebindingsControllerClusterrolebindingDeployerYaml, nil
}

func v3110OpenshiftControllerManagerDefaultRolebindingsControllerClusterrolebindingDeployerYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerDefaultRolebindingsControllerClusterrolebindingDeployerYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/default-rolebindings-controller-clusterrolebinding-deployer.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerDefaultRolebindingsControllerClusterrolebindingImageBuilderYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: system:openshift:openshift-controller-manager:image-builder
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:image-builder
subjects:
- kind: ServiceAccount
  name: default-rolebindings-controller
  namespace: openshift-infra
`)

func v3110OpenshiftControllerManagerDefaultRolebindingsControllerClusterrolebindingImageBuilderYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerDefaultRolebindingsControllerClusterrolebindingImageBuilderYaml, nil
}

func v3110OpenshiftControllerManagerDefaultRolebindingsControllerClusterrolebindingImageBuilderYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerDefaultRolebindingsControllerClusterrolebindingImageBuilderYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/default-rolebindings-controller-clusterrolebinding-image-builder.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerDefaultRolebindingsControllerClusterrolebindingImagePullerYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: system:openshift:openshift-controller-manager:image-puller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:image-puller
subjects:
- kind: ServiceAccount
  name: default-rolebindings-controller
  namespace: openshift-infra
`)

func v3110OpenshiftControllerManagerDefaultRolebindingsControllerClusterrolebindingImagePullerYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerDefaultRolebindingsControllerClusterrolebindingImagePullerYaml, nil
}

func v3110OpenshiftControllerManagerDefaultRolebindingsControllerClusterrolebindingImagePullerYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerDefaultRolebindingsControllerClusterrolebindingImagePullerYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/default-rolebindings-controller-clusterrolebinding-image-puller.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerDefaultRolebindingsControllerClusterrolebindingYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: system:openshift:openshift-controller-manager:default-rolebindings-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:openshift:openshift-controller-manager:default-rolebindings-controller
subjects:
- kind: ServiceAccount
  name: default-rolebindings-controller
  namespace: openshift-infra
`)

func v3110OpenshiftControllerManagerDefaultRolebindingsControllerClusterrolebindingYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerDefaultRolebindingsControllerClusterrolebindingYaml, nil
}

func v3110OpenshiftControllerManagerDefaultRolebindingsControllerClusterrolebindingYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerDefaultRolebindingsControllerClusterrolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/default-rolebindings-controller-clusterrolebinding.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerDeployerControllerClusterroleYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: system:openshift:openshift-controller-manager:deployer-controller
rules:
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - watch
- apiGroups:
  - ""
  resources:
  - replicationcontrollers
  verbs:
  - delete
- apiGroups:
  - ""
  resources:
  - replicationcontrollers
  verbs:
  - get
  - list
  - update
  - watch
- apiGroups:
  - ""
  resources:
  - replicationcontrollers/scale
  verbs:
  - get
  - update
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - patch
  - update
  `)

func v3110OpenshiftControllerManagerDeployerControllerClusterroleYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerDeployerControllerClusterroleYaml, nil
}

func v3110OpenshiftControllerManagerDeployerControllerClusterroleYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerDeployerControllerClusterroleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/deployer-controller-clusterrole.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerDeployerControllerClusterrolebindingYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: system:openshift:openshift-controller-manager:deployer-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:openshift:openshift-controller-manager:deployer-controller
subjects:
- kind: ServiceAccount
  name: deployer-controller
  namespace: openshift-infra
`)

func v3110OpenshiftControllerManagerDeployerControllerClusterrolebindingYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerDeployerControllerClusterrolebindingYaml, nil
}

func v3110OpenshiftControllerManagerDeployerControllerClusterrolebindingYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerDeployerControllerClusterrolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/deployer-controller-clusterrolebinding.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerDeploymentconfigControllerClusterroleYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: system:openshift:openshift-controller-manager:deploymentconfig-controller
rules:
- apiGroups:
  - ""
  resources:
  - replicationcontrollers
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - ""
  resources:
  - replicationcontrollers/scale
  verbs:
  - get
  - update
- apiGroups:
  - ""
  - apps.openshift.io
  resources:
  - deploymentconfigs/status
  verbs:
  - update
- apiGroups:
  - ""
  - apps.openshift.io
  resources:
  - deploymentconfigs/finalizers
  verbs:
  - update
- apiGroups:
  - ""
  - apps.openshift.io
  resources:
  - deploymentconfigs
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - patch
  - update
`)

func v3110OpenshiftControllerManagerDeploymentconfigControllerClusterroleYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerDeploymentconfigControllerClusterroleYaml, nil
}

func v3110OpenshiftControllerManagerDeploymentconfigControllerClusterroleYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerDeploymentconfigControllerClusterroleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/deploymentconfig-controller-clusterrole.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerDeploymentconfigControllerClusterrolebindingYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: system:openshift:openshift-controller-manager:deploymentconfig-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:openshift:openshift-controller-manager:deploymentconfig-controller
subjects:
- kind: ServiceAccount
  name: deploymentconfig-controller
  namespace: openshift-infra
`)

func v3110OpenshiftControllerManagerDeploymentconfigControllerClusterrolebindingYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerDeploymentconfigControllerClusterrolebindingYaml, nil
}

func v3110OpenshiftControllerManagerDeploymentconfigControllerClusterrolebindingYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerDeploymentconfigControllerClusterrolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/deploymentconfig-controller-clusterrolebinding.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerDsYaml = []byte(`apiVersion: apps/v1
kind: DaemonSet
metadata:
  namespace: openshift-controller-manager
  name: controller-manager
  labels:
    app: openshift-controller-manager
    controller-manager: "true"
spec:
  updateStrategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 3
  selector:
    matchLabels:
      app: openshift-controller-manager
      controller-manager: "true"
  template:
    metadata:
      name: openshift-controller-manager
      annotations:
        target.workload.openshift.io/management: '{"effect": "PreferredDuringScheduling"}'
      labels:
        app: openshift-controller-manager
        controller-manager: "true"
    spec:
      priorityClassName: system-node-critical 
      serviceAccountName: openshift-controller-manager-sa
      containers:
      - name: controller-manager
        image: ${IMAGE}
        imagePullPolicy: IfNotPresent
        command: ["openshift-controller-manager", "start"]
        args:
        - "--config=/var/run/configmaps/config/config.yaml"
        resources:
          requests:
            memory: 100Mi
            cpu: 100m
        ports:
        - containerPort: 8443
        terminationMessagePolicy: FallbackToLogsOnError
        volumeMounts:
        - mountPath: /var/run/configmaps/config
          name: config
        - mountPath: /var/run/configmaps/client-ca
          name: client-ca
        - mountPath: /var/run/secrets/serving-cert
          name: serving-cert
        - mountPath: /etc/pki/ca-trust/extracted/pem
          name: proxy-ca-bundles
      volumes:
      - name: config
        configMap:
          name: config
      - name: client-ca
        configMap:
          name: client-ca
      - name: serving-cert
        secret:
          secretName: serving-cert
      - name: proxy-ca-bundles
        configMap:
          name: openshift-global-ca
          items:
            - key: ca-bundle.crt
              path: tls-ca-bundle.pem
      nodeSelector:
        node-role.kubernetes.io/master: ""
      tolerations:
      - operator: Exists
`)

func v3110OpenshiftControllerManagerDsYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerDsYaml, nil
}

func v3110OpenshiftControllerManagerDsYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerDsYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/ds.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerImageImportControllerClusterroleYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: system:openshift:openshift-controller-manager:image-import-controller
rules:
- apiGroups:
  - ""
  - image.openshift.io
  resources:
  - imagestreams
  verbs:
  - create
  - get
  - list
  - update
  - watch
- apiGroups:
  - ""
  - image.openshift.io
  resources:
  - images
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - ""
  - image.openshift.io
  resources:
  - imagestreamimports
  verbs:
  - create
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - patch
  - update
`)

func v3110OpenshiftControllerManagerImageImportControllerClusterroleYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerImageImportControllerClusterroleYaml, nil
}

func v3110OpenshiftControllerManagerImageImportControllerClusterroleYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerImageImportControllerClusterroleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/image-import-controller-clusterrole.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerImageImportControllerClusterrolebindingYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: system:openshift:openshift-controller-manager:image-import-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:openshift:openshift-controller-manager:image-import-controller
subjects:
- kind: ServiceAccount
  name: image-import-controller
  namespace: openshift-infra
`)

func v3110OpenshiftControllerManagerImageImportControllerClusterrolebindingYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerImageImportControllerClusterrolebindingYaml, nil
}

func v3110OpenshiftControllerManagerImageImportControllerClusterrolebindingYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerImageImportControllerClusterrolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/image-import-controller-clusterrolebinding.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerImageTriggerControllerClusterroleYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: system:openshift:openshift-controller-manager:image-trigger-controller
rules:
- apiGroups:
  - ""
  - image.openshift.io
  resources:
  - imagestreams
  verbs:
  - list
  - watch
- apiGroups:
  - extensions
  resources:
  - daemonsets
  verbs:
  - get
  - update
- apiGroups:
  - apps
  - extensions
  resources:
  - deployments
  verbs:
  - get
  - update
- apiGroups:
  - apps
  resources:
  - statefulsets
  verbs:
  - get
  - update
- apiGroups:
  - batch
  resources:
  - cronjobs
  verbs:
  - get
  - update
- apiGroups:
  - ""
  - apps.openshift.io
  resources:
  - deploymentconfigs
  verbs:
  - get
  - update
- apiGroups:
  - ""
  - build.openshift.io
  resources:
  - buildconfigs/instantiate
  verbs:
  - create
- apiGroups:
  - ""
  - build.openshift.io
  resources:
  - builds/custom
  - builds/docker
  - builds/jenkinspipeline
  - builds/optimizeddocker
  - builds/source
  verbs:
  - create
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - patch
  - update
`)

func v3110OpenshiftControllerManagerImageTriggerControllerClusterroleYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerImageTriggerControllerClusterroleYaml, nil
}

func v3110OpenshiftControllerManagerImageTriggerControllerClusterroleYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerImageTriggerControllerClusterroleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/image-trigger-controller-clusterrole.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerImageTriggerControllerClusterrolebindingYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: system:openshift:openshift-controller-manager:image-trigger-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:openshift:openshift-controller-manager:image-trigger-controller
subjects:
- kind: ServiceAccount
  name: image-trigger-controller
  namespace: openshift-infra
`)

func v3110OpenshiftControllerManagerImageTriggerControllerClusterrolebindingYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerImageTriggerControllerClusterrolebindingYaml, nil
}

func v3110OpenshiftControllerManagerImageTriggerControllerClusterrolebindingYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerImageTriggerControllerClusterrolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/image-trigger-controller-clusterrolebinding.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerInformerClusterroleYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: system:openshift:openshift-controller-manager
rules:
# we run cluster resource quota, so we have to be able to see all resources
- apiGroups:
  - "*"
  resources:
  - "*"
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  - events.k8s.io
  resources:
  - events
  verbs:
  - create
  - patch
  - update
`)

func v3110OpenshiftControllerManagerInformerClusterroleYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerInformerClusterroleYaml, nil
}

func v3110OpenshiftControllerManagerInformerClusterroleYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerInformerClusterroleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/informer-clusterrole.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerInformerClusterrolebindingYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: system:openshift:openshift-controller-manager
roleRef:
  kind: ClusterRole
  name: system:openshift:openshift-controller-manager
subjects:
- kind: ServiceAccount
  namespace: openshift-controller-manager
  name: openshift-controller-manager-sa
`)

func v3110OpenshiftControllerManagerInformerClusterrolebindingYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerInformerClusterrolebindingYaml, nil
}

func v3110OpenshiftControllerManagerInformerClusterrolebindingYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerInformerClusterrolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/informer-clusterrolebinding.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerIngressToRouteControllerClusterroleYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: system:openshift:openshift-controller-manager:ingress-to-route-controller
rules:
- apiGroups:
  - ""
  resources:
  - secrets
  - services
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - networking.k8s.io
  resources:
  - ingress
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - networking.k8s.io
  resources:
  - ingresses/status
  verbs:
  - update
- apiGroups:
  - route.openshift.io
  resources:
  - routes
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - route.openshift.io
  resources:
  - routes/custom-host
  verbs:
  - create
  - update
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - patch
  - update
`)

func v3110OpenshiftControllerManagerIngressToRouteControllerClusterroleYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerIngressToRouteControllerClusterroleYaml, nil
}

func v3110OpenshiftControllerManagerIngressToRouteControllerClusterroleYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerIngressToRouteControllerClusterroleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/ingress-to-route-controller-clusterrole.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerIngressToRouteControllerClusterrolebindingYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: system:openshift:openshift-controller-manager:ingress-to-route-controller
roleRef:
  kind: ClusterRole
  name: system:openshift:openshift-controller-manager:ingress-to-route-controller
subjects:
- kind: ServiceAccount
  namespace: openshift-infra
  name: ingress-to-route-controller
`)

func v3110OpenshiftControllerManagerIngressToRouteControllerClusterrolebindingYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerIngressToRouteControllerClusterrolebindingYaml, nil
}

func v3110OpenshiftControllerManagerIngressToRouteControllerClusterrolebindingYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerIngressToRouteControllerClusterrolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/ingress-to-route-controller-clusterrolebinding.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerLeaderRoleYaml = []byte(`# needed to get the legacy lock that we used to use
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: system:openshift:leader-locking-openshift-controller-manager
  namespace: openshift-controller-manager
rules:
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - create
- apiGroups:
  - ""
  resourceNames:
  - openshift-master-controllers
  resources:
  - configmaps
  verbs:
  - get
  - create
  - update
  - patch
`)

func v3110OpenshiftControllerManagerLeaderRoleYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerLeaderRoleYaml, nil
}

func v3110OpenshiftControllerManagerLeaderRoleYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerLeaderRoleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/leader-role.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerLeaderRolebindingYaml = []byte(`# needed to get the legacy lock that we used to use
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  namespace: openshift-controller-manager
  name: system:openshift:leader-locking-openshift-controller-manager
roleRef:
  kind: Role
  name: system:openshift:leader-locking-openshift-controller-manager
subjects:
- kind: ServiceAccount
  namespace: openshift-controller-manager
  name: openshift-controller-manager-sa
`)

func v3110OpenshiftControllerManagerLeaderRolebindingYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerLeaderRolebindingYaml, nil
}

func v3110OpenshiftControllerManagerLeaderRolebindingYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerLeaderRolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/leader-rolebinding.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerNsYaml = []byte(`apiVersion: v1
kind: Namespace
metadata:
  name: openshift-controller-manager
  annotations:
    openshift.io/node-selector: ""
    workload.openshift.io/allowed: "management"
  labels:
    openshift.io/cluster-monitoring: "true"
`)

func v3110OpenshiftControllerManagerNsYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerNsYaml, nil
}

func v3110OpenshiftControllerManagerNsYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerNsYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/ns.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerOldLeaderRoleYaml = []byte(`# needed to get the legacy lock that we used to use
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: system:openshift:leader-locking-openshift-controller-manager
  namespace: kube-system
rules:
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - create
- apiGroups:
  - ""
  resourceNames:
  - openshift-master-controllers
  resources:
  - configmaps
  verbs:
  - get
  - create
  - update
  - patch`)

func v3110OpenshiftControllerManagerOldLeaderRoleYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerOldLeaderRoleYaml, nil
}

func v3110OpenshiftControllerManagerOldLeaderRoleYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerOldLeaderRoleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/old-leader-role.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerOldLeaderRolebindingYaml = []byte(`# needed to get the legacy lock that we used to use
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  namespace: kube-system
  name: system:openshift:leader-locking-openshift-controller-manager
roleRef:
  kind: Role
  name: system:openshift:leader-locking-openshift-controller-manager
subjects:
- kind: ServiceAccount
  namespace: openshift-controller-manager
  name: openshift-controller-manager-sa
`)

func v3110OpenshiftControllerManagerOldLeaderRolebindingYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerOldLeaderRolebindingYaml, nil
}

func v3110OpenshiftControllerManagerOldLeaderRolebindingYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerOldLeaderRolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/old-leader-rolebinding.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerOpenshiftGlobalCaCmYaml = []byte(`kind: ConfigMap
apiVersion: v1
metadata:
  name: openshift-global-ca
  namespace: openshift-controller-manager
  labels: 
    config.openshift.io/inject-trusted-cabundle: "true"
data: {}
`)

func v3110OpenshiftControllerManagerOpenshiftGlobalCaCmYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerOpenshiftGlobalCaCmYaml, nil
}

func v3110OpenshiftControllerManagerOpenshiftGlobalCaCmYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerOpenshiftGlobalCaCmYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/openshift-global-ca-cm.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerOpenshiftServiceCaCmYaml = []byte(`kind: ConfigMap
apiVersion: v1
metadata:
  name: openshift-service-ca
  namespace: openshift-controller-manager
  annotations: 
    service.beta.openshift.io/inject-cabundle: "true"
data: {}
`)

func v3110OpenshiftControllerManagerOpenshiftServiceCaCmYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerOpenshiftServiceCaCmYaml, nil
}

func v3110OpenshiftControllerManagerOpenshiftServiceCaCmYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerOpenshiftServiceCaCmYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/openshift-service-ca-cm.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerOriginNamespaceControllerClusterroleYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: system:openshift:openshift-controller-manager:origin-namespace-controller
rules:
- apiGroups:
  - ""
  resources:
  - namespaces
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - namespaces/finalize
  - namespaces/status
  verbs:
  - update
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - patch
  - update
`)

func v3110OpenshiftControllerManagerOriginNamespaceControllerClusterroleYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerOriginNamespaceControllerClusterroleYaml, nil
}

func v3110OpenshiftControllerManagerOriginNamespaceControllerClusterroleYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerOriginNamespaceControllerClusterroleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/origin-namespace-controller-clusterrole.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerOriginNamespaceControllerClusterrolebindingYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  annotations:
    rbac.authorization.kubernetes.io/autoupdate: "true"
  creationTimestamp: null
  name: system:openshift:openshift-controller-manager:origin-namespace-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:openshift:openshift-controller-manager:origin-namespace-controller
subjects:
- kind: ServiceAccount
  name: origin-namespace-controller
  namespace: openshift-infra
`)

func v3110OpenshiftControllerManagerOriginNamespaceControllerClusterrolebindingYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerOriginNamespaceControllerClusterrolebindingYaml, nil
}

func v3110OpenshiftControllerManagerOriginNamespaceControllerClusterrolebindingYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerOriginNamespaceControllerClusterrolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/origin-namespace-controller-clusterrolebinding.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerSaYaml = []byte(`apiVersion: v1
kind: ServiceAccount
metadata:
  namespace: openshift-controller-manager
  name: openshift-controller-manager-sa
`)

func v3110OpenshiftControllerManagerSaYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerSaYaml, nil
}

func v3110OpenshiftControllerManagerSaYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerSaYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/sa.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerSeparateSaRoleYaml = []byte(`# needed to support the "use separate service accounts" feature.
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: system:openshift:sa-creating-openshift-controller-manager
  namespace: openshift-infra
rules:
- apiGroups:
  - ""
  resources:
  - serviceaccounts
  verbs:
  - get
  - create
  - update
- apiGroups:
  - ""
  resources:
  - serviceaccounts/token
  verbs:
  - create
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - get
  - list
  - create
`)

func v3110OpenshiftControllerManagerSeparateSaRoleYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerSeparateSaRoleYaml, nil
}

func v3110OpenshiftControllerManagerSeparateSaRoleYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerSeparateSaRoleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/separate-sa-role.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerSeparateSaRolebindingYaml = []byte(`# needed to support the "use separate service accounts" feature.
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  namespace: openshift-infra
  name: system:openshift:sa-creating-openshift-controller-manager
roleRef:
  kind: Role
  name: system:openshift:sa-creating-openshift-controller-manager
subjects:
- kind: ServiceAccount
  namespace: openshift-controller-manager
  name: openshift-controller-manager-sa
`)

func v3110OpenshiftControllerManagerSeparateSaRolebindingYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerSeparateSaRolebindingYaml, nil
}

func v3110OpenshiftControllerManagerSeparateSaRolebindingYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerSeparateSaRolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/separate-sa-rolebinding.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerServiceIngressIpControllerClusterroleYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: system:openshift:openshift-controller-manager:service-ingress-ip-controller
rules:
- apiGroups:
  - ""
  resources:
  - services
  verbs:
  - list
  - update
  - watch
- apiGroups:
  - ""
  resources:
  - services/status
  verbs:
  - update
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - patch
  - update
`)

func v3110OpenshiftControllerManagerServiceIngressIpControllerClusterroleYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerServiceIngressIpControllerClusterroleYaml, nil
}

func v3110OpenshiftControllerManagerServiceIngressIpControllerClusterroleYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerServiceIngressIpControllerClusterroleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/service-ingress-ip-controller-clusterrole.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerServiceIngressIpControllerClusterrolebindingYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: system:openshift:openshift-controller-manager:service-ingress-ip-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:openshift:openshift-controller-manager:service-ingress-ip-controller
subjects:
- kind: ServiceAccount
  name: service-ingress-ip-controller
  namespace: openshift-infra
`)

func v3110OpenshiftControllerManagerServiceIngressIpControllerClusterrolebindingYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerServiceIngressIpControllerClusterrolebindingYaml, nil
}

func v3110OpenshiftControllerManagerServiceIngressIpControllerClusterrolebindingYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerServiceIngressIpControllerClusterrolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/service-ingress-ip-controller-clusterrolebinding.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerServiceaccountControllerClusterroleYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: system:openshift:openshift-controller-manager:serviceaccount-controller
rules:
- apiGroups:
  - ""
  resources:
  - serviceaccounts
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
- apiGroups:
  - ""
  resources:
  - serviceaccounts/token
  verbs:
  - create
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - patch
  - update
`)

func v3110OpenshiftControllerManagerServiceaccountControllerClusterroleYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerServiceaccountControllerClusterroleYaml, nil
}

func v3110OpenshiftControllerManagerServiceaccountControllerClusterroleYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerServiceaccountControllerClusterroleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/serviceaccount-controller-clusterrole.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerServiceaccountControllerClusterrolebindingYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: system:openshift:openshift-controller-manager:serviceaccount-controller
roleRef:
  kind: ClusterRole
  name: system:openshift:openshift-controller-manager:serviceaccount-controller
subjects:
- kind: ServiceAccount
  namespace: openshift-infra
  name: serviceaccount-controller
`)

func v3110OpenshiftControllerManagerServiceaccountControllerClusterrolebindingYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerServiceaccountControllerClusterrolebindingYaml, nil
}

func v3110OpenshiftControllerManagerServiceaccountControllerClusterrolebindingYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerServiceaccountControllerClusterrolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/serviceaccount-controller-clusterrolebinding.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerServiceaccountPullSecretsControllerClusterroleYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: system:openshift:openshift-controller-manager:serviceaccount-pull-secrets-controller
rules:
- apiGroups:
  - ""
  resources:
  - serviceaccounts
  verbs:
  - create
  - get
  - list
  - update
  - watch
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - ""
  resources:
  - services
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - patch
  - update
  `)

func v3110OpenshiftControllerManagerServiceaccountPullSecretsControllerClusterroleYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerServiceaccountPullSecretsControllerClusterroleYaml, nil
}

func v3110OpenshiftControllerManagerServiceaccountPullSecretsControllerClusterroleYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerServiceaccountPullSecretsControllerClusterroleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/serviceaccount-pull-secrets-controller-clusterrole.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerServiceaccountPullSecretsControllerClusterrolebindingYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: system:openshift:openshift-controller-manager:serviceaccount-pull-secrets-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:openshift:openshift-controller-manager:serviceaccount-pull-secrets-controller
subjects:
- kind: ServiceAccount
  name: serviceaccount-pull-secrets-controller
  namespace: openshift-infra
`)

func v3110OpenshiftControllerManagerServiceaccountPullSecretsControllerClusterrolebindingYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerServiceaccountPullSecretsControllerClusterrolebindingYaml, nil
}

func v3110OpenshiftControllerManagerServiceaccountPullSecretsControllerClusterrolebindingYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerServiceaccountPullSecretsControllerClusterrolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/serviceaccount-pull-secrets-controller-clusterrolebinding.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerServicemonitorRoleYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: prometheus-k8s
  namespace: openshift-controller-manager
rules:
- apiGroups:
  - ""
  resources:
  - services
  - endpoints
  - pods
  verbs:
  - get
  - list
  - watch
`)

func v3110OpenshiftControllerManagerServicemonitorRoleYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerServicemonitorRoleYaml, nil
}

func v3110OpenshiftControllerManagerServicemonitorRoleYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerServicemonitorRoleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/servicemonitor-role.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerServicemonitorRolebindingYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: prometheus-k8s
  namespace: openshift-controller-manager
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: prometheus-k8s
subjects:
- kind: ServiceAccount
  name: prometheus-k8s
  namespace: openshift-monitoring
`)

func v3110OpenshiftControllerManagerServicemonitorRolebindingYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerServicemonitorRolebindingYaml, nil
}

func v3110OpenshiftControllerManagerServicemonitorRolebindingYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerServicemonitorRolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/servicemonitor-rolebinding.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerSvcYaml = []byte(`apiVersion: v1
kind: Service
metadata:
  namespace: openshift-controller-manager
  name: controller-manager
  annotations:
    service.alpha.openshift.io/serving-cert-secret-name: serving-cert
    prometheus.io/scrape: "true"
    prometheus.io/scheme: https
spec:
  selector:
    controller-manager: "true"
  ports:
  - name: https
    port: 443
    targetPort: 8443
`)

func v3110OpenshiftControllerManagerSvcYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerSvcYaml, nil
}

func v3110OpenshiftControllerManagerSvcYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerSvcYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/svc.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerTemplateInstanceControllerClusterroleYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: system:openshift:openshift-controller-manager:template-instance-controller
rules:
- apiGroups:
  - authorization.k8s.io
  resources:
  - subjectaccessreviews
  verbs:
  - create
- apiGroups:
  - template.openshift.io
  resources:
  - templateinstances/status
  verbs:
  - update
`)

func v3110OpenshiftControllerManagerTemplateInstanceControllerClusterroleYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerTemplateInstanceControllerClusterroleYaml, nil
}

func v3110OpenshiftControllerManagerTemplateInstanceControllerClusterroleYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerTemplateInstanceControllerClusterroleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/template-instance-controller-clusterrole.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerTemplateInstanceControllerClusterrolebindingAdminYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: system:openshift:openshift-controller-manager:template-instance-controller:admin
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: admin
subjects:
- kind: ServiceAccount
  name: template-instance-controller
  namespace: openshift-infra
`)

func v3110OpenshiftControllerManagerTemplateInstanceControllerClusterrolebindingAdminYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerTemplateInstanceControllerClusterrolebindingAdminYaml, nil
}

func v3110OpenshiftControllerManagerTemplateInstanceControllerClusterrolebindingAdminYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerTemplateInstanceControllerClusterrolebindingAdminYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/template-instance-controller-clusterrolebinding-admin.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerTemplateInstanceControllerClusterrolebindingYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: system:openshift:openshift-controller-manager:template-instance-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:openshift:openshift-controller-manager:template-instance-controller`)

func v3110OpenshiftControllerManagerTemplateInstanceControllerClusterrolebindingYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerTemplateInstanceControllerClusterrolebindingYaml, nil
}

func v3110OpenshiftControllerManagerTemplateInstanceControllerClusterrolebindingYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerTemplateInstanceControllerClusterrolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/template-instance-controller-clusterrolebinding.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerTemplateInstanceFinalizerControllerClusterroleYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: system:openshift:openshift-controller-manager:template-instance-finalizer-controller
rules:
- apiGroups:
  - template.openshift.io
  resources:
  - templateinstances/status
  verbs:
  - update
`)

func v3110OpenshiftControllerManagerTemplateInstanceFinalizerControllerClusterroleYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerTemplateInstanceFinalizerControllerClusterroleYaml, nil
}

func v3110OpenshiftControllerManagerTemplateInstanceFinalizerControllerClusterroleYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerTemplateInstanceFinalizerControllerClusterroleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/template-instance-finalizer-controller-clusterrole.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerTemplateInstanceFinalizerControllerClusterrolebindingAdminYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: system:openshift:openshift-controller-manager:template-instance-finalizer-controller:admin
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: admin
subjects:
- kind: ServiceAccount
  name: template-instance-finalizer-controller
  namespace: openshift-infra
`)

func v3110OpenshiftControllerManagerTemplateInstanceFinalizerControllerClusterrolebindingAdminYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerTemplateInstanceFinalizerControllerClusterrolebindingAdminYaml, nil
}

func v3110OpenshiftControllerManagerTemplateInstanceFinalizerControllerClusterrolebindingAdminYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerTemplateInstanceFinalizerControllerClusterrolebindingAdminYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/template-instance-finalizer-controller-clusterrolebinding-admin.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerTemplateInstanceFinalizerControllerClusterrolebindingYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: system:openshift:openshift-controller-manager:template-instance-finalizer-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:openshift:openshift-controller-manager:template-instance-finalizer-controller
subjects:
- kind: ServiceAccount
  name: template-instance-finalizer-controller
  namespace: openshift-infra
`)

func v3110OpenshiftControllerManagerTemplateInstanceFinalizerControllerClusterrolebindingYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerTemplateInstanceFinalizerControllerClusterrolebindingYaml, nil
}

func v3110OpenshiftControllerManagerTemplateInstanceFinalizerControllerClusterrolebindingYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerTemplateInstanceFinalizerControllerClusterrolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/template-instance-finalizer-controller-clusterrolebinding.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerTokenreviewClusterroleYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: system:openshift:tokenreview-openshift-controller-manager
rules:
- apiGroups:
  - authentication.k8s.io
  resources:
  - tokenreviews
  verbs:
  - create
- apiGroups:
  - authorization.k8s.io
  resources:
  - subjectaccessreviews
  verbs:
  - create
`)

func v3110OpenshiftControllerManagerTokenreviewClusterroleYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerTokenreviewClusterroleYaml, nil
}

func v3110OpenshiftControllerManagerTokenreviewClusterroleYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerTokenreviewClusterroleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/tokenreview-clusterrole.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerTokenreviewClusterrolebindingYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: system:openshift:tokenreview-openshift-controller-manager
roleRef:
  kind: ClusterRole
  name: system:openshift:tokenreview-openshift-controller-manager
subjects:
- kind: ServiceAccount
  namespace: openshift-controller-manager
  name: openshift-controller-manager-sa
`)

func v3110OpenshiftControllerManagerTokenreviewClusterrolebindingYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerTokenreviewClusterrolebindingYaml, nil
}

func v3110OpenshiftControllerManagerTokenreviewClusterrolebindingYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerTokenreviewClusterrolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/tokenreview-clusterrolebinding.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerUnidlingControllerClusterroleYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: system:openshift:openshift-controller-manager:unidling-controller
rules:
- apiGroups:
  - ""
  resources:
  - endpoints
  - replicationcontrollers/scale
  - services
  verbs:
  - get
  - update
- apiGroups:
  - ""
  resources:
  - replicationcontrollers
  verbs:
  - get
  - patch
  - update
- apiGroups:
  - ""
  - apps.openshift.io
  resources:
  - deploymentconfigs
  verbs:
  - get
  - patch
  - update
- apiGroups:
  - apps
  - extensions
  resources:
  - deployments/scale
  - replicasets/scale
  verbs:
  - get
  - update
- apiGroups:
  - ""
  - apps.openshift.io
  resources:
  - deploymentconfigs/scale
  verbs:
  - get
  - update
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - patch
  - update
`)

func v3110OpenshiftControllerManagerUnidlingControllerClusterroleYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerUnidlingControllerClusterroleYaml, nil
}

func v3110OpenshiftControllerManagerUnidlingControllerClusterroleYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerUnidlingControllerClusterroleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/unidling-controller-clusterrole.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerUnidlingControllerClusterrolebindingYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: system:openshift:openshift-controller-manager:unidling-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:openshift:openshift-controller-manager:unidling-controller
subjects:
- kind: ServiceAccount
  name: unidling-controller
  namespace: openshift-infra
`)

func v3110OpenshiftControllerManagerUnidlingControllerClusterrolebindingYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerUnidlingControllerClusterrolebindingYaml, nil
}

func v3110OpenshiftControllerManagerUnidlingControllerClusterrolebindingYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerUnidlingControllerClusterrolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/unidling-controller-clusterrolebinding.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"v3.11.0/config/defaultconfig.yaml":                                                                          v3110ConfigDefaultconfigYaml,
	"v3.11.0/openshift-controller-manager/build-config-change-controller-clusterrole.yaml":                       v3110OpenshiftControllerManagerBuildConfigChangeControllerClusterroleYaml,
	"v3.11.0/openshift-controller-manager/build-config-change-controller-clusterrolebinding.yaml":                v3110OpenshiftControllerManagerBuildConfigChangeControllerClusterrolebindingYaml,
	"v3.11.0/openshift-controller-manager/build-controller-clusterrole.yaml":                                     v3110OpenshiftControllerManagerBuildControllerClusterroleYaml,
	"v3.11.0/openshift-controller-manager/build-controller-clusterrolebinding.yaml":                              v3110OpenshiftControllerManagerBuildControllerClusterrolebindingYaml,
	"v3.11.0/openshift-controller-manager/buildconfigstatus-clusterrole.yaml":                                    v3110OpenshiftControllerManagerBuildconfigstatusClusterroleYaml,
	"v3.11.0/openshift-controller-manager/buildconfigstatus-clusterrolebinding.yaml":                             v3110OpenshiftControllerManagerBuildconfigstatusClusterrolebindingYaml,
	"v3.11.0/openshift-controller-manager/cm.yaml":                                                               v3110OpenshiftControllerManagerCmYaml,
	"v3.11.0/openshift-controller-manager/default-rolebindings-controller-clusterrole.yaml":                      v3110OpenshiftControllerManagerDefaultRolebindingsControllerClusterroleYaml,
	"v3.11.0/openshift-controller-manager/default-rolebindings-controller-clusterrolebinding-deployer.yaml":      v3110OpenshiftControllerManagerDefaultRolebindingsControllerClusterrolebindingDeployerYaml,
	"v3.11.0/openshift-controller-manager/default-rolebindings-controller-clusterrolebinding-image-builder.yaml": v3110OpenshiftControllerManagerDefaultRolebindingsControllerClusterrolebindingImageBuilderYaml,
	"v3.11.0/openshift-controller-manager/default-rolebindings-controller-clusterrolebinding-image-puller.yaml":  v3110OpenshiftControllerManagerDefaultRolebindingsControllerClusterrolebindingImagePullerYaml,
	"v3.11.0/openshift-controller-manager/default-rolebindings-controller-clusterrolebinding.yaml":               v3110OpenshiftControllerManagerDefaultRolebindingsControllerClusterrolebindingYaml,
	"v3.11.0/openshift-controller-manager/deployer-controller-clusterrole.yaml":                                  v3110OpenshiftControllerManagerDeployerControllerClusterroleYaml,
	"v3.11.0/openshift-controller-manager/deployer-controller-clusterrolebinding.yaml":                           v3110OpenshiftControllerManagerDeployerControllerClusterrolebindingYaml,
	"v3.11.0/openshift-controller-manager/deploymentconfig-controller-clusterrole.yaml":                          v3110OpenshiftControllerManagerDeploymentconfigControllerClusterroleYaml,
	"v3.11.0/openshift-controller-manager/deploymentconfig-controller-clusterrolebinding.yaml":                   v3110OpenshiftControllerManagerDeploymentconfigControllerClusterrolebindingYaml,
	"v3.11.0/openshift-controller-manager/ds.yaml":                                                               v3110OpenshiftControllerManagerDsYaml,
	"v3.11.0/openshift-controller-manager/image-import-controller-clusterrole.yaml":                              v3110OpenshiftControllerManagerImageImportControllerClusterroleYaml,
	"v3.11.0/openshift-controller-manager/image-import-controller-clusterrolebinding.yaml":                       v3110OpenshiftControllerManagerImageImportControllerClusterrolebindingYaml,
	"v3.11.0/openshift-controller-manager/image-trigger-controller-clusterrole.yaml":                             v3110OpenshiftControllerManagerImageTriggerControllerClusterroleYaml,
	"v3.11.0/openshift-controller-manager/image-trigger-controller-clusterrolebinding.yaml":                      v3110OpenshiftControllerManagerImageTriggerControllerClusterrolebindingYaml,
	"v3.11.0/openshift-controller-manager/informer-clusterrole.yaml":                                             v3110OpenshiftControllerManagerInformerClusterroleYaml,
	"v3.11.0/openshift-controller-manager/informer-clusterrolebinding.yaml":                                      v3110OpenshiftControllerManagerInformerClusterrolebindingYaml,
	"v3.11.0/openshift-controller-manager/ingress-to-route-controller-clusterrole.yaml":                          v3110OpenshiftControllerManagerIngressToRouteControllerClusterroleYaml,
	"v3.11.0/openshift-controller-manager/ingress-to-route-controller-clusterrolebinding.yaml":                   v3110OpenshiftControllerManagerIngressToRouteControllerClusterrolebindingYaml,
	"v3.11.0/openshift-controller-manager/leader-role.yaml":                                                      v3110OpenshiftControllerManagerLeaderRoleYaml,
	"v3.11.0/openshift-controller-manager/leader-rolebinding.yaml":                                               v3110OpenshiftControllerManagerLeaderRolebindingYaml,
	"v3.11.0/openshift-controller-manager/ns.yaml":                                                               v3110OpenshiftControllerManagerNsYaml,
	"v3.11.0/openshift-controller-manager/old-leader-role.yaml":                                                  v3110OpenshiftControllerManagerOldLeaderRoleYaml,
	"v3.11.0/openshift-controller-manager/old-leader-rolebinding.yaml":                                           v3110OpenshiftControllerManagerOldLeaderRolebindingYaml,
	"v3.11.0/openshift-controller-manager/openshift-global-ca-cm.yaml":                                           v3110OpenshiftControllerManagerOpenshiftGlobalCaCmYaml,
	"v3.11.0/openshift-controller-manager/openshift-service-ca-cm.yaml":                                          v3110OpenshiftControllerManagerOpenshiftServiceCaCmYaml,
	"v3.11.0/openshift-controller-manager/origin-namespace-controller-clusterrole.yaml":                          v3110OpenshiftControllerManagerOriginNamespaceControllerClusterroleYaml,
	"v3.11.0/openshift-controller-manager/origin-namespace-controller-clusterrolebinding.yaml":                   v3110OpenshiftControllerManagerOriginNamespaceControllerClusterrolebindingYaml,
	"v3.11.0/openshift-controller-manager/sa.yaml":                                                               v3110OpenshiftControllerManagerSaYaml,
	"v3.11.0/openshift-controller-manager/separate-sa-role.yaml":                                                 v3110OpenshiftControllerManagerSeparateSaRoleYaml,
	"v3.11.0/openshift-controller-manager/separate-sa-rolebinding.yaml":                                          v3110OpenshiftControllerManagerSeparateSaRolebindingYaml,
	"v3.11.0/openshift-controller-manager/service-ingress-ip-controller-clusterrole.yaml":                        v3110OpenshiftControllerManagerServiceIngressIpControllerClusterroleYaml,
	"v3.11.0/openshift-controller-manager/service-ingress-ip-controller-clusterrolebinding.yaml":                 v3110OpenshiftControllerManagerServiceIngressIpControllerClusterrolebindingYaml,
	"v3.11.0/openshift-controller-manager/serviceaccount-controller-clusterrole.yaml":                            v3110OpenshiftControllerManagerServiceaccountControllerClusterroleYaml,
	"v3.11.0/openshift-controller-manager/serviceaccount-controller-clusterrolebinding.yaml":                     v3110OpenshiftControllerManagerServiceaccountControllerClusterrolebindingYaml,
	"v3.11.0/openshift-controller-manager/serviceaccount-pull-secrets-controller-clusterrole.yaml":               v3110OpenshiftControllerManagerServiceaccountPullSecretsControllerClusterroleYaml,
	"v3.11.0/openshift-controller-manager/serviceaccount-pull-secrets-controller-clusterrolebinding.yaml":        v3110OpenshiftControllerManagerServiceaccountPullSecretsControllerClusterrolebindingYaml,
	"v3.11.0/openshift-controller-manager/servicemonitor-role.yaml":                                              v3110OpenshiftControllerManagerServicemonitorRoleYaml,
	"v3.11.0/openshift-controller-manager/servicemonitor-rolebinding.yaml":                                       v3110OpenshiftControllerManagerServicemonitorRolebindingYaml,
	"v3.11.0/openshift-controller-manager/svc.yaml":                                                              v3110OpenshiftControllerManagerSvcYaml,
	"v3.11.0/openshift-controller-manager/template-instance-controller-clusterrole.yaml":                         v3110OpenshiftControllerManagerTemplateInstanceControllerClusterroleYaml,
	"v3.11.0/openshift-controller-manager/template-instance-controller-clusterrolebinding-admin.yaml":            v3110OpenshiftControllerManagerTemplateInstanceControllerClusterrolebindingAdminYaml,
	"v3.11.0/openshift-controller-manager/template-instance-controller-clusterrolebinding.yaml":                  v3110OpenshiftControllerManagerTemplateInstanceControllerClusterrolebindingYaml,
	"v3.11.0/openshift-controller-manager/template-instance-finalizer-controller-clusterrole.yaml":               v3110OpenshiftControllerManagerTemplateInstanceFinalizerControllerClusterroleYaml,
	"v3.11.0/openshift-controller-manager/template-instance-finalizer-controller-clusterrolebinding-admin.yaml":  v3110OpenshiftControllerManagerTemplateInstanceFinalizerControllerClusterrolebindingAdminYaml,
	"v3.11.0/openshift-controller-manager/template-instance-finalizer-controller-clusterrolebinding.yaml":        v3110OpenshiftControllerManagerTemplateInstanceFinalizerControllerClusterrolebindingYaml,
	"v3.11.0/openshift-controller-manager/tokenreview-clusterrole.yaml":                                          v3110OpenshiftControllerManagerTokenreviewClusterroleYaml,
	"v3.11.0/openshift-controller-manager/tokenreview-clusterrolebinding.yaml":                                   v3110OpenshiftControllerManagerTokenreviewClusterrolebindingYaml,
	"v3.11.0/openshift-controller-manager/unidling-controller-clusterrole.yaml":                                  v3110OpenshiftControllerManagerUnidlingControllerClusterroleYaml,
	"v3.11.0/openshift-controller-manager/unidling-controller-clusterrolebinding.yaml":                           v3110OpenshiftControllerManagerUnidlingControllerClusterrolebindingYaml,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"v3.11.0": {nil, map[string]*bintree{
		"config": {nil, map[string]*bintree{
			"defaultconfig.yaml": {v3110ConfigDefaultconfigYaml, map[string]*bintree{}},
		}},
		"openshift-controller-manager": {nil, map[string]*bintree{
			"build-config-change-controller-clusterrole.yaml":        {v3110OpenshiftControllerManagerBuildConfigChangeControllerClusterroleYaml, map[string]*bintree{}},
			"build-config-change-controller-clusterrolebinding.yaml": {v3110OpenshiftControllerManagerBuildConfigChangeControllerClusterrolebindingYaml, map[string]*bintree{}},
			"build-controller-clusterrole.yaml":                      {v3110OpenshiftControllerManagerBuildControllerClusterroleYaml, map[string]*bintree{}},
			"build-controller-clusterrolebinding.yaml":               {v3110OpenshiftControllerManagerBuildControllerClusterrolebindingYaml, map[string]*bintree{}},
			"buildconfigstatus-clusterrole.yaml":                     {v3110OpenshiftControllerManagerBuildconfigstatusClusterroleYaml, map[string]*bintree{}},
			"buildconfigstatus-clusterrolebinding.yaml":              {v3110OpenshiftControllerManagerBuildconfigstatusClusterrolebindingYaml, map[string]*bintree{}},
			"cm.yaml": {v3110OpenshiftControllerManagerCmYaml, map[string]*bintree{}},
			"default-rolebindings-controller-clusterrole.yaml":                      {v3110OpenshiftControllerManagerDefaultRolebindingsControllerClusterroleYaml, map[string]*bintree{}},
			"default-rolebindings-controller-clusterrolebinding-deployer.yaml":      {v3110OpenshiftControllerManagerDefaultRolebindingsControllerClusterrolebindingDeployerYaml, map[string]*bintree{}},
			"default-rolebindings-controller-clusterrolebinding-image-builder.yaml": {v3110OpenshiftControllerManagerDefaultRolebindingsControllerClusterrolebindingImageBuilderYaml, map[string]*bintree{}},
			"default-rolebindings-controller-clusterrolebinding-image-puller.yaml":  {v3110OpenshiftControllerManagerDefaultRolebindingsControllerClusterrolebindingImagePullerYaml, map[string]*bintree{}},
			"default-rolebindings-controller-clusterrolebinding.yaml":               {v3110OpenshiftControllerManagerDefaultRolebindingsControllerClusterrolebindingYaml, map[string]*bintree{}},
			"deployer-controller-clusterrole.yaml":                                  {v3110OpenshiftControllerManagerDeployerControllerClusterroleYaml, map[string]*bintree{}},
			"deployer-controller-clusterrolebinding.yaml":                           {v3110OpenshiftControllerManagerDeployerControllerClusterrolebindingYaml, map[string]*bintree{}},
			"deploymentconfig-controller-clusterrole.yaml":                          {v3110OpenshiftControllerManagerDeploymentconfigControllerClusterroleYaml, map[string]*bintree{}},
			"deploymentconfig-controller-clusterrolebinding.yaml":                   {v3110OpenshiftControllerManagerDeploymentconfigControllerClusterrolebindingYaml, map[string]*bintree{}},
			"ds.yaml": {v3110OpenshiftControllerManagerDsYaml, map[string]*bintree{}},
			"image-import-controller-clusterrole.yaml":            {v3110OpenshiftControllerManagerImageImportControllerClusterroleYaml, map[string]*bintree{}},
			"image-import-controller-clusterrolebinding.yaml":     {v3110OpenshiftControllerManagerImageImportControllerClusterrolebindingYaml, map[string]*bintree{}},
			"image-trigger-controller-clusterrole.yaml":           {v3110OpenshiftControllerManagerImageTriggerControllerClusterroleYaml, map[string]*bintree{}},
			"image-trigger-controller-clusterrolebinding.yaml":    {v3110OpenshiftControllerManagerImageTriggerControllerClusterrolebindingYaml, map[string]*bintree{}},
			"informer-clusterrole.yaml":                           {v3110OpenshiftControllerManagerInformerClusterroleYaml, map[string]*bintree{}},
			"informer-clusterrolebinding.yaml":                    {v3110OpenshiftControllerManagerInformerClusterrolebindingYaml, map[string]*bintree{}},
			"ingress-to-route-controller-clusterrole.yaml":        {v3110OpenshiftControllerManagerIngressToRouteControllerClusterroleYaml, map[string]*bintree{}},
			"ingress-to-route-controller-clusterrolebinding.yaml": {v3110OpenshiftControllerManagerIngressToRouteControllerClusterrolebindingYaml, map[string]*bintree{}},
			"leader-role.yaml":                                    {v3110OpenshiftControllerManagerLeaderRoleYaml, map[string]*bintree{}},
			"leader-rolebinding.yaml":                             {v3110OpenshiftControllerManagerLeaderRolebindingYaml, map[string]*bintree{}},
			"ns.yaml":                                             {v3110OpenshiftControllerManagerNsYaml, map[string]*bintree{}},
			"old-leader-role.yaml":                                {v3110OpenshiftControllerManagerOldLeaderRoleYaml, map[string]*bintree{}},
			"old-leader-rolebinding.yaml":                         {v3110OpenshiftControllerManagerOldLeaderRolebindingYaml, map[string]*bintree{}},
			"openshift-global-ca-cm.yaml":                         {v3110OpenshiftControllerManagerOpenshiftGlobalCaCmYaml, map[string]*bintree{}},
			"openshift-service-ca-cm.yaml":                        {v3110OpenshiftControllerManagerOpenshiftServiceCaCmYaml, map[string]*bintree{}},
			"origin-namespace-controller-clusterrole.yaml":        {v3110OpenshiftControllerManagerOriginNamespaceControllerClusterroleYaml, map[string]*bintree{}},
			"origin-namespace-controller-clusterrolebinding.yaml": {v3110OpenshiftControllerManagerOriginNamespaceControllerClusterrolebindingYaml, map[string]*bintree{}},
			"sa.yaml":                      {v3110OpenshiftControllerManagerSaYaml, map[string]*bintree{}},
			"separate-sa-role.yaml":        {v3110OpenshiftControllerManagerSeparateSaRoleYaml, map[string]*bintree{}},
			"separate-sa-rolebinding.yaml": {v3110OpenshiftControllerManagerSeparateSaRolebindingYaml, map[string]*bintree{}},
			"service-ingress-ip-controller-clusterrole.yaml":                       {v3110OpenshiftControllerManagerServiceIngressIpControllerClusterroleYaml, map[string]*bintree{}},
			"service-ingress-ip-controller-clusterrolebinding.yaml":                {v3110OpenshiftControllerManagerServiceIngressIpControllerClusterrolebindingYaml, map[string]*bintree{}},
			"serviceaccount-controller-clusterrole.yaml":                           {v3110OpenshiftControllerManagerServiceaccountControllerClusterroleYaml, map[string]*bintree{}},
			"serviceaccount-controller-clusterrolebinding.yaml":                    {v3110OpenshiftControllerManagerServiceaccountControllerClusterrolebindingYaml, map[string]*bintree{}},
			"serviceaccount-pull-secrets-controller-clusterrole.yaml":              {v3110OpenshiftControllerManagerServiceaccountPullSecretsControllerClusterroleYaml, map[string]*bintree{}},
			"serviceaccount-pull-secrets-controller-clusterrolebinding.yaml":       {v3110OpenshiftControllerManagerServiceaccountPullSecretsControllerClusterrolebindingYaml, map[string]*bintree{}},
			"servicemonitor-role.yaml":                                             {v3110OpenshiftControllerManagerServicemonitorRoleYaml, map[string]*bintree{}},
			"servicemonitor-rolebinding.yaml":                                      {v3110OpenshiftControllerManagerServicemonitorRolebindingYaml, map[string]*bintree{}},
			"svc.yaml":                                                             {v3110OpenshiftControllerManagerSvcYaml, map[string]*bintree{}},
			"template-instance-controller-clusterrole.yaml":                        {v3110OpenshiftControllerManagerTemplateInstanceControllerClusterroleYaml, map[string]*bintree{}},
			"template-instance-controller-clusterrolebinding-admin.yaml":           {v3110OpenshiftControllerManagerTemplateInstanceControllerClusterrolebindingAdminYaml, map[string]*bintree{}},
			"template-instance-controller-clusterrolebinding.yaml":                 {v3110OpenshiftControllerManagerTemplateInstanceControllerClusterrolebindingYaml, map[string]*bintree{}},
			"template-instance-finalizer-controller-clusterrole.yaml":              {v3110OpenshiftControllerManagerTemplateInstanceFinalizerControllerClusterroleYaml, map[string]*bintree{}},
			"template-instance-finalizer-controller-clusterrolebinding-admin.yaml": {v3110OpenshiftControllerManagerTemplateInstanceFinalizerControllerClusterrolebindingAdminYaml, map[string]*bintree{}},
			"template-instance-finalizer-controller-clusterrolebinding.yaml":       {v3110OpenshiftControllerManagerTemplateInstanceFinalizerControllerClusterrolebindingYaml, map[string]*bintree{}},
			"tokenreview-clusterrole.yaml":                                         {v3110OpenshiftControllerManagerTokenreviewClusterroleYaml, map[string]*bintree{}},
			"tokenreview-clusterrolebinding.yaml":                                  {v3110OpenshiftControllerManagerTokenreviewClusterrolebindingYaml, map[string]*bintree{}},
			"unidling-controller-clusterrole.yaml":                                 {v3110OpenshiftControllerManagerUnidlingControllerClusterroleYaml, map[string]*bintree{}},
			"unidling-controller-clusterrolebinding.yaml":                          {v3110OpenshiftControllerManagerUnidlingControllerClusterrolebindingYaml, map[string]*bintree{}},
		}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
