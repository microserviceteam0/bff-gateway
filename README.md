# BFF Gateway

**BFF Gateway** — учебный проект, демонстрирующий реализацию паттерна Backend for Frontend на Go.

Система состоит из нескольких микросервисов, объединённых через API Gateway:
- 🛒 **Order Service** — управление заказами
- 📦 **Product Service** — каталог товаров
- 👤 **User & Auth Service** — пользователи и JWT-авторизация
- 🚀 **BFF Gateway** — агрегация данных, кэширование Redis, rate limiting

Проект полностью контейнеризирован и запускается через `docker compose up -d`.

---

## 👥 Авторы

- Балакин Кирилл
- Батршин Денис
- Малышев Георгий

---

## 📑 Оглавление

- [📦 Модули проекта](#-модули-проекта)
    - [Monitoring](#monitoring)
    - [Order Service](#order-service)
    - [Product Service](#product-service)
    - [User & Auth Service](#user--auth-service)
    - [Shared](#shared)
- [BFF Gateway](#bff-gateway-1)
  - [🐳 Быстрый старт проекта через Docker](#-быстрый-старт-через-docker)
  - [🌟 Основные возможности](#-основные-возможности)
  - [📁 Структура проекта](#-структура-проекта)
  - [🔌 API Endpoints](#-api-endpoints)
  - [🔄 Пример агрегации данных](#-пример-агрегации-данных)
  - [⚙️ Конфигурация](#️-конфигурация)
  - [🛡️ Паттерны отказоустойчивости](#️-паттерны-отказоустойчивости)
  - [🔐 Аутентификация](#-аутентификация)
  - [📊 Мониторинг и метрики](#-мониторинг-и-метрики)
  - [📚 Swagger документация](#-swagger-документация)
  - [🧪 Тестирование](#-тестирование)
  - [🔗 URL-адреса сервисов](#-url-адреса-сервисов)

---

## 📦 Модули проекта

### [Monitoring](./monitoring)
Набор конфигураций для мониторинга микросервисов с использованием **Prometheus** и **Grafana**.  
Включает готовые дашборды и datasource конфигурации для автоматического запуска мониторинга.

─────────────────────────────

### [Order Service](./order-service)
Микросервис управления заказами пользователей.  
Реализован на **Go** с использованием **gRPC** и **PostgreSQL**.  
Обеспечивает создание, обновление, отмену и получение заказов, а также агрегацию статистики и экспорт метрик для мониторинга.

─────────────────────────────

### [Product Service](./product-service)
Микросервис управления каталогом товаров.  
Реализован на **Go** с **REST API** и **gRPC API** для межсервисного взаимодействия.  
Поддерживает полный CRUD с товарами, проверку остатков, обновление запасов и структурированное логирование.

─────────────────────────────

### [User & Auth Service](./user)
Микросервис управления пользователями и аутентификацией.  
Реализован на **Go** с использованием **PostgreSQL**, **gRPC** и **HTTP API**.  
Включает **User Service** (CRUD пользователей, управление данными) и **Auth Service** (JWT, логин/refresh токены, проверка прав доступа, мониторинг).

─────────────────────────────

### [Shared](./shared)
Общий модуль, содержащий переиспользуемые компоненты инфраструктурного уровня.  
Включает middleware и interceptor'ы для сбора метрик Prometheus (HTTP, gRPC, база данных) и упрощает мониторинг микросервисов.

---

## [BFF Gateway](./bff)
Центральный сервис **Backend for Frontend (BFF)**, объединяющий данные из нескольких микросервисов и предоставляющий клиентам единое REST API.  
Реализован на **Go 1.25** с использованием **Gin** web-фреймворка, **gRPC** для межсервисного взаимодействия и **Redis** для кэширования.

─────────────────────────────

## 🐳 Быстрый старт через Docker

Сборка всего проекта:
```bash
  docker compose up -d 
```

─────────────────────────────

## 🌟 Основные возможности

- ✅ **Агрегация данных** — объединение ответов из User, Order и Product сервисов в единый JSON-ответ
- ✅ **Кэширование Redis** — автоматическое кэширование GET-запросов с настраиваемым TTL (по умолчанию 30 секунд)
- ✅ **JWT авторизация** — защита маршрутов через валидацию токенов в Auth Service
- ✅ **Rate Limiting** — ограничение количества запросов (Token Bucket алгоритм)
- ✅ **Retry с Fallback** — автоматические повторные попытки при недоступности сервисов с graceful degradation
- ✅ **gRPC + HTTP клиенты** — гибридный подход к межсервисной коммуникации
- ✅ **Swagger документация** — автоматически генерируемая API документация
- ✅ **Prometheus метрики** — экспорт метрик для мониторинга
- ✅ **Structured Logging** — JSON логирование через slog

─────────────────────────────

## 📁 Структура проекта

```
bff/
├── main.go                    # Точка входа приложения
├── go.mod                     # Go модуль с зависимостями
├── Dockerfile                 # Multi-stage Docker сборка
├── api/
│   └── proto/                 # Скомпилированные protobuf файлы
│       ├── order/v1/          # Order Service proto
│       ├── product/           # Product Service proto
│       └── user/              # User Service proto
├── docs/                      # Swagger документация
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
└── internal/
    ├── api/
    │   └── middleware/        # API middleware
    ├── apperr/
    │   └── errors.go          # Определение ошибок приложения
    ├── clients/               # Клиенты для внешних сервисов
    │   ├── auth_client.go     # HTTP клиент Auth Service
    │   ├── user_http_client.go# HTTP клиент User Service
    │   ├── product_http_client.go # HTTP клиент с retry для Product Service
    │   ├── error_mapper.go    # Маппинг ошибок HTTP/gRPC → AppError
    │   └── grpc/
    │       ├── user. go        # gRPC клиент User Service
    │       ├── order.go       # gRPC клиент Order Service с retry
    │       └── product.go     # gRPC клиент Product Service
    ├── config/
    │   └── config.go          # Конфигурация через переменные окружения
    ├── dto/                   # Data Transfer Objects
    │   ├── order_dto.go
    │   ├── product_dto. go
    │   └── user_dto.go
    ├── handler/               # HTTP обработчики
    │   ├── handler.go         # Базовый handler с error handling
    │   ├── order_handler.go   # Обработчики заказов
    │   ├── product_handler.go # Обработчики товаров
    │   └── user_handler.go    # Обработчики пользователей и авторизации
    ├── middleware/            # Middleware слой
    │   ├── auth.go            # JWT авторизация через Auth Service
    │   ├── cache.go           # Redis кэширование с генерацией ключей
    │   ├── logger.go          # Structured logging
    │   └── ratelimit.go       # Rate limiting (Token Bucket)
    ├── router/
    │   └── router.go          # Настройка маршрутов Gin
    └── service/               # Бизнес-логика
        ├── bff_service.go     # Главный сервис с интерфейсом
        ├── order_service.go   # Логика заказов с агрегацией
        ├── product_service.go # Логика товаров
        └── user_service.go    # Логика пользователей и профилей
```

─────────────────────────────

## 🔌 API Endpoints

### Публичные маршруты

| Метод | Endpoint | Описание |
|-------|----------|----------|
| `POST` | `/api/v1/register` | Регистрация нового пользователя |
| `POST` | `/api/v1/login` | Авторизация и получение JWT токена |
| `GET` | `/api/v1/products` | Получение списка всех товаров |

### Защищённые маршруты (требуют JWT)

| Метод | Endpoint | Описание |
|-------|----------|----------|
| `GET` | `/api/v1/profile` | Получение профиля пользователя с историей заказов |
| `POST` | `/api/v1/orders` | Создание нового заказа |
| `GET` | `/api/v1/orders/{id}` | Получение деталей заказа с агрегацией данных |
| `POST` | `/api/v1/orders/{id}/cancel` | Отмена заказа |

### Служебные маршруты

| Метод | Endpoint | Описание |
|-------|----------|----------|
| `GET` | `/metrics` | Prometheus метрики |
| `GET` | `/swagger/index.html` | Swagger UI документация |

─────────────────────────────

## 🔄 Пример агрегации данных

**Запрос:** `GET /api/v1/profile`

BFF Gateway параллельно запрашивает данные из трёх сервисов:

```
┌─────────────┐     ┌──────────────────┐     ┌───────────────┐
│ User Service│     │  Order Service   │     │Product Service│
│   (gRPC)    │     │     (gRPC)       │     │    (gRPC)     │
└──────┬──────┘     └────────┬─────────┘     └───────┬───────┘
       │                     │                       │
       └─────────────────────┼───────────────────────┘
                             │
                    ┌────────▼────────┐
                    │   BFF Gateway   │
                    │  (агрегация)    │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │  Единый JSON    │
                    │    ответ        │
                    └─────────────────┘
```

**Ответ:**
```json
{
  "user":  {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com"
  },
  "orders": [
    {
      "id":  123,
      "user":  {"id": 1, "name": "John Doe", "email": "john@example.com"},
      "items": [
        {
          "product_id": 10,
          "product_name": "Laptop",
          "quantity": 1,
          "unit_price": 999.99
        }
      ],
      "status": "completed",
      "total_sum": 999.99,
      "created_at":  "2025-12-25T10:00:00Z"
    }
  ]
}
```

─────────────────────────────

## ⚙️ Конфигурация

Сервис конфигурируется через переменные окружения:

| Переменная | Описание | Значение по умолчанию |
|------------|----------|----------------------|
| `PORT` | Порт HTTP сервера | `8080` |
| `USER_SERVICE_ADDR` | gRPC адрес User Service | `localhost:50053` |
| `ORDER_SERVICE_ADDR` | gRPC адрес Order Service | `localhost:50051` |
| `PRODUCT_SERVICE_ADDR` | gRPC адрес Product Service | `localhost:50052` |
| `AUTH_SERVICE_URL` | HTTP URL Auth Service | `http://localhost:8084` |
| `USER_SERVICE_HTTP_ADDR` | HTTP URL User Service | `http://localhost:8081` |
| `PRODUCT_SERVICE_HTTP_ADDR` | HTTP URL Product Service | `http://localhost:8082` |
| `REDIS_ADDR` | Адрес Redis сервера | `localhost:6379` |
| `CACHE_TTL_SECONDS` | TTL кэша в секундах | `30` |
| `RATE_LIMIT_RPS` | Лимит запросов в секунду | `10. 0` |
| `RATE_LIMIT_BURST` | Burst размер для rate limit | `20` |
| `RETRY_ATTEMPTS` | Количество повторных попыток | `3` |
| `RETRY_DELAY_MS` | Задержка между попытками (мс) | `200` |
| `HTTP_CLIENT_TIMEOUT_MS` | Таймаут HTTP клиента (мс) | `5000` |
| `SHUTDOWN_TIMEOUT_SECONDS` | Таймаут graceful shutdown | `5` |

─────────────────────────────

## 🛡️ Паттерны отказоустойчивости

### 1. Retry с экспоненциальной задержкой
Используется библиотека `avast/retry-go` для автоматических повторных попыток при транзиентных ошибках:

```go
retry.Do(
    func() error { ...  },
    retry. Attempts(3),
    retry. Delay(200 * time.Millisecond),
    retry.DelayType(retry. BackOffDelay),
)
```

### 2. Fallback (Graceful Degradation)
При недоступности сервисов возвращаются частичные данные вместо ошибки:

```go
if err != nil {
    slog.Error("Fallback:  GetUserOrders failed.  Returning empty list")
    return &orderv1.GetUserOrdersResponse{Orders: []*orderv1.Order{}}, nil
}
```

### 3. Параллельные запросы с errgroup
Агрегация данных выполняется параллельно для минимизации latency:

```go
g, ctx := errgroup.WithContext(ctx)
g.Go(func() error { /* User Service */ })
g.Go(func() error { /* Order Service */ })
if err := g.Wait(); err != nil { ...  }
```

─────────────────────────────

## 🔐 Аутентификация

BFF Gateway использует JWT-токены для авторизации защищённых маршрутов:

1.  Клиент получает токен через `POST /api/v1/login`
2. Токен передаётся в заголовке:  `Authorization: Bearer <token>`
3. Middleware валидирует токен через Auth Service
4. При успешной валидации в контекст добавляются `userID`, `userRole`, `userEmail`

```go
// Middleware извлекает и валидирует токен
resp, err := authClient.ValidateToken(ctx, tokenString)
c.Set("userID", resp.UserID)
c.Set("userRole", resp.Role)
```

─────────────────────────────

## 📊 Мониторинг и метрики

### Prometheus метрики
Доступны по адресу `/metrics`:
- HTTP запросы (latency, count, errors)
- gRPC вызовы
- Redis connection pool stats

### Структурированное логирование
JSON-формат логов через `slog`:

```json
{
  "time": "2025-12-26T10:00:00Z",
  "level": "INFO",
  "msg": "Server listening",
  "port": "8080"
}
```

─────────────────────────────

## 📚 Swagger документация

После запуска сервиса документация доступна по адресу:
```
http://localhost:8080/swagger/index.html
```

─────────────────────────────

## 🧪 Тестирование

Запуск unit-тестов:
```bash
cd bff
go test ./...
```

─────────────────────────────

## 🔗 URL-адреса сервисов

После запуска проекта через `docker compose up -d` доступны следующие адреса:

### 📊 Мониторинг и документация

| Сервис | URL | Описание | Учётные данные |
|--------|-----|----------|----------------|
| **Grafana** | [http://localhost:3000](http://localhost:3000) | Дашборды мониторинга | `admin` / `admin` |
| **Prometheus** | [http://localhost:9090](http://localhost:9090) | Метрики и алерты | — |
| **Swagger UI (BFF)** | [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html) | API документация BFF Gateway | — |

### 🚀 API сервисов

| Сервис | HTTP API | gRPC | Метрики |
|--------|----------|------|---------|
| **BFF Gateway** | [http://localhost:8080](http://localhost:8080) | — | [/metrics](http://localhost:8080/metrics) |
| **Product Service** | [http://localhost:8083](http://localhost:8083) | `localhost:50051` | [/metrics](http://localhost:8083/metrics) |
| **Order Service** | [http://localhost:8082](http://localhost:8082) | `localhost:50052` | [/metrics](http://localhost:8082/metrics) |
| **User Service** | [http://localhost:8081](http://localhost:8081) | `localhost:50053` | [/metrics](http://localhost:8081/metrics) |
| **Auth Service** | [http://localhost:8084](http://localhost:8084) | — | [/metrics](http://localhost:8081/metrics) |

### 🗄️ Базы данных и инфраструктура

| Сервис | Хост | Порт | База данных | Пользователь |
|--------|------|------|-------------|--------------|
| **Product DB** | `localhost` | `5433` | `products_db` | `product_user` |
| **Order DB** | `localhost` | `5434` | `orders_db` | `order_user` |
| **User DB** | `localhost` | `5435` | `users_db` | `user_user` |
| **Redis** | `localhost` | `6379` | — | — |

### 🔍 Health-check эндпоинты

| Сервис | URL |
|--------|-----|
| Product Service | [http://localhost:8083/health](http://localhost:8083/health) |
| Order Service | [http://localhost:8082/health](http://localhost:8082/health) |
| User Service | [http://localhost:8081/health](http://localhost:8081/health) |
