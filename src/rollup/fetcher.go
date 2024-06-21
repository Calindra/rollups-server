package rollup

import (
	"net/http"

	DA "github.com/calindra/rollups-server/src/dataavailability"
	"github.com/labstack/echo/v4"
)

func (r *RollupAPI) Fetcher(ctx echo.Context, request GioJSONRequestBody) (*GioResponseRollup, *DA.HttpCustomError) {
	var (
		espresso uint16 = 2222
		syscoin  uint16 = 5700
		its_ok   uint16 = 42
	)

	switch request.Domain {
	case espresso:
		espressoFetcher := DA.NewEspressoFetcher(r.model.GetInputRepository())
		data, err := espressoFetcher.Fetch(ctx, request.Id)

		if err != nil {
			return nil, err
		}

		return &GioResponseRollup{Data: *data, Code: its_ok}, nil
	case syscoin:
		syscoinFetcher := DA.NewSyscoinClient()
		data, err := syscoinFetcher.Fetch(ctx, request.Id)

		if err != nil {
			return nil, err
		}

		return &GioResponseRollup{Data: *data, Code: its_ok}, nil
	default:
		unsupported := "Unsupported domain"
		return nil, DA.NewHttpCustomError(http.StatusBadRequest, &unsupported)
	}
}
