# Java

## JVM (Java Virtual Machine)

### Java Virtual Machine

**Java Virtual Machine** — это виртуальная машина Java, ключевой компонент платформы Java, который выполняет байт-код Java и обеспечивает переносимость программ между разными операционными системами и устройствами.

### Основные функции JVM:

* Исполнение байт-кода (`*.class` файлов) — JVM интерпретирует или компилирует байт-код в машинный код.
* Управление памятью — автоматическое выделение и сборка мусора (Garbage Collection).
* Обеспечение безопасности — контроль доступа к ресурсам через механизмы безопасности Java.
* Платформонезависимость — один и тот же байт-код работает на любой ОС, где есть JVM.

### Как работает JVM?

* **Компиляция**: Исходный код Java (`*.java`) компилируется в байт-код (`*.class`).
    ```java
    javac Main.java → Main.class
    ```
* **Загрузка классов**: ClassLoader загружает `.class`-файлы в память.
* **Верификация**: JVM проверяет байт-код на безопасность.
* **Исполнение**:
    * **Интерпретатор** выполняет байт-код построчно.
    * **JIT-компилятор** (Just-In-Time) оптимизирует часто используемый код в машинный.

### Архитектура JVM:

* **Class Loader** — загружает классы.
* **Runtime Data Areas** (Heap, Stack, Method Area и др.) — управление памятью.
* **Execution Engine** (Интерпретатор + JIT) — исполнение кода.
* **Garbage Collector** — автоматическое освобождение неиспользуемой памяти.

### Почему JVM важна?

* **Write Once, Run Anywhere** — код работает везде, где есть JVM (Windows, Linux, macOS и др.).
* **Автоматическое управление памятью** — не нужно вручную освобождать память, как в C/C++.
* **Оптимизация производительности** — JIT-компилятор ускоряет выполнение.

### Популярные реализации JVM:

* **HotSpot** (от Oracle, самая распространённая, обеспечивает высокую производительность, оптимизации JIT-компиляции и сборки мусора)
* **OpenJDK** (открытая реализация JVM, которая является основой для большинства дистрибутивов Java)
* **OpenJ9** (от Eclipse, оптимизирована для облачных сред).
* **GraalVM** — поддерживает многоязыковое выполнение (JavaScript, Python, Ruby и др.).


## Java Collections

В Java коллекции (Collections) — это набор классов и интерфейсов в рамках Java Collections Framework (JCF), которые предназначены для хранения, обработки и управления группами объектов. Они предоставляют готовые реализации распространённых структур данных.

Все коллекции находятся в пакете `java.util` и делятся на 3 основные категории:

1. **List** (Списки) – Упорядоченные коллекции с возможностью дублирования элементов.
2. **Set** (Множества) – Неупорядоченные коллекции без дубликатов.
3. **Map** (Ассоциативные массивы) – Пары "ключ-значение" (не входят в `Collection`, но относятся к JCF).

### 1.List (Списки)

Хранят элементы в порядке добавления и позволяют дублирование.

**ArrayList** - Динамический массив, который автоматически расширяется по мере добавления элементов. Быстрый доступ по индексу (O(1)), но медленные вставка/удаление в середине (O(n)).
Когда количество добавленных элементов достигает текущей ёмкости внутреннего массива, происходит его расширение. Создаётся новый массив с увеличенным размером. Размер нового массива обычно увеличивается в 1.5 или 2 раза по сравнению с предыдущим (зависит от реализации). Все существующие элементы копируются из старого массива в новый. Происходит обновление ссылки, внутренний массив `ArrayList` теперь указывает на новый массив.

Рекомендуется заранее задавать начальный размер, если ожидается большое количество элементов, чтобы уменьшить количество расширений.

```java
ArrayList<Integer> list = new ArrayList<>(100); // начальный размер 100
```

**LinkedList** - Двусвязный список. Быстрая вставка/удаление (O(1)), но медленный доступ по индексу (O(n)).

~~**Vector**~~ (устаревший)
~~**Stack**~~ (устаревший)

### 2.Set (Множества)

Хранят только уникальные элементы. Порядок зависит от реализации.

**HashSet** - Хранит элементы в хеш-таблице. Порядок не гарантируется. Вставка/поиск – `O(1)`.

**LinkedHashSet** - Сохраняет порядок добавления элементов.

**TreeSet** - Хранит элементы в отсортированном порядке (по возрастанию). Основан на красно-чёрном дереве. Вставка/поиск – `O(log n)`.

### 3.Map (Ассоциативные массивы)

Хранят пары "ключ-значение". Ключи уникальны.

