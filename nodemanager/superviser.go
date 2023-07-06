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
	"fmt"
	"github.com/ShinyTrinkets/overseer"
	"github.com/eoscanada/eos-go"
	nodeManager "github.com/streamingfast/node-manager"
	logplugin "github.com/streamingfast/node-manager/log_plugin"
	"github.com/streamingfast/node-manager/superviser"
	"go.uber.org/zap"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type NodeosSuperviser struct {
	*superviser.Superviser
	name string

	api          *eos.API
	blocksDir    string
	options      *NodeosOptions
	snapshotsDir string

	chainID       eos.SHA256Bytes
	lastBlockSeen uint32

	serverVersion       string
	serverVersionString string

	snapshotRestorePath string

	headBlockUpdateFunc nodeManager.HeadBlockUpdater

	logger *zap.Logger
}

func (s *NodeosSuperviser) GetName() string {
	return "nodeos"
}

type NodeosOptions struct {

	// LocalNodeEndpoint is the URL to reach the locally managed node (`http://localhost:8888` if empty)
	LocalNodeEndpoint string

	// ConfigPath points to the path where the config.ini lives (`/etc/nodeos` if empty)
	ConfigDir string

	// BinPath points to the file system location of the`nodeos` binary. Required.
	BinPath string

	// DataDir points to the location of the nodeos data dir. Required.
	DataDir string

	// BootstrapSnapshotUrl points to
	BootstrapSnapshotUrl string

	// AdditionalArgs are parameters you want to pass down to `nodeos`
	// in addition to the ones `node manager` would add itself.  You're
	// better off, putting long-running parameters in the `config.ini`
	// though.
	AdditionalArgs []string

	// Redirects all output to zlog instance configured for this process
	// instead of the standard console output
	LogToZap bool
}

func NewSuperviser(
	debugDeepMind bool,
	headBlockUpdateFunc nodeManager.HeadBlockUpdater,
	nodeosOptions *NodeosOptions,
	logger *zap.Logger,
	nodeosLogger *zap.Logger,
) (*NodeosSuperviser, error) {
	// Ensure process manager line buffer is large enough (50 MiB) for our Deep Mind instrumentation outputting lots of text.
	overseer.DEFAULT_LINE_BUFFER_SIZE = 50 * 1024 * 1024

	if !strings.HasPrefix(nodeosOptions.LocalNodeEndpoint, "http") {
		nodeosOptions.LocalNodeEndpoint = fmt.Sprintf("http://%s", nodeosOptions.LocalNodeEndpoint)
	}

	s := &NodeosSuperviser{
		// The arguments field is actually `nil` because arguments are re-computed upon each start
		Superviser:          superviser.New(logger, nodeosOptions.BinPath, nil),
		api:                 eos.New(nodeosOptions.LocalNodeEndpoint),
		blocksDir:           filepath.Join(nodeosOptions.DataDir, "blocks"),
		snapshotsDir:        filepath.Join(nodeosOptions.DataDir, "snapshots"),
		options:             nodeosOptions,
		headBlockUpdateFunc: headBlockUpdateFunc,
		logger:              logger,
	}

	if nodeosOptions.LogToZap {
		s.RegisterLogPlugin(newToZapLogPlugin(debugDeepMind, nodeosLogger))
	} else {
		s.RegisterLogPlugin(logplugin.NewToConsoleLogPlugin(debugDeepMind))
	}

	return s, nil
}

func (s *NodeosSuperviser) GetCommand() string {
	return s.options.BinPath + " " + strings.Join(s.getArguments(), " ")
}

func (s *NodeosSuperviser) GetBlocksDir() string {
	return s.blocksDir
}

func (s *NodeosSuperviser) HasData() bool {
	dir, err := os.ReadDir(s.blocksDir)
	if err != nil || len(dir) == 0 {
		return false
	}

	return true
}

func (s *NodeosSuperviser) removeState() error {
	stateDir := path.Join(s.options.DataDir, "state")
	dir, err := os.ReadDir(stateDir)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("cannot read state directory: %w", err)
	}

	for _, file := range dir {
		err = os.RemoveAll(path.Join(stateDir, file.Name()))
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("cannot delete state element: %w", err)
		}
	}

	return nil
}

func (s *NodeosSuperviser) removeBlocksLog() error {
	err := os.Remove(path.Join(s.blocksDir, "blocks.log"))
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("cannot delete blocks.log before starting: %w", err)
	}
	err = os.Remove(path.Join(s.blocksDir, "blocks.index"))
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("cannot delete blocks.index directory before starting: %w", err)
	}
	return nil
}

func (s *NodeosSuperviser) removeReversibleBlocks() error {
	err := os.RemoveAll(path.Join(s.blocksDir, "reversible"))
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("cannot delete blocks/reversible directory before starting: %w", err)
	}
	return nil
}

func (s *NodeosSuperviser) Start(options ...nodeManager.StartOption) error {
	s.Logger.Info("updating nodeos arguments before starting binary")
	s.Superviser.Arguments = s.getArguments()

	err := s.Superviser.Start(options...)
	if err != nil {
		return err
	}

	return nil
}

func (s *NodeosSuperviser) IsRunning() bool {
	isRunning := s.Superviser.IsRunning()

	isRunningMetricsValue := float64(0)
	if isRunning {
		isRunningMetricsValue = float64(1)
	}
	leapStatus.SetFloat64(isRunningMetricsValue)

	return isRunning
}

func (s *NodeosSuperviser) LastSeenBlockNum() uint64 {
	return uint64(s.lastBlockSeen)
}

func (s *NodeosSuperviser) ServerID() (string, error) {
	return os.Hostname()
}

func (s *NodeosSuperviser) getArguments() []string {
	var args []string
	args = append(args, fmt.Sprintf("--config-dir=%s", s.options.ConfigDir))
	args = append(args, fmt.Sprintf("--data-dir=%s", s.options.DataDir))

	if s.snapshotRestorePath != "" {
		args = append(args, fmt.Sprintf("--snapshot=%s", s.snapshotRestorePath))
	} else {
		args = append(args, fmt.Sprintf("--genesis-json=%s", filepath.Join(s.options.ConfigDir, "genesis.json")))
	}

	args = append(args, s.options.AdditionalArgs...)

	return args
}
