# Spring Boot

## Часть 6. Жизненный цикл бинов

К этому моменту у нас есть:

* Готовый `ApplicationContext` (Часть III)
* Зарегистрированные `BeanDefinition` от сканирования (Раздел 4)
* Загруженные автоконфигурации (Раздел 5)

Теперь `refresh()` должен превратить все эти `BeanDefinition` в **живые объекты** — бины. Разберём по шагам.

### 6.1. BeanDefinition — «чертёж» бина

Прежде чем говорить о жизненном цикле, надо понять, что такое `BeanDefinition`.
```java
public interface BeanDefinition extends AttributeAccessor, BeanMetadataElement {
    String getBeanClassName();
    String getScope();
    boolean isLazyInit();
    boolean isPrimary();
    String[] getDependsOn();
    boolean isAutowireCandidate();
    ConstructorArgumentValues getConstructorArgumentValues();
    MutablePropertyValues getPropertyValues();
    String getInitMethodName();
    String getDestroyMethodName();
    int getRole();
    // ...
}
```

`BeanDefinition` — это метаданные о бине, а не сам бин. Из него Spring узнаёт:

- Какой класс инстанциировать
- Какой scope (singleton / prototype / request / session)
- Что инжектить (constructor args / property values)
- Какой init/destroy method
- Lazy или нет

Иерархия реализаций:

| Класс | Когда используется
|--|--
| `RootBeanDefinition` | Основная реализация (merge-результат)
| `ChildBeanDefinition` | Устарел, для parent-child
| `GenericBeanDefinition` | Современная (используется в конфигурациях)
| `AnnotatedGenericBeanDefinition` | Для классов с аннотациями (`@Configuration`, `@Component`)
| `ScannedGenericBeanDefinition` | Для классов, найденных сканированием
| `ConfigurationClassBeanDefinition` | Для `@Bean`-методов

**Ключевая особенность**: `BeanDefinition` может быть parent для других `BeanDefinition`. При создании бина Spring делает `merge` — объединяет parent и child в RootBeanDefinition. Это называется `MergedBeanDefinition`.

### 6.2. Полный жизненный цикл бина — обзор

Вот полная схема. Сейчас разберём каждый шаг:
```
1.  BeanDefinition зарегистрирован (сканирование / @Bean / .imports)
2.  BeanFactoryPostProcessor.postProcessBeanFactory()
3.  BeanPostProcessor регистрация
4.  getBean() вызван
5.  ─── для singleton ───
6.  MergedBeanDefinition создаётся (merge parent/child)
7.  InstantiationAwareBeanPostProcessor.postProcessBeforeInstantiation()
8.  InstantiationAwareBeanPostProcessor.determineCandidateConstructors()
9.  Конструктор вызван (или factory method)
10. MergedBeanDefinitionPostProcessor.postProcessMergedBeanDefinition()
11. InstantiationAwareBeanPostProcessor.postProcessAfterInstantiation()
12. InstantiationAwareBeanPostProcessor.postProcessProperties()
       └─ @Autowired / @Value / @Resource
13. Aware-интерфейсы: BeanNameAware, BeanClassLoaderAware, BeanFactoryAware
       └─ ApplicationContextAware, EnvironmentAware, ...
14. BeanPostProcessor.postProcessBeforeInitialization()
       └─ @PostConstruct (CommonAnnotationBeanPostProcessor)
15. InitializingBean.afterPropertiesSet()
16. @Bean(initMethod="...")
17. BeanPostProcessor.postProcessAfterInitialization()
       └─ AOP-прокси здесь (AbstractAutoProxyCreator)
18. ─── singleton cache ───
19. Bean готов к использованию
20. ─── при shutdown ───
21. DestructionAwareBeanPostProcessor.postProcessBeforeDestruction()
       └─ @PreDestroy
22. DisposableBean.destroy()
23. @Bean(destroyMethod="...")
```

### 6.3. BeanFactoryPostProcessor — работа с метаданными

Это первый этап. `BeanFactoryPostProcessor` (BFPP) работает с `BeanDefinition`-ами, до создания бинов.

