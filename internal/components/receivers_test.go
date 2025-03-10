// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Skip tests on Windows temporarily, see https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/11451
//go:build !windows
// +build !windows

package components

import (
	"context"
	"errors"
	"path/filepath"
	"runtime"
	"testing"

	promconfig "github.com/prometheus/prometheus/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer/consumertest"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/adapter"
	tcpop "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/input/tcp"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/carbonreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/chronyreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filelogreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/otlpjsonfilereceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/syslogreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/tcplogreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/udplogreceiver"
)

func TestDefaultReceivers(t *testing.T) {
	allFactories, err := Components()
	assert.NoError(t, err)

	rcvrFactories := allFactories.Receivers

	tests := []struct {
		receiver     config.Type
		skipLifecyle bool
		getConfigFn  getReceiverConfigFn
	}{
		{
			receiver:     "active_directory_ds",
			skipLifecyle: true, // Requires a running windows service
		},
		{
			receiver: "aerospike",
		},
		{
			receiver: "apache",
		},
		{
			receiver: "awscontainerinsightreceiver",
			// TODO: skipped since it will only function in a container environment with procfs in expected location.
			skipLifecyle: true,
		},
		{
			receiver:     "awsecscontainermetrics",
			skipLifecyle: true, // Requires container metaendpoint to be running
		},
		{
			receiver: "awsfirehose",
		},
		{
			receiver:     "awsxray",
			skipLifecyle: true, // Requires AWS endpoint to check identity to run
		},
		{
			receiver: "azureeventhub",
		},
		{
			receiver: "bigip",
		},
		{
			receiver: "carbon",
			getConfigFn: func() config.Receiver {
				cfg := rcvrFactories["carbon"].CreateDefaultConfig().(*carbonreceiver.Config)
				cfg.Endpoint = "0.0.0.0:0"
				return cfg
			},
			skipLifecyle: true, // Panics after test have completed, requires a wait group
		},
		{
			receiver:     "cloudfoundry",
			skipLifecyle: true, // Requires UAA (auth) endpoint to run
		},
		{
			receiver: "chrony",
			getConfigFn: func() config.Receiver {
				cfg := rcvrFactories["chrony"].CreateDefaultConfig().(*chronyreceiver.Config)
				cfg.Endpoint = "udp://localhost:323"
				return cfg
			},
		},
		{
			receiver: "collectd",
		},
		{
			receiver: "couchdb",
		},
		{
			receiver:     "docker_stats",
			skipLifecyle: true,
		},
		{
			receiver:     "dotnet_diagnostics",
			skipLifecyle: true, // Requires a running .NET process to examine
		},
		{
			receiver: "elasticsearch",
		},
		{
			receiver: "expvar",
		},
		{
			receiver: "filelog",
			getConfigFn: func() config.Receiver {
				cfg := rcvrFactories["filelog"].CreateDefaultConfig().(*filelogreceiver.FileLogConfig)
				cfg.Include = []string{filepath.Join(t.TempDir(), "*")}
				return cfg
			},
		},
		{
			receiver: "flinkmetrics",
		},
		{
			receiver: "fluentforward",
		},
		{
			receiver: "googlecloudspanner",
		},
		{
			receiver:     "googlecloudpubsub",
			skipLifecyle: true, // Requires a pubsub subscription
		},
		{
			receiver: "hostmetrics",
		},
		{
			receiver: "influxdb",
		},
		{
			receiver:     "iis",
			skipLifecyle: true, // Requires a running windows process
		},
		{
			receiver: "jaeger",
		},
		{
			receiver:     "jmx",
			skipLifecyle: true, // Requires a running instance with JMX
		},
		{
			receiver:     "journald",
			skipLifecyle: runtime.GOOS != "linux",
		},
		{
			receiver:     "k8s_events",
			skipLifecyle: true, // need a valid Kubernetes host and port
		},
		{
			receiver:     "kafka",
			skipLifecyle: true, // TODO: It needs access to internals to successful start.
		},
		{
			receiver: "kafkametrics",
		},
		{
			receiver:     "k8s_cluster",
			skipLifecyle: true, // Requires access to the k8s host and port in order to run
		},
		{
			receiver:     "kubeletstats",
			skipLifecyle: true, // Requires access to certificates to auth against kubelet
		},
		{
			receiver: "memcached",
		},
		{
			receiver: "mongodb",
		},
		{
			receiver: "mongodbatlas",
		},
		{
			receiver: "mysql",
		},
		{
			receiver: "nginx",
		},
		{
			receiver: "nsxt",
		},
		{
			receiver:     "opencensus",
			skipLifecyle: true, // TODO: Usage of CMux doesn't allow proper shutdown.
		},
		{
			receiver: "otlp",
		},
		{
			receiver: "otlpjsonfile",
			getConfigFn: func() config.Receiver {
				cfg := rcvrFactories["otlpjsonfile"].CreateDefaultConfig().(*otlpjsonfilereceiver.Config)
				cfg.Include = []string{"/tmp/*.log"}
				return cfg
			},
		},
		{
			receiver:     "podman_stats",
			skipLifecyle: true, // Requires a running podman daemon
		},
		{
			receiver: "postgresql",
		},
		{
			receiver: "prometheus",
			getConfigFn: func() config.Receiver {
				cfg := rcvrFactories["prometheus"].CreateDefaultConfig().(*prometheusreceiver.Config)
				cfg.PrometheusConfig = &promconfig.Config{
					ScrapeConfigs: []*promconfig.ScrapeConfig{
						{JobName: "test"},
					},
				}
				return cfg
			},
		},
		{
			receiver:     "prometheus_exec",
			skipLifecyle: true, // Requires running a subproccess that can not be easily set across platforms
		},
		{
			receiver: "rabbitmq",
		},
		{
			receiver: "receiver_creator",
		},
		{
			receiver: "redis",
		},
		{
			receiver: "riak",
		},
		{
			receiver: "sapm",
		},
		{
			receiver: "saphana",
		},
		{
			receiver: "signalfx",
		},
		{
			receiver: "prometheus_simple",
		},
		{
			receiver: "skywalking",
		},
		{
			receiver: "splunk_hec",
		},
		{
			receiver: "sqlquery",
		},
		{
			receiver:     "sqlserver",
			skipLifecyle: true, // Requires a running windows process
		},
		{
			receiver: "statsd",
		},
		{
			receiver:     "wavefront",
			skipLifecyle: true, // Depends on carbon receiver to be running correctly
		},
		{
			receiver:     "windowseventlog",
			skipLifecyle: true, // Requires a running windows process
		},
		{
			receiver:     "windowsperfcounters",
			skipLifecyle: true, // Requires a running windows process
		},
		{
			receiver: "zipkin",
		},
		{
			receiver: "zookeeper",
		},
		{
			receiver: "syslog",
			getConfigFn: func() config.Receiver {
				cfg := rcvrFactories["syslog"].CreateDefaultConfig().(*syslogreceiver.SysLogConfig)
				cfg.TCP = &tcpop.NewConfig("tcp_input").BaseConfig
				cfg.TCP.ListenAddress = "0.0.0.0:0"
				cfg.Protocol = "rfc5424"
				return cfg
			},
		},
		{
			receiver: "tcplog",
			getConfigFn: func() config.Receiver {
				cfg := rcvrFactories["tcplog"].CreateDefaultConfig().(*tcplogreceiver.TCPLogConfig)
				cfg.Input = adapter.InputConfig{
					"listen_address": "0.0.0.0:0",
				}
				return cfg
			},
		},
		{
			receiver: "udplog",
			getConfigFn: func() config.Receiver {
				cfg := rcvrFactories["udplog"].CreateDefaultConfig().(*udplogreceiver.UDPLogConfig)
				cfg.Input = adapter.InputConfig{
					"listen_address": "0.0.0.0:0",
				}
				return cfg
			},
		},
		{
			receiver: "vcenter",
		},
		{
			receiver:     "solace",
			skipLifecyle: true, // Requires a solace broker to connect to
		},
	}

	assert.Len(t, tests, len(rcvrFactories), "All receivers must be added to the lifecycle suite")
	for _, tt := range tests {
		t.Run(string(tt.receiver), func(t *testing.T) {
			factory, ok := rcvrFactories[tt.receiver]
			require.True(t, ok)
			assert.Equal(t, tt.receiver, factory.Type())
			assert.Equal(t, config.NewComponentID(tt.receiver), factory.CreateDefaultConfig().ID())

			if tt.skipLifecyle {
				t.Skip("Skipping lifecycle test", tt.receiver)
				return
			}

			verifyReceiverLifecycle(t, factory, tt.getConfigFn)
		})
	}
}

