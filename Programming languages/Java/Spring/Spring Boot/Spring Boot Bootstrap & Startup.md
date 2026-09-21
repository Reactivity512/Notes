# Spring Boot

## Часть I. Bootstrap и запуск

### 1.1. Точка входа main() — анатомия старта

Каждое Spring Boot приложение начинается одинаково:
```java
@SpringBootApplication
public class MyApplication {
    public static void main(String[] args) {
        SpringApplication.run(MyApplication.class, args);
    }
}
```
Эта одна строка запускает весь механизм.

### 1.2. Статический `run()` vs Экземпляр `SpringApplication`

Есть два способа запуска:

**Способ 1 — статический метод (95% случаев):**
```java
SpringApplication.run(MyApplication.class, args);
```
Внутри это делает:
```java
public static ConfigurableApplicationContext run(Class<?> primarySource, String... args) {
    return new SpringApplication(primarySource).run(args);
}
```

То есть сначала создаётся экземпляр `SpringApplication`, затем вызывается его instance-метод `run()`. Это важно понимать: **вся настройка происходит на этапе конструктора**, а ``run()`` — уже исполнение.

**Способ 2 — ручное создание экземпляра (когда нужна кастомизация до запуска):**

```java
public static void main(String[] args) {
    SpringApplication app = new SpringApplication(MyApplication.class);
    app.setBannerMode(Banner.Mode.OFF);
    app.setAdditionalProfiles("dev");
    app.setWebApplicationType(WebApplicationType.SERVLET);
    app.run(args);
}
```
Именно поэтому существует отдельный класс `SpringApplicationBuilder` — для fluent-настройки:

```java
new SpringApplicationBuilder(MyApplication.class)
    .profiles("dev")
    .bannerMode(Banner.Mode.OFF)
    .web(WebApplicationType.SERVLET)
    .run(args);
```

`SpringApplicationBuilder` особенно полезен, когда нужно построить иерархию контекстов (parent-child):
```java
new SpringApplicationBuilder(ParentConfig.class)
    .child(ChildConfig.class)
    .run(args);
```

Такой подход используется в Spring Cloud (bootstrap context) и в тестовых сценариях

### 1.3. Конструктор `SpringApplication` — что происходит до `run()`

Конструктор делает три ключевые вещи:

**1. Определение типа приложения (WebApplicationType)**
```java
private WebApplicationType deduceWebApplicationType() {
    if (ClassUtils.isPresent(REACTIVE_WEB_ENVIRONMENT_CLASS, null)
            && !ClassUtils.isPresent(MVC_WEB_ENVIRONMENT_CLASS, null)) {
        return WebApplicationType.REACTIVE;
    }
    // ... проверки для SERVLET и NONE
}
```
Spring Boot сканирует classpath:
* Есть `DispatcherServlet` (Spring MVC) → `SERVLET`
* Есть `DispatcherHandler` (WebFlux) без MVC → `REACTIVE`
* Нет веб-классов → `NONE`

**2. Загрузка `ApplicationContextInitializer` и `ApplicationListener`**

Через `SpringFactoriesLoader` читается `META-INF/spring.factories:`
```declarative
org.springframework.context.ApplicationContextInitializer=\
org.springframework.boot.context.ConfigurationWarningsApplicationContextInitializer,\
org.springframework.boot.context.ContextIdApplicationContextInitializer,\
...
```

|Характеристика|`ApplicationContextInitializer`|`ApplicationListener`
|--|--|--
|Паттерн|Стратегия инициализации (Template method)|Наблюдатель (Observer)
|Главная цель|Настроить / Изменить контекст и окружение.|Отреагировать на то, что происходит с приложением.
|Тайминг|Строго один раз: до refresh() (до создания бинов).|Множество раз: на старте, в рантайме, при шатдауне.
|Доступ к бинам|Нет (бины еще не созданы).|Да (если слушаешь поздние события или работаешь внутри бина).
|Типичный пример|Подтянуть секреты из Vault в Environment.|Отправить алерт в Telegram при ApplicationReadyEvent.

**3. Установка primary sources**

`MyApplication.class` сохраняется как `primarySources` — это корневой `@Configuration`-класс, от которого пойдёт сканирование.

### 1.4. Instance-метод run() — пошаговый разбор

