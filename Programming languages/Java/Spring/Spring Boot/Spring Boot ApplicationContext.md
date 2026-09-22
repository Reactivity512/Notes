# Spring Boot

## Часть 3. Создание ApplicationContext

### 3.1. Что такое `ApplicationContext` и почему их несколько

`ApplicationContext` — это расширенный `BeanFactory`, который добавляет:
* Поддержку ApplicationEvent (publish/subscribe)
* MessageSource (i18n)
* ResourceLoader (доступ к файлам)
* Автоматическую регистрацию BeanPostProcessor / BeanFactoryPostProcessor
* Иерархию контекстов (parent/child)

В Spring Boot существует три базовых типа контекста, по одному на каждый WebApplicationType:

| WebApplicationType | Класс контекста | Модуль (Boot 4)
|--|--|--
| SERVLET | AnnotationConfigServletWebServerApplicationContext | spring-boot-webmvc
| REACTIVE | AnnotationConfigReactiveWebServerApplicationContext | spring-boot-webflux
| NONE | AnnotationConfigApplicationContext | spring-boot (core)

Ключевое отличие `AnnotationConfigServletWebServerApplicationContext` от обычного `AnnotationConfigApplicationContext` — он умеет создавать и управлять встроенным веб-сервером (Tomcat/Jetty). Он наследуется от `ServletWebServerApplicationContext`, который реализует `WebServerApplicationContext` и содержит методы `createWebServer()`, `getWebServer()` и т.д.

### 3.2. `createApplicationContext()` — выбор контекста

В `SpringApplication.run()` этот шаг выглядит так:

```java
protected ConfigurableApplicationContext createApplicationContext() {
    return this.applicationContextFactory.create(this.webApplicationType);
}
```

`ApplicationContextFactory` — это функциональный интерфейс (начиная с Spring Boot 2.4), который отвечает за создание контекста на основе типа приложения:

```java
@FunctionalInterface
public interface ApplicationContextFactory {
    ConfigurableApplicationContext create(WebApplicationType webApplicationType);

    // ... фабричные методы of(...)
}
```

Дефолтная реализация (`ApplicationContextFactory.DEFAULT`) внутри выглядит примерно так:
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
        throw new IllegalStateException(
            "Unable create a default ApplicationContext instance, "
            + "you may need a custom ApplicationContextFactory", ex);
    }
};
```

Как кастомизировать
```java
// Способ 1: через SpringApplication
SpringApplication app = new SpringApplication(MyApp.class);
app.setApplicationContextFactory(ctx -> new MyCustomContext());

// Способ 2: через SpringApplicationBuilder
new SpringApplicationBuilder(MyApp.class)
    .contextFactory(ApplicationContextFactory.of(MyCustomContext::new))
    .run(args);
