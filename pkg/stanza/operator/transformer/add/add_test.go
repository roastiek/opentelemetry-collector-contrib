// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package add

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/entry"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/testutil"
)

type testCase struct {
	name      string
	op        *Config
	input     func() *entry.Entry
	output    func() *entry.Entry
	expectErr bool
}

func TestProcessAndBuild(t *testing.T) {
	now := time.Now()
	newTestEntry := func() *entry.Entry {
		e := entry.New()
		e.ObservedTimestamp = now
		e.Timestamp = time.Unix(1586632809, 0)
		e.Body = map[string]interface{}{
			"key": "val",
			"nested": map[string]interface{}{
				"nestedkey": "nestedval",
			},
		}
		return e
	}

	cases := []testCase{
		{
			"add_value",
			func() *Config {
				cfg := defaultCfg()
				cfg.Field = entry.NewBodyField("new")
				cfg.Value = "randomMessage"
				return cfg
			}(),
			newTestEntry,
			func() *entry.Entry {
				e := newTestEntry()
				e.Body.(map[string]interface{})["new"] = "randomMessage"
				return e
			},
			false,
		},
		{
			"add_expr",
			func() *Config {
				cfg := defaultCfg()
				cfg.Field = entry.NewBodyField("new")
				cfg.Value = `EXPR(body.key + "_suffix")`
				return cfg
			}(),
			newTestEntry,
			func() *entry.Entry {
				e := newTestEntry()
				e.Body.(map[string]interface{})["new"] = "val_suffix"
				return e
			},
			false,
		},
		{
			"add_nest",
			func() *Config {
				cfg := defaultCfg()
				cfg.Field = entry.NewBodyField("new")
				cfg.Value = map[interface{}]interface{}{
					"nest": map[interface{}]interface{}{
						"key": "val",
					},
				}
				return cfg
			}(),
			newTestEntry,
			func() *entry.Entry {
				e := newTestEntry()
				e.Body = map[string]interface{}{
					"key": "val",
					"nested": map[string]interface{}{
						"nestedkey": "nestedval",
					},
					"new": map[interface{}]interface{}{
						"nest": map[interface{}]interface{}{
							"key": "val",
						},
					},
				}
				return e
			},
			false,
		},
		{
			"add_attribute",
			func() *Config {
				cfg := defaultCfg()
				cfg.Field = entry.NewAttributeField("new")
				cfg.Value = "some.attribute"
				return cfg
			}(),
			newTestEntry,
			func() *entry.Entry {
				e := newTestEntry()
				e.Attributes = map[string]interface{}{"new": "some.attribute"}
				return e
			},
			false,
		},
		{
			"add_resource",
			func() *Config {
				cfg := defaultCfg()
				cfg.Field = entry.NewResourceField("new")
				cfg.Value = "some.resource"
				return cfg
			}(),
			newTestEntry,
			func() *entry.Entry {
				e := newTestEntry()
				e.Resource = map[string]interface{}{"new": "some.resource"}
				return e
			},
			false,
		},
		{
			"add_resource_expr",
			func() *Config {
				cfg := defaultCfg()
				cfg.Field = entry.NewResourceField("new")
				cfg.Value = `EXPR(body.key + "_suffix")`
				return cfg
			}(),
			newTestEntry,
			func() *entry.Entry {
				e := newTestEntry()
				e.Resource = map[string]interface{}{"new": "val_suffix"}
				return e
			},
			false,
		},
		{
			"add_int_to_body",
			func() *Config {
				cfg := defaultCfg()
				cfg.Field = entry.NewBodyField("new")
				cfg.Value = 1
				return cfg
			}(),
			newTestEntry,
			func() *entry.Entry {
				e := newTestEntry()
				e.Body = map[string]interface{}{
					"key": "val",
					"nested": map[string]interface{}{
						"nestedkey": "nestedval",
					},
					"new": 1,
				}
				return e
			},
			false,
		},
		{
			"add_array_to_body",
			func() *Config {
				cfg := defaultCfg()
				cfg.Field = entry.NewBodyField("new")
				cfg.Value = []int{1, 2, 3, 4}
				return cfg
			}(),
			newTestEntry,
			func() *entry.Entry {
				e := newTestEntry()
				e.Body = map[string]interface{}{
					"key": "val",
					"nested": map[string]interface{}{
						"nestedkey": "nestedval",
					},
					"new": []int{1, 2, 3, 4},
				}
				return e
			},
			false,
		},
		{
			"overwrite",
			func() *Config {
				cfg := defaultCfg()
				cfg.Field = entry.NewBodyField("key")
				cfg.Value = []int{1, 2, 3, 4}
				return cfg
			}(),
			newTestEntry,
			func() *entry.Entry {
				e := newTestEntry()
				e.Body = map[string]interface{}{
					"key": []int{1, 2, 3, 4},
					"nested": map[string]interface{}{
						"nestedkey": "nestedval",
					},
				}
				return e
			},
			false,
		},
		{
			"add_int_to_resource",
			func() *Config {
				cfg := defaultCfg()
				cfg.Field = entry.NewResourceField("new")
				cfg.Value = 1
				return cfg
			}(),
			newTestEntry,
			func() *entry.Entry {
				e := newTestEntry()
				e.Resource = map[string]interface{}{
					"new": 1,
				}
				return e
			},
			false,
		},
		{
			"add_int_to_attributes",
			func() *Config {
				cfg := defaultCfg()
				cfg.Field = entry.NewAttributeField("new")
				cfg.Value = 1
				return cfg
			}(),
			newTestEntry,
			func() *entry.Entry {
				e := newTestEntry()
				e.Attributes = map[string]interface{}{
					"new": 1,
				}
				return e
			},
			false,
		},
		{
			"add_nested_to_attributes",
			func() *Config {
				cfg := defaultCfg()
				cfg.Field = entry.NewAttributeField("one", "two")
				cfg.Value = map[string]interface{}{
					"new": 1,
				}
				return cfg
			}(),
			newTestEntry,
			func() *entry.Entry {
				e := newTestEntry()
				e.Attributes = map[string]interface{}{
					"one": map[string]interface{}{
						"two": map[string]interface{}{
							"new": 1,
						},
					},
				}
				return e
			},
			false,
		},
		{
			"add_nested_to_resource",
			func() *Config {
				cfg := defaultCfg()
				cfg.Field = entry.NewResourceField("one", "two")
				cfg.Value = map[string]interface{}{
					"new": 1,
				}
				return cfg
			}(),
			newTestEntry,
			func() *entry.Entry {
				e := newTestEntry()
				e.Resource = map[string]interface{}{
					"one": map[string]interface{}{
						"two": map[string]interface{}{
							"new": 1,
						},
					},
				}
				return e
			},
			false,
		},
	}
	for _, tc := range cases {
		t.Run("BuildandProcess/"+tc.name, func(t *testing.T) {
			cfg := tc.op
			cfg.OutputIDs = []string{"fake"}
			cfg.OnError = "drop"
			op, err := cfg.Build(testutil.Logger(t))
			require.NoError(t, err)

			add := op.(*Transformer)
			fake := testutil.NewFakeOutput(t)
			require.NoError(t, add.SetOutputs([]operator.Operator{fake}))
			val := tc.input()
			err = add.Process(context.Background(), val)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				fake.ExpectEntry(t, tc.output())
			}
		})
	}
}
