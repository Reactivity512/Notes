# Java 17

Java 17 была выпущенная в сентябре 2021 года.

## Sealed Classes (Закрытые классы) — JEP 409

Теперь мы можем ограничивать иерархию наследования.

Суть: Класс или интерфейс может объявить, какие именно классы могут его наследовать (или реализовывать).

```java
// Разрешено наследовать только Dog, Cat (Bird нельзя)
public sealed class Animal permits Dog, Cat {
}

final class Dog extends Animal {}
final class Cat extends Animal {}

// Ошибка компиляции!
// class Bird extends Animal {}
```

Это дает лучший контроль над доменной моделью и идеально работает с `switch` (компилятор знает все возможные подтипы).

## Pattern Matching for switch (Preview) — JEP 406

В 17 версии это еще предварительная функция (preview), но очень долгожданная.

Суть: Можно передавать в `switch` любой объект, а в кейсах проверять его тип и автоматически кастовать.

```java
Object obj = "Hello";

String result = switch (obj) {
    case Integer i -> "Это число: " + i;
    case String s -> "Это строка длины: " + s.length();
    case null -> "Это null"; // Обработка null наконец-то!
    default -> "Что-то другое";
};
```

## Text Blocks (Стали окончательными) — JEP 378

Блоки текста появились еще в 13/14/15, но в 17 они стали окончательными (final). Больше не нужно экранировать кавычки и склеивать строки.


```java
// Было:
String json = "{\n" +
              "  \"name\": \"John\"\n" +
              "}";

// Стало:
String json = """
              {
                "name": "John"
              }
              """;
```

## Records (Records) — JEP 384

Тоже появились раньше, но в 17 стали окончательными. Идеальный способ для DTO, Value Objects.

```java
public record Point(int x, int y) {}

// Автоматически: конструктор, getters (x(), y()), equals, hashCode, toString.
Point p = new Point(3, 4);
System.out.println(p.x()); // 3
```

## `instanceof` Pattern Matching (Окончательно) — JEP 394

Упрощает работу с `instanceof`, убирая явный каст.

Было:
```java
if (obj instanceof String) {
    String s = (String) obj;
    System.out.println(s.length());
}
```

Стало:
```java
if (obj instanceof String s) {
    System.out.println(s.length()); // s уже готов
}
```

## New Random Generators Interface — JEP 356

Новый API для генерации случайных чисел (`java.util.random`). Добавлен интерфейс `RandomGenerator` и куча новых алгоритмов.

```java
// Старый способ (еще работает)
Random old = new Random();

// Новый способ
RandomGenerator random = RandomGenerator.of("L64X128MixRandom");
random.nextInt();
// Или фабричные методы:
RandomGenerator.getDefault();
RandomGeneratorFactory.all().forEach(f -> System.out.println(f.name()));
```

## Deprecation of Security Manager (JEP 411)

**Важно**: SecurityManager был помечен как @Deprecated(forRemoval = true). Его собираются убрать в будущих версиях. Он практически не использовался в обычных приложениях, но некоторые старые системы должны быть готовы к этому.

## Сборщики мусора

* Удален Experimental G1 GC (но сам G1 остался сборщиком по умолчанию).
* ZGC стал production-ready (раньше был экспериментальным).
* Удален CMS Garbage Collector (окончательно).

## MacOS/AArch64 Port (JEP 391)

Официальная поддержка новых чипов Apple M1/M2.

## Новые методы в стандартных API

`String::formatted` — аналог `String.format`.

```java
"%s is %d".formatted("Age", 30);
```

`Files.writeString`, `Files.readString` — удобная работа с текстовыми файлами.
`Instant::getEpochSecond` / `getNano` — вместо костылей с `getEpochSecond()` из ZonedDateTime.
