# Java 21

Java 21 была выпущенная в сентябре 2023.

## Virtual Threads (Виртуальные потоки) — JEP 444

Самая важная фича. Легковесные потоки, которые позволяют писать блокирующий код (как синхронный) с производительностью асинхронного (как Netty/Spring WebFlux).

```java
// Раньше: платформенный поток (тяжелый)
Thread.startVirtualThread(() -> {
    // Блокирующий вызов больше не убивает производительность
    var result = httpClient.send(request);
});

// Или через Executors
try (var executor = Executors.newVirtualThreadPerTaskExecutor()) {
    executor.submit(() -> doSomething());
}
```

Итог: миллионы потоков вместо тысяч. Tomcat/Jetty уже умеют использовать их "из коробки".

### Как работало раньше (Platform Threads)

Модель: Каждый поток Java = 1 поток операционной системы (OS thread).

```java
// Платформенный поток (стоит ~1 МБ памяти)
Thread thread = new Thread(() -> {
    // Блокирующий вызов (БД, HTTP, сокет)
    var result = httpClient.send(request); // Поток "спит", ОС держит ресурсы
    process(result);
});
thread.start();
```

* Ограниченное количество — больше 10-50 тысяч потоков убивают ОС (stack memory + context switching)
* Блокирующие операции — поток простаивает, но ресурсы зарезервированы
* Реактивный подход (Spring WebFlox, Netty) — нужен асинхронный код с колбэками/Mono/Flux

```java
// Асинхронный код до Virtual Threads
webClient.get()
    .uri("/api/user")
    .retrieve()
    .bodyToMono(User.class)
    .flatMap(user -> webClient.get()
        .uri("/api/order/" + user.id())
        .retrieve()
        .bodyToMono(Order.class))
    .subscribe(order -> process(order));
// Сложно дебажить, стэктрейс "рвется"
```

### Как стало (Virtual Threads)

Модель: Миллионы легковесных потоков на JVM → десятки OS threads (carrier threads).

```java
// Виртуальный поток (стоит ~200 байт памяти)
Thread.startVirtualThread(() -> {
    var user = fetchUser();   // Блокируется? JVM "монтирует" на другой carrier thread
    var order = fetchOrder(); // Остальные виртуальные потоки работают параллельно
    process(user, order);
});
```

Виртуальный поток при блокировке (I/O, сокет, БД) отмонтируется от carrier thread, который идет выполнять другой виртуальный поток. Когда результат готов — монтируется обратно (на любой свободный carrier).

```
Фазы работы:
[Virtual Thread #1] --блокировка--> [Unmounted] --готово--> [Mounted на любом carrier]
[Carrier Thread ] ----выполняет VT #2, #3, #4...--------
```

### Главные фишки Virtual Threads

1. Миллионы потоков без боли

```java
// До: out-of-memory или замедление
for (int i = 0; i < 100_000; i++) {
    new Thread(() -> fetchUrl()).start(); // падает/тормозит
}

// После: работает легко
try (var executor = Executors.newVirtualThreadPerTaskExecutor()) {
    for (int i = 0; i < 1_000_000; i++) {
        executor.submit(() -> fetchUrl()); // миллион потоков OK
    }
}
```

2. Блокирующий код = производительный. Не нужны `CompletableFuture`, `Mono`, `Flux`

```java
// Синхронный код на виртуальных потоках
var user = userService.findById(userId);      // Блокирующая JDBC
var orders = orderService.findByUser(userId); // Другой запрос
var recommendations = recService.get(orders); // HTTP client

// Под капотом: блокировки не убивают производительность
```

3. `Thread.Builder` API

```java
// Создание с именем, демоном, обработчиком исключений
Thread vThread = Thread.ofVirtual()
    .name("user-fetcher-", 1)    // именованные: user-fetcher-1, -2...
    .inheritInheritableThreadLocals(false)
    .uncaughtExceptionHandler((t, e) -> log.error("Error", e))
    .start(() -> fetchUser());

// Проверка
vThread.isVirtual(); // true
vThread.threadId();  // обычный ID
```