```java
@FunctionalInterface
public interface BeanFactoryPostProcessor {
    void postProcessBeanFactory(ConfigurableListableBeanFactory beanFactory);
}
```

Вызывается в AbstractApplicationContext.invokeBeanFactoryPostProcessors(). Порядок:

1. `BeanDefinitionRegistryPostProcessor.postProcessBeanDefinitionRegistry()` — регистрируют новые `BeanDefinition`

2. `BeanFactoryPostProcessor.postProcessBeanFactory()` — модифицируют существующие

**Главный BFPP в Spring **— `ConfigurationClassPostProcessor`. Он:
- Обрабатывает `@Configuration`, `@Bean`, `@ComponentScan`, `@Import`, `@EnableAutoConfiguration`
- Регистрирует `BeanDefinition` для всех найденных бинов

Именно здесь происходят Части III–IV. `ConfigurationClassPostProcessor` имеет наивысший приоритет (`Ordered.HIGHEST_PRECEDENCE`), поэтому выполняется первым.

Свой BFPP:
```java
public class MyBeanFactoryPostProcessor implements BeanFactoryPostProcessor {
    @Override
    public void postProcessBeanFactory(ConfigurableListableBeanFactory beanFactory) {
        BeanDefinition bd = beanFactory.getBeanDefinition("myService");
        bd.setScope(BeanDefinition.SCOPE_PROTOTYPE);
    }
}
```

Регистрация: `@Component` или `@Bean`. Не инжектит другие бины — на этом этапе их ещё нет.

### 6.4. `BeanPostProcessor` — работа с бинами

`BeanPostProcessor` (BPP) вызывается для каждого бина при его создании.

```java
public interface BeanPostProcessor {
    default Object postProcessBeforeInitialization(Object bean, String beanName) {
        return bean;
    }
    default Object postProcessAfterInitialization(Object bean, String beanName) {
        return bean;
    }
}
```

Два метода — до и после инициализации. Здесь строится вся магия Spring: `@Autowired`, `@PostConstruct`, AOP-прокси, `@Transactional`.

Иерархия BPP
```
BeanPostProcessor
  ├─ InstantiationAwareBeanPostProcessor
  │     ├─ postProcessBeforeInstantiation()  ← до создания бина
  │     ├─ postProcessAfterInstantiation()   ← после создания, до инжекта
  │     └─ postProcessProperties()           ← инжект зависимостей
  │
  ├─ SmartInstantiationAwareBeanPostProcessor
  │     └─ determineCandidateConstructors()  ← выбор конструктора
  │
  ├─ MergedBeanDefinitionPostProcessor
  │     └─ postProcessMergedBeanDefinition() ← сбор metadata для инжекта
  │
  └─ DestructionAwareBeanPostProcessor
        └─ postProcessBeforeDestruction()    ← @PreDestroy
```

Встроенные BPP Spring

| BPP | Что делает | Порядок
|--|--|--
| `AutowiredAnnotationBeanPostProcessor` | `@Autowired`, `@Value`, `@Inject` | `Ordered.LOWEST_PRECEDENCE - 2`
| `CommonAnnotationBeanPostProcessor` | `@PostConstruct`, `@PreDestroy`, `@Resource` | `Ordered.LOWEST_PRECEDENCE - 3`
| `ConfigurationPropertiesBindingPostProcessor` | `@ConfigurationProperties` | `Ordered.HIGHEST_PRECEDENCE + 1`
| `ApplicationListenerDetector` | Регистрирует `ApplicationListener`-бины | `LOWEST_PRECEDENCE`
| `AbstractAutoProxyCreator` | AOP-прокси (`@Transactional`, `@Async`, `@Cacheable`) | зависит от реализации

**Важно**: порядок BPP критичен. Например, `@PostConstruct` должен выполниться до AOP-проксирования, иначе `@PostConstruct`-метод вызовется на прокси, а не на целевом объекте.

### 6.5. Пошаговый разбор создания бина

Разберём `doCreateBean()` из `AbstractAutowireCapableBeanFactory` — центральный метод создания бина.