**HashMap** - Основан на хеш-таблице. Порядок не гарантируется. Операции – O(1)

**LinkedHashMap** - Сохраняет порядок добавления пар.

**TreeMap** - Сортирует пары по ключам (в естественном порядке или через `Comparator`). Основан на красно-чёрном дереве (O(log n)).
(`Comparator` — это интерфейс, который используется для определения порядка объектов при их сравнении. `int compare(T o1, T o2);`)

~~**HashTable**~~ (устаревший)

### 4.Queue (Очереди) и Deque (Двусторонние очереди)

Используются для работы по принципу FIFO (первым пришёл — первым ушёл) или LIFO (стек).

**PriorityQueue** - Очередь с приоритетом (элементы сортируются).

```java
Queue<Integer> pq = new PriorityQueue<>();
pq.add(5);
pq.add(1); // Извлечётся сначала 1, затем 5
```

**ArrayDeque** - Двусторонняя очередь на основе массива. Быстрее LinkedList для операций добавления/удаления.

```java
Deque<String> deque = new ArrayDeque<>();
deque.addFirst("A");
deque.addLast("B");
```

### 5.Итераторы и Stream API

`Iterator` / `ListIterator` – Для обхода коллекций.

```java
Iterator<String> it = list.iterator();
while (it.hasNext()) System.out.println(it.next());
```

**Stream API** (Java 8+) – Функциональная обработка коллекций.

```java
list.stream().filter(s -> s.startsWith("A")).forEach(System.out::println);
```

### Сравнение коллекций

| Коллекция | Доступ по индексу | Время доступа |
|--|--|--|
|ArrayList  | Да (`O(1)`)    | `O(1)`    |
|LinkedList | Нет (`O(n)`)   | `O(n)`    |
|HashSet    | Нет            | `O(1)`    |
|TreeSet    | Нет            | `O(log n)`|
|HashMap    | Нет (по ключу) | `O(1)`    |
|TreeMap    | Нет (по ключу) | `O(log n)`|


## Статические анализаторы кода (Static Code Analysis, SCA)

Анализируют исходный код или байт-код без запуска программы.

### Checkstyle

* Проверяет соответствие кода стилевым стандартам (Google Java Style, Sun Code Conventions).
* Пример: отступы, именование переменных, длина методов.

### PMD

Находит "плохие" практики: дублирование кода, неиспользуемые переменные, сложные условия.

### SpotBugs (ранее FindBugs)

Ищет потенциальные баги: NPE, утечки памяти, неэффективные операции.

### SonarQube / SonarLint

Комплексный анализ (баги, уязвимости, "запахи кода"), интеграция с CI/CD.

### Error Prone (от Google)

Находит ошибки на этапе компиляции (например, == вместо equals()).

```java
// Пример ошибки, которую ловит Error Prone:
if (str == "hello") { ... } // Должно быть str.equals("hello")
```


## Динамические анализаторы (Dynamic Analysis)

Анализируют код во время выполнения (runtime).

### JVMTI-инструменты (Java VisualVM, JProfiler, YourKit)

* Профилирование памяти, CPU, потоков.
* Поиск утечек памяти, "узких" мест в коде.

### Java Flight Recorder (JFR) + JDK Mission Control

Встроенный в JVM инструмент для диагностики производительности.

### OWASP ZAP / Burp Suite

Тестирование безопасности (SQL-инъекции, XSS, CSRF).


## Анализаторы зависимостей (Dependency Analysis)

Проверяют уязвимости в используемых библиотеках.

### OWASP Dependency-Check

Сканирует `pom.xml` / `build.gradle` на наличие CVE-уязвимостей.

### Snyk / GitHub Dependabot

Автоматическое обновление зависимостей с фиксами уязвимостей.


## Инструменты для рефакторинга и улучшения кода

### IntelliJ IDEA Inspections

Встроенный анализатор в IDE (оптимизация, замена `for` на `stream` и т. д.).

### ArchUnit

Проверяет архитектурные правила (например, "Сервисы не должны зависеть от контроллеров").

| Инструмент | Тип анализа | Что ищет? | Интеграция |
|--|--|--|--|
|Checkstyle | Статический | Стиль кода | Maven/Gradle/IDE |
|PMD | Статический | Плохие практики, дублирование | CLI/Maven/Gradle |
|SpotBugs | Статический | Баги (NPE, утечки) | Maven/Gradle |
|SonarQube | Статический | Всё (баги, уязвимости, метрики) | CI/CD (Jenkins) |
|JProfiler | Динамический | Производительность, память | GUI/Серверное |
|Dependabot | Зависимости | Уязвимости в библиотеках | GitHub/GitLab |

