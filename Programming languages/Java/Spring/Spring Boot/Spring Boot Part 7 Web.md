# Spring Boot

## Часть 7. Веб

Мы дошли до момента, когда `refresh()` вызывает `onRefresh()` у контекста. Для веб-приложений `onRefresh()` — это точка, где создаётся и запускается встроенный веб-сервер. Разберём, как Spring Boot выбирает сервер, создаёт `DispatcherServlet` и обрабатывает ошибки.

### 7.1. Servlet vs Reactive — два параллельных мира

Spring Boot поддерживает два стека:

| Стек | WebApplicationType | Сервер по умолчанию | Контекст
|--|--|--|--
| Spring MVC | `SERVLET` | Tomcat | `AnnotationConfigServletWebServerApplicationContext`
| Spring WebFlux | `REACTIVE` | Reactor Netty | `AnnotationConfigReactiveWebServerApplicationContext`

Выбор происходит в конструкторе `SpringApplication` через `deduceWebApplicationType()`. Если в classpath есть `DispatcherServlet` (Spring MVC) — `SERVLET`. Если есть `DispatcherHandler` (WebFlux) без MVC — `REACTIVE`.

**Важно**: если оба в classpath — побеждает `SERVLET`. Spring Boot отдаёт приоритет MVC как «более старому» стеку.

### 7.2. Автоконфигурация веб-сервера — `ServletWebServerFactoryAutoConfiguration`

Главный класс для Servlet-стека:
```java
@AutoConfiguration
@AutoConfigureOrder(Ordered.HIGHEST_PRECEDENCE)
@ConditionalOnClass(ServletRequest.class)
@ConditionalOnWebApplication(type = Type.SERVLET)
@EnableConfigurationProperties(ServerProperties.class)
@Import({ 
    ServletWebServerFactoryAutoConfiguration.BeanPostProcessorsRegistrar.class,
    ServletWebServerFactoryConfiguration.EmbeddedTomcat.class,
    ServletWebServerFactoryConfiguration.EmbeddedJetty.class,
    ServletWebServerFactoryConfiguration.EmbeddedUndertow.class 
})
public class ServletWebServerFactoryAutoConfiguration {
    
    @Bean
    public ServletWebServerFactoryCustomizer servletWebServerFactoryCustomizer(
            ServerProperties serverProperties) {
        return new ServletWebServerFactoryCustomizer(serverProperties);
    }
}
```

Ключевые моменты:

* `@AutoConfigureOrder(Ordered.HIGHEST_PRECEDENCE)` — сервер создаётся раньше большинства других автоконфигураций.

* `@Import` подтягивает три вложенных конфигурации — для Tomcat, Jetty и Undertow. Каждая активируется через `@ConditionalOnClass`.

**7.2.1. `EmbeddedTomcat` — выбор сервера**

```java
@Configuration(proxyBeanMethods = false)
@ConditionalOnClass({ Servlet.class, Tomcat.class, UpgradeProtocol.class })
@ConditionalOnMissingBean(value = ServletWebServerFactory.class, 
    search = SearchStrategy.CURRENT)
static class EmbeddedTomcat {
    
    @Bean
    TomcatServletWebServerFactory tomcatServletWebServerFactory(
            ObjectProvider<TomcatConnectorCustomizer> connectorCustomizers,
            ObjectProvider<TomcatContextCustomizer> contextCustomizers,
            ObjectProvider<TomcatProtocolHandlerCustomizer<?>> protocolHandlerCustomizers) {
        // ...
        return factory;
    }
}
```

Логика выбора:
1. Если в classpath есть `Tomcat.class` и нет своего `ServletWebServerFactory` — создаётся `TomcatServletWebServerFactory`.
2. Если добавить `spring-boot-starter-jetty` вместо Tomcat — `@ConditionalOnClass(Tomcat.class)` не сработает, активируется `EmbeddedJetty`.
3. Если объявить свой `ServletWebServerFactory` — `@ConditionalOnMissingBean` отступит, и твоя фабрика победит.

### 7.3. `ServletWebServerApplicationContext.onRefresh()` — создание сервера

