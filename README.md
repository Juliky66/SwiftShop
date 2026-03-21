# ShopKuber

Маркетплейс в стиле Wildberries на Go микросервисах + React Native (Expo).

## Стек технологий

| Слой | Технологии |
|------|-----------|
| Бэкенд | Go, chi, sqlx, PostgreSQL, JWT RS256 |
| Мобильное приложение | Expo / React Native, React Query, Zustand |
| Инфраструктура | Docker Compose, Kubernetes (манифесты в `deploy/k8s/`) |
| Тесты | Go testify, Playwright |

## Архитектура

5 микросервисов, каждый на своём порту:

| Сервис | Порт | Назначение |
|--------|------|-----------|
| auth | 8081 | Регистрация, вход, JWT токены |
| catalog | 8082 | Категории, товары, поиск |
| orders | 8083 | Корзина, заказы |
| seller | 8084 | Управление товарами продавца |
| reviews | 8085 | Отзывы, платежи |

---

## Быстрый старт

### Требования

- [Docker](https://docs.docker.com/get-docker/) + Docker Compose
- [Go 1.22+](https://go.dev/dl/)
- [Node.js 18+](https://nodejs.org/)
- [psql](https://www.postgresql.org/download/) (клиент PostgreSQL)
- [golang-migrate](https://github.com/golang-migrate/migrate) (`brew install golang-migrate` или `go install`)

---

### 1. Первый запуск

```bash
# Клонируйте репозиторий
git clone <repo-url>
cd ShopKuber

# Сгенерируйте RSA-ключи для JWT (нужно один раз)
make keys

# Установите зависимости мобильного приложения
make mobile-install
```

---

### 2. Запуск бэкенда

**Терминал 1 — бэкенд (PostgreSQL + все 5 сервисов):**

```bash
# Собрать образы и запустить всё
make start

# Применить миграции БД (только первый раз или после сброса)
make migrate-up

# Загрузить тестовые данные
make seed
```

---

### 3. Запуск мобильного приложения

**Терминал 2 — Expo web:**

```bash
make mobile-web
```

Откройте браузер: [http://localhost:8086](http://localhost:8086)

---

### 4. Тестовые аккаунты

| Роль | Email | Пароль |
|------|-------|--------|
| Покупатель | `buyer@example.com` | `password123` |
| Продавец | `seller@example.com` | `password123` |
| Администратор | `admin@example.com` | `password123` |

---

## Основные команды

```bash
# Запуск / остановка
make start          # сборка и запуск всего через Docker Compose
make stop           # остановить все контейнеры

# База данных
make migrate-up     # применить все миграции
make migrate-down   # откатить последнюю миграцию
make seed           # сбросить и загрузить тестовые данные заново
make migrate-create # создать новую миграцию (спросит имя)

# Разработка
make mobile-web     # запустить Expo web на http://localhost:8086
make build-all      # собрать все Go бинарники в ./bin/

# Тесты и линтер
make test           # go test ./...
make lint           # golangci-lint run ./...

# Playwright (запись сценариев)
make playwright-install   # установить Playwright и Chromium
make record               # запустить тесты headless
make record-headed        # запустить тесты с видимым браузером
make playwright-report    # открыть отчёт последнего запуска
```

---

## Структура проекта

```
ShopKuber/
├── services/
│   ├── auth/          # сервис аутентификации
│   ├── catalog/       # каталог товаров
│   ├── orders/        # корзина и заказы
│   ├── seller/        # кабинет продавца
│   └── reviews/       # отзывы и платежи
├── shared/            # общий код (middleware, ошибки, пагинация)
├── migrations/        # SQL миграции и seed.sql
├── mobile/            # Expo React Native приложение
├── playwright/        # E2E тесты
├── deploy/
│   ├── docker-compose.dev.yaml
│   └── k8s/           # Kubernetes манифесты
├── keys/              # RSA ключи (не коммитить private.pem!)
├── go.work            # Go workspace
└── Makefile
```

---

## Сброс данных

```bash
# Полный сброс тестовых данных (товары, заказы, пользователи)
make seed

# Полный сброс БД (удалить все таблицы и пересоздать)
make migrate-down   # повторить нужное количество раз
make migrate-up
make seed
```

---

## Переменные окружения

Скопируйте `.env.example` в `.env` и при необходимости измените значения:

```bash
cp .env.example .env
```

Основные переменные:

| Переменная | По умолчанию | Описание |
|-----------|-------------|---------|
| `DB_DSN` | `postgres://shopkuber:shopkuber@localhost:5432/shopkuber` | Строка подключения к PostgreSQL |
| `AUTH_JWT_ACCESS_TTL` | `15m` | Время жизни access токена |
| `AUTH_JWT_REFRESH_TTL` | `720h` | Время жизни refresh токена |
