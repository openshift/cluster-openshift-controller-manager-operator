module github.com/openshift/cluster-openshift-controller-manager-operator

go 1.13

require (
	github.com/ghodss/yaml v1.0.0
	github.com/go-bindata/go-bindata v3.1.2+incompatible
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/openshift/api v0.0.0-20220531073726-6c4f186339a7
	github.com/openshift/build-machinery-go v0.0.0-20220429084610-baff9f8d23b3
	github.com/openshift/client-go v0.0.0-20220603133046-984ee5ebedcf
	github.com/openshift/library-go v0.0.0-20220525173854-9b950a41acdc
	github.com/prometheus/client_golang v1.12.1
	github.com/spf13/cobra v1.4.0
	k8s.io/api v0.24.0
	k8s.io/apimachinery v0.24.0
	k8s.io/client-go v0.24.0
	k8s.io/component-base v0.24.0
	k8s.io/klog/v2 v2.60.1
)

replace vbom.ml/util => github.com/fvbommel/util v0.0.0-20180919145318-efcd4e0f9787