Точка входа — `onRefresh()` в `ServletWebServerApplicationContext`:
```java
@Override
protected void onRefresh() {
    super.onRefresh();
    try {
        createWebServer();
    } catch (Throwable ex) {
        throw new ApplicationContextException("Unable to start web server", ex);
    }
}
```

`createWebServer()`:

```java
private void createWebServer() {
    WebServer webServer = this.webServer;
    ServletContext servletContext = getServletContext();
    if (webServer == null && servletContext == null) {
        StartupStep createWebServer = getApplicationStartup()
            .start("spring.boot.webserver.create");
        ServletWebServerFactory factory = getWebServerFactory();
        this.webServer = factory.getWebServer(getSelfInitializer());
        createWebServer.tag("factory", factory.getClass().toString());
        getBeanFactory().registerSingleton("webServerGracefulShutdown", 
            new WebServerGracefulShutdownLifecycle(this.webServer));
        getBeanFactory().registerSingleton("webServerStartStop", 
            new WebServerStartStopLifecycle(this, this.webServer));
    } else if (servletContext != null) {
        // ... embedded container
    }
}
```

Ключевые шаги:
1. `getWebServerFactory()` — достаёт `ServletWebServerFactory` из контекста.
2. `factory.getWebServer(getSelfInitializer())` — создаёт `WebServer` и передаёт callback для инициализации ServletContext.
3. Регистрируются два lifecycle-бина:
    - `webServerStartStop` — запускает/останавливает сервер
    - `webServerGracefulShutdown` — graceful shutdown

**7.3.1. `getSelfInitializer()` — мост к `ServletContextInitializer`**

```java
private org.springframework.boot.web.servlet.ServletContextInitializer getSelfInitializer() {
    return this::selfInitialize;
}

private void selfInitialize(ServletContext servletContext) throws ServletException {
    prepareWebApplicationContext(servletContext);
    registerApplicationScope(servletContext);
    WebApplicationContextUtils.registerEnvironmentBeans(getBeanFactory(), servletContext);
    for (ServletContextInitializer beans : getServletContextInitializerBeans()) {
        beans.onStartup(servletContext);
    }
}
```

`ServletContextInitializer` — SPI-интерфейс Spring Boot:
```java
@FunctionalInterface
public interface ServletContextInitializer {
    void onStartup(ServletContext servletContext) throws ServletException;
}
```

Все бины, реализующие `ServletContextInitializer`, вызываются при старте сервера. Именно так регистрируются `DispatcherServlet`, фильтры и слушатели.

**7.3.2. `ServletContextInitializerBeans` — сбор всех инициализаторов**

```java
public class ServletContextInitializerBeans 
        extends AbstractCollection<ServletContextInitializer> {
    
    public ServletContextInitializerBeans(ListableBeanFactory beanFactory, 
            Class<? extends ServletContextInitializer>... initializerTypes) {
        // 1. Собирает все бины ServletContextInitializer
        // 2. Собирает все ServletRegistrationBean, FilterRegistrationBean, 
        //    ServletListenerRegistrationBean
        // 3. Сортирует через @Order / Ordered
    }
}
```

Что сюда попадает:
- `DispatcherServletRegistrationBean` — регистрация `DispatcherServlet` (от `DispatcherServletAutoConfiguration`)
- Пользовательские `Filter` (Spring Security и т.д.)
- `@WebServlet`, `@WebFilter`, `@WebListener` (при `@ServletComponentScan`)

### 7.4. `WebServerStartStopLifecycle` — запуск Tomcat

После создания `WebServer` регистрируется lifecycle-бин:
```java
class WebServerStartStopLifecycle implements SmartLifecycle {
    private final ServletWebServerApplicationContext applicationContext;
    private final WebServer webServer;
    private volatile boolean running;
    
    @Override
    public void start() {
        this.webServer.start();
        this.running = true;
        this.applicationContext.publishEvent(
            new ServletWebServerInitializedEvent(this.webServer, this.applicationContext));
    }
    
    @Override
    public void stop() {
        this.webServer.stop();
        this.running = false;
    }
    
    @Override
    public int getPhase() {
        return Integer.MAX_VALUE - 1;  // запускается почти последним
    }
}
```

