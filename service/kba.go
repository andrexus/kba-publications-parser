package service

import "bytes"

type ParseResponse struct {
	Success bool          `json:"success"`
	Data    []*ParseResult `json:"data"`
}

type ParseResult struct {
	Total      int         `json:"total"`
	EntityType string      `json:"entityType"`
	Items      interface{} `json:"items"`
}

type VehicleCategoryM struct {
	ManufacturerCodeNumber      *string `json:"manufacturerCodeNumber"`
	TypeCodeNumber              *string `json:"typeCodeNumber"`
	ManufacturerPlaintext       *string `json:"manufacturerPlaintext"`
	ManufacturerTradeName       *string `json:"manufacturerTradeName"`
	CommercialName              *string `json:"commercialName"`
	TypeCodeNumberAllotmentDate *string `json:"typeCodeNumberAllotmentDate"`
	VehicleCategory             *string `json:"vehicleCategory"`
	BodyworkCode                *string `json:"bodyworkCode"`
	FuelCode                    *string `json:"fuelCode"`
	MaxNetPower                 *int    `json:"maxNetPower"`
	EngineCapacity              *int    `json:"engineCapacity"`
	MaxAxles                    *int    `json:"maxAxles"`
	MaxPoweredAxles             *int    `json:"maxPoweredAxles"`
	MaxSeats                    *int    `json:"maxSeats"`
	MaxPermissibleMass          *int    `json:"maxPermissibleMass"`
}

type KBAService interface {
	ParseVehiclesCategoryM(data []byte) (*ParseResult, error)
}

type KBAServiceImpl struct {
}

func (c *KBAServiceImpl) ParseVehiclesCategoryM(data []byte) (*ParseResult, error) {
	items, err := parseVehiclesCategoryM(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	result := &ParseResult{
		Total:      len(items),
		EntityType: "VehicleCategoryM",
		Items:      items,
	}
	return result, nil
}
