# Spring Boot

## Часть 9. Инструменты и фичи

Здесь собраны инструменты, которые не входят в ядро работы приложения, но без которых не обходится ни один production-проект: мониторинг, логирование, ускорение разработки, тестирование, сборка и нативная компиляция.

### 9.1. Actuator — production-ready мониторинг

Actuator добавляет в приложение готовые HTTP/JMX-эндпоинты для мониторинга и управления.

**9.1.1. Подключение**
```xml
<dependency>
    <groupId>org.springframework.boot</groupId>
    <artifactId>spring-boot-starter-actuator</artifactId>
</dependency>
```

По умолчанию доступны только `/actuator/health` и `/actuator/info`. Остальные включаются через `management.endpoints.web.exposure.include`:

```yaml
management:
  endpoints:
    web:
      exposure:
        include: health,info,metrics,env,beans
```

**9.1.2. Встроенные эндпоинты**

| ID | Что показывает
|--|--
| `health` | Состояние приложения (UP/DOWN)
| `info` | Произвольная информация (`info.*` properties)
| `metrics` | Метрики Micrometer (JVM, HTTP, DataSource)
| `env` | Все PropertySource
| `beans` | Полный список бинов
| `mappings` | Все `@RequestMapping` пути
| `configprops` | Все `@ConfigurationProperties`
| `threaddump` | Thread dump JVM
| `heapdump` | Heap dump (hprof)
| `loggers` | Просмотр/изменение уровней логирования
| `shutdown` | Корректное завершение работы (graceful shutdown) — удалено в Boot 3.4; используйте вместо этого `server.shutdown=graceful`.

**9.1.3. Свой `@Endpoint`**

```java
@Component
@Endpoint(id = "feature-flags")
public class FeatureFlagsEndpoint {

    @ReadOperation
    public Map<String, Boolean> getFlags() {
        return Map.of("newCheckout", true, "darkMode", false);
    }

    @WriteOperation
    public void setFlag(@Selector String name, boolean value) {
        // установка флага
    }
}
```

