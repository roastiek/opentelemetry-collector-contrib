module github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/signalfx

go 1.18

require (
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/common v0.58.0
	github.com/signalfx/com_signalfx_metrics_protobuf v0.0.3
	github.com/stretchr/testify v1.8.0
	go.opentelemetry.io/collector/pdata v0.58.0
	go.uber.org/multierr v1.8.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/kr/pretty v0.2.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	golang.org/x/net v0.0.0-20201021035429-f5854403a974 // indirect
	golang.org/x/sys v0.0.0-20210119212857-b64e53b001e4 // indirect
	golang.org/x/text v0.3.3 // indirect
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013 // indirect
	google.golang.org/grpc v1.48.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/open-telemetry/opentelemetry-collector-contrib/internal/common => ../../../internal/common
