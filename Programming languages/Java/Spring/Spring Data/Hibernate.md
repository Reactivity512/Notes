# Hibernate

## Архитектура Hibernate и ключевые компоненты: SessionFactory и Session

Hibernate — это ORM-фреймворк, который стоит между вашим Java-кодом и JDBC. Чтобы понимать его поведение, нужно чётко представлять архитектуру и роль двух центральных объектов: SessionFactory и Session. Всё остальное — производные от них.

Общая архитектура (сверху вниз)
```
Application
    ↓
SessionFactory  (тяжёлый, один на приложение)
    ↓
Session         (лёгкий, один на единицу работы)
    ↓
Persistence Context (кэш первого уровня внутри Session)
    ↓
JDBC Connection / DataSource
    ↓
Database
```

И ещё сбоку: кэш второго уровня, Connection Pool, Transaction Manager, Dialect, Interceptor/Listener — всё это живёт на уровне `SessionFactory`.

### SessionFactory

`SessionFactory` — это неизменяемая, потокобезопасная, тяжеловесная фабрика, которая создаётся один раз на всё приложение (на одну базу данных). Внутри неё хранится:

* Метаданные маппинга — все сущности, их поля, связи, стратегии загрузки. Строятся один раз при старте.
* Кэш второго уровня (если включён).
* Настройки подключения — DataSource, пул соединений, Dialect.
* SQL-запросы, скомпилированные из HQL/JPQL (query plan cache).
* Слушатели событий (EventListener) — interceptors, callbacks.
* Настройки кэша запросов, batching, statistics.

**Как создаётся:**

В современном Spring Boot вы этого не делаете вручную — Spring создаёт SessionFactory (точнее, EntityManagerFactory) при старте контекста. Но понимать, что происходит, важно.

В «чистом» Hibernate:
```java
Configuration cfg = new Configuration()
    .addAnnotatedClass(User.class)
    .configure("hibernate.cfg.xml");

SessionFactory sessionFactory = cfg.buildSessionFactory();
```

`buildSessionFactory()` — дорогая операция. Она:

1. Читает все маппинги.
2. Строит метамодель.
3. Проверяет связи.
4. Создаёт пул соединений.
5. Инициализирует кэши.
6. Компилирует именованные запросы.

На старте приложения это занимает секунды, иногда десятки секунд — на больших проектах с тысячами сущностей.

**Ключевые характеристики**

| Свойство | Значение
|--|--
| Потокобезопасность | Да
| Время жизни | Всё приложение (singleton)
| Стоимость создания | Высокая
| Состояние | Неизменяемое после создания (immutable)
| Один на | Приложение × БД
| Spring-аналог | `LocalContainerEntityManagerFactoryBean`

**Типичные ошибки**
* Создавать `SessionFactory` на каждый запрос. Это катастрофа по производительности: каждый раз перестраивается вся метамодель и пул соединений.
* Держать несколько `SessionFactory` без нужды. Обычно это нужно только для multi-tenancy или работы с несколькими БД.
* Считать его «фабрикой соединений». Нет, соединениями управляет `DataSource` / Connection Pool. `SessionFactory` — фабрика сессий, а не соединений.

### Session

`Session` — это лёгкий, недолговечный, НЕ потокобезопасный объект, представляющий одну единицу работы (unit of work) с базой данных. В JPA-терминологии ему соответствует `EntityManager`.

Внутри `Session` живёт Persistence Context — кэш первого уровня, где хранятся все загруженные и управляемые сущности.

**Время жизни**
`Session` создаётся на:
* одну транзакцию (transaction-scoped) — самый частый вариант в Spring;
* один HTTP-запрос (request-scoped) — через `OpenSessionInViewFilter` (спорная практика, см. ниже);
* один пользовательский диалог (long session, редко).

В Spring Boot `Session` открывается и закрывается автоматически благодаря `@Transactional`:
```java
@Transactional
public void updateUser(Long id, String name) {
    User user = session.get(User.class, id); // session открыта
    user.setName(name);
} // session закрыта, изменения зафлашены
```

