// Copyright 2020 lack
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

package source

import "errors"

type noopWriter struct {
	exit chan struct{}
}

func (w *noopWriter) Next() (*ChangeSet, error) {
	<-w.exit

	return nil, errors.New("noopWatcher stopped")
}

func (w *noopWriter) Stop() error {
	close(w.exit)
	return nil
}

// NewNoopWatcher returns a watcher that blocks on Next() util Stop() is called.
func NewNoopWatcher() (Watcher, error) {
	return &noopWriter{exit: make(chan struct{})}, nil
}
