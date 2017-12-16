package api

import (
	"net/http"

	"io/ioutil"

	"github.com/labstack/echo"
	"github.com/andrexus/kba-publications-parser/service"
)

func (api *API) UploadPDF(ctx echo.Context) error {
	file, err := ctx.FormFile("file")
	if err != nil {
		return err
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	data, err := ioutil.ReadAll(src)
	if err != nil {
		return err
	}

	parseResult, err := api.kba.ParseVehiclesCategoryM(data)
	if err != nil {
		response := &MessageResponse{Message: err.Error()}
		return ctx.JSON(http.StatusInternalServerError, response)
	}

	var parsedResults []*service.ParseResult
	parsedResults = append(parsedResults, parseResult)
	resp := &service.ParseResponse{
		Success: true,
		Data: parsedResults,
	}
	return ctx.JSON(http.StatusOK, resp)
}
