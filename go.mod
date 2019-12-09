module github.com/openshift/cluster-openshift-controller-manager-operator

go 1.12

require (
	github.com/certifi/gocertifi v0.0.0-20180905225744-ee1a9a0726d2 // indirect
	github.com/coreos/bbolt v1.3.3 // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/getsentry/raven-go v0.0.0-20190513200303-c977f96e1095 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/gorilla/websocket v1.4.1 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.1.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.11.3 // indirect
	github.com/jteeuwen/go-bindata v0.0.0-00010101000000-000000000000
	github.com/onsi/ginkgo v1.10.2 // indirect
	github.com/openshift/api v3.9.1-0.20191111211345-a27ff30ebf09+incompatible
	github.com/openshift/client-go v0.0.0-20191022152013-2823239d2298
	github.com/openshift/library-go v0.0.0-20191029093732-d368763acf52
	github.com/pkg/profile v1.3.0 // indirect
	github.com/prometheus/client_golang v1.0.0
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/tmc/grpc-websocket-proxy v0.0.0-20190109142713-0ad062ec5ee5 // indirect
	go.uber.org/zap v1.11.0 // indirect
	k8s.io/api v0.0.0-20191122220107-b5267f2975e0
	k8s.io/apiextensions-apiserver v0.0.0-20191016113550-5357c4baaf65 // indirect
	k8s.io/apimachinery v0.0.0-20191121175448-79c2a76c473a
	k8s.io/client-go v0.0.0-20191122220542-ed16ecbdf3a0
	k8s.io/component-base v0.0.0-20191122220729-2684fb322cb9
	k8s.io/klog v1.0.0
	k8s.io/kube-aggregator v0.0.0-20191016112429-9587704a8ad4 // indirect
)

replace (
	github.com/jteeuwen/go-bindata => github.com/jteeuwen/go-bindata v0.0.0-20151023091102-a0ff2567cfb7
	k8s.io/api => k8s.io/api v0.0.0-20191122220107-b5267f2975e0
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20191122222427-64482ea217ff
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20191121175448-79c2a76c473a
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20191122221311-9d521947b1e1
	k8s.io/client-go => k8s.io/client-go v0.0.0-20191122220542-ed16ecbdf3a0
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20191121175249-e95606b614f0
	k8s.io/component-base => k8s.io/component-base v0.0.0-20191122220729-2684fb322cb9
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20191122221605-1e8d331e4dcc
)
