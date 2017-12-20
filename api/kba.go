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

	parseResults, err := api.kba.ParsePDF(data)
	//parseResults, err := api.kba.ParseTaxonomyDirectory(data)
	if err != nil {
		response := &MessageResponse{Message: err.Error()}
		return ctx.JSON(http.StatusInternalServerError, response)
	}

	resp := &service.ParseResponse{
		Success: true,
		Data: parseResults,
	}
	return ctx.JSON(http.StatusOK, resp)
}
