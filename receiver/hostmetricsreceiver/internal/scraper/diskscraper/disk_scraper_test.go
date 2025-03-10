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

package diskscraper

import (
	"context"
	"errors"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/processor/filterset"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/diskscraper/internal/metadata"
)

func TestScrape(t *testing.T) {
	type testCase struct {
		name              string
		config            Config
		bootTimeFunc      func() (uint64, error)
		newErrRegex       string
		initializationErr string
		expectMetrics     int
		expectedStartTime pcommon.Timestamp
		mutateScraper     func(*scraper)
	}

	metricsWithDirection := 3
	if runtime.GOOS == "linux" {
		metricsWithDirection++
	}

	testCases := []testCase{
		{
			name:          "Standard",
			config:        Config{Metrics: metadata.DefaultMetricsSettings()},
			expectMetrics: metricsLen,
		},
		{
			name:          "With direction removed",
			config:        Config{Metrics: metadata.DefaultMetricsSettings()},
			expectMetrics: metricsLen + metricsWithDirection,
			mutateScraper: func(s *scraper) {
				s.emitMetricsWithDirectionAttribute = false
				s.emitMetricsWithoutDirectionAttribute = true
			},
		},
		{
			name:              "Validate Start Time",
			config:            Config{Metrics: metadata.DefaultMetricsSettings()},
			bootTimeFunc:      func() (uint64, error) { return 100, nil },
			expectMetrics:     metricsLen,
			expectedStartTime: 100 * 1e9,
		},
		{
			name:              "Boot Time Error",
			config:            Config{Metrics: metadata.DefaultMetricsSettings()},
			bootTimeFunc:      func() (uint64, error) { return 0, errors.New("err1") },
			initializationErr: "err1",
			expectMetrics:     metricsLen,
		},
		{
			name: "Include Filter that matches nothing",
			config: Config{
				Metrics: metadata.DefaultMetricsSettings(),
				Include: MatchConfig{filterset.Config{MatchType: "strict"}, []string{"@*^#&*$^#)"}},
			},
			expectMetrics: 0,
		},
		{
			name: "Invalid Include Filter",
			config: Config{
				Metrics: metadata.DefaultMetricsSettings(),
				Include: MatchConfig{Devices: []string{"test"}},
			},
			newErrRegex: "^error creating device include filters:",
		},
		{
			name: "Invalid Exclude Filter",
			config: Config{
				Metrics: metadata.DefaultMetricsSettings(),
				Exclude: MatchConfig{Devices: []string{"test"}},
			},
			newErrRegex: "^error creating device exclude filters:",
		},
		{
			name: "Disable one metric",
			config: (func() Config {
				config := Config{Metrics: metadata.DefaultMetricsSettings()}
				config.Metrics.SystemDiskIo.Enabled = false
				return config
			})(),
			expectMetrics: metricsLen - 1,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			scraper, err := newDiskScraper(context.Background(), componenttest.NewNopReceiverCreateSettings(), &test.config)
			if test.mutateScraper != nil {
				test.mutateScraper(scraper)
			}
			if test.newErrRegex != "" {
				require.Error(t, err)
				require.Regexp(t, test.newErrRegex, err)
				return
			}
			require.NoError(t, err, "Failed to create disk scraper: %v", err)

			if test.bootTimeFunc != nil {
				scraper.bootTime = test.bootTimeFunc
			}

			err = scraper.start(context.Background(), componenttest.NewNopHost())
			if test.initializationErr != "" {
				assert.EqualError(t, err, test.initializationErr)
				return
			}
			require.NoError(t, err, "Failed to initialize disk scraper: %v", err)

			md, err := scraper.scrape(context.Background())
			require.NoError(t, err, "Failed to scrape metrics: %v", err)

			assert.Equal(t, test.expectMetrics, md.MetricCount())
			if md.ResourceMetrics().Len() == 0 {
				return
			}

			metrics := md.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics()
			assert.Equal(t, test.expectMetrics, metrics.Len())

			reportedMetricsCount := map[string]int{}
			for i := 0; i < metrics.Len(); i++ {
				metric := metrics.At(i)
				reportedMetricsCount[metric.Name()]++
				switch metric.Name() {
				case "system.disk.io":
					assertInt64DiskMetricValid(t, metric, true, test.expectedStartTime)
				case "system.disk.io.read":
					assertInt64DiskMetricValid(t, metric, false, test.expectedStartTime)
				case "system.disk.io.write":
					assertInt64DiskMetricValid(t, metric, false, test.expectedStartTime)
				case "system.disk.io_time":
					assertDoubleDiskMetricValid(t, metric, false, test.expectedStartTime)
				case "system.disk.operation_time":
					assertDoubleDiskMetricValid(t, metric, true, test.expectedStartTime)
				case "system.disk.operation_time.read":
					assertDoubleDiskMetricValid(t, metric, false, test.expectedStartTime)
				case "system.disk.operation_time.write":
					assertDoubleDiskMetricValid(t, metric, false, test.expectedStartTime)
				case "system.disk.operations":
					assertInt64DiskMetricValid(t, metric, true, test.expectedStartTime)
				case "system.disk.operations.read":
					assertInt64DiskMetricValid(t, metric, false, test.expectedStartTime)
				case "system.disk.operations.write":
					assertInt64DiskMetricValid(t, metric, false, test.expectedStartTime)
				case "system.disk.weighted.io.time":
					assertDoubleDiskMetricValid(t, metric, false, test.expectedStartTime)
				case "system.disk.merged":
					assertInt64DiskMetricValid(t, metric, true, test.expectedStartTime)
				case "system.disk.merged.read":
					assertInt64DiskMetricValid(t, metric, false, test.expectedStartTime)
				case "system.disk.merged.write":
					assertInt64DiskMetricValid(t, metric, false, test.expectedStartTime)
				case "system.disk.pending_operations":
					assertDiskPendingOperationsMetricValid(t, metric)
				case "system.disk.weighted_io_time":
					assertDoubleDiskMetricValid(t, metric, false, test.expectedStartTime)
				default:
					assert.Failf(t, "unexpected-metric", "metric %q is not expected", metric.Name())
				}
			}
			for m, c := range reportedMetricsCount {
				assert.Equal(t, 1, c, "metric %q reported %d times", m, c)
			}

			internal.AssertSameTimeStampForAllMetrics(t, metrics)
		})
	}
}

