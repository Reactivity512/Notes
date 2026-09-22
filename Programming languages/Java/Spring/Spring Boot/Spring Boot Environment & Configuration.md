# Spring Boot

## Часть 2. Environment & Configuration

### 2.1. Что такое `Environment` и зачем он нужен

`Environment` — это центральный объект Spring, который отвечает за две вещи:
1. Property sources — откуда читаются настройки (файлы, env vars, args).
2. Profiles — какие профили активны.

Иерархия интерфейсов:
```declarative
Environment
  └─ ConfigurableEnvironment
        ├─ StandardEnvironment                   (NONE)
        ├─ StandardServletEnvironment            (SERVLET)
        └─ StandardReactiveWebEnvironment        (REACTIVE)
```

`StandardServletEnvironment` дополнительно подмешивает `ServletConfig` и `ServletContext` как property sources.

### 2.2. Создание Environment — `getOrCreateEnvironment()`

```java
private ConfigurableEnvironment getOrCreateEnvironment() {
    if (this.environment != null) {
        return this.environment;
    }
    return this.applicationContextFactory.createEnvironment(
        this.webApplicationType);
}
```

В ApplicationContextFactory (Boot 3):
```java
case SERVLET:
    return new StandardServletEnvironment();
case REACTIVE:
    return new StandardReactiveWebEnvironment();
default:
    return new StandardEnvironment();
```

**Boot 4**: фабрика удалена, Environment создаётся модулем (`spring-boot-webmvc` → `StandardServletEnvironment`).

### 2.3. `MutablePropertySources` — порядок приоритетов

Внутри `Environment` лежит `MutablePropertySources` — упорядоченный список `PropertySource`. **Побеждает тот, что выше.**

Порядок для `StandardServletEnvironment` (сверху = высший приоритет):


| # | PropertySource | Пример
|--|--|--
| 1 | `commandLineArgs` | `--server.port=9090`
| 2 | `servletConfigInitParams` | `<init-param>`
| 3 | `servletContextInitParams` | `<context-param>`
| 4 | `systemProperties` | `-Dserver.port=9090`
| 5 | `systemEnvironment` | `SERVER_PORT=9090`
| 6 | `random` | `random.int`, `random.uuid`
| 7 | Config resource 'application.yml' | файл конфигурации
| 8 | `defaultProperties` | `SpringApplication.setDefaultProperties()`

**Ключевой момент**: `@PropertySource`, `application-{profile}.yml`, `spring.config.import` добавляются между системными и default, каждый со своим индексом.

Проверить можно так:
```java
@Autowired
private ConfigurableEnvironment env;

env.getPropertySources().forEach(ps -> 
    System.out.println(ps.getName()));
```

### 2.4. `configureEnvironment()` — что происходит до чтения файлов

```java
protected void configureEnvironment(ConfigurableEnvironment environment, 
        String[] args) {
    if (this.addCommandLineProperties) {
        addCommandLineProperties(environment, args);   // CommandLinePropertySource
    }
    configureProfiles(environment, args);              // active profiles
}
```

Здесь:
* `SimpleCommandLinePropertySource` парсит `--key=value`.
* Профили читаются из `spring.profiles.active` (если переданы в args или системных свойствах) до загрузки файлов.

### 2.5. `EnvironmentPostProcessor` — точка расширения
Ключевой SPI. Загружается из `META-INF/spring.factories`:
```java
org.springframework.boot.env.EnvironmentPostProcessor=\
org.springframework.boot.context.config.ConfigDataEnvironmentPostProcessor,\
org.springframework.boot.env.RandomValuePropertySourceEnvironmentPostProcessor,\
...
```

Интерфейс:
```java
public interface EnvironmentPostProcessor {
    void postProcessEnvironment(ConfigurableEnvironment environment, 
        SpringApplication application);
}
```

Порядок через `@Order`. Вызывается в `listeners.environmentPrepared()`, то есть до создания ApplicationContext.

