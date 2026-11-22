# Go Learn

## Запуск Gateway

Выполните из корня репозитория.

```bash
go -C gateway run .
```

Gateway слушает порт `8080` и загружает начальные бюджеты из [ledger/budgets.json](ledger/budgets.json). Транзакции и изменения бюджетов хранятся в памяти до перезапуска.

## Тесты и покрытие

Из корня репозитория запустите тесты обоих модулей. Флаг `-v` показывает отдельные сценарии.

```bash
go -C ledger test -v ./...
go -C gateway test -v ./...
```

Покрытие операторов в Gateway и Ledger составляет **68,3%**. Чтобы получить общий профиль для двух модулей и HTML-отчёт, выполните из корня репозитория следующее.

```bash
(
  coverage_workspace=$(mktemp -d) &&
  GOWORK=off go -C "$coverage_workspace" work init "$PWD/gateway" "$PWD/ledger" &&
  export GOWORK="$coverage_workspace/go.work" &&
  go test ./gateway/... ./ledger/... -cover -coverpkg=./gateway/...,./ledger/... -coverprofile=cover.out &&
  go tool cover -func=cover.out &&
  go tool cover -html=cover.out -o cover.html
)
```

Итоговый процент указан в строке `total`. Откройте `cover.html` в браузере для просмотра покрытия по файлам.

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

Создайте транзакцию и запросите список. Поле `date` принимает дату `YYYY-MM-DD` или метку времени в формате RFC 3339. Дата без времени возвращается как полночь UTC в формате RFC 3339. Укажите одну и ту же категорию в бюджете и транзакции, чтобы применился заданный лимит.

```bash
curl -i -X POST http://localhost:8080/api/transactions \
  -H 'Content-Type: application/json' \
  -d '{"amount":450,"category":"Еда","description":"ланч","date":"2025-09-10"}'

curl -i http://localhost:8080/api/transactions
```

Повторите запрос на создание транзакции. При лимите 500 две транзакции по 450 превышают бюджет, поэтому повторный запрос вернёт `409 Conflict` с телом `{"error":"budget exceeded"}`.
