package onair

type FBO struct {
	Id                             string  `json:"Id"`
	CompanyId                      string  `json:"CompanyId"`
	WorldId                        string  `json:"WorldId"`
	Airport                        Airport `json:"Airport"`
	AirportId                      string  `json:"AirportId"`
	Name                           string  `json:"Name"`
	CargoWeightCapacity            float64 `json:"CargoWeightCapacity"`
	Fuel100LLCapacity              float64 `json:"Fuel100LLCapacity"`
	FuelJetCapacity                float64 `json:"FuelJetCapacity"`
	SleepingCapacity               float64 `json:"SleepingCapacity"`
	AircraftTieDownCapacity        int     `json:"AircraftTieDownCapacity"`
	AircraftHangarCapacity         int     `json:"AircraftHangarCapacity"`
	Fuel100LLQuantity              float64 `json:"Fuel100LLQuantity"`
	FuelJetQuantity                float64 `json:"FuelJetQuantity"`
	Fuel100LLSellPrice             float64 `json:"Fuel100LLSellPrice"`
	FuelJetSellPrice               float64 `json:"FuelJetSellPrice"`
	AllowFuel100LLSelling          bool    `json:"AllowFuel100LLSelling"`
	AllowFuelJetSelling            bool    `json:"AllowFuelJetSelling"`
	AllowWorkshopSelling           bool    `json:"AllowWorkshopSelling"`
	MarkupWorkshopSelling          float64 `json:"MarkupWorkshopSelling"`
	WorkshopSEP                    bool    `json:"WorkshopSEP"`
	WorkshopMEP                    bool    `json:"WorkshopMEP"`
	WorkshopTurboProp              bool    `json:"WorkshopTurboProp"`
	WorkshopJet                    bool    `json:"WorkshopJet"`
	WorkshopHeavyJet               bool    `json:"WorkshopHeavyJet"`
	WorkshopHelicopter             bool    `json:"WorkshopHelicopter"`
	WorkshopAnnualSEP              bool    `json:"WorkshopAnnualSEP"`
	WorkshopAnnualMEP              bool    `json:"WorkshopAnnualMEP"`
	WorkshopAnnualTurboProp        bool    `json:"WorkshopAnnualTurboProp"`
	WorkshopAnnualJet              bool    `json:"WorkshopAnnualJet"`
	WorkshopAnnualHeavyJet         bool    `json:"WorkshopAnnualHeavyJet"`
	WorkshopAnnualHelicopter       bool    `json:"WorkshopAnnualHelicopter"`
	Fuel100LLOrderedQuantity       float64 `json:"Fuel100LLOrderedQuantity"`
	FuelJetOrderedQuantity         float64 `json:"FuelJetOrderedQuantity"`
	FuelDeliveryDate               string  `json:"FuelDeliveryDate"`
	FuelJetOrderedCapacity         float64 `json:"FuelJetOrderedCapacity"`
	Fuel100LLOrderedCapacity       float64 `json:"Fuel100LLOrderedCapacity"`
	CargoOrderedCapacity           float64 `json:"CargoOrderedCapacity"`
	SleepingOrderedCapacity        float64 `json:"SleepingOrderedCapacity"`
	AircraftHangarOrderedCapacity  int     `json:"AircraftHangarOrderedCapacity"`
	AircraftTieDownOrderedCapacity int     `json:"AircraftTieDownOrderedCapacity"`
	WorkshopUnderConstruction      bool    `json:"WorkshopUnderConstruction"`
	LastWeeklyOwnershipPaymentDate string  `json:"LastWeeklyOwnershipPaymentDate"`
}
