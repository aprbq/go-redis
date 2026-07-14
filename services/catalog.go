package services

type Product struct {
	ID       int
	Name     string
	Quantity int
}

type CatalogService interface {
	GetProducts() ([]Product, error)
}