**Что делает Session**

1. Предоставляет API для работы с БД: `save`, `persist`, `merge`, `delete`, `get`, `load`, `createQuery`, `flush`.
2. Управляет Persistence Context: отслеживает все загруженные сущности, их изменения, грязные поля (dirty checking).
3. Содержит кэш первого уровня: каждый `get`/`load` сначала ищет в кэше, потом в БД.
4. Управляет транзакцией: через `Transaction` (или `EntityTransaction` в JPA).
5. Выполняет flush: синхронизирует изменения из памяти в БД.

**Ключевые характеристики**
| Свойство | Значение
|--|--
| Потокобезопасность | Нет (нельзя шарить между потоками)
| Время жизни | Одна единица работы
| Стоимость создания | Низкая (лёгкий объект)
| Состояние | Мутабельное
| Один на | Транзакцию / запрос / единицу работы

**Типичные ошибки**
* Шарить `Session` между потоками. Не потокобезопасна, приведёт к порче Persistence Context.
* Держать `Session` открытой долго. Кэш первого уровня растёт, память утекает, данные устаревают.
* Использовать `OpenSessionInView` для ленивых загрузок в контроллере. Это антипаттерн: N+1-запросы, непредсказуемое поведение, проблемы с транзакциями. Лучше явно загружать нужные данные через `JOIN FETCH` / `@EntityGraph`.


### Persistence Context — сердце Session

Это самая важная часть, которую нужно понять. `Persistence Context` — это кэш первого уровня, живущий внутри `Session`.

**Что он делает**
1. Хранит все управляемые (persistent) сущности. Загрузили `User` — он здесь.
2. Гарантирует identity. Два вызова `session.get(User.class, 1L)` в одной сессии вернут один и тот же объект (identity map).
3. Dirty checking. При `flush` Hibernate сравнивает текущее состояние сущностей с их снимком (snapshot) на момент загрузки и генерирует `UPDATE` только для изменённых полей.
4. Write-behind. `save`/`persist` не пишут сразу в БД — они регистрируют сущность, а `flush` синхронизирует всё пакетно.

**Состояния сущностей**

| Состояние | Что значит | Как попасть
|--|--|--
| Transient | Новый объект, не связан с сессией | `new User()`
| Persistent | Управляется сессией, в Persistence Context | `session.get()`, `session.persist()`, `session.merge()`
| Detached | Был persistent, но сессия закрыта или объект эвиктнут | После закрытия `Session`, `session.evict()`
| Removed | Помечен на удаление | `session.remove()`

**Пример dirty checking**
```java
@Transactional
public void doWork() {
    User u = session.get(User.class, 1L); // загружен, snapshot сохранён
    u.setName("Bob");                      // изменено в памяти
    // flush при коммите сравнит с snapshot и выполнит UPDATE
}
```

Никакого явного `session.update()` не нужно. Это главная магия Hibernate.


### Как они взаимодействуют

```
Application → SessionFactory.openSession() → Session
                  ↑                              ↓
                  │                     Persistence Context (кэш 1-го уровня)
                  │                              ↓
            (создаётся 1 раз)            JDBC Connection (из пула)
                                                 ↓
                                          Transaction
```

Порядок типичной работы:

1. Старт приложения: создаётся `SessionFactory` (один раз).
2. Запрос пользователя: открывается `Session` (на время запроса или транзакции).
3. Работа: `Session` загружает сущности в Persistence Context, отслеживает изменения.
4. Flush: при коммите или явно — изменения синхронизируются в БД.
5. Закрытие: `Session` закрывается, Persistence Context уничтожается. Сущности становятся Detached.

### Spring-специфика
В Spring Boot всё немного иначе:

1. `SessionFactory` → создаётся Spring как `EntityManagerFactory` (обычно `LocalContainerEntityManagerFactoryBean`).
2. `Session` → обёрнут в `EntityManager` и привязан к текущему потоку через `EntityManagerHolder`.
3. `@Transactional` → управляет открытием/закрытием сессии и транзакции.
4. `OpenEntityManagerInView` → держит сессию открытой до конца HTTP-запроса.

