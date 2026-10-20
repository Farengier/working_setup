# Собираем image
docker buildx build -t <имя> .

# Запускаем image
## Просто запуск с последующим удалением
docker run --rm <имя>
## Запуск интерактивно
docker run -it <имя>
