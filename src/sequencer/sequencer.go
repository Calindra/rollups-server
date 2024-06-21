package sequencer

import (
	"github.com/calindra/rollups-server/src/model"
)

type InputBoxSequencer struct {
	model *model.AppModel
}

func NewInputBoxSequencer(model *model.AppModel) *InputBoxSequencer {
	return &InputBoxSequencer{model: model}
}

func NewEspressoSequencer(model *model.AppModel) *EspressoSequencer {
	return &EspressoSequencer{model: model}
}

type EspressoSequencer struct {
	model *model.AppModel
}

func (es *EspressoSequencer) FinishAndGetNext(accept bool) model.Input {
	return FinishAndGetNext(es.model, accept)
}

func FinishAndGetNext(m *model.AppModel, accept bool) model.Input {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	// finish current input
	var status model.CompletionStatus
	if accept {
		status = model.CompletionStatusAccepted
	} else {
		status = model.CompletionStatusRejected
	}
	m.State.Finish(status)

	// try to get first unprocessed inspect
	for _, input := range m.Inspects {
		if input.Status == model.CompletionStatusUnprocessed {
			m.State = model.NewRollupsStateInspect(input, m.GetProcessedInputCount)
			return *input
		}
	}

	// try to get first unprocessed advance
	input, err := m.InputRepository.FindByStatus(model.CompletionStatusUnprocessed)
	if err != nil {
		panic(err)
	}
	if input != nil {
		m.State = model.NewRollupsStateAdvance(
			input,
			m.Decoder,
			m.ReportRepository,
			m.InputRepository,
		)
		return *input
	}

	// if no input was found, set state to idle
	m.State = model.NewRollupsStateIdle()
	return nil
}

func (ibs *InputBoxSequencer) FinishAndGetNext(accept bool) model.Input {
	return FinishAndGetNext(ibs.model, accept)
}

type Sequencer interface {
	FinishAndGetNext(accept bool) model.Input
}
