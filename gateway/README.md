# Gateway

Gateway обслуживает прежние HTTP-эндпоинты бюджетов и транзакций на порту `8080`. Слой Ledger хранит данные в PostgreSQL. Отчёт обслуживает Ledger на порту `8081` и кеширует его в Redis на 30 секунд.

## Запуск

Создайте базу `cashapp` командой `createdb -h localhost -U postgres cashapp` или при запуске контейнера с `POSTGRES_DB=cashapp`. Установите Goose и запустите Redis по [инструкции Ledger](../ledger/README.md). Из корня репозитория выполните следующие команды.

```bash
export DATABASE_URL='postgres://postgres:postgres@localhost:5432/cashapp?sslmode=disable'
make migrate-up
export REDIS_ADDR=localhost:6379
export REDIS_DB=0
export REDIS_PASSWORD=''
go -C ledger run . -listen 127.0.0.1:8081
```

В другом терминале задайте ту же `DATABASE_URL` и запустите Gateway.

```bash
export DATABASE_URL='postgres://postgres:postgres@localhost:5432/cashapp?sslmode=disable'
go -C gateway run .
```

## Проверка через cURL

```bash
curl -i http://127.0.0.1:8080/ping

curl -i -X POST http://127.0.0.1:8080/api/budgets \
  -H 'Content-Type: application/json' \
  -d '{"category":"Еда","limit":500}'

curl -i -X POST http://127.0.0.1:8080/api/transactions \
  -H 'Content-Type: application/json' \
  -d '{"amount":450,"category":"Еда","description":"ланч","date":"2025-09-10"}'

curl -i http://127.0.0.1:8080/api/budgets
curl -i http://127.0.0.1:8080/api/transactions
```

При первом прогоне на пустой базе оба запроса `POST` вернут `201 Created`. Повторная транзакция превысит лимит и вернёт `409 Conflict`. После записи запросите сводку у Ledger дважды в пределах 30 секунд. Для нового ключа заголовок `X-Cache` изменится с `MISS` на `HIT`.

```bash
curl -i 'http://127.0.0.1:8081/api/reports/summary?from=2025-09-01&to=2025-09-30'
curl -i 'http://127.0.0.1:8081/api/reports/summary?from=2025-09-01&to=2025-09-30'
```
