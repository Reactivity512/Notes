# Spring Boot

## Часть 8. Старт и остановка приложения

В части 1 мы остановились на том, что `SpringApplication.run()` вызывает `callRunners()` и публикует `ApplicationReadyEvent`. Теперь разберём эти шаги детально, а также полный цикл остановки приложения — от JVM shutdown hook до `@PreDestroy`.

### 8.1. `ApplicationRunner` и `CommandLineRunner` — точка входа в «готовое» приложение

После того как контекст полностью поднят (все бины созданы, Tomcat запущен), Spring Boot вызывает бины, реализующие `ApplicationRunner` или `CommandLineRunner`.

```java
@FunctionalInterface
public interface ApplicationRunner {
    void run(ApplicationArguments args) throws Exception;
}

@FunctionalInterface
public interface CommandLineRunner {
    void run(String... args) throws Exception;
}
```

Ключевое отличие — формат аргументов:

| Интерфейс | Параметр | Пример доступа
|--|--|--
| `CommandLineRunner` | `String[]` (raw) | `args[0]` → `--server.port=9090`
| `ApplicationRunner`| `ApplicationArguments` | `args.getOptionNames()` → `[server.port]`

`ApplicationArguments` — это уже распарсенный объект, который Spring Boot создаёт из `String[]`:

```java
DefaultApplicationArguments args = new DefaultApplicationArguments(args);
// args.getOptionNames() → Set<String>
// args.getOptionValues("server.port") → List<String>
// args.getSourceArgs() → String[]
// args.getNonOptionArgs() → List<String>
```

### 8.2. Порядок вызова runners

В `SpringApplication.callRunners()`:

```java
private void callRunners(ApplicationContext context, ApplicationArguments args) {
    List<Object> runners = new ArrayList<>();
    runners.addAll(context.getBeansOfType(ApplicationRunner.class).values());
    runners.addAll(context.getBeansOfType(CommandLineRunner.class).values());
    AnnotationAwareOrderComparator.sort(runners);
    for (Object runner : new LinkedHashSet<>(runners)) {
        if (runner instanceof ApplicationRunner applicationRunner) {
            callRunner(applicationRunner, args);
        }
        if (runner instanceof CommandLineRunner commandLineRunner) {
            callRunner(commandLineRunner, args);
        }
    }
}
```

Важные детали:
1. Оба типа собираются в один список.
2. Сортировка через `AnnotationAwareOrderComparator` — по `@Order` / `Ordered`.
3. `@Order` влияет только на порядок вызова `run()`, но не на порядок создания бинов.
4. Если `@Order` не указан — порядок неопределён.

### 8.3. Правило: runners вместо `@PostConstruct`

**Ключевое правило Spring Boot**: задачи, которые должны выполняться при старте приложения (после полной инициализации контекста), следует реализовывать через `ApplicationRunner` или `CommandLineRunner`, а не через `@PostConstruct`.

Почему:
- `@PostConstruct` вызывается на этапе инициализации бина, когда не все бины готовы.
- Runners вызываются после того, как весь контекст refreshed, все бины созданы, веб-сервер запущен.
- Runners получают доступ к `ApplicationArguments`.

### 8.4. Graceful shutdown — «мягкая» остановка

Начиная со Spring Boot 2.3, встроена поддержка graceful shutdown.

Что происходит:
1. Приложение получает сигнал `SIGTERM` (например, от Kubernetes).
2. JVM shutdown hook инициирует `context.close()`.
3. Первая фаза — остановка приёма новых запросов.
4. Существующие запросы получают grace period на завершение.
5. После завершения всех запросов — уничтожение бинов.

Конфигурация:
```yaml
server:
  shutdown: graceful
spring:
  lifecycle:
    timeout-per-shutdown-phase: 30s
```

`server.shutdown` может быть `graceful` или `immediate` (по умолчанию). `timeout-per-shutdown-phase` — максимальное время ожидания завершения запросов (по умолчанию 30 секунд).

Поведение по серверам:

| Сервер | Как отклоняет новые запросы
|--|--
| Tomcat | На сетевом уровне (перестаёт принимать соединения)
| Jetty | На сетевом уровне
| Reactor Netty | На сетевом уровне
| Undertow | Принимает соединение, но отвечает `503 Service Unavailable`

