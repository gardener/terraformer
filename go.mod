module github.com/gardener/terraformer

go 1.16

require (
	github.com/aws/aws-sdk-go v1.35.26
	github.com/fsnotify/fsnotify v1.4.9
	github.com/gardener/gardener v1.17.1-0.20210222145700-bc28019dd75a
	github.com/go-logr/logr v0.3.0
	github.com/golang/mock v1.4.4-0.20200731163441-8734ec565a4d
	github.com/hashicorp/go-multierror v1.0.0
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.5
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	go.uber.org/zap v1.15.0
	k8s.io/api v0.19.6
	k8s.io/apimachinery v0.19.6
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/utils v0.0.0-20200912215256-4140de9c8800
	sigs.k8s.io/controller-runtime v0.7.1
)

replace (
	k8s.io/api => k8s.io/api v0.19.6
	k8s.io/apimachinery => k8s.io/apimachinery v0.19.6
	k8s.io/client-go => k8s.io/client-go v0.19.6
)
