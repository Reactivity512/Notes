# Spring Boot

## Часть 5. Auto-configuration

Auto-configuration — это сердце Spring Boot. Именно она позволяет добавить `spring-boot-starter-web` в `pom.xml`, и у тебя автоматически появляется Tomcat, DispatcherServlet, Jackson и куча других бинов без единой строчки конфигурации.

Разберём, как это работает «до винтика».

### 5.1. Где живёт точка входа

Ты уже знаешь, что `@SpringBootApplication` включает `@EnableAutoConfiguration`. Посмотрим на неё:
```java
@Target(ElementType.TYPE)
@Retention(RetentionPolicy.RUNTIME)
@Documented
@Inherited
@AutoConfigurationPackage
@Import(AutoConfigurationImportSelector.class)
public @interface EnableAutoConfiguration {
    Class<?>[] exclude() default {};
    String[] excludeName() default {};
}
```

Два ключевых момента:
* `@AutoConfigurationPackage` — запоминает пакет главного класса (мы разбирали в Части III).
* `@Import(AutoConfigurationImportSelector.class)` — импортирует селектор, который и делает всю магию.

`@Import` — это механизм Spring Framework, который позволяет программно добавить BeanDefinition-ы в контекст, минуя сканирование. Селектор реализует ImportSelector и возвращает массив имён классов, которые нужно зарегистрировать

### 5.2. `AutoConfigurationImportSelector` — главный дирижёр

```java
public class AutoConfigurationImportSelector 
        implements DeferredImportSelector, BeanClassLoaderAware, 
                   ResourceLoaderAware, BeanFactoryAware, EnvironmentAware, Ordered {
```

**Обрати внимание**: это `DeferredImportSelector`, а не просто `ImportSelector`. Разница принципиальная.

**5.2.1. DeferredImportSelector vs ImportSelector**

| Тип | Когда обрабатывается
|--|--
| `ImportSelector` | Сразу при парсинге `@Configuration`
| `DeferredImportSelector` | После обработки всех обычных `@Configuration`

Это гарантирует, что автоконфигурации обрабатываются после пользовательских бинов. Именно поэтому `@ConditionalOnMissingBean` в автоконфигурациях работает корректно — к моменту их обработки все твои `@Bean` уже зарегистрированы.

**5.2.2. `selectImports()` — точка входа**

```java
@Override
public String[] selectImports(AnnotationMetadata importingClassMetadata) {
    AutoConfigurationEntry autoConfigurationEntry = 
        getAutoConfigurationEntry(annotationMetadata);
    return StringUtils.toStringArray(autoConfigurationEntry.getConfigurations());
}
```

Метод `getAutoConfigurationEntry()` — центральный. Разберём его.

### 5.3. `getAutoConfigurationEntry()` — пошаговый разбор

```java
protected AutoConfigurationEntry getAutoConfigurationEntry(
        AnnotationMetadata annotationMetadata) {
    
    // 1. Проверка: включена ли автоконфигурация вообще
    if (!isEnabled(annotationMetadata)) {
        return EMPTY_ENTRY;
    }
    
    // 2. Получение атрибутов @EnableAutoConfiguration (exclude, excludeName)
    AnnotationAttributes attributes = getAttributes(annotationMetadata);
    
    // 3. Загрузка ВСЕХ кандидатов из classpath
    List<String> configurations = getCandidateConfigurations(
        annotationMetadata, attributes);
    
    // 4. Дедупликация
    configurations = removeDuplicates(configurations);
    
    // 5. Применение exclude
    Set<String> exclusions = getExclusions(annotationMetadata, attributes);
    checkExcludedClasses(configurations, exclusions);
    configurations.removeAll(exclusions);
    
    // 6. Фильтрация по условиям (ConfigurationClassFilter)
    configurations = getConfigurationClassFilter().filter(configurations);
    
    // 7. Публикация события AutoConfigurationImportEvent
    fireAutoConfigurationImportEvents(configurations, exclusions);
    
    return new AutoConfigurationEntry(configurations, exclusions);
}
```

Ключевые шаги — 3 и 6.

### 5.4. Загрузка кандидатов — .imports вместо spring.factories

**5.4.1. Историческая справка**