`SmartLifecycle` вызывается в `finishRefresh()` → `lifecycleProcessor.onRefresh()`. `getPhase()` = `Integer.MAX_VALUE - 1` означает, что сервер стартует после всех остальных lifecycle-бинов.

Порядок:
- `refresh()` → `finishRefresh()` → `lifecycleProcessor.onRefresh()`
- `WebServerStartStopLifecycle.start()` → `tomcat.start()`
- Публикуется `ServletWebServerInitializedEvent`

### 7.5. `DispatcherServletAutoConfiguration` — создание `DispatcherServlet`

```java
@AutoConfiguration(after = ServletWebServerFactoryAutoConfiguration.class)
@ConditionalOnClass(DispatcherServlet.class)
@AutoConfigureAfter(ServletWebServerFactoryAutoConfiguration.class)
public class DispatcherServletAutoConfiguration {
    
    @Configuration(proxyBeanMethods = false)
    @ConditionalOnClass(DispatcherServlet.class)
    @EnableConfigurationProperties(WebMvcProperties.class)
    protected static class DispatcherServletConfiguration {
        
        @Bean(name = DispatcherServletAutoConfiguration.DEFAULT_DISPATCHER_SERVLET_BEAN_NAME)
        public DispatcherServlet dispatcherServlet(WebMvcProperties webMvcProperties) {
            DispatcherServlet dispatcherServlet = new DispatcherServlet();
            dispatcherServlet.setDispatchOptionsRequest(
                webMvcProperties.isDispatchOptionsRequest());
            dispatcherServlet.setDispatchTraceRequest(
                webMvcProperties.isDispatchTraceRequest());
            // ...
            return dispatcherServlet;
        }
        
        @Bean
        public DispatcherServletRegistrationBean dispatcherServletRegistration(
                DispatcherServlet dispatcherServlet, WebMvcProperties webMvcProperties,
                ObjectProvider<MultipartConfigElement> multipartConfig) {
            DispatcherServletRegistrationBean registration = 
                new DispatcherServletRegistrationBean(dispatcherServlet, 
                    webMvcProperties.getServlet().getPath());
            registration.setName(DEFAULT_DISPATCHER_SERVLET_BEAN_NAME);
            registration.setLoadOnStartup(webMvcProperties.getServlet().getLoadOnStartup());
            multipartConfig.ifAvailable(registration::setMultipartConfig);
            return registration;
        }
    }
}
```

Два ключевых бина:
- `dispatcherServlet` — сам `DispatcherServlet`.
- `dispatcherServletRegistration` — `DispatcherServletRegistrationBean`, который реализует `ServletContextInitializer`. При старте сервера он регистрирует `DispatcherServlet` в `ServletContext`.

**Порядок:** `@AutoConfigureAfter(ServletWebServerFactoryAutoConfiguration.class)` гарантирует, что `DispatcherServlet` создаётся после того, как определён `ServletWebServerFactory`.

### 7.6. `WebMvcAutoConfiguration` — настройка Spring MVC

```java
@AutoConfiguration(after = { DispatcherServletAutoConfiguration.class, 
    TaskExecutionAutoConfiguration.class, ValidationAutoConfiguration.class })
@ConditionalOnWebApplication(type = Type.SERVLET)
@ConditionalOnClass({ Servlet.class, DispatcherServlet.class, WebMvcConfigurer.class })
@ConditionalOnMissingBean(WebMvcConfigurationSupport.class)
@AutoConfigureOrder(Ordered.HIGHEST_PRECEDENCE + 10)
@ImportRuntimeHints(WebResourcesRuntimeHints.class)
public class WebMvcAutoConfiguration {
    
    @Bean
    @ConditionalOnMissingBean
    public InternalResourceViewResolver defaultViewResolver() { ... }
    
    @Bean
    @ConditionalOnMissingBean
    public RequestMappingHandlerAdapter requestMappingHandlerAdapter(...) { ... }
    
    @Bean
    @ConditionalOnMissingBean
    public RequestMappingHandlerMapping requestMappingHandlerMapping(...) { ... }
    
    @Bean
    @ConditionalOnMissingBean
    public HttpMessageConverters messageConverters(...) { ... }
}
```

