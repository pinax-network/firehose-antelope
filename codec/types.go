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

package codec

import (
	"github.com/eoscanada/eos-go"
	"github.com/pinax-network/firehose-antelope/codec/antelope"
	"github.com/pinax-network/firehose-antelope/types/pb/sf/antelope/type/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Specification struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
type SubjectiveRestrictions struct {
	Enable                        bool         `json:"enable"`
	PreactivationRequired         bool         `json:"preactivation_required"`
	EarliestAllowedActivationTime eos.JSONTime `json:"earliest_allowed_activation_time"`
}

//
/// Permission OP
//

type PermOp struct {
	Operation   string            `json:"op"`
	ActionIndex int               `json:"action_idx"`
	OldPerm     *permissionObject `json:"old,omitempty"`
	NewPerm     *permissionObject `json:"new,omitempty"`
}

// permissionObject represent the `nodeos` permission object that is stored
// in chainbase. Used to deserialize deep mind JSON to a correct `PermOp`.
type permissionObject struct {
	Owner       eos.AccountName `json:"owner"`
	Name        string          `json:"name"`
	ParentID    uint64          `json:"parent"`
	LastUpdated eos.JSONTime    `json:"last_updated"`
	Auth        *eos.Authority  `json:"auth"`
}

func (p *permissionObject) ToProto() *pbantelope.PermissionObject {
	return &pbantelope.PermissionObject{
		ParentId:    p.ParentID,
		Owner:       string(p.Owner),
		Name:        p.Name,
		LastUpdated: timestamppb.New(p.LastUpdated.Time),
		Authority:   antelope.AuthoritiesToDEOS(p.Auth),
	}
}

type rlimitState struct {
	ID                   uint32            `json:"id"`
	AverageBlockNetUsage *usageAccumulator `json:"average_block_net_usage"`
	AverageBlockCpuUsage *usageAccumulator `json:"average_block_cpu_usage"`
	PendingNetUsage      eos.Uint64        `json:"pending_net_usage"`
	PendingCpuUsage      eos.Uint64        `json:"pending_cpu_usage"`
	TotalNetWeight       eos.Uint64        `json:"total_net_weight"`
	TotalCpuWeight       eos.Uint64        `json:"total_cpu_weight"`
	TotalRamBytes        eos.Uint64        `json:"total_ram_bytes"`
	VirtualNetLimit      eos.Uint64        `json:"virtual_net_limit"`
	VirtualCpuLimit      eos.Uint64        `json:"virtual_cpu_limit"`
}

func (s *rlimitState) ToProto() *pbantelope.RlimitOp_State {
	return &pbantelope.RlimitOp_State{
		State: &pbantelope.RlimitState{
			AverageBlockNetUsage: s.AverageBlockNetUsage.ToProto(),
			AverageBlockCpuUsage: s.AverageBlockCpuUsage.ToProto(),
			PendingNetUsage:      uint64(s.PendingNetUsage),
			PendingCpuUsage:      uint64(s.PendingCpuUsage),
			TotalNetWeight:       uint64(s.TotalNetWeight),
			TotalCpuWeight:       uint64(s.TotalCpuWeight),
			TotalRamBytes:        uint64(s.TotalRamBytes),
			VirtualNetLimit:      uint64(s.VirtualNetLimit),
			VirtualCpuLimit:      uint64(s.VirtualCpuLimit),
		},
	}
}

type rlimitConfig struct {
	ID                           uint32                 `json:"id"`
	CPULimitParameters           elasticLimitParameters `json:"cpu_limit_parameters"`
	NetLimitParameters           elasticLimitParameters `json:"net_limit_parameters"`
	AccountCpuUsageAverageWindow uint32                 `json:"account_cpu_usage_average_window"`
	AccountNetUsageAverageWindow uint32                 `json:"account_net_usage_average_window"`
}

func (c *rlimitConfig) ToProto() *pbantelope.RlimitOp_Config {
	return &pbantelope.RlimitOp_Config{
		Config: &pbantelope.RlimitConfig{
			CpuLimitParameters:           c.CPULimitParameters.ToProto(),
			NetLimitParameters:           c.NetLimitParameters.ToProto(),
			AccountCpuUsageAverageWindow: c.AccountCpuUsageAverageWindow,
			AccountNetUsageAverageWindow: c.AccountNetUsageAverageWindow,
		},
	}
}

type rlimitAccountLimits struct {
	Owner     eos.AccountName `json:"owner"`
	NetWeight eos.Int64       `json:"net_weight"`
	CpuWeight eos.Int64       `json:"cpu_weight"`
	RamBytes  eos.Int64       `json:"ram_bytes"`
}

func (u *rlimitAccountLimits) ToProto() *pbantelope.RlimitOp_AccountLimits {
	return &pbantelope.RlimitOp_AccountLimits{
		AccountLimits: &pbantelope.RlimitAccountLimits{
			Owner:     string(u.Owner),
			NetWeight: int64(u.NetWeight),
			CpuWeight: int64(u.CpuWeight),
			RamBytes:  int64(u.RamBytes),
		},
	}
}

type rlimitAccountUsage struct {
	Owner    eos.AccountName   `json:"owner"`
	NetUsage *usageAccumulator `json:"net_usage"`
	CpuUsage *usageAccumulator `json:"cpu_usage"`
	RamUsage eos.Uint64        `json:"ram_usage"`
}

func (c *rlimitAccountUsage) ToProto() *pbantelope.RlimitOp_AccountUsage {
	return &pbantelope.RlimitOp_AccountUsage{
		AccountUsage: &pbantelope.RlimitAccountUsage{
			Owner:    string(c.Owner),
			NetUsage: c.NetUsage.ToProto(),
			CpuUsage: c.CpuUsage.ToProto(),
			RamUsage: uint64(c.RamUsage),
		},
	}
}

type usageAccumulator struct {
	LastOrdinal uint32     `json:"last_ordinal"`
	ValueEx     eos.Uint64 `json:"value_ex"`
	Consumed    eos.Uint64 `json:"consumed"`
}

func (a *usageAccumulator) ToProto() *pbantelope.UsageAccumulator {
	return &pbantelope.UsageAccumulator{
		LastOrdinal: a.LastOrdinal,
		ValueEx:     uint64(a.ValueEx),
		Consumed:    uint64(a.Consumed),
	}
}

type elasticLimitParameters struct {
	Target        eos.Uint64 `json:"target"`
	Max           eos.Uint64 `json:"max"`
	Periods       uint32     `json:"periods"`
	MaxMultiplier uint32     `json:"max_multiplier"`
	ContractRate  ratio      `json:"contract_rate"`
	ExpandRate    ratio      `json:"expand_rate"`
}

func (p *elasticLimitParameters) ToProto() *pbantelope.ElasticLimitParameters {
	return &pbantelope.ElasticLimitParameters{
		Target:        uint64(p.Target),
		Max:           uint64(p.Max),
		Periods:       p.Periods,
		MaxMultiplier: p.MaxMultiplier,
		ContractRate:  p.ContractRate.ToProto(),
		ExpandRate:    p.ExpandRate.ToProto(),
	}
}

type ratio struct {
	Numerator   eos.Uint64 `json:"numerator"`
	Denominator eos.Uint64 `json:"denominator"`
}

func (r *ratio) ToProto() *pbantelope.Ratio {
	return &pbantelope.Ratio{
		Numerator:   uint64(r.Numerator),
		Denominator: uint64(r.Denominator),
	}
}