| Версия | Механизм
|--|--
| Spring Boot 1.x – 2.6 | `META-INF/spring.factories` (ключ EnableAutoConfiguration)
| Spring Boot 2.7 | Оба механизма параллельно (обратная совместимость)
| Spring Boot 3.0+ | Только `.imports`, `spring.factories` для автоконфигураций удалён

Это подтверждается официальной документацией и release notes: «Spring Boot 2.7 introduced a new `META-INF/spring/org.springframework.boot.autoconfigure.AutoConfiguration.imports` file for registering auto-configurations, while maintaining backwards compatibility with registration in `spring.factories`. With this release, support for registering auto-configurations in `spring.factories` has been removed in favor of the imports file».

**5.4.2. Как выглядит `.imports`**

Путь к файлу:
```
META-INF/spring/org.springframework.boot.autoconfigure.AutoConfiguration.imports
```

Формат — одна строка = один класс:
```
org.springframework.boot.autoconfigure.admin.SpringApplicationAdminJmxAutoConfiguration
org.springframework.boot.autoconfigure.aop.AopAutoConfiguration
org.springframework.boot.autoconfigure.web.servlet.WebMvcAutoConfiguration
org.springframework.boot.autoconfigure.jackson.JacksonAutoConfiguration
...
```

Для сравнения, старый `spring.factories` использовал key-value:
```
org.springframework.boot.autoconfigure.EnableAutoConfiguration=\
org.springframework.boot.autoconfigure.aop.AopAutoConfiguration,\
org.springframework.boot.autoconfigure.web.servlet.WebMvcAutoConfiguration
```

Преимущества `.imports`: проще парсинг, нет проблем с escaping, лучше работает с AOT.

**5.4.3. `ImportCandidates` — загрузчик**

Загрузка кандидатов происходит через `ImportCandidates.load()`:
```java
public static ImportCandidates load(Class<?> annotation, ClassLoader classLoader) {
    String location = String.format(
        "META-INF/spring/%s.imports", annotation.getName());
    List<String> candidates = new ArrayList<>();
    // ... чтение всех файлов по этому пути из всех JAR'ов
    return new ImportCandidates(candidates);
}
```

Ключевой момент: `ImportCandidates` не использует `SpringFactoriesLoader` (который читает `spring.factories`). Это отдельный механизм, оптимизированный под AOT и нативный образ.

**5.4.4. `@AutoConfiguration` — новая аннотация**

Начиная с Spring Boot 2.7, классы автоконфигурации помечаются `@AutoConfiguration` вместо `@Configuration`:

```java
@Target(ElementType.TYPE)
@Retention(RetentionPolicy.RUNTIME)
@Documented
@Configuration(proxyBeanMethods = false)
@AutoConfigureBefore
@AutoConfigureAfter
public @interface AutoConfiguration {
    @AliasFor(annotation = AutoConfigureBefore.class, attribute = "value")
    Class<?>[] before() default {};
    
    @AliasFor(annotation = AutoConfigureAfter.class, attribute = "value")
    Class<?>[] after() default {};
}
```

Три ключевых отличия от `@Configuration`:

1. `proxyBeanMethods = false` — автоконфигурации не проксируются CGLIB. Это быстрее и не создаёт circular dependencies.
2. Наследует `@AutoConfigureBefore` / `@AutoConfigureAfter` — порядок можно указывать прямо в аннотации.
3. Семантически отделена от пользовательских `@Configuration`.

### 5.5. Условные аннотации (`@Conditional*`) — фильтр кандидатов

После загрузки всех кандидатов (в Spring Boot 3.x их около 150+) начинается фильтрация. Именно здесь работают условные аннотации.

**5.5.1. Как это устроено**

Фильтрация происходит в `ConfigurationClassFilter`:

```java
class ConfigurationClassFilter {
    private final List<AutoConfigurationImportFilter> filters;
    
    List<String> filter(List<String> configurations) {
        String[] candidates = configurations.toArray(new String[0]);
        boolean[] skip = new boolean[candidates.length];
        boolean skipped = false;
        for (AutoConfigurationImportFilter filter : this.filters) {
            boolean[] match = filter.match(candidates, this.autoConfigurationMetadata);
            for (int i = 0; i < match.length; i++) {
                if (!match[i]) {
                    skip[i] = true;
                    skipped = true;
                }
            }
        }
        // ... возврат отфильтрованного списка
    }
}
```

