package product

type Product struct {
	ID    uint64  `json:"id"`
	Title string  `json:"title"`
	Price float64 `json:"price"`
	Stock uint64  `json:"stock"`
}