Шаг 1: `createBean()` — вход
```java
@Override
protected Object createBean(String beanName, RootBeanDefinition mbd, Object[] args) {
    // 1. Дать BPP шанс вернуть прокси вместо реального бина
    Object bean = resolveBeforeInstantiation(beanName, mbdToUse);
    if (bean != null) {
        return bean;
    }
    
    // 2. Реальное создание
    return doCreateBean(beanName, mbdToUse, args);
}
```

Шаг 2: `resolveBeforeInstantiation()` — ранний прокси
```java
protected Object resolveBeforeInstantiation(String beanName, RootBeanDefinition mbd) {
    Object bean = null;
    if (!Boolean.FALSE.equals(mbd.beforeInstantiationResolved)) {
        if (!mbd.isSynthetic() && hasInstantiationAwareBeanPostProcessors()) {
            Class<?> targetType = determineTargetType(beanName, mbd);
            bean = applyBeanPostProcessorsBeforeInstantiation(targetType, beanName);
            if (bean != null) {
                bean = applyBeanPostProcessorsAfterInitialization(bean, beanName);
            }
        }
    }
    return bean;
}
```

`InstantiationAwareBeanPostProcessor.postProcessBeforeInstantiation()` может вернуть готовый объект, и тогда обычный цикл создания пропускается. Это используется, например, для `@Configuration`-прокси и для некоторых AOP-сценариев.

Шаг 3: `doCreateBean()` — основная логика
```java
protected Object doCreateBean(String beanName, RootBeanDefinition mbd, Object[] args) {
    // 1. Создание экземпляра (конструктор / factory method)
    BeanWrapper instanceWrapper = null;
    if (mbd.isSingleton()) {
        instanceWrapper = this.factoryBeanInstanceCache.remove(beanName);
    }
    if (instanceWrapper == null) {
        instanceWrapper = createBeanInstance(beanName, mbd, args);
    }
    Object bean = instanceWrapper.getWrappedInstance();
    
    // 2. MergedBeanDefinitionPostProcessor
    applyMergedBeanDefinitionPostProcessors(mbd, beanType, beanName);
    
    // 3. Ранняя регистрация (для circular dependencies)
    boolean earlySingletonExposure = (mbd.isSingleton() && this.allowCircularReferences 
        && isSingletonCurrentlyInCreation(beanName));
    if (earlySingletonExposure) {
        addSingletonFactory(beanName, () -> getEarlyBeanReference(beanName, mbd, bean));
    }
    
    // 4. Populate — инжект зависимостей
    Object exposedObject = bean;
    populateBean(beanName, mbd, instanceWrapper);
    
    // 5. Initialize — инициализация
    exposedObject = initializeBean(beanName, exposedObject, mbd);
    
    // 6. Проверка circular dependencies
    if (earlySingletonExposure) {
        Object earlySingletonReference = getSingleton(beanName, false);
        // ...
    }
    
    return exposedObject;
}
```

Шаг 4: `createBeanInstance()` — выбор конструктора
```java
protected BeanWrapper createBeanInstance(String beanName, RootBeanDefinition mbd, Object[] args) {
    // 1. Если есть factory method — используем его
    if (mbd.getFactoryMethodName() != null) {
        return instantiateUsingFactoryMethod(beanName, mbd, args);
    }
    
    // 2. Определение конструктора
    Constructor<?>[] ctors = determineConstructorsFromBeanPostProcessors(beanClass, beanName);
    if (ctors != null || mbd.getResolvedAutowireMode() == AUTOWIRE_CONSTRUCTOR 
            || mbd.hasConstructorArgumentValues() || !ObjectUtils.isEmpty(args)) {
        return autowireConstructor(beanName, mbd, ctors, args);
    }
    
    // 3. Дефолтный конструктор
    return instantiateBean(beanName, mbd);
}
```

`determineConstructorsFromBeanPostProcessors()` вызывает `SmartInstantiationAwareBeanPostProcessor.determineCandidateConstructors()`. Именно здесь `AutowiredAnnotationBeanPostProcessor` находит `@Autowired`-конструктор (или единственный конструктор).