Вот скелет метода run(String... args) (Spring Boot 3.x):
```java
public ConfigurableApplicationContext run(String... args) {
    // 1. Создание BootstrapContext
    DefaultBootstrapContext bootstrapContext = createBootstrapContext();
    ConfigurableApplicationContext context = null;
    
    // 2. Headless-режим (java.awt.headless)
    configureHeadlessProperty();
    
    // 3. Получение RunListeners
    SpringApplicationRunListeners listeners = getRunListeners(args);
    listeners.starting(bootstrapContext, this.mainApplicationClass);
    
    try {
        // 4. Подготовка Environment
        ApplicationArguments applicationArguments = new DefaultApplicationArguments(args);
        ConfigurableEnvironment environment = prepareEnvironment(listeners, 
            bootstrapContext, applicationArguments);
        
        // 5. Печать баннера
        Banner printedBanner = printBanner(environment);
        
        // 6. Создание ApplicationContext
        context = createApplicationContext();
        context.setApplicationStartup(this.applicationStartup);
        
        // 7. Подготовка контекста
        prepareContext(bootstrapContext, context, environment, listeners, 
            applicationArguments, printedBanner);
        
        // 8. Refresh — ключевой этап
        refreshContext(context);
        
        // 9. AfterRefresh (hook)
        afterRefresh(context, applicationArguments);
        
        // 10. Публикация ApplicationStartedEvent
        listeners.started(context, timeTakenToStartup);
        
        // 11. Вызов ApplicationRunner / CommandLineRunner
        callRunners(context, applicationArguments);
    } catch (Throwable ex) {
        handleRunFailure(context, ex, listeners);
        throw new IllegalStateException(ex);
    }
    
    // 12. Публикация ApplicationReadyEvent
    listeners.ready(context, timeTakenToReady);
    return context;
}
```

**Шаг 1: `DefaultBootstrapContext`**

`BootstrapContext` — это контейнер для объектов, доступных до создания `ApplicationContext`. Он живёт от начала `run()` до момента, когда контекст готов. Используется для регистрации early beans, например, `EnvironmentPostProcessor`.

**Шаг 2: configureHeadlessProperty()**
```java
private void configureHeadlessProperty() {
    System.setProperty(SYSTEM_PROPERTY_JAVA_AWT_HEADLESS,
        System.getProperty(SYSTEM_PROPERTY_JAVA_AWT_HEADLESS, 
            Boolean.toString(this.headless)));
}
```

Устанавливает `java.awt.headless=true` по умолчанию. Это важно для серверных приложений — отключает GUI-компоненты AWT.

**Шаг 3: `SpringApplicationRunListeners` — сердце событий**

Это SPI-интерфейс для слушателей жизненного цикла запуска:
```java
public interface SpringApplicationRunListener {
    void starting(ConfigurableBootstrapContext bootstrapContext);
    void environmentPrepared(ConfigurableBootstrapContext bootstrapContext, 
        ConfigurableEnvironment environment);
    void contextPrepared(ConfigurableApplicationContext context);
    void contextLoaded(ConfigurableApplicationContext context);
    void started(ConfigurableApplicationContext context, Duration timeTaken);
    void ready(ConfigurableApplicationContext context, Duration timeTaken);
    void failed(ConfigurableApplicationContext context, Throwable exception);
}
```

Загружается через `SpringFactoriesLoader` из `META-INF/spring.factories`. Spring Boot поставляет одну реализацию — `EventPublishingRunListener`.

`EventPublishingRunListener` — это мост между `SpringApplicationRunListener` и `ApplicationEvent`. Он транслирует каждый lifecycle-колбэк в соответствующее событие.

|Callback|Event
|--|--
|`starting()`|`ApplicationStartingEvent`
|`environmentPrepared()`|`ApplicationEnvironmentPreparedEvent`
|`contextPrepared()`|`ApplicationContextInitializedEvent`
|`contextLoaded()`|`ApplicationPreparedEvent`
|`started()`|`ApplicationStartedEvent`
|`ready()`|`ApplicationReadyEvent`
|`failed()`|`ApplicationFailedEvent`

Полная цепочка событий:

|Event| Когда                |Context доступен?
|--|----------------------|--
|`ApplicationStartingEvent`| До создания Environment и Context|Нет
|`ApplicationEnvironmentPreparedEvent`| Environment готов, Context ещё нет |Нет
|`ApplicationContextInitializedEvent`| Context создан, бины ещё не загружены |Да (пустой)
|`ApplicationPreparedEvent`| Бины загружены, refresh ещё не вызван |Да (не refreshed)
|`ApplicationStartedEvent`| Refresh выполнен, runners ещё не вызваны |Да (refreshed)
|`AvailabilityChangeEvent`(Liveness)| Сразу после `ApplicationStartedEvent` |Да
|`ApplicationReadyEvent`| Все runners выполнены|Да (refreshed)
|`AvailabilityChangeEvent` (Readiness)| Сразу после `ApplicationReadyEvent` |Да
|`ApplicationFailedEvent`| При любом необработанном исключении|Частично

