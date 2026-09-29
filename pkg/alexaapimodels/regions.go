package alexaapimodels

// Region identifies an Alexa service region.
type Region string

const (
	// RegionUS identifies the North America service region.
	RegionUS Region = "us"
	// RegionEU identifies the Europe service region.
	RegionEU Region = "eu"
	// RegionJP identifies the Japan service region.
	RegionJP Region = "jp"
)
