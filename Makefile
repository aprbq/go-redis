.PHONY: up down run test check logs clean

# Start services ทั้งหมด (PostgreSQL, Redis, InfluxDB, Grafana)
up:
	docker compose up -d postgres redis influxdb grafana

# Stop และลบ containers
down:
	docker compose down

# รัน Go server (ต้อง make up ก่อน)
run:
	go run main.go

# รัน load test ด้วย k6 (server ต้องรันอยู่)
test:
	docker compose run --rm k6 run /scripts/test.js

# ทดสอบยิง endpoint
check:
	curl http://localhost:8000/products

# ดู logs ของ services
logs:
	docker compose logs -f

# ลบ containers พร้อม volume data (PostgreSQL/Redis/InfluxDB/Grafana จะถูกล้าง)
clean:
	docker compose down
	rm -rf data
