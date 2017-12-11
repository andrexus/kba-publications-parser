package service

import "bytes"

type Vehicle struct {
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
	ParseVehiclesPDF(data []byte) ([]Vehicle, error)
}

type KBAServiceImpl struct {
}

func (c *KBAServiceImpl) ParseVehiclesPDF(data []byte) ([]Vehicle, error) {
	return parseVehicles(bytes.NewReader(data))
}
