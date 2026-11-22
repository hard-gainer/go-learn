# Go Learn

## Запуск Gateway

Выполните из корня репозитория.

```bash
go -C gateway run .
```

Gateway слушает порт `8080` и хранит бюджеты и транзакции в PostgreSQL. Примените миграции перед запуском. В новой базе бюджеты создаются через `POST /api/budgets`.

## Миграции PostgreSQL

Из корня репозитория установите Goose в локальный каталог Ledger. [Документация по установке Goose](https://pressly.github.io/goose/installation/)

```bash
GOBIN="$PWD/ledger/.bin" go install github.com/pressly/goose/v3/cmd/goose@v3.30.1
```

Запустите PostgreSQL локально или в Docker и создайте базу `cashapp`. Укажите её адрес в `DATABASE_URL`. Makefile объединяет два исходных SQL-файла во временный файл миграции Goose. [Документация по SQL-миграциям Goose](https://pressly.github.io/goose/documentation/annotations/)

```bash
export DATABASE_URL='postgres://postgres:postgres@localhost:5432/cashapp?sslmode=disable'
make migrate-up
```

Чтобы откатить миграцию, выполните команду в том же терминале. Откат удаляет обе таблицы вместе с данными.

```bash
make migrate-down
```

Применяйте миграции перед первым запуском Ledger и после каждого изменения схемы. После успешного `make migrate-up` запустите Ledger из того же терминала, чтобы он получил `DATABASE_URL`.

```bash
go -C ledger run .
```

При успешном подключении Ledger выведет `Ledger connected to PostgreSQL`. Затем запустите Gateway командой `go -C gateway run .` с той же `DATABASE_URL` и проверьте HTTP-эндпоинты.

Если `DATABASE_URL` не задана, Ledger собирает адрес из `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASS` и `DB_NAME`. Значения по умолчанию для локальной разработки — `localhost`, `5432`, `postgres`, `postgres` и `cashapp` соответственно, с `sslmode=disable`.

## Redis и отчёт Ledger

Запустите Redis, например в Docker. Ledger читает `REDIS_ADDR` (по умолчанию `localhost:6379`), `REDIS_DB` (по умолчанию `0`) и необязательный `REDIS_PASSWORD`.

```bash
docker run -d --name cashcraft-redis -p 127.0.0.1:6379:6379 redis:7-alpine
go -C ledger run . -from 2025-09-01 -to 2025-09-30
```

Ledger выводит JSON-массив с суммой расходов по каждой категории. Обе даты входят в отчёт. Сводка хранится в Redis 30 секунд по ключу `report:summary:<from>:<to>`. При недоступности Redis Ledger строит отчёт напрямую из PostgreSQL.

## Тесты и покрытие

Из корня репозитория запустите тесты обоих модулей. Для интеграционных тестов задайте `TEST_DATABASE_URL` с адресом тестовой базы PostgreSQL. Тесты создают и удаляют отдельную схему для каждого сценария. Для проверки кеша также задайте `TEST_REDIS_ADDR` с адресом тестового Redis и при необходимости `REDIS_DB` и `REDIS_PASSWORD`. Без этих адресов соответствующие тесты пропускаются. Флаг `-v` показывает отдельные сценарии.

```bash
go -C ledger test -v ./...
go -C gateway test -v ./...
```

Покрытие операторов в Gateway и Ledger с тестовыми PostgreSQL и Redis составляет **75,5%**. Чтобы получить общий профиль для двух модулей и HTML-отчёт, выполните из корня репозитория следующее.

```bash
(
  coverage_workspace=$(mktemp -d) &&
  GOWORK=off go -C "$coverage_workspace" work init "$PWD/gateway" "$PWD/ledger" &&
  export GOWORK="$coverage_workspace/go.work" &&
  go test ./gateway/... ./ledger/... -cover \
    -coverpkg=./gateway,./gateway/internal/api,./ledger,./ledger/core,./ledger/internal/db,./ledger/internal/cache,./ledger/internal/report \
    -coverprofile=cover.out &&
  go tool cover -func=cover.out &&
  go tool cover -html=cover.out -o cover.html
)
```

Итоговый процент указан в строке `total`. Откройте `cover.html` в браузере для просмотра покрытия по файлам.

## Эндпоинты

| Метод и путь | Результат |
| --- | --- |
| `GET /ping` | `200 OK` и JSON-строка `"pong"` |
| `GET /api/transactions` | `200 OK` и массив транзакций от новых дат к старым. Для одной даты первыми идут записи с большим `id` |
| `POST /api/transactions` | `201 Created` и созданная транзакция с `id` |
| `GET /api/budgets` | `200 OK` и массив бюджетов по категориям |
| `POST /api/budgets` | `201 Created` и сохранённый бюджет. Для существующей категории запрос заменяет лимит |
| `PUT /api/budgets/{category}` | Зарезервированный маршрут, который возвращает `501 Not Implemented` |

Ответы API имеют заголовок `Content-Type: application/json; charset=utf-8`. Ошибки возвращаются в формате `{"error":"описание"}`. Некорректный JSON или ошибка валидации дают `400 Bad Request`, превышение бюджета даёт `409 Conflict`, внутренняя ошибка даёт `500 Internal Server Error`.

## Проверка через cURL

В другом терминале проверьте, что Gateway отвечает.

```bash
curl -i http://localhost:8080/ping
```

Установите бюджет для категории «Еда» и запросите список бюджетов.

```bash
curl -i -X POST http://localhost:8080/api/budgets \
  -H 'Content-Type: application/json' \
  -d '{"category":"Еда","limit":500}'

curl -i http://localhost:8080/api/budgets
```

Создайте транзакцию и запросите список. Поле `date` принимает дату `YYYY-MM-DD` или метку времени в формате RFC 3339. Дата без времени возвращается как полночь UTC в формате RFC 3339. Укажите одну и ту же категорию в бюджете и транзакции, чтобы применился заданный лимит.

```bash
curl -i -X POST http://localhost:8080/api/transactions \
  -H 'Content-Type: application/json' \
  -d '{"amount":450,"category":"Еда","description":"ланч","date":"2025-09-10"}'

curl -i http://localhost:8080/api/transactions
```

Повторите запрос на создание транзакции. При лимите 500 две транзакции по 450 превышают бюджет, поэтому повторный запрос вернёт `409 Conflict` с телом `{"error":"budget exceeded"}`.
