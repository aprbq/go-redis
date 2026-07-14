package repositories

import (
	"gorm.io/gorm"
)

type productRepositoryDB struct {
	db *gorm.DB
}

// constructor
func NewProductRepositoryDB(db *gorm.DB) ProductRepository {
	// return productRepositoryDB{db} //อันไหนก็ได้
	db.AutoMigrate(&product{})
	mockData(db)
	return productRepositoryDB{db: db}
}

func (r productRepositoryDB) GetProducts() (products []product, err error) {

	err = r.db.Order("quantity desc").Limit(30).Find(&products).Error
	return products, err
}
