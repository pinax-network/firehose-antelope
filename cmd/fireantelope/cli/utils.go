// Copyright 2021 dfuse Platform Inc.
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

package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lithammer/dedent"
	"github.com/logrusorgru/aurora"
	"github.com/spf13/viper"
	"github.com/streamingfast/dstore"
	"github.com/streamingfast/logging"
	"go.uber.org/zap"
)

var DefaultLevelInfo = logging.LoggerDefaultLevel(zap.InfoLevel)

func MustReplaceDataDir(dataDir, in string) string {
	d, err := filepath.Abs(dataDir)
	if err != nil {
		panic(fmt.Errorf("file path abs: %w", err))
	}

	in = strings.Replace(in, "{sf-data-dir}", d, -1)
	return in
}

var commonStoresCreated bool
var indexStoreCreated bool

func mustGetCommonStoresURLs(dataDir string) (mergedBlocksStoreURL, oneBlocksStoreURL, forkedBlocksStoreURL string) {
	var err error
	mergedBlocksStoreURL, oneBlocksStoreURL, forkedBlocksStoreURL, err = getCommonStoresURLs(dataDir)
	if err != nil {
		panic(err)
	}

	return
}

func getCommonStoresURLs(dataDir string) (mergedBlocksStoreURL, oneBlocksStoreURL, forkedBlocksStoreURL string, err error) {
	mergedBlocksStoreURL = MustReplaceDataDir(dataDir, viper.GetString("common-merged-blocks-store-url"))
	oneBlocksStoreURL = MustReplaceDataDir(dataDir, viper.GetString("common-one-block-store-url"))
	forkedBlocksStoreURL = MustReplaceDataDir(dataDir, viper.GetString("common-forked-blocks-store-url"))

	if commonStoresCreated {
		return
	}

	if err = mkdirStorePathIfLocal(forkedBlocksStoreURL); err != nil {
		return
	}
	if err = mkdirStorePathIfLocal(oneBlocksStoreURL); err != nil {
		return
	}
	if err = mkdirStorePathIfLocal(mergedBlocksStoreURL); err != nil {
		return
	}
	commonStoresCreated = true
	return
}

func GetIndexStore(dataDir string) (indexStore dstore.Store, possibleIndexSizes []uint64, err error) {
	indexStoreURL := MustReplaceDataDir(dataDir, viper.GetString("common-index-store-url"))

	if indexStoreURL != "" {
		s, err := dstore.NewStore(indexStoreURL, "", "", false)
		if err != nil {
			return nil, nil, fmt.Errorf("couldn't create index store: %w", err)
		}
		if !indexStoreCreated {
			if err = mkdirStorePathIfLocal(indexStoreURL); err != nil {
				return nil, nil, err
			}
		}
		indexStoreCreated = true
		indexStore = s
	}

	for _, size := range viper.GetIntSlice("common-block-index-sizes") {
		if size < 0 {
			return nil, nil, fmt.Errorf("invalid negative size for common-block-index-sizes: %d", size)
		}
		possibleIndexSizes = append(possibleIndexSizes, uint64(size))
	}
	return
}

func mkdirStorePathIfLocal(storeURL string) (err error) {
	if dirs := getDirsToMake(storeURL); len(dirs) > 0 {
		zlog.Debug("creating directory and its parent(s)", zap.String("directory", storeURL))
		err = makeDirs(dirs)
	}
	return
}

func getDirsToMake(storeURL string) []string {
	parts := strings.Split(storeURL, "://")
	if len(parts) > 1 {
		if parts[0] != "file" {
			// Not a local store, nothing to do
			return nil
		}
		storeURL = parts[1]
	}

	// Some of the store URL are actually a file directly, let's try our best to cope for that case
	filename := filepath.Base(storeURL)
	if strings.Contains(filename, ".") {
		storeURL = filepath.Dir(storeURL)
	}

	// If we reach here, it's a local store path
	return []string{storeURL}
}

func makeDirs(directories []string) error {
	for _, directory := range directories {
		err := os.MkdirAll(directory, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory %q: %w", directory, err)
		}
	}

	return nil
}

func dfuseAbsoluteDataDir() (string, error) {
	return filepath.Abs(viper.GetString("global-data-dir"))
}

func cliErrorAndExit(message string) {
	fmt.Println(aurora.Red(message).String())
	os.Exit(1)
}

func dedentf(format string, args ...interface{}) string {
	return fmt.Sprintf(dedent.Dedent(strings.TrimPrefix(format, "\n")), args...)
}
