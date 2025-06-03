# PHP 8.3

PHP 8.3 был выпущен 23 ноября 2023 года

## Что нового в PHP 8.3

1. Типизированные константы классов
2.  Динамическое получение констант класса
3. Новый атрибут #[\Override]
4. Глубокое клонирование readonly-свойств
5. Новая функция json_validate()
6. Новый метод Randomizer::getBytesFromString()
7. Новые методы Randomizer::getFloat() и Randomizer::nextFloat()
8. Линтер командной строки поддерживает несколько файлов

### Устаревшая функциональность и изменения в обратной совместимости

* Более подходящие исключения в модуле `Date/Time`.
* Присвоение отрицательного индекса n пустому массиву теперь гарантирует, что следующим индексом будет n + 1, а не 0.
* Изменения в функции `range()`.
* Изменения в повторном объявлении статических свойств в трейтах.
* Константа `U_MULTIPLE_DECIMAL_SEPERATORS` объявлена устаревшей, вместо неё рекомендуется использовать константу `U_MULTIPLE_DECIMAL_SEPARATORS`.
* Вариант Mt19937 MT_RAND_PHP объявлен устаревшим.
* `ReflectionClass::getStaticProperties()` теперь не возвращает значение null.
* Параметры INI assert.active, assert.bail, assert.callback, assert.exception и assert.warning объявлены устаревшими.
* Вызов функции `get_class()` и `get_parent_class()` без аргументов объявлен устаревшим.

### Новые классы, интерфейсы и функции

* Новые методы `DOMElement::getAttributeNames()`, `DOMElement::insertAdjacentElement()`, `DOMElement::insertAdjacentText()`, `DOMElement::toggleAttribute()`, `DOMNode::contains()`, `DOMNode::getRootNode()`, `DOMNode::isEqualNode()`, `DOMNameSpaceNode::contains()` и `DOMParentNode::replaceChildren()`.
* Новые методы `IntlCalendar::setDate()`, `IntlCalendar::setDateTime()`, `IntlGregorianCalendar::createFromDate()` и `IntlGregorianCalendar::createFromDateTime()`.
* Новые функции `ldap_connect_wallet()` и `ldap_exop_sync()`.
* Новая функция `mb_str_pad()`.
* Новые функции `posix_sysconf()`, `posix_pathconf()`, `posix_fpathconf()` и `posix_eaccess()`.
* Новый метод `ReflectionMethod::createFromMethodName()`.
* Новая функция `socket_atmark()`.
* Новые функции `str_increment()`, `str_decrement()` и `stream_context_set_options()`.
* Новый метод `ZipArchive::getArchiveFlag()`.
* Поддержка генерации EC-ключей с пользовательскими EC-параметрами в модуле OpenSSL.
* Новый параметр INI zend.max_allowed_stack_size для установки максимально допустимого размера стека.
* `php.ini` теперь поддерживает синтаксис резервных значений/значений по умолчанию.
* Анонимные классы теперь доступны только для чтения.

### 1. PHP RFC: Typed class constants (Типизированные константы классов)

Введение автора RFC: Несмотря на огромные усилия, вложенные в улучшение системы типов PHP из года в год, по-прежнему невозможно объявлять типы констант. Это в меньшей степени касается глобальных констант, но может быть источником ошибок и путаницы для констант классов. По умолчанию дочерние классы могут переопределять константы своих родителей, что затрудняет определение их значения и типа, если определяющий класс или константа не являются final:

```php
interface I {
   const TEST = "Test";  // Мы можем предположить, что константа TEST всегда является строкой
}

class Foo implements I {
   const TEST = [];      // Но это может быть массив...
}

class Bar extends Foo {
   const TEST = null;    // или null
}
```

В php 8.3 можно указывать тип констант:

```php
interface I {
   const string PHP = 'PHP 8.3';
}

class Foo implements I {
   const string PHP = [];
}
// Fatal error: Cannot use array as value for class constant
// Foo::PHP of type string
```

### 2. PHP RFC: Dynamic class constant fetch (Динамическое получение констант класса)

Введение автора RFC: PHP реализует различные поиска елемента:
* Переменные `$$foo`
* Свойства `$foo->$bar`
* Статические свойства `Foo::${$bar}`
* Методы `$foo->{$bar}()`
* Статические методы `Foo::{$bar}()`
* Классы для статических свойств `$foo::$bar`
* Классы для статических методов `$foo::bar()`

Одним заметным исключением являются константы класса.

