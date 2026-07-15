.PHONY: up stop start down run test check logs clean

# Start services ทั้งหมด (PostgreSQL, Redis, InfluxDB, Grafana)
up:
	@echo "Starting services (postgres, redis, influxdb, grafana)..."
	docker compose up -d postgres redis influxdb grafana

# หยุด services ชั่วคราว (ไม่ลบ containers)
stop:
	@echo "Stopping services..."
	docker compose stop postgres redis influxdb grafana

# start services ที่หยุดไว้
start:
	@echo "Starting stopped services..."
	docker compose start postgres redis influxdb grafana

# Stop และลบ containers
down:
	@echo "Removing containers..."
	docker compose down postgres redis influxdb grafana

# รัน Go server (ต้อง make up ก่อน)
run:
	@echo "Running Go server on :8000..."
	go mod tidy
	go run main.go

# รัน load test ด้วย k6 (server ต้องรันอยู่)
test:
	@echo "Running k6 load test..."
	docker compose run --rm k6 run /scripts/test.js

# ทดสอบยิง endpoint
check:
	@echo "Checking GET /products..."
	curl http://localhost:8000/products

# ดู logs ของ services
logs:
	@echo "Tailing service logs..."
	docker compose logs -f

# ลบ containers พร้อม volume data (PostgreSQL/Redis/InfluxDB/Grafana จะถูกล้าง)
clean:
	@echo "Removing containers and data volumes..."
	docker compose down
	rm -rf data
