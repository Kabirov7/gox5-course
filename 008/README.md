# HTTP-сервер

Go 1.22+, без внешних зависимостей. Данные хранятся в памяти и сбрасываются при перезапуске. Хранилище защищено `sync.RWMutex`, HTTP-обработчики работают через интерфейс `UserService`.

## Запуск

Из каталога `008`:

```sh
go run ./cmd/server
PORT=9000 go run ./cmd/server
go run ./cmd/server -port 9000
```

По умолчанию порт — `8080`. Флаг `-port` имеет приоритет над `PORT`. `SIGINT` и `SIGTERM` запускают graceful shutdown.

## API

| Метод | Путь | Успешный ответ |
| --- | --- | --- |
| GET | `/healthz` | `200`, текст `ok` |
| POST | `/users` | `200`, `{"id":"<UUIDv7>"}` |
| GET | `/users` | `200`, массив объектов `id`, `name`, `age`; `[]`, если пользователей нет |
| GET | `/users/{user_id}` | `200`, объект пользователя |
| PUT | `/users/{user_id}` | `200`, обновлённый объект пользователя |
| DELETE | `/users/{user_id}` | `204`, без тела |

`POST` и `PUT` принимают `{"name":"Alice","age":30}`. Оба поля обязательны: имя — непустая строка (пробелы по краям удаляются), возраст — целое число от нуля. Неизвестные поля, неверные типы, `null`, несколько JSON-объектов и тело более 1 МиБ отклоняются с `400`. Ошибки имеют вид `{"error":"описание"}`. Для отсутствующего пользователя `GET`, `PUT` и `DELETE` возвращают `404`.

```sh
curl -i localhost:8080/healthz
curl -i -X POST localhost:8080/users \
  -H 'Content-Type: application/json' \
  -d '{"name":"Alice","age":30}'
curl -i localhost:8080/users
# Подставьте id из ответа POST:
curl -i localhost:8080/users/USER_ID
curl -i -X PUT localhost:8080/users/USER_ID \
  -H 'Content-Type: application/json' \
  -d '{"name":"Alice","age":31}'
curl -i -X DELETE localhost:8080/users/USER_ID
```

Каждый запрос получает сгенерированный `X-Request-ID`. Он добавляется в JSON-логи `slog` как `request_id`; лог запроса содержит `method`, `path`, `status`, `duration_ms`. Логгер запроса доступен через `httpapi.Logger(ctx)`. Middleware перехватывает панику, записывает stack trace в лог и возвращает `500` без внутренних деталей.

## Проверки

```sh
go test -race ./...
go vet ./...
```