`AutoConfigurationImportFilter` — SPI-интерфейс, реализации которого загружаются через `spring.factories` (да, `spring.factories` ещё используется для внутренних SPI, просто не для регистрации автоконфигураций). Spring Boot поставляет три фильтра:

| Фильтр | Что проверяет
|--|--
| `OnClassCondition` | `@ConditionalOnClass` / `@ConditionalOnMissingClass`
| `OnBeanCondition` | `@ConditionalOnBean` / `@ConditionalOnMissingBean`
| `OnWebApplicationCondition` | `@ConditionalOnWebApplication` / `@ConditionalOnNotWebApplication`

**Важно**: это первый этап фильтрации — грубый. Он проверяет только наличие классов и бинов, не загружая сами автоконфигурации. Более тонкие проверки (property, resource, expression) происходят позже, когда `ConfigurationClassParser` уже обрабатывает конкретный класс.

**5.5.2. Основные условные аннотации**

`@ConditionalOnClass `/ `@ConditionalOnMissingClass`

```java
@AutoConfiguration
@ConditionalOnClass(DataSource.class)   // активируется, только если DataSource есть в classpath
public class DataSourceAutoConfiguration { }
```

Проверка происходит через ASM — класс не загружается в JVM, читается только метаданные `.class`-файла. Поэтому можно безопасно ссылаться на классы, которых может не быть в runtime.

`@ConditionalOnBean` / `@ConditionalOnMissingBean`

```java
@Bean
@ConditionalOnMissingBean
public DataSource dataSource() {
    return new HikariDataSource();
}
```

**Ключевой механизм кастомизации**. Если ты объявил свой `DataSource` — автоконфигурация не создаст свой. Именно поэтому автоконфигурации обрабатываются как `DeferredImportSelector` — чтобы все твои бины уже были зарегистрированы к моменту проверки.

> **Нюанс**: `@ConditionalOnMissingBean` работает только для `@Bean`-методов и `@Configuration`-классов, которые обрабатываются после твоих. Если поместить его на обычный `@Configuration`, результат может быть непредсказуемым — `@Conditional` обрабатываются при парсинге `@Configuration`, и порядок парсинга может отличаться от порядка регистрации бинов.

`@ConditionalOnProperty`

```java
@AutoConfiguration
@ConditionalOnProperty(
    prefix = "spring.datasource", 
    name = "url", 
    matchIfMissing = false
)
public class DataSourceAutoConfiguration { }
```

Проверяет значение property. `matchIfMissing = true` означает «активировать, если property вообще отсутствует».

`@ConditionalOnResource`
```java
@ConditionalOnResource(resources = "classpath:my-config.properties")
```

`@ConditionalOnWebApplication` / `@ConditionalOnNotWebApplication`
```java
@ConditionalOnWebApplication(type = Type.SERVLET)
```

`@ConditionalOnExpression`
```java
@ConditionalOnExpression("${my.feature.enabled:false} and '${my.mode}' == 'advanced'")
```

Использует SpEL. Не рекомендуется для автоконфигураций — медленно и плохо работает с AOT.

**5.5.3. `ConfigurationPhase` — тонкость***

`@Conditional` может обрабатываться на двух фазах:
```java
public enum ConfigurationPhase {
    PARSE_CONFIGURATION,  // при парсинге @Configuration
    REGISTER_BEAN         // при регистрации бина
}
```

`@ConditionalOnBean` и `@ConditionalOnMissingBean` используют `REGISTER_BEAN`, потому что им нужно знать, какие бины уже зарегистрированы. Остальные (`@ConditionalOnClass`, `@ConditionalOnProperty`) — `PARSE_CONFIGURATION`.

### 5.6. Порядок автоконфигураций
Когда у тебя 150+ автоконфигураций, порядок их применения критичен. Например, `DataSourceAutoConfiguration` должен быть до `JpaRepositoriesAutoConfiguration`.

**5.6.1. `@AutoConfigureBefore` / `@AutoConfigureAfter`**
```java
@AutoConfiguration(after = DataSourceAutoConfiguration.class)
public class JpaRepositoriesAutoConfiguration { }
```