**Полезные настройки**

```properties
# Logical? Нет — физическое соединение берётся из пула
spring.jpa.properties.hibernate.connection.provider_disables_autocommit=true

# Batching — критично для производительности
spring.jpa.properties.hibernate.jdbc.batch_size=50
spring.jpa.properties.hibernate.order_inserts=true
spring.jpa.properties.hibernate.order_updates=true

# Не давать Hibernate трогать БД при старте (используйте Flyway/Liquibase)
spring.jpa.hibernate.ddl-auto=validate

# Открывать сессию в view — не использовать в проде
spring.jpa.open-in-view=false
```

### Итог

| Критерий | `SessionFactory` | `Session`
|--|--|--
| Вес | Тяжёлый | Лёгкий
| Потокобезопасность | Да | Нет
| Время жизни | Приложение | Транзакция / запрос
| Создание | Дорого | Дёшево
| Состояние | Immutable | Mutable
| Содержит | Метаданные, кэш 2-го уровня, пул | Persistence Context, кэш 1-го уровня
| Spring-аналог | `EntityManagerFactory` | `EntityManager`


1. `SessionFactory` — один на приложение, тяжёлый, потокобезопасный. Создаётся при старте, живёт до конца.
2. `Session` — один на единицу работы, лёгкий, не потокобезопасный. Открывается и закрывается на транзакцию.
3. Persistence Context — кэш первого уровня внутри `Session`. Обеспечивает identity, dirty checking, write-behind.
4. Не шарьте `Session` между потоками. Никогда.
5. Не создавайте SessionFactory на каждый запрос. Никогда.
6. `OpenSessionInView` — антипаттерн. Отключайте в проде.
7. После закрытия `Session` сущности становятся Detached. Попытка ленивой загрузки → `LazyInitializationException`.


## Жизненный цикл сущностей и Persistence Context

Это, пожалуй, самая важная тема в Hibernate. Если вы понимаете состояния сущностей и что происходит внутри Persistence Context — 80% багов с Hibernate перестают быть загадкой.

### Что такое Persistence Context

**Persistence Context** — это контейнер управляемых (managed) сущностей внутри `Session`/`EntityManager`. Его ключевые свойства:

1. Это кэш первого уровня. Все загруженные сущности живут здесь до закрытия сессии.
2. Он гарантирует identity. В рамках одной сессии `get(User.class, 1L)` дважды вернёт один и тот же Java-объект (identity map).
3. Он отслеживает изменения. При `flush` Hibernate сравнивает текущее состояние с snapshot и генерирует `UPDATE` только для изменённых полей (dirty checking).
4. Он реализует write-behind. `save`/`persist` не пишут сразу в БД — изменения накапливаются и синхронизируются при `flush`.

```java
User u1 = session.get(User.class, 1L);
User u2 = session.get(User.class, 1L);
System.out.println(u1 == u2); // true — один и тот же объект
```

Это не «оптимизация ради оптимизации». Это семантическая гарантия: в рамках одной единицы работы вы всегда работаете с одним и тем же объектом, и не получится ситуации «два объекта, оба про User с id=1, у одного name=«A», у другого name=«B»».

### Четыре состояния сущности:

### 1. Transient (Преходящий / Новый / Транзиентный)

Объект создан через `new`, но Hibernate о нём ничего не знает.

```java
User user = new User();  // Transient
user.setName("Alice");
```

Это обычный Java-объект, созданный через оператор `new`. Он никак не связан с базой данных и текущей сессией Hibernate.

* **Связь с БД:** Отсутствует. В базе нет записи, соответствующей этому объекту.
* **Идентификатор (ID):** Обычно `null` (если только вы не задали его вручную, но Hibernate это проигнорирует).
* **Отслеживание:** Hibernate не следит за изменениями его полей. Если вы измените поле и ничего не сделаете, в БД ничего не уйдет.
* **Аналогия:** Черновик документа, который вы написали у себя на столе. Он еще не зарегистрирован в архиве.

