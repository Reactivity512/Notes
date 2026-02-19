# Java 8

Java 8 была выпущена 18 марта 2014 года.

## Устаревшие (deprecated) методы и классы

### `java.lang`

* `Thread.destroy()` и `Thread.stop(Throwable)` – окончательно помечены как `@Deprecated` (опасные методы для остановки потоков).

* `Object.finalize()` – всё ещё работает, но официально не рекомендуется (лучше использовать `Cleaner` или `try-with-resources`).

### `java.util`

* `Hashtable` – не deprecated, но рекомендуется использовать  `ConcurrentHashMap` или `HashMap`.
* `Vector` – устарел в пользу `ArrayList` + `Collections.synchronizedList()`.
* `Stack` – лучше использовать `Deque` (например, `ArrayDeque`). (Stack — это класс, основанный на наследовании от `Vector`, и его методы синхронизированы, что может негативно влиять на производительность.)

### `java.util.Date` и Calendar

* `java.util.Date` – не deprecated, но все его конструкторы и большинство методов устарели. Вместо него: `java.time.LocalDate`, `LocalDateTime`, `ZonedDateTime`.

* `java.util.Calendar` – также считается устаревшим.

### `java.io`

`FileInputStream.finalize()` и `FileOutputStream.finalize()` – deprecated (уязвимы к утечкам памяти).


## Изменения в Java 8

### Удаление PermGen → Metaspace

* PermGen (Permanent Generation) – удалён, теперь метаданные классов хранятся в Metaspace (в native-памяти).
* Больше не нужно настраивать `-XX:MaxPermSize`.
* Проблема: Metaspace может бесконечно расти (нужно лимитировать `-XX:MaxMetaspaceSize`).

Метапространство (Metaspace) — это область памяти в Java, которая используется для хранения метаданных классов, таких как информация о классах, методах, полях и других структурах, связанных с загрузкой классов.

### Изменения в String

* До Java 8 подстроки (`substring`) ссылались на исходный массив` char[]`, что могло приводить к утечкам памяти.
* После Java 8 – `substring` создаёт новый массив.

### Изменения в HashMap

В Java 8 `HashMap` при коллизиях использует сбалансированные деревья вместо связных списков (при большом количестве коллизий).
Если количество элементов в бакете превышает определенный порог (по умолчанию 8), связный список преобразуется в сбалансированное дерево (обычно красно-черное дерево). Это значительно ускоряет операции поиска, вставки и удаления при большом количестве коллизий.

### Nashorn вместо Rhino

Rhino (старый JS-движок, разработанный Mozilla) удалён, теперь только Nashorn.
Nashorn — это более современный и производительный движок JavaScript, который:
* Производительный: Nashorn обеспечивает более быструю работу скриптов благодаря использованию новых технологий JDK.
* Лучше интегрирован: лучше интегрируется с Java, поддерживает новые стандарты ECMAScript.
* Лучше поддержка: Nashorn был включен по умолчанию в JDK 8 и далее.


## Список нововведений java 8

В Java 8 появилось множество важных нововведений:

1. Лямбда-выражения (Lambda Expressions)
2. Функциональные интерфейсы (Functional Interfaces)
3. Stream API
4. Методы в интерфейсах (Default & Static Methods)
5. Ссылки на методы (Method References)
6. Новое API для работы с датами (java.time)
7. Optional
8. Nashorn – JavaScript-движок
9. Параллельная сортировка (Arrays.parallelSort)
10. CompletableFuture
11. Аннотации на типы (Type Annotations)

**Список дополнительных нововведений Java 8**

12. String.join()
13. Comparator.comparing() и методы для сравнения
14. Параллельные операции с массивами (parallelPrefix, parallelSetAll)
15. Улучшения в ConcurrentHashMap
16. Удаление PermGen (метаданные классов в Metaspace)
17. java.util.Base64
18. StampedLock
19. Расширение аннотаций (@Repeatable)
20. java.util.Random → ThreadLocalRandom
21. DoubleAccumulator, DoubleAdder
22. Files.list() и Files.walk()
23. JJS (Nashorn Command-Line Tool)
24. Улучшенные аннотации (@FunctionalInterface)

### 1.Лямбда-выражения (Lambda Expressions)

Лямбда-выражения позволяют писать более компактный и функциональный код. Они особенно полезны при работе с коллекциями, потоками (Streams) и функциональными интерфейсами.

