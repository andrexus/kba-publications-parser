package service

import (
	"fmt"
	"io"
	"log"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/Sirupsen/logrus"
	pdfcontent "github.com/unidoc/unidoc/pdf/contentstream"
	pdfcore "github.com/unidoc/unidoc/pdf/core"
	pdf "github.com/unidoc/unidoc/pdf/model"
)

const (
	minHeightVehiclePage      = 134
	maxHeightVehiclePage      = 575
	minHeightEnergySourcePage = 66
	maxHeightEnergySourcePage = 690
	lineDeviation             = 2
	prefixEnergySource        = "Kraftstoffart bzw. Energiequelle"
)

var (
	columnsVehicle              = []int{35, 67, 95, 220, 337, 460, 544, 598, 620, 652, 676, 712, 736, 767, 790}
	manufacturerCodeNumber      = columnsVehicle[0]
	typeCodeNumber              = columnsVehicle[1]
	manufacturerPlaintext       = columnsVehicle[2]
	manufacturerTradeName       = columnsVehicle[3]
	commercialName              = columnsVehicle[4]
	typeCodeNumberAllotmentDate = columnsVehicle[5]
	vehicleCategory             = columnsVehicle[6]
	bodyworkCode                = columnsVehicle[7]
	fuelCode                    = columnsVehicle[8]
	maxNetPower                 = columnsVehicle[9]
	engineCapacity              = columnsVehicle[10]
	maxAxles                    = columnsVehicle[11]
	maxPoweredAxles             = columnsVehicle[12]
	maxSeats                    = columnsVehicle[13]
	maxPermissibleMass          = columnsVehicle[14]

	columnsEnergySource = []int{191, 337}
	esShortName         = columnsEnergySource[0]
	esCode              = columnsEnergySource[1]
)

func parseEnergySources(rs io.ReadSeeker) ([]EnergySource, error) {
	pdfReader, err := pdf.NewPdfReader(rs)
	if err != nil {
		return nil, err
	}

	numPages, err := pdfReader.GetNumPages()
	items := make([]EnergySource, 0)
	if err != nil {
		return nil, err
	}

	for i := 0; i < numPages; i++ {
		pageNum := i + 1

		page, err := pdfReader.GetPage(pageNum)
		if err != nil {
			return nil, err
		}

		text, err := parseText(page)
		if err != nil {
			logrus.Error(err.Error())
			continue
		}

		if !strings.HasPrefix(text, prefixEnergySource) {
			continue
		}
		fmt.Println("Page: ", i+1)
		tableElements, err := parseEnergySourcePage(page)
		if err != nil {
			return nil, err
		}
		//fmt.Println("Found table elements: ", len(tableElements))
		//for _, elem := range tableElements {
		//	fmt.Printf("%d, %d : %s\n", elem.X, elem.Y, elem.Text)
		//}
		items = append(items, parseEnergySourceTableElements(tableElements)...)
	}

	for _, item := range items {
		fmt.Printf("%s - %s\n", item.Code, item.ShortName)
	}

	return items, nil
}

func parseText(page *pdf.PdfPage) (string, error) {
	pageContentStr, err := page.GetAllContentStreams()
	if err != nil {
		return "", err
	}

	cstreamParser := pdfcontent.NewContentStreamParser(pageContentStr)
	if err != nil {
		return "", err
	}
	return cstreamParser.ExtractText()
}

func parseVehiclesCategoryM(rs io.ReadSeeker) ([]VehicleCategoryM, error) {
	pdfReader, err := pdf.NewPdfReader(rs)
	if err != nil {
		return nil, err
	}

	numPages, err := pdfReader.GetNumPages()
	items := make([]VehicleCategoryM, 0)
	if err != nil {
		return nil, err
	}

	for i := 0; i < numPages; i++ {
		pageNum := i + 1

		page, err := pdfReader.GetPage(pageNum)
		if err != nil {
			return nil, err
		}

		tableElements, err := parseVehiclePage(page)
		if err != nil {
			return nil, err
		}
		items = append(items, parseVehicleTableElements(tableElements)...)
	}

	return items, nil
}

