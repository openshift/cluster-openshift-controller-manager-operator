module github.com/openshift/cluster-openshift-controller-manager-operator

go 1.13

require (
	github.com/ghodss/yaml v1.0.0
	github.com/go-bindata/go-bindata v3.1.2+incompatible
	github.com/gorilla/websocket v1.4.1 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.1.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.11.3 // indirect
	github.com/openshift/api v0.0.0-20200824162408-86251274d75f
	github.com/openshift/build-machinery-go v0.0.0-20200819073603-48aa266c95f7
	github.com/openshift/client-go v0.0.0-20200722173614-5a1b0aaeff15
	github.com/openshift/library-go v0.0.0-20200731134909-dbf343342338
	github.com/prometheus/client_golang v1.7.1
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	go.uber.org/zap v1.11.0 // indirect
	k8s.io/api v0.19.0-rc.3
	k8s.io/apimachinery v0.19.0-rc.3
	k8s.io/apiserver v0.19.0-rc.3 // indirect
	k8s.io/client-go v0.19.0-rc.3
	k8s.io/component-base v0.19.0-rc.3
	k8s.io/klog v1.0.0
)
