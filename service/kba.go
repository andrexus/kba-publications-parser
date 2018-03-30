package service

import "bytes"

type ParseResponse struct {
	Success bool           `json:"success"`
	Data    []*ParseResult `json:"data"`
}

type ParseResult struct {
	Total      int         `json:"total"`
	EntityType string      `json:"entityType"`
	Items      interface{} `json:"items"`
}

type EnergySource struct {
	ShortName string `json:"shortName"`
	Code      string `json:"code"`
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
	ParsePDF(data []byte) ([]*ParseResult, error)

	// SV 1 - Verzeichnis zur Systematisierung von Kraftfahrzeugen und ihren Anhängern
	ParseTaxonomyDirectory(data []byte) ([]*ParseResult, error)

	// SV 4.2 - Verzeichnis der Hersteller und Typen der für die Personenbeförderung ausgelegten und gebauten
	// Kraftfahrzeuge mit mindestens vier Rädern (Klasse M)
	ParseVehiclesCategoryM(data []byte) (*ParseResult, error)
}

type KBAServiceImpl struct {
}

func (c *KBAServiceImpl) ParsePDF(data []byte) ([]*ParseResult, error) {
	var results = make([]*ParseResult, 0)

	taxonomyResults, err := c.ParseTaxonomyDirectory(data)
	if err != nil {
		return nil, err
	}
	results = append(results, taxonomyResults...)

	return results, nil
}

func (c *KBAServiceImpl) ParseTaxonomyDirectory(data []byte) ([]*ParseResult, error) {
	energySources, err := parseEnergySources(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	var results = make([]*ParseResult, 0)
	result := &ParseResult{
		Total:      len(energySources),
		EntityType: "EnergySource",
		Items:      energySources,
	}
	results = append(results, result)
	return results, nil
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
