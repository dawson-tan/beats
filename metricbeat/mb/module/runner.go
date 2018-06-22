// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package module

import (
	"fmt"
	"sync"

	"github.com/elastic/beats/libbeat/beat"
)

// Runner is a facade for a Wrapper that provides a simple interface
// for starting and stopping a Module.
type Runner interface {
	// Start starts the Module. If Start is called more than once, only the
	// first will start the Module.
	Start()

	// Stop stops the Module and waits for module's MetricSets to exit. The
	// publisher.Client will be closed by Stop. If Stop is called more than
	// once, only the first stop the Module and wait for it to exit.
	Stop()
}

// NewRunner returns a Runner facade. The events generated by
// the Module will be published to a new publisher.Client generated from the
// pubClientFactory.
func NewRunner(client beat.Client, mod *Wrapper) Runner {
	return &runner{
		done:   make(chan struct{}),
		mod:    mod,
		client: client,
	}
}

type runner struct {
	done      chan struct{}
	wg        sync.WaitGroup
	startOnce sync.Once
	stopOnce  sync.Once
	mod       *Wrapper
	client    beat.Client
}

func (mr *runner) Start() {
	mr.startOnce.Do(func() {
		output := mr.mod.Start(mr.done)
		mr.wg.Add(1)
		go func() {
			defer mr.wg.Done()
			PublishChannels(mr.client, output)
		}()
	})
}

func (mr *runner) Stop() {
	mr.stopOnce.Do(func() {
		close(mr.done)
		mr.client.Close()
		mr.wg.Wait()
	})
}

func (mr *runner) String() string {
	return fmt.Sprintf("%s [metricsets=%d]", mr.mod.Name(), len(mr.mod.metricSets))
}
