# Spring Security

## Глава 1. Архитектура и базовые концепции

### 1. Аутентификация (Authentication) vs Авторизация (Authorization)

Это база, но в Spring Security эти термины имеют строгие границы.
- **Аутентификация (AuthN — "Кто ты?"):** Процесс проверки личности. Ты вводишь логин и пароль, система проверяет их. Если всё верно — ты аутентифицирован.
- **Авторизация (AuthZ — "Что тебе можно?"):** Процесс проверки прав после того, как система поняла, кто ты. Например, аутентифицированный пользователь Иван может только читать статьи, а администратор Петр — удалять их.

**Важно:** Аутентификация всегда предшествует авторизации. Нельзя проверить права, если не понятно, кто их запрашивает.

### 2. Сердце Spring Security: Цепочка фильтров (Security Filter Chain)

Spring Security не перехватывает запросы с помощью магии или прокси. Он использует стандартный механизм Servlet API — **Фильтры (Filters)**.

Когда HTTP-запрос приходит в твое приложение, он не летит сразу в Controller. Сначала он проходит через цепочку фильтров Spring Security.

- **Как это работает:** Запрос передается от фильтра к фильтру по порядку.
- **Примеры фильтров:**
    * `CsrfFilter` — проверяет CSRF-токены.
    * `UsernamePasswordAuthenticationFilter` — перехватывает POST-запросы на /login, достает логин/пароль и пытается аутентифицировать пользователя.
    * `BearerTokenAuthenticationFilter` — ищет JWT-токен в заголовке Authorization.
    * `AuthorizationFilter` — самый последний фильтр. Он смотрит, аутентифицирован ли пользователь и есть ли у него права (Авторизация).
- **Главное правило:** Если какой-то фильтр в цепочке решает, что запрос невалиден (например, пароль неверный), он обрывает цепочку и сразу возвращает ошибку (например, 401 Unauthorized). До Controller запрос просто не дойдет.
- **Кастомизация:** Ты можешь добавлять свои собственные фильтры в цепочку (до, после или вместо стандартных), если тебе нужна специфическая логика.

### 3. Ключевые объекты (Словарь Spring Security)

Чтобы понимать код и логи, нужно знать 4 главных интерфейса/класса.

**А. `GrantedAuthority` (Полномочия / Права)**

Отвечает на вопрос "Что можно делать?".
Это строки, обозначающие права. Обычно это роли (например, `ROLE_ADMIN`, `ROLE_USER`) или более мелкие привилегии (`READ_PRIVILEGE`, `WRITE_PRIVILEGE`).

**Б. `Principal` (Субъект / Идентичность)**

Отвечает на вопрос "Кто это?".
Чаще всего это просто имя пользователя (строка `username`), но технически это может быть любой объект, представляющий текущего пользователя (например, сущность `User` из базы данных).

**В. `Authentication` (Аутентификация - объект)**

Это главный объект в процессе аутентификации. Он объединяет в себе всё:
1. `Principal` (кто).
2. `Credentials` (доказательство: пароль или токен. *После успешной аутентификации обычно стирается из соображений безопасности*).
3. `Authorities` (список `GrantedAuthority` — какие права есть).
4. `isAuthenticated` (флаг: true, если пользователь подтвержден, false, если это просто анонимный запрос).

**Г. `SecurityContext` (Контекст безопасности)**

Это "коробка", в которой хранится объект `Authentication`.
- Контекст привязан к текущему потоку выполнения (Thread) с помощью `ThreadLocal`.
- Это значит, что в любом месте твоего кода (в Controller, в Service) ты можешь достать текущего пользователя, не передавая его через параметры методов.
- Как достать: `SecurityContextHolder.getContext().getAuthentication()`.

### 4. Краткий пайплайн: Как это работает вместе (Lifecycle)

Давай соберем всё воедино на примере запроса с JWT-токеном:
1. Запрос приходит с заголовком `Authorization: Bearer <token>`.
2. Фильтр (`BearerTokenAuthenticationFilter`) перехватывает запрос, читает токен.
3. Фильтр создает неаутентифицированный объект `Authentication` (где токен — это credentials, а флаг `isAuthenticated = false`).
4. Этот объект передается в `AuthenticationManager`, который проверяет токен (валиден ли он, не истек ли).
5. Если токен верный, создается новый, аутентифицированный `Authentication` (с флагом `true`, заполненными `Principal` и `Authorities`).
6. Этот объект кладется в `SecurityContext`.
7. Запрос идет дальше по фильтрам, доходит до `AuthorizationFilter`, который проверяет права.
8. Запрос попадает в Controller. Ты можешь достать пользователя из `SecurityContext`.


