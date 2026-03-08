# In-Memory Database

Простая key-value база данных в памяти с поддержкой команд `SET`, `GET`, `DEL`.
Работает как TCP-сервер с текстовым протоколом.

## Быстрый старт

```bash
# запустить сервер
make run-server

# подключиться клиентом (в другом терминале)
make run-client
```

```
Connected to 127.0.0.1:3223 (type EXIT to quit)
> SET name alice
[ok]
> GET name
[ok] alice
> DEL name
[ok]
> GET name
[not found]
> EXIT
Bye!
```

## Команды

| Команда | Описание | Ответ |
|---|---|---|
| `SET key value` | Сохранить значение | `[ok]` |
| `GET key` | Получить значение | `[ok] value` / `[not found]` |
| `DEL key` | Удалить ключ | `[ok]` |

## Конфигурация

По умолчанию сервер запускается с дефолтными параметрами.
Чтобы задать свои — передайте YAML-файл через флаг `--config`:

```bash
go run ./cmd/server/... --config=server.config.yaml
```

Пример файла конфигурации:

```yaml
engine:
  type: in_memory

network:
  address: "127.0.0.1:3223"
  max_connections: 100
  max_message_size: "4KB"
  idle_timeout: "5m"

logging:
  level: "info"    # debug | info | warn | error
  output: "app.log"
```

Любое незаданное поле заполняется значением по умолчанию.
Если файл не найден — сервер стартует с дефолтами без ошибки.

## CLI-клиент

```bash
# подключиться к нестандартному адресу
go run ./cmd/client/... --address=192.168.1.10:3223
```

## Make-цели

| Цель | Описание |
|---|---|
| `make run-server` | Запустить сервер |
| `make run-client` | Запустить интерактивный клиент |
| `make test` | Прогнать все тесты |
| `make test-cover` | Тесты с отчётом о покрытии |
| `make get-deps` | Обновить зависимости |
| `make install-deps` | Установить кодогенератор minimock |

## Разработка

```bash
make test                        # все тесты
go test -race ./...              # тесты с детектором гонок
go generate ./mocks/...          # перегенерировать моки
```

Логи сервера пишутся в файл, указанный в `logging.output` (`app.log` по умолчанию).

## Архитектура

```
TCP Client  →  cmd/server  →  database.HandleQuery
                               ├── compute/parser  (парсинг строки в Query)
                               └── storage         (SET/GET/DEL + txID в context)
                                    └── engine/in_memory  (map + RWMutex)
```