```

`ApplicationContextFactory.of(Supplier<ConfigurableApplicationContext>)` создаёт фабрику, которая просто вызывает переданный `Supplier`

> Подводный камень. Возвращаемый контекст должен быть в «сыром» виде — Spring Boot сам вызовет `setEnvironment()`, `prepareContext()` и `refresh()`. Если ты вернёшь уже refreshed-контекст, получишь двойную инициализацию.

### 3.3. `prepareContext()` — подготовка контекста до refresh

Это самый насыщенный шаг. Вот его полный скелет (Spring Boot 3.x):
```java
private void prepareContext(DefaultBootstrapContext bootstrapContext,
        ConfigurableApplicationContext context,
        ConfigurableEnvironment environment,
        SpringApplicationRunListeners listeners,
        ApplicationArguments applicationArguments,
        Banner printedBanner) {

    // 1. Установка Environment
    context.setEnvironment(environment);

    // 2. Пост-обработка контекста (bean name generator, resource loader, ConversionService)
    postProcessApplicationContext(context);

    // 3. Применение ApplicationContextInitializer
    applyInitializers(context);

    // 4. Публикация ApplicationContextInitializedEvent
    listeners.contextPrepared(context);

    // 5. Логирование startup info и профилей
    if (this.logStartupInfo) {
        logStartupInfo(context.getParent() == null);
        logStartupProfileInfo(context);
    }

    // 6. Регистрация boot-специфичных singleton-бинов
    ConfigurableListableBeanFactory beanFactory = context.getBeanFactory();
    beanFactory.registerSingleton("springApplicationArguments", applicationArguments);
    if (printedBanner != null) {
        beanFactory.registerSingleton("springBootBanner", printedBanner);
    }

    // 7. Настройка allowBeanDefinitionOverriding
    if (beanFactory instanceof DefaultListableBeanFactory dlbf) {
        dlbf.setAllowBeanDefinitionOverriding(this.allowBeanDefinitionOverriding);
    }

    // 8. Lazy initialization
    if (this.lazyInitialization) {
        context.addBeanFactoryPostProcessor(
            new LazyInitializationBeanFactoryPostProcessor());
    }

    // 9. Загрузка всех sources (главный @Configuration-класс и др.)
    Set<Object> sources = getAllSources();
    load(context, sources.toArray(new Object[0]));

    // 10. Публикация ApplicationPreparedEvent
    listeners.contextLoaded(context);
}
```

Разберём ключевые шаги.

**Шаг 1: `context.setEnvironment(environment)`**

Environment, подготовленный в Части II, передаётся в контекст. С этого момента `EnvironmentAware`-бины и `@Value` смогут читать свойства.

**Шаг 2: `postProcessApplicationContext()`**

```java
protected void postProcessApplicationContext(ConfigurableApplicationContext context) {
    if (this.beanNameGenerator != null) {
        context.getBeanFactory().registerSingleton(
            AnnotationConfigUtils.CONFIGURATION_BEAN_NAME_GENERATOR,
            this.beanNameGenerator);
    }
    if (this.resourceLoader != null) {
        if (context instanceof GenericApplicationContext gac) {
            gac.setResourceLoader(this.resourceLoader);
        }
        context.getBeanFactory().registerResolvableDependency(
            ResourceLoader.class, this.resourceLoader);
    }
    if (this.addConversionService) {
        context.getBeanFactory().setConversionService(
            ApplicationConversionService.getSharedInstance());
    }
}
```

Три вещи:
* BeanNameGenerator — стратегия именования бинов (по умолчанию `AnnotationBeanNameGenerator`).
* ResourceLoader — откуда читать ресурсы (по умолчанию `DefaultResourceLoader`).
* ConversionService — `ApplicationConversionService` для конвертации типов (используется в `@ConfigurationProperties` и `@Value`).

**Шаг 3: `applyInitializers()` — точка расширения №1**

```java
protected void applyInitializers(ConfigurableApplicationContext context) {
    for (ApplicationContextInitializer initializer : getInitializers()) {
        Class<?> requiredType = GenericTypeResolver.resolveTypeArgument(
            initializer.getClass(), ApplicationContextInitializer.class);
        Assert.isInstanceOf(requiredType, context, 
            "Unable to call initializer.");
        initializer.initialize(context);
    }
}
```

`ApplicationContextInitializer` — это SPI-интерфейс Spring Framework:

```java
@FunctionalInterface
public interface ApplicationContextInitializer<C extends ConfigurableApplicationContext> {
    void initialize(C applicationContext);
}
```

Он вызывается до `refresh()`, когда контекст ещё пустой. Типичные сценарии:
* Программная активация профилей: `context.getEnvironment().addActiveProfile("dev")`
* Регистрация `PropertySource` из нестандартного источника
* Установка `parent`-контекста

Регистрация через `spring.factories`:

```java
# META-INF/spring.factories
org.springframework.context.ApplicationContextInitializer=\
com.example.MyInitializer
```

Программная регистрация:
```java
SpringApplication app = new SpringApplication(MyApp.class);
app.addInitializers(ctx -> ctx.getEnvironment().addActiveProfile("metrics"));
```

> Важно: встроенные `ApplicationContextInitializer` Spring Boot (например, `ConfigurationWarningsApplicationContextInitializer`, `ContextIdApplicationContextInitializer`) загружаются на этапе конструктора `SpringApplication`, а не здесь. `applyInitializers()` вызывает и их, и добавленные пользователем.

**Шаг 4: `listeners.contextPrepared()`**

Публикуется `ApplicationContextInitializedEvent`. На этот момент контекст создан, Environment установлен, инициализаторы применены — но бины ещё не загружены. Это последняя точка, где можно вмешаться в контекст без риска сломать граф зависимостей.

**Шаг 5: логирование профилей**

```java
protected void logStartupProfileInfo(ConfigurableApplicationContext context) {
    Log log = getApplicationLog();
    if (log.isInfoEnabled()) {
        String[] activeProfiles = context.getEnvironment().getActiveProfiles();
        // ... печатает: "The following 1 profile is active: \"dev\""
    }
}
```

Именно эту строчку в логах ты видишь при старте:
```java
The following 1 profile is active: "dev"
```

**Шаг 6: регистрация singleton-бинов**

```java
beanFactory.registerSingleton("springApplicationArguments", applicationArguments);
beanFactory.registerSingleton("springBootBanner", printedBanner);
```

Эти объекты не являются `@Component` — они регистрируются вручную. Внутри `registerSingleton()` вызывает `addSingleton()` в `DefaultSingletonBeanRegistry`, то есть кладёт объект в singleton cache напрямую, минуя обычный жизненный цикл бина.

Это значит:
* Их нельзя проксировать через `BeanPostProcessor`
* Они не проходят `@PostConstruct` / `InitializingBean`
* Но их можно инжектить через `@Autowired ApplicationArguments`

**Шаг 7: `allowBeanDefinitionOverriding`**

```java
if (beanFactory instanceof DefaultListableBeanFactory dlbf) {
    dlbf.setAllowBeanDefinitionOverriding(this.allowBeanDefinitionOverriding);
}
```

По умолчанию в Boot 2.1+ — `false`. Если два `@Bean`-метода определяют бин с одним именем → `BeanDefinitionOverrideException`. Можно разрешить через `spring.main.allow-bean-definition-overriding=true`.

**Шаг 8: `getAllSources()` + `load()` — вход в мир BeanDefinition**

```java
protected Set<Object> getAllSources() {
    Set<Object> allSources = new LinkedHashSet<>();
    if (!CollectionUtils.isEmpty(this.primarySources)) {
        allSources.addAll(this.primarySources);
    }
    if (!CollectionUtils.isEmpty(this.sources)) {
        allSources.addAll(this.sources);
    }
    return Collections.unmodifiableSet(allSources);
}
```

`primarySources` — это твой `@SpringBootApplication`-класс. Дополнительные sources можно добавить через `SpringApplicationBuilder.sources(...)`.

Затем — `load()`:

```java
protected void load(ApplicationContext context, Object[] sources) {
    BeanDefinitionLoader loader = createBeanDefinitionLoader(
        getBeanDefinitionRegistry(context), sources);
    // ...
    loader.load();
}
```

**Шаг 9: `listeners.contextLoaded()`**

Публикуется `ApplicationPreparedEvent`. Контекст готов к `refresh()`.

### 3.4. `BeanDefinitionLoader` — «фасад» для загрузки бинов

Это ключевой класс, который превращает твои `@Configuration`-классы в `BeanDefinition`-ы. Он не из Spring Framework, а из Spring Boot.

Что он умеет
`BeanDefinitionLoader` — это фасад над тремя читателями:

| Источник | Читатель
|--|--
| `Class<?>` (твой `@Configuration`) | `AnnotatedBeanDefinitionReader`
| `Resource` (XML-файл) | `XmlBeanDefinitionReader`
| `Package` (базовый пакет) | `ClassPathBeanDefinitionScanner`
| `CharSequence` (строка) | Пробует Class → Resource → Package

Как устроен
```java
class BeanDefinitionLoader {
    private final BeanDefinitionRegistry registry;
    private final AnnotatedBeanDefinitionReader annotatedReader;
    private final XmlBeanDefinitionReader xmlReader;
    private final ClassPathBeanDefinitionScanner scanner;