> **Важно:** graceful shutdown работает только при получении корректного SIGTERM. Если остановить приложение из IDE (кнопкой Stop), IDE может послать SIGKILL, и graceful shutdown не сработает.

### 8.5. `WebServerGracefulShutdownLifecycle` — реализация

В части 6 мы упоминали, что при создании `WebServer` регистрируются два lifecycle-бина:
```java
getBeanFactory().registerSingleton("webServerGracefulShutdown", 
    new WebServerGracefulShutdownLifecycle(this.webServer));
getBeanFactory().registerSingleton("webServerStartStop", 
    new WebServerStartStopLifecycle(this, this.webServer));
```

`WebServerGracefulShutdownLifecycle` реализует `SmartLifecycle`:
```java
class WebServerGracefulShutdownLifecycle implements SmartLifecycle {
    private final WebServer webServer;
    private volatile boolean running;
    
    @Override
    public void start() {
        this.running = true;
    }
    
    @Override
    public void stop(Runnable callback) {
        this.running = false;
        this.webServer.shutDownGracefully((result) -> callback.run());
    }
    
    @Override
    public int getPhase() {
        return Integer.MAX_VALUE;  // ← самая высокая фаза
    }
}
```

**Ключевой момент:** `getPhase() = Integer.MAX_VALUE`. При остановке `SmartLifecycle`-бины останавливаются в порядке убывания phase, поэтому graceful shutdown веб-сервера происходит первым — до того, как начнут уничтожаться бины, которые могут обрабатывать запросы (DataSource, сервисы и т.д.).

### 8.6. `SmartLifecycle` и `LifecycleProcessor` — порядок start/stop

`SmartLifecycle` расширяет `Lifecycle` и `Phased`:

```java
public interface SmartLifecycle extends Lifecycle, Phased {
    boolean isAutoStartup();
    void stop(Runnable callback);
    
    @Override
    default int getPhase() {
        return 0;
    }
}
```

Правила порядка:

| Фаза | Старт | Остановка
|--|--|--
| `Integer.MIN_VALUE` | Раньше всех | Позже всех
| `0` (default) | Середина | Середина
| `Integer.MAX_VALUE` | Позже всех | Раньше всех

**При старте** (`onRefresh()`): фазы сортируются по возрастанию — `MIN_VALUE` запускается первым, `MAX_VALUE` — последним.

**При остановке** (`onClose()`): фазы сортируются по убыванию — `MAX_VALUE` останавливается первым, `MIN_VALUE` — последним.

`DefaultLifecycleProcessor` управляет этим:
```java
public class DefaultLifecycleProcessor implements LifecycleProcessor, BeanFactoryAware {
    private long timeoutPerShutdownPhase = 30000;  // 30 секунд
    
    @Override
    public void onRefresh() {
        startBeans(true);  // все SmartLifecycle с isAutoStartup() == true
    }
    
    @Override
    public void onClose() {
        stopBeans();       // все запущенные Lifecycle
    }
}
```

`timeoutPerShutdownPhase` — максимальное время ожидания для одной фазы. Если бин не завершил `stop()` за это время — Spring продолжает, не дожидаясь его.

### 8.7. JVM shutdown hook — как JVM узнаёт, что нужно остановиться

`SpringApplication.refreshContext()` автоматически регистрирует shutdown hook:
```java
private void refreshContext(ConfigurableApplicationContext context) {
    if (this.registerShutdownHook) {
        try {
            context.registerShutdownHook();
        } catch (AccessControlException ex) {
            // Not allowed in some environments
        }
    }
    refresh(context);
}
```

`registerShutdownHook` по умолчанию `true`. Отключается через `SpringApplication.setRegisterShutdownHook(false)`.

Внутри `AbstractApplicationContext`:
```java
public void registerShutdownHook() {
    if (this.shutdownHook == null) {
        this.shutdownHook = new Thread(SHUTDOWN_HOOK_THREAD_NAME) {
            @Override
            public void run() {
                synchronized (startupShutdownMonitor) {
                    doClose();
                }
            }
        };
        Runtime.getRuntime().addShutdownHook(this.shutdownHook);
    }
}
```

