package repositories

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/glebarez/sqlite"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// สร้าง Redis จำลองด้วย miniredis
func newTestRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()

	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return mr, client
}

func TestNewProductRepositoryRedis(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	_, client := newTestRedis(t)

	repo := NewProductRepositoryRedis(db, client)
	require.NotNil(t, repo)

	var count int64
	require.NoError(t, db.Model(&product{}).Count(&count).Error)
	assert.EqualValues(t, 5000, count, "constructor ต้อง migrate และ seed mock data")
}

func TestProductRepositoryRedis_GetProducts(t *testing.T) {

	key := "repository::GetProducts"

	t.Run("cache miss: queries db then sets cache with ttl", func(t *testing.T) {
		db := newTestDB(t)
		seedProducts(t, db, []product{{Name: "A", Quantity: 7}})
		mr, client := newTestRedis(t)
		repo := productRepositoryRedis{db: db, redisClient: client}

		products, err := repo.GetProducts()

		require.NoError(t, err)
		require.Len(t, products, 1)
		assert.Equal(t, "A", products[0].Name)

		// ต้องมี cache ถูกเซตไว้พร้อม TTL 10 วินาที
		cached, err := mr.Get(key)
		require.NoError(t, err)
		assert.Contains(t, cached, `"Name":"A"`)
		assert.Equal(t, 10*time.Second, mr.TTL(key))
	})

	t.Run("cache hit: returns cached data without touching db", func(t *testing.T) {
		db := newTestDB(t) // DB ว่าง — ถ้าคำตอบมาจาก DB จะได้ slice เปล่า
		mr, client := newTestRedis(t)

		cachedProducts := []product{{ID: 1, Name: "FromCache", Quantity: 42}}
		data, err := json.Marshal(cachedProducts)
		require.NoError(t, err)
		require.NoError(t, mr.Set(key, string(data)))

		repo := productRepositoryRedis{db: db, redisClient: client}

		products, err := repo.GetProducts()

		require.NoError(t, err)
		require.Len(t, products, 1)
		assert.Equal(t, "FromCache", products[0].Name)
	})

	t.Run("corrupted cache: falls back to db", func(t *testing.T) {
		db := newTestDB(t)
		seedProducts(t, db, []product{{Name: "B", Quantity: 3}})
		mr, client := newTestRedis(t)
		require.NoError(t, mr.Set(key, "not-valid-json"))

		repo := productRepositoryRedis{db: db, redisClient: client}

		products, err := repo.GetProducts()

		require.NoError(t, err)
		require.Len(t, products, 1)
		assert.Equal(t, "B", products[0].Name)
	})

	t.Run("db error: returns error when table is missing", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		}) // ไม่ migrate — ให้ query พังเพราะไม่มีตาราง
		require.NoError(t, err)
		_, client := newTestRedis(t)
		repo := productRepositoryRedis{db: db, redisClient: client}

		products, err := repo.GetProducts()

		assert.Error(t, err)
		assert.Nil(t, products)
	})

	t.Run("redis down: still serves from db", func(t *testing.T) {
		db := newTestDB(t)
		seedProducts(t, db, []product{{Name: "C", Quantity: 1}})
		mr := miniredis.RunT(t)
		client := redis.NewClient(&redis.Options{
			Addr:        mr.Addr(),
			MaxRetries:  -1, // ปิด retry ให้เทสต์จบเร็ว
			DialTimeout: 200 * time.Millisecond,
		})
		mr.Close() // จำลอง redis ล่ม

		repo := productRepositoryRedis{db: db, redisClient: client}

		products, err := repo.GetProducts()

		// การเขียน cache เป็น best-effort — redis ล่มต้องยังตอบข้อมูลจาก DB ได้
		require.NoError(t, err)
		require.Len(t, products, 1)
		assert.Equal(t, "C", products[0].Name)
	})
}
