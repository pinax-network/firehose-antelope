// Copyright 2019 dfuse Platform Inc.
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

package nodemanager

import (
	"context"
	"fmt"
	"github.com/streamingfast/dstore"
	"go.uber.org/zap"
	"io"
	"os"
	"path/filepath"
	"time"
)

func (s *NodeosSuperviser) Bootstrap() error {

	// no snapshot url given, nothing to do here
	if s.options.BootstrapSnapshotUrl == "" {
		return nil
	}

	s.Logger.Info("bootstrapping nodeos from snapshot file as node-reader-bootstrap-snapshot-url is set")
	s.Logger.Info("trying to download snapshot file...", zap.String("node-reader-bootstrap-snapshot-url", s.options.BootstrapSnapshotUrl))

	// todo is 30 min an appropriate timeout to download snapshots???
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	reader, _, snapshotFilename, err := dstore.OpenObject(ctx, s.options.BootstrapSnapshotUrl)
	if err != nil {
		return fmt.Errorf("cannot get snapshot from data store: %w", err)
	}
	defer reader.Close()

	s.snapshotRestoreFilename = snapshotFilename

	return storeSnapshotFile(reader, s.snapshotsDir, s.snapshotRestoreFilename)
}

func storeSnapshotFile(reader io.Reader, snapshotsDir, snapshotFilename string) error {
	err := os.MkdirAll(snapshotsDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to create snapshot directory: %w", err)
	}

	file, err := os.Create(filepath.Join(snapshotsDir, snapshotFilename))
	if err != nil {
		return fmt.Errorf("unable to create snapshot file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, reader)
	if err != nil {
		return fmt.Errorf("unable to write snapshot file: %w", err)
	}

	return nil
}
