module github.com/openshift/cluster-openshift-controller-manager-operator

go 1.13

require (
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/go-bindata/go-bindata v3.1.2+incompatible
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.1.0 // indirect
	github.com/openshift/api v0.0.0-20210730095913-85e1d547cdee
	github.com/openshift/build-machinery-go v0.0.0-20210712174854-1bb7fd1518d3
	github.com/openshift/client-go v0.0.0-20210730113412-1811c1b3fc0e
	github.com/openshift/library-go v0.0.0-20210804131825-452f351d83b0
	github.com/prometheus/client_golang v1.11.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.22.0-rc.0
	k8s.io/apimachinery v0.22.0-rc.0
	k8s.io/client-go v0.22.0-rc.0
	k8s.io/component-base v0.22.0-rc.0
	k8s.io/klog/v2 v2.9.0
	k8s.io/kube-openapi v0.0.0-20210421082810-95288971da7e // indirect
)

replace (
	// This replace directive comes from: https://github.com/etcd-io/etcd/issues/12569
	github.com/coreos/etcd => github.com/coreos/etcd v3.3.13+incompatible
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.1
	go.etcd.io/bbolt => go.etcd.io/bbolt v1.3.5
	go.etcd.io/etcd => go.etcd.io/etcd v0.5.0-alpha.5.0.20200910180754-dd1b699fc489 // ae9734ed278b is the SHA for git tag v3.4.13
	google.golang.org/grpc => google.golang.org/grpc v1.29.1
	k8s.io/apiserver => github.com/openshift/kubernetes-apiserver v0.0.0-20210419140141-620426e63a99
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20210421082810-95288971da7e
	vbom.ml/util => github.com/fvbommel/util v0.0.0-20180919145318-efcd4e0f9787
)
