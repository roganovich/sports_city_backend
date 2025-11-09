package models

type AddressRequest struct {
	Query string `json:"query"`
}

type Geo struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

type AddressSuggestion struct {
	Value string `json:"value"`
	Geo   Geo    `json:"geo"`
}

type AddressResponse struct {
	Suggestions []AddressSuggestion `json:"suggestions"`
}