Лямбда-выражения:
* Не имеет имени.
* Может быть передана как аргумент метода.
* Может использоваться для реализации функциональных интерфейсов.

```java
// До Java 8
Runnable runnable = new Runnable() {
    @Override
    public void run() {
        System.out.println("Hello, world!");
    }
};

// В Java 8
Runnable runnable = () -> System.out.println("Hello, world!");

runnable.run();
```

Пример лямбды с параметрами:

```java
// Функциональный интерфейс
interface MathOperation {
    int operate(int a, int b);
}

// Использование лямбды
MathOperation add = (a, b) -> a + b;
MathOperation multiply = (x, y) -> x * y;

System.out.println(add.operate(5, 3));       // 8
System.out.println(multiply.operate(2, 4));  // 8
```

Пример лямбды в forEach (со списком):

```java
List<String> names = Arrays.asList("Alice", "Bob", "Charlie");

// Старый способ (анонимный класс)
names.forEach(new Consumer<String>() {
    @Override
    public void accept(String name) {
        System.out.println(name);
    }
});

// С лямбдой
names.forEach(name -> System.out.println(name));

// Ссылка на метод (Method Reference)
names.forEach(System.out::println);
```

Лямбда может использовать:

* **Локальные переменные** (должны быть final или effectively final).
* **Поля класса** и **статические переменные** (можно изменять).

```java
// Пример с effectively final
int x = 10;
Runnable r = () -> System.out.println(x);  // OK, x не меняется
// x = 20;  // Ошибка: переменная должна быть effectively final
```

```java
// Пример с полем класса
class Example {
    int count = 0;

    void increment() {
        Runnable r = () -> count++;  // Можно изменять поле
        r.run();
    }
}
```

Разница между лямбдой и анонимным классом

| Характеристика | Лямбда | Анонимный класс |
|--|--|--|
| Синтаксис |   Короткий (x -> x + 1) | Громоздкий (new Runnable() {...}) |
| this | Ссылается на внешний класс | Ссылается на себя |
| Захват переменных | Только final/effectively final | Может использовать любые |

Ограничения лямбд
* Не могут изменять локальные переменные.
* Не могут содержать return с меткой.
* Не могут иметь перегрузку (т.к. нет имени).

### 2.Функциональные интерфейсы (Functional Interfaces)

Лямбда работает только с функциональными интерфейсами — теми, у которых один абстрактный метод (SAM).

```java
@FunctionalInterface
interface MyFunction {
    int apply(int a, int b);
}

MyFunction add = (a, b) -> a + b;
System.out.println(add.apply(2, 3)); // 5
```

Встроенные функциональные интерфейсы (`java.util.function`)

| Интерфейс | Сигнатура  |Пример использования|
|--|--|--|
|Consumer<T>| void accept(T t)| x -> System.out.println(x)
|Function<T, R>| R apply(T t)| s -> s.length()
|Predicate<T>| boolean test(T t)| num -> num > 0
|Runnable| void run()| () -> System.out.println("Hi")

```java
Predicate<Integer> isEven = n -> n % 2 == 0;
System.out.println(isEven.test(4));  // true

Function<String, Integer> length = s -> s.length();
System.out.println(length.apply("Java"));  // 4
```

### 3.Stream API
Stream позволяет работать с коллекциями в функциональном стиле (фильтрация, преобразование, агрегация).

* Stream (поток) — это последовательность элементов, поддерживающая различные операции.
* Не хранит данные, а лишь обрабатывает их «на лету».
* Не изменяет исходную коллекцию (все операции возвращают новый Stream).
* Может быть последовательным (stream()) или параллельным (parallelStream()).

Этапы работы со Stream:

1. Создание (`stream()`, `Arrays.stream()`, `Stream.of()`).
2. Промежуточные операции (**Intermediate operations**) - эти операции возвращают новый поток и позволяют трансформировать или фильтровать данные (`filter()`, `map()`, `distinct()`, `sorted()`, `limit()`, `skip()`).
3. Терминальная операция (**Terminal operations**) - эти операции завершают обработку потока и возвращают результат или побочные эффекты (`forEach()`, `collect()`, `reduce()`, `count()`, `anyMatch()`, `allMatch()`).

**Создание Stream:**

