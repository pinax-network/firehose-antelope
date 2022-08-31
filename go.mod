module github.com/streamingfast/firehose-acme

go 1.16

require (
	github.com/ShinyTrinkets/overseer v0.3.0
	github.com/spf13/cobra v1.4.0
	github.com/spf13/viper v1.8.1
	github.com/streamingfast/bstream v0.0.2-0.20220831142019-0a0b0caa04c3
	github.com/streamingfast/cli v0.0.4-0.20220113202443-f7bcefa38f7e
	github.com/streamingfast/dauth v0.0.0-20220404140613-a40f4cd81626
	github.com/streamingfast/derr v0.0.0-20220526184630-695c21740145
	github.com/streamingfast/dgrpc v0.0.0-20220829132503-5a93eff857a6
	github.com/streamingfast/dlauncher v0.0.0-20220307153121-5674e1b64d40
	github.com/streamingfast/dmetering v0.0.0-20220307162406-37261b4b3de9
	github.com/streamingfast/dmetrics v0.0.0-20220811180000-3e513057d17c
	github.com/streamingfast/dstore v0.1.1-0.20220607202639-35118aeaf648
	github.com/streamingfast/dtracing v0.0.0-20220305214756-b5c0e8699839 // indirect
	github.com/streamingfast/firehose v0.1.1-0.20220830152055-a665bb4336d3
	github.com/streamingfast/firehose-acme/types v0.0.0-20220831185201-05ffef22e3a0
	github.com/streamingfast/logging v0.0.0-20220511154537-ce373d264338
	github.com/streamingfast/merger v0.0.3-0.20220728182037-7b5841d3c98f
	github.com/streamingfast/node-manager v0.0.2-0.20220728194124-81a202e6dcfa
	github.com/streamingfast/pbgo v0.0.6-0.20220630154121-2e8bba36234e
	github.com/streamingfast/relayer v0.0.2-0.20220729104233-021c788e2cc1
	github.com/streamingfast/sf-tools v0.0.0-20220727183125-3348eaca0e25
	github.com/streamingfast/substreams v0.0.21-0.20220831182608-4b383d73e6dc
	github.com/stretchr/testify v1.8.0
	go.uber.org/zap v1.21.0
	google.golang.org/grpc v1.49.0
)

replace github.com/ShinyTrinkets/overseer => github.com/streamingfast/overseer v0.2.1-0.20210326144022-ee491780e3ef