type TableElement struct {
	X    int
	Y    int
	Text string
}

func parseEnergySourcePage(page *pdf.PdfPage) ([]*TableElement, error) {
	pageContentStr, err := page.GetAllContentStreams()
	fmt.Println(pageContentStr)
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

	parsing := true
	x := 0.0
	y := 0.0

	txt := ""
	for _, op := range *operations {
		if op.Operand == "BT" {
			txt = ""
		} else if op.Operand == "ET" {
			if len(txt) > 0 {
				fmt.Printf("%f, %f: %s\n", x, y, txt)
			}
		}
		if op.Operand == "Tm" && len(op.Params) == 6 {
			if v, ok := op.Params[4].(*pdfcore.PdfObjectFloat); ok {
				x = float64(*v)
			}
			if v, ok := op.Params[5].(*pdfcore.PdfObjectFloat); ok {
				y = float64(*v)
			}
		}
		if y < minHeightEnergySourcePage || y > maxHeightEnergySourcePage {
			parsing = false
			continue
		} else {
			parsing = true
		}

		if parsing {
			if op.Operand == "TJ" {
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
				if len(txt) > 0 && txt != " " {
					result = append(result, &TableElement{X: int(math.Ceil(x)), Y: int(math.Ceil(y)), Text: decodeString([]byte(txt))})
				}
			} else if op.Operand == "Tj" {
				if len(op.Params) < 1 {
					continue
				}
				param, ok := op.Params[0].(*pdfcore.PdfObjectString)
				if !ok {
					return nil, fmt.Errorf("invalid parameter type, no array (%T)", op.Params[0])
				}
				txt := string(*param)
				if len(txt) > 0 && txt != " " {
					result = append(result, &TableElement{X: int(math.Ceil(x)), Y: int(math.Ceil(y)), Text: decodeString([]byte(txt))})
				}
			}
		}
	}

	return result, nil
}

func parseVehiclePage(page *pdf.PdfPage) ([]*TableElement, error) {
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
				if y < minHeightVehiclePage || y > maxHeightVehiclePage {
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

func parseEnergySourceTableElements(tableElements []*TableElement) []EnergySource {
	result := make([]EnergySource, 0)
	row := make([]*TableElement, 0)

	if len(tableElements) < 2 {
		return nil
	}

	for i := 0; i < len(tableElements); i++ {
		element := tableElements[i]
		row = append(row, element)
		if (i == len(tableElements)-1) || (element.Y < tableElements[i+1].Y-lineDeviation || element.Y > tableElements[i+1].Y+lineDeviation) {
			item := parseEnergySourceRow(row)
			result = append(result, *item)
			row = make([]*TableElement, 0)
		}
	}

	return result
}

func parseVehicleTableElements(tableElements []*TableElement) []VehicleCategoryM {
	result := make([]VehicleCategoryM, 0)
	row := make([]*TableElement, 0)

	if len(tableElements) < 4 {
		return nil
	}

	for i := 0; i < len(tableElements); i++ {
		element := tableElements[i]
		row = append(row, element)
		if (i == len(tableElements)-1) || (element.Y < tableElements[i+1].Y-lineDeviation || element.Y > tableElements[i+1].Y+lineDeviation) {
			item := parseVehicleRow(row)
			result = append(result, *item)
			row = make([]*TableElement, 0)
		}
	}

	return result
}

func parseEnergySourceRow(row []*TableElement) *EnergySource {
	item := &EnergySource{}
	for _, tableElement := range row {
		closestPosition := closest(tableElement.X, columnsEnergySource)
		switch closestPosition {
		case esShortName:
			item.ShortName = tableElement.Text
		case esCode:
			item.Code = tableElement.Text

		default:
			log.Printf("Unknown position: %d", closestPosition)
		}
	}
	return item
}

func parseVehicleRow(row []*TableElement) *VehicleCategoryM {
	vehicle := &VehicleCategoryM{}
	for _, tableElement := range row {
		closestPosition := closest(tableElement.X, columnsVehicle)
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