    BeanDefinitionLoader(BeanDefinitionRegistry registry, Object... sources) {
        this.registry = registry;
        this.annotatedReader = new AnnotatedBeanDefinitionReader(registry);
        this.xmlReader = new XmlBeanDefinitionReader(registry);
        this.scanner = new ClassPathBeanDefinitionScanner(registry, false);
        // ... setEnvironment, setResourceLoader
    }
}
```

Обрати внимание: `ClassPathBeanDefinitionScanner` создаётся с флагом `false` — без дефолтных include-фильтров. Это значит, что он не сканирует `@Component` автоматически, пока ты не вызовешь `scan()`. Сканирование произойдёт позже, когда `ConfigurationClassPostProcessor` обработает `@ComponentScan` из твоего `@SpringBootApplication`.

Как работает `load()`

```java
int load() {
    int count = 0;
    for (Object source : this.sources) {
        count += load(source);
    }
    return count;
}

private int load(Object source) {
    if (source instanceof Class<?> clazz)    return load(clazz);
    if (source instanceof Resource res)     return load(res);
    if (source instanceof Package pkg)      return load(pkg);
    if (source instanceof CharSequence cs)  return load(cs);
    throw new IllegalArgumentException("Invalid source type " + source.getClass());
}
```

Для твоего `@SpringBootApplication`-класса вызовется `load(Class<?>)`:
```java
private int load(Class<?> source) {
    if (isGroovyPresent() && GroovyBeanDefinitionSource.class.isAssignableFrom(source)) {
        // ... Groovy
    }
    if (isComponent(source)) {
        this.annotatedReader.register(source);
        return 1;
    }
    return 0;
}
```

То есть если класс помечен `@Component` (а `@SpringBootApplication` включает `@Component` через мета-аннотации) — он регистрируется через `AnnotatedBeanDefinitionReader`. Создаётся `AnnotatedGenericBeanDefinition`, в котором хранится metadata аннотаций.

> **Важно**: на этом этапе создаются только `BeanDefinition` для самого `@SpringBootApplication`-класса. Все остальные бины (сервисы, контроллеры, `@Bean`-методы) будут зарегистрированы позже — во время `refresh()`, когда `ConfigurationClassPostProcessor` обработает `@ComponentScan`, `@Import`, `@Bean` и автоконфигурацию.

### 3.5. Диаграмма: от `createApplicationContext()` до `ApplicationPreparedEvent`

```java
run()
 │
 ├─ createApplicationContext()
 │     └─ ApplicationContextFactory.create(webApplicationType)
 │           ├─ SERVLET   → AnnotationConfigServletWebServerApplicationContext
 │           ├─ REACTIVE  → AnnotationConfigReactiveWebServerApplicationContext
 │           └─ NONE      → AnnotationConfigApplicationContext
 │
 ├─ context.setApplicationStartup()
 │
 └─ prepareContext()
       │
       ├─ context.setEnvironment(environment)
       │
       ├─ postProcessApplicationContext()
       │     ├─ BeanNameGenerator
       │     ├─ ResourceLoader
       │     └─ ConversionService
       │
       ├─ applyInitializers()
       │     └─ ApplicationContextInitializer.initialize(context)
       │           (ConfigurationWarnings, ContextId, кастомные)
       │
       ├─ listeners.contextPrepared()
       │     └─ ApplicationContextInitializedEvent
       │
       ├─ logStartupProfileInfo()
       │
       ├─ beanFactory.registerSingleton("springApplicationArguments")
       ├─ beanFactory.registerSingleton("springBootBanner")
       │
       ├─ setAllowBeanDefinitionOverriding()
       │
       ├─ getAllSources() → {MyApplication.class}
       │
       ├─ load(context, sources)
       │     └─ BeanDefinitionLoader.load()
       │           └─ AnnotatedBeanDefinitionReader.register(MyApplication.class)
       │                 → BeanDefinition для @SpringBootApplication
       │
       └─ listeners.contextLoaded()
             └─ ApplicationPreparedEvent
                   │
                   ▼
             refreshContext()   ← Часть V