// getReceiverConfigFn is used customize the configuration passed to the verification.
// This is used to change ports or provide values required but not provided by the
// default configuration.
type getReceiverConfigFn func() config.Receiver

// verifyReceiverLifecycle is used to test if a receiver type can handle the typical
// lifecycle of a component. The getConfigFn parameter only need to be specified if
// the test can't be done with the default configuration for the component.
func verifyReceiverLifecycle(t *testing.T, factory component.ReceiverFactory, getConfigFn getReceiverConfigFn) {
	ctx := context.Background()
	host := newAssertNoErrorHost(t)
	receiverCreateSet := componenttest.NewNopReceiverCreateSettings()

	if getConfigFn == nil {
		getConfigFn = factory.CreateDefaultConfig
	}

	createFns := []createReceiverFn{
		wrapCreateLogsRcvr(factory),
		wrapCreateTracesRcvr(factory),
		wrapCreateMetricsRcvr(factory),
	}

	for _, createFn := range createFns {
		firstRcvr, err := createFn(ctx, receiverCreateSet, getConfigFn())
		if errors.Is(err, component.ErrDataTypeIsNotSupported) {
			continue
		}
		require.NoError(t, err)
		require.NoError(t, firstRcvr.Start(ctx, host))
		require.NoError(t, firstRcvr.Shutdown(ctx))

		secondRcvr, err := createFn(ctx, receiverCreateSet, getConfigFn())
		require.NoError(t, err)
		require.NoError(t, secondRcvr.Start(ctx, host))
		require.NoError(t, secondRcvr.Shutdown(ctx))
	}
}

// assertNoErrorHost implements a component.Host that asserts that there were no errors.
type createReceiverFn func(
	ctx context.Context,
	set component.ReceiverCreateSettings,
	cfg config.Receiver,
) (component.Receiver, error)

func wrapCreateLogsRcvr(factory component.ReceiverFactory) createReceiverFn {
	return func(ctx context.Context, set component.ReceiverCreateSettings, cfg config.Receiver) (component.Receiver, error) {
		return factory.CreateLogsReceiver(ctx, set, cfg, consumertest.NewNop())
	}
}

func wrapCreateMetricsRcvr(factory component.ReceiverFactory) createReceiverFn {
	return func(ctx context.Context, set component.ReceiverCreateSettings, cfg config.Receiver) (component.Receiver, error) {
		return factory.CreateMetricsReceiver(ctx, set, cfg, consumertest.NewNop())
	}
}

func wrapCreateTracesRcvr(factory component.ReceiverFactory) createReceiverFn {
	return func(ctx context.Context, set component.ReceiverCreateSettings, cfg config.Receiver) (component.Receiver, error) {
		return factory.CreateTracesReceiver(ctx, set, cfg, consumertest.NewNop())
	}
}
