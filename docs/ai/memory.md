# Запускаем векторную базу
```bash
# запускаем в detached формате с указанием volume для бд чтобы сохранить состояние
docker run -d --name qdrant -p 6333:6333 -p 6334:6334 -v qdrant_storage:/qdrant/storage qdrant/qdrant

# Можно проверить volume командами
docker inspect qdrant | grep -A 10 "Mounts"
docker volume ls
```


При крнфигурировании zoo code (vscode ai agent) можно указать http://localhost:6333 как адрес БД quadrant