Шаг 5: `populateBean()` — инжект зависимостей
```java
protected void populateBean(String beanName, RootBeanDefinition mbd, BeanWrapper bw) {
    // 1. postProcessAfterInstantiation — BPP может отменить инжект
    if (!mbd.isSynthetic() && hasInstantiationAwareBeanPostProcessors()) {
        for (InstantiationAwareBeanPostProcessor bp : getBeanPostProcessorCache().instantiationAware) {
            if (!bp.postProcessAfterInstantiation(bw.getWrappedInstance(), beanName)) {
                return;
            }
        }
    }
    
    // 2. postProcessProperties — инжект
    PropertyValues pvs = (mbd.hasPropertyValues() ? mbd.getPropertyValues() : null);
    if (pvs != null || hasInstantiationAwareBeanPostProcessors()) {
        for (InstantiationAwareBeanPostProcessor bp : getBeanPostProcessorCache().instantiationAware) {
            PropertyValues pvsToUse = bp.postProcessProperties(pvs, bw.getWrappedInstance(), beanName);
            // ...
        }
    }
    
    // 3. applyPropertyValues — установка значений (для XML-конфигов)
    if (pvs != null) {
        applyPropertyValues(beanName, mbd, bw, pvs);
    }
}
```

Именно здесь `AutowiredAnnotationBeanPostProcessor.postProcessProperties()` находит `@Autowired`-поля и сеттеры и инжектит их.

Шаг 6: `initializeBean()` — инициализация
```java
protected Object initializeBean(String beanName, Object bean, RootBeanDefinition mbd) {
    // 1. Aware-интерфейсы
    invokeAwareMethods(beanName, bean);   // BeanNameAware, BeanClassLoaderAware, BeanFactoryAware
    
    // 2. postProcessBeforeInitialization
    Object wrappedBean = bean;
    if (mbd == null || !mbd.isSynthetic()) {
        wrappedBean = applyBeanPostProcessorsBeforeInitialization(wrappedBean, beanName);
    }
    
    // 3. init-method
    try {
        invokeInitMethods(beanName, wrappedBean, mbd);
    } catch (Throwable ex) {
        throw new BeanCreationException(...);
    }
    
    // 4. postProcessAfterInitialization
    if (mbd == null || !mbd.isSynthetic()) {
        wrappedBean = applyBeanPostProcessorsAfterInitialization(wrappedBean, beanName);
    }
    
    return wrappedBean;
}
```

**6.1. `invokeAwareMethods()`**

```java
private void invokeAwareMethods(String beanName, Object bean) {
    if (bean instanceof Aware) {
        if (bean instanceof BeanNameAware bna) {
            bna.setBeanName(beanName);
        }
        if (bean instanceof BeanClassLoaderAware bcla) {
            ClassLoader bcl = getBeanClassLoader();
            if (bcl != null) bcla.setBeanClassLoader(bcl);
        }
        if (bean instanceof BeanFactoryAware bfa) {
            bfa.setBeanFactory(this);
        }
    }
}
```

Это только три Aware-интерфейса. Остальные (`ApplicationContextAware`, `EnvironmentAware`, `ResourceLoaderAware`, `ApplicationEventPublisherAware`, `MessageSourceAware`) обрабатываются `ApplicationContextAwareProcessor` — это `BeanPostProcessor`, зарегистрированный в контексте.

**6.2. `applyBeanPostProcessorsBeforeInitialization()`**

Здесь вызывается `CommonAnnotationBeanPostProcessor.postProcessBeforeInitialization()` — который вызывает `@PostConstruct`-методы.

Порядок для нескольких BPP — `@Order` / `Ordered`. `CommonAnnotationBeanPostProcessor` имеет `Ordered.LOWEST_PRECEDENCE - 3`, поэтому `@PostConstruct` вызывается позже других `beforeInitialization`, но раньше AOP-проксирования (которое идёт в `afterInitialization`).

**6.3. invokeInitMethods()**

```java
protected void invokeInitMethods(String beanName, Object bean, RootBeanDefinition mbd) {
    boolean isInitializingBean = (bean instanceof InitializingBean);
    if (isInitializingBean && (mbd == null || !mbd.hasAnyExternallyManagedInitMethod("afterPropertiesSet"))) {
        ((InitializingBean) bean).afterPropertiesSet();
    }
    if (mbd != null && bean.getClass() != NullBean.class) {
        String initMethodName = mbd.getInitMethodName();
        if (StringUtils.hasLength(initMethodName) 
                && !(isInitializingBean && "afterPropertiesSet".equals(initMethodName))
                && !mbd.hasAnyExternallyManagedInitMethod(initMethodName)) {
            invokeCustomInitMethod(beanName, bean, mbd);
        }
    }
}
```

