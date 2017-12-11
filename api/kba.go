package api

import (
	"net/http"

	"io/ioutil"

	"github.com/labstack/echo"
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

	vehicles, err := api.kba.ParseVehiclesPDF(data)
	if err != nil {
		response := &Response{Message: err.Error()}
		return ctx.JSON(http.StatusInternalServerError, response)
	}
	return ctx.JSON(http.StatusOK, &ListResponse{Total: len(vehicles), Items: vehicles})
}
