// Copyright (c) Gabriel de Quadros Ligneul
// SPDX-License-Identifier: Apache-2.0 (see LICENSE)

// The nonodo model uses a shared-memory paradigm to synchronize between threads.
package model

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/jmoiron/sqlx"
)

// Nonodo model shared among the internal workers.
// The model store inputs as pointers because these pointers are shared with the rollup state.
type AppModel struct {
	Mutex            sync.Mutex
	Inspects         []*InspectInput
	State            rollupsState
	Decoder          Decoder
	ReportRepository *ReportRepository
	InputRepository  *InputRepository
}

func NewAppModel(decoder Decoder, db *sqlx.DB) *AppModel {

	reportRepository := ReportRepository{Db: db}
	err := reportRepository.CreateTables()
	if err != nil {
		panic(err)
	}
	inputRepository := InputRepository{Db: db}
	err = inputRepository.CreateTables()
	if err != nil {
		panic(err)
	}
	return &AppModel{
		State:            &RollupsStateIdle{},
		Decoder:          decoder,
		ReportRepository: &reportRepository,
		InputRepository:  &inputRepository,
	}
}

func (m *AppModel) GetInputRepository() *InputRepository {
	return m.InputRepository
}

//
// Methods for Inputter
//

// Add an advance input to the model.
func (m *AppModel) AddAdvanceInput(
	sender common.Address,
	payload []byte,
	blockNumber uint64,
	timestamp time.Time,
	index int,
) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	input := AdvanceInput{
		Index:          index,
		Status:         CompletionStatusUnprocessed,
		MsgSender:      sender,
		Payload:        payload,
		BlockTimestamp: timestamp,
		BlockNumber:    blockNumber,
	}
	_, err := m.InputRepository.Create(input)
	if err != nil {
		panic(err)
	}
	slog.Info("nonodo: added advance input", "index", input.Index, "sender", input.MsgSender,
		"payload", hexutil.Encode(input.Payload))
}

//
// Methods for Inspector
//

// Add an inspect input to the model.
// Return the inspect input index that should be used for polling.
func (m *AppModel) AddInspectInput(payload []byte) int {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	index := len(m.Inspects)
	input := InspectInput{
		Index:   index,
		Status:  CompletionStatusUnprocessed,
		Payload: payload,
	}
	m.Inspects = append(m.Inspects, &input)
	slog.Info("nonodo: added inspect input", "index", input.Index,
		"payload", hexutil.Encode(input.Payload))

	return index
}

// Get the inspect input from the model.
func (m *AppModel) GetInspectInput(index int) InspectInput {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	if index >= len(m.Inspects) {
		panic(fmt.Sprintf("invalid inspect input index: %v", index))
	}
	return *m.Inspects[index]
}

//
// Methods for Rollups
//

// Finish the current input and get the next one.
// If there is no input to be processed return nil.
//
// Note: use in v2 the sequencer instead.
func (m *AppModel) FinishAndGetNext(accepted bool) Input {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	// finish current input
	var status CompletionStatus
	if accepted {
		status = CompletionStatusAccepted
	} else {
		status = CompletionStatusRejected
	}
	m.State.Finish(status)

	// try to get first unprocessed inspect
	for _, input := range m.Inspects {
		if input.Status == CompletionStatusUnprocessed {
			m.State = NewRollupsStateInspect(input, m.GetProcessedInputCount)
			return *input
		}
	}

	// try to get first unprocessed advance
	input, err := m.InputRepository.FindByStatus(CompletionStatusUnprocessed)
	if err != nil {
		panic(err)
	}
	if input != nil {
		m.State = NewRollupsStateAdvance(
			input,
			m.Decoder,
			m.ReportRepository,
			m.InputRepository,
		)
		return *input
	}

	// if no input was found, set state to idle
	m.State = NewRollupsStateIdle()
	return nil
}

// Add a voucher to the model.
// Return the voucher index within the input.
// Return an error if the state isn't advance.
func (m *AppModel) AddVoucher(destination common.Address, payload []byte) (int, error) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	return m.State.AddVoucher(destination, payload)
}

// Add a notice to the model.
// Return the notice index within the input.
// Return an error if the state isn't advance.
func (m *AppModel) AddNotice(payload []byte) (int, error) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	return m.State.AddNotice(payload)
}

// Add a report to the model.
// Return an error if the state isn't advance or inspect.
func (m *AppModel) AddReport(payload []byte) error {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	return m.State.AddReport(payload)
}

// Finish the current input with an exception.
// Return an error if the state isn't advance or inspect.
func (m *AppModel) RegisterException(payload []byte) error {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	err := m.State.RegisterException(payload)
	if err != nil {
		return err
	}

	// set state to idle
	m.State = NewRollupsStateIdle()
	return nil
}

//
// Auxiliary Methods
//

func (m *AppModel) GetProcessedInputCount() int {
	filter := []*ConvenienceFilter{}
	field := "Status"
	value := fmt.Sprintf("%d", CompletionStatusUnprocessed)
	filter = append(filter, &ConvenienceFilter{
		Field: &field,
		Ne:    &value,
	})
	total, err := m.InputRepository.Count(filter)
	if err != nil {
		panic(err)
	}
	return int(total)
}