Или через отдельные аннотации:
```java
@AutoConfiguration
@AutoConfigureAfter(DataSourceAutoConfiguration.class)
public class JpaRepositoriesAutoConfiguration { }
```

Эти аннотации не гарантируют абсолютный порядок — они лишь задают относительный. Spring Boot сортирует граф зависимостей через `AutoConfigurationSorter`.

**5.6.2. `@AutoConfigureOrder`**
```java
@AutoConfiguration
@AutoConfigureOrder(Ordered.HIGHEST_PRECEDENCE)
public class MyEarlyAutoConfiguration { }
```

Семантика как у `@Order`, но для автоконфигураций. Используется, когда автоконфигурации не знают друг о друге и не могут ссылаться через `before`/`after`.

**5.6.3. `AutoConfigurationSorter` — как работает сортировка**
```java
class AutoConfigurationSorter {
    void sort(List<String> classNames) {
        // 1. Читает метаданные каждого класса (ASM)
        // 2. Строит граф: A before B, A after C
        // 3. Топологическая сортировка
        // 4. Разрешает конфликты через @AutoConfigureOrder
    }
}
```

Приоритет: `@AutoConfigureBefore` / `@AutoConfigureAfter` > `@AutoConfigureOrder`.

### 5.7. Создание своей автоконфигурации

**5.7.1. Класс автоконфигурации**

```java
@AutoConfiguration
@ConditionalOnClass(MyService.class)
@EnableConfigurationProperties(MyServiceProperties.class)
public class MyServiceAutoConfiguration {

    @Bean
    @ConditionalOnMissingBean
    public MyService myService(MyServiceProperties properties) {
        return new MyService(properties.getEndpoint(), properties.getTimeout());
    }
}
```

**5.7.2. Properties**
```java
@ConfigurationProperties(prefix = "my.service")
public class MyServiceProperties {
    private String endpoint = "http://localhost:8080";
    private Duration timeout = Duration.ofSeconds(5);
    // getters/setters
}
```

**5.7.3. Регистрация в `.imports`**

Файл:
```
src/main/resources/META-INF/spring/org.springframework.boot.autoconfigure.AutoConfiguration.imports
```

Содержимое:
```
com.example.autoconfigure.MyServiceAutoConfiguration
```

**5.7.4. Starter**
Чтобы пользователи могли добавить одну зависимость, создаётся starter — пустой JAR, который тянет:
```xml
<dependency>
    <groupId>com.example</groupId>
    <artifactId>my-service-spring-boot-starter</artifactId>
</dependency>
```

Внутри `pom.xml` starter'а:
```xml
<dependencies>
    <dependency>
        <groupId>com.example</groupId>
        <artifactId>my-service</artifactId>
    </dependency>
    <dependency>
        <groupId>com.example</groupId>
        <artifactId>my-service-spring-boot-autoconfigure</artifactId>
    </dependency>
</dependencies>
```

Соглашение об именовании: `xxx-spring-boot-starter` (для пользователей) и `xxx-spring-boot-autoconfigure` (для кода автоконфигурации).

### 5.8. Spring Boot 3 vs Spring Boot 4: ключевые отличия

**5.8.1. Модуляризация spring-boot-autoconfigure**

Это самое большое изменение в Boot 4.

| Аспект | Spring Boot 3.x | Spring Boot 4.x
|--|--|--
| Артефакт | Один монолитный `spring-boot-autoconfigure` (2 MiB в 3.5) | 47 модульных JAR'ов, каждый для своей технологии
| Размер | Всё в одном месте | Только нужные модули попадают в classpath
| Автоконфигурации Web | В `spring-boot-autoconfigure` | В `spring-boot-webmvc` / `spring-boot-webflux`
| Автоконфигурации JPA | В `spring-boot-autoconfigure` | В `spring-boot-data-jpa`
| IDE-подсказки | Все классы всех технологий | Только те, что реально используются

Официальный анонс объясняет мотивацию: «Instead of a single, monolithic `spring-boot-autoconfigure` jar, we are now splitting functionality into small and more focused modules. This change is motivated by maintainability, clarity, and a leaner runtime footprint».

**5.8.2. Что осталось без изменений**

* `@EnableAutoConfiguration` — работает так же.
* `AutoConfigurationImportSelector` — тот же механизм.
* `.imports`-файлы — формат тот же.
* `@Conditional*` — те же аннотации.
* `@AutoConfigureBefore` / `@AutoConfigureAfter` — та же семантика.

