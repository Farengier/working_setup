# Собираем image
docker buildx build -t <имя> .

# Запускаем image
## Просто запуск с последующим удалением
docker run --rm <имя>
## Запуск интерактивно
docker run -it <имя>

# Compose
## Пересоздание
Если надо пересобрать контейнер (изменилась конфигурация, обновился код, etc) - поможет пересоздание контейнера
`docker compose recreate`

## Несколько инстансов одного сервиса
Если надо несколько инстансов, то можно указать в сервисе диапазон портов
```yaml
  # var 1. use range
  ports:
    - "8080-8089:80"
  deploy:
    replicas: 4 # или можно без replicas, но запуская `docker compose up -d --scale web_app=4`

  # var 3. anchoring
  x-base-app: &base-app
    image: ...
    // etc
  services:
    app-1:
      <<: *base-app
      ports: ["8080:80"]
    app-2:
      <<: *base-app
      ports: ["8081:80"]
```