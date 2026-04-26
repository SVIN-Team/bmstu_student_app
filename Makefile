.PHONY: up down build-front build-back build-all fill-db clean

# Запустить все сервисы (бэкенд, фронтенд, базы данных) в фоновом режиме
up:
	docker compose up -d

# Остановить все сервисы
down:
	docker compose down

# Собрать образ только для фронтенда
build-front:
	docker compose build frontend

# Собрать образ только для бэкенда
build-back:
	docker compose build app

run-front:
	docker compose up -d frontend

run-back:
	docker compose up -d app

# Собрать образы для всех сервисов
build-all:
	docker compose build

# Заполнить базу данных тестовыми данными (используя виртуальное окружение Python локально)
# Требует запущенной базы данных (make up)
fill-db:
	cd backend/deploy/scripts && \
	python3 -m venv .venv && \
	. .venv/bin/activate && \
	pip install -r requirements.txt && \
	python3 fill_db.py

# Полная очистка: остановить контейнеры и удалить все данные (включая volume с базой данных)
clean:
	docker compose down -v

