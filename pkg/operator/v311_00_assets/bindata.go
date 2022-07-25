// Code generated for package v311_00_assets by go-bindata DO NOT EDIT. (@generated)
// sources:
// bindata/v3.11.0/config/defaultconfig.yaml
// bindata/v3.11.0/openshift-controller-manager/buildconfigstatus-clusterrole.yaml
// bindata/v3.11.0/openshift-controller-manager/buildconfigstatus-clusterrolebinding.yaml
// bindata/v3.11.0/openshift-controller-manager/cm.yaml
// bindata/v3.11.0/openshift-controller-manager/ds.yaml
// bindata/v3.11.0/openshift-controller-manager/informer-clusterrole.yaml
// bindata/v3.11.0/openshift-controller-manager/informer-clusterrolebinding.yaml
// bindata/v3.11.0/openshift-controller-manager/ingress-to-route-controller-clusterrole.yaml
// bindata/v3.11.0/openshift-controller-manager/ingress-to-route-controller-clusterrolebinding.yaml
// bindata/v3.11.0/openshift-controller-manager/leader-ingress-to-route-controller-role.yaml
// bindata/v3.11.0/openshift-controller-manager/leader-ingress-to-route-controller-rolebinding.yaml
// bindata/v3.11.0/openshift-controller-manager/leader-role.yaml
// bindata/v3.11.0/openshift-controller-manager/leader-rolebinding.yaml
// bindata/v3.11.0/openshift-controller-manager/ns.yaml
// bindata/v3.11.0/openshift-controller-manager/old-leader-role.yaml
// bindata/v3.11.0/openshift-controller-manager/old-leader-rolebinding.yaml
// bindata/v3.11.0/openshift-controller-manager/openshift-global-ca-cm.yaml
// bindata/v3.11.0/openshift-controller-manager/openshift-service-ca-cm.yaml
// bindata/v3.11.0/openshift-controller-manager/route-controller-ns.yaml
// bindata/v3.11.0/openshift-controller-manager/sa.yaml
// bindata/v3.11.0/openshift-controller-manager/separate-sa-role.yaml
// bindata/v3.11.0/openshift-controller-manager/separate-sa-rolebinding.yaml
// bindata/v3.11.0/openshift-controller-manager/servicemonitor-role.yaml
// bindata/v3.11.0/openshift-controller-manager/servicemonitor-rolebinding.yaml
// bindata/v3.11.0/openshift-controller-manager/svc.yaml
// bindata/v3.11.0/openshift-controller-manager/tokenreview-clusterrole.yaml
// bindata/v3.11.0/openshift-controller-manager/tokenreview-clusterrolebinding.yaml
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
      securityContext:
        runAsNonRoot: true
        seccompProfile:
          type: RuntimeDefault
      priorityClassName: system-node-critical 
      serviceAccountName: openshift-controller-manager-sa
      containers:
      - name: controller-manager
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            drop:
            - ALL
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

var _v3110OpenshiftControllerManagerLeaderIngressToRouteControllerRoleYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: system:openshift:openshift-controller-manager:leader-locking-ingress-to-route-controller
  namespace: openshift-route-controller-manager
rules:
- apiGroups:
  - "coordination.k8s.io"
  resources:
  - leases
  verbs:
  - get
  - create
  - update
`)

func v3110OpenshiftControllerManagerLeaderIngressToRouteControllerRoleYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerLeaderIngressToRouteControllerRoleYaml, nil
}

func v3110OpenshiftControllerManagerLeaderIngressToRouteControllerRoleYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerLeaderIngressToRouteControllerRoleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/leader-ingress-to-route-controller-role.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v3110OpenshiftControllerManagerLeaderIngressToRouteControllerRolebindingYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: system:openshift:openshift-controller-manager:leader-locking-ingress-to-route-controller
  namespace: openshift-route-controller-manager
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: system:openshift:openshift-controller-manager:leader-locking-ingress-to-route-controller
subjects:
- kind: ServiceAccount
  namespace: openshift-infra
  name: ingress-to-route-controller
`)

func v3110OpenshiftControllerManagerLeaderIngressToRouteControllerRolebindingYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerLeaderIngressToRouteControllerRolebindingYaml, nil
}

func v3110OpenshiftControllerManagerLeaderIngressToRouteControllerRolebindingYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerLeaderIngressToRouteControllerRolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/leader-ingress-to-route-controller-rolebinding.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
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