Свой пример — добавить свой property source из БД:
```java
public class DbPropertySourcePostProcessor implements EnvironmentPostProcessor {
    @Override
    public void postProcessEnvironment(ConfigurableEnvironment env, 
            SpringApplication app) {
        env.getPropertySources().addLast(
            new MapPropertySource("dbConfig", loadFromDb()));
    }
}
```

Регистрация:
```declarative
# META-INF/spring.factories
org.springframework.boot.env.EnvironmentPostProcessor=\
com.example.DbPropertySourcePostProcessor
```

### 2.6. Загрузка конфигурационных файлов (Boot 2.4+)

Это самая «мясная» часть. С версии 2.4 полностью переписан механизм — появился `ConfigDataEnvironmentPostProcessor`.

Цепочка
```declarative
ConfigDataEnvironmentPostProcessor.postProcessEnvironment()
  └─ ConfigDataEnvironment.processAndApply()
        ├─ ConfigDataLocationResolver   (для каждого location)
        │     └─ StandardConfigDataLocationResolver
        ├─ ConfigDataLoader             (для каждого ресурса)
        │     └─ StandardConfigDataLoader → PropertySource
        └─ Environment.getPropertySources().addXxx(...)
```

Ключевые классы:
* `ConfigDataEnvironment` — оркестратор.
* `ConfigDataLocationResolver` — превращает строку `file:./config/` в `ConfigDataResource`.
* `ConfigDataLoader` — загружает `ConfigDataResource` в `ConfigData` (набор PropertySource).
* `ConfigData` — immutable-контейнер: имя + `PropertySource`'ы.

**Что ищется по умолчанию**
Стандартные locations (в порядке возрастания приоритета):
1. `classpath:/`
2. `classpath:/config/`
3. `file:./`
4. `file:./config/`
5. `file:./config/*/`

Файлы: `application.properties` / `application.yml` / `application.yaml`.

Плюс profile-specific: `application-{profile}.yml`.

`spring.config.*`

| Свойство | Что делает
|--|--
| `spring.config.name` | базовое имя (по умолчанию `application`)
| `spring.config.location` | заменяет дефолтные locations
| `spring.config.additional-location` | добавляет к дефолтным
| `spring.config.import` | импорт других источников
| `spring.config.on-not-found` | `fail` / `ignore`

Пример:
```yaml
spring:
  config:
    import:
      - optional:file:./external.yml
      - configtree:/run/secrets/
```

Префиксы
* `optional:` — не падать, если нет.
* `file:` / `classpath:` — явный протокол.
* `configtree:` — каталог с файлами, каждый файл = property (имя файла = ключ).

YAML multi-document
```yaml
# application.yml
server:
  port: 8080
---
spring:
  config:
    activate:
      on-profile: prod
server:
  port: 9090
```

Второй документ активируется только в профиле `prod`.

Boot 4: весь механизм остался, но модульный: `spring-boot-config-data` — отдельный JAR.

### 2.7. Профили (Profiles)

Активация
Способы (по убыванию приоритета):
1. `--spring.profiles.active=dev,metrics` (args)
2. `-Dspring.profiles.active=dev` (system prop)
3. `SPRING_PROFILES_ACTIVE=dev` (env)
4. `spring.profiles.active` в `application.yml`
5. `SpringApplication.setAdditionalProfiles("dev")` (программно)

Группы профилей (2.4+)
```yaml
spring:
  profiles:
    group:
      "prod": "proddb,prodmq"
      "dev":  "devdb,devmq"
```

Активируешь `prod` → включаются `proddb`, `prodmq`.

`include` / `default`
```yaml
spring:
  profiles:
    include: common
    default: local
```

* `include` — добавить профили к активным.
* `default` — если ничего не активировано.

Проверки в коде
```java
@Profile("dev")
@Configuration
class DevConfig {}

@Autowired
private Environment env;

if (env.acceptsProfiles(Profiles.of("prod & !test"))) { ... }
```

