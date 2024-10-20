package data

type FeatureCollection struct {
	Type     string `json:"type"`
	Features []struct {
		Type     string `json:"type"`
		Geometry struct {
			Type        string    `json:"type"`
			Coordinates []float64 `json:"coordinates"`
		} `json:"geometry"`
		Properties struct {
			Name        string `json:"name"`
			FullAddress string `json:"full_address"`
		} `json:"properties"`
	} `json:"features"`
}