Порядок: `InitializingBean.afterPropertiesSet()` → `@Bean(initMethod=...)`. `@PostConstruct` уже был вызван раньше, в `postProcessBeforeInitialization`.

Итоговый порядок инициализации:
1. `@PostConstruct`
2. `InitializingBean.afterPropertiesSet()`
3. `@Bean(initMethod = "...")`

**6.4. `applyBeanPostProcessorsAfterInitialization()`**

Здесь создаётся AOP-прокси. `AbstractAutoProxyCreator.postProcessAfterInitialization()` проверяет, есть ли у бина `@Transactional`, `@Async`, `@Cacheable` и т.д., и если да — оборачивает его в прокси (CGLIB или JDK dynamic proxy).

Ключевой момент: после `afterInitialization` в singleton cache кладётся прокси, а не оригинальный объект. Все дальнейшие `@Autowired` получат прокси.

### 6.6. Порядок вызовов — сводная таблица

Для бина `MyService` с полным набором callback'ов:

| # | Callback | Кто вызывает
|--|--|--
| 1 | Конструктор | `createBeanInstance()`
| 2 | `@Autowired` поля/сеттеры | `AutowiredAnnotationBeanPostProcessor`
| 3 | `BeanNameAware.setBeanName()` | `invokeAwareMethods()`
| 4 | `BeanFactoryAware.setBeanFactory()` | `invokeAwareMethods()`
| 5 | `ApplicationContextAware.setApplicationContext()` | `ApplicationContextAwareProcessor`
| 6 | `@PostConstruct` | `CommonAnnotationBeanPostProcessor`
| 7 | `InitializingBean.afterPropertiesSet()` | `invokeInitMethods()`
| 8 | `@Bean(initMethod)` | `invokeInitMethods()`
| 9 | AOP-прокси | `AbstractAutoProxyCreator`
| 10 | Бин готов | Singleton cache
| 11 | `@PreDestroy` | `CommonAnnotationBeanPostProcessor`
| 12 | `DisposableBean.destroy()` | `DisposableBeanAdapter`
| 13 | `@Bean(destroyMethod)` | `DisposableBeanAdapter`

### 6.7. Циклические зависимости

Проблема
```
A → B → A
```

Если `A` требует `B`, а `B` требует `A` — классический deadlock.

Решение Spring: three-level cache
```java
public class DefaultSingletonBeanRegistry {
    // Level 1: готовые singleton-бины
    private final Map<String, Object> singletonObjects = new ConcurrentHashMap<>(256);
    
    // Level 2: early singleton (до populate/init)
    private final Map<String, Object> earlySingletonObjects = new ConcurrentHashMap<>(16);
    
    // Level 3: singleton factories (лямбды для получения early reference)
    private final Map<String, ObjectFactory<?>> singletonFactories = new HashMap<>(16);
}
```

Алгоритм `getSingleton()`:

1. Проверить `singletonObjects` → есть — вернуть.
2. Проверить `earlySingletonObjects` → есть — вернуть.
3. Проверить `singletonFactories` → есть — вызвать фабрику, положить в `earlySingletonObjects`, вернуть.
4. Иначе — создать бин.

Пример:
- Создаём `A` → кладём `ObjectFactory` в level 3.
- Инжектим `B` в `A` → создаём `B`.
- `B` требует `A` → `getSingleton("A")` → level 3 → early reference → возвращает недоделанный `A`.
- `B` инициализируется полностью.
- Продолжаем инициализацию `A` → `A` готов.

Ограничения:
- Работает только для singleton.
- Не работает для constructor injection — оба бина не могут быть созданы через конструктор, потому что конструктор вызывается до регистрации в кэше. Решение — `@Lazy` на одном из параметров.
- Spring Boot 2.6+ запрещает circular references по умолчанию. Включается через `spring.main.allow-circular-references=true`.

