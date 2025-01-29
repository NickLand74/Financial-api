.PHONY: build run stop test

build:
	# Собираем образы с помощью docker-compose
	docker-compose build

run:
	# Поднимаем контейнеры в фоне
	# --build тоже пересобирает образы при каждом запуске
	docker-compose up --build -d
	# Если у вас есть миграции через goose или другой инструмент,
	# можно добавить команду здесь, например:
	# docker-compose run --rm app migrate

stop:
	# Останавливаем и удаляем контейнеры
	docker-compose down

test:
	# Запускаем go-тесты локально (не в Docker)
	go test -v ./...