Как перевести в другое состояние:
* В **Persistent**: вызвать `session.save()` или `session.persist()`.
* В **Detached**: (напрямую почти не переводят, обычно объект сначала сохраняют, а потом закрывают сессию).

### 2. Persistent (Управляемый / Сохраненный / Персистентный)

```java
session.persist(user);   // user стал Persistent
user.setName("Bob");     // изменение отслеживается
// при flush Hibernate сгенерирует UPDATE
```

Объект находится под управлением текущей сессии (`Session`). У него есть представление в базе данных (он там сохранен или будет сохранен при `flush`).

* **Связь с БД:** Прямая. Объекту соответствует строка в таблице.
* **Идентификатор (ID):** Обязательно имеет значение (ID из БД).
* **Отслеживание:** Hibernate внимательно следит за объектом (механизм Dirty Checking). Если вы измените любое поле этого объекта, Hibernate это заметит и при следующем `flush()` автоматически сгенерирует и выполнит SQL-запрос `UPDATE`.
* **Аналогия:** Документ, который вы достали из архива, открыли в системе и редактируете. Система видит все ваши изменения и сохранит их, когда вы нажмете "Сохранить" (flush).

Как попасть в это состояние:

* Из **Transient**: `session.save()`, `session.persist()`.
* Из **Detached**: `session.update()`, `session.saveOrUpdate()`, `session.merge()`.

Как выйти из этого состояния:

* В **Detached**: закрыть сессию (`session.close()`), очистить её (`session.clear()`) или удалить объект из кэша сессии (`session.evict()`).
* В **Removed**: вызвать `session.delete()` или `session.remove()`.

### 3. Detached (Отсоединенный)

```java
session.close();
// все загруженные сущности теперь Detached
```

Объект был в состоянии Persistent (у него есть ID и он есть в БД), но потерял связь с текущей сессией (сессия была закрыта или объект был из неё извлечен).

* **Связь с БД:** Объект существует в базе данных.
* **Идентификатор (ID):** Обязательно имеет значение.
* **Отслеживание:** Hibernate не следит за ним. Если вы измените поле Detached-объекта, в БД ничего не уйдет, пока вы не "подключите" его обратно к сессии.
* **Аналогия:** Документ, который вы физически забрали из архива к себе на стол. Вы можете его читать и даже менять, но архивная система не знает об этом. Чтобы изменения попали в архив, документ нужно вернуть (сделать merge/update).

Как попасть в это состояние:

* Из **Persistent**: закрыть сессию (`session.close()`), очистить кэш (`session.clear()`), или "вытолкнуть" объект (`session.evict()`).

Как перевести в другое состояние:

* В **Persistent**: вызвать `session.update()`, `session.saveOrUpdate()` или `session.merge()` (в новой открытой сессии).

### 4. Removed (Удаленный)

```java
session.remove(user);  // Removed
// при flush → DELETE FROM users WHERE id = ?
```

Объект помечен на удаление. Он все еще связан с сессией, но "приговорен".

* **Связь с БД:** Запись в БД еще существует, но помечена на удаление.
* **Отслеживание:** При следующем `flush()` или `commit()` Hibernate сгенерирует SQL-запрос `DELETE` и удалит строку из базы данных.
* **Аналогия:** Документ, на котором поставили красный штамп "УНИЧТОЖИТЬ". Он еще лежит в системе, но в конце рабочего дня (при commit) его отправят в шредер.

Как попасть в это состояние:

Из **Persistent**: вызвать `session.delete()` или `session.remove()`.

### Итог

Самая частая ошибка новичков — попытка изменить Detached объект и ожидание, что он сам обновится в БД.

Помните: автоматически изменения в БД уходят только из состояния Persistent. Если объект стал Detached (например, вы достали его из БД в одном методе, а обновляете в другом, где сессия уже закрыта), его нужно обязательно "переподключать" через `merge()` или `update()`.

