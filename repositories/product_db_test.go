package repositories

import (
	"fmt"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// สร้าง DB จำลองด้วย SQLite in-memory แล้ว migrate ตาราง product
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	require.NoError(t, db.AutoMigrate(&product{}))
	return db
}

func seedProducts(t *testing.T, db *gorm.DB, products []product) {
	t.Helper()
	require.NoError(t, db.Create(&products).Error)
}

func TestNewProductRepositoryDB(t *testing.T) {

	t.Run("migrates table and seeds 5000 mock products", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		require.NoError(t, err)

		repo := NewProductRepositoryDB(db)
		require.NotNil(t, repo)

		var count int64
		require.NoError(t, db.Model(&product{}).Count(&count).Error)
		assert.EqualValues(t, 5000, count)
	})

	t.Run("does not reseed when data already exists", func(t *testing.T) {
		db := newTestDB(t)
		seedProducts(t, db, []product{{Name: "existing", Quantity: 1}})

		NewProductRepositoryDB(db)

		var count int64
		require.NoError(t, db.Model(&product{}).Count(&count).Error)
		assert.EqualValues(t, 1, count, "mockData ต้องไม่เติมข้อมูลซ้ำเมื่อมีข้อมูลอยู่แล้ว")
	})
}

func TestProductRepositoryDB_GetProducts(t *testing.T) {

	t.Run("returns products ordered by quantity desc", func(t *testing.T) {
		db := newTestDB(t)
		seedProducts(t, db, []product{
			{Name: "A", Quantity: 5},
			{Name: "B", Quantity: 99},
			{Name: "C", Quantity: 42},
		})
		repo := productRepositoryDB{db: db}

		products, err := repo.GetProducts()

		require.NoError(t, err)
		require.Len(t, products, 3)
		assert.Equal(t, "B", products[0].Name)
		assert.Equal(t, "C", products[1].Name)
		assert.Equal(t, "A", products[2].Name)
	})

	t.Run("limits results to 30 products", func(t *testing.T) {
		db := newTestDB(t)
		many := []product{}
		for i := 0; i < 40; i++ {
			many = append(many, product{Name: fmt.Sprintf("P%d", i), Quantity: i})
		}
		seedProducts(t, db, many)
		repo := productRepositoryDB{db: db}

		products, err := repo.GetProducts()

		require.NoError(t, err)
		assert.Len(t, products, 30)
	})

	t.Run("returns empty when table has no rows", func(t *testing.T) {
		db := newTestDB(t)
		repo := productRepositoryDB{db: db}

		products, err := repo.GetProducts()

		require.NoError(t, err)
		assert.Empty(t, products)
	})
}