```php
class Foo {
   const BAR = 'bar';
}
$bar = 'BAR';

// В настоящее время это синтаксическая ошибка.
echo Foo::{$bar};

// Вместо этого необходимо использовать функцию `constant`
echo constant(Foo::class . '::' . $bar);
```

В PHP 8.3 извлечение констант класса и членов перечисления с именами переменных становится более простым:

```php
class Foo {
   const BAR = 'bar';
}
$bar = 'BAR';

echo Foo::{$bar};
// Выведет: bar
```

### 3. PHP RFC: Marking overridden methods #[\Override] (Новый атрибут #[\Override])

Введение автора RFC: При реализации интерфейса или наследовании от другого класса PHP выполняет различные проверки, чтобы гарантировать совместимость реализованных методов с ограничениями, налагаемыми интерфейсом или родительским классом. Однако есть одна вещь, которую он не может проверить: Намерение.
Новый атрибут `#[Override]` используется для отображения намерений программиста. В основном это говорит: Я знаю, что этот метод переопределяет родительский метод. Если это когда-нибудь изменится, пожалуйста, дайте мне знать.
Если добавить к методу атрибут `#[\Override]`, то PHP убедится, что метод с таким же именем существует в родительском классе или в реализованном интерфейсе. Добавление атрибута даёт понять, что переопределение родительского метода является намеренным, а также упрощает рефакторинг, поскольку удаление переопределённого родительского метода будет обнаружено.

```php
namespace MyApp\Tests;

use PHPUnit\Framework\TestCase;

final class MyTest extends TestCase
{
   protected bool $myProp;

   // В setUp() была допущена опечатка, и этот метод никогда
   // не будет вызван, поскольку он защищен в финальном
   // классе, который на него не ссылается.
   protected function setUpp(): void
   {
       $this->myProp = true;
   }
   public function testItWorks(): void
   {
       $this->assertTrue($this->myProp);
   }
}
```

```php
namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Support\Facades\Http;

class RssFeed extends Model {

   // В Laravel 5.4 добавили метод refresh() в Eloquent, но у
   // нас уже есть пользовательский метод с тем же именем и
   // сигнатурой, который делает нечто совершенно иное.            
   public function refresh()
   {
       $this->message = Http::get($this->url);
       $this->save();
   }
}
```

```php
use PHPUnit\Framework\TestCase;

final class MyTest extends TestCase {
   protected $logFile;

   protected function setUp(): void {
       $this->logFile = fopen('/tmp/logfile', 'w');
   }

   #[\Override]
   protected function taerDown(): void {
       fclose($this->logFile);
       unlink('/tmp/logfile');
   }
}

// Фатальная ошибка: MyTest::taerDown() имеет атрибут #[\Override],
// но соответствующего родительского метода не существует
```

### 4. PHP RFC: Readonly amendments (Глубокое клонирование readonly-свойств)

Введение автора RFC: PHP 8.1 добавил поддержку свойств только для чтения через PHP RFC: Readonly properties 2.0, а PHP 8.2 добавил поддержку классов только для чтения через PHP RFC: Readonly classes. Однако эти функции все еще имеют некоторые серьезные недостатки, которые следует устранить. Поэтому этот RFC предлагает следующие поправки к исходным RFC:
1. Классы, не предназначенные только для чтения, могут расширять классы, предназначенные только для чтения. ❌ **Отклонено**
2. Свойства, доступные только для чтения, могут быть повторно инициализированы во время клонирования. ✅ **Принято**

В предложение №2: В настоящее время readonly-свойства не могут быть «глубоко клонированы», поскольку при повторном назначении им любого значения выдается ошибка. Это серьезное неудобство, которое ограничивает использование. Предложение устраняет этот недостаток, позволяя повторно инициализировать readonly-свойства при клонировании. Каждое свойство можно инициализировать только один раз; последующие изменения вызовут ошибку. Семантика свойств только для чтения остается прежней, и их изменение разрешено только в закрытой области видимости. Повторная инициализация означает либо присвоение нового значения, либо отмену.

```php
class PHP {
   public string $version = '8.2';
}

readonly class Foo {
   public function __construct(
       public PHP $php
   ) {}

   public function __clone(): void {
       $this->php = clone $this->php;
   }
}

$instance = new Foo(new PHP());
$cloned = clone $instance;

// Fatal error: невозможно изменить свойство Foo::$php, доступное только для чтения
```

