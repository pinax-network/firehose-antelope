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

	s.Logger.Info("bootstrapping nodeos from snapshot file as reader-node-bootstrap-snapshot-url is set")
	s.Logger.Info("trying to download snapshot file...", zap.String("reader-node-bootstrap-snapshot-url", s.options.BootstrapSnapshotUrl))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	reader, _, snapshotFilename, err := dstore.OpenObject(ctx, s.options.BootstrapSnapshotUrl)
	if err != nil {
		return fmt.Errorf("cannot get snapshot from data store: %w", err)
	}
	defer reader.Close()

	s.snapshotRestorePath = filepath.Join(s.snapshotsDir, snapshotFilename)

	// we check whether we have already downloaded the snapshot before and ignore the reader-node-bootstrap-snapshot-url
	// if so. This is so the node manager does not replay the snapshot again if it gets restarted.
	if _, err := os.Stat(s.snapshotRestorePath); err == nil {
		s.Logger.Warn("a snapshot with this name already has been downloaded before, the reader-node-bootstrap-snapshot-url will be ignored!")
		s.Logger.Warn("if you want to replay the snapshot again, deleted the local snapshot and restart the node manager", zap.String("snapshot_file", s.snapshotRestorePath))
		s.snapshotRestorePath = ""
		return nil
	}


	err = storeSnapshotFile(reader, s.snapshotsDir, s.snapshotRestorePath)
	if err != nil {
		return err
	}

	s.Logger.Info("successfully downloaded snapshot", zap.String("snapshot_path", s.snapshotRestorePath))

	err = s.removeState()
	if err != nil {
		return err
	}

	err = s.removeBlocksLog()
	if err != nil {
		return err
	}

	err = s.removeReversibleBlocks()
	if err != nil {
		return err
	}

	return err
}

func storeSnapshotFile(reader io.Reader, snapshotsDir, snapshotPath string) error {
	err := os.MkdirAll(snapshotsDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to create snapshot directory: %w", err)
	}

	file, err := os.Create(snapshotPath)
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
