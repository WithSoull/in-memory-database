# PR: TCP сервер, сетевой CLI-клиент и YAML-конфигурация

## Цель

Реализовать сетевой режим работы in-memory базы данных:

- TCP сервер базы данных с текстовым протоколом (`SET key value`, `GET key`, `DEL key`).
- Отдельное CLI-приложение клиента для взаимодействия с сервером по TCP.
- Конфигурирование сервера через YAML с безопасными значениями по умолчанию.
- Ограничение одновременных подключений.
- Потокобезопасность хранилища при конкурентной работе клиентов.
- Тесты для сетевого слоя и конфигурации (в первую очередь).

## План действий

### Шаг 1. Базовая инфраструктура конфигурации

- Добавить пакет `internal/config/app` c загрузкой YAML.
- Описать структуру:
  - `engine.type`
  - `network.address`
  - `network.max_connections`
  - `network.max_message_size`
  - `network.idle_timeout`
  - `logging.level`
  - `logging.output`
- Реализовать `DefaultConfig()` и `Load(path string) (Config, error)`:
  - если поле не задано в YAML, использовать дефолт;
  - если файл не задан, использовать дефолтный конфиг;
  - ошибкой считать только невалидный YAML/формат.
- Добавить парсинг размера сообщения (`4KB`, `1MB`) и `time.Duration`.

### Шаг 2. Потокобезопасность storage/engine

- Обеспечить безопасный доступ к `hashtable` под конкурентной нагрузкой.
  - Вариант: `sync.RWMutex` + `map[string]string`.
- Сделать генератор txID потокобезопасным (через `atomic`).
- Добавить/обновить тесты конкурентного доступа в `storage`/`engine`.

### Шаг 3. Сетевой слой сервера

- Добавить пакет `internal/network/server`.
- Интерфейс сервера:
  - `NewServer(cfg Config, handler QueryHandler, logger *zap.Logger)`,
  - `Run(ctx context.Context) error`,
  - `Shutdown(ctx context.Context) error` (если нужен graceful shutdown).
- Принцип обработки:
  - `net.Listen("tcp", address)`,
  - каждое соединение в отдельной горутине,
  - recovery от panic внутри клиентской горутины,
  - `idle_timeout` через deadline на `conn`,
  - ограничитель `max_connections` через semaphore / buffered channel.
- Протокол:
  - одна строка = один запрос;
  - ответ одной строкой;
  - формат ответа полностью совпадает с текущим (`[ok]`, `[ok] value`, `[not found]`, `[error] ...`).
- Ограничение размера сообщения:
  - читать через `bufio.Reader`;
  - отбрасывать/ошибаться, если длина строки > `max_message_size`.

### Шаг 4. Приложение сервера

- Добавить отдельную точку входа, например `cmd/server/main.go`.
- Параметры CLI сервера:
  - `--config=path/to/config.yaml` (опционально),
  - при отсутствии файла используется дефолтная конфигурация.
- Инициализация зависимостей:
  - logger из конфига,
  - engine по `engine.type` (на текущий момент только `in_memory`),
  - storage + parser + database,
  - server.Run.

### Шаг 5. CLI-клиент

- Добавить `cmd/client/main.go`.
- Аргументы:
  - `--address=host:port` (по умолчанию `127.0.0.1:3223`).
- UX:
  - интерактивный режим, похожий на текущий CLI (`>`),
  - `EXIT` завершает клиент,
  - каждая команда отправляется серверу, ответ печатается в stdout.

### Шаг 6. Тесты (приоритет)

- Конфиг:
  - загрузка полного YAML,
  - частично заполненный YAML + проверка дефолтов,
  - отсутствие файла (если поддерживаем режим с дефолтами),
  - невалидные значения `max_message_size`, `idle_timeout`.
- Сервер:
  - обработка `SET/GET/DEL`,
  - лимит подключений,
  - `max_message_size`,
  - idle timeout,
  - panic-safe обработка клиента (сервер не падает).
- Интеграционный smoke:
  - поднять сервер, подключить клиента через TCP, выполнить базовый сценарий.

### Шаг 7. Документация и make-цели

- Обновить README (запуск сервера/клиента, пример конфига).
- Добавить make-цели:
  - `make run-server`
  - `make run-client`
  - `make test`
  - `make test-cover`

## Предлагаемая структура изменений

- `cmd/server/main.go`
- `cmd/client/main.go`
- `internal/config/app/config.go`
- `internal/network/server/server.go`
- `internal/network/server/session.go` (опционально)
- `internal/network/server/tests/...`
- `internal/config/app/tests/...`
- доработки:
  - `internal/database/storage/engine/in_memory/hashtable.go`
  - `internal/database/storage/id_generator.go`
  - `internal/config/zap/config.go` (под output path из конфига)

## Критерии готовности (Definition of Done)

- Сервер принимает TCP подключения и обрабатывает текстовые команды.
- Клиентское CLI работает через сеть и поддерживает `--address`.
- Конфиг читается из YAML; отсутствующие поля корректно заполняются дефолтами.
- Есть ограничение на число соединений.
- Есть защита от паник в клиентских горутинах.
- Хранилище безопасно при конкурентном доступе.
- Тесты покрывают сетевой слой и конфиг (основной фокус).
- `go test ./...` проходит.

## Риски и меры

- Риск: гонки в map/счетчике txID.
  - Мера: `RWMutex`/`atomic`, запуск `go test -race ./...`.
- Риск: зависание соединений/утечки goroutine.
  - Мера: deadlines + корректное закрытие `conn` + graceful shutdown.
- Риск: неоднозначность поведения на переполнении сообщения.
  - Мера: явный контракт ответа (`[error] message too large`), тест.