### 6.8. `@Lazy` — отложенная инициализация

```java
@Service
public class A {
    @Autowired
    @Lazy
    private B b;
}
```

Spring инжектит прокси `B`, реальный бин создаётся при первом обращении. Это обходит circular dependencies.

`@Lazy` на `@Configuration`-классе — вся конфигурация lazy. `@Lazy` на `@Bean` — бин lazy.

### 6.9. `ObjectProvider` и `@Lookup`

`ObjectProvider`
```java
@Service
public class MyService {
    @Autowired
    private ObjectProvider<PrototypeBean> prototypeBeanProvider;
    
    public void doWork() {
        PrototypeBean bean = prototypeBeanProvider.getObject();  // новый каждый раз
    }
}
```

`ObjectProvider` — это ленивый резолвер. `getObject()` каждый раз запрашивает бин у контейнера. Работает и для prototype-scope.

`@Lookup`
```java
@Service
public abstract class MyService {
    @Lookup
    protected abstract PrototypeBean createPrototypeBean();
}
```

Spring генерирует подкласс (CGLIB), который переопределяет метод и запрашивает бин у контейнера. Устаревший способ, `ObjectProvider` предпочтительнее.

### 6.10. Уничтожение бинов

**`registerShutdownHook()`**
Вызывается в `SpringApplication.refreshContext()`:
```java
private void refreshContext(ConfigurableApplicationContext context) {
    if (this.registerShutdownHook) {
        try {
            context.registerShutdownHook();
        } catch (AccessControlException ex) { }
    }
    refresh(context);
}
```

`registerShutdownHook()` добавляет JVM shutdown hook, который вызовет `context.close()` при завершении JVM.

`close()` → `doClose()`

```java
@Override
protected void doClose() {
    // 1. LifecycleProcessor.onClose() — останавливает Lifecycle-бины
    if (this.lifecycleProcessor != null) {
        this.lifecycleProcessor.onClose();
    }
    
    // 2. Уничтожение бинов
    destroyBeans();
    
    // 3. Закрытие BeanFactory
    closeBeanFactory();
    
    // 4. onClose()
    onClose();
    
    // 5. Сброс listeners
    if (this.earlyApplicationListeners != null) {
        this.applicationListeners.clear();
        this.applicationListeners.addAll(this.earlyApplicationListeners);
    }
    
    // 6. active = false
    this.active.set(false);
}
```

`destroyBeans()` → `destroySingleton()`

```java
public void destroySingleton(String beanName) {
    removeSingleton(beanName);
    DisposableBean disposableBean = this.disposableBeans.remove(beanName);
    if (disposableBean != null) {
        disposableBean.destroy();
    }
}
```

`DisposableBeanAdapter.destroy()` вызывает по порядку:
1. `DestructionAwareBeanPostProcessor.postProcessBeforeDestruction()` → `@PreDestroy`
2. `DisposableBean.destroy()`
3. `@Bean(destroyMethod)` или `AutoCloseable.close()`

**Важно:** Spring автоматически вызывает `close()` для бинов, реализующих `AutoCloseable` / `Closeable`, если не указано иное. Отключается через `@Bean(destroyMethod = "")`.

### 6.11. `SmartLifecycle` и `Lifecycle`

`Lifecycle` — интерфейс для бинов с фазами start/stop:
```java
public interface Lifecycle {
    void start();
    void stop();
    boolean isRunning();
}
```

`SmartLifecycle` расширяет его:
```java
public interface SmartLifecycle extends Lifecycle, Phased {
    boolean isAutoStartup();
    void stop(Runnable callback);
    int getPhase();
}
```

`LifecycleProcessor` (по умолчанию `DefaultLifecycleProcessor`) управляет:

- `onRefresh()` — `start()` для всех `isAutoStartup()`, в порядке `getPhase()` (по возрастанию)
- `onClose()` — `stop()` для всех, в порядке убывания `getPhase()`

Пример: `WebServerStartStopLifecycle` — запускает/останавливает Tomcat.

### 6.12. Spring Boot 3 vs Spring Boot 4