`@ConditionalOnMissingBean(WebMvcConfigurationSupport.class)` — ключевой момент. Если ты добавишь `@EnableWebMvc`, Spring Boot отступит, и вся автоконфигурация MVC отключится. Ты получишь «чистый» Spring MVC без Boot-настроек.

Что создаёт `WebMvcAutoConfiguration`:
- `RequestMappingHandlerMapping` — маппинг URL → методы контроллеров
- `RequestMappingHandlerAdapter` — вызов методов контроллеров
- `HttpMessageConverters` — Jackson, String, ByteArray и т.д.
- `InternalResourceViewResolver` — для JSP
- `MessageSource` — i18n
- `Validator` — JSR-380

### 7.7. `ErrorMvcAutoConfiguration` — обработка ошибок

`/error` — это не Spring MVC. Это Spring Boot-надстройка.

```java
@AutoConfiguration
@ConditionalOnWebApplication(type = Type.SERVLET)
@ConditionalOnClass({ Servlet.class, DispatcherServlet.class })
@AutoConfigureBefore(WebMvcAutoConfiguration.class)
@EnableConfigurationProperties({ ServerProperties.class, WebMvcProperties.class })
public class ErrorMvcAutoConfiguration {
    
    @Bean
    @ConditionalOnMissingBean(value = ErrorAttributes.class, search = SearchStrategy.CURRENT)
    public DefaultErrorAttributes errorAttributes() {
        return new DefaultErrorAttributes();
    }
    
    @Bean
    @ConditionalOnMissingBean(value = ErrorController.class, search = SearchStrategy.CURRENT)
    public BasicErrorController basicErrorController(ErrorAttributes errorAttributes,
            ObjectProvider<ErrorViewResolver> errorViewResolvers) {
        return new BasicErrorController(errorAttributes, this.serverProperties.getError(),
            errorViewResolvers.orderedStream().toList());
    }
    
    @Bean
    public ErrorPageCustomizer errorPageCustomizer(DispatcherServletPath dispatcherServletPath) {
        return new ErrorPageCustomizer(this.serverProperties, dispatcherServletPath);
    }
}
```

**7.7.1. `BasicErrorController` — обработчик `/error`**

```java
@Controller
@RequestMapping("${server.error.path:${error.path:/error}}")
public class BasicErrorController extends AbstractErrorController {
    
    @RequestMapping(produces = MediaType.TEXT_HTML_VALUE)
    public ModelAndView errorHtml(HttpServletRequest request, HttpServletResponse response) {
        HttpStatus status = getStatus(request);
        Map<String, Object> model = Collections.unmodifiableMap(
            getErrorAttributes(request, getErrorAttributeOptions(request, MediaType.TEXT_HTML)));
        response.setStatus(status.value());
        ModelAndView modelAndView = resolveErrorView(request, response, status, model);
        return (modelAndView != null) ? modelAndView : new ModelAndView("error", model);
    }
    
    @RequestMapping
    public ResponseEntity<Map<String, Object>> error(HttpServletRequest request) {
        HttpStatus status = getStatus(request);
        if (status == HttpStatus.NO_CONTENT) {
            return new ResponseEntity<>(status);
        }
        Map<String, Object> body = getErrorAttributes(request, 
            getErrorAttributeOptions(request, MediaType.ALL));
        return new ResponseEntity<>(body, status);
    }
}
```

Что делает `BasicErrorController`:
- Обрабатывает `/error`
- Если Accept: `text/html` → рендерит HTML-страницу (Whitelabel Error Page)
- Если Accept: `application/json` → возвращает JSON

**7.7.2. `DefaultErrorAttributes` — модель ошибки**