```java
// Из коллекции
List<String> names = List.of("Alice", "Bob", "Charlie");
Stream<String> stream = names.stream();

// Из массива
String[] arr = {"A", "B", "C"};
Stream<String> stream = Arrays.stream(arr);

// Из значений
Stream<Integer> numbers = Stream.of(1, 2, 3);

// Бесконечные Stream
Stream<Integer> infiniteNumbers = Stream.iterate(0, n -> n + 1); // 0, 1, 2, 3, ...
Stream<Double> randomNumbers = Stream.generate(Math::random); // 0.12, 0.95, ...
```

**Промежуточные операции (Intermediate Operations):**

```java
// filter(Predicate<T>) — фильтрация
List<String> names = List.of("Alice", "Bob", "Charlie");
List<String> longNames = names.stream()
                              .filter(name -> name.length() > 4)
                              .toList(); // ["Alice", "Charlie"]

//map(Function<T, R>) — преобразование
List<String> names = List.of("Alice", "Bob", "Charlie");
List<Integer> nameLengths = names.stream()
                                 .map(String::length)
                                 .toList(); // [5, 3, 7]

// sorted() / sorted(Comparator<T>) — сортировка
List<Integer> numbers = List.of(3, 1, 4, 2);
List<Integer> sorted = numbers.stream()
                              .sorted()
                              .toList(); // [1, 2, 3, 4]

// distinct() — удаление дубликатов
List<Integer> nums = List.of(1, 2, 2, 3, 3, 3);
List<Integer> unique = nums.stream()
                           .distinct()
                           .toList(); // [1, 2, 3]

// limit(n) / skip(n) — ограничение
Stream.iterate(0, n -> n + 1)
      .limit(5)       // берёт первые 5 элементов
      .forEach(System.out::println); // 0, 1, 2, 3, 4
```

**Терминальные операции (Terminal Operations):**

```java
// forEach(Consumer<T>) — выполнение действия
List.of("A", "B", "C").stream()
                      .forEach(System.out::println); // A B C

// collect(Collector) — сбор в коллекцию
List<String> names = Stream.of("Alice", "Bob")
                           .collect(Collectors.toList());

// count() — подсчёт элементов
long count = Stream.of(1, 2, 3).count(); // 3

// reduce() — агрегация
int sum = Stream.of(1, 2, 3)
                .reduce(0, (a, b) -> a + b); // 6

// anyMatch / allMatch / noneMatch — проверка условий
boolean hasEven = List.of(1, 3, 5).stream()
                                  .anyMatch(n -> n % 2 == 0); // false

boolean allPositive = List.of(1, 2, 3).stream()
                                      .allMatch(n -> n > 0); // true
```

Примеры использования Stream API:

```java
List<String> names = Arrays.asList("Alice", "Bob", "Charlie");
names.stream()
    .filter(name -> name.startsWith("A"))
    .forEach(System.out::println); // Alice

или

List<String> names = Arrays.asList("Alice", "Bob", "Charlie");
names.stream()
     .filter(name -> name.length() > 4)
     .map(String::toUpperCase)
     .forEach(System.out::println);  // "ALICE", "CHARLIE"

или

//  Поиск самого длинного слова
List<String> words = List.of("Java", "Python", "C++");
String longest = words.stream()
                      .max(Comparator.comparing(String::length))
                      .orElse(""); // "Python"

// Группировка по длине строки
Map<Integer, List<String>> byLength =
    Stream.of("a", "bb", "cc", "ddd")
          .collect(Collectors.groupingBy(String::length));
// {1=["a"], 2=["bb", "cc"], 3=["ddd"]}

или

// Объединение строк через запятую
String joined = Stream.of("A", "B", "C")
                      .collect(Collectors.joining(", ")); // "A, B, C"
```

**Параллельные Stream (parallelStream)**

Позволяет ускорить обработку больших данных за счёт многопоточности.

```java
List<Integer> numbers = List.of(1, 2, 3, 4, 5);
int sum = numbers.parallelStream()
                 .mapToInt(n -> n)
                 .sum(); // 15
```

**Осторожно:**
* Подходит только для неблокирующих операций.
* Может быть медленнее для маленьких данных.

Stream API не подходит:
* Если нужно изменить исходную коллекцию.
* Для сложных условий с состоянием.

### 4.Методы в интерфейсах (Default & Static Methods)

Теперь интерфейсы могут иметь реализацию методов:
* **default**-методы – методы с реализацией по умолчанию.
* **static**-методы – статические методы в интерфейсах.