| Аспект | Spring Boot 3 | Spring Boot 4
|--|--|--
| Bean lifecycle | Без изменений | Без изменений
| BeanPostProcessor | Те же интерфейсы | Те же интерфейсы
| Circular references | Запрещены по умолчанию (с 2.6) | Запрещены
| AOT | Есть, но opt-in | Более агрессивный, reflection-free по умолчанию
| `@ConstructorBinding` | Опционально | Legacy
| `BeanRegistrar` | Появился в Spring 7 | Основной способ регистрации для AOT

Для AOT (Boot 3.x / 4.x): Spring генерирует код, который регистрирует `BeanDefinition` программно, без reflection. `@PostConstruct`, `@Autowired` и другие callback'и превращаются в прямой вызов кода. Это работает быстрее и совместимо с GraalVM native image.

### 6.13. Диаграмма: полный lifecycle бина

```
BeanDefinition
    │
    ▼
BeanFactoryPostProcessor.postProcessBeanFactory()   ← работа с metadata
    │
    ▼
BeanPostProcessor регистрируются в контексте
    │
    ▼
getBean(beanName)
    │
    ├─ singletonObjects?  → да → return (готовый)
    │
    ▼
createBean()
    │
    ├─ resolveBeforeInstantiation()
    │     └─ InstantiationAwareBeanPostProcessor.postProcessBeforeInstantiation()
    │           └─ если вернул объект → return (shortcut)
    │
    ▼
doCreateBean()
    │
    ├─ 1. createBeanInstance()
    │       ├─ determineCandidateConstructors()    ← @Autowired конструктор
    │       └─ new Instance() / factory method
    │
    ├─ 2. applyMergedBeanDefinitionPostProcessors()
    │       └─ @Autowired metadata сканируется
    │
    ├─ 3. addSingletonFactory()                    ← ранний кэш (для circular)
    │
    ├─ 4. populateBean()
    │       ├─ postProcessAfterInstantiation()
    │       └─ postProcessProperties()             ← @Autowired / @Value
    │
    ├─ 5. initializeBean()
    │       ├─ invokeAwareMethods()                ← BeanNameAware, BeanFactoryAware
    │       ├─ postProcessBeforeInitialization()
    │       │     ├─ ApplicationContextAwareProcessor
    │       │     └─ CommonAnnotationBeanPostProcessor → @PostConstruct
    │       ├─ invokeInitMethods()
    │       │     ├─ InitializingBean.afterPropertiesSet()
    │       │     └─ @Bean(initMethod)
    │       └─ postProcessAfterInitialization()
    │             └─ AbstractAutoProxyCreator → AOP-прокси
    │
    ├─ 6. addSingleton()                           ← готовый бин в кэш
    │
    ▼
Бин готов
    │
    ▼ (при shutdown)
destroySingleton()
    ├─ DestructionAwareBeanPostProcessor.postProcessBeforeDestruction()
    │     └─ @PreDestroy
    ├─ DisposableBean.destroy()
    └─ @Bean(destroyMethod) / AutoCloseable.close()
```

### 6.14. Ключевые моменты

1. `BeanDefinition` — метаданные, не бин. Из него Spring узнаёт всё о будущем объекте.
2. `BeanFactoryPostProcessor` работает с metadata до создания бинов. Главный BFPP — `ConfigurationClassPostProcessor`.
3. `BeanPostProcessor` работает с бинами. Каждый бин проходит через цепочку BPP.
4. Порядок инициализации: `@PostConstruct` → `InitializingBean.afterPropertiesSet()` → `@Bean(initMethod)`.
5. AOP-прокси создаётся в `postProcessAfterInitialization()`. После этого в singleton cache лежит прокси, не оригинал.
6. Циклические зависимости разрешаются через three-level cache, но только для singleton и не для constructor injection.
7. `@Lazy` и `ObjectProvider` — обход circular dependencies.
8. Уничтожение: `@PreDestroy` → `DisposableBean.destroy()` → `@Bean(destroyMethod)` → `AutoCloseable.close()`.
9. `SmartLifecycle` — для бинов с фазами start/stop (например, web-сервер).
10. Boot 4 + AOT: lifecycle превращается в сгенерированный код, reflection минимизируется.