```java
public class DefaultErrorAttributes implements ErrorAttributes {
    
    @Override
    public Map<String, Object> getErrorAttributes(WebRequest webRequest, 
            ErrorAttributeOptions options) {
        Map<String, Object> errorAttributes = new LinkedHashMap<>();
        errorAttributes.put("timestamp", new Date());
        addStatus(errorAttributes, webRequest);
        addErrorDetails(errorAttributes, webRequest, options);
        addPath(errorAttributes, webRequest, options);
        return errorAttributes;
    }
}
```

Стандартный JSON:
```json
{
    "timestamp": "2026-09-23T12:00:00.000+00:00",
    "status": 500,
    "error": "Internal Server Error",
    "message": "...",
    "path": "/api/order"
}
```

**7.7.3. `ErrorPageCustomizer` — регистрация error page**

```java
private static class ErrorPageCustomizer implements ErrorPageRegistrar, Ordered {
    
    @Override
    public void registerErrorPages(ErrorPageRegistry errorPageRegistry) {
        ErrorPage errorPage = new ErrorPage(
            this.dispatcherServletPath.getRelativePath(this.properties.getError().getPath()));
        errorPageRegistry.addErrorPages(errorPage);
    }
}
```

Регистрирует `/error` как страницу ошибки в Tomcat. Когда Tomcat ловит необработанное исключение (или 404/405), он форвардит на `/error`, где `BasicErrorController` формирует ответ.

Цепочка:
```
Controller → DispatcherServlet → HandlerExceptionResolver → не обработано
    → Tomcat → forward на /error → BasicErrorController
```

### 7.8. Настройка сервера — `ServerProperties`

Все `server.*` настройки читаются из `ServerProperties`:
```java
@ConfigurationProperties(prefix = "server", ignoreUnknownFields = true)
public class ServerProperties {
    private Integer port;
    private InetAddress address;
    private String contextPath;
    private final Servlet servlet = new Servlet();
    private final Tomcat tomcat = new Tomcat();
    private final Jetty jetty = new Jetty();
    // ...
}
```

Ключевые настройки:

| Свойство | Что делает
|--|--
| `server.port` | порт (по умолчанию 8080)
| `server.address` | адрес прослушивания
| `server.servlet.context-path` | корневой путь приложения
| `server.servlet.session.timeout` | таймаут сессии
| `server.tomcat.threads.max` | max потоков Tomcat
| `server.tomcat.threads.min-spare` | min spare потоков
| `server.shutdown=graceful` | graceful shutdown
| `spring.lifecycle.timeout-per-shutdown-phase` | таймаут graceful shutdown

**7.8.1. `ServletWebServerFactoryCustomizer` — применение свойств**
```java
public class ServletWebServerFactoryCustomizer 
        implements WebServerFactoryCustomizer<ConfigurableServletWebServerFactory>, Ordered {
    
    private final ServerProperties serverProperties;
    
    @Override
    public void customize(ConfigurableServletWebServerFactory factory) {
        PropertyMapper map = PropertyMapper.get().alwaysApplyingWhenNonNull();
        map.from(this.serverProperties::getPort).to(factory::setPort);
        map.from(this.serverProperties::getAddress).to(factory::setAddress);
        map.from(this.serverProperties.getServlet()::getContextPath).to(factory::setContextPath);
        map.from(this.serverProperties.getServlet()::getSession).to(factory::setSession);
        map.from(this.serverProperties::getSsl).to(factory::setSsl);
        map.from(this.serverProperties::getCompression).to(factory::setCompression);
        map.from(this.serverProperties::getHttp2).to(factory::setHttp2);
        map.from(this.serverProperties::getError).to(factory::setErrorPages);
    }
    
    @Override
    public int getOrder() {
        return 0;  // до пользовательских customizer'ов
    }
}
```

`@Order(0)` — автоконфигурационные customizer'ы применяются до пользовательских. Если хочешь переопределить настройку, объяви свой `WebServerFactoryCustomizer<ConfigurableServletWebServerFactory>` — он выполнится после и перезапишет значение.

**7.8.2. Свой `WebServerFactoryCustomizer`**
```java
@Component
public class MyTomcatCustomizer 
        implements WebServerFactoryCustomizer<TomcatServletWebServerFactory> {
    
    @Override
    public void customize(TomcatServletWebServerFactory factory) {
        factory.addConnectorCustomizers(connector -> {
            connector.setProperty("maxKeepAliveRequests", "100");
        });
    }
}
```