**5.8.3. BeanRegistrar — новая альтернатива**

`BeanRegistrar` появился в Spring Framework 7 / Boot 4 как альтернатива `@Bean`-методам и `BeanDefinitionRegistryPostProcessor` для программной регистрации бинов.

```java
public class MyBeanRegistrar implements BeanRegistrar {
    @Override
    public void register(BeanRegistry registry, Environment env) {
        if (env.acceptsProfiles(Profiles.of("dev"))) {
            registry.registerBean("myService", MyService.class, 
                spec -> spec.supplier(ctx -> new MyService("dev")));
        }
    }
}
```

Регистрируется через `@Import(MyBeanRegistrar.class)`. Это ближе к AOT-подходу: бины регистрируются программно, а не через reflection.

**5.8.4. Starter'ы**

| Boot 3 | Boot 4
|--|--
| `spring-boot-starter-web` | `spring-boot-starter-webmvc`
| `spring-boot-starter-webflux` | `spring-boot-starter-webflux` (остался)
| `spring-boot-starter-data-jpa` | `spring-boot-starter-data-jpa` (остался)

### 5.9. Диаграмма: полный путь автоконфигурации

```
@SpringBootApplication
  └─ @EnableAutoConfiguration
        └─ @Import(AutoConfigurationImportSelector.class)
              │
              └─ refresh() → invokeBeanFactoryPostProcessors()
                    └─ ConfigurationClassPostProcessor
                          └─ ConfigurationClassParser
                                └─ DeferredImportSelectorGroupingHandler
                                      └─ AutoConfigurationImportSelector.selectImports()
                                            │
                                            ├─ 1. isEnabled()?
                                            ├─ 2. getAttributes() → exclude/excludeName
                                            ├─ 3. getCandidateConfigurations()
                                            │     └─ ImportCandidates.load()
                                            │           └─ META-INF/spring/...AutoConfiguration.imports
                                            │                 (все JAR'ы)
                                            ├─ 4. removeDuplicates()
                                            ├─ 5. getExclusions() → removeAll()
                                            ├─ 6. ConfigurationClassFilter.filter()
                                            │     ├─ OnClassCondition
                                            │     ├─ OnBeanCondition
                                            │     └─ OnWebApplicationCondition
                                            └─ 7. fireAutoConfigurationImportEvents()
                                                  │
                                                  ▼
                                            Список имён классов
                                                  │
                                                  ▼
                                            ConfigurationClassParser обрабатывает
                                            каждый класс как @Configuration
                                                  │
                                                  ▼
                                            @Bean-методы → BeanDefinition
                                            @Conditional* → финальная фильтрация
```

### 5.10. Ключевые моменты

1. `@EnableAutoConfiguration` → `AutoConfigurationImportSelector` — единственная точка входа. Селектор — `DeferredImportSelector`, поэтому автоконфигурации обрабатываются после пользовательских бинов.

2. `.imports`-файлы (Boot 3+) заменили `spring.factories` для автоконфигураций. Путь: `META-INF/spring/org.springframework.boot.autoconfigure.AutoConfiguration.imports`.

3. `@AutoConfiguration` — новая аннотация для классов автоконфигурации. `proxyBeanMethods = false`, наследует `@AutoConfigureBefore/After`.

4. `@Conditional*` — работают в две фазы: грубая фильтрация через `AutoConfigurationImportFilter` (по классам/бинам), затем тонкая через `ConfigurationClassParser`.

5. `@ConditionalOnMissingBean` — механизм «бэк-оффа»: если ты объявил свой бин, автоконфигурация отступит.

6. Порядок задаётся `@AutoConfigureBefore` / `@AutoConfigureAfter` (относительный) и `@AutoConfigureOrder` (абсолютный). Сортировка — через `AutoConfigurationSorter`.

7. Boot 4 — модуляризация: 47 модулей вместо одного `spring-boot-autoconfigure`. Механизм автоконфигурации не изменился, изменилась только упаковка.

8. Своя автоконфигурация = @AutoConfiguration + @ConditionalOnClass + `@ConditionalOnMissingBean` + `.imports`-файл + starter.