**Шаг 4: `prepareEnvironment()` — подготовка окружения**

```java
private ConfigurableEnvironment prepareEnvironment(
        SpringApplicationRunListeners listeners,
        ConfigurableBootstrapContext bootstrapContext,
        ApplicationArguments applicationArguments) {
    
    // 1. Создание или переиспользование Environment
    ConfigurableEnvironment environment = getOrCreateEnvironment();
    
    // 2. Конфигурация: command line args, профили
    configureEnvironment(environment, applicationArguments.getSourceArgs());
    
    // 3. Привязка ConfigurationPropertySources
    ConfigurationPropertySources.attach(environment);
    
    // 4. Публикация ApplicationEnvironmentPreparedEvent
    listeners.environmentPrepared(bootstrapContext, environment);
    
    // 5. Привязка spring.main.*
    DefaultPropertiesPropertySource.moveToEnd(environment);
    bindToSpringApplication(environment);
    
    // 6. Конвертация Environment при необходимости
    if (!this.isCustomEnvironment) {
        environment = convertEnvironment(environment);
    }
    
    ConfigurationPropertySources.attach(environment);
    return environment;
}
```

Ключевые моменты:
* `getOrCreateEnvironment()` — создаёт `StandardServletEnvironment` / `StandardReactiveEnvironment` / `StandardEnvironment`
* `configureEnvironment()` — добавляет `CommandLinePropertySource`, устанавливает active profiles
* `listeners.environmentPrepared()` — здесь срабатывают `EnvironmentPostProcessor` (например, загрузка `application.yml`)

**Шаг 5: printBanner() — печать баннера**

```java
private Banner printBanner(ConfigurableEnvironment environment) {
    if (this.bannerMode == Banner.Mode.OFF) {
        return null;
    }
    ResourceLoader resourceLoader = (this.resourceLoader != null) 
        ? this.resourceLoader : new DefaultResourceLoader(null);
    SpringApplicationBannerPrinter bannerPrinter = new SpringApplicationBannerPrinter(
        resourceLoader, this.banner);
    
    if (this.bannerMode == Banner.Mode.LOG) {
        return bannerPrinter.print(environment, this.mainApplicationClass, logger);
    }
    return bannerPrinter.print(environment, this.mainApplicationClass, System.out);
}
```

* Файл `banner.txt` в `src/main/resources/` — подхватывается автоматически
* `spring.banner.location` — указать путь к файлу
* `spring.main.banner-mode=off` — отключить
* Spring Boot 3.0.0 M2+: поддержка PNG/JPEG/GIF удалена, только `banner.txt`

Переменные в баннере: `${spring-boot.version}`, `${application.version}`, `${application.formatted-version}`

**Шаг 6: `createApplicationContext()` — выбор типа контекста**

```java
protected ConfigurableApplicationContext createApplicationContext() {
    return this.applicationContextFactory.create(this.webApplicationType);
}
```

В Spring Boot 3 используется ApplicationContextFactory (enum-подобная фабрика):
```java
ApplicationContextFactory DEFAULT = (webApplicationType) -> {
    try {
        switch (webApplicationType) {
            case SERVLET:
                return new AnnotationConfigServletWebServerApplicationContext();
            case REACTIVE:
                return new AnnotationConfigReactiveWebServerApplicationContext();
            default:
                return new AnnotationConfigApplicationContext();
        }
    } catch (Exception ex) {
        throw new IllegalStateException("Unable create a default ApplicationContext instance, "
            + "you may need a custom ApplicationContextFactory", ex);
    }
};
```

**В Spring Boot 4**: `ApplicationContextFactory` **удалён**. Вместо него используется модульная система — каждый модуль (spring-boot-webmvc, spring-boot-webflux) регистрирует свой контекст через модульные зависимости. `AbstractApplicationContextFactory` помечен как `@Deprecated(since="6.0", forRemoval=true)`

**Шаг 7: `prepareContext()` — подготовка контекста**