func assertInt64DiskMetricValid(t *testing.T, metric pmetric.Metric, expectDirectionLabels bool, startTime pcommon.Timestamp) {
	if startTime != 0 {
		internal.AssertSumMetricStartTimeEquals(t, metric, startTime)
	}

	expectedDataPointsLen := 2
	if !expectDirectionLabels {
		expectedDataPointsLen = 1
	}
	assert.GreaterOrEqual(t, metric.Sum().DataPoints().Len(), expectedDataPointsLen)

	internal.AssertSumMetricHasAttribute(t, metric, 0, "device")
	if expectDirectionLabels {
		internal.AssertSumMetricHasAttributeValue(t, metric, 0, "direction",
			pcommon.NewValueString(metadata.AttributeDirectionRead.String()))
		internal.AssertSumMetricHasAttributeValue(t, metric, 1, "direction",
			pcommon.NewValueString(metadata.AttributeDirectionWrite.String()))
	}
}

func assertDoubleDiskMetricValid(t *testing.T, metric pmetric.Metric, expectDirectionLabels bool, startTime pcommon.Timestamp) {
	if startTime != 0 {
		internal.AssertSumMetricStartTimeEquals(t, metric, startTime)
	}

	minExpectedPoints := 1
	if expectDirectionLabels {
		minExpectedPoints = 2
	}
	assert.GreaterOrEqual(t, metric.Sum().DataPoints().Len(), minExpectedPoints)

	internal.AssertSumMetricHasAttribute(t, metric, 0, "device")
	if expectDirectionLabels {
		internal.AssertSumMetricHasAttributeValue(t, metric, 0, "direction",
			pcommon.NewValueString(metadata.AttributeDirectionRead.String()))
		internal.AssertSumMetricHasAttributeValue(t, metric, metric.Sum().DataPoints().Len()-1, "direction",
			pcommon.NewValueString(metadata.AttributeDirectionWrite.String()))
	}
}

func assertDiskPendingOperationsMetricValid(t *testing.T, metric pmetric.Metric) {
	assert.GreaterOrEqual(t, metric.Sum().DataPoints().Len(), 1)
	internal.AssertSumMetricHasAttribute(t, metric, 0, "device")
}