**`persist()` vs `save()`**

| | `persist()` (JPA) | `save()` (Hibernate)
|--|--|--
| Возвращает | `void` | `Serializable` (id)
| `INSERT` | Отложенный (при flush) | Может быть немедленным (если id нужен)
| Вне транзакции | Ошибка | Работает (но плохо)
| Для detached | Ошибка | Создаст новый объект (если id есть — может быть ошибка)

Используйте `persist()`, если работаете через JPA API. Это стандарт.

**`merge()` vs `update()`**

| | `merge()` (JPA) | `update()` (Hibernate)
|--|--|--
| Что делает | Загружает текущее состояние из БД, копирует в новый managed-объект, возвращает его | «Присоединяет» переданный detached-объект к сессии
| Возвращает | Управляемую копию | `void`
| Исходный объект | Остаётся detached | Становится persistent
| Если запись удалена | Создаст новую | Бросит исключение

`merge()` не делает переданный объект managed. Он возвращает другую managed-копию.

```java
User detached = ...;               // Detached
User managed = session.merge(detached);
// detached — по-прежнему Detached
// managed — Persistent
```

**`get()` vs `load()`**

| | get() | load()
|--|--|--
| Запрос к БД | Сразу | Ленивый (прокси)
| Если не найдено | `null` | `EntityNotFoundException` при обращении
| Возвращает | Реальный объект или `null` | Прокси или реальный объект

`load()` возвращает прокси — Hibernate подсовывает подкласс, который загрузит данные при первом обращении к полю (кроме id).

**`refresh()`**

Перечитывает состояние объекта из БД и перезаписывает его в памяти. Полезно, если данные могли измениться извне (другой процесс, native SQL).

```java
session.refresh(user);  // user перезагружен из БД
```

**`evict()` / `detach()`**

Удаляет объект из Persistence Context. Объект становится Detached.

```java
session.evict(user);  // user больше не managed
```

**`clear()`**

Полностью очищает Persistence Context. Все объекты становятся Detached.

```java
session.clear();  // кэш 1-го уровня пуст
```

**`flush()`**

Синхронизирует Persistence Context с БД. Не коммитит транзакцию — только выполняет `INSERT`/`UPDATE`/`DELETE`.

```java
session.flush();  // SQL выполнен, но транзакция не закоммичена
```

Когда происходит автоматический `flush`:

* Перед коммитом транзакции.
* Перед выполнением запроса, который может затронуть затрагиваемые таблицы (auto-flush).
* При явном `flush()`.
* При `session.close()` (если не было коммита — откат).


### Dirty Checking — как Hibernate узнаёт об изменениях

При загрузке сущности Hibernate сохраняет snapshot — копию всех её полей. При flush он сравнивает текущее состояние с snapshot.

```java
@Transactional
public void rename(Long id, String newName) {
    User u = session.get(User.class, id);  // snapshot сохранён
    u.setName(newName);                    // изменено в памяти
    // flush при коммите: snapshot.name != u.name → UPDATE users SET name = ?
}
```

Что это даёт:
* Не нужно вызывать `update()` — изменения подхватываются автоматически.
* Генерируется `UPDATE` только для изменённых полей (если не отключено).
* Если ничего не изменилось — `UPDATE` не выполняется вообще.

Dirty checking срабатывает при `flush`, а не при `set`. Если вы изменили поле и откатили транзакцию — изменение не попадёт в БД (это ок). Но если вы изменили поле и не закоммитили — оно останется в памяти объекта.

### Write-behind — почему save не пишет сразу

`persist()` не выполняет `INSERT` немедленно. Он регистрирует сущность в Persistence Context, а SQL генерируется при `flush`.

Зачем:
* Батчинг. Hibernate может собрать несколько `INSERT` и отправить одним пакетом (`hibernate.jdbc.batch_size`).
* Порядок операций. Hibernate сам выстраивает порядок `INSERT`/`UPDATE`/`DELETE`, чтобы не нарушить foreign keys.
* Экономия. Если транзакция откатится — `INSERT` вообще не выполнялся.