```java
private void prepareContext(DefaultBootstrapContext bootstrapContext,
        ConfigurableApplicationContext context,
        ConfigurableEnvironment environment,
        SpringApplicationRunListeners listeners,
        ApplicationArguments applicationArguments,
        Banner printedBanner) {
    
    // 1. Установка Environment
    context.setEnvironment(environment);
    
    // 2. PostProcessApplicationContext (bean name generator, resource loader)
    postProcessApplicationContext(context);
    
    // 3. ApplicationContextInitializer
    applyInitializers(context);
    
    // 4. Публикация ApplicationContextInitializedEvent
    listeners.contextPrepared(context);
    
    // 5. Регистрация beans: springApplicationArguments, springBootBanner
    if (this.logStartupInfo) {
        logStartupInfo(context.getParent() == null);
        logStartupProfileInfo(context);
    }
    ConfigurableListableBeanFactory beanFactory = context.getBeanFactory();
    beanFactory.registerSingleton("springApplicationArguments", applicationArguments);
    if (printedBanner != null) {
        beanFactory.registerSingleton("springBootBanner", printedBanner);
    }
    if (beanFactory instanceof DefaultListableBeanFactory dlbf) {
        dlbf.setAllowBeanDefinitionOverriding(this.allowBeanDefinitionOverriding);
    }
    if (this.lazyInitialization) {
        context.addBeanFactoryPostProcessor(
            new LazyInitializationBeanFactoryPostProcessor());
    }
    
    // 6. Загрузка всех sources
    Set<Object> sources = getAllSources();
    load(context, sources.toArray(new Object[0]));
    
    // 7. Публикация ApplicationPreparedEvent
    listeners.contextLoaded(context);
}
```

**Ключевой момент** — `applyInitializers()`: здесь вызываются все `ApplicationContextInitializer`, загруженные из `spring.factories`. Это механизм для кастомизации контекста до загрузки бинов.

**Шаг 8: `refreshContext()` — запуск Spring-контейнера**

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

Метод `refresh()` — это `AbstractApplicationContext.refresh()` из Spring Framework. Именно здесь:
* Создаются все singleton-бины
* Запускается `BeanFactoryPostProcessor` → `ConfigurationClassPostProcessor` обрабатывает `@Configuration`
* Запускается `AutoConfigurationImportSelector` → автоконфигурация
* Запускаются `BeanPostProcessor`
* Стартует встроенный веб-сервер

**Шаг 9: `afterRefresh()` — hook**

```java
protected void afterRefresh(ConfigurableApplicationContext context, 
        ApplicationArguments args) {
}
```
Пустой hook для переопределения в подклассах.

**Шаг 10: `listeners.started()` — `ApplicationStartedEvent`**

Публикуется `ApplicationStartedEvent` и `AvailabilityChangeEvent(LivenessState.CORRECT)`. С этого момента приложение считается запущенным, но ещё не готовым принимать трафик.

**Шаг 11: `callRunners()` — `ApplicationRunner` и `CommandLineRunner`**

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

* `ApplicationRunner` — получает `ApplicationArguments` (parsed)
* `CommandLineRunner` — получает `String[]` (raw)
* Сортируются через `@Order` / `Ordered`

**Шаг 12: `listeners.ready()` — `ApplicationReadyEvent`**
Публикуется `ApplicationReadyEvent` и `AvailabilityChangeEvent(ReadinessState.ACCEPTING_TRAFFIC)`. **С этого момента приложение готово принимать трафик.**

### 1.5. Обработка ошибок — `handleRunFailure()`

```java
private void handleRunFailure(ConfigurableApplicationContext context,
        Throwable exception, SpringApplicationRunListeners listeners) {
    try {
        try {
            handleExitCode(context, exception);
            if (listeners != null) {
                listeners.failed(context, exception);
            }
        } finally {
            reportFailure(getExceptionReporters(context), exception);
            if (context != null) {
                context.close();
            }
        }
    } catch (Exception ex) {
        logger.warn("Unable to close ApplicationContext", ex);
    }
    ReflectionUtils.rethrowRuntimeException(exception);
}
```

* `handleExitCode()` — определяет код выхода через `ExitCodeExceptionMapper`
* `listeners.failed()` — публикует `ApplicationFailedEvent`
* `FailureAnalyzer` — пытается дать human-readable описание ошибки
* Контекст закрывается

### 1.6. Диаграмма последовательности запуска

