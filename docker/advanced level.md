# Docker. Продвинутый уровнень

## Многоэтапная сборка (Multi-stage builds) и оптимизация образов

### 1. Написание эффективных `Dockerfile` с минимизацией слоёв.

Каждая инструкция в `Dockerfile` (`RUN`, `COPY`, `ADD`) создаёт новый слой. Чем больше слоёв — тем:
* больше размер образа,
* медленнее сборка и загрузка,
* сложнее отладка и анализ

Необходимо объединять команды в один `RUN`, где это логично:
```bash
# Плохо: 3 слоя
RUN apt-get update
RUN apt-get install -y curl
RUN apt-get clean

# Хорошо: 1 слой (и кэш не сломается)
RUN apt-get update && \
    apt-get install -y --no-install-recommends curl && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*
```

* Группируйте логически связанные операции: установка зависимостей, настройка, очистка — в одной строке.
* Не копируйте файлы по одному, если можно одним `COPY`.
* В Go: `go mod download` + `go build` — можно оставить отдельно (для кэширования).
* В PHP: установка расширений через `docker-php-ext-install` — объедините в один `RUN`.

***Один слой — одна логическая операция***

### 2. Использование `scratch` или `distroless`-образов для ультракомпактных и безопасных финальных образов.

`scratch` - Это пустой образ — буквально ноль байт. Подходит только для статически скомпилированных бинарников.
```docker
FROM golang:1.23 AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

FROM scratch
COPY --from=builder /app/server /server
CMD ["/server"]
```

Минусы:

* Нет `ca-certificates` → HTTPS не работает,
* Нет `tzdata` → проблемы с временем,
* Нет `sh` → нельзя сделать docker exec.

`distroless` - Образы от Google, содержащие только runtime и необходимые зависимости (например, libc, CA certs), но без shell и менеджеров пакетов.
```docker
FROM gcr.io/distroless/static-debian12
COPY --from=builder /app/server /server
CMD ["/server"]
```

* Работает HTTPS,
* Безопасный (нет shell),
* Размер ~20 МБ.

- Для Go → `distroless/static` (если нужен HTTPS) или `scratch` (если нет).
- Для PHP/Python/JS → `distroless` не подходит (требуется интерпретатор); используйте alpine или debian-slim.

### 3. Кэширование слоёв и как его использовать/обходить.

Docker кэширует каждый слой, если:

* Инструкция не изменилась,
* Все предыдущие слои тоже не изменились.

Разделяйте «редко меняющееся» и «часто меняющееся»:
```docker
# Сначала — зависимости (меняются редко)
COPY go.mod go.sum ./
RUN go mod download

# Потом — код (меняется часто)
COPY . .
RUN go build -o app .
```

При изменении .go-файлов не перекачиваются зависимости — сборка быстрая.

Когда кэш мешает?
* При обновлении зависимостей: изменение `go.mod` → всё после этого пересобирается.
* При использовании `.env` или динамических данных — кэш может использовать устаревшие значения.

Как обойти кэш?
* `docker build --no-cache` — полная пересборка.
* Использование *BuildKit* с секретами или динамическими аргументами.

**Best practice:** порядок `COPY` — сначала конфигурация зависимостей, потом код.

### 4. Применение `.dockerignore` для исключения ненужных файлов.

Это:
* Ускоряет сборку (меньше данных копируется в build context),
* Уменьшает размер образа,
* Повышает безопасность (не попадают `.env`, `.git`, ключи).

Пример `.dockerignore`:
```
# Исходники контроля версий
.git
.gitignore

# Локальные файлы
.env
.env.local
.DS_Store

# IDE / редакторы
.vscode/
.idea/

# Зависимости (они устанавливаются внутри!)
node_modules/
vendor/

# Бинарники и временные файлы
*.exe
*.log
tmp/
```

* `.dockerignore` применяется на этапе копирования в build context, до запуска Dockerfile.
* Без него `COPY . .` может скопировать гигабайты ненужных данных.

***Копируй только то, что нужно для сборки и запуска***

### 5. Сборка с аргументами (`ARG`, `--build-arg`) и целевыми стадиями (`--target`).

`ARG` — параметры сборки. Позволяют делать гибкие Dockerfile, адаптируемые под окружение.
```docker
ARG GO_VERSION=1.23
ARG APP_ENV=production

FROM golang:${GO_VERSION}-alpine AS builder
# ...
```

```bash
docker build --build-arg GO_VERSION=1.22 --build-arg APP_ENV=staging -t my-app .
```

***`ARG` не виден во время выполнения контейнера (в отличие от `ENV`). Используйте `ENV` для переменных окружения при запуске.***

`--target` — сборка до определённой стадии

Полезно для:

* Тестирования: собрать только builder, запустить тесты.
* Dev-сборки: оставить отладочные инструменты.

```docker
FROM golang:1.23 AS builder
# ... сборка

FROM alpine AS runtime
# ... финальный образ

FROM runtime AS debug
RUN apk add curl netcat-openbsd
```

Сборка только для тестов:
```bash
docker build --target builder -t my-app-builder .
```

Сборка с отладкой:
```bash
docker build --target debug -t my-app-debug .
```

Один Dockerfile — три режима: `prod`, `debug`, `test`.

## Docker Compose — управление многоконтейнерными приложениями

### 1. Написание `docker-compose.yml` с несколькими сервисами (бэкенд, фронтенд, БД, кэш).

Пример: Laravel-приложение (PHP) + PostgreSQL + Redis + Go-микросервис
```yaml
# docker-compose.yml
version: '3.8'

services:
  # Основной бэкенд на PHP (Laravel/Symfony)
  app:
    build:
      context: .
      dockerfile: ./docker/php/Dockerfile
    ports:
      - "8000:8000"
    volumes:
      - .:/var/www/html
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_started
    networks:
      - app-network
    environment:
      DB_HOST: db
      REDIS_HOST: redis
      APP_ENV: local

  # Go-микросервис (например, обработчик очередей)
  worker:
    build:
      context: ./worker
      dockerfile: Dockerfile
    depends_on:
      rabbitmq:
        condition: service_healthy
    networks:
      - app-network
    environment:
      RABBITMQ_URL: amqp://guest:guest@rabbitmq:5672

  # База данных
  db:
    image: postgres:16-alpine
    restart: always
    volumes:
      - postgres_data:/var/lib/postgresql/data
    environment:
      POSTGRES_DB: myapp
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U user -d myapp"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - app-network

  # Кэш (Redis)
  redis:
    image: redis:7-alpine
    restart: always
    volumes:
      - redis_data:/data
    networks:
      - app-network

  # Брокер сообщений
  rabbitmq:
    image: rabbitmq:3-management-alpine
    restart: always
    ports:
      - "15672:15672"  # UI для отладки
    volumes:
      - rabbitmq_data:/var/lib/rabbitmq
    networks:
      - app-network

  # Фронтенд (Vue.js)
  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    ports:
      - "3000:3000"
    volumes:
      - ./frontend:/app
    networks:
      - app-network

# Сети и тома
networks:
  app-network:
    driver: bridge

volumes:
  postgres_data:
  redis_data:
  rabbitmq_data:
```

* Каждый сервис — изолирован, но может общаться через общую сеть `app-network`.
* Имена сервисов (`db`, `redis`) становятся DNS-именами внутри сети.
* Тома (`volumes`) обеспечивают сохранение данных между перезапусками.

### 2. Использование сетей, томов, зависимостей (`depends_on`), health checks.