var _v3110OpenshiftControllerManagerRouteControllerNsYaml = []byte(`apiVersion: v1
kind: Namespace
metadata:
  name: openshift-route-controller-manager
  annotations:
    openshift.io/node-selector: ""
    workload.openshift.io/allowed: "management"
  labels:
    openshift.io/cluster-monitoring: "true"
`)

func v3110OpenshiftControllerManagerRouteControllerNsYamlBytes() ([]byte, error) {
	return _v3110OpenshiftControllerManagerRouteControllerNsYaml, nil
}

func v3110OpenshiftControllerManagerRouteControllerNsYaml() (*asset, error) {
	bytes, err := v3110OpenshiftControllerManagerRouteControllerNsYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v3.11.0/openshift-controller-manager/route-controller-ns.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
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
    service.beta.openshift.io/serving-cert-secret-name: serving-cert
  labels:
    prometheus: openshift-controller-manager
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
	"v3.11.0/config/defaultconfig.yaml":                                                        v3110ConfigDefaultconfigYaml,
	"v3.11.0/openshift-controller-manager/buildconfigstatus-clusterrole.yaml":                  v3110OpenshiftControllerManagerBuildconfigstatusClusterroleYaml,
	"v3.11.0/openshift-controller-manager/buildconfigstatus-clusterrolebinding.yaml":           v3110OpenshiftControllerManagerBuildconfigstatusClusterrolebindingYaml,
	"v3.11.0/openshift-controller-manager/cm.yaml":                                             v3110OpenshiftControllerManagerCmYaml,
	"v3.11.0/openshift-controller-manager/ds.yaml":                                             v3110OpenshiftControllerManagerDsYaml,
	"v3.11.0/openshift-controller-manager/informer-clusterrole.yaml":                           v3110OpenshiftControllerManagerInformerClusterroleYaml,
	"v3.11.0/openshift-controller-manager/informer-clusterrolebinding.yaml":                    v3110OpenshiftControllerManagerInformerClusterrolebindingYaml,
	"v3.11.0/openshift-controller-manager/ingress-to-route-controller-clusterrole.yaml":        v3110OpenshiftControllerManagerIngressToRouteControllerClusterroleYaml,
	"v3.11.0/openshift-controller-manager/ingress-to-route-controller-clusterrolebinding.yaml": v3110OpenshiftControllerManagerIngressToRouteControllerClusterrolebindingYaml,
	"v3.11.0/openshift-controller-manager/leader-ingress-to-route-controller-role.yaml":        v3110OpenshiftControllerManagerLeaderIngressToRouteControllerRoleYaml,
	"v3.11.0/openshift-controller-manager/leader-ingress-to-route-controller-rolebinding.yaml": v3110OpenshiftControllerManagerLeaderIngressToRouteControllerRolebindingYaml,
	"v3.11.0/openshift-controller-manager/leader-role.yaml":                                    v3110OpenshiftControllerManagerLeaderRoleYaml,
	"v3.11.0/openshift-controller-manager/leader-rolebinding.yaml":                             v3110OpenshiftControllerManagerLeaderRolebindingYaml,
	"v3.11.0/openshift-controller-manager/ns.yaml":                                             v3110OpenshiftControllerManagerNsYaml,
	"v3.11.0/openshift-controller-manager/old-leader-role.yaml":                                v3110OpenshiftControllerManagerOldLeaderRoleYaml,
	"v3.11.0/openshift-controller-manager/old-leader-rolebinding.yaml":                         v3110OpenshiftControllerManagerOldLeaderRolebindingYaml,
	"v3.11.0/openshift-controller-manager/openshift-global-ca-cm.yaml":                         v3110OpenshiftControllerManagerOpenshiftGlobalCaCmYaml,
	"v3.11.0/openshift-controller-manager/openshift-service-ca-cm.yaml":                        v3110OpenshiftControllerManagerOpenshiftServiceCaCmYaml,
	"v3.11.0/openshift-controller-manager/route-controller-ns.yaml":                            v3110OpenshiftControllerManagerRouteControllerNsYaml,
	"v3.11.0/openshift-controller-manager/sa.yaml":                                             v3110OpenshiftControllerManagerSaYaml,
	"v3.11.0/openshift-controller-manager/separate-sa-role.yaml":                               v3110OpenshiftControllerManagerSeparateSaRoleYaml,
	"v3.11.0/openshift-controller-manager/separate-sa-rolebinding.yaml":                        v3110OpenshiftControllerManagerSeparateSaRolebindingYaml,
	"v3.11.0/openshift-controller-manager/servicemonitor-role.yaml":                            v3110OpenshiftControllerManagerServicemonitorRoleYaml,
	"v3.11.0/openshift-controller-manager/servicemonitor-rolebinding.yaml":                     v3110OpenshiftControllerManagerServicemonitorRolebindingYaml,
	"v3.11.0/openshift-controller-manager/svc.yaml":                                            v3110OpenshiftControllerManagerSvcYaml,
	"v3.11.0/openshift-controller-manager/tokenreview-clusterrole.yaml":                        v3110OpenshiftControllerManagerTokenreviewClusterroleYaml,
	"v3.11.0/openshift-controller-manager/tokenreview-clusterrolebinding.yaml":                 v3110OpenshiftControllerManagerTokenreviewClusterrolebindingYaml,
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
			"buildconfigstatus-clusterrole.yaml":        {v3110OpenshiftControllerManagerBuildconfigstatusClusterroleYaml, map[string]*bintree{}},
			"buildconfigstatus-clusterrolebinding.yaml": {v3110OpenshiftControllerManagerBuildconfigstatusClusterrolebindingYaml, map[string]*bintree{}},
			"cm.yaml":                          {v3110OpenshiftControllerManagerCmYaml, map[string]*bintree{}},
			"ds.yaml":                          {v3110OpenshiftControllerManagerDsYaml, map[string]*bintree{}},
			"informer-clusterrole.yaml":        {v3110OpenshiftControllerManagerInformerClusterroleYaml, map[string]*bintree{}},
			"informer-clusterrolebinding.yaml": {v3110OpenshiftControllerManagerInformerClusterrolebindingYaml, map[string]*bintree{}},
			"ingress-to-route-controller-clusterrole.yaml":        {v3110OpenshiftControllerManagerIngressToRouteControllerClusterroleYaml, map[string]*bintree{}},
			"ingress-to-route-controller-clusterrolebinding.yaml": {v3110OpenshiftControllerManagerIngressToRouteControllerClusterrolebindingYaml, map[string]*bintree{}},
			"leader-ingress-to-route-controller-role.yaml":        {v3110OpenshiftControllerManagerLeaderIngressToRouteControllerRoleYaml, map[string]*bintree{}},
			"leader-ingress-to-route-controller-rolebinding.yaml": {v3110OpenshiftControllerManagerLeaderIngressToRouteControllerRolebindingYaml, map[string]*bintree{}},
			"leader-role.yaml":                    {v3110OpenshiftControllerManagerLeaderRoleYaml, map[string]*bintree{}},
			"leader-rolebinding.yaml":             {v3110OpenshiftControllerManagerLeaderRolebindingYaml, map[string]*bintree{}},
			"ns.yaml":                             {v3110OpenshiftControllerManagerNsYaml, map[string]*bintree{}},
			"old-leader-role.yaml":                {v3110OpenshiftControllerManagerOldLeaderRoleYaml, map[string]*bintree{}},
			"old-leader-rolebinding.yaml":         {v3110OpenshiftControllerManagerOldLeaderRolebindingYaml, map[string]*bintree{}},
			"openshift-global-ca-cm.yaml":         {v3110OpenshiftControllerManagerOpenshiftGlobalCaCmYaml, map[string]*bintree{}},
			"openshift-service-ca-cm.yaml":        {v3110OpenshiftControllerManagerOpenshiftServiceCaCmYaml, map[string]*bintree{}},
			"route-controller-ns.yaml":            {v3110OpenshiftControllerManagerRouteControllerNsYaml, map[string]*bintree{}},
			"sa.yaml":                             {v3110OpenshiftControllerManagerSaYaml, map[string]*bintree{}},
			"separate-sa-role.yaml":               {v3110OpenshiftControllerManagerSeparateSaRoleYaml, map[string]*bintree{}},
			"separate-sa-rolebinding.yaml":        {v3110OpenshiftControllerManagerSeparateSaRolebindingYaml, map[string]*bintree{}},
			"servicemonitor-role.yaml":            {v3110OpenshiftControllerManagerServicemonitorRoleYaml, map[string]*bintree{}},
			"servicemonitor-rolebinding.yaml":     {v3110OpenshiftControllerManagerServicemonitorRolebindingYaml, map[string]*bintree{}},
			"svc.yaml":                            {v3110OpenshiftControllerManagerSvcYaml, map[string]*bintree{}},
			"tokenreview-clusterrole.yaml":        {v3110OpenshiftControllerManagerTokenreviewClusterroleYaml, map[string]*bintree{}},
			"tokenreview-clusterrolebinding.yaml": {v3110OpenshiftControllerManagerTokenreviewClusterrolebindingYaml, map[string]*bintree{}},
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