Цепочка остановки:
```
SIGTERM от ОС / Ctrl+C / kill <pid>
    │
    ▼
JVM Shutdown Hook (Thread "SpringContextShutdownHook")
    │
    ▼
AbstractApplicationContext.doClose()
    │
    ├─ 1. LifecycleProcessor.onClose()
    │       └─ SmartLifecycle.stop() — по убыванию phase
    │             ├─ WebServerGracefulShutdownLifecycle (MAX_VALUE) — graceful shutdown
    │             ├─ WebServerStartStopLifecycle (MAX-1) — tomcat.stop()
    │             └─ ... пользовательские SmartLifecycle
    │
    ├─ 2. destroyBeans()
    │       └─ для каждого singleton:
    │             ├─ @PreDestroy (DestructionAwareBeanPostProcessor)
    │             ├─ DisposableBean.destroy()
    │             └─ @Bean(destroyMethod) / AutoCloseable.close()
    │
    ├─ 3. closeBeanFactory()
    │
    └─ 4. active.set(false)
```

### 8.8. Ручное закрытие контекста

Если нужно закрыть контекст программно (например, в тестах или в CLI-приложении):
```java
ConfigurableApplicationContext context = SpringApplication.run(MyApp.class, args);
// ... работа
context.close();  // → doClose()
```

Или через `SpringApplication.exit()`:
```java
int exitCode = SpringApplication.exit(context, () -> 0);
System.exit(exitCode);
```

### 8.9. Spring Boot 3 vs Spring Boot 4 — отличия

| Аспект | Spring Boot 3.x | Spring Boot 4.x
|--|--|--
| Graceful shutdown | Встроен, `server.shutdown=graceful` | Встроен, включён по умолчанию для всех встроенных серверов
| Undertow | Поддерживается | Удалён (несовместим с Servlet 6.1)
| Jetty graceful shutdown | `StatisticsHandler` | `GracefulHandler` (изменено в 4.2)
| `WebServerGracefulShutdownLifecycle` | `org.springframework.boot.web.server` | Модульная структура, пакеты переехали
| `ApplicationRunner` / `CommandLineRunner` | Без изменений | Без изменений
| `SmartLifecycle` / `LifecycleProcessor` | Без изменений | Без изменений
| JVM shutdown hook | По умолчанию `true` | По умолчанию `true`
| Исправление deadlock | — | Spring Framework 7.0.4: исправлен deadlock при конкурентных ShutdownHook (issue #36260)

**Ключевое изменение в Boot 4:** graceful shutdown включён по умолчанию для всех встроенных серверов. Это значит, что при получении `SIGTERM` приложение автоматически перестаёт принимать новые запросы и ждёт завершения существующих — без явной конфигурации.

### 8.10. Что запомнить

1. `ApplicationRunner` vs `CommandLineRunner` — разница только в формате аргументов (`ApplicationArguments` vs `String[]`). `ApplicationRunner` предпочтительнее.
2. Порядок runners — через `@Order` / `Ordered`. Runners вызываются после полной инициализации контекста, но до `ApplicationReadyEvent`.
3. `@PostConstruct` ≠ runner. Задачи «при старте приложения» — через runners.
4. **Graceful shutdown** — `server.shutdown=graceful` + `spring.lifecycle.timeout-per-shutdown-phase`. В Boot 4 включён по умолчанию.
5. `SmartLifecycle` — порядок: старт по возрастанию `phase`, остановка по убыванию. `WebServerGracefulShutdownLifecycle` имеет `phase = MAX_VALUE`, поэтому останавливается первым.
6. **JVM shutdown hook** регистрируется автоматически (`registerShutdownHook = true`). Через него `context.close()` вызывается при `SIGTERM` / `Ctrl+C`.
7. **Цепочка остановки:** shutdown hook → `doClose()` → `LifecycleProcessor.onClose()` → `destroyBeans()` → `@PreDestroy` → `DisposableBean.destroy()`.
8. **Boot 4:** graceful shutdown по умолчанию, Undertow удалён, исправлен deadlock в shutdown hook.
