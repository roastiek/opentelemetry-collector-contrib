// Copyright 2021 OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package syslogreceiver

import (
	"context"
	"fmt"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/service/servicetest"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/adapter"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/input/syslog"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/input/tcp"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/input/udp"
)

func TestSyslogWithTcp(t *testing.T) {
	testSyslog(t, testdataConfigYaml())
}

func TestSyslogWithUdp(t *testing.T) {
	testSyslog(t, testdataUDPConfig())
}

func testSyslog(t *testing.T, cfg *SysLogConfig) {
	numLogs := 5

	f := NewFactory()
	sink := new(consumertest.LogsSink)
	rcvr, err := f.CreateLogsReceiver(context.Background(), componenttest.NewNopReceiverCreateSettings(), cfg, sink)
	require.NoError(t, err)
	require.NoError(t, rcvr.Start(context.Background(), componenttest.NewNopHost()))

	var conn net.Conn
	if cfg.Config.TCP != nil {
		conn, err = net.Dial("tcp", "0.0.0.0:29018")
		require.NoError(t, err)
	} else {
		conn, err = net.Dial("udp", "0.0.0.0:29018")
		require.NoError(t, err)
	}

	for i := 0; i < numLogs; i++ {
		msg := fmt.Sprintf("<86>1 2021-02-28T00:0%d:02.003Z 192.168.1.1 SecureAuth0 23108 ID52020 [SecureAuth@27389] test msg %d\n", i, i)
		_, err = conn.Write([]byte(msg))
		require.NoError(t, err)
	}
	require.NoError(t, conn.Close())

	require.Eventually(t, expectNLogs(sink, numLogs), 2*time.Second, time.Millisecond)
	require.NoError(t, rcvr.Shutdown(context.Background()))
	require.Len(t, sink.AllLogs(), 1)

	resourceLogs := sink.AllLogs()[0].ResourceLogs().At(0)
	logs := resourceLogs.ScopeLogs().At(0).LogRecords()

	for i := 0; i < numLogs; i++ {
		log := logs.At(i)

		require.Equal(t, log.Timestamp(), pcommon.Timestamp(1614470402003000000+i*60*1000*1000*1000))
		msg, ok := log.Attributes().AsRaw()["message"]
		require.True(t, ok)
		require.Equal(t, msg, fmt.Sprintf("test msg %d", i))
	}
}

func TestLoadConfig(t *testing.T) {
	factories, err := componenttest.NopFactories()
	assert.Nil(t, err)

	factory := NewFactory()
	factories.Receivers[typeStr] = factory
	cfg, err := servicetest.LoadConfigAndValidate(filepath.Join("testdata", "config.yaml"), factories)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, len(cfg.Receivers), 1)
	assert.Equal(t, testdataConfigYaml(), cfg.Receivers[config.NewComponentID(typeStr)])
}

func testdataConfigYaml() *SysLogConfig {
	return &SysLogConfig{
		BaseConfig: adapter.BaseConfig{
			ReceiverSettings: config.NewReceiverSettings(config.NewComponentID(typeStr)),
			Operators:        adapter.OperatorConfigs{},
			Converter: adapter.ConverterConfig{
				FlushInterval: 100 * time.Millisecond,
				WorkerCount:   1,
			},
		},
		Config: func() syslog.Config {
			c := syslog.NewConfig("syslog_input")
			c.TCP = &tcp.NewConfig("tcp_input").BaseConfig
			c.TCP.ListenAddress = "0.0.0.0:29018"
			c.Protocol = "rfc5424"
			return *c
		}(),
	}
}

func testdataUDPConfig() *SysLogConfig {
	return &SysLogConfig{
		BaseConfig: adapter.BaseConfig{
			ReceiverSettings: config.NewReceiverSettings(config.NewComponentID(typeStr)),
			Operators:        adapter.OperatorConfigs{},
			Converter: adapter.ConverterConfig{
				FlushInterval: 100 * time.Millisecond,
				WorkerCount:   1,
			},
		},
		Config: func() syslog.Config {
			c := syslog.NewConfig("syslog_input")
			c.UDP = &udp.NewConfig("udp_input").BaseConfig
			c.UDP.ListenAddress = "0.0.0.0:29018"
			c.Protocol = "rfc5424"
			return *c
		}(),
	}
}

func TestDecodeInputConfigFailure(t *testing.T) {
	sink := new(consumertest.LogsSink)
	factory := NewFactory()
	badCfg := &SysLogConfig{
		BaseConfig: adapter.BaseConfig{
			ReceiverSettings: config.NewReceiverSettings(config.NewComponentID(typeStr)),
			Operators:        adapter.OperatorConfigs{},
		},
		Config: func() syslog.Config {
			c := syslog.NewConfig("syslog_input")
			c.TCP = &tcp.NewConfig("tcp_input").BaseConfig
			c.Protocol = "fake"
			return *c
		}(),
	}
	receiver, err := factory.CreateLogsReceiver(context.Background(), componenttest.NewNopReceiverCreateSettings(), badCfg, sink)
	require.Error(t, err, "receiver creation should fail if input config isn't valid")
	require.Nil(t, receiver, "receiver creation should fail if input config isn't valid")
}

func expectNLogs(sink *consumertest.LogsSink, expected int) func() bool {
	return func() bool {
		return sink.LogRecordCount() == expected
	}
}