## Глава 2. Аутентификация (Классическая)

### 1. `UserDetails` и `UserDetailsService` — "Где взять пользователя?"

Spring Security не знает, где хранятся твои пользователи (в БД, в LDAP, в памяти, в файле). Ему нужен контракт — способ получить данные о пользователе.

**А. `UserDetails` (Интерфейс)**
Это объект, описывающий пользователя с точки зрения Spring Security. Он должен реализовать несколько методов:
- `getUsername()` — логин.
- `getPassword()` — хешированный пароль (да, хранится именно хеш!).
- `getAuthorities()` — список прав (`GrantedAuthority`).
- `isAccountNonExpired()`, `isAccountNonLocked()`, `isCredentialsNonExpired()`, `isEnabled()` — флаги состояния аккаунта (заблокирован, истек ли пароль и т.д.).

**Лайфхак:** В своих проектах ты обычно не пишешь класс с нуля, а наследуешься от готового `org.springframework.security.core.userdetails.User` или делаешь свою сущность (например, JPA-Entity), реализующую этот интерфейс.

**Б. `UserDetailsService` (Интерфейс)**

Это "поставщик" пользователей. У него всего один метод:
```java
UserDetails loadUserByUsername(String username) throws UsernameNotFoundException;
```

Ты пишешь свою реализацию, например:
```java
@Service
public class MyUserDetailsService implements UserDetailsService {
    @Autowired
    private UserRepository userRepository;

    @Override
    public UserDetails loadUserByUsername(String username) {
        User user = userRepository.findByUsername(username)
            .orElseThrow(() -> new UsernameNotFoundException("User not found"));
        // user должен реализовывать UserDetails
        return user;
    }
}
```

**Главное правило:** Если пользователь не найден — всегда бросай `UsernameNotFoundException`. Никогда не возвращай `null`.

### 2. PasswordEncoder — "Как хранить пароли?"

Никогда не храни пароли в открытом виде. Это правило №1.

Spring Security предоставляет интерфейс PasswordEncoder с двумя методами:
- `String encode(CharSequence rawPassword)` — превращает сырой пароль в хеш.
- `boolean matches(CharSequence rawPassword, String encodedPassword)` — сравнивает сырой пароль с хешем.

**Самый популярный вариант: `BCryptPasswordEncoder`**
```java
@Bean
public PasswordEncoder passwordEncoder() {
    return new BCryptPasswordEncoder();
}
```

- BCrypt автоматически добавляет "соль" (salt) к каждому паролю.
- Результат выглядит так: `$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy`
- Одна и та же строка при каждом хешировании даст разный результат (из-за соли), но `matches()` все равно сработает корректно.

### 3. `AuthenticationManager` и `AuthenticationProvider` — "Кто проверяет?"

Когда `UsernamePasswordAuthenticationFilter` получает логин и пароль, он не проверяет их сам. Он делегирует это дальше по цепочке.

**А. `AuthenticationManager` (Интерфейс)**

Это "менеджер" аутентификации. У него один метод:
```java
Authentication authenticate(Authentication authentication) throws AuthenticationException;
```

На практике почти всегда используется реализация `ProviderManager`. Её задача — передать запрос нужному провайдеру.

**Б. `AuthenticationProvider` (Интерфейс)**

Это конкретный "исполнитель". Он знает, как аутентифицировать пользователя определенным способом.
```java
public interface AuthenticationProvider {
    Authentication authenticate(Authentication authentication) throws AuthenticationException;
    boolean supports(Class<?> authentication); // Поддерживаю ли я этот тип?
}
```

**Как они работают вместе (стандартный сценарий):**
1. Фильтр создает `UsernamePasswordAuthenticationToken` (это подтип `Authentication` с логином и паролем).
2. Передает его в `AuthenticationManager`.
3. `ProviderManager` перебирает список `AuthenticationProvider`-ов и спрашивает каждого: `supports()`?
4. Находит `DaoAuthenticationProvider` (стандартный, для работы с БД).
5. `DaoAuthenticationProvider` делает две вещи:
    - Вызывает твой `UserDetailsService.loadUserByUsername()` — получает `UserDetails`.
    - С помощью `PasswordEncoder.matches()` сравнивает введенный пароль с тем, что в `UserDetails`.
