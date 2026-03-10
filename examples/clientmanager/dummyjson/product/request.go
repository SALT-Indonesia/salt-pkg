package product

type Request struct {
	Title string  `json:"title"`
	Price float64 `json:"price"`
	Stock uint64  `json:"stock"`
}
