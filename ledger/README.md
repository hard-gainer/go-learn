# Ledger

Ledger хранит бюджеты и транзакции в PostgreSQL. Сводки по категориям кешируются в Redis на 30 секунд. Пул PostgreSQL допускает до 10 открытых и до 5 простаивающих соединений.

## База данных и миграции

Запустите PostgreSQL и Redis. Для локального запуска подойдут контейнеры ниже. PostgreSQL создаст базу `cashapp` при первом запуске контейнера.

```bash
docker run -d --name cashcraft-postgres -p 127.0.0.1:5432:5432 \
  -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=cashapp postgres:17-alpine
docker run -d --name cashcraft-redis -p 127.0.0.1:6379:6379 redis:7-alpine
```

Если PostgreSQL уже запущен локально, создайте базу командой `createdb -h localhost -U postgres cashapp`. Из корня репозитория установите Goose и, когда PostgreSQL готов принимать соединения, примените миграции.

```bash
GOBIN="$PWD/ledger/.bin" go install github.com/pressly/goose/v3/cmd/goose@v3.30.1
export DATABASE_URL='postgres://postgres:postgres@localhost:5432/cashapp?sslmode=disable'
make migrate-up
```

Для отката используйте `make migrate-down`. Откат начальной миграции удаляет обе таблицы вместе с данными.

## Запуск и проверка отчёта

```bash
export REDIS_ADDR=localhost:6379
export REDIS_DB=0
export REDIS_PASSWORD=''
go -C ledger run . -listen 127.0.0.1:8081
```

Пустой `REDIS_PASSWORD` подходит для Redis без аутентификации. С включённой аутентификацией задайте пароль вашего Redis. Ledger логирует успешное подключение к PostgreSQL и Redis.

В другом терминале выполните два запроса подряд. Для диапазона, которого ещё нет в кеше, первый ответ содержит `X-Cache: MISS`, второй — `X-Cache: HIT`. При недоступном кеше Ledger читает PostgreSQL и возвращает `X-Cache: BYPASS`.

```bash
curl -i 'http://127.0.0.1:8081/api/reports/summary?from=2025-09-01&to=2025-09-30'
curl -i 'http://127.0.0.1:8081/api/reports/summary?from=2025-09-01&to=2025-09-30'
```

Обе даты входят в диапазон. Ответ содержит суммы расходов по категориям, например `[{"category":"Еда","total":450}]`. Для заполнения базы используйте [примеры Gateway](../gateway/README.md).
