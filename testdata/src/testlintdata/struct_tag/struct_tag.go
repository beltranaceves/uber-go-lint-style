package struct_tag

import "encoding/json"

type Stock struct {
	Price int    // want "exported field 'Price' should have a tag for marshaling"
	Name  string // want "exported field 'Name' should have a tag for marshaling"
}

type StockGood struct {
	Price int    `json:"price"`
	Name  string `json:"name"`
}

func ExampleBad() {
	_, _ = json.Marshal(Stock{
		Price: 137,
		Name:  "UBER",
	})
}

func ExampleGood() {
	_, _ = json.Marshal(StockGood{
		Price: 137,
		Name:  "UBER",
	})
}