4. `Executors.newVirtualThreadPerTaskExecutor()`

```java
// Старый способ: фиксированный пул под железо
ExecutorService old = Executors.newFixedThreadPool(200);

// Новый способ: новый виртуальный поток на каждую задачу
ExecutorService executor = Executors.newVirtualThreadPerTaskExecutor();
executor.submit(task1);
executor.submit(task2);
// Закрыть (ждет завершения всех, освобождает carrier threads)
executor.close();
```

5. Pin-thread (редкая проблема)

Пиннинг — когда виртуальный поток не может отмонтироваться (привязан к carrier). Причины:

* `synchronized` блок (не `ReentrantLock`)
* Вызов native метода (JNI)

```java
// Плохо (пиннинг)
synchronized(lock) {
    Thread.sleep(1000); // виртуальный поток заморозит carrier
}

// Хорошо
lock.lock();
try {
    Thread.sleep(1000); // отмонтируется нормально
} finally {
    lock.unlock();
}
```

## Pattern Matching for switch (Окончательно) — JEP 441

То, что в 17 было preview, стало production-ready.

```java
Object obj = ...;
switch (obj) {
    case String s when s.length() > 5 -> System.out.println("Long string");
    case String s -> System.out.println("Short string");
    case Integer i -> System.out.println(i * 2);
    case null -> System.out.println("It's null");
    default -> System.out.println("Unknown");
}
```

## Record Patterns — JEP 440

Можно деструктурировать `Record` прямо в `switch` или `instanceof`.

```java
record Point(int x, int y) {}

if (obj instanceof Point(int x, int y)) {
    System.out.println(x + y); // x и y уже извлечены
}

switch (obj) {
    case Point(int x, int y) -> System.out.println(x + y);
    case ColorPoint(Point p, Color c) -> ... // вложенные рекорды
}
```

## Sequenced Collections — JEP 431

У коллекций наконец появился порядок (первый/последний элемент).

```java
SequencedCollection<String> list = new ArrayList<>();
list.addFirst("first");
list.addLast("last");
String first = list.getFirst();
String last = list.getLast();

// Работает для List, Deque, LinkedHashSet, SortedSet...
```

## String Templates (Preview) — JEP 430

Безопасная замена `+` и `String.format()`.

```java
String name = "Ivan";
String info = STR."Hello \{name}!"; // "Hello Ivan!"

// С шаблонами
JSONObject doc = JSON."\{name}: \{age}";

// С безопасным экранированием (SQL, HTML)
String query = SQL."SELECT * FROM users WHERE name = \{name}";
```

## Scoped Values (Preview) — JEP 446

Замена `ThreadLocal` для виртуальных потоков (наследуется, но неизменяем).

```java
static final ScopedValue<String> TOKEN = ScopedValue.newInstance();

ScopedValue.where(TOKEN, "secret").run(() -> {
    // Внутри любого потока (даже виртуального) доступно:
    String token = TOKEN.get();
});
```

## Structured Concurrency (Preview) — JEP 453

Управление группой задач как единым целым (остановить все, если одна упала).

```java
try (var scope = new StructuredTaskScope.ShutdownOnFailure()) {
    Future<String> user = scope.fork(() -> fetchUser());
    Future<Integer> order = scope.fork(() -> fetchOrder());
    
    scope.join();            // ждем оба
    scope.throwIfFailed();   // если любой упал — бросаем исключение
    
    return new Response(user.resultNow(), order.resultNow());
}
```

## Unnamed Patterns & Variables (Preview) — JEP 443

`_` для игнорирования ненужных значений.

```java
try (var _ = conn.getResultSet()) { ... } // ресурс открыт, но не нужен

switch (obj) {
    case Point(int x, int _) -> ... // y не нужен
    case _ -> System.out.println("default");
}
```

## Generational ZGC — JEP 439

ZGC научили делить объекты на "молодые" и "старые" (как G1). Стало быстрее для большинства приложений.

## Deprecations

* Finalization for removal (метод `finalize()` наконец-то хоронят)
* Память о 32-bit Windows (удалена поддержка)