`Profiles.of()` — expression-синтаксис: `&` (and), `|` (or), `!` (not).

### 2.8. `@ConfigurationProperties` — типизированный доступ

Базовый пример
```java
@ConfigurationProperties(prefix = "app.mail")
@Validated
public record MailProperties(
    @NotBlank String host,
    @Min(1) @Max(65535) int port,
    Duration timeout,
    List<String> recipients
) {}
```

Активация:
```java
@SpringBootApplication
@ConfigurationPropertiesScan   // ← сканирует все @ConfigurationProperties
public class App {}
```

Или точечно:
```java
@EnableConfigurationProperties(MailProperties.class)
```

Как работает биндинг
Цепочка:
```java
ConfigurationPropertiesBindingPostProcessor   (BeanPostProcessor)
  └─ ConfigurationPropertiesBinder
        ├─ Binder                    (Spring Framework)
        ├─ PropertySources           ← из Environment
        ├─ ConversionService         ← конвертация типов
        └─ Validator                 ← JSR-380
```

`ConfigurationPropertiesBindingPostProcessor` регистрируется автоматически через `ConfigurationPropertiesAutoConfiguration`.

Релаксация имён
Spring Boot матчит ключи в любом стиле:

| Формат в файле | Матчится на поле
|--|--
| `app.mail.host-name` | `hostName`
| `app.mail.host_name` | `hostName`
| `app.mail.hostName` | `hostName`
| `APP_MAIL_HOSTNAME` | `hostname`

Канонический формат — **kebab-case.**

`@ConstructorBinding` — эволюция
* Boot 2.x — требовался для immutable (constructor-binding) бинов.
* Boot 3.0+ — не нужен для record'ов и классов с единственным конструктором.
* Boot 3.x — остался для случаев, когда у класса несколько конструкторов.
* Boot 4 — фактически legacy, используется редко.

Валидация
```java
@ConfigurationProperties(prefix = "app.mail")
@Validated
public class MailProperties {
    @NotBlank private String host;
    @Min(1) private int port;
    // getters/setters
}
```

При старте, если валидация падает → `BindValidationException` → `ApplicationFailedEvent`.

Разрешение сложных типов
* `Duration` — `10s`, `5m`, `1h`.
* `DataSize` — `10MB`, `512KB`.
* `List<String>` — `app.mail.recipients=a,b,c` или YAML-список.
* `Map<String, X>` — вложенные ключи.
* Enum — по имени (case-insensitive).

**Boot 3 vs Boot 4**

| Аспект | Boot 3 | Boot 4
|--|--|--
| `@ConstructorBinding` | Опционально | Legacy
| Модуль | `spring-boot` | `spring-boot-configuration-properties` (выделен)
| Record-binding | Поддерживается | Поддерживается
| AOT-обработка | Есть | Улучшена (reflection-free по умолчанию)


### 2.9. Диаграмма: путь property от файла до бина

```java
application.yml
      │
      ▼
StandardConfigDataLoader
      │
      ▼
ConfigData (PropertySource "Config resource '...'")
      │
      ▼
MutablePropertySources.addLast()        ← внутри Environment
      │
      ▼
Binder.bind(prefix, targetClass)
      │  (через ConfigurationPropertiesBinder)
      ▼
MailProperties (bean)
```

### 2.10. Ключевые моменты

1. `Environment` создаётся до `ApplicationContext` — поэтому `EnvironmentPostProcessor` может всё.
2. `ConfigDataEnvironmentPostProcessor` — входная точка загрузки `application.yml` (с Boot 2.4).
3. Порядок property sources решает всё — выше в списке = побеждает.
4. `spring.config.import` — единственный правильный способ подключать внешние конфиги в Boot 3/4.
5. `@ConfigurationProperties` + record — современный стандарт, `@ConstructorBinding` больше не нужен.
6. Профили-группы (`spring.profiles.group.*`) — удобно для композиции.
7. Boot 4: конфиг-механизм вынесен в отдельный модуль `spring-boot-config-data`