### 7.9. Reactive-стек — WebFlux и Netty

**7.9.1. `ReactiveWebServerFactoryAutoConfiguration`**

```java
@AutoConfiguration
@AutoConfigureOrder(Ordered.HIGHEST_PRECEDENCE)
@ConditionalOnClass(ReactiveHttpInputMessage.class)
@ConditionalOnWebApplication(type = Type.REACTIVE)
@EnableConfigurationProperties(ServerProperties.class)
@Import({ 
    ReactiveWebServerFactoryAutoConfiguration.BeanPostProcessorsRegistrar.class,
    ReactiveWebServerFactoryConfiguration.EmbeddedTomcat.class,
    ReactiveWebServerFactoryConfiguration.EmbeddedJetty.class,
    ReactiveWebServerFactoryConfiguration.EmbeddedUndertow.class,
    ReactiveWebServerFactoryConfiguration.EmbeddedNetty.class 
})
public class ReactiveWebServerFactoryAutoConfiguration {
    
    @Bean
    public ReactiveWebServerFactoryCustomizer reactiveWebServerFactoryCustomizer(
            ServerProperties serverProperties) {
        return new ReactiveWebServerFactoryCustomizer(serverProperties);
    }
}
```

`EmbeddedNetty` активируется через `@ConditionalOnClass({ HttpServer.class, ... })` и создаёт `NettyReactiveWebServerFactory`.

**7.9.2. `NettyReactiveWebServerFactory`**

```java
public class NettyReactiveWebServerFactory extends AbstractReactiveWebServerFactory {
    
    @Override
    public WebServer getWebServer(HttpHandler httpHandler) {
        HttpServer httpServer = createHttpServer();
        // ... настройка
        return new NettyWebServer(httpServer, handler, 
            getRouteProvider(), this::getServerShutdownTimeout);
    }
}
```

`HttpHandler` — это `WebHttpHandlerBuilder`, который оборачивает `DispatcherHandler` WebFlux. Нет Servlet API — Netty работает напрямую с `HttpHandler`.

**7.9.3. Ключевые отличия от Servlet**

| Аспект | Servlet (Tomcat) | Reactive (Netty)
|--|--|--
| Модель | Thread-per-request | Event loop
| API | `ServletContext`, `HttpServletRequest` | `ServerHttpRequest`, `ServerHttpResponse`
| Обработчик | `DispatcherServlet` | `DispatcherHandler`
| Ошибки | `/error` + `BasicErrorController` | `DefaultErrorWebExceptionHandler`
| Lifecycle | `WebServerStartStopLifecycle` | `NettyWebServer`

### 7.10. Spring Boot 3 vs Spring Boot 4 — ключевые отличия

**7.10.1. Модуляризация**

Самое большое изменение. В Boot 3 всё лежит в spring-boot-autoconfigure. В Boot 4 — 47 модулей:

| Boot 3 | Boot 4
|--|--
| `spring-boot-autoconfigure` (всё) | `spring-boot-webmvc` (MVC)
| | `spring-boot-webflux` (WebFlux)
| | `spring-boot-tomcat` (Tomcat)
| | `spring-boot-jetty` (Jetty)
| | `spring-boot-jackson` (Jackson)
| `spring-boot-starter-web` | `spring-boot-starter-webmvc` (переименован)
| `spring-boot-starter-webflux` | `spring-boot-starter-webflux` (остался)

Пакеты автоконфигураций тоже переехали:
- `org.springframework.boot.autoconfigure.web.servlet` → `org.springframework.boot.webmvc.autoconfigure`
- `org.springframework.boot.autoconfigure.web.reactive` → `org.springframework.boot.webflux.autoconfigure`

**7.10.2. Удалён Undertow**

Spring Boot 4 требует Servlet 6.1, с которым Undertow несовместим. Undertow полностью удалён:

> «Spring Boot 4.0 requires a Servlet 6.1 baseline, with which Undertow is not yet compatible. As a result, Undertow support is dropped, including the Undertow starter and the ability to use Undertow as an embedded server.»

