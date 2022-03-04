module github.com/streamingfast/firehose-acme

go 1.16

require (
	cloud.google.com/go/monitoring v1.4.0 // indirect
	cloud.google.com/go/trace v1.2.0 // indirect
	github.com/ShinyTrinkets/overseer v0.3.0
	github.com/abourget/llerrgroup v0.2.0 // indirect
	github.com/golang/protobuf v1.5.2
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.8.1
	github.com/streamingfast/bstream v0.0.2-0.20220301162141-6630bbe5996c
	github.com/streamingfast/cli v0.0.4-0.20220113202443-f7bcefa38f7e
	github.com/streamingfast/dauth v0.0.0-20210812020920-1c83ba29add1
	github.com/streamingfast/dbin v0.0.0-20210809205249-73d5eca35dc5
	github.com/streamingfast/derr v0.0.0-20220301163149-de09cb18fc70
	github.com/streamingfast/dgrpc v0.0.0-20220301153539-536adf71b594
	github.com/streamingfast/dlauncher v0.0.0-20220304184304-e5045b95f9ad
	github.com/streamingfast/dmetering v0.0.0-20220301165106-a642bb6a21bd
	github.com/streamingfast/dmetrics v0.0.0-20210811180524-8494aeb34447
	github.com/streamingfast/dstore v0.1.1-0.20220304164644-696f9c5fc231 // indirect
	github.com/streamingfast/firehose v0.1.1-0.20220301165907-a8296f77f3bd
	github.com/streamingfast/logging v0.0.0-20220304183711-ddba33d79e27
	github.com/streamingfast/merger v0.0.3-0.20220301162603-c0129b6f1ad4
	github.com/streamingfast/node-manager v0.0.2-0.20220301170656-5dbc7988e730
	github.com/streamingfast/pbgo v0.0.6-0.20220228185940-1bbaafec7d8a
	github.com/streamingfast/relayer v0.0.2-0.20220301162545-2db510359d2a
	github.com/streamingfast/sf-tools v0.0.0-20220301170200-43b1f43dde6f
	github.com/stretchr/testify v1.7.0
	go.uber.org/zap v1.21.0
	google.golang.org/grpc v1.44.0
	google.golang.org/protobuf v1.27.1
)

replace github.com/ShinyTrinkets/overseer => github.com/streamingfast/overseer v0.2.1-0.20210326144022-ee491780e3ef