Исключение: если стратегия генерации id — `IDENTITY` (auto-increment), Hibernate вынужден выполнить `INSERT` сразу, чтобы получить сгенерированный id. Это ломает батчинг.

### Identity Map — гарантия уникальности

```java
User u1 = session.get(User.class, 1L);
User u2 = session.get(User.class, 1L);
assert u1 == u2;  // true
```

Как это работает: Persistence Context — это `Map<EntityKey, Object>`, где ключ — `(entityClass, id)`. При `get` Hibernate сначала смотрит в map. Если объект есть — возвращает его. Если нет — загружает из БД и кладёт в map.

Зачем это нужно:
* Гарантия консистентности: не будет двух разных объектов с одинаковым id.
* Производительность: повторные `get` не ходят в БД.
* Корректный dirty checking: Hibernate знает, какой именно объект отслеживать.

> identity map живёт только внутри сессии. Между сессиями — это разные объекты.

```java
// Сессия 1
User u1 = session1.get(User.class, 1L);
session1.close();

// Сессия 2
User u2 = session2.get(User.class, 1L);
assert u1 != u2;  // разные объекты, оба с id=1
```

### Типичные ошибки и подводные камни

1. `LazyInitializationException`
```java
@Transactional
public User getUser(Long id) {
    return session.get(User.class, id);  // возвращает managed
}
// сессия закрыта, user — Detached

// в контроллере:
user.getOrders().size();  // LazyInitializationException!
```

Решение: загружать нужные данные внутри транзакции (`JOIN FETCH`, `@EntityGraph`), либо использовать DTO.

2. `merge()` не делает объект managed

```java
User detached = ...;
session.merge(detached);
detached.setName("new");  // НЕ отслеживается! detached всё ещё Detached
```

Решение: работать с результатом merge():
```java
User managed = session.merge(detached);
managed.setName("new");  // отслеживается
```

3. Изменение detached-объекта без `merge()`

```java
@Transactional
public void update(User user) {  // user пришёл извне (detached)
    user.setName("new");
    // НИЧЕГО не произойдёт — user не managed
}
```

Решение: `session.merge(user)` или загрузить заново и изменить.

4. `save()` vs `persist()` с id

Если у transient-объекта уже есть id (например, присвоен вручную), `save()` может решить, что объект detached, и выполнить `UPDATE` вместо `INSERT`. `persist()` в такой ситуации бросит исключение — это безопаснее.


5. Долгая сессия → утечка памяти

Persistence Context растёт с каждой загруженной сущностью. Если сессия открыта на весь HTTP-запрос (OpenSessionInView) и загружает тысячи объектов — память утекает.

Решение: не держать сессию дольше, чем нужно; периодически `flush()` + `clear()` в пакетной обработке.

6. `flush()` не коммитит

```java
session.flush();  // SQL выполнен, но не закоммичен
// если дальше исключение — откат
```

Решение: `flush()` — для синхронизации, `commit()` — для фиксации.

7. Auto-flush перед запросом

Hibernate делает `flush` перед `createQuery()`, если запрос затрагивает таблицы с «грязными» сущностями. Это может привести к неожиданным `UPDATE` в середине метода.

Решение: понимать это поведение; при необходимости — `FlushMode.MANUAL` / `COMMIT`.

**Пример: все состояния**

```java
@Transactional
public void demo() {
    // 1. Transient
    User user = new User();
    user.setName("Alice");
    
    // 2. Transient → Persistent
    session.persist(user);
    // при flush → INSERT
    
    // 3. Persistent (managed)
    user.setName("Bob");  // dirty checking подхватит при flush
    
    // 4. Persistent → Detached
    session.evict(user);
    user.setName("Charlie");  // НЕ отслеживается
    
    // 5. Detached → Persistent (через merge)
    User managed = session.merge(user);
    // managed — Persistent, user — всё ещё Detached
    
    // 6. Persistent → Removed
    session.remove(managed);
    // при flush → DELETE
}
```

