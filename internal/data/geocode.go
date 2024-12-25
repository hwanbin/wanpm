package data

type FeatureCollection struct {
	Type     string    `json:"type"`
	Features []Feature `json:"features"`
}

type Feature struct {
	Type     string `json:"type"`
	Geometry struct {
		Type        string    `json:"type"`
		Coordinates []float64 `json:"coordinates"`
	} `json:"geometry"`
	Properties struct {
		Name        string `json:"name"`
		FullAddress string `json:"full_address"`
	} `json:"properties"`
}

func (f *Feature) IsValidFeature() bool {
	return f.Type == "Feature" &&
		f.Geometry.Type == "Point" &&
		len(f.Geometry.Coordinates) == 2 &&
		f.Properties.Name != "" &&
		f.Properties.FullAddress != ""
}

func (f *Feature) IsEmptyFeature() bool {
	return f.Type == "" &&
		f.Geometry.Type == "" &&
		f.Geometry.Coordinates == nil &&
		f.Properties.Name == "" &&
		f.Properties.FullAddress == ""
}

func (f *Feature) IsValidFeatureType() bool {
	return f.Type == "Feature"
}

func (f *Feature) IsValidPointGeometryType() bool {
	return f.Geometry.Type == "Point"
}

func (f *Feature) IsValidPointGeometryCoordinates() bool {
	return len(f.Geometry.Coordinates) == 2
}

func (f *Feature) IsValidPropertiesName() bool {
	return f.Properties.Name != ""
}

func (f *Feature) IsValidPropertiesFullAddress() bool {
	return f.Properties.FullAddress != ""
}

func (f *Feature) IsValidLongitude() bool {
	return f.Geometry.Coordinates[0] >= -180 && f.Geometry.Coordinates[0] <= 180
}

func (f *Feature) IsValidLatitude() bool {
	return f.Geometry.Coordinates[1] >= -90 && f.Geometry.Coordinates[1] <= 90
}
