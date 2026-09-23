# Spring Boot

## Часть 4. Component Scan и @SpringBootApplication

Мы дошли до момента, когда `BeanDefinitionLoader` зарегистрировал `BeanDefinition` для твоего главного класса. Но на этом магия `@SpringBootApplication` только начинается — сам класс помечен этой аннотацией, а она запускает три независимых механизма. Разберём каждый.

### 4.1. `@SpringBootApplication` — композитная аннотация

Посмотрим на её определение (Spring Boot 3.x):
```java
@Target(ElementType.TYPE)
@Retention(RetentionPolicy.RUNTIME)
@Documented
@Inherited
@SpringBootConfiguration
@EnableAutoConfiguration
@ComponentScan(excludeFilters = { 
    @Filter(type = FilterType.CUSTOM, classes = TypeExcludeFilter.class),
    @Filter(type = FilterType.CUSTOM, classes = AutoConfigurationExcludeFilter.class) 
})
public @interface SpringBootApplication {
    // алиасы для @ComponentScan
    @AliasFor(annotation = ComponentScan.class, attribute = "basePackages")
    String[] scanBasePackages() default {};
    
    @AliasFor(annotation = ComponentScan.class, attribute = "basePackageClasses")
    Class<?>[] scanBasePackageClasses() default {};
    
    // алиасы для @EnableAutoConfiguration
    @AliasFor(annotation = EnableAutoConfiguration.class)
    Class<?>[] exclude() default {};
    
    @AliasFor(annotation = EnableAutoConfiguration.class)
    String[] excludeName() default {};
}
```

**Ключевой момент**: `@SpringBootApplication` — это не «одна аннотация, которая делает всё», а три отдельные аннотации, каждая из которых запускает свой механизм. Причём каждая из них обрабатывается в разное время и разными классами.

### 4.2. Три механизма и когда они срабатывают

| Аннотация | Что делает | Когда срабатывает | Кто обрабатывает
|--|--|--|--
| `@SpringBootConfiguration` | Помечает класс как `@Configuration` | `ConfigurationClassPostProcessor` | `ConfigurationClassParser`
| `@ComponentScan` | Ищет `@Component` в пакете | `ConfigurationClassPostProcessor` | `ComponentScanAnnotationParser`
| `@EnableAutoConfiguration` | Загружает автоконфигурации | `ConfigurationClassPostProcessor` | `AutoConfigurationImportSelector`

Все три обрабатываются внутри `ConfigurationClassPostProcessor` — это `BeanFactoryPostProcessor`, который вызывается в `refreshContext()` на этапе `invokeBeanFactoryPostProcessors()`. То есть до создания обычных бинов.

Диаграмма:
```java
refreshContext()
  └─ AbstractApplicationContext.refresh()
        └─ invokeBeanFactoryPostProcessors()
              └─ ConfigurationClassPostProcessor.postProcessBeanDefinitionRegistry()
                    └─ ConfigurationClassParser.parse()
                          ├─ @SpringBootConfiguration → registerAsConfigClass()
                          ├─ @ComponentScan         → doProcessConfigurationClass() → scan()
                          └─ @EnableAutoConfiguration → AutoConfigurationImportSelector
```

### 4.3. `@SpringBootConfiguration` — «я конфигурация»
```java
@Target(ElementType.TYPE)
@Retention(RetentionPolicy.RUNTIME)
@Documented
@Configuration
@Indexed
public @interface SpringBootConfiguration {
    @AliasFor(annotation = Configuration.class)
    boolean proxyBeanMethods() default true;
}
```

Это просто `@Configuration` с меткой «я — главный конфиг Spring Boot». Зачем отдельная аннотация? Чтобы:
1. Отличить главный класс от других `@Configuration` (используется в тестах: `@SpringBootTest` ищет именно `@SpringBootConfiguration`).
2. Добавить `@Indexed` (об этом ниже).

`proxyBeanMethods = true` (по умолчанию) означает, что класс будет обёрнут в CGLIB-прокси, и вызовы `@Bean`-методов внутри класса будут перехвачены — это гарантирует, что бин создаётся один раз (singleton-семантика).

### 4.4. `@ComponentScan` — как работает сканирование

**4.4.1. Откуда берётся базовый пакет**
```java
@Target(ElementType.TYPE)
@Retention(RetentionPolicy.RUNTIME)
@Documented
@Repeatable(ComponentScans.class)
public @interface ComponentScan {
    @AliasFor("basePackages")
    String[] value() default {};
    
    @AliasFor("value")
    String[] basePackages() default {};
    
    Class<?>[] basePackageClasses() default {};
    
    // ...
}
```