Default методы (методы по умолчанию):

```java
public interface MyInterface {
    void abstractMethod(); // абстрактный метод

    default void defaultMethod() {
        System.out.println("Это метод по умолчанию");
    }
}

public class MyClass implements MyInterface {
    @Override
    public void abstractMethod() {
        System.out.println("Реализация абстрактного метода");
    }
}

public class Main {
    public static void main(String[] args) {
        MyClass obj = new MyClass();
        obj.abstractMethod();   // вызов реализованного метода
        obj.defaultMethod();    // вызов метода по умолчанию
    }
}
```

Static методы:

```java
// Вызываются через имя интерфейса, а не через экземпляр.
public interface MathUtils {
    static int add(int a, int b) {
        return a + b;
    }
}

public class Main {
    public static void main(String[] args) {
        int sum = MathUtils.add(5, 10);
        System.out.println("Сумма: " + sum);
    }
}
```

### 5.Ссылки на методы (Method References)

Позволяют передавать методы как аргументы.

```java
List<String> names = Arrays.asList("Alice", "Bob", "Charlie");
names.forEach(System.out::println); // Аналог: x -> System.out.println(x)
```

### 6.Новое API для работы с датам`и (java.time)
Замена устаревших `java.util.Date` и `java.util.Calendar`:

`LocalDate`, `LocalTime`, `LocalDateTime`
`ZonedDateTime`, `Period`, `Duration`

```java
LocalDate today = LocalDate.now();
LocalDate tomorrow = today.plusDays(1);
```

### 7.Optional

Помогает избежать `NullPointerException`, оборачивая nullable-значения.

```java
Optional<String> name = Optional.ofNullable(getName());
name.ifPresent(System.out::println); // Выведет имя, если оно не null
```
| Метод| Описание | Пример использования |
|--|--|--|
| `isPresent()` | Проверяет наличие значения | `if (opt.isPresent()) { ... }` |
| `get()` | Возвращает значение, если есть, иначе выбрасывает исключение `NoSuchElementException` | `String value = opt.get();` |
| `orElse(T other)` | Возвращает значение или альтернативное значение, если пустой |  `opt.orElse("Default")` |
| `orElseGet(Supplier<? extends T> supplier)` | Возвращает значение или результат функции, если пустой |` opt.orElseGet(() -> computeDefault())` |
| `orElseThrow(Supplier<? extends X> exceptionSupplier)` | Возвращает значение или выбрасывает исключение | `opt.orElseThrow(() -> new RuntimeException("Нет значения"))` |
| `ifPresent(Consumer<? super T> consumer)` | Выполняет действие, если значение есть | `opt.ifPresent(System.out::println);` |
| `filter(Predicate<? super T> predicate)` | Возвращает Optional с тем же значением, если оно соответствует условию; иначе — пустой | `opt.filter(s -> s.length() > 3)` |
| `map(Function<? super T, ? extends U> mapper)` | Преобразует значение внутри Optional, если оно есть | `opt.map(String::toUpperCase)` |

### 8.Nashorn – JavaScript-движок

Встроенный движок для выполнения JavaScript-кода в Java.

```java
ScriptEngine engine = new ScriptEngineManager().getEngineByName("nashorn");
engine.eval("print('Hello Nashorn!')");
```

### 9.Параллельная сортировка (Arrays.parallelSort)

Быстрая сортировка массивов с использованием нескольких потоков. Особенно эффективно для больших массивов на многоядерных процессорах. (Использует Fork/Join фреймворк для распараллеливания процесса сортировки)

Как `Arrays.parallelSort()` использует Fork/Join
1. **Разделение массива:** Большой массив разбивается на части, которые могут сортироваться независимо.
2. **Параллельная сортировка:** Каждая часть сортируется в отдельном потоке, используя рекурсивное деление.
3. **Объединение результатов:** После сортировки подмассивов происходит слияние отсортированных частей, что реализуется внутри фреймворка.
4. **Автоматический выбор стратегии:** В зависимости от размера массива и доступных ядер, `parallelSort()` выбирает оптимальный уровень параллелизма.

```java
int[] numbers = {5, 3, 9, 1};
Arrays.parallelSort(numbers); // [1, 3, 5, 9]
```

### 10.CompletableFuture

Улучшенная работа с асинхронными операциями.

