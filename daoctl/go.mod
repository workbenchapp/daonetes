module github.com/workbenchapp/worknet/daoctl

go 1.18

// 	github.com/pion/ice/v2 v2.2.11-0.20221008025019-af9281dc76df
// due to https://github.com/pion/ice/pull/477 on windows

require (
	github.com/alecthomas/kong v0.6.1
	github.com/alecthomas/kong-yaml v0.1.1
	github.com/davecgh/go-spew v1.1.1
	github.com/docker/docker v20.10.20+incompatible
	github.com/gagliardetto/binary v0.6.1
	github.com/gagliardetto/gofuzz v1.2.2
	github.com/gagliardetto/solana-go v1.5.0
	github.com/gagliardetto/treeout v0.1.4
	github.com/go-logr/logr v1.2.3
	github.com/go-logr/zapr v1.2.3
	github.com/google/gops v0.3.25
	github.com/grandcat/zeroconf v1.0.0
	github.com/honeycombio/honeycomb-opentelemetry-go v0.2.0
	github.com/honeycombio/opentelemetry-go-contrib/launcher v0.0.0-20220824095536-e0b3dd3fbfe7
	github.com/jpillora/backoff v1.0.0
	github.com/kardianos/service v1.2.1
	github.com/miekg/dns v1.1.27
	github.com/mitchellh/mapstructure v1.5.0
	github.com/mr-tron/base58 v1.2.0
	github.com/northbright/iocopy v1.0.0
	github.com/pion/ice/v2 v2.2.11-0.20221008025019-af9281dc76df
	github.com/portto/solana-go-sdk v1.19.1
	github.com/rs/cors v1.8.2
	github.com/stretchr/testify v1.8.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.36.1
	go.opentelemetry.io/otel v1.10.0
	go.opentelemetry.io/otel/trace v1.10.0
	go.uber.org/zap v1.22.0
	golang.org/x/sys v0.0.0-20220728004956-3c1f35247d10
	golang.zx2c4.com/wireguard v0.0.0-20220407013110-ef5c587f782d
	golang.zx2c4.com/wireguard/tun/netstack v0.0.0-20220703234212-c31a7b1ab478
	golang.zx2c4.com/wireguard/wgctrl v0.0.0-20220504211119-3d4a969bb56b
	gopkg.in/yaml.v2 v2.4.0
)

require (
	contrib.go.opencensus.io/exporter/stackdriver v0.13.4 // indirect
	filippo.io/edwards25519 v1.0.0-rc.1 // indirect
	github.com/Microsoft/go-winio v0.5.0 // indirect
	github.com/andres-erbsen/clock v0.0.0-20160526145045-9e14626cd129 // indirect
	github.com/aybabtme/rgbterm v0.0.0-20170906152045-cc83f3b3ce59 // indirect
	github.com/blendle/zapdriver v1.3.1 // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cenkalti/backoff/v4 v4.1.3 // indirect
	github.com/dfuse-io/logging v0.0.0-20201110202154-26697de88c79 // indirect
	github.com/docker/distribution v2.7.1-0.20190205005809-0d3efadf0154+incompatible // indirect
	github.com/docker/go-connections v0.3.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/fatih/color v1.9.0 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/btree v1.0.1 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/gorilla/rpc v1.2.0 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.7.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.13.6 // indirect
	github.com/logrusorgru/aurora v2.0.3+incompatible // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/mattn/go-isatty v0.0.11 // indirect
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/mostynb/zstdpool-freelist v0.0.0-20201229113212-927304c0c3b1 // indirect
	github.com/near/borsh-go v0.3.2-0.20220516180422-1ff87d108454 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/pion/dtls/v2 v2.1.5 // indirect
	github.com/pion/logging v0.2.2 // indirect
	github.com/pion/mdns v0.0.5 // indirect
	github.com/pion/randutil v0.1.0 // indirect
	github.com/pion/stun v0.3.5 // indirect
	github.com/pion/transport v0.13.1 // indirect
	github.com/pion/turn/v2 v2.0.8 // indirect
	github.com/pion/udp v0.1.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/sethvargo/go-envconfig v0.6.2 // indirect
	github.com/shirou/gopsutil/v3 v3.22.6 // indirect
	github.com/sirupsen/logrus v1.7.0 // indirect
	github.com/streamingfast/logging v0.0.0-20220405224725-2755dab2ce75 // indirect
	github.com/teris-io/shortid v0.0.0-20201117134242-e59966efd125 // indirect
	github.com/tidwall/gjson v1.9.3 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/tklauser/go-sysconf v0.3.10 // indirect
	github.com/tklauser/numcpus v0.4.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	go.opencensus.io v0.23.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/host v0.33.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/runtime v0.33.0 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.8.0 // indirect
	go.opentelemetry.io/contrib/propagators/ot v1.8.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/internal/retry v1.9.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric v0.31.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.31.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v0.31.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.9.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.8.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.9.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.9.0 // indirect
	go.opentelemetry.io/otel/metric v0.32.1 // indirect
	go.opentelemetry.io/otel/sdk v1.9.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v0.31.0 // indirect
	go.opentelemetry.io/proto/otlp v0.18.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	go.uber.org/ratelimit v0.2.0 // indirect
	golang.org/x/crypto v0.0.0-20220427172511-eb4f295cb31f // indirect
	golang.org/x/net v0.0.0-20221002022538-bcab6841153b // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	golang.zx2c4.com/wintun v0.0.0-20211104114900-415007cec224 // indirect
	google.golang.org/genproto v0.0.0-20220112215332-a9c7c0acf9f2 // indirect
	google.golang.org/grpc v1.47.0 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gvisor.dev/gvisor v0.0.0-20211020211948-f76a604701b6 // indirect
)
