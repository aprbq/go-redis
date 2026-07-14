package main

import (
	"goredis/handlers"
	"goredis/repositories"
	"goredis/services"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {

	db := initDatabase()
	redisClient := initRedis()
	// _ = redisClient

	productRepo := repositories.NewProductRepositoryDB(db) //ไม่มี redis
	// productRepo := repositories.NewProductRepositoryRedis(db, redisClient) //มี redis

	productservice := services.NewCatalogService(productRepo)
	// productservice := services.NewCatalogServiceRedis(productRepo, redisClient)

	// productHandler := handlers.NewCatalogHandler(productservice)
	productHandler := handlers.NewCatalogHandlerRedis(productservice, redisClient)

	app := fiber.New()

	app.Get("/products", productHandler.GetProducts)

	app.Listen(":8000")
	// products, err := productservice.GetProducts()
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// fmt.Println(products)
}

func initDatabase() *gorm.DB {
	dial := postgres.Open("host=localhost port=5433 user=postgres password=kook0990 dbname=infinitas sslmode=disable")

	db, err := gorm.Open(dial, &gorm.Config{})
	if err != nil {
		panic(err)
	}

	return db
}

func initRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}
