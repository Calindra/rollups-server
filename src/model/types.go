// Copyright (c) Gabriel de Quadros Ligneul
// SPDX-License-Identifier: Apache-2.0 (see LICENSE)

package model

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
)

const INPUT_INDEX = "InputIndex"
const EXECUTED = "Executed"
const DESTINATION = "Destination"

// Rollups voucher type.
type Voucher struct {
	Index       int
	InputIndex  int
	Destination common.Address
	Payload     []byte
}

func (v Voucher) GetInputIndex() int {
	return v.InputIndex
}

// Rollups notice type.
type Notice struct {
	Index      int
	InputIndex int
	Payload    []byte
}

func (n Notice) GetInputIndex() int {
	return n.InputIndex
}

// Rollups report type.
type Report struct {
	Index      int
	InputIndex int
	Payload    []byte
}

func (r Report) GetInputIndex() int {
	return r.InputIndex
}

// Completion status for inputs.
type CompletionStatus int

const (
	CompletionStatusUnprocessed CompletionStatus = iota
	CompletionStatusAccepted
	CompletionStatusRejected
	CompletionStatusException
)

// Rollups input, which can be advance or inspect.
type Input interface{}

// Rollups advance input type.
type AdvanceInput struct {
	Index          int
	Status         CompletionStatus
	MsgSender      common.Address
	Payload        []byte
	BlockNumber    uint64
	BlockTimestamp time.Time
	PrevRandao     uint64
	Vouchers       []Voucher
	Notices        []Notice
	Reports        []Report
	Exception      []byte
}

// Rollups inspect input type.
type InspectInput struct {
	Index               int
	Status              CompletionStatus
	Payload             []byte
	ProcessedInputCount int
	Reports             []Report
	Exception           []byte
}

type ConvenienceFilter struct {
	Field *string              `json:"field,omitempty"`
	Eq    *string              `json:"eq,omitempty"`
	Ne    *string              `json:"ne,omitempty"`
	Gt    *string              `json:"gt,omitempty"`
	Gte   *string              `json:"gte,omitempty"`
	Lt    *string              `json:"lt,omitempty"`
	Lte   *string              `json:"lte,omitempty"`
	In    []*string            `json:"in,omitempty"`
	Nin   []*string            `json:"nin,omitempty"`
	And   []*ConvenienceFilter `json:"and,omitempty"`
	Or    []*ConvenienceFilter `json:"or,omitempty"`
}

// Voucher metadata type
type ConvenienceVoucher struct {
	Destination common.Address `db:"destination"`
	Payload     string         `db:"payload"`
	InputIndex  uint64         `db:"input_index"`
	OutputIndex uint64         `db:"output_index"`
	Executed    bool           `db:"executed"`

	// Proof we can fetch from the original GraphQL

	// future improvements
	// Contract        common.Address
	// Beneficiary     common.Address
	// Label           string
	// Amount          uint64
	// ExecutedAt      uint64
	// ExecutedBlock   uint64
	// InputIndex      int
	// OutputIndex     int
	// MethodSignature string
	// ERCX            string
}

type ConvenienceNotice struct {
	Payload     string `db:"payload"`
	InputIndex  uint64 `db:"input_index"`
	OutputIndex uint64 `db:"output_index"`
}
