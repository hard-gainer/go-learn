# Go Learn

## Запуск Gateway

Выполните из корня репозитория.

```bash
go -C gateway run .
```

Gateway слушает порт `8080` и загружает начальные бюджеты из [ledger/budgets.json](ledger/budgets.json). Транзакции и изменения бюджетов хранятся в памяти до перезапуска.

## Эндпоинты

| Метод и путь | Результат |
| --- | --- |
| `GET /ping` | `200 OK` и JSON-строка `"pong"` |
| `GET /api/transactions` | `200 OK` и массив транзакций. Если записей нет, возвращается `[]` |
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

Создайте транзакцию и запросите список. Для поля `date` используйте дату со временем и часовым поясом в формате RFC 3339. Категория в обоих запросах на создание должна совпадать, чтобы применился заданный лимит.

```bash
curl -i -X POST http://localhost:8080/api/transactions \
  -H 'Content-Type: application/json' \
  -d '{"amount":450,"category":"Еда","description":"ланч","date":"2025-09-10T12:00:00Z"}'

curl -i http://localhost:8080/api/transactions
```

Повторите запрос на создание транзакции. При лимите 500 две транзакции по 450 превышают бюджет, поэтому повторный запрос вернёт `409 Conflict` с телом `{"error":"budget exceeded"}`.