```

### 3.6. Spring Boot 3 vs Spring Boot 4: что изменилось

`ApplicationContextFactory` — удалён

В Spring Boot 3 `ApplicationContextFactory` — публичный интерфейс, часть `spring-boot` ядра. В Spring Boot 4 он удалён. Вместо него работает модульная система: каждый модуль (`spring-boot-webmvc`, `spring-boot-webflux`, `spring-boot-core`) сам регистрирует свой контекст через `META-INF/spring/...imports` и `spring.factories`.

`AbstractApplicationContextFactory` помечен как `@Deprecated(since="6.0", forRemoval=true)` и запланирован к удалению. То есть переходный период: в 3.x ещё есть, в 4.x — нет.

`prepareContext()` — почти без изменений

Логика `prepareContext()` в Boot 4 осталась той же. Изменилось только то, откуда берётся `ApplicationContextFactory` — если в Boot 3 он был полем `SpringApplication`, то в Boot 4 контекст создаётся через модульный `ApplicationContextFactory` из соответствующего starter'а.

`BeanDefinitionLoader` — без изменений

Этот класс не менялся ни в 3.x, ни в 4.x. Он по-прежнему package-private, по-прежнему фасад над тремя reader'ами.

**Starter'ы**

| Boot 3 | Boot 4
|--|--
| `spring-boot-starter-web` | `spring-boot-starter-webmvc`
| `spring-boot-starter-webflux` | `spring-boot-starter-webflux` (остался)


### 3.7. Ключевые моменты

1. `ApplicationContextFactory` — стратегия выбора контекста. В Boot 3 — публичный интерфейс, в Boot 4 — удалён в пользу модульности.
2. `prepareContext()` — это 10 шагов, каждый из которых можно расширить: Initializer, BeanNameGenerator, ResourceLoader, ConversionService.
3. `ApplicationContextInitializer` — единственная точка, где можно вмешаться в контекст до загрузки бинов.
4. `BeanDefinitionLoader` — не создаёт бины, а только регистрирует `BeanDefinition` для primary source. Остальное — работа `ConfigurationClassPostProcessor` в `refresh()`.
5. `springApplicationArguments` и `springBootBanner` — регистрируются как singleton напрямую, минуя `BeanPostProcessor`.
6. `ApplicationPreparedEvent` — последнее событие перед `refresh()`. После него контекст «заморожен» для внешних изменений.
7. Boot 4: `ApplicationContextFactory` удалён, контекст определяется модулями. `prepareContext()` и `BeanDefinitionLoader` — без изменений.
