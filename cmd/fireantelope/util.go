package main

import (
	"fmt"
	"strconv"
)

func mustParseUint64(in string) uint64 {
	out, err := strconv.ParseUint(in, 10, 64)
	if err != nil {
		panic(fmt.Errorf("unable to parse %q as uint64: %w", in, err))
	}

	return out
}
