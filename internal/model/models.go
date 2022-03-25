package model

import "encoding/json"

type Order struct {
	ID                 string
	Symbol             string
	AccountID          string
	Price              float64
	Status             string
	Side               string
	Time               int64
	Leverage           bool
	TakeProfit         float64
	StopLoss           float64
	GuaranteedStopLoss bool
	Quantity           float64
}

type GeneratedPrice struct {
	Ask    float64 `json:"ask"`
	Bid    float64 `json:"bid"`
	Symbol string  `json:"symbol"`
}

func DecodePrice(data []byte) (GeneratedPrice, error) {
	var msg GeneratedPrice
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return GeneratedPrice{}, err
	}
	return msg, nil
}
