// Copyright 2021 FerretDB Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package diag contains tests for diagnostics commands:
//   - buildInfo;
//   - collStats;
//   - connectionStatus;
//   - dataSize;
//   - dbStats;
//   - explain;
//   - getCmdLineOpts;
//   - getLog;
//   - hostInfo;
//   - listCommands;
//   - ping;
//   - serverStatus;
//   - validate;
//   - whatsmyuri.
package diag

import (
	"testing"

	"github.com/FerretDB/FerretDB/v2/integration/setup"
)

func TestMain(m *testing.M) {
	setup.Main(m)
}