```php
class PHP {
   public string $version = '8.2';
}

readonly class Foo {
   public function __construct(
       public PHP $php
   ) {}

   public function __clone(): void {
       $this->php = clone $this->php;
   }
}

$instance = new Foo(new PHP());
$cloned = clone $instance;
$cloned->php->version = '8.3'; // Успешно
```

### 5. PHP RFC: json_validate (Новая функция json_validate())

Введение автора RFC: Большинство реализаций пользовательского пространства используют `json_decode()`, который по замыслу генерирует ZVAL (объект/массив/и т. д.) при разборе строки, следовательно, используя память и обработку, которые можно было бы сэкономить.
Предлагаемая функция будет использовать тот же самый парсер JSON, который уже существует в ядре PHP, который также используется `json_decode()`, это гарантирует, что то, что допустимо в `json_validate()`, также допустимо в `json_decode()`.

```php
json_validate(string $json, int $depth = 512, int $flags = 0): bool
```

Функция `json_validate()` позволяет проверить, является ли строка синтаксически корректным JSON, при этом она более эффективна, чем функция `json_decode()`.

**Предостережение:** Вызов функции `json_validate()` непосредственно перед функцией `json_decode()` приведёт к ненужному двойному разбору строки, поскольку функция `json_decode()` неявно выполняет такую проверку при декодировании. Функцию `json_validate()` вызывают, когда данные JSON нужны не сразу, а необходимо проверить, является ли строка допустимый JSON.

```php
function json_validate(string $string): bool
{
   json_decode($string);

   return json_last_error() === JSON_ERROR_NONE;
}

var_dump(
   json_validate('{ "test": { "foo": "bar" } }')
); // true
```

В php 8.3:
```php
var_dump(
   json_validate('{ "test": { "foo": "bar" } }')
); // true
```

### 6. PHP RFC: Randomizer Additions (Новый метод Randomizer::getBytesFromString(), Randomizer::getFloat() и Randomizer::nextFloat())

Введение автора RFC:
Генерация случайной строки с определенными символами часто используется для создания идентификаторов и кодов ваучеров. Реализация в пользовательской среде требует выбора случайных смещений в цикле, что приводит к избыточному коду и возможным ошибкам, например, при определении максимального индекса строки. Использование `Randomizer::getInt()` для выбора смещений неэффективно, так как требует обращения к движку для каждого символа, в то время как 64-разрядный движок может генерировать случайность для 8 символов одновременно. Генерация случайных значений с плавающей запятой также полезна, но деление случайного целого числа на другое может привести к ошибкам округления и снижению плотности чисел для больших значений.

```php
Random\Randomizer::getBytesFromString(string $string, int $length): string — Получает случайные байты из исходной строки.
```

Вероятность выбора байта пропорциональна его доле во входной строке string. Если каждый байт встречается одинаковое количество раз, вероятность выбора каждого байта будет одинаковой.

```php
// Эту функцию необходимо реализовать вручную.
function getBytesFromString(string $string, int $length) {
   $stringLength = strlen($string);

   $result = '';
   for ($i = 0; $i < $length; $i++) {
       // random_int не подходит для тестирования, но безопасен.
       $result .= $string[random_int(0, $stringLength - 1)];
   }

   return $result;
}

$randomDomain = sprintf(
   "%s.example.com",
   getBytesFromString(
       'abcdefghijklmnopqrstuvwxyz0123456789',
       16,
   ),
);

echo $randomDomain;
```

```php
// Для заполнения можно передать \Random\ Engine,
// по умолчанию используется безопасный движок.
$randomizer = new \Random\Randomizer();

$randomDomain = sprintf(
   "%s.example.com",
   $randomizer->getBytesFromString(
       'abcdefghijklmnopqrstuvwxyz0123456789',
       16,
   ),
);

echo $randomDomain;
```

Объяснение алгоритма:
Рассмотрим представление с плавающей точкой с 3-битной мантиссой, позволяющее представлять 8 значений между степенями двойки. Например, между 1.0 и 2.0 шаги 0.125, а между 2.0 и 4.0 — шаги 0.25. В PHP используется 52-битная мантисса, что позволяет представлять 252 значения между каждой степенью двойки. Например, между 1.0 и 4.0 доступны 16 точных значений:
1.0, 1.125, 1.25, 1.375, 1.5, 1.625, 1.75, 1.875, 2.0, 2.25, 2.5, 2.75, 3.0, 3.25, 3.5, 3.75 и 4.0.
При вызове `$randomizer->getFloat(1.625, 2.5, IntervalBoundary::ClosedOpen)` запрашивается случайное число в интервале от 1.625 до 2.5 (не включая 2.5). Алгоритм определяет размер шага на границе с большим значением (2.5), равный 0.25.
Запрошенный интервал составляет 0.875, что не является точным кратным 0.25. Поэтому алгоритм начинает с верхней границы 2.5 и доступные значения: 2.25, 2.0, 1.75 и 1.625. Значение 2.5 не включается, но 1.625 включено, так как нижняя граница закрыта. В итоге алгоритм случайным образом выбирает одно из четырёх значений и возвращает его.

