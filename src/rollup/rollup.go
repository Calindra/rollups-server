// Copyright (c) Gabriel de Quadros Ligneul
// SPDX-License-Identifier: Apache-2.0 (see LICENSE)

// This package contains the bindings for the rollup OpenAPI spec.
package rollup

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -config=oapi.yaml ../../api/rollup.yaml

import (
	"log/slog"
	"net/http"

	"strings"
	"time"

	mdl "github.com/calindra/rollups-server/src/model"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/labstack/echo/v4"
)

const FinishRetries = 50
const FinishPollInterval = time.Millisecond * 100

// Register the rollup API to echo
func Register(e *echo.Echo, model *mdl.AppModel, sequencer Sequencer) {
	var rollupAPI ServerInterface = &RollupAPI{model, sequencer}
	RegisterHandlers(e, rollupAPI)
}

// Shared struct for request handlers.
type RollupAPI struct {
	model     *mdl.AppModel
	sequencer Sequencer
}

type Sequencer interface {
	FinishAndGetNext(accept bool) mdl.Input
}

// Gio implements ServerInterface.
func (r *RollupAPI) Gio(ctx echo.Context) error {

	if !checkContentType(ctx) {
		return ctx.String(http.StatusUnsupportedMediaType, "invalid content type")
	}

	// parse body
	var request GioJSONRequestBody
	if err := ctx.Bind(&request); err != nil {
		return err
	}

	fetch, err := r.Fetcher(ctx, request)

	if err != nil {
		slog.Debug("Error in Fetcher: %s %d", err.Error(), err.Status())
		return ctx.String(int(err.Status()), err.Error())
	}

	if fetch == nil {
		return ctx.String(http.StatusNotFound, "Not found")
	}

	return ctx.JSON(http.StatusOK, fetch)
}

// Handle requests to /finish.
func (r *RollupAPI) Finish(c echo.Context) error {
	if !checkContentType(c) {
		return c.String(http.StatusUnsupportedMediaType, "invalid content type")
	}

	// parse body
	var request FinishJSONRequestBody
	if err := c.Bind(&request); err != nil {
		return err
	}

	// validate fields
	var accepted bool
	switch request.Status {
	case Accept:
		accepted = true
	case Reject:
		accepted = false
	default:
		return c.String(http.StatusBadRequest, "invalid value for status")
	}

	// talk to model
	if r.sequencer == nil {
		return c.String(http.StatusInternalServerError, "sequencer not available")
	}
	for i := 0; i < FinishRetries; i++ {
		input := r.sequencer.FinishAndGetNext(accepted)
		if input != nil {
			resp := convertInput(input)
			return c.JSON(http.StatusOK, &resp)
		}
		ctx := c.Request().Context()
		select {
		case <-ctx.Done():
			return c.String(http.StatusInternalServerError, ctx.Err().Error())
		case <-time.After(FinishPollInterval):
		}
	}
	return c.String(http.StatusAccepted, "no rollup request available")
}

// Handle requests to /voucher.
func (r *RollupAPI) AddVoucher(c echo.Context) error {
	if !checkContentType(c) {
		return c.String(http.StatusUnsupportedMediaType, "invalid content type")
	}

	// parse body
	var request AddVoucherJSONRequestBody
	if err := c.Bind(&request); err != nil {
		return err
	}

	// validate fields
	destination, err := hexutil.Decode(request.Destination)
	if err != nil {
		return c.String(http.StatusBadRequest, "invalid hex payload")
	}
	if len(destination) != common.AddressLength {
		return c.String(http.StatusBadRequest, "invalid address length")
	}
	payload, err := hexutil.Decode(request.Payload)
	if err != nil {
		return c.String(http.StatusBadRequest, "invalid hex payload")
	}

	// talk to model
	index, err := r.model.AddVoucher(common.Address(destination), payload)
	if err != nil {
		return c.String(http.StatusForbidden, err.Error())
	}
	resp := IndexResponse{
		Index: uint64(index),
	}
	return c.JSON(http.StatusOK, &resp)
}

// Handle requests to /notice.
func (r *RollupAPI) AddNotice(c echo.Context) error {
	if !checkContentType(c) {
		return c.String(http.StatusUnsupportedMediaType, "invalid content type")
	}

	// parse body
	var request AddNoticeJSONRequestBody
	if err := c.Bind(&request); err != nil {
		return err
	}

	// validate fields
	payload, err := hexutil.Decode(request.Payload)
	if err != nil {
		return c.String(http.StatusBadRequest, "invalid hex payload")
	}

	// talk to model
	index, err := r.model.AddNotice(payload)
	if err != nil {
		return c.String(http.StatusForbidden, err.Error())
	}
	resp := IndexResponse{
		Index: uint64(index),
	}
	return c.JSON(http.StatusOK, &resp)
}

// Handle requests to /report.
func (r *RollupAPI) AddReport(c echo.Context) error {
	if !checkContentType(c) {
		return c.String(http.StatusUnsupportedMediaType, "invalid content type")
	}

	// parse body
	var request AddReportJSONRequestBody
	if err := c.Bind(&request); err != nil {
		return err
	}

	// validate fields
	payload, err := hexutil.Decode(request.Payload)
	if err != nil {
		return c.String(http.StatusBadRequest, "invalid hex payload")
	}

	// talk to model
	err = r.model.AddReport(payload)
	if err != nil {
		return c.String(http.StatusForbidden, err.Error())
	}
	return c.NoContent(http.StatusOK)
}

// Handle requests to /exception.
func (r *RollupAPI) RegisterException(c echo.Context) error {
	if !checkContentType(c) {
		return c.String(http.StatusUnsupportedMediaType, "invalid content type")
	}

	// parse body
	var request RegisterExceptionJSONRequestBody
	if err := c.Bind(&request); err != nil {
		return err
	}

	// validate fields
	payload, err := hexutil.Decode(request.Payload)
	if err != nil {
		return c.String(http.StatusBadRequest, "invalid hex payload")
	}

	// talk to model
	err = r.model.RegisterException(payload)
	if err != nil {
		return c.String(http.StatusForbidden, err.Error())
	}
	return c.NoContent(http.StatusOK)
}

// Check whether the content type is application/json.
func checkContentType(c echo.Context) bool {
	cType := c.Request().Header.Get(echo.HeaderContentType)
	return strings.HasPrefix(cType, echo.MIMEApplicationJSON)
}

// Convert model input to API type.
func convertInput(input mdl.Input) RollupRequest {
	var resp RollupRequest
	switch input := input.(type) {
	case mdl.AdvanceInput:
		advance := Advance{
			Metadata: Metadata{
				BlockNumber:    input.BlockNumber,
				InputIndex:     uint64(input.Index),
				MsgSender:      hexutil.Encode(input.MsgSender[:]),
				BlockTimestamp: uint64(input.BlockTimestamp.Unix()),
			},
			Payload: hexutil.Encode(input.Payload),
		}
		err := resp.Data.FromAdvance(advance)
		if err != nil {
			panic("failed to convert advance")
		}
		resp.RequestType = AdvanceState
	case mdl.InspectInput:
		inspect := Inspect{
			Payload: hexutil.Encode(input.Payload),
		}
		err := resp.Data.FromInspect(inspect)
		if err != nil {
			panic("failed to convert inspect")
		}
		resp.RequestType = InspectState
	default:
		panic("invalid input from model")
	}
	return resp
}
