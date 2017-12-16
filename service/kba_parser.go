package service

import (
	"fmt"

	"math"

	"log"

	"io"

	"strconv"

	pdfcontent "github.com/unidoc/unidoc/pdf/contentstream"
	pdfcore "github.com/unidoc/unidoc/pdf/core"
	pdf "github.com/unidoc/unidoc/pdf/model"
	"unicode/utf8"
)

const (
	minHeight     = 134
	maxHeight     = 575
	lineDeviation = 2
)

var (
	columns                     = []int{35, 67, 95, 220, 337, 460, 544, 598, 620, 652, 676, 712, 736, 767, 790}
	manufacturerCodeNumber      = columns[0]
	typeCodeNumber              = columns[1]
	manufacturerPlaintext       = columns[2]
	manufacturerTradeName       = columns[3]
	commercialName              = columns[4]
	typeCodeNumberAllotmentDate = columns[5]
	vehicleCategory             = columns[6]
	bodyworkCode                = columns[7]
	fuelCode                    = columns[8]
	maxNetPower                 = columns[9]
	engineCapacity              = columns[10]
	maxAxles                    = columns[11]
	maxPoweredAxles             = columns[12]
	maxSeats                    = columns[13]
	maxPermissibleMass          = columns[14]
)

func parseVehiclesCategoryM(rs io.ReadSeeker) ([]VehicleCategoryM, error) {
	pdfReader, err := pdf.NewPdfReader(rs)
	if err != nil {
		return nil, err
	}

	isEncrypted, err := pdfReader.IsEncrypted()
	if err != nil {
		return nil, err
	}

	if isEncrypted {
		_, err = pdfReader.Decrypt([]byte(""))
		if err != nil {
			return nil, err
		}
	}

	numPages, err := pdfReader.GetNumPages()
	vehicles := make([]VehicleCategoryM, 0)
	if err != nil {
		return nil, err
	}

	for i := 0; i < numPages; i++ {
		pageNum := i + 1

		page, err := pdfReader.GetPage(pageNum)
		if err != nil {
			return nil, err
		}

		tableElements, err := parsePage(page)
		if err != nil {
			return nil, err
		}
		vehicles = append(vehicles, parseTableElements(tableElements)...)
	}

	return vehicles, nil
}

type TableElement struct {
	X    int
	Y    int
	Text string
}

func parsePage(page *pdf.PdfPage) ([]*TableElement, error) {
	pageContentStr, err := page.GetAllContentStreams()
	if err != nil {
		return nil, err
	}

	cstreamParser := pdfcontent.NewContentStreamParser(pageContentStr)
	if err != nil {
		return nil, err
	}

	operations, err := cstreamParser.Parse()
	if err != nil {
		return nil, err
	}

	result := make([]*TableElement, 0)

	parsing := false
	x := 0.0
	y := 0.0

	for _, op := range *operations {
		if op.Operand == "Td" && len(op.Params) == 2 {
			if v, ok := op.Params[0].(*pdfcore.PdfObjectFloat); ok {
				x = float64(*v)
			}
			if v, ok := op.Params[1].(*pdfcore.PdfObjectFloat); ok {
				y = float64(*v)
				if y < minHeight || y > maxHeight {
					parsing = false
					continue
				} else {
					parsing = true
				}
			}
		}

		if parsing && op.Operand == "TJ" {
			if len(op.Params) < 1 {
				continue
			}
			paramList, ok := op.Params[0].(*pdfcore.PdfObjectArray)
			if !ok {
				return nil, fmt.Errorf("invalid parameter type, no array (%T)", op.Params[0])
			}
			txt := ""
			for _, obj := range *paramList {
				if strObj, ok := obj.(*pdfcore.PdfObjectString); ok {
					txt += string(*strObj)
				}
			}
			result = append(result, &TableElement{X: int(math.Ceil(x)), Y: int(math.Ceil(y)), Text: decodeString([]byte(txt))})
		}
	}
	return result, nil
}

func parseTableElements(tableElements []*TableElement) []VehicleCategoryM {
	result := make([]VehicleCategoryM, 0)
	row := make([]*TableElement, 0)

	if len(tableElements) < 2 {
		return nil
	}

	for i := 0; i < len(tableElements); i++ {
		element := tableElements[i]
		row = append(row, element)
		if (i == len(tableElements)-1) || (element.Y < tableElements[i+1].Y-lineDeviation || element.Y > tableElements[i+1].Y+lineDeviation) {
			vehicle := parseRow(row)
			result = append(result, *vehicle)
			row = make([]*TableElement, 0)
		}
	}

	return result
}

func parseRow(row []*TableElement) *VehicleCategoryM {
	vehicle := &VehicleCategoryM{}
	for _, tableElement := range row {
		closestPosition := closest(tableElement.X, columns)
		switch closestPosition {
		case manufacturerCodeNumber:
			vehicle.ManufacturerCodeNumber = &tableElement.Text
		case typeCodeNumber:
			vehicle.TypeCodeNumber = &tableElement.Text
		case manufacturerPlaintext:
			vehicle.ManufacturerPlaintext = &tableElement.Text
		case manufacturerTradeName:
			vehicle.ManufacturerTradeName = &tableElement.Text
		case commercialName:
			vehicle.CommercialName = &tableElement.Text
		case typeCodeNumberAllotmentDate:
			vehicle.TypeCodeNumberAllotmentDate = &tableElement.Text
		case vehicleCategory:
			vehicle.VehicleCategory = &tableElement.Text
		case bodyworkCode:
			vehicle.BodyworkCode = &tableElement.Text
		case fuelCode:
			vehicle.FuelCode = &tableElement.Text
		case maxNetPower:
			if v, err := strconv.Atoi(tableElement.Text); err == nil {
				vehicle.MaxNetPower = &v
			}
		case engineCapacity:
			if v, err := strconv.Atoi(tableElement.Text); err == nil {
				vehicle.EngineCapacity = &v
			}
		case maxAxles:
			if v, err := strconv.Atoi(tableElement.Text); err == nil {
				vehicle.MaxAxles = &v
			}
		case maxPoweredAxles:
			if v, err := strconv.Atoi(tableElement.Text); err == nil {
				vehicle.MaxPoweredAxles = &v
			}
		case maxSeats:
			if v, err := strconv.Atoi(tableElement.Text); err == nil {
				vehicle.MaxSeats = &v
			}
		case maxPermissibleMass:
			if v, err := strconv.Atoi(tableElement.Text); err == nil {
				vehicle.MaxPermissibleMass = &v
			}

		default:
			log.Printf("Unknown position: %d", closestPosition)
		}
	}
	return vehicle
}

func closest(num int, arr []int) int {
	curr := arr[0]
	for _, val := range arr {
		if math.Abs(float64(num-val)) < math.Abs(float64(num-curr)) {
			curr = val
		}
	}
	return curr
}

func decodeString(input []byte) string {
	var result string
	for i := 0; i < utf8.RuneCount(input); i++ {
		r := rune(input[i])
		buf := make([]byte, utf8.RuneLen(r))
		utf8.EncodeRune(buf, r)
		result += string(buf)
	}
	return result
}