6. Если всё ок — создает новый `Authentication` объект, уже с флагом `isAuthenticated = true`, и возвращает его.
7. Если ошибка — бросает исключение (`BadCredentialsException`, `DisabledException` и т.д.).

**Кастомный `AuthenticationProvider`**

Если тебе нужна нестандартная логика (например, аутентификация по SMS-коду, или через внешний API), ты пишешь свой провайдер:
```java
@Component
public class SmsAuthenticationProvider implements AuthenticationProvider {
    @Override
    public Authentication authenticate(Authentication auth) {
        String phone = auth.getName();
        String code = auth.getCredentials().toString();
        // своя логика проверки кода...
        return authenticatedToken;
    }
    @Override
    public boolean supports(Class<?> authentication) {
        return SmsAuthenticationToken.class.isAssignableFrom(authentication);
    }
}
```

### 4. Краткая схема всего процесса (Login с формой)

1. Пользователь отправляет POST на `/login` с `username` и `password`.
2. `UsernamePasswordAuthenticationFilter` перехватывает запрос.
3. Создает неаутентифицированный `UsernamePasswordAuthenticationToken`.
4. Передает в `AuthenticationManager`.
5. `DaoAuthenticationProvider` вызывает твой `UserDetailsService` → получает Us`erDetails.
6. Сравнивает пароли через `PasswordEncoder.matches()`.
7. Если ок → создает аутентифицированный `Authentication` → кладет в `SecurityContext`.
8. Возвращает успешный ответ (редирект или JSON).
9. Если не ок → бросает исключение, которое обрабатывает `AuthenticationFailureHandler` (обычно редирект на `/login?error`).

## Глава 3. Авторизация и контроль доступа

### 1. Роли vs Привилегии (Authorities)

В Spring Security все права — это строки, которые реализуют интерфейс `GrantedAuthority`. Но исторически сложилось два подхода к их именованию:

- Роли (Roles): Широкие категории. Обычно начинаются с префикса `ROLE_`.
    * Примеры: `ROLE_ADMIN`, `ROLE_USER`, `ROLE_MODERATOR`.
    * Используются, когда права групповые: "админ может всё", "пользователь — только читать".
- Привилегии (Authorities/Privileges): Более мелкие, точные права.
    * Примеры: `READ_PRIVILEGE`, `WRITE_PRIVILEGE`, `DELETE_ARTICLE`.
    * Используются в сложных системах, где нужны гибкие настройки.

**Важный момент:** Spring Security не делает между ними технической разницы. И то, и другое — просто строки в списке `authorities`. Префикс `ROLE_` — это просто соглашение, но методы `hasRole()` и `hasAuthority()` работают по-разному (об этом ниже).

### 2. Защита URL (HTTP Security)

Это самый распространенный способ. Ты говоришь Spring Security: "Этот URL доступен только админам, а этот — всем".

Настраивается в классе конфигурации через `SecurityFilterChain`:
```java
@Bean
public SecurityFilterChain filterChain(HttpSecurity http) throws Exception {
    http
        .authorizeHttpRequests(auth -> auth
            // Публичные эндпоинты
            .requestMatchers("/login", "/register", "/public/**").permitAll()
            
            // Админ-зона: только для пользователей с ролью ADMIN
            .requestMatchers("/admin/**").hasRole("ADMIN")
            
            // Эндпоинт для модераторов ИЛИ админов
            .requestMatchers("/moderator/**").hasAnyRole("MODERATOR", "ADMIN")
            
            // Эндпоинт для конкретной привилегии
            .requestMatchers("/api/articles/delete").hasAuthority("DELETE_ARTICLE")
            
            // Любой аутентифицированный пользователь
            .requestMatchers("/api/**").authenticated()
            
            // Всё остальное — запретить
            .anyRequest().denyAll()
        )
        // ... другие настройки (csrf, cors, formLogin и т.д.)
        ;
    return http.build();
}
```

Ключевые методы:
- `.permitAll()` — доступно всем (анонимусам тоже).
- `.authenticated()` — только аутентифицированным.
- `.hasRole("ADMIN")` — нужна роль `ROLE_ADMIN` (префикс добавляется автоматически!).
- `.hasAuthority("DELETE_ARTICLE")` — нужна привилегия `DELETE_ARTICLE` (без префикса).
- `.hasAnyRole(...)` — любая из перечисленных ролей.
- `.access(...)` — для сложных условий через SpEL (см. ниже).

**Внимание:** Правила проверяются сверху вниз. Первое совпадение — победа. Поэтому более специфичные правила (`/admin/**`) должны идти раньше общих (`/api/**`).

### 3. Защита методов (Method Security)

Иногда нужно защитить не URL, а конкретный метод в Service. Например, метод `deleteUser()` должен быть доступен только админу, независимо от того, с какого URL его вызвали.

**Шаг 1: Включить метод-секьюрити**

В Spring Security 6+ это делается одной аннотацией на конфигурации:
```java
@Configuration
@EnableMethodSecurity // <-- волшебная строчка
public class SecurityConfig { ... }
```

**Шаг 2: Использовать аннотации на методах**

А. `@PreAuthorize` — самая мощная
Проверяет условие до вызова метода. Поддерживает SpEL (Spring Expression Language):
```java
@Service
public class ArticleService {
    
    // Простая проверка роли
    @PreAuthorize("hasRole('ADMIN')")
    public void deleteAnyArticle(Long id) { ... }
    
    // Проверка привилегии
    @PreAuthorize("hasAuthority('WRITE_ARTICLE')")
    public Article createArticle(Article article) { ... }
    
    // Сложная логика: пользователь может редактировать ТОЛЬКО свои статьи
    @PreAuthorize("#article.author.id == authentication.principal.id")
    public void editArticle(Article article) { ... }
    
    // Комбинированные условия
    @PreAuthorize("hasRole('ADMIN') or #article.author.id == authentication.principal.id")
    public void deleteArticle(Article article) { ... }
}
```

Б. `@PostAuthorize` — проверка ПОСЛЕ выполнения

Полезна, когда нужно проверить что-то, что известно только после выполнения метода (например, вернуть данные, если пользователь имеет право их видеть):
```java
@PostAuthorize("returnObject.author.id == authentication.principal.id")
public Article getArticle(Long id) { ... }
```

В. `@Secured` — старая, простая аннотация

Не поддерживает SpEL, только роли:
```java
@Secured("ROLE_ADMIN")
public void adminOnly() { ... }
```

Г. `@RolesAllowed` — из стандарта JSR-250

Аналог `@Secured`, но без префикса `ROLE_`:
```java
@RolesAllowed("ADMIN") // эквивалентно ROLE_ADMIN
public void adminOnly() { ... }
```

### 4. Доступ к текущему пользователю в Controller

Часто нужно получить данные текущего пользователя прямо в контроллере. Есть 3 способа:

**Способ 1: Через `@AuthenticationPrincipal` (рекомендуемый)**

Если твой `UserDetails` — это сущность `User` из БД:
```java
@GetMapping("/profile")
public String profile(@AuthenticationPrincipal User user) {
    return "Привет, " + user.getUsername();
}
```

**Способ 2: Через `Authentication` напрямую**
```java
@GetMapping("/info")
public String info(Authentication authentication) {
    String username = authentication.getName();
    Collection<? extends GrantedAuthority> roles = authentication.getAuthorities();
    return "User: " + username + ", Roles: " + roles;
}
```

**Способ 3: Через `SecurityContextHolder` (в любом месте кода)**
```java
public void someBusinessLogic() {
    Authentication auth = SecurityContextHolder.getContext().getAuthentication();
    String username = auth.getName();
}
```

### 5. SpEL в @PreAuthorize — шпаргалка

SpEL позволяет писать очень гибкие условия. Вот частые конструкции:

| Выражение | Что означает
|--|--
| `hasRole('ADMIN')` | Есть роль `ROLE_ADMIN`
| `hasAuthority('READ')` | Есть привилегия `READ`
| `hasAnyRole('A', 'B')` | Есть роль A ИЛИ B
| `authentication.name` | Имя текущего пользователя
| `authentication.principal` | Объект `UserDetails` текущего пользователя
| `#id` | Ссылка на параметр метода с именем `id`
| `#user.id == authentication.principal.id` | Сравнение (например, "это мой объект")
| `permitAll` | Всегда разрешено
| `denyAll` | Всегда запрещено

### 6. Что выбрать: URL или Method Security?

- **URL Security (HttpSecurity):** Глобальные правила. Хорошо для разграничения "публичное / для юзеров / для админов". Работает на уровне HTTP-запросов.
- **Method Security:** Точечные правила. Хорошо для бизнес-логики ("пользователь может редактировать только свою статью"). Работает на уровне вызовов методов, включая внутренние вызовы между сервисами.

**Лучшая практика:** Используй оба. URL Security — для общей структуры, Method Security — для тонкой бизнес-логики.


## Глава 4. Безопасность веб-приложений (Web Security)

### 1. CSRF (Cross-Site Request Forgery) — "Защита от подделки запросов"

**Что это за атака?**

Представь: ты залогинен в своем банке в одной вкладке. В другой вкладке ты открываешь фишинговый сайт, который незаметно отправляет POST-запрос в банк ("перевести 10000 руб. на счет хакера"). Браузер автоматически подставит твои куки сессии, и банк подумает, что это ты.

**Как защищает Spring Security?**

По умолчанию Spring Security требует, чтобы каждый небезопасный HTTP-метод (POST, PUT, DELETE, PATCH) содержал специальный CSRF-токен.

- Токен генерируется сервером и кладется в сессию.
- При рендере формы он вставляется в скрытое поле `<input type="hidden" name="_csrf" value="...">`.
- При отправке формы сервер проверяет: совпадает ли токен из запроса с тем, что в сессии?
- Хакер с другого сайта не может узнать этот токен (Same-Origin Policy не даст), поэтому его запрос будет отклонен.

**Как выглядит в коде (для Thymeleaf/JSP):**

```java
<form th:action="@{/transfer}" method="post">
    <!-- Spring автоматически добавит CSRF-токен -->
    <input type="hidden" th:name="${_csrf.parameterName}" th:value="${_csrf.token}"/>
    <input type="number" name="amount"/>
    <button>Перевести</button>
</form>
```

**Когда CSRF нужно ОТКЛЮЧАТЬ?**

Для REST API, который использует JWT или Basic Auth (без куки-сессий).
Почему? Потому что JWT обычно передается в заголовке `Authorization`, а не в куках. Хакер не может подставить этот заголонок с другого сайта — атака невозможна в принципе.

```java
@Bean
public SecurityFilterChain filterChain(HttpSecurity http) throws Exception {
    http
        .csrf(csrf -> csrf.disable()) // <-- для REST API
        // ...
    ;
    return http.build();
}
```

Правило:
- Используются куки-сессии + формы → CSRF ВКЛЮЧЕН (по умолчанию).
- Stateless REST API + JWT/токены в заголовках → CSRF ОТКЛЮЧЕН.

### 2. CORS (Cross-Origin Resource Sharing) — "Разрешаем запросы с других доменов"

**Что это?**
Если твой фронтенд на `http://localhost:3000` (React/Vue), а бэкенд на `http://localhost:8080` (Spring), браузер заблокирует запросы из-за политики Same-Origin. CORS — это способ сказать браузеру: "Я разрешаю запросы с этого домена".

**Важный нюанс!**
В Spring Security CORS нужно настраивать в ДВУХ местах:
1. В Spring MVC (или через `@CrossOrigin`) — чтобы добавить правильные заголовки в ответ.
2. В Spring Security — чтобы пропустить preflight-запрос (OPTIONS) через цепочку фильтров.

Если настроить только Spring MVC — Spring Security может отклонить OPTIONS-запрос до того, как дойдет до контроллера.

**Правильная настройка:**
```java
@Bean
public SecurityFilterChain filterChain(HttpSecurity http) throws Exception {
    http
        .cors(cors -> cors.configurationSource(corsConfigurationSource()))
        // ...
    ;
    return http.build();
}

@Bean
public CorsConfigurationSource corsConfigurationSource() {
    CorsConfiguration configuration = new CorsConfiguration();
    configuration.setAllowedOrigins(List.of("http://localhost:3000")); // фронтенд
    configuration.setAllowedMethods(List.of("GET", "POST", "PUT", "DELETE", "OPTIONS"));
    configuration.setAllowedHeaders(List.of("*"));
    configuration.setAllowCredentials(true); // если используем куки
    
    UrlBasedCorsConfigurationSource source = new UrlBasedCorsConfigurationSource();
    source.registerCorsConfiguration("/**", configuration);
    return source;
}
```

**Важно:** `setAllowCredentials(true)` не работает с `allowedOrigins("*")`. Нужно указывать конкретные домены.

### 3. Управление сессиями (Session Management)

Стратегии создания сессий
```java
http.sessionManagement(session -> session
    .sessionCreationPolicy(SessionCreationPolicy.IF_REQUIRED) // по умолчанию
);
```

Варианты `SessionCreationPolicy`:
- `ALWAYS` — сессия создается всегда.
- `IF_REQUIRED` — создается только если нужна (по умолчанию).
- `NEVER` — Spring не будет создавать сессию, но будет использовать, если она уже есть.
- `STATELESS` — сессии не создаются и не используются. Обязательно для JWT/REST API.

**Защита от Session Fixation**

Атака: хакер "подсовывает" тебе свою сессию (через URL), ты в ней логинишься, и хакер получает доступ к твоему аккаунту.

Spring Security по умолчанию защищает от этого: после успешного логина старая сессия уничтожается и создается новая.
```java
http.sessionManagement(session -> session
    .sessionFixation(sessionFixation -> sessionFixation.migrateSession()) // по умолчанию
);
```

Варианты:
- `none()` — ничего не делать (опасно!).
- `migrateSession()` — создать новую сессию, перенести атрибуты (по умолчанию).
- `newSession()` — создать абсолютно чистую сессию.

**Ограничение количества одновременных сессий**

Хочешь, чтобы пользователь мог быть залогинен только в одном месте?
```java
http.sessionManagement(session -> session
    .maximumSessions(1)
    .maxSessionsPreventsLogin(false) // если true — новый логин будет отклонен
    .expiredSessionStrategy(event -> {
        // что делать при истечении старой сессии
    })
);
```

### 4. HTTP Security Headers — "Заголовки безопасности"

Spring Security по умолчанию добавляет в ответы набор заголовков, которые защищают от различных атак:
```java
http.headers(headers -> headers
    .frameOptions(frame -> frame.deny()) // защита от clickjacking
    .contentTypeContentType(Customizer.withDefaults()) // защита от MIME-sniffing
    .cacheControl(Customizer.withDefaults()) // для защищенных страниц
    // и другие...
);
```

Самые важные:
- **X-Frame-Options: DENY** — запрещает встраивать страницу в `<iframe>` (защита от clickjacking).
- **X-Content-Type-Options: nosniff** — запрещает браузеру угадывать MIME-тип.
- **Strict-Transport-Security (HSTS)** — заставляет браузер использовать только HTTPS.
- **X-XSS-Protection** — включает встроенный XSS-фильтр браузера.

**Важно:** Для REST API часто `frameOptions` и `cacheControl` не нужны, их можно отключить.

### 5. Logout — "Настройка выхода"

```java
http.logout(logout -> logout
    .logoutUrl("/logout") // URL для выхода (POST)
    .logoutSuccessUrl("/login?logout") // куда редиректить после
    .invalidateHttpSession(true) // уничтожить сессию
    .clearAuthentication(true) // очистить SecurityContext
    .deleteCookies("JSESSIONID") // удалить куки
    .logoutSuccessHandler((req, resp, auth) -> {
        // кастомная логика (например, вернуть 200 OK для REST)
        resp.setStatus(HttpServletResponse.SC_OK);
    })
);
```

Для REST API обычно используют кастомный `LogoutSuccessHandler`, который возвращает JSON, а не делает редирект.

### 6. Итоговый пример конфигурации для REST API (JWT)

Соберем всё вместе для типичного современного бэкенда:
```java
@Bean
public SecurityFilterChain filterChain(HttpSecurity http) throws Exception {
    http
        // 1. Отключаем CSRF (используем JWT, не куки)
        .csrf(csrf -> csrf.disable())
        
        // 2. Настраиваем CORS
        .cors(cors -> cors.configurationSource(corsConfigurationSource()))
        
        // 3. Stateless сессии (JWT)
        .sessionManagement(session -> 
            session.sessionCreationPolicy(SessionCreationPolicy.STATELESS))
        
        // 4. Правила авторизации
        .authorizeHttpRequests(auth -> auth
            .requestMatchers("/api/auth/**").permitAll()
            .requestMatchers("/api/admin/**").hasRole("ADMIN")
            .anyRequest().authenticated()
        )
        
        // 5. Добавляем JWT-фильтр ДО стандартного UsernamePasswordFilter
        .addFilterBefore(jwtAuthenticationFilter, 
            UsernamePasswordAuthenticationFilter.class)
        
        // 6. Отключаем ненужные заголовки для REST
        .headers(headers -> headers.frameOptions(frame -> frame.disable()));
    
    return http.build();
}
```