Если ни `basePackages`, ни `basePackageClasses` не указаны — сканирование идёт от пакета класса, на котором стоит аннотация.

```java
package com.example.myapp;   // ← корневой пакет

@SpringBootApplication
public class MyApplication { }
```

Всё, что лежит в `com.example.myapp` и вложенных пакетах (`com.example.myapp.service`, `com.example.myapp.web` и т.д.) — будет найдено.

**4.4.2. Почему «корневой пакет» — это важно**

Если положить главный класс в `com.example.myapp`, а сервис — в `com.other.service`, он не найдётся. Это не ошибка Spring, а стандартное поведение сканирования.

Правильно:
```
com.example.myapp
  ├─ MyApplication.java      ← @SpringBootApplication
  ├─ service/
  │    └─ UserService.java   ← @Service  ✅ найдётся
  └─ web/
       └─ UserController.java ← @Controller ✅ найдётся
```

Неправильно:
```
com.example.myapp
  └─ MyApplication.java

com.other.service            ← вне корневого пакета
  └─ UserService.java        ← @Service  ❌ НЕ найдётся
```

Исключение — если явно указать `scanBasePackages`:
```java
@SpringBootApplication(scanBasePackages = {"com.example.myapp", "com.other.service"})
```

**4.4.3. `basePackageClasses` — type-safe альтернатива**

Вместо строки можно передать класс-маркер:
```java
@SpringBootApplication(
    scanBasePackageClasses = {MyApplication.class, AnotherMarker.class}
)
```

Пакет определится автоматически из пакета переданных классов. Это безопаснее: при рефакторинге (переименовании пакета) компилятор поймает ошибку.

**4.4.4. Что именно ищет сканер**

`ClassPathBeanDefinitionScanner` ищет классы, помеченные стереотипными аннотациями:

| Аннотация | Семантика
|--|--
| `@Component` | базовый стереотип
| `@Service` | бизнес-логика
| `@Repository` | доступ к данным (+ трансляция исключений)
| `@Controller` | MVC-контроллер
| `@RestController` | `@Controller` + `@ResponseBody`
| `@Configuration` | конфигурационный класс (тоже компонент)

Важно: все они мета-аннотированы `@Component`. Сканер ищет именно `@Component` по цепочке мета-аннотаций.

**4.4.5. Механика сканирования — `ClassPathScanningCandidateComponentProvider`**

Этот класс делает основную работу:

```java
public class ClassPathScanningCandidateComponentProvider {
    private final List<TypeFilter> includeFilters = new ArrayList<>();
    private final List<TypeFilter> excludeFilters = new ArrayList<>();
    
    public Set<BeanDefinition> findCandidateComponents(String basePackage) {
        // 1. Собрать все .class-файлы в пакете
        // 2. Для каждого прочитать metadata (через ASM, не через reflection!)
        // 3. Применить include/exclude фильтры
        // 4. Вернуть совпавшие BeanDefinition
    }
}
```

**Ключевой момент:** metadata читается через ASM (`SimpleMetadataReader`), а не через `Class.forName()`. Это значит, что сканер не загружает классы на этапе сканирования — он читает только заголовки `.class`-файлов. Reflection используется позже, при создании бинов.

### 4.5. Фильтры: `includeFilters` и `excludeFilters`

**4.5.1. `FilterType` — пять стратегий**

```java
public enum FilterType {
    ANNOTATION,        // по аннотации
    ASSIGNABLE_TYPE,   // по типу (класс/интерфейс)
    ASPECTJ,           // по AspectJ-выражению
    REGEX,             // по regex на имя класса
    CUSTOM             // свой TypeFilter
}
```

Примеры:
```java
// ANNOTATION: включить всё, что помечено @MyService
@ComponentScan(includeFilters = @Filter(type = FilterType.ANNOTATION, 
    classes = MyService.class))

// ASSIGNABLE_TYPE: включить всё, что реализует Animal
@ComponentScan(includeFilters = @Filter(type = FilterType.ASSIGNABLE_TYPE, 
    classes = Animal.class))

// CUSTOM: своя логика
@ComponentScan(includeFilters = @Filter(type = FilterType.CUSTOM, 
    classes = MyCustomFilter.class))
```

При указании нескольких классов в одном `@Filter` применяется логика OR: «тип помечен `@Foo` ИЛИ `@Bar`».

**4.5.2. `useDefaultFilters` — важный флаг**

```java
@ComponentScan(
    useDefaultFilters = false,
    includeFilters = @Filter(type = FilterType.ANNOTATION, 
        classes = MyService.class)
)
```

