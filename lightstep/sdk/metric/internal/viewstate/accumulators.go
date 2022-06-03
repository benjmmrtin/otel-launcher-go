// Copyright The OpenTelemetry Authors
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

package viewstate // import "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/viewstate"

import (
	"sync"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
)

// multiAccumulator
type multiAccumulator[N number.Any] []Accumulator

func (acc multiAccumulator[N]) SnapshotAndProcess() {
	for _, coll := range acc {
		coll.SnapshotAndProcess()
	}
}

func (acc multiAccumulator[N]) Update(value N) {
	for _, coll := range acc {
		coll.(Updater[N]).Update(value)
	}
}

// syncAccumulator
type syncAccumulator[N number.Any, Storage any, Methods aggregator.Methods[N, Storage]] struct {
	// syncLock prevents two readers from calling
	// SnapshotAndProcess at the same moment.
	syncLock    sync.Mutex
	current     Storage
	snapshot    Storage
	findStorage func() *Storage
}

func (acc *syncAccumulator[N, Storage, Methods]) Update(number N) {
	var methods Methods
	methods.Update(&acc.current, number)
}

func (acc *syncAccumulator[N, Storage, Methods]) SnapshotAndProcess() {
	var methods Methods
	acc.syncLock.Lock()
	defer acc.syncLock.Unlock()
	methods.Move(&acc.current, &acc.snapshot)
	methods.Merge(&acc.snapshot, acc.findStorage())
}

// asyncAccumulator
type asyncAccumulator[N number.Any, Storage any, Methods aggregator.Methods[N, Storage]] struct {
	asyncLock   sync.Mutex
	current     N
	findStorage func() *Storage
}

func (acc *asyncAccumulator[N, Storage, Methods]) Update(number N) {
	acc.asyncLock.Lock()
	defer acc.asyncLock.Unlock()
	acc.current = number
}

func (acc *asyncAccumulator[N, Storage, Methods]) SnapshotAndProcess() {
	acc.asyncLock.Lock()
	defer acc.asyncLock.Unlock()

	var methods Methods
	methods.Update(acc.findStorage(), acc.current)
}
