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

type BootstrapOptions struct {

	// todo replace stores with dstores??
	BackupTag            string
	BackupStoreURL       string
	SnapshotStoreURL     string
	VolumeSnapshotAppVer string
	Namespace            string //k8s namespace
	Pod                  string //k8s podname
	PVCPrefix            string
	Project              string //gcp project

	BootstrapSnapshotName   string
	BootstrapDataURL        string
	AutoRestoreSource       string
	NumberOfSnapshotsToKeep int
	RestoreBackupName       string
	RestoreSnapshotName     string
	BlocksDir               string
}

type NodeosBootstrapper struct {
	Options BootstrapOptions
	Logger  *zap.Logger
}

func (b *NodeosBootstrapper) Bootstrap() error {
	b.Logger.Info("bootstrapping blocks.log from pre-built data", zap.String("bootstrap_data_url", b.Options.BootstrapDataURL))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	reader, _, _, err := dstore.OpenObject(ctx, b.Options.BootstrapDataURL)
	if err != nil {
		return fmt.Errorf("cannot get snapshot from data store: %w", err)
	}
	defer reader.Close()

	return b.createBlocksLogFile(reader)
}

//func (s *NodeosSuperviser) Bootstrap(bootstrapDataURL string) error {
//	s.Logger.Info("bootstrapping blocks.log from pre-built data", zap.String("bootstrap_data_url", bootstrapDataURL))
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
//	defer cancel()
//
//	reader, _, _, err := dstore.OpenObject(ctx, bootstrapDataURL)
//	if err != nil {
//		return fmt.Errorf("cannot get snapshot from data store: %w", err)
//	}
//	defer reader.Close()
//
//	return s.createBlocksLogFile(reader)
//}

func (b *NodeosBootstrapper) createBlocksLogFile(reader io.Reader) error {
	err := os.MkdirAll(b.Options.BlocksDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to create blocks log file: %w", err)
	}

	file, err := os.Create(filepath.Join(b.Options.BlocksDir, "blocks.log"))
	if err != nil {
		return fmt.Errorf("unable to create blocks log file: %w", err)
	}

	defer file.Close()

	_, err = io.Copy(file, reader)
	if err != nil {
		return fmt.Errorf("unable to create blocks log file: %w", err)
	}

	return nil
}