По умолчанию `useDefaultFilters = true`, то есть сканер автоматически включает `@Component`, `@Service`, `@Repository`, `@Controller`. Если хочешь искать только по своим аннотациям — ставь `false`.

**4.5.3. `TypeExcludeFilter` — расширяемый фильтр Spring Boot**

Ты помнишь, что `@SpringBootApplication` содержит:

```java
@ComponentScan(excludeFilters = { 
    @Filter(type = CUSTOM, classes = TypeExcludeFilter.class),
    @Filter(type = CUSTOM, classes = AutoConfigurationExcludeFilter.class) 
})
```

`TypeExcludeFilter` — это точка расширения. Он не фильтрует ничего сам по себе. Вместо этого он при старте получает из `BeanFactory` все бины, реализующие `TypeExcludeFilter`, и применяет их `match()`-методы:

```java
public class TypeExcludeFilter implements TypeFilter, BeanFactoryAware {
    @Override
    public boolean match(MetadataReader metadataReader, 
            MetadataReaderFactory metadataReaderFactory) {
        if (this.beanFactory instanceof ListableBeanFactory lbf) {
            Collection<TypeExcludeFilter> filters = lbf.getBeansOfType(
                TypeExcludeFilter.class).values();
            return filters.stream().anyMatch(filter -> 
                filter.match(metadataReader, metadataReaderFactory));
        }
        return false;
    }
}
```

Это позволяет добавлять свои фильтры через `@Bean`:

```java
@Bean
MyExcludeFilter myExcludeFilter() {
    return new MyExcludeFilter();
}
```

Твой фильтр будет автоматически применён при сканировании. Это мощный, но малоизвестный механизм.

**4.5.4. `AutoConfigurationExcludeFilter` — второй встроенный фильтр**

Этот фильтр исключает из сканирования автоконфигурации. Логика:

```java
@Override
public boolean match(MetadataReader metadataReader, 
        MetadataReaderFactory factory) {
    return isConfiguration(metadataReader) && isAutoConfiguration(metadataReader);
}
```

То есть если класс — `@Configuration` и является автоконфигурацией (зарегистрирован в `.imports / spring.factories`), он не подхватывается компонентным сканированием. Автоконфигурации загружаются отдельно, через `@EnableAutoConfiguration`. Без этого фильтра они бы дублировались.

### 4.6. `@AutoConfigurationPackage` — «скрытая» аннотация

Посмотрим ещё раз на определение `@EnableAutoConfiguration`:

```java
@Target(ElementType.TYPE)
@Retention(RetentionPolicy.RUNTIME)
@Documented
@Inherited
@AutoConfigurationPackage
@Import(AutoConfigurationImportSelector.class)
public @interface EnableAutoConfiguration {
    // ...
}
```

`@AutoConfigurationPackage` — это то, что запоминает пакет твоего главного класса:
```java
@Target(ElementType.TYPE)
@Retention(RetentionPolicy.RUNTIME)
@Documented
@Inherited
@Import(AutoConfigurationPackages.Registrar.class)
public @interface AutoConfigurationPackage { }
```

`AutoConfigurationPackages.Registrar` регистрирует бин `AutoConfigurationPackages.BasePackages` со списком базовых пакетов. Зачем? Чтобы другие автоконфигурации знали, где искать твои сущности:

* **Spring Data JPA** (`@EntityScan`) — использует базовый пакет для поиска `@Entity`.
* **Spring Data MongoDB** — аналогично.
* **Spring Boot DevTools** — для перезапуска.

Если положить `@Entity` вне корневого пакета, JPA её не найдёт — потому что `@AutoConfigurationPackage` запомнил только `com.example.myapp`.

### 4.7. @Indexed и META-INF/spring.components — ускорение сканирования

**4.7.1. Проблема**
При старте Spring Boot сканирует classpath. В больших проектах это может занимать секунды. Особенно если в classpath много JAR'ов.

**4.7.2. Решение: `@Indexed`**

Spring Framework 5.0 ввёл `@Indexed`:
```java
@Target(ElementType.TYPE)
@Retention(RetentionPolicy.RUNTIME)
public @interface Indexed { }
```

Все стереотипные аннотации (`@Component`, `@Service`, `@Repository`, `@Controller`, `@Configuration`, `@SpringBootConfiguration`) мета-аннотированы `@Indexed`.

**4.7.3. Как это работает**

При компиляции annotation processor (spring-context-indexer) генерирует файл:
```
META-INF/spring.components
```

Пример содержимого:
```
com.example.myapp.service.UserService=org.springframework.stereotype.Component
com.example.myapp.web.UserController=org.springframework.stereotype.Component
com.example.myapp.config.AppConfig=org.springframework.context.annotation.Configuration
```