CompletableFuture?
* Обеспечивает возможность выполнения задач асинхронно (в фоновом режиме).
* Позволяет цеплять последовательные операции, обрабатывать результаты или исключения.
* Поддерживает комбинирование нескольких асинхронных задач.
* Обеспечивает более читаемый и управляемый код по сравнению с использованием Future и ExecutorService.

Запуск асинхронных задач:
```java
CompletableFuture.supplyAsync(() -> { // долгие вычисления  return "Результат";});
```

Обработка результата после завершения:
```java
future.thenAccept(result -> {
    System.out.println("Результат: " + result);
});
```

Обработка ошибок:
```java
future.exceptionally(ex -> {
    ex.printStackTrace();
    return null;
});
```

Цепочка операций:
```java
CompletableFuture.supplyAsync(() -> "Hello")
                 .thenApply(s -> s + " World")
                 .thenAccept(System.out::println);
```

Пример использования:

```java
import java.util.concurrent.CompletableFuture;

CompletableFuture.supplyAsync(() -> "Hello")
    .thenApply(s -> s + " World!")
    .thenAccept(System.out::println); // Hello World!

или

public class CompletableFutureExample {
    public static void main(String[] args) {
        // Запускаем асинхронную задачу
        CompletableFuture<String> future = CompletableFuture.supplyAsync(() -> {
            // Имитация долгой операции
            try {
                Thread.sleep(1000);
            } catch (InterruptedException e) {
                e.printStackTrace();
            }
            return "Асинхронный результат";
        });

        // Обработка результата после завершения
        future.thenAccept(result -> System.out.println("Получено: " + result));

        // Ждем завершения всех задач, чтобы программа не завершилась раньше времени
        future.join();
    }
}
```

### 11.Аннотации на типы (Type Annotations)

Теперь аннотации можно ставить не только над методами/классами, но и в объявлениях типов.

```java
List<@NonNull String> names = new ArrayList<>();

или

@interface NonNull {}
public class Example {
    public static void main(String[] args) {
        @NonNull String s = "Hello"; // Аннотация на типе String
        @NonNull Integer n = 42;     // Аннотация на типе Integer
    }
}
```

### 12.String.join()

Упрощённая конкатенация строк через разделитель.

```java
String joined = String.join(", ", "Java", "C++", "Python");  
// "Java, C++, Python"  
```

### 13.Comparator.comparing() и методы для сравнения

Упрощённое создание компараторов

```java
List<Person> people = ...;
people.sort(Comparator.comparing(Person::getName).thenComparing(Person::getAge));
```

### 14.Параллельные операции с массивами (parallelPrefix, parallelSetAll)

```java
int[] arr = {1, 2, 3, 4};
Arrays.parallelPrefix(arr, (a, b) -> a + b); // [1, 3, 6, 10]
```

### 15.Улучшения в ConcurrentHashMap

* Новые методы: `forEach`, `search`, `reduce`.
* Более эффективная параллельная обработка.

### 16.Удаление PermGen (метаданные классов в Metaspace)

`PermGen` заменён на `Metaspace`, что уменьшает риск `OutOfMemoryError`.

### 17.java.util.Base64

Встроенная поддержка кодирования/декодирования Base64.

```java
String encoded = Base64.getEncoder().encodeToString("Java".getBytes());
```

### 18.StampedLock

Новый тип блокировки с оптимизированным чтением-записью.

StampedLock — это класс, который предоставляет механизм для управления конкурентным доступом к разделяемым ресурсам с возможностью использования трех режимов блокировки: чтение, запись и оптимистичное чтение. Он был введён в Java 8 как более гибкая альтернатива `ReentrantReadWriteLock`, позволяющая повысить производительность в сценариях с большим количеством чтений и редких записей.

Режимы блокировки:
* **Write Lock** (запись): эксклюзивный доступ, только один поток может писать.
* **Read Lock** (чтение): совместный доступ, несколько потоков могут читать одновременно.
* **Optimistic Read** (оптимистичное чтение): без блокировки, предполагая, что данные не изменятся. Проверка на изменение после чтения.

```java
import java.util.concurrent.locks.StampedLock;

public class Counter {
    private int count = 0;
    private final StampedLock lock = new StampedLock();

    public void increment() {
        long stamp = lock.writeLock(); // захватить эксклюзивную блокировку
        try {
            count++;
        } finally {
            lock.unlockWrite(stamp); // освободить
        }
    }

    public int get() {
        long stamp = lock.tryOptimisticRead(); // попытка неблокирующего чтения
        int currentCount = count;
        if (!lock.validate(stamp)) { // проверить, не было ли изменений
            // если было изменение, захватить полноценную блокировку
            stamp = lock.readLock();
            try {
                currentCount = count;
            } finally {
                lock.unlockRead(stamp);
            }
        }
        return currentCount;
    }
}
```

