# Shortcut

Shortcut — сервис для оркестрации и выполнения вычислительных графов с поддержкой отказоустойчивости и восстановления после сбоев.

## Описание

Граф состоит из узлов (nodes), каждый из которых может выполнять:
- HTTP-запросы к внешним сервисам
- Операции с кешем (Valkey/Redis)
- Вызовы подграфов

Узлы выполняются параллельно с учётом зависимостей (DAG). Для каждого запроса собирается трейс с детальной информацией по каждому узлу. При сбое запрос сохраняется, и к нему можно применить стратегию восстановления.

### Стратегии восстановления

| Стратегия | Описание |
|-----------|----------|
| `retry`   | Повторить запрос |
| `revert`  | Откатить выполненные шаги |
| `finish`  | Завершить обработку с частичным результатом |
| `save`    | Сохранить для ручного разбора |
| `custom`  | Пользовательская логика |

### API

```
POST   /run/:namespace_id/*path                         — запустить граф
GET    /trace/:request_id                               — получить трейс запроса
GET    /errors/:namespace_id/:graph_id                  — список упавших запросов
DELETE /errors/:namespace_id/:graph_id/:request_id      — удалить запись об ошибке
POST   /process/:request_id/:strategy                   — обработать сбой стратегией
GET    /health                                          — healthcheck
```

Метрики (Prometheus) доступны на порту `2112`.

## Стек

- **Go 1.25** — основной сервис
- **PostgreSQL 17** — хранение упавших запросов
- **MongoDB 8** — хранение трейсов (с TTL)
- **Valkey 8** — кеш
- **React 19 + TypeScript** — веб-интерфейс для просмотра трейсов и управления сбоями

## Требования

- Go 1.25+
- Docker и Docker Compose
- Node.js 22+ (для сборки фронтенда)

## Запуск

### Через Docker Compose

```bash
docker-compose up
```

Поднимает весь стек: MongoDB, PostgreSQL, Valkey, mock-сервис и само приложение на порту `8080`.

### Локально

1. Запустить инфраструктуру:
   ```bash
   docker-compose up mongo postgres valkey -d
   ```

2. Собрать и запустить сервис:
   ```bash
   make run
   ```

3. Запустить mock-сервис (опционально, для тестов):
   ```bash
   make run/mock
   ```

## Конфигурация

Конфиги находятся в `configs/`. По умолчанию используется `base.yaml`, перекрытый `dev.yaml`.

Ключевые параметры:

```yaml
http:
  port: 8080

mongo:
  uri: "mongodb://localhost:27017"
  database: "shortcut"

postgres:
  url: "postgres://shortcut:shortcut@localhost:5432/shortcut"

cache:
  addr: "localhost:6379"

trace:
  ttl: 168h  # срок хранения трейсов

failure-worker:
  interval: 5s
  batch-size: 32
  visibility-timeout: 30s
```

## Разработка

### Установка инструментов

```bash
make install
```

Устанавливает `golangci-lint` и `mockery`.

### Генерация моков

```bash
make mock
```

### Сборка

```bash
make build
```

## Тесты

```bash
make test/unit   # юнит-тесты
make test/e2e    # интеграционные тесты (требуют Docker)
make test        # все тесты
make test-coverage  # покрытие (результат в coverage.html)
```

E2E-тесты поднимают необходимые контейнеры автоматически через testcontainers.

## Линтинг и форматирование

```bash
make fmt    # форматирование
make vet    # go vet
make lint   # golangci-lint
make check  # всё сразу + тесты
```

## Grafana-дашборды

Дашборды Grafana генерируются из директории с конфигами графов скриптом `scripts/gen_dashboards.py`. Скрипт читает `<configs-dir>/<namespace>/http_router.yaml` и `<configs-dir>/<namespace>/graphs/*.yaml`, и собирает один JSON-дашборд `k8s/dashboards/shortcut.json` со строкой Cluster resources (CPU/память) и свёрнутой строкой на каждый граф (RPS / p95 latency / error ratio по ручкам и по нодам).

```bash
pip install -r scripts/requirements.txt
make gen-dashboards                             
make gen-dashboards DASHBOARDS_CONFIGS_DIR=demo/configs/shop
```

Полученный `k8s/dashboards/shortcut.json` коммитится в репозиторий. На `helm upgrade` чарт оборачивает каждый JSON из `k8s/dashboards/` в `ConfigMap` с лейблом `grafana_dashboard: "1"`, а Grafana sidecar автоматически подхватывает их и складывает в папку `shortcut`. В `docker-compose.yml` те же файлы примонтированы в Grafana через file-провайдер `tests/infra/grafana/dashboards/dashboards.yaml`.
