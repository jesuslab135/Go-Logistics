package dto

// FacetCount is one bucket of a grouped count: the raw grouping value, its
// human label (equal to Value when the dimension is its own label), and how
// many rows in the filtered set fall into it.
type FacetCount struct {
	Value string `json:"value"`
	Label string `json:"label"`
	Count int64  `json:"count"`
}

type WorkOrderFacetsResponse struct {
	Status []FacetCount `json:"status"`
}

type IssueFacetsResponse struct {
	State []FacetCount `json:"state"`
}

type AssetFacetsResponse struct {
	VehicleType []FacetCount `json:"vehicle_type"`
	TrailerType []FacetCount `json:"trailer_type"`
	Status      []FacetCount `json:"status"`
}