Остались: Tomcat, Jetty, Netty (для WebFlux).

**7.10.3. Перемещение классов**

Классы встроенных серверов переехали:

| Boot 3 | Boot 4
|--|--
| `org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory` | `org.springframework.boot.tomcat.TomcatServletWebServerFactory`
| `org.springframework.boot.web.embedded.netty.NettyReactiveWebServerFactory` | `org.springframework.boot.netty.NettyReactiveWebServerFactory`

**7.10.4. Что не изменилось**

- Механизм `ServletWebServerFactory` — та же абстракция.
- `ServletContextInitializer` — тот же SPI.
- `ErrorMvcAutoConfiguration` / `BasicErrorController` — та же логика.
- `WebMvcAutoConfiguration` — те же бины, только пакет другой.
- `WebServerFactoryCustomizer` — тот же интерфейс.

### 7.11. Диаграмма: создание веб-сервера (Servlet)

```
refresh()
  └─ AbstractApplicationContext.refresh()
        ├─ ...
        └─ onRefresh()                          ← переопределён в ServletWebServerApplicationContext
              │
              └─ createWebServer()
                    ├─ getWebServerFactory()
                    │     └─ TomcatServletWebServerFactory (от ServletWebServerFactoryAutoConfiguration)
                    │
                    ├─ factory.getWebServer(getSelfInitializer())
                    │     └─ TomcatServletWebServerFactory.getWebServer()
                    │           ├─ new Tomcat()
                    │           ├─ context = tomcat.addContext()
                    │           ├─ initializer.onStartup(servletContext)   ← callback
                    │           │     └─ ServletContextInitializerBeans
                    │           │           ├─ DispatcherServletRegistrationBean
                    │           │           ├─ FilterRegistrationBean
                    │           │           └─ ServletListenerRegistrationBean
                    │           └─ return new TomcatWebServer(tomcat)
                    │
                    ├─ registerSingleton("webServerStartStop", 
                    │     new WebServerStartStopLifecycle(...))
                    │
                    └─ registerSingleton("webServerGracefulShutdown", 
                          new WebServerGracefulShutdownLifecycle(...))
        │
        └─ finishRefresh()
              └─ lifecycleProcessor.onRefresh()
                    └─ WebServerStartStopLifecycle.start()    ← phase = MAX-1
                          ├─ tomcat.start()
                          ├─ publishEvent(ServletWebServerInitializedEvent)
                          └─ running = true
```

### 7.12. Ключевые моменты

1. `ServletWebServerFactoryAutoConfiguration` — выбирает Tomcat/Jetty/Undertow по classpath. `@AutoConfigureOrder(HIGHEST_PRECEDENCE)`.
2. `onRefresh()` → `createWebServer()` — точка создания сервера. Вызывается в `refresh()` до `finishRefresh()`.
3. `ServletContextInitializer` — SPI для регистрации сервлетов, фильтров, слушателей. `DispatcherServletRegistrationBean` — одна из реализаций.
4. `WebServerStartStopLifecycle` — `SmartLifecycle` с `phase = MAX-1`, запускает Tomcat после всех остальных бинов.
5. `DispatcherServletAutoConfiguration` — создаёт `DispatcherServlet` и его регистрацию. Порядок: после `ServletWebServerFactoryAutoConfiguration`.
6. `WebMvcAutoConfiguration` — настраивает MVC: `HandlerMapping`, `HandlerAdapter`, `HttpMessageConverters`. Отключается через `@EnableWebMvc`.
7. `ErrorMvcAutoConfiguration` — `/error` + `BasicErrorController`. Это не MVC, а Boot-надстройка.
8. `ServletWebServerFactoryCustomizer` — применяет `server.*` properties. `@Order(0)`, пользовательские customizer'ы выполняются после.
9. Reactive-стек — `NettyReactiveWebServerFactory` + `DispatcherHandler`. Другая модель, другие API.
10. Boot 4: модуляризация (47 модулей), `spring-boot-starter-webmvc`, Undertow удалён, пакеты переехали. Механизм — тот же.
