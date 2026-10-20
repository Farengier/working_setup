По умолчанию кластер main, но можно там же создать еще кластеров
см pg_lsclusters
Каждый кластер берет свой следующий порт

Доступы pg_hba.conf - откда, куда, какие методы аутентификации
Параметры postgresql.conf - миллиард параметров (max_connections, shared_buffers, work_mem, effective_cache_size)

Можно использовать pgtune

\l+ - список баз данных
\с <db> - "подключиться" к БД
\dt - список таблиц
\du - список пользователей

Пул подключений PgBouncer
Мониторинг через `SELECT * FROM pg_stat_activity`
Расширение ps_stat_statements. `CREATE EXTENSION pg_stat_statements;`

Резервное копирование pg_dump, pg_dumpall // базово однопоточный