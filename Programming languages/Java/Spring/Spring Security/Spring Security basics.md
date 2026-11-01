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