Сети (`networks`):
* По умолчанию каждый `docker-compose.yml` создаёт свою отдельную bridge-сеть.
* Все сервисы в этом файле автоматически подключаются к ней и могут общаться по имени сервиса (http://app:8000, redis://redis:6379).
* Можно создавать кастомные сети (как в примере) или подключаться к внешним.

Тома (`volumes`):
* **Named volumes** (`postgres_data:`) — управляются Docker’ом, идеальны для БД.
* **Bind mounts** (`./:/var/www/html`) — проброс папки с хоста, удобен в разработке.
* Тома сохраняются даже после `docker-compose down` (если не указать `-v`).

`depends_on` - указывает порядок запуска, но не ждёт полной готовности сервиса по умолчанию
```yaml
depends_on:
  db:
    condition: service_healthy  # ← ждёт, пока healthcheck не станет OK
```

Без `condition: service_healthy` `depends_on` только гарантирует запуск контейнера, но не то, что БД уже принимает подключения. Это частая ошибка.

Health checks - показывают, действительно ли сервис готов к работе:
```yaml
healthcheck:
  test: ["CMD-SHELL", "pg_isready -U user"]
  interval: 10s
  timeout: 5s
  retries: 5
  start_period: 10s  # даём время на старт
```

* Для PHP: можно проверить через `curl -f http://localhost:8000/health`
* Для Go: `/health` эндпоинт
* Для Redis: `redis-cli ping`

**Best practice:** Используйте `depends_on` + `healthcheck` для надёжного старта.

### 3. Переопределение конфигурации (dev vs prod через `docker-compose.override.yml`).

При запуске `docker-compose up` автоматически сливается с `docker-compose.override.yml`, если он существует. Это позволяет не дублировать основную конфигурацию.

Пример: `docker-compose.override.yml` (для разработки)
```yaml
# docker-compose.override.yml
version: '3.8'

services:
  app:
    # Пробрасываем код для "горячей" перезагрузки
    volumes:
      - .:/var/www/html
    # Включаем Xdebug или отладку
    environment:
      - XDEBUG_MODE=develop,debug

  frontend:
    volumes:
      - ./frontend:/app
    environment:
      - NODE_ENV=development

  # Добавляем pgAdmin для удобства
  pgadmin:
    image: dpage/pgadmin4
    ports:
      - "5050:80"
    environment:
      PGADMIN_DEFAULT_EMAIL: admin@admin.com
      PGADMIN_DEFAULT_PASSWORD: admin
    networks:
      - app-network
```

В продакшене этого файла нет, поэтому:
* Нет проброса кода,
* Нет отладочных инструментов,
* Нет лишних портов.

Можно использовать несколько файлов (`docker-compose.prod.yml`, `docker-compose.dev.yml`) и запуск через:
```bash
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up
```

### 4. Переменные окружения, секреты, `.env`-файлы.

Docker Compose автоматически читает переменные из `.env` в той же папке:
```
# .env
DB_USER=user
DB_PASSWORD=strongpassword123
APP_DEBUG=true
```

Использование в `docker-compose.yml`:
```yaml
services:
  app:
    environment:
      DB_USER: ${DB_USER}
      DB_PASSWORD: ${DB_PASSWORD}
      APP_DEBUG: ${APP_DEBUG}
```

`.env` не передаётся в контейнеры напрямую — только через `environment` или `env_file`.

`env_file` - Можно загружать целый файл:
```yaml
services:
  app:
    env_file:
      - .env
      - .env.local
```

* В разработке: `.env` (но не коммить в Git).
* В продакшене: использовать менеджеры секретов (HashiCorp Vault, AWS Secrets Manager) или Docker secrets (в Swarm).
* В Compose (для имитации):
    ```yaml
    services:
    app:
        secrets:
        - db_password

    secrets:
    db_password:
        file: ./secrets/db_password.txt
    ```

**Best practice:** `.env` в `.gitignore`, шаблон — `.env.example`.

### 5. Запуск в фоне, управление жизненным циклом.

|Команда|Назначение|
|---|---|
|`docker-compose up`|Запуск всех сервисов в *foreground* (логи в терминале)|
|`docker-compose up -d`|Запуск в фоне (detached mode)|
|`docker-compose down`|Остановка и удаление контейнеров, сетей. Тома остаются|
|`docker-compose down -v`|То же + удаление томов (осторожно: потеря данных)|
|`docker-compose logs app`|Просмотр логов сервиса `app`|
|`docker-compose logs -f app`|Логи в реальном времени|
|`docker-compose exec app sh`|Зайти внутрь контейнера `app`|
|`docker-compose build`|Пересобрать образы|
|`docker-compose restart app`|Перезапустить один сервис|

Полезные флаги:
* `--build`: пересобрать при `up` → `docker-compose up -d --build`
* `--force-recreate`: пересоздать контейнеры даже без изменений

Пример workflow:
```bash
# Собрать и запустить в фоне
docker-compose up -d --build

# Посмотреть логи бэкенда
docker-compose logs -f app

# Зайти в PHP-контейнер для artisan
docker-compose exec app php artisan tinker

# Остановить всё
docker-compose down
```

## Сетевое взаимодействие в Docker

### 1. Типы сетей: `bridge`, `host`, `none`, `overlay`.

Docker поддерживает несколько драйверов сетей. Каждый решает разные задачи.

* `bridge` (по умолчанию)
  * Частная внутренняя сеть на хосте.
  * Как работает:
    - Каждый контейнер получает виртуальный IP в диапазоне (например, `172.17.0.0/16`).
    - Контейнеры в одной bridge-сети могут общаться по IP.
    - Для связи с внешним миром используется NAT (masquerading).
  * Когда использовать: почти всегда для локальной разработки и изолированных сервисов.
  * Ограничение: по умолчанию контейнеры из разных bridge-сетей не видят друг друга.

**По умолчанию все контейнеры без указания сети попадают в default bridge, где нет DNS-имён — только IP. Это плохо! Лучше создавать пользовательские bridge-сети**

* `host`
  * Контейнер делит сетевое пространство с хостом.
  * Как работает:
    - Нет изоляции сети.
    - Контейнер использует реальные порты хоста напрямую.
    - Нет NAT, нет виртуальных интерфейсов.
  * Когда использовать:
    - Высокопроизводительные приложения (минимизация overhead),
    - Сетевые демоны (например, Prometheus node_exporter),
    - Когда нужно слушать на localhost хоста.
  * Минусы:
    - Нет изоляции → конфликты портов,
    - Небезопасно (контейнер видит все сетевые интерфейсы хоста).

`Не работает на Docker Desktop для Mac/Windows (только Linux)`

* `none`
  * Полное отключение сети.
  * Как работает:
    - У контейнера только lo (loopback).
    - Нет доступа к интернету, нет связи с другими контейнерами.
  * Когда использовать:
    - Изолированные задачи без сети (например, обработка файла),
    - Безопасность (если приложению сеть не нужна).

```bash
docker run --network none alpine ip addr
# Только lo: 127.0.0.1
```

* `overlay`
  * Сеть для многонодовых кластеров (Docker Swarm, Kubernetes через CNI).
  * Как работает:
    - Использует VXLAN для туннелирования трафика между хостами.
    - Все контейнеры в overlay-сети видят друг друга, даже если на разных серверах.
  * Когда использовать:
    - Только в оркестрации (Swarm или k8s).
    - Не используется в обычном docker run.

`В standalone Docker (docker-compose) overlay не применяется`

### 2. Создание пользовательских bridge-сетей для изоляции и DNS-разрешения имён контейнеров.

В `default bridge`:
* Нет DNS → нельзя писать curl http://redis:6379, только по IP.
* Все контейнеры в одной плоскости → меньше изоляции.

Пользовательская bridge-сеть:
```bash
# Создаём сеть
docker network create my-app-net

# Запускаем контейнеры в ней
docker run -d --name redis --network my-app-net redis:7
docker run -it --name app --network my-app-net alpine

# Теперь внутри app можно:
ping redis        # ← работает!
curl http://redis:6379
```

1. Docker автоматически запускает встроенный DNS-сервер.
2. Имя контейнера = DNS-имя.
3. Можно задать алиас: --network-alias cache.

В `docker-compose.yml`:

```yaml
services:
  app:
    networks:
      - mynet

  redis:
    networks:
      - mynet

networks:
  mynet:
    driver: bridge
```

`Best practice: всегда используйте пользовательские bridge-сети, даже для одного сервиса — ради DNS и чистоты`

### 3. Проброс портов (`-p`), различие между EXPOSE и публикацией.

EXPOSE в Dockerfile:
* Это документация
* Говорит: «приложение внутри слушает на этом порту».
* Не публикует порт наружу.
```dockerfile
EXPOSE 8080
```

Без `-p` вы не сможете обратиться к контейнеру с хоста. `-p` (или `--publish`) фактически пробрасывает порт с хоста в контейнер.
```bash
docker run -p 8000:8080 my-app
```
http://localhost:8000 → контейнер:8080.

Форматы:
* `-p 8080:8080` — явное сопоставление,
* `-p 8080` — случайный порт на хосте (редко используется),
* `-p 127.0.0.1:8080:8080` — только для localhost (безопаснее).

Важно:
* Если контейнеры в одной сети — проброс портов не нужен! Они общаются напрямую по внутреннему порту.
* `-p` нужен только для доступа с хоста или извне.

Пример:
* Бэкенд (:8080) и фронтенд (:3000) — общаются внутри сети без `-p`.
* Но чтобы открыть фронтенд в браузере — нужен `-p 3000:3000`

### 4. Отладка сетей: `docker network inspect`, `nsenter`, `ping`, `curl` между контейнерами.

`docker network inspect <network>` показывает:
* Все подключённые контейнеры
* Их IP-адреса
* Настройки DNS, gateway
```bash
docker network inspect my-app-net
```
Ищите секцию "Containers" — там IP и MAC.

`docker exec` + `ping / curl` - Самый простой способ проверить связь:
```bash
# Зайти в контейнер app
docker exec -it app sh

# Проверить, видит ли он redis
ping redis
curl http://redis:6379
telnet db 5432
```
`Убедитесь, что в образе есть ping, curl, telnet (в alpine их нет по умолчанию — ставьте apk add curl).`

`nsenter` — продвинутая отладка (на уровне ядра). Позволяет «войти» в сетевое пространство контейнера с хоста, даже если в контейнере нет `shell`.
```bash
# Найти PID контейнера
PID=$(docker inspect -f '{{.State.Pid}}' my-container)

# Войти в его net namespace
sudo nsenter -t $PID -n

# Теперь вы «внутри» сети контейнера
ip addr
ping 8.8.8.8
```
Полезно, когда контейнер на `scratch` или `distroless`.

`docker port <container>` — показывает, какие порты опубликованы.
`iptables -L -n -t nat` — смотреть правила NAT (если интересно, как работает `-p`).

### 5. Использование `--network host` и его последствия.

Запустить:
```bash
docker run --network host nginx
```

* Nginx слушает напрямую на порту 80 хоста.
* Нет NAT, нет моста — максимальная производительность.

Плюсы:
* Нулевой сетевой overhead,
* Простота (не нужно `-p`),
* Доступ ко всем интерфейсам хоста (например, `eth0`, `lo`).

Минусы и риски:
* Нет изоляции → Контейнер видит все процессы и порты хоста
* Конфликты портов → Нельзя запустить два контейнера на одном порту
* Безопасность → Контейнер может сканировать сеть хоста
* Нет DNS по имени → В host-сети нет встроенного DNS Docker
* Не переносимо → Не работает на Mac/Windows (Docker Desktop)

`Если нужна производительность — используйте пользовательскую bridge-сеть + оптимизацию приложения. --network host — крайняя мера`

## Управление томами (Volumes) и хранение данных

### 1. Разница между `bind mounts`, `named volumes`, `tmpfs`.

|Тип|Где хранится|Кто управляет|Используется для|Особенности|
|--|--|--|--|--|
|**Bind mount**|Любая папка/файл на хосте (`/home/user/data`)|Пользователь|Разработка, конфиги, логи|Прямая привязка к ФС хоста. Может создать файл как папку|
|**Named volume**|Внутри Docker (`/var/lib/docker/volumes/...`)|Docker|БД, постоянные данные|Абстрагирован от хоста. Безопаснее, переносимее|
|**tmpfs**|Только в оперативной памяти хоста|Ядро Linux|Временные данные (сессии, кэш)|Данные исчезают при остановке. Не работает на Windows/macOS|

**Bind mount:**
```bash
docker run -v /host/path:/container/path image
# или
docker run --mount type=bind,source=/host/path,target=/container/path image
```

Используйте для:
* Проброса кода в dev-режиме (`$(pwd):/app`),
* Конфигов (`.env`, `nginx.conf`),
* Логов (`/logs:/app/logs`).

**Named volume:**
```bash
docker volume create mydata
docker run -v mydata:/app/data image
```

Используйте для:
* PostgreSQL, MySQL, Redis, RabbitMQ,
* Любых данных, которые должны пережить пересоздание контейнера.

**tmpfs:**
```bash
docker run --tmpfs /app/cache image
# или
docker run --mount type=tmpfs,destination=/app/cache image
```

Используйте для:
* Сессий, временных токенов, кэша,
* Чувствительных данных, которые не должны записываться на диск.

### 2. Создание и управление томами через CLI и Compose.

```bash
# Создать named volume
docker volume create app_data

# Посмотреть список
docker volume ls

# Инспектировать (узнать путь на хосте)
docker volume inspect app_data

# Удалить (только если не используется!)
docker volume rm app_data

# Удалить все неиспользуемые тома
docker volume prune
```

`docker-compose.yml`
```yaml
version: '3.8'

services:
  db:
    image: postgres:16
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:  # ← Docker сам создаст named volume
    # можно указать драйвер, опции и т.д.
    # driver: local
    # driver_opts:
    #   type: 'none'
    #   o: 'bind'
    #   device: '/mnt/data/postgres'
```
Если имя тома не указано в корне (`volumes:`), Compose создаст том с префиксом (например, `myproject_postgres_data`).

### 3. Инициализация томов данными из контейнера (volume initialization pattern).

Вы создаёте `named volume` для БД. При первом запуске том пустой → БД инициализируется «с нуля».
Но что, если вы хотите заполнить том начальными данными (например, SQL-дампом, конфигами)?

Volume Initialization Pattern - Docker автоматически копирует содержимое точки монтирования из образа в том, если том пустой.

В `Dockerfile`:
```dockerfile
FROM alpine
RUN mkdir -p /init-data
COPY init.sql /init-data/
```

В `docker-compose.yml`:
```yaml
services:
  init:
    build: .
    volumes:
      - app_data:/init-data

  app:
    image: my-app
    volumes:
      - app_data:/app/data

volumes:
  app_data:
```

При первом запуске:
* Том `app_data` пуст → Docker копирует `/init-data` из контейнера `init` в том.
* Теперь `app` видит эти данные.

`Копирование происходит ТОЛЬКО если том пуст! После этого изменения в образе не попадут в том.`

Альтернатива: init-контейнер
В продвинутых сценариях (особенно в Kubernetes) используют init-контейнер, который заполняет том перед стартом основного сервиса.

### 4. Резервное копирование и восстановление данных из томов.

Резервное копирование.
Так как тома — это просто папки на хосте (или в облаке), резервная копия = архив этой папки.

Способ 1: через временный контейнер
```bash
# Создаём контейнер, который монтирует том и архивирует его
docker run --rm \
  -v postgres_data:/volume \
  -v $(pwd):/backup \
  alpine \
  tar czf /backup/postgres_backup.tar.gz -C /volume .
```
→ Получаем `postgres_backup.tar.gz` в текущей папке.

Способ 2: напрямую (если знаете путь)

```bash
# Найти путь тома
docker volume inspect postgres_data

# Архивировать
sudo tar czf backup.tar.gz -C /var/lib/docker/volumes/postgres_data/_data .
```

`Не рекомендуется напрямую работать с /var/lib/docker — лучше через контейнер.`

Восстановление:
```bash
# Остановить сервис
docker-compose down

# Удалить старый том (осторожно!)
docker volume rm myproject_postgres_data

# Создать новый том
docker volume create myproject_postgres_data

# Распаковать данные в него
docker run --rm \
  -v myproject_postgres_data:/volume \
  -v $(pwd):/backup \
  alpine \
  tar xzf /backup/postgres_backup.tar.gz -C /volume
```

→ Запускаем `docker-compose up` — данные восстановлены.

`Для PostgreSQL/MySQL лучше использовать логические дампы (pg_dump, mysqldump), а не копирование файлов — особенно при разных версиях`

### 5. Использование томов для shared state между контейнерами.

Тома — отличный способ делиться данными между контейнерами без сети.

Общий кэш или upload-директория:
```yaml
services:
  frontend:
    image: nginx
    volumes:
      - shared_uploads:/var/www/uploads

  backend:
    image: my-go-app
    volumes:
      - shared_uploads:/app/uploads

volumes:
  shared_uploads:
```

Оба сервиса читают/пишут в одну папку.

CI-агент + runner
* Агент кладёт артефакты в том,
* Runner забирает их оттуда.

Sidecar-контейнер для логов
* Основной контейнер пишет логи в /logs,
* Sidecar (Fluentd, Filebeat) читает их оттуда и отправляет в централизованную систему.

Важно:
`Убедитесь, что нет конфликтов записи (лучше один writer + много readers),
следите за правами доступа (UID/GID в контейнерах должны совпадать или быть настроены правильно)`

## Безопасность контейнеров

### 1. Запуск без root: `USER` в Dockerfile, `--user`, user namespace remapping.

* `USER` в Dockerfile: Директива, указывающая, от какого пользователя (обычно не-root, например, USER 1000) будут запускаться последующие инструкции и сам контейнер. Всегда используйте это, даже если приложение требует root для сборки, запуск должен быть от непривилегированного пользователя.

```dockerfile
FROM alpine
RUN adduser -D myuser && chown -R myuser /app
USER myuser
CMD ["/app/start.sh"]
```

* `--user` в `docker run`: Переопределяет пользователя при запуске. Например, `docker run --user 1000:1000 myimage`. Полезно для локальной отладки или в CI/CD, где ID пользователя на хосте известен.

* User Namespace Remapping (отображение пространства имен пользователей): Самая мощная изоляция на уровне ОС. Включается в демоне Docker (`/etc/docker/daemon.json`). Позволяет отображать root внутри контейнера (UID 0) на непривилегированного пользователя на хосте (например, UID 100000). Даже если атакующий сломает контейнер и станет root внутри, его привилегии на хосте будут ограничены. **Обязательная настройка для shared-хостов (K8s нод).**

Best practice: `Всегда запускайте приложения от непривилегированного пользователя`

### 2. Минимизация attack surface: устранение shell, утилит (`rm /bin/sh` в `scratch`).

Чем меньше пакетов и бинарников в контейнере, тем меньше векторов для атаки.

* Устранение shell (`/bin/sh`, `/bin/bash`): Без shell многие RCE-уязвимости становятся бесполезны. Используйте `scratch` (пустой образ) или `distroless`-образы от Google. Проверьте: `docker run --rm your-image sh`. Если команда не найдена — отлично.

* Удаление отладочных утилит (`curl`, `ps`, `netstat`): Они помогают атакующему исследовать окружение и эскалировать привилегии. В финальном образе оставляйте только бинарник приложения и его минимальные зависимости.

* Многоступенчатая сборка (multi-stage) — ключевой инструмент:

```dockerfile
# Этап сборки (здесь может быть всё)
FROM golang:alpine AS builder
WORKDIR /app
COPY . . && go build -o myapp

# Финальный образ (только бинарник)
FROM scratch
COPY --from=builder /app/myapp /myapp
USER 1000
CMD ["/myapp"]
```

### 3. Сканирование уязвимостей: `docker scout`, `trivy`, `snyk`.

Это обязательный шаг в CI/CD пайплайне. Позволяет находить известные уязвимости (CVEs) в базовых образах и зависимостях.

* `docker scout` (ранее `docker scan`): Интегрирован в Docker CLI, использует базу Snyk. Просто: `docker scout cves my-image`.

* `trivy` (Aqua Security): Лидер по популярности. Бесплатный, быстрый, очень глубокий анализ (OS пакеты, языковые зависимости, конфиги). Интегрируется в GitLab CI, GitHub Actions.

```bash
trivy image --severity HIGH,CRITICAL my-registry/my-image:tag
```

* `snyk`: Облачный коммерческий инструмент с обширной БД. Хорош для приложений (SCA).

`Сканируйте не только финальный образ, но и базовый образ на этапе сборки. Блокируйте сборку при критических уязвимостях. Обновляйте базовые образы регулярно.`

### 4. Ограничение capabilities (`--cap-drop`, `--cap-add`).

Linux Capabilities — это дробные привилегии, которые обычно есть у root (например, CAP_NET_ADMIN для управления сетью). Контейнеру по умолчанию дают множество таких capabilities.

`--cap-drop ALL` / `--cap-add`: Всегда сбрасывайте все и добавляйте только необходимое. Например, контейнеру Nginx может понадобиться только CAP_NET_BIND_SERVICE для слушания портов <1024.

```bash
docker run --cap-drop ALL --cap-add NET_BIND_SERVICE nginx
```

Избегайте опасных capabilities:
* CAP_SYS_ADMIN — почти как --privileged.
* CAP_CHOWN, CAP_DAC_OVERRIDE — позволяют обходить права на файлы.
* CAP_SYS_MODULE — загрузка модулей ядра.

### 5. Запрет привилегированных контейнеров (`--privileged` = зло в prod).

`--privileged` — это абсолютное зло в production. Контейнер получает все capabilities и отключает многие механизмы изоляции (seccomp, AppArmor). Фактически получает доступ к хостовому ядру. Может понадобиться только для очень специфичных низкоуровневых задач (например, запуск Docker-in-Docker в CI). Platform Engineer должен на уровне оркестратора (K8s Pod Security Admission, OPA) запрещать такие контейнеры.

### 6. Использование `seccomp`, `AppArmor`, `SELinux` (на уровне хоста).

Это профили безопасности уровня ядра.

* `seccomp` (Secure Computing Mode): Фильтрует системные вызовы (syscalls), которые может делать процесс. Docker имеет разумный дефолтный профиль, который запрещает опасные вызовы (например, `reboot`). Для особо критичных приложений можно писать кастомные профили, запрещая всё, кроме необходимого минимума (`unshare`, `clone` и т.д.).

* `AppArmor` (Ubuntu/Debian) и `SELinux` (RHEL/Fedora): Контролируют доступ процессов к ресурсам (файлам, портам, мьютексам). Можно запретить контейнеру писать в домашнюю директорию хоста или читать `/proc`. В Docker можно указывать профиль через `--security-opt apparmor=my-profile`. В Kubernetes настройка сложнее, но возможна.

### 7. Управление секретами (через Compose secrets или внешние vault'ы).

Секреты (пароли, токены, ключи) никогда не должны быть в Dockerfile или в образе.

* Docker Compose secrets: Использует tmpfs (в памяти) для монтирования секретов в контейнер. Удобно для локальной разработки.

```yaml
services:
  app:
    image: myapp
    secrets:
      - db_password
secrets:
  db_password:
    file: ./secrets/db_password.txt  # Файл вне репозитория!
```

* Внешние vault'ы (Хранилища) — промышленный стандарт для Platform Engineering:
  - **HashiCorp Vault**: Лидер. Динамически генерирует секреты, есть leasing, аудит.
  - **AWS Secrets Manager / Azure Key Vault / GCP Secret Manager**: Нативные сервисы облаков.
  - **Kubernetes Secrets** (базовый, но лучше использовать вместе с внешними vault'ами через CSI Driver или sidecar-инжекторы, например, `vault-agent`).

Принципы:
* Никогда не логировать секреты.
* Ротация секретов без пересборки образа.
* Выдавать минимальные права (принцип наименьших привилегий) для доступа к секретам.
* Аудит — кто и когда запросил секрет.

## Производительность и мониторинг

### 1. Ограничение ресурсов: `--memory`, `--cpus`, `--pids-limit`.**

**Память (`--memory` или `-m`)**

* Docker использует `cgroups` для ограничения памяти. Контейнер не сможет использовать больше указанного лимита.
```bash
docker run -m 512m ...              # Лимит 512 МБ
docker run -m 1g --memory-swap 2g ... # Swap = 2GB, swap = memory + swap_limit
docker run -m 500m --memory-reservation 300m ... # Мягкий лимит, ядро пытается удерживаться
```

* **OOM Killer**: При превышении лимита контейнер будет убит OOM Killer. Важно: В Kubernetes при OOM контейнер перезапускается.

* `Всегда устанавливайте memory limit в production`

**CPU (`--cpus`, `--cpu-quota`, `--cpu-shares`)**

* `--cpus 1.5`: Контейнер может использовать максимум 1.5 ядра CPU. Самый простой и понятный способ.
* CPU shares (`--cpu-shares 512`): Относительный вес при конкуренции за CPU. По умолчанию 1024. Если два контейнера с shares 512 и 1024, второй получит в 2 раза больше CPU при нагрузке.
* CPU period и quota: Более точный контроль:
```bash
docker run --cpu-period=100000 --cpu-quota=50000 ...
```
Период 100000 мкс (100 мс), квота 50000 мкс = контейнер получает 50% CPU каждые 100 мс.

* CPU sets (`--cpuset-cpus 0-2`): Привязка к конкретным ядрам. Полезно для low-latency приложений или изоляции noisy соседей.

**PIDs limit (`--pids-limit`)**

* Защита от fork-бомб. Ограничивает максимальное количество процессов в контейнере.
```bash
docker run --pids-limit 100 ...
```

`Один контейнер не сможет исчерпать PID space всей ноды (обычно 32768). Стандартный лимит в K8s — 100-200 процессов.`

### 2. Мониторинг через `docker stats`, `docker top`.

* `docker stats` - лайв-мониторинг в реальном времени:
```bash
docker stats --no-stream  # Однократный вывод
docker stats --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}"  # Кастомный формат
```

Отображает:
* CPU %: Процент использования от всех ядер хоста
* MEM USAGE / LIMIT: Использование относительно лимита
* MEM %: Процент от установленного лимита
* NET I/O, BLOCK I/O: Сетевой и дисковый ввод-вывод
* PIDS: Количество процессов

`Ограничения: Только текущие метрики, нет истории. Для production нужны Prometheus + cAdvisor.`

* `docker top` - просмотр процессов внутри контейнера:
```bash
docker top <container_id> -ef  # Аналог `ps -ef`
docker top <container_id> -o pid,ppid,user,cmd  # Выбор полей
```

* Диагностика runaway-процессов
* Поиск утечек памяти (сравнить с docker stats)
* Проверка, от какого пользователя запущены процессы

### 3. Профилирование использования диска (`docker system df`).

```bash
docker system df -v  # Подробный вывод
```

Вывод показывает:
* Images: Занято образами и их слоями
* Containers: Размер контейнеров (R/W слой поверх образа)
* Local Volumes: Данные в volumes
* Build Cache: Кэш сборки (может занимать гигабайты!)

Детализация:
```bash
docker system df --format '
Тип           Всего     Активно   Размер     %
{{range .}}
{{.Type}}     {{.TotalCount}} {{.ActiveCount}} {{.Size}} {{.ReclaimablePercent}}%
{{end}}'
```

Поиск "жирных" объектов:
```bash
# Самые большие образы
docker images --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}" | sort -k3 -h -r

# Размер контейнеров
docker ps -s --format "table {{.Names}}\t{{.Size}}"
```

### 4. Очистка системы: `docker system prune`, управление мусором.

```bash
# Безопасная очистка (удаляет всё остановленное)
docker system prune

# Агрессивная очистка (включает volumes и build cache)
docker system prune -a --volumes

# Целевая очистка
docker image prune -a --filter "until=24h"  # Образы старше 24 часов
docker container prune --filter "until=2h"  # Остановленные контейнеры
docker volume prune  # Неиспользуемые volumes
docker builder prune  # Кэш сборки
```

Автоматизация очистки (для Platform Engineer):

1. Cron-задания на нодах:
```bash
# Ежедневная очистка в 3:00
0 3 * * * docker system prune -f --filter "until=48h"
```

2. Garbage Collection в Kubernetes:
* Настраивается в kubelet: `--eviction-hard`, `--image-gc-high-threshold`
* K8s сам удаляет неиспользуемые образы при нехватке места

3. Registry cleanup:
* Удаление старых тегов из registry (Harbor, GitLab Registry)
* Использовать `skopeo` или registry API

### 5. Понимание overhead'а контейнеров по сравнению с bare metal.

На что тратятся ресурсы:


|Компонент| Overhead | Комментарий|
|--|--|--|
|Память| 10-100 МБ на контейнер | Сам Docker + runtime (runc) + изоляция namespaces|
|CPU | 1-5% | Обработка системных вызовов, network virtualization|
|Диск | Минимальный | Copy-on-Write (CoW) на уровне слоев, overlayfs|
|Сеть | 1-3% | latency	iptables, bridge, overlay networks|
|I/O | 5-15% | OverlayFS, device mapper|

**Детальный разбор overhead'а:**

1. Память:
* RSS каждого контейнера = память приложения + shared libraries + page cache
* Page cache дублируется: Два контейнера с одинаковым файлом имеют свои копии в page cache
* Решение: Использовать --memory-swappiness=0 для уменьшения swap контейнеров

2. CPU и производительность:
* syscall overhead: Каждый вызов проходит через несколько слоев (runC → containerd → dockerd)
* Пример: open() в контейнере на 10-20% медленнее, чем на хосте
* Сетевой overhead:
  - Bridge: +15% latency
  - Overlay (Swarm/K8s): +25-30% latency
  - Решение: Использовать host network (`--net=host`) или CNI plugins с прямым доступом (SR-IOV)

3. . Дисковый I/O:
* OverlayFS (стандартный драйвер):
  - Write: Copy-up операция (копирование файла из нижнего слоя в верхний)
  - Read: Проверка по слоям (L1 → L2 → ...)
  - Проблема: Множественные маленькие файлы = большие потери
* Device mapper (для CentOS/RHEL):
  - Еще больший overhead, но стабильнее для production
* Решение для high I/O:
```bash
docker run -v /mnt/fast-ssd:/data ...  # Монтировать быстрый диск напрямую
```

4. Сетевой overhead в цифрах:
```bash
# TCP_RR тест (транзакции/секунду)
Bare metal: 15000 trans/sec
Docker bridge: 12000 trans/sec (-20%)
Docker overlay: 10000 trans/sec (-33%)
```
---

**Как минимизировать overhead (практические советы Platform Engineer):**

1. Плотная упаковка (high density):
* Запускать больше контейнеров на одной ноде
* Но оставлять headroom для burst нагрузок (обычно 20-30%)

2. Выбор runtime:
* **runc** (стандартный): Balanced overhead
* **gVisor** (Google): Высокая безопасность, высокий overhead (до 2x)
* **Kata Containers**: Виртуализация, минимальный overhead безопасности, но выше потребление памяти

3. Оптимизация образа:
* Меньше слоев → быстрее запуск
* Использовать `ADD` вместо `RUN curl ...` (один слой)

4. Мониторинг реального overhead'а:
```bash
# Сравнение производительности
docker run --rm --net=host alpine ping -c 1000 localhost
docker run --rm alpine ping -c 1000 google.com

# Измерение syscall overhead
docker run --rm ubuntu strace -c ls
```

5. Использование инструментов:
* `perf` для профилирования ядра
* `bpftrace` для трассировки системных вызовов
* `node-exporter` + `cAdvisor` для сбора метрик

**Overhead критичен:**
* HPC/High-Frequency Trading: Использовать специальные решения (NVIDIA enroot, Singularity)
* High-throughput databases: Рассмотреть Docker только для оркестрации, данные на хосте
* Real-time системы: Тщательно тестировать под нагрузкой, использовать реальные priorities

## Интеграция с CI/CD

### 1. Сборка и пуш образов в registry из GitLab CI / GitHub Actions.

*  GitLab CI (`.gitlab-ci.yml`)
```yaml
build-and-push:
  image: docker:24.0
  services:
    - docker:24.0-dind
  variables:
    DOCKER_TLS_CERTDIR: "/certs"
    IMAGE_NAME: $CI_REGISTRY_IMAGE/myapp
    IMAGE_TAG: $CI_COMMIT_SHA
  before_script:
    - docker login -u $CI_REGISTRY_USER -p $CI_REGISTRY_PASSWORD $CI_REGISTRY
  script:
    - docker build -t $IMAGE_NAME:$IMAGE_TAG .
    - docker push $IMAGE_NAME:$IMAGE_TAG
```

`CI_REGISTRY`, `CI_REGISTRY_USER`, `CI_REGISTRY_PASSWORD` — встроенные переменные GitLab при использовании GitLab Container Registry.

* GitHub Actions (`.github/workflows/ci.yml`)
```yaml
jobs:
  build-and-push:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Log in to registry
        uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Build and push
        uses: docker/build-push-action@v6
        with:
          context: .
          push: true
          tags: ghcr.io/${{ github.repository }}/myapp:${{ github.sha }}
```

Официальные Docker GitHub Actions (`docker/setup-buildx-action`, `docker/login-action`, `docker/build-push-action`) — они поддерживают BuildKit, кэширование и multi-platform builds.

### 2. Использование build cache в CI (например, через `--cache-from`).

Без кэша каждая сборка начинается с нуля → медленно и дорого.

* Способ 1: Pull/push промежуточного образа как кэша. *(Устаревший подход (работает, но не оптимален))*.
```bash
# Pull предыдущий образ (если есть)
docker pull $IMAGE_NAME:buildcache || true

# Build с использованием кэша
docker build \
  --cache-from $IMAGE_NAME:buildcache \
  --tag $IMAGE_NAME:$IMAGE_TAG \
  --tag $IMAGE_NAME:buildcache \
  .

# Push обновлённый кэш
docker push $IMAGE_NAME:buildcache
```

* Способ 2: BuildKit + inline cache

В Dockerfile — ничего менять не нужно.

В CI (с BuildKit):
```yaml
# GitHub Actions пример
- name: Build and push
  uses: docker/build-push-action@v6
  with:
    context: .
    push: true
    tags: ghcr.io/user/app:${{ github.sha }}
    cache-from: type=registry,ref=ghcr.io/user/app:buildcache
    cache-to: type=registry,ref=ghcr.io/user/app:buildcache,mode=max
```

`mode=max` сохраняет больше слоёв для кэширования.
Для GitLab CI используйте `DOCKER_BUILDKIT=1` и аналогичные флаги через `docker buildx`.

### 3. Тегирование образов по ветке, хешу коммита, семантической версии.

Правильное тегирование — основа traceability и rollback.

| Тип тега | Пример | Когда использовать |
|--|--|--|
| Хеш коммита | `sha256:abc123...` или `abc123` | Все сборки (immutable) |
| Имя ветки | `feature/auth` | Для preview-сред |
| SemVer | v1.2.3 | Релизы (теги в Git) |
| latest | latest | **Избегать в prod** |

GitLab CI:
```yaml
variables:
  COMMIT_SHORT: $CI_COMMIT_SHA[0:8]
  BRANCH_SLUG: $CI_COMMIT_REF_SLUG  # безопасное имя ветки
  SEMVER_TAG: $CI_COMMIT_TAG        # если запущено по git tag
```

GitHub Actions:
```yaml
env:
  COMMIT_SHORT: ${{ github.sha }}
  BRANCH: ${{ github.ref_name }}
  IS_TAG: ${{ startsWith(github.ref, 'refs/tags/') }}
```

**Best practice:**
* Каждый коммит → образ с тегом по хешу.
* Git tag v1.2.3 → пушить v1.2.3, 1.2, 1, и latest (осторожно с latest).

### 4. Работа с private registry (авторизация, `docker login`).

GitLab CI:
* Для GitLab Registry — встроенные переменные (`$CI_REGISTRY_USER`, `$CI_REGISTRY_PASSWORD`).
* Для внешнего registry (Docker Hub, AWS ECR, Harbor):
```yaml
before_script:
  - echo "$MY_REGISTRY_PASSWORD" | docker login --username $MY_REGISTRY_USER --password-stdin my-registry.com
```
Где `MY_REGISTRY_PASSWORD` — masked CI variable.

GitHub Actions: Используйте `secrets`
```yaml
- name: Login to private registry
  uses: docker/login-action@v3
  with:
    registry: my-registry.com
    username: ${{ secrets.REGISTRY_USER }}
    password: ${{ secrets.REGISTRY_PASSWORD }}
```

* Никогда не храните креды в коде.
* Используйте robot accounts с минимальными правами (push-only).
* Для AWS ECR — используйте IAM Roles вместо кредов (через `aws-actions/configure-aws-credentials`).

### 5. Использование *BuildKit* (ускоренная сборка, параллелизм, секреты в build).

BuildKit — современный backend для `docker build`, даёт:
* Параллельную сборку стадий
* Улучшенное кэширование
* Монтирование секретов без их попадания в слой образа
* Поддержку `--cache-from`/`--cache-to`
* Multi-platform builds (`--platform linux/amd64,linux/arm64`)

Включение:
* Локально: `DOCKER_BUILDKIT=1 docker build ...`
* В CI: большинство современных образов (например, `docker:24.0`) уже используют BuildKit по умолчанию.

Секреты в build (без утечки в образ)
```dockerfile
# syntax=docker/dockerfile:1.4
FROM alpine
RUN --mount=type=secret,id=token cat /run/secrets/token
```

В CI:
```bash
docker build --secret id=token,src=./token-file ...
```

Файл `token-file` никогда не копируется в образ — он доступен только во время сборки.

Пример с GitHub Actions и BuildKit-фичами:
```yaml
- name: Set up Docker Buildx
  uses: docker/setup-buildx-action@v3

- name: Build with secrets and cache
  run: |
    docker buildx build \
      --secret id=ssh_key,src=~/.ssh/id_rsa \
      --cache-from type=registry,ref=ghcr.io/user/app:buildcache \
      --cache-to type=registry,ref=ghcr.io/user/app:buildcache,mode=max \
      --output type=image,name=ghcr.io/user/app:${{ github.sha }},push=true \
      .
```

Рекомендации для Platform Engineering
* Всегда используйте BuildKit — он стал стандартом.
* Кэшируйте агрессивно — экономия времени и ресурсов.
* Тегируйте по хешу коммита — обеспечивает иммутабельность.
* Никогда не пушьте latest в production — нарушает воспроизводимость.
* Автоматизируйте политики: например, запрещайте сборку без --no-cache в prod-ветках.
* Сканьте образы сразу после сборки (Trivy/Snyk в том же pipeline).

## Работа с registry и образами

### 1. Понимание формата образов (OCI, Docker Image Spec).

Образ — это не один файл, а набор слоёв (layers) + манифест (manifest) + конфигурация.

* Docker Image Manifest v2 Schema 2 — оригинальный формат от Docker.
* OCI (Open Container Initiative) Image Format — открытый стандарт, созданный на основе Docker v2 Schema 2. → Сегодня OCI — де-факто стандарт. Поддерживается всеми современными инструментами: Docker, Podman, containerd, Kubernetes, Helm, BuildKit и т.д.

Структура OCI-образа:
1. Layer (слои) — tar-архивы с файловой системой, каждый слой — diff по сравнению с предыдущим.
2. Image Config — JSON-файл с метаданными: `Cmd`, `Env`, `User`, `ExposedPorts`, история слоёв и т.д.
3. Manifest — указывает, какие слои использовать и ссылается на config. Может быть multi-arch (через manifest list / index).
4. Index (для multi-platform) — список манифестов под разные архитектуры (linux/amd64, linux/arm64 и др.).

`Все современные registry (Docker Hub, ECR, GCR, GHCR) совместимы с OCI. Вы можете пушить OCI-образы даже в Docker Hub.`

Инструменты для работы с OCI:
* `oras` — push/pull любых артефактов в OCI registry (не только образы).
* `crane` — CLI от Google для анализа и манипуляции образами.
* `skopeo` — копирование, инспекция, проверка образов между registry.

### 2. Push/pull из Docker Hub, GitLab Registry, AWS ECR, GCR и др.

Все registry работают по Docker Registry HTTP API v2 (совместимы с OCI), но есть нюансы авторизации.

| Registry | URL пример | Авторизация |
|--|--|--|
| Docker Hub | `docker.io/library/nginx` | `docker login` + username/password или PAT |
| GitLab Registry | `registry.gitlab.com/group/project/app` | `$CI_REGISTRY_USER` + `$CI_REGISTRY_PASSWORD` (в CI) или personal access token |
| AWS ECR | `123456789012.dkr.ecr.us-east-1.amazonaws.com/app` | IAM-роли или `aws ecr get-login-password \| docker login --password-stdin ...` |
| GCR | `gcr.io/project-id/app` | `gcloud auth configure-docker` или workload identity |
| GHCR | `ghcr.io/username/app` | `GITHUB_TOKEN` или PAT с правами `write:packages` |
| Harbor | `harbor.example.com/project/app` | LDAP/OIDC или базовая аутентификация |

AWS ECR (CLI):
```bash
# Получить токен и залогиниться
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin 123456789012.dkr.ecr.us-east-1.amazonaws.com

# Push
docker tag myapp:latest 123456789012.dkr.ecr.us-east-1.amazonaws.com/myapp:latest
docker push 123456789012.dkr.ecr.us-east-1.amazonaws.com/myapp:latest
```

GCR (с gcloud):
```bash
gcloud auth configure-docker
docker push gcr.io/my-project/myapp:latest
```

В CI используйте официальные actions или роли (например, `aws-actions/configure-aws-credentials` для ECR).

### 3. Подпись образов (Docker Content Trust, Notary).

Цель: убедиться, что образ не был изменён и поступил от доверенного источника.

**Docker Content Trust (DCT) + Notary**
* Использует Notary v1 (устаревает, но ещё работает).
* Криптографическая подпись манифеста.
* Включается через `DOCKER_CONTENT_TRUST=1`.
```bash
export DOCKER_CONTENT_TRUST=1
docker push my-registry/myapp:v1.0  # автоматически подписывает
docker pull my-registry/myapp:v1.0  # проверяет подпись
```

`Notary v1 считается legacy. Сообщество переходит на Sigstore`

**Современный стандарт: Sigstore (cosign + Fulcio + Rekor)**
* Open-source, без управления ключами (использует OIDC).
* Интеграция с Kubernetes (Kyverno, OPA/Gatekeeper), GitHub Actions.

Пример с `cosign`:
```bash
# Подписать образ
cosign sign --key cosign.key ghcr.io/user/app:sha123

# Или без ключей (через GitHub OIDC!)
cosign sign ghcr.io/user/app:sha123

# Проверить
cosign verify ghcr.io/user/app:sha123 --key cosign.pub
```

`Platform engineering рекомендация: внедряйте cosign + policy engine, чтобы запрещать запуск неподписанных образов.`

### 4. Анализ слоёв образа: `docker history`, `dive`-утилита.

Чтобы:
* Уменьшить размер образа
* Найти секреты или лишние зависимости
* Понять, почему образ "тяжёлый"

`docker history`. Показывает историю слоёв:
```bash
docker history nginx:alpine
```
Видно, какие команды создали слои и их размер. Минус: не показывает содержимое слоя.

`dive` — лучший инструмент для анализа. Устанавливается отдельно: https://github.com/wagoodman/dive
```bash
dive nginx:alpine
```

Фичи:
* Интерактивный просмотр файловой системы по слоям
* Анализ эффективности (wasted space)
* Подсветка дублирующихся/удаляемых файлов
* Оценка "score" образа

`Используйте dive при оптимизации Dockerfile`

*Альтернативы:*
* `syft` + `grype` — SBOM + сканирование уязвимостей
* `crane blob` — анализ отдельных слоёв по digest

### 5. Управление тегами и иммутабельностью (лучше хеши, чем latest).

Проблема `latest`
* `latest` — mutable тег: сегодня он может указывать на один образ, завтра — на другой.
* Нарушает воспроизводимость, traceability, безопасность.
* Не позволяет сделать надёжный rollback.

Решение: **immutable references**
* Digest (хеш манифеста): `nginx@sha256:abc123...` — уникален, неизменяем, гарантирует один и тот же образ.
* Тег по хешу коммита: `myapp:git-abc123` — привязка к коду.
* SemVer для релизов: `myapp:v1.2.3` — но только если вы никогда не перезаписываете этот тег.

Иммутабельность на уровне registry. Многие registry поддерживают immutable tags:
* GitLab: включается в настройках проекта → "Prevent tag overwrites".
* AWS ECR: lifecycle policy + запрет на `PutImage` с существующим тегом.
* Harbor: "Immutable Tag Rule".

**Best practice:**
* В CI пушьте образ с тегом по хешу коммита.
* В deployment манифестах (Helm, K8s) используйте digest, а не тег.
* Если используете теги — делайте их immutable в registry

Как получить digest после push
```bash
# Docker возвращает digest при push
docker push my-registry/myapp:abc123
# → digest: sha256:...

# Или через inspect
docker inspect my-registry/myapp:abc123 --format='{{.RepoDigests}}'
```

В GitHub Actions:
```yaml
- name: Build and output digest
  id: build
  uses: docker/build-push-action@v6
  with:
    push: true
    tags: ghcr.io/user/app:${{ github.sha }}
- run: echo "Digest: ${{ steps.build.outputs.digest }}"
```

Важно для Platform Engineer:
* Используйте OCI, забудьте про "Docker-only"
* Выбирайте с поддержкой OCI, immutable tags, scanning
* Переходите на Sigstore/cosign, а не DCT
* Встраивайте dive и trivy в CI
* Хеш коммита → тег; digest → deployment
* Запрещайте latest, требуйте подписи, сканируйте

## Отладка и диагностика

### 1. Вход в контейнер: `docker exec -it`.

Используется для интерактивного доступа к уже запущенному контейнеру.
```bash
docker exec -it <container_name_or_id> /bin/sh
```

`В минимальных образах (distroless, scratch) может не быть shell (/bin/sh).
В таком случае используйте ephemeral debug-контейнеры.`

Если нет shell:
* Kubernetes: kubectl debug --copy-to=... — создаёт копию пода с debug-контейнером.
* Локально: запустите временный контейнер в той же сети и PID namespace
```bash
docker run -it --pid=container:<target> --net=container:<target> --privileged alpine sh
```

* Избегайте exec в production без причины — это нарушает принцип immutable infrastructure.
* Используйте только для диагностики, не для "чинки на лету".

### 2. Чтение логов: `docker logs`, `--tail`, `--follow`.

Контейнеры должны писать логи в stdout/stderr — тогда Docker перехватывает их.

```bash
# Последние 100 строк
docker logs --tail 100 myapp

# Следить в реальном времени
docker logs --follow myapp

# С метками времени
docker logs --timestamps myapp

# Объединить всё
docker logs --tail 50 --follow --timestamps myapp
```

Логирование в production:
* Docker по умолчанию использует json-file driver → логи хранятся в `/var/lib/docker/containers/<id>/<id>-json.log`.
* Для централизованного сбора используйте:
  - Fluentd, Fluent Bit, Filebeat
  - Docker logging drivers: `--log-driver=fluentd`, `--log-driver=syslog`, `--log-driver=awslogs` и др.

`Best practice: Приложение пишет в stdout → Docker перехватывает → агент отправляет в Loki, ELK, CloudWatch и т.д.`

### 3. Анализ зависших контейнеров, OOM-killer.

1. Проверить статус:
```bash
docker ps -a | grep myapp
```

Возможные состояния:
* `Up` — работает
* `Exited (0)` — завершился успешно
* `Exited (137)` — убит извне (обычно OOM или `kill -9`)
* `Created` — не запустился

2. Проверить exit code
```bash
docker inspect myapp --format='{{.State.ExitCode}}'
```

3. OOM-killer (Out-Of-Memory)

* Если контейнер убит с кодом 137, это почти всегда OOM.
* Подтвердить через системные логи:
```bash
dmesg -T | grep -i "killed process"
# Или
journalctl -u docker | grep -i oom
```

4. Мониторинг ресурсов
```bash
# Текущее потребление
docker stats myapp

# Или через inspect
docker inspect myapp --format='{{.HostConfig.Memory}}'  # лимит
```

Варианты решений:
* Увеличьте `--memory` лимит.
* Оптимизируйте приложение (утечки памяти).
* Настройте liveness/readiness probes в оркестраторе (K8s), чтобы перезапускать "зависшие" поды.

### 4. Использование `docker inspect` для просмотра деталей.

`docker inspect` - мощный инструмент для получения полной информации о контейнере, образе или volume.
```bash
docker inspect myapp
```

Что можно найти:
* IP-адрес и сетевые настройки (`NetworkSettings`)
* Mount-точки (`Mounts`)
* Переменные окружения (`Config.Env`)
* Лимиты CPU/Memory (`HostConfig`)
* Состояние (`State.Running`, `State.ExitCode`)
* Health check (`State.Health`)

Фильтрация через Go-templates:
```bash
# Только IP
docker inspect myapp --format='{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}'

# Только переменные окружения
docker inspect myapp --format='{{.Config.Env}}'

# Состояние health check
docker inspect myapp --format='{{.State.Health.Status}}'
```

Инспекция образа:
```bash
docker inspect nginx:alpine --format='{{.Architecture}} {{.Os}}'
```

Используйте `jq`, если не хотите возиться с шаблонами:
```bash
docker inspect myapp | jq '.[0].NetworkSettings.IPAddress'
```

### 5. Проверка health checks и перезапуск по состоянию.

Health check в Dockerfile:
```dockerfile
HEALTHCHECK --interval=30s --timeout=3s --start-period=60s --retries=3 \
  CMD curl -f http://localhost/health || exit 1
```

* Docker периодически выполняет команду.
* Статус: `starting`, `healthy`, `unhealthy`.
* Виден в `docker ps` и `docker inspect`.

Docker сам по себе НЕ перезапускает контейнер при `unhealthy`
Это делают оркестраторы:
* Kubernetes: `livenessProbe` → перезапускает под.
* Docker Swarm: поддерживает restart по health check.
* systemd или внешние скрипты — можно написать вручную, но не рекомендуется.

Как проверить вручную:
```bash
# Посмотреть статус
docker inspect myapp --format='{{.State.Health.Status}}'

# Последний лог проверки
docker inspect myapp --format='{{json .State.Health.Log}}' | jq
```

Platform engineering:
* Всегда определяйте `HEALTHCHECK` в образах.
* В Kubernetes используйте `livenessProbe` + `readinessProbe`.
* Не полагайтесь на `docker restart` — проектируйте систему как ephemeral.

Полезные one-liners:
```
# Все unhealthy контейнеры
docker ps --filter "health=unhealthy"

# Последние 10 логов всех контейнеров
docker ps -q | xargs -I {} docker logs --tail 10 {}

# Найти контейнер по IP
docker inspect $(docker ps -q) | jq -r '.[] | select(.NetworkSettings.IPAddress == "172.17.0.3") | .Name'

# Проверить, есть ли в контейнере curl/wget
docker exec myapp which curl || echo "no curl"
```

Советы для Platform Engineer:

* **Контейнер не стартует** - `docker logs`, `docker inspect .State.Error`
* **Сетевые проблемы** - `docker exec ... nslookup`, `telnet`, `curl`; или `nsenter`
* **Файловая система повреждена** - `docker cp` для экспорта данных
* **Процесс "завис" внутри** - `docker top`, `strace` через debug-контейнер
* **Мониторинг в продакшене** - Prometheus + cAdvisor + Grafana

## Подготовка к оркестрации (Kubernetes, Swarm)

### 1. Понимание, что Docker — не оркестратор.

Что делает Docker (engine):
* Собирает образы (`docker build`)
* Запускает/останавливает контейнеры (`docker run`)
* Управляет сетью, volume’ами, образами на одном хосте

Что НЕ делает Docker:
* Не управляет множеством нод
* Не обеспечивает высокую доступность
* Не делает rolling updates, autoscaling, self-healing
* Не предоставляет service discovery, внутренний DNS, сетевые политики

Оркестраторы:
* Kubernetes — де-факто стандарт (сложный, гибкий)
* Docker Swarm — встроенный в Docker, проще, но почти не развивается
* Nomad, OpenShift, ECS — альтернативы

**Docker** — инструмент для упаковки и запуска. 
**Оркестратор** — для управления жизненным циклом в распределённой системе.

### 2. Написание «k8s-friendly» образов: без `VOLUME` в Dockerfile, stateless, health checks.

Kubernetes предъявляет особые требования к контейнерам:

**1. Избегайте `VOLUME` в Dockerfile (если не нужно)**
```dockerfile
# ПЛОХО (в большинстве случаев):
VOLUME ["/var/lib/mysql"]
```

При сборке создаётся анонимный volume, который:
* Не удаляется при удалении контейнера (в Docker)
* Ломает кэширование слоёв
* В Kubernetes игнорируется — вы всё равно монтируете `PersistentVolume` через Pod-спецификацию

Правильно:
* Храните данные вне контейнера (PV/PVC в K8s).
* В Dockerfile не указывайте `VOLUME`, если это не часть официального образа (например, `postgres`)

**2. Stateless по умолчанию**

* Контейнер не должен хранить состояние внутри.
* Все данные — во внешних хранилищах: DB, S3, Redis, PV.
* Это позволяет:
  - Безболезненно убивать и пересоздавать поды
  - Масштабировать горизонтально

**3. Обязательные health checks**

* `HEALTHCHECK` в Dockerfile → `livenessProbe` / `readinessProbe` в K8s
* Без этого Kubernetes не знает, жив ли ваш сервис

**4. Минимализм**

* Используйте `distroless` или `scratch`
* Убирайте shell, если не нужен для отладки
* Чем меньше attack surface — тем безопаснее

### 3. Использование `HEALTHCHECK` в Dockerfile.

`HEALTHCHECK` - это мост между Docker и Kubernetes.

```dockerfile
HEALTHCHECK --interval=20s --timeout=5s --start-period=60s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1
```

* В Docker: статус `healthy`/`unhealthy` (виден в `docker ps`)
* В Kubernetes: можно использовать как основу для `livenessProbe`

`Kubernetes не читает HEALTHCHECK из образа автоматически. Вы должны явно задать livenessProbe в манифесте.`

### 4. Понимание разницы между `docker run` и Pod-спецификацией.

|Фича|docker run|Kubernetes Pod|
|--|--|--|
|Сетевой namespace|Один контейнер = один IP|Все контейнеры в Pod делят один IP и порты|
|Sidecar-контейнеры|Нельзя|Можно (например, fluentd, istio-proxy)|
|Перезапуск|`--restart=always`|`restartPolicy: Always` (по умолчанию)|
|Health checks|`HEALTHCHECK`|`livenessProbe`, `readinessProbe`, `startupProbe`|
|Volumes|`-v /host:/container`|`emptyDir`, `configMap`, `secret`, `persistentVolumeClaim`|
|Переменные|`-e KEY=VAL`|`env`, `envFrom: { configMapRef, secretRef }`|
|Security|`--user`, `--cap-drop`|`securityContext` (на уровне контейнера и Pod)|
|Resource limits|`--memory=512m`|`resources.limits.memory`|

Пример миграции:

`docker run`:
```bash
docker run -d \
  --name myapp \
  --memory=512m \
  -e DB_HOST=db \
  -v ./logs:/app/logs \
  --restart=always \
  my-registry/myapp:v1
```

Kubernetes Pod:
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: myapp
spec:
  containers:
  - name: myapp
    image: my-registry/myapp:v1
    resources:
      limits:
        memory: "512Mi"
    env:
    - name: DB_HOST
      value: "db"
    volumeMounts:
    - name: logs
      mountPath: /app/logs
    livenessProbe:
      httpGet:
        path: /health
        port: 8080
      initialDelaySeconds: 60
      periodSeconds: 20
  volumes:
  - name: logs
    emptyDir: {}
  restartPolicy: Always
```

`Pod — это группа контейнеров, которые всегда запускаются вместе на одной ноде. Это фундаментальная абстракция K8s.`

5. Знание `containerd` как runtime’а под Kubernetes.

Архитектура Docker vs Kubernetes:

Docker:
```
User → Docker CLI → Docker Daemon → containerd → runc → Kernel
```

Kubernetes:
```
kubelet → CRI (Container Runtime Interface) → containerd → runc → Kernel
```

`containerd` это:
* CRI-совместимый контейнерный runtime
* Выполняет:
  - Pull образов
  - Создание/удаление контейнеров
  - Управление сетью и volume’ами (через CNI/CSI)
* Не имеет CLI для пользователя — работает «под капотом»

Почему Kubernetes использует `containerd`, а не Docker

* Docker daemon — тяжёлый, с кучей функций, не нужных в K8s (build, registry login и т.д.)
`containerd` — легковесный, безопасный, соответствует CRI
С Docker Engine v20.10+ Docker сам использует `containerd` как backend

Вывод:
* В Kubernetes Docker как runtime устарел (dockershim удалён в v1.24).
* Сегодня стандарт — `containerd` или CRI-O.

Полезные команды (на ноде с `containerd`):
```bash
# Посмотреть образы
crictl images

# Посмотреть поды
crictl pods

# Посмотреть логи контейнера
crictl logs <container-id>

# Pull образа
crictl pull nginx
```

`crictl` — аналог docker для CRI-совместимых runtimes.

### Вывод

* Docker — для сборки и локального запуска; K8s — для управления кластером
* Stateless, без VOLUME, с health checks
* Полезен, но в K8s нужно дублировать в livenessProbe
* Pod — группа контейнеров с общей сетью и storage
* Современный runtime для K8s; Docker больше не используется напрямую

## Инструменты

* **BuildKit** — современный builder с кэшированием, параллелизмом, секретами.
* **Dive** — анализ слоёв образа.
* **Trivy / Grype / Snyk** — сканирование уязвимостей.
* **Podman** — альтернатива без демона (rootless by default).
* **docker scan** — встроенный анализ безопасности.