| Операция | Transient | Persistent | Detached | Removed
|--|--|--|--|--
| `persist()` | → Persistent | — | ❌ ошибка | → Persistent (иногда)
| `merge()` | → Persistent (копия) | → Persistent (копия) | → Persistent (копия) | → Persistent (копия)
| `remove()` | ❌ ошибка | → Removed | ❌ ошибка | —
| `refresh()` | ❌ ошибка | Перечитывает | ❌ ошибка | ❌ ошибка
| `evict()` | — | → Detached | — | → Detached
| `get()`/`load()` | — | Возвращает | — | —
| Dirty checking | ❌ | ✅ | ❌ | ✅ (для DELETE)
| Lazy loading | ❌ | ✅ | ❌ (LazyInitException) | ✅

**Главное**

1. Persistence Context = кэш 1-го уровня + identity map + dirty checking + write-behind. Это сердце `Session`.
2. Четыре состояния: Transient, Persistent, Detached, Removed. Переходы — через методы сессии.
3. `persist()` ≠ `save()`, `merge()` ≠ `update()`. Понимайте разницу.
4. `merge()` возвращает копию, исходный объект остаётся detached.
5. Dirty checking срабатывает при flush, а не при `set`.
6. `flush()` не коммитит — только синхронизирует.
7. После закрытия сессии объекты Detached → ленивые поля недоступны.
8. Identity map живёт только внутри сессии — между сессиями объекты разные.
9. Долгая сессия = утечка памяти. Не держите её дольше необходимого.
10. `LazyInitializationException` — симптом неправильной работы с состояниями, а не «баг Hibernate».

## Маппинг сущностей (Entity Mapping)

Маппинг сущностей — это фундамент работы с Hibernate/JPA. Он определяет, как Java-объекты связаны с таблицами базы данных.

### Базовые аннотации: с чего начинается сущность

Любой маппинг начинается с объявления класса сущностью и определения его первичного ключа.

- `@Entity`: Помечает класс как сущность, которая будет связана с таблицей в БД. Это обязательная аннотация.

- `@Table`: Позволяет явно указать имя таблицы. Если её не использовать, Hibernate возьмёт имя класса (например, для `User` таблица будет `User`).

- `@Id`: Определяет поле, которое является первичным ключом (Primary Key).

- `@GeneratedValue`: Указывает, что значение первичного ключа будет генерироваться автоматически. Стратегия `GenerationType.IDENTITY` (автоинкремент в БД) — одна из самых распространённых и простых.

- `@Column`: Используется для тонкой настройки колонки: можно задать её имя, длину, точность, запретить `null` и т.д..

Пример простой сущности:
```java
@Entity
@Table(name = "users")
public class User {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(name = "full_name", nullable = false)
    private String name;
    
    // ... getters and setters
}
```

### Типы связей: от одного к одному до многих ко многим

В JPA существует четыре основных типа связей. Ключевой момент для понимания — это владеющая сторона (owning side). Именно на владеющей стороне определяется, как связь будет отражена в базе данных (например, в какой таблице будет внешний ключ или таблица-связка).

1. `@ManyToOne` (Многие к одному): Самая распространённая и базовая связь. Она напрямую маппится на внешний ключ в таблице. Важно: по умолчанию JPA использует для неё `FetchType.EAGER` (жадная загрузка), что почти всегда является нежелательным.

```java
@Entity
public class Book {
    // ...
    @ManyToOne(fetch = FetchType.LAZY) // Явно указываем ленивую загрузку!
    @JoinColumn(name = "publisher_id") // Имя колонки с внешним ключом
    private Publisher publisher;
}
```

2. `@OneToMany` (Один ко многим): Обратная сторона `@ManyToOne`. В JPA эта связь почти всегда двунаправленная и должна быть невладеющей. Для этого используется атрибут `mappedBy`, который указывает на поле в дочерней сущности, владеющее связью.