* Метод `increment()` использует эксклюзивную блокировку для безопасного увеличения счетчика.
* Метод `get()` использует оптимистичное чтение для повышения производительности. Если данные изменились во время чтения, он повторно захватывает полноценную блокировку.

### 19.Расширение аннотаций (@Repeatable)

Теперь одну аннотацию можно применить несколько раз.

```java
@Schedule(dayOfMonth="last")
@Schedule(dayOfWeek="Fri")
public void doSomething() {}
```

### 20.java.util.Random → ThreadLocalRandom

Более эффективная генерация случайных чисел в многопоточных приложениях.

### 21.DoubleAccumulator, DoubleAdder

Классы для эффективного накопления значений в многопоточной среде.

**DoubleAdder** — быстрый способ суммировать значения в многопоточном окружении.
**DoubleAccumulator** — универсальный инструмент для агрегации с произвольной ассоциативной функцией.

Пример **DoubleAdder**:

```java
import java.util.concurrent.atomic.DoubleAdder;

public class Example {
    public static void main(String[] args) throws InterruptedException {
        DoubleAdder adder = new DoubleAdder();

        // Создаём несколько потоков, которые добавляют значения
        Thread t1 = new Thread(() -> {
            for (int i = 0; i < 1000; i++) {
                adder.add(1.0);
            }
        });
        Thread t2 = new Thread(() -> {
            for (int i = 0; i < 1000; i++) {
                adder.add(2.0);
            }
        });

        t1.start();
        t2.start();
        t1.join();
        t2.join();

        System.out.println("Sum: " + adder.sum()); // Ожидаемый результат: 3000.0
    }
}
```

Пример **DoubleAccumulator** (позволяет выполнять произвольную ассоциативную операцию), нахождение максимума:
```java
import java.util.concurrent.atomic.DoubleAccumulator;
import java.util.function.DoubleBinaryOperator;

public class Example {
    public static void main(String[] args) throws InterruptedException {
        DoubleBinaryOperator maxOperator = Math::max;
        DoubleAccumulator maxAccumulator = new DoubleAccumulator(maxOperator, Double.NEGATIVE_INFINITY);

        // Потоки обновляют максимум
        Thread t1 = new Thread(() -> {
            double[] values = {1.5, 3.2, 4.8};
            for (double v : values) {
                maxAccumulator.accumulate(v);
            }
        });
        Thread t2 = new Thread(() -> {
            double[] values = {2.5, 5.1};
            for (double v : values) {
                maxAccumulator.accumulate(v);
            }
        });

        t1.start();
        t2.start();
        t1.join();
        t2.join();

        System.out.println("Maximum value: " + maxAccumulator.get()); // Ожидаемый результат: 5.1
    }
}
```

### 22.Files.list() и Files.walk()

Упрощённая работа с файлами через Stream API.

```java
Files.list(Paths.get(".")).forEach(System.out::println);
```

### 23.JJS (Nashorn Command-Line Tool)

Консольный инструмент для выполнения JavaScript-кода.

### 24.Улучшенные аннотации (@FunctionalInterface)

Помечает интерфейсы как функциональные.

`@FunctionalInterface` используется для обозначения интерфейсов, предназначенных для использования в качестве функциональных — то есть таких, которые содержат ровно один абстрактный метод. Это помогает компилятору проверять правильность определения интерфейса и повышает читаемость кода.

* **Обозначение**: Аннотация `@FunctionalInterface` явно указывает, что интерфейс предназначен для использования как функциональный интерфейс.
* **Требования**: Интерфейс должен содержать ровно один абстрактный метод. Другие методы могут быть дефолтными или статическими.
* **Преимущество**: Компилятор выдаст ошибку, если интерфейс нарушает эти правила, что помогает избежать ошибок при определении.

```java
@FunctionalInterface
public interface Converter<F, T> {
    T convert(F from);
}
```

```java
@FunctionalInterface
public interface InvalidInterface {
    void method1();
    void method2(); // Ошибка! Более одного абстрактного метода
}
```