- `@ReadOperation` → HTTP GET (доступен http://localhost:8080/actuator/featureFlags)
- `@WriteOperation` → HTTP POST
- `@DeleteOperation` → HTTP DELETE

**9.1.4. HealthIndicators**

```java
@Component
public class ExternalServiceHealthIndicator implements HealthIndicator {
    @Override
    public Health health() {
        if (isServiceUp()) {
            return Health.up().withDetail("responseTime", "45ms").build();
        }
        return Health.down().withDetail("error", "Connection refused").build();
    }
}
```

Actuator собирает все `HealthIndicator`-бины и агрегирует их в общий статус `/actuator/health`. Статус может быть: `UP`, `DOWN`, `OUT_OF_SERVICE`, `UNKNOWN`. Общий статус — «худший» из всех.

**9.1.5. Micrometer**

Micrometer — фасад метрик. Actuator автоматически регистрирует `MeterRegistry` и собирает JVM-метрики (heap, threads, GC).

```java
@Component
public class OrderMetrics {
    private final Counter orderCounter;
    private final Timer orderTimer;

    public OrderMetrics(MeterRegistry registry) {
        this.orderCounter = Counter.builder("orders.created")
            .description("Количество созданных заказов")
            .register(registry);
        this.orderTimer = Timer.builder("orders.processing.time")
            .register(registry);
    }

    public void onOrderCreated() { orderCounter.increment(); }
    public void recordTime(Runnable task) { orderTimer.record(task); }
}
```

Интеграция с Prometheus:
```xml
<dependency>
    <groupId>io.micrometer</groupId>
    <artifactId>micrometer-registry-prometheus</artifactId>
</dependency>
```

Эндпоинт `/actuator/prometheus` отдаёт метрики в формате Prometheus.

**9.1.6. Boot 4: изменения в Actuator**

- **Пакеты переехали:** `org.springframework.boot.actuate.health.Health` → `org.springframework.boot.health.contributor.Health`.

- **Модуляризация:** Actuator разбит на множество модулей (actuator-web, actuator-jmx, actuator-sbom и т.д.).

- `/actuator/info` расширен: добавлена информация о процессе.

### 9.2. Logging

**9.2.1. Дефолты**

Spring Boot использует SLF4J как фасад и Logback как реализацию. При старте `LoggingApplicationListener` (из `spring.factories`) инициализирует логирование раньше, чем создаётся `ApplicationContext`.

**9.2.2. Конфигурация**

| Система | Файл
|--|--
| Logback | `logback-spring.xml` (рекомендуется), `logback.xml`
| Log4j2 | `log4j2-spring.xml`, `log4j2.xml`
| JUL | `logging.properties`

Рекомендуется `-spring`-варианты: они позволяют использовать `<springProfile>` и `<springProperty>`.

```xml
<!-- logback-spring.xml -->
<configuration>
    <springProfile name="dev">
        <appender name="CONSOLE" class="ch.qos.logback.core.ConsoleAppender">
            <encoder>
                <pattern>%d{HH:mm:ss.SSS} [%thread] %-5level %logger{36} - %msg%n</pattern>
            </encoder>
        </appender>
    </springProfile>

    <springProfile name="prod">
        <appender name="FILE" class="ch.qos.logback.core.rolling.RollingFileAppender">
            <file>${LOG_FILE}</file>
            <encoder>
                <pattern>%d{yyyy-MM-dd HH:mm:ss} [%thread] %-5level %logger - %msg%n</pattern>
            </encoder>
        </appender>
    </springProfile>
</configuration>
```

**9.2.3. Уровни через properties**

```yaml
logging:
  level:
    root: INFO
    com.example.myapp: DEBUG
    org.springframework.web: WARN
```

**9.2.4. Структурированное логирование (Boot 3.4+)**

Spring Boot 3.4 добавил нативную поддержку structured logging без дополнительных зависимостей:

```yaml
logging:
  structured:
    format:
      console: ecs   # Elastic Common Schema
      file: logstash
```

Форматы: `ecs`, `logstash`, `gelf`.

**9.2.5. Boot 4: изменения**

- Кодировка по умолчанию — UTF-8 для Logback, как в Log4j2.
- Нет автоматической корреляции trace/span ID — нужно настраивать appender вручную.
- Logback обновлён до новой версии, некоторые appenders удалены.

### 9.3. DevTools — ускорение разработки

```xml
<dependency>
    <groupId>org.springframework.boot</groupId>
    <artifactId>spring-boot-devtools</artifactId>
    <optional>true</optional>
</dependency>
```

**9.3.1. Автоматический перезапуск**

DevTools использует два ClassLoader:

| ClassLoader | Что загружает
|--|--
| Base ClassLoader | Неизменяемые JAR'ы (Spring, сторонние библиотеки)
| Restart ClassLoader | Классы твоего приложения (часто меняются)

При изменении файла в classpath DevTools пересоздаёт только Restart ClassLoader, а Base остаётся. Это делает перезапуск в разы быстрее полного.

**9.3.2. LiveReload**

Встроенный LiveReload-сервер автоматически обновляет браузер при изменении ресурсов (HTML, CSS, JS). Требует расширения LiveReload в браузере. Отключается через `spring.devtools.livereload.enabled=false`.

**9.3.3. Настройка include/exclude**

```properties
# META-INF/spring-devtools.properties
restart.exclude.companycommonlibs=/mycorp-common-[\\w-]+\.jar
restart.include.projectcommon=/mycorp-myproj-[\\w-]+\.jar
```

`exclude` — JAR'ы, которые загружаются Base ClassLoader'ом. `include` — наоборот, Restart ClassLoader'ом.

**9.3.4. Глобальные настройки**

Файл `.spring-boot-devtools.properties` в `$HOME` применяется ко всем проектам.

> **Важно:** DevTools не должен попадать в production. optional=true + Maven/Gradle исключают его из финального JAR.

### 9.4. Тестирование

**9.4.1. `@SpringBootTest` — полный контекст**

```java
@SpringBootTest(webEnvironment = WebEnvironment.RANDOM_PORT)
class MyApplicationTests {

    @Autowired
    private TestRestTemplate restTemplate;

    @Test
    void contextLoads() {
        ResponseEntity<String> response = restTemplate.getForEntity("/api/hello", String.class);
        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
    }
}
```

`webEnvironment`:
- `MOCK` (по умолчанию) — без реального сервера, `MockMvc`
- `RANDOM_PORT` — реальный сервер на случайном порту
- `DEFINED_PORT` — на порту из конфигурации
- `NONE` — без веб-окружения

**9.4.2. Slice-тесты — «тонкие» контексты**

| Аннотация | Что загружает | Что мокать
|--|--|--
| `@WebMvcTest` | Только MVC: контроллеры, `@ControllerAdvice`, фильтры | Сервисы через `@MockitoBean`
| `@DataJpaTest` | Только JPA: репозитории, `@Entity`, DataSource | Обычно H2 in-memory
| `@JsonTest` | Только Jackson/Gson | —
| `@RestClientTest` | Только REST-клиент (`RestTemplate`, `WebClient`) | Сервер через MockWebServer

```java
@WebMvcTest(UserController.class)
class UserControllerTest {

    @Autowired
    private MockMvc mockMvc;

    @MockitoBean
    private UserService userService;

    @Test
    void shouldReturnUser() throws Exception {
        when(userService.findById(1L)).thenReturn(new User(1L, "Alice"));

        mockMvc.perform(get("/api/users/1"))
            .andExpect(status().isOk())
            .andExpect(jsonPath("$.name").value("Alice"));
    }
}
```

**Ключевое преимущество slice-тестов:** контекст загружается быстрее, потому что включается только нужная часть автоконфигурации.

**9.4.3. `@MockitoBean` — замена `@MockBean` в Boot 4**

Критическое breaking change в Boot 4:

```java
// Boot 3.x (deprecated в 3.4, удалено в 4.0)
@MockBean
private UserService userService;

// Boot 4.x
import org.springframework.test.context.bean.override.mockito.MockitoBean;
@MockitoBean
private UserService userService;
```

`@MockBean` и `@SpyBean` удалены в Boot 4. Вместо них — `@MockitoBean` и `@MockitoSpyBean` из Spring Framework.

**9.4.4. `ApplicationContextRunner` — тестирование автоконфигураций**

Для тестирования своих автоконфигураций:
```java
class MyAutoConfigurationTests {

    private final ApplicationContextRunner contextRunner = new ApplicationContextRunner()
        .withConfiguration(AutoConfigurations.of(MyServiceAutoConfiguration.class))
        .withPropertyValues("my.service.endpoint=http://localhost:9090");

    @Test
    void shouldCreateMyService() {
        contextRunner.run(context -> {
            assertThat(context).hasSingleBean(MyService.class);
        });
    }

    @Test
    void shouldBackOffWhenCustomBeanPresent() {
        contextRunner
            .withUserConfiguration(CustomConfig.class)
            .run(context -> {
                assertThat(context).hasSingleBean(MyService.class);
                assertThat(context.getBean(MyService.class)).isInstanceOf(CustomMyService.class);
            });
    }
}
```

`ApplicationContextRunner` позволяет изолированно тестировать условия автоконфигурации без загрузки всего приложения.

### 9.5. Сборка и упаковка

**9.5.1. `spring-boot-maven-plugin`**

```xml
<build>
    <plugins>
        <plugin>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-maven-plugin</artifactId>
        </plugin>
    </plugins>
</build>
```

Goal `repackage` создаёт executable JAR (fat jar):

```bash
mvn clean package
java -jar target/myapp-0.0.1-SNAPSHOT.jar
```

**9.5.2. Layered JAR**

Начиная с Boot 2.3, JAR можно разделить на слои для эффективного Docker-кэширования:

```xml
<plugin>
    <groupId>org.springframework.boot</groupId>
    <artifactId>spring-boot-maven-plugin</artifactId>
    <configuration>
        <layers>
            <enabled>true</enabled>
        </layers>
    </configuration>
</plugin>
```

Слои (от редко меняющихся к часто меняющимся):

1. `dependencies` — стабильные зависимости
2. `spring-boot-loader` — загрузчик
3. `snapshot-dependencies` — snapshot'ы
4. `application` — твой код

Dockerfile:
```dockerfile
FROM eclipse-temurin:21-jre AS builder
WORKDIR /app
COPY target/myapp-*.jar app.jar
RUN java -Djarmode=tools -jar app.jar extract --layers --launcher

FROM eclipse-temurin:21-jre
WORKDIR /app
COPY --from=builder /app/dependencies/ ./
COPY --from=builder /app/spring-boot-loader/ ./
COPY --from=builder /app/snapshot-dependencies/ ./
COPY --from=builder /app/application/ ./
ENTRYPOINT ["java", "org.springframework.boot.loader.launch.JarLauncher"]
```

**9.5.3. `JarLauncher`**

`JarLauncher` — это точка входа fat jar'а. Он находится в `spring-boot-loader` и отвечает за:

1. Чтение `MANIFEST.MF` (`Main-Class: org.springframework.boot.loader.launch.JarLauncher`)
2. Построение `URLClassLoader` со всеми вложенными JAR'ами
3. Делегирование `Start-Class` (твой `main`-класс)

Также есть `WarLauncher` и `PropertiesLauncher` (для внешних конфигураций).

**9.5.4. Buildpacks**
```bash
mvn spring-boot:build-image
```

Spring Boot использует Paketo Buildpacks для создания OCI-образа без Dockerfile:
```xml
<plugin>
    <groupId>org.springframework.boot</groupId>
    <artifactId>spring-boot-maven-plugin</artifactId>
    <configuration>
        <image>
            <name>myregistry/myapp:${project.version}</name>
        </image>
    </configuration>
</plugin>
```

**9.5.5. Boot 4: изменения**

- Starter'ы переименованы: `spring-boot-starter-web` →` spring-boot-starter-webmvc`.
- Undertow удалён (несовместим с Servlet 6.1).
- `spring-boot-loader` остался, но пакеты переехали.

### 9.6. AOT и GraalVM Native Image (обзорно)

**9.6.1. Что такое AOT**

Ahead-of-Time — обработка приложения во время сборки, а не во время запуска. Spring Boot 3+ выполняет AOT-обработку, которая:

1. Сканирует все `@Configuration`, `@Bean`, `@Component`.
2. Генерирует Java-код для регистрации бинов (вместо reflection).
3. Регистрирует `RuntimeHints` — подсказки для GraalVM.

Зачем: GraalVM Native Image — компиляция Java в нативный исполняемый файл. Reflection, dynamic proxies и загрузка ресурсов в native image работают только если о них сообщить на этапе сборки.

**9.6.2. `RuntimeHintsRegistrar`**

```java
public class MyRuntimeHints implements RuntimeHintsRegistrar {
    @Override
    public void registerHints(RuntimeHints hints, ClassLoader classLoader) {
        hints.reflection().registerType(MyDto.class, MemberCategory.values());
        hints.resources().registerPattern("my-config/*.json");
    }
}
```

Регистрация:
```java
@ImportRuntimeHints(MyRuntimeHints.class)
@Configuration
public class MyConfig { }
```

**9.6.3. `@RegisterReflectionForBinding` — упрощение**

```java
@Configuration
@RegisterReflectionForBinding({User.class, UserDto.class})
public class JacksonConfig { }
```

Аннотация регистрирует reflection-подсказки для DTO, которые сериализуются Jackson'ом.

**9.6.4. Сборка native image**
```bash
mvn -Pnative native:compile
```

Результат — нативный исполняемый файл:
```bash
./target/myapp
```

Преимущества:
- Мгновенный старт (миллисекунды вместо секунд)
- Меньшее потребление памяти
- Меньший размер образа

Ограничения:
- Нет динамической загрузки классов
- Нет JIT (пиковая производительность ниже)
- Сборка дольше
- Не все библиотеки поддерживаются

**9.6.5. Boot 4: AOT по умолчанию**

В Boot 4 AOT-обработка более агрессивна: reflection-free по умолчанию для большинства сценариев. `BeanRegistrar` (появившийся в Spring Framework 7) — основной способ регистрации бинов для AOT, заменяющий reflection-based `@Bean`-методы.

### 9.7. Сводная таблица: Boot 3 vs Boot 4

| Аспект | Spring Boot 3.x | Spring Boot 4.x
|--|--|--
| Actuator пакеты | `org.springframework.boot.actuate.health` | `org.springframework.boot.health.contributor`
| Actuator модули | Единый `spring-boot-actuator` | Множество модулей (web, jmx, sbom...)
| Logging по умолчанию | Logback, charset зависит от ОС | Logback, UTF-8 всегда
| Structured logging | Появился в 3.4 | Полностью встроен
| Тестовые аннотации | `@MockBean`, `@SpyBean` | Удалены, вместо них `@MockitoBean`, `@MockitoSpyBean`
| Starter'ы | `spring-boot-starter-web` | `spring-boot-starter-webmvc`
| AOT | Opt-in, `-Pnative` | Более агрессивный, reflection-free
| Buildpacks | Поддерживаются | Поддерживаются, улучшены
| Layered JAR | Есть | Есть, улучшены

### 9.8. Что запомнить

1. **Actuator** — не только `/health`. Свои эндпоинты через `@Endpoint` + `@ReadOperation`/`@WriteOperation`. `HealthIndicator` — для проверки внешних сервисов. Micrometer — для метрик.
2. **Logging** инициализируется до `ApplicationContext`. Используй `logback-spring.xml`, не `logback.xml`. С Boot 3.4 — structured logging из коробки.
3. **DevTools** использует два ClassLoader: Base (стабильные JAR'ы) и Restart (твой код). Перезапускает только второй.
4. **Тестирование:** `@SpringBootTest` — полный контекст, slice-тесты — тонкие контексты. `ApplicationContextRunner` — для тестирования автоконфигураций.
5. `@MockBean` → `@MockitoBean` — критическое breaking change в Boot 4.
6. **Layered JAR** — разделение на слои для Docker-кэширования. `JarLauncher` — точка входа fat jar'а.
7. **Buildpacks** — OCI-образ без Dockerfile: `mvn spring-boot:build-image`.
8. **AOT** — обработка во время сборки для GraalVM. `@RegisterReflectionForBinding` — для DTO. Native image: мгновенный старт, но ограничения.