При старте `CandidateComponentsIndexLoader` читает этот файл и создаёт `CandidateComponentsIndex`. `ClassPathScanningCandidateComponentProvider` вместо сканирования classpath просто читает индекс.

**4.7.4. Как включить**

Вариант 1 — аннотация на главном классе:
```java
@SpringBootApplication
@Indexed
public class MyApplication { }
```

Вариант 2 — зависимость + annotation processor:
```xml
<dependency>
    <groupId>org.springframework</groupId>
    <artifactId>spring-context-indexer</artifactId>
    <optional>true</optional>
</dependency>
```

> **Важный нюанс**: `@Indexed` включает индекс. Если индекс есть, сканирование идёт по нему. Если индекс есть, но ты добавил новый `@Component` и не перекомпилировал — он не найдётся. Индекс нужно перестраивать при каждой сборке.

### 4.8. Диаграмма: полный путь `@ComponentScan`

```
ConfigurationClassPostProcessor.postProcessBeanDefinitionRegistry()
  │
  └─ ConfigurationClassParser.parse()
        │
        ├─ processConfigurationClass(MyApplication.class)
        │     │
        │     ├─ @SpringBootConfiguration → registerAsConfigClass()
        │     │
        │     ├─ @ComponentScan → doProcessConfigurationClass()
        │     │     │
        │     │     ├─ ComponentScanAnnotationParser.parse()
        │     │     │     ├─ basePackages = {pkg(MyApplication)}
        │     │     │     ├─ excludeFilters = [TypeExcludeFilter, AutoConfigurationExcludeFilter]
        │     │     │     └─ useDefaultFilters = true
        │     │     │
        │     │     └─ ClassPathBeanDefinitionScanner.doScan(basePackages)
        │     │           │
        │     │           ├─ CandidateComponentsIndexLoader.loadIndex()   ← @Indexed
        │     │           │     └─ если индекс есть → читаем spring.components
        │     │           │
        │     │           └─ findCandidateComponents(basePackage)
        │     │                 ├─ для каждого .class:
        │     │                 │     ├─ SimpleMetadataReader (ASM)
        │     │                 │     ├─ isCandidateComponent() → include/exclude
        │     │                 │     └─ если прошёл → BeanDefinition
        │     │                 │
        │     │                 └─ registerBeanDefinition()
        │     │
        │     └─ @EnableAutoConfiguration → AutoConfigurationImportSelector
        │           └─ загрузка автоконфигураций   ← Часть IV
        │
        └─ ...
```

### 4.9. Spring Boot 3 vs Spring Boot 4

| Аспект | Spring Boot 3.x | Spring Boot 4.x
|--|--|--
| `@ComponentScan` | Без изменений | Без изменений
| `@SpringBootApplication` | Без изменений | Без изменений
| Новый механизм | — | `BeanRegistrar` — программная регистрация бинов как альтернатива сканированию
| Модульность | `spring-boot-autoconfigure` — монолит | 47 модульных JAR'ов, каждый со своим сканированием
| Starter'ы | `spring-boot-starter-web` | `spring-boot-starter-webmvc`
| `@Indexed` | Работает | Работает, но менее актуален из-за AOT

**Главное в Boot 4:** базовое поведение @ComponentScan и @SpringBootApplication не изменилось. Но появился BeanRegistrar — альтернативный способ регистрации бинов без сканирования:
```java
public class MyBeanRegistrar implements BeanRegistrar {
    @Override
    public void register(BeanRegistry registry, Environment env) {
        if (env.acceptsProfiles(Profiles.of("dev"))) {
            registry.registerBean("myService", MyService.class);
        }
    }
}
```

Это ближе к AOT-подходу: бины регистрируются программно, а не ищутся сканированием.

### 4.10. Что запомнить
1. `@SpringBootApplication` = три аннотации, обрабатываемые в `ConfigurationClassPostProcessor` во время `refresh()`.
2. Базовый пакет для сканирования = пакет главного класса. Всё, что вне его — не найдётся без явного `scanBasePackages`.
3. `ClassPathBeanDefinitionScanner` читает metadata через ASM, не загружая классы.
4. `TypeExcludeFilter` — расширяемая точка: можно добавить свой фильтр через `@Bean`.
5. `AutoConfigurationExcludeFilter` — исключает автоконфигурации из сканирования, чтобы не дублировать их.
6. `@AutoConfigurationPackage` — запоминает базовый пакет для JPA/Data и других автоконфигураций.
7. `@Indexed` + `META-INF/spring.components` — ускоряет сканирование в больших проектах.
8. **Boot 4**: базовое поведение без изменений, но появился `BeanRegistrar` как альтернатива сканированию.
