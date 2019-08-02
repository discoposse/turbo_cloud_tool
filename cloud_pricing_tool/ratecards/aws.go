package aws

import (
	"time"
)

type ProductAttributes struct {
	ServiceCode  string `json:"servicecode"`
	Location     string `json:"location"`
	LocationType string `json:"locationType"`
	InstanceType string `json:"instanceType"`
	// TODO: Need a custom unmarshaller for this `json:"currentGeneration"`
	CurrentGeneration bool   `json:"foo,omitempty"`
	InstanceFamily    string `json:"instanceFamily"`
	// TODO: Need a custom unmarshaller to handle non int VCPUs and convert them
	VCPU              string `json:"vcpu"`
	PhysicalProcessor string `json:"physicalProcessor"`
	// TODO: In a custom unmarshaller, convert this to an int, containing KB of mem
	Memory                string `json:"memory"`
	Storage               string `json:"storage"`
	NetworkPerformance    string `json:"networkPerformance"`
	ProcessorArchitecture string `json:"processorArchitecture"`
	Tenancy               string `json:"tenancy"`
	OperatingSystem       string `json:"operatingSystem"`
	LicenseModel          string `json:"licenseModel"`
	UsageType             string `json:"usagetype"`
	Operation             string `json:"operation"`
	CapacityStatus        string `json:"capacitystatus"`
	ECU                   string `json:"ecu"`
	InstanceSKU           string `json:"instancesku"`
	// TODO: Need a custom unmarshaller to handle non int and convert them
	NormalizationSizeFactor string `json:"normalizationSizeFactor"`
	PreInstalledSw          string `json:"preInstalledSw"`
	ServiceName             string `json:"serviceName"`
}

type Product struct {
	SKU           string            `json:"sku"`
	ProductFamily string            `json:"productFamily"`
	Attributes    ProductAttributes `json:"attributes"`
}

type TermAttributes struct {
	LeaseContractLength string
	OfferingClass       string
	PurchaseOption      string
}

type PriceDimension struct {
	RateCode    string `json:"rateCode"`
	Description string `json:"description"`
	// TODO: These can be a string "Inf", in many cases, need a custom unmarshaller, but
	// I'm not using these values for now, so I'm gonna skip it.
	// BeginRange     int                // `json:"beginRange"`
	// EndRange       int                // `json:"endRange"`
	Unit string `json:"unit"`
	// TODO: Another which needs custom unmarshaling for mixed numeric and string vals
	PricePerUnit map[string]string `json:"pricePerUnit"`
	AppliesTo    []string          `json:"appliesTo"`
}

type Terms struct {
	OnDemand map[string]map[string]OfferTerm `json:"OnDemand"`
	Reserved map[string]map[string]OfferTerm `json:"Reserved"`
}

// type InstanceTerm struct {
// 	OfferTerm map[string]OfferTerm
// }

type OfferTerm struct {
	OfferTermCode   string                    `json:"offerTermCode"`
	SKU             string                    `json:"sku"`
	EffectiveDate   time.Time                 `json:"effectiveDate"`
	PriceDimensions map[string]PriceDimension `json:"priceDimensions"`
	TermAttributes  TermAttributes            `json:"termAttributes"`
}

type Root struct {
	Products map[string]Product `json:"products"`
	Terms    Terms              `json:"terms"`
}