```java
@Entity
public class Publisher {
    // ...
    @OneToMany(mappedBy = "publisher", cascade = CascadeType.ALL, orphanRemoval = true)
    private List<Book> books = new ArrayList<>();
}
```

3. `@OneToOne` (Один к одному): Связь, где одна запись в таблице A соответствует ровно одной записи в таблице B. Как и `@ManyToOne`, по умолчанию использует `EAGER` загрузку. Для ленивой загрузки на дочерней стороне (где находится внешний ключ) необходимо использовать `@JoinColumn`.

4. `@ManyToMany` (Многие ко многим): Требует наличия промежуточной таблицы-связки (join table). Владеющая сторона определяет эту таблицу с помощью аннотации `@JoinTable`.

```java
// Владеющая сторона
@Entity
public class Student {
    @ManyToMany
    @JoinTable(
        name = "student_course",
        joinColumns = @JoinColumn(name = "student_id"),
        inverseJoinColumns = @JoinColumn(name = "course_id")
    )
    private Set<Course> courses = new HashSet<>();
}

// Невладеющая сторона
@Entity
public class Course {
    @ManyToMany(mappedBy = "courses")
    private Set<Student> students = new HashSet<>();
}
```

###  Каскады (`CascadeType`) и `orphanRemoval`

Эти механизмы определяют, как операции над одной сущностью влияют на связанные с ней сущности.

- `CascadeType`: Определяет, какие операции (сохранение, удаление, обновление) должны «каскадно» применяться к связанным сущностям.
    - `CascadeType.PERSIST`: При сохранении родителя автоматически сохраняются и новые дочерние сущности.
    - `CascadeType.REMOVE`: При удалении родителя удаляются и все связанные дочерние сущности.
    - `CascadeType.ALL`: Применяет все операции. Используйте с осторожностью, только если дочерняя сущность не может существовать без родителя.
- `orphanRemoval = true`: Более специфичный и мощный инструмент. Он удаляет дочернюю сущность, если её «отвязали» от родителя (например, удалили из коллекции `List<Book>` у `Publisher`), даже если саму родительскую сущность не удаляют. Это идеально подходит для моделирования отношений «композиция».

### Стратегии загрузки (`FetchType`): `LAZY` vs `EAGER`

Это одна из самых критичных для производительности тем.

- `FetchType.LAZY` (Ленивая): Hibernate загружает связанную сущность только в момент первого обращения к ней. Это рекомендуемый подход по умолчанию для всех связей, особенно для коллекций (`@OneToMany`, `@ManyToMany`), так как он предотвращает проблему N+1 и избыточную загрузку данных.
- `FetchType.EAGER `(Жадная): Hibernate загружает связанную сущность немедленно, вместе с родительской. Это может показаться удобным, но на практике часто ведёт к загрузке огромного объёма ненужных данных и проблемам с производительностью.

Главное правило: Явно указывайте `fetch = FetchType.LAZY` для всех связей `@ManyToOne` и `@OneToOne`, так как для них `EAGER` является значением по умолчанию.

### Лучшие практики и типичные ошибки

1. Всегда используйте `FetchType.LAZY` для `@ManyToOne` и `@OneToOne`. Это самое важное правило для избежания проблем с производительностью.
2. Правильно определяйте владеющую сторону. Для `@OneToMany` используйте `mappedBy` на стороне «один», а `@JoinColumn` — на стороне «многие» (владеющей).
3. Будьте осторожны с `CascadeType.ALL`. Применяйте его только для настоящих композиций, где дочерний объект не имеет смысла без родительского.
4. Используйте `orphanRemoval = true` для чистого удаления дочерних сущностей при их удалении из коллекции.
5. Не забывайте про `@JsonIgnoreProperties` при сериализации двунаправленных связей в JSON, чтобы избежать бесконечной рекурсии.
6. Для загрузки ленивых связей используйте `JOIN FETCH` или `@EntityGraph`, а не полагайтесь на `OpenSessionInView`, который считается антипаттерном.
