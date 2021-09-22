module github.com/openshift/cluster-openshift-controller-manager-operator

go 1.13

require (
	github.com/ghodss/yaml v1.0.0
	github.com/go-bindata/go-bindata v3.1.2+incompatible
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/openshift/api v0.0.0-20210927171657-636513e97fda
	github.com/openshift/build-machinery-go v0.0.0-20210806203541-4ea9b6da3a37
	github.com/openshift/client-go v0.0.0-20210916133943-9acee1a0fb83
	github.com/openshift/library-go v0.0.0-20210929123655-f91754363286
	github.com/prometheus/client_golang v1.11.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.22.1
	k8s.io/apimachinery v0.22.1
	k8s.io/client-go v0.22.1
	k8s.io/component-base v0.22.1
	k8s.io/klog/v2 v2.9.0
)

replace vbom.ml/util => github.com/fvbommel/util v0.0.0-20180919145318-efcd4e0f9787
