module inspection-server

go 1.22.2

replace (
	github.com/rancher/kubernetes-provider-detector => github.com/cnrancher/kubernetes-provider-detector v0.1.6-0.20231201070144-30b5716078ac
	k8s.io/api => k8s.io/api v0.28.6
	k8s.io/apimachinery => k8s.io/apimachinery v0.28.6
	k8s.io/client-go => k8s.io/client-go v0.28.6
)

require (
	github.com/go-rod/rod v0.116.1
	github.com/go-sql-driver/mysql v1.8.1
	github.com/google/uuid v1.6.0
	github.com/gorilla/mux v1.8.1
	github.com/larksuite/oapi-sdk-go/v3 v3.2.9
	github.com/mattn/go-sqlite3 v1.14.22
	github.com/rancher/kubernetes-provider-detector v0.1.5
	github.com/robfig/cron/v3 v3.0.1
	github.com/signintech/gopdf v0.26.1
	github.com/sirupsen/logrus v1.9.3
	github.com/urfave/cli v1.22.15
	k8s.io/api v0.30.3
	k8s.io/apimachinery v0.30.3
	k8s.io/client-go v0.30.1
	k8s.io/utils v0.0.0-20240711033017-18e509b52bc8
	sigs.k8s.io/yaml v1.4.0
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.4 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/emicklei/go-restful/v3 v3.11.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.22.4 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/gnostic-models v0.6.8 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/imdario/mergo v0.3.13 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/moby/spdystream v0.2.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/onsi/ginkgo/v2 v2.19.0 // indirect
	github.com/onsi/gomega v1.33.1 // indirect
	github.com/phpdave11/gofpdi v1.0.14-0.20211212211723-1f10f9844311 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/ysmood/fetchup v0.2.4 // indirect
	github.com/ysmood/goob v0.4.0 // indirect
	github.com/ysmood/got v0.40.0 // indirect
	github.com/ysmood/gson v0.7.3 // indirect
	github.com/ysmood/leakless v0.9.0 // indirect
	golang.org/x/net v0.27.0 // indirect
	golang.org/x/oauth2 v0.16.0 // indirect
	golang.org/x/sys v0.22.0 // indirect
	golang.org/x/term v0.22.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/kube-openapi v0.0.0-20240228011516-70dd3763d340 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.4.1 // indirect
)