`Randomizer::nextFloat()` - Возвращает равномерно выбранное равнораспределённое число с плавающей точкой из открытого справа интервала от 0.0 до 1.0, но не включая саму единицу.

```php
$randomizer = new \Random\Randomizer(new MaxEngine);
$min = 3.5;
$max = 4.5;

// ❌ НЕ ДЕЛАЙТЕ ЭТОГО:
// Это выведет значение 4.5, несмотря на выборку метода nextFloat()
// из открытого справа интервала, который никогда не вернёт значение 1.
printf("Неправильное масштабирование: %.17g", $randomizer->nextFloat() * ($max - $min) + $min);

// ✅ Правильно:
$randomizer->getFloat($min, $max, \Random\IntervalBoundary::ClosedOpen);
```

```php
// Возвращает случайное число с плавающей точкой между $min и
// $max, включая оба значения.
function getFloat(float $min, float $max) {
   // Этот алгоритм смещен для определенных входных
   // данных и может возвращать значения за пределами
   // заданного диапазона. Это невозможно обойти в
   // пользовательском пространстве.
   $offset = random_int(0, PHP_INT_MAX) / PHP_INT_MAX;

   return $offset * ($max - $min) + $min;
}

$temperature = getFloat(-89.2, 56.7);

$chanceForTrue = 0.1;

// getFloat(0, 1) может вернуть верхнюю границу, т. е. 1, внося небольшую погрешность.
$myBoolean = getFloat(0, 1) < $chanceForTrue;
```

```php
$randomizer = new \Random\Randomizer();

$temperature = $randomizer->getFloat(
   -89.2,
   56.7,
   \Random\IntervalBoundary::ClosedClosed,
);

$chanceForTrue = 0.1;
// Randomizer::nextFloat() эквивалентно
// Randomizer::getFloat(0, 1, \Random\IntervalBoundary::ClosedOpen).
// Верхняя граница, т. е. 1, не будет возвращена.
$myBoolean = $randomizer->nextFloat() < $chanceForTrue;
```

### Линтер командной строки поддерживает несколько файлов

PHP CLI Lint поддерживает одновременный анализ нескольких файлов

PHP 8.2:
```bash
php -l foo.php bar.php
❌ No syntax errors detected in foo.php
```

PHP 8.3:
```bash
php -l foo.php bar.php
✅ No syntax errors detected in foo.php
✅ No syntax errors detected in bar.php
```

### Отрицательные индексы в массивах

Если у вас есть пустой массив, добавьте элемент с отрицательным индексом, а затем ещё один элемент, этот второй элемент всегда будет начинаться с индекса 0:

```php
$array = [];

$array[-5] = 'a';
$array[] = 'b';

В PHP 8.2
print_r($array); // Array ( [-5] => a [0] => b )

В PHP 8.3
print_r($array); // Array ( [-5] => a [-4] => b )
```

### Магические замыкания методов и именованные аргументы, Анонимные readonly классы

Допустим, у вас есть класс, поддерживающий магические методы:

```php
class Test {
   public function __call($name, $args)
   {
       var_dump($name, $args);
   }

   public static function __callStatic($name, $args) {
       var_dump($name, $args);
   }
}
```

PHP 8.3 позволяет создавать замыкания из этих методов, а затем передавать именованные аргументы этим замыканиям. Ранее это было невозможно.

```php
$test = new Test();
$closure = $test->magic(...);
$closure(a: 'hello', b: 'world');
```

Ранее нельзя было пометить анонимные классы, как readonly. Это исправлено в PHP 8.3:

```php
$class = new readonly class {
   public function __construct(
       public string $foo = 'bar',
   ) {}
};
```

### Инвариантная видимость констант

Ранее видимость констант не проверялась при реализации интерфейса. В PHP 8.3 эта ошибка исправлена, но в некоторых местах она может привести к поломке кода, если вы не знали о таком поведении.

```php
interface I {
   public const FOO = 'foo';
}

class C implements I {
   private const FOO = 'foo'; // Ошибка в PHP 8.3
}
```
