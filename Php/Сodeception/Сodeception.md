# Codeception


Codeception работает только на PHP версии 7.1 или выше, и для него необходим **Composer**. Codeception инициируется командой `codecept bootstrap`

Основная структура папок:

```
tests/
_data/         - тестовые данные
_output/       - отчеты и логи
_support/      - вспомогательные классы
acceptance/    - acceptance-тесты
functional/    - functional-тесты
unit/          - unit-тесты
codeception.yml - основной конфигурационный файл
```



Основной конфигурационный файл `codeception.yml`:

```
actor_suffix: Tester
paths:
tests: tests
output: tests/_output
data: tests/_data
support: tests/_support
envs: tests/_envs
settings:
bootstrap: _bootstrap.php
colors: true
memory_limit: 1024M
extensions:
enabled:
\- Codeception\\Extension\\RunFailed
```

Для разных типов тестов (*acceptance*, *functional*, *unit*, *api*) есть отдельные конфигурационные файлы в соответствующих папках.

Сначала необходимо создать suite: `codecept generate:suite acceptance` или `codecept generate:suite api`.

Будет сгенерирован файл `tests/acceptance.suite.yml` с примерным содержанием:

```
actor: AcceptanceTester
modules:
enabled:
\- PhpBrowser:
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;url: http://localhost/myapp/
\- \\Helper\\Acceptance
```

**Suite** (сьют) в Codeception — это логическая группа тестов, объединенных по типу тестирования или функциональности. Это основная организационная единица тестов в Codeception: *unit*, *functional*, *acceptance*, *api*.

**1. Unit-тесты (Модульные тесты)**

* Проверяют отдельные классы и методы в изоляции.
* Не требуют окружения (базы данных, HTTP-запросов)

Нужны для:

* Для тестирования отдельных классов и методов
* Для проверки бизнес-логики
* Для тестирования утилитарных функций

Создание unit теста: `codecept generate:test unit Example`

**2. Functional-тесты (Функциональные тесты)**

* Тестируют контроллеры и middleware
* Эмулируют HTTP-запросы без реального веб-сервера
* Используют фреймворк приложения

Нужны для:

* Для тестирования MVC-контроллеров
* Для проверки маршрутизации
* Для тестирования форм и валидации

Создание functional-теста: `codecept generate:cest functional Example`

**3. Acceptance-тесты (Приемочные тесты)**

* Тестируют приложение через реальные HTTP-запросы
* Могут использовать браузер (через Selenium/WebDriver)
* Максимально приближены к действиям пользователя
* Самые медленные тесты

Нужны для:

* Для проверки полного сценария использования
* Для end-to-end тестирования
* Для тестирования UI и JavaScript

Создание Acceptance-теста: `codecept generate:cest acceptance Example`

**4.API-тесты**

* Тестируют REST, GraphQL, SOAP и другие API
* Работают с JSON/XML ответами
* Могут проверять структуру ответов

Нужны для:

* Для тестирования API endpoints
* Для проверки мобильного бэкенда
* Для тестирования микросервисов

Создание API-теста: `codecept generate:cest api Example`

Лучшая практика - комбинировать все типы тестов в пропорции:

* 70% unit-тестов
* 15% functional-тестов
* 10% API-тестов
* 5% acceptance-тестов

Запуск конкретного **suite**:

`codecept run unit`&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\# Все unit-тесты

`codecept run acceptance`&nbsp;&nbsp;&nbsp;&nbsp;\# Все acceptance-тесты

`codecept run functional`&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\# Все Функциональные тесты

`codecept run api`&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\# Все API-тесты

---

`codecept run` \# Запуск всех тестов

`codecept run tests/acceptance/LoginCest.php` \# Запуск конкретного теста

`codecept run --html`   \# Запуск с генерацией отчета

**Основные форматы отчетов:**

1. HTML-отчет:

   `codecept run --html`

   Создает отчет в *tests/_output/report.html*

2. XML-отчет (JUnit style):

   `codecept run --xml`

   Генерирует *tests/_output/report.xml* для интеграции с CI-системами

3. JSON-отчет:

   `codecept run --json`

   Создает *tests/_output/report.json* для машинной обработки

**Анализ покрытия кода тестами:**

1. Необходимо добавить в codeception.yml:

```
coverage:
    enabled: true
    include:
        - src/*
    exclude:
        - src/config/*
    c3_url: http://your-app.test
```

2. Запуск с анализом покрытия (Локальный запуск):

* `codecept run --coverage`
* `codecept run --coverage-html`  \# *HTML-отчет*
* `codecept run --coverage-xml`   \# *XML для CI*


HTML-отчет будет доступен в *tests/_output/coverage*