```
main()
  │
  ├─ new SpringApplication(primarySource)
  │     ├─ deduceWebApplicationType()          → SERVLET / REACTIVE / NONE
  │     ├─ load SpringFactories (Initializers, Listeners)
  │     └─ set primarySources
  │
  └─ .run(args)
        │
        ├─ createBootstrapContext()
        ├─ configureHeadlessProperty()
        ├─ getRunListeners()                    → EventPublishingRunListener
        ├─ listeners.starting()                 → ApplicationStartingEvent
        │
        ├─ prepareEnvironment()
        │     ├─ getOrCreateEnvironment()       → StandardServletEnvironment
        │     ├─ configureEnvironment()         → CommandLinePropertySource, profiles
        │     ├─ ConfigurationPropertySources.attach()
        │     └─ listeners.environmentPrepared()→ ApplicationEnvironmentPreparedEvent
        │           └─ EnvironmentPostProcessor → ConfigDataEnvironmentPostProcessor
        │                 └─ загрузка application.yml/properties
        │
        ├─ printBanner()                        → banner.txt → System.out
        │
        ├─ createApplicationContext()
        │     └─ ApplicationContextFactory      → AnnotationConfigServletWebServerApplicationContext
        │
        ├─ prepareContext()
        │     ├─ setEnvironment()
        │     ├─ postProcessApplicationContext()
        │     ├─ applyInitializers()            → ApplicationContextInitializer
        │     ├─ listeners.contextPrepared()    → ApplicationContextInitializedEvent
        │     ├─ registerSingleton(springApplicationArguments)
        │     ├─ registerSingleton(springBootBanner)
        │     ├─ load(sources)                  → BeanDefinitionLoader
        │     └─ listeners.contextLoaded()      → ApplicationPreparedEvent
        │
        ├─ refreshContext()
        │     ├─ registerShutdownHook()
        │     └─ AbstractApplicationContext.refresh()
        │           ├─ invokeBeanFactoryPostProcessors()
        │           │     └─ ConfigurationClassPostProcessor → @Configuration, @Bean
        │           │     └─ AutoConfigurationImportSelector → автоконфигурация
        │           ├─ registerBeanPostProcessors()
        │           ├─ initMessageSource()
        │           ├─ initApplicationEventMulticaster()
        │           ├─ onRefresh()
        │           │     └─ ServletWebServerApplicationContext.createWebServer()
        │           │           └─ TomcatServletWebServerFactory → Tomcat.start()
        │           └─ finishRefresh()
        │
        ├─ afterRefresh()                       → hook
        ├─ listeners.started()                  → ApplicationStartedEvent
        │                                          AvailabilityChangeEvent(Liveness)
        │
        ├─ callRunners()
        │     ├─ ApplicationRunner
        │     └─ CommandLineRunner
        │
        └─ listeners.ready()                    → ApplicationReadyEvent
                                                   AvailabilityChangeEvent(Readiness)
```

### 1.7. Spring Boot 3 vs Spring Boot 4 — ключевые отличия в bootstrap

| Аспект |Spring Boot 3.x| Spring Boot 4.x
|--|--|--
| `ApplicationContextFactory` | Enum-подобная фабрика `DEFAULT`, выбирает контекст по `WebApplicationType`| **Удалён.** Контекст определяется модульными зависимостями (`spring-boot-webmvc` / `spring-boot-webflux`)
| `spring-boot-autoconfigure` | Один монолитный JAR (6.2MB) | 47 модульных JAR'ов, каждый для своей технологии
| Starter'ы | `spring-boot-starter-web` | `spring-boot-starter-webmvc` (переименован)
| `spring.factories` | SPI для auto-configuration | `META-INF/spring/...AutoConfiguration.imports` (уже в 3.x, в 4.x — обязательный)
| Undertow | Поддерживается | **Удалён** (несовместим с Servlet 6.1)
| Java baseline | Java 17+ | Java 21+
| Spring Framework | 6.x | 7.0
| Jackson | Jackson 2.x | Jackson 3.x (namespace changes)
| Banner (image) | PNG/JPEG/GIF удалены с 3.0 M2 | Только `banner.txt`

### Вывод

* `SpringApplication.run()` — это фасад. Вся реальная работа — в instance-методе `run()`.
* Конструктор `SpringApplication` определяет `WebApplicationType` по classpath и загружает SPI.
* `SpringApplicationRunListener` — точка расширения для вмешательства в lifecycle. `EventPublishingRunListener` транслирует вызовы в `ApplicationEvent`.
* `Environment` готовится до `ApplicationContext` — поэтому `EnvironmentPostProcessor` может добавлять property sources до создания бинов.
* `refreshContext()` — самый важный шаг. Здесь создаются все бины, обрабатываются `@Configuration`, запускается Tomcat/Netty.
* `ApplicationStartedEvent` ≠ `ApplicationReadyEvent`. Первое — контекст поднят, второе — runners выполнены, приложение готово принимать трафик.
* Spring Boot 4 — модульный. `ApplicationContextFactory` удалён, контекст определяется зависимостями.

