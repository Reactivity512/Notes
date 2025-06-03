# PHP 8.4

PHP 8.4 был выпущен 21 ноября 2024 года

## Что нового в PHP 8.4

1. Хуки свойств 
2. Асимметричная область видимости свойств
3. Атрибут #[\Deprecated]
4. Новые возможности ext-dom и поддержка HTML5
5. Объектно-ориентированный API для BCMath 
6. Новые функции array_*()
7. SQL-парсеры, специфичные для драйверов PDO
8. new MyClass()->method() без скобок

### Устаревшая функциональность и изменения в обратной совместимости

* Модули IMAP, OCI8, PDO_OCI и pspell перенесены из ядра в PECL.
* Типы параметров, неявно допускающие значение null объявлены устаревшими.
* Использование _ в качестве имени класса объявлено устаревшим.
* Возведение нуля в степень отрицательного числа объявлено устаревшим.
* Передача некорректного режима в функцию `round()` выбрасывает ошибку ValueError.
* Константы классов модулей `date`, `intl`, `pdo`, `reflection`, `spl`, `sqlite` и `xmlreader` типизированы.
* Класс GMP теперь является окончательным.
* Удалены константы `MYSQLI_SET_CHARSET_DIR`, `MYSQLI_STMT_ATTR_PREFETCH_ROWS`, `MYSQLI_CURSOR_TYPE_FOR_UPDATE`, `MYSQLI_CURSOR_TYPE_SCROLLABLE` и `MYSQLI_TYPE_INTERVAL`.
* Функции `mysqli_ping()`,` mysqli_kill()`, `mysqli_refresh()`, методы `mysqli::ping()`, `mysqli::kill()`, `mysqli::refresh()` и константы MYSQLI_REFRESH_* объявлены устаревшими.
* Функции `stream_bucket_make_writeable()` и `stream_bucket_new()` теперь возвращают экземпляр класса StreamBucket вместо stdClass.
* Изменение поведения языковой конструкции `exit()`.
* Константа `E_STRICT` объявлена устаревшей.

### Новые классы, интерфейсы и функции

* Добавлены ленивые объекты.
* Новая реализация JIT на основе IR Framework.
* Добавлена функция `request_parse_body()`.
* Добавлены функции b`cceil()`, `bcdivmod()`, `bcfloor()` и `bcround()`.
* Добавлено перечисление RoundingMode для функции `round()` с 4 режимами: `TowardsZero`, `AwayFromZero`, `NegativeInfinity` и `PositiveInfinity`.
* Добавлены методы `DateTime::createFromTimestamp()`,` DateTime::getMicrosecond()`, `DateTime::setMicrosecond()`, `DateTimeImmutable::createFromTimestamp()`, `DateTimeImmutable::getMicrosecond()` и `DateTimeImmutable::setMicrosecond()`.
* Добавлены функции `mb_trim()`, `mb_ltrim()`, `mb_rtrim()`, `mb_ucfirst()` и `mb_lcfirst()`.
* Добавлены функции `pcntl_getcpu()`, `pcntl_getcpuaffinity()`, `pcntl_getqos_class()`, `pcntl_setns()` и `pcntl_waitid()`.
* Добавлены методы `ReflectionClassConstant::isDeprecated()`, `ReflectionGenerator::isClosed()` и `ReflectionProperty::isDynamic()`.
* Добавлены функции `http_get_last_response_headers()`, `http_clear_last_response_headers()`, `fpow()`.
* Добавлены методы `XMLReader::fromStream()`, `XMLReader::fromUri()`, `XMLReader::fromString()`, `XMLWriter::toStream()`, `XMLWriter::toUri()` и `XMLWriter::toMemory()`.
* Добавлена функция `grapheme_str_split()`.

### 1. PHP RFC: Property hooks (Хуки свойств)

Введение автора RFC: Разработчики часто создают методы getFoo/setFoo не из-за текущей необходимости, а чтобы иметь возможность изменить свойства на методы без изменения API в будущем. Альтернативой являются методы __get и __set, которые перехватывают чтение и запись свойств, но это плохой подход, который перехватывает все неопределенные (и некоторые определенные) свойства безоговорочно. Хуки свойств предлагают более целенаправленный подход для взаимодействия с общими свойствами, позволяя прикреплять шаблоны getFoo/setFoo к свойствам. Можно добавить это поведение позже, без изменения API и без создания лишних методов для каждого свойства «на всякий случай».

PHP 8.3:

```php
class Locale
{
   private string $languageCode;
   private string $countryCode;

   public function __construct(string $languageCode, string $countryCode)
   {
       $this->setLanguageCode($languageCode);
       $this->setCountryCode($countryCode);
   }

   public function getLanguageCode(): string
   {
       return $this->languageCode;
   }

   public function setLanguageCode(string $languageCode): void
   {
       $this->languageCode = $languageCode;
   }

   public function getCountryCode(): string
   {
       return $this->countryCode;
   }

   public function setCountryCode(string $countryCode): void
   {
       $this->countryCode = strtoupper($countryCode);
   }

   public function setCombinedCode(string $combinedCode): void
   {
       [$languageCode, $countryCode] = explode('_', $combinedCode, 2);

       $this->setLanguageCode($languageCode);
       $this->setCountryCode($countryCode);
   }

   public function getCombinedCode(): string
   {
       return \sprintf("%s_%s", $this->languageCode, $this->countryCode);
   }
}
```

PHP 8.4:

```php
class Locale
{
   public string $languageCode;
   public string $countryCode
   {
       set (string $countryCode) {
           $this->countryCode = strtoupper($countryCode);
       }
   }

   public string $combinedCode
   {
       get => \sprintf("%s_%s", $this->languageCode, $this->countryCode);
       set (string $value) {
           [$this->languageCode, $this->countryCode] = explode('_', $value, 2);
       }
   }

   public function __construct(string $languageCode, string $countryCode)
   {
       $this->languageCode = $languageCode;
       $this->countryCode = $countryCode;
   }
}
```

Пример:

```php
class Person
{
   // «Виртуальное» свойство. Невозможно установить значение виртуального свойства явным образом
   public string $fullName {
       get => $this->firstName . ' ' . $this->lastName;
   }

   // Каждая операция записи значения свойства пройдёт через хук. В свойство в итоге запишется значение, которое вернётся из хука.
   // Доступ к свойству для чтения значения проходит в стандартном режиме
   public string $firstName {
       set => mb_ucfirst(strtolower($value));
   }

   // Каждая операция записи значения свойства пройдёт через хук, который сам запишет реальное значение свойства.
   // Доступ к свойству для чтения значения проходит в стандартном режиме
   public string $lastName {
       set {
           if (strlen($value) < 2) {
               throw new \InvalidArgumentException('Слишком короткая фамилия');
           }

           $this->lastName = $value;
       }
   }
}

$p = new Person();

$p->firstName = 'пётр';
print $p->firstName; // Конструкция выведет "Пётр"

$p->lastName = 'Петров';
print $p->fullName; // Конструкция выведет "Пётр Петров"
```

### 2. PHP RFC: Asymmetric Visibility v2 (Асимметричная область видимости свойств)

Введение автора RFC: PHP уже давно имеет возможность контролировать видимость свойств объектов — публичных, закрытых или защищенных. Однако этот контроль всегда одинаков для операций get и set. То есть они «симметричны». Этот RFC предлагает разрешить свойствам иметь отдельную («асимметричную») видимость, с отдельной видимостью для операций чтения и записи.

```php
class Foo
{
   public private(set) string $bar = 'baz';
}

$foo = new Foo();
var_dump($foo->bar); // prints "baz"
$foo->bar = 'beep';  // Visibility error
```

Область видимости записи свойства теперь может контролироваться независимо от области видимости чтения свойства.

```php
class PhpVersion
{
   private string $version = '8.3';

   public function getVersion(): string
   {
       return $this->version;
   }

   public function increment(): void
   {
       [$major, $minor] = explode('.', $this->version);
       $minor++;
       $this->version = "{$major}.{$minor}";
   }
}
```

```php
class PhpVersion
{
   public private(set) string $version = '8.4';

   public function increment(): void
   {
       [$major, $minor] = explode('.', $this->version);
       $minor++;
       $this->version = "{$major}.{$minor}";
   }
}
```

### 3. PHP RFC: #[\Deprecated] Attribute (Атрибут #[\Deprecated])

Введение автора RFC: Внутренние функции PHP и (класс-)константы могут быть помечены как устаревшие, что делает эту информацию доступной для Reflection и выдает ошибки устаревания (**E_DEPRECATED**), но эквивалентной функциональности для функций, определенных в пользовательском пространстве, не существует.
Атрибут помечает функциональность устаревшей. Устаревшая функциональность вызывает ошибки уровня **E_USER_DEPRECATED**.

Обзор класса:

```php
final class Deprecated {
   /* Свойства */
   public readonly ?string $message;
   public readonly ?string $since;
   /* Методы */
   public __construct(?string $message = null, ?string $since = null)
}
```

Новый атрибут #[\Deprecated] расширяет существующий механизм объявления сущности устаревшей для пользовательских функций, методов и констант классов.

```php
class PhpVersion
{
   /**
    * @deprecated 8.3 use PhpVersion::getVersion() instead
    */
   public function getPhpVersion(): string
   {
       return $this->getVersion();
   }

   public function getVersion(): string
   {
       return '8.3';
   }
}

$phpVersion = new PhpVersion();
// Нет никаких указаний на то, что метод устарел.
echo $phpVersion->getPhpVersion();
```

```php
class PhpVersion
{
   #[\Deprecated(
       message: "use PhpVersion::getVersion() instead",
       since: "8.4",
   )]
   public function getPhpVersion(): string
   {
       return $this->getVersion();
   }

   public function getVersion(): string
   {
       return '8.4';
   }
}

$phpVersion = new PhpVersion();

// Deprecated: Method PhpVersion::getPhpVersion() is deprecated since 8.4,
// use PhpVersion::getVersion() instead in php-wasm run script on line 20
echo $phpVersion->getPhpVersion();  
```

### 4. PHP RFC: New ext-dom features in PHP 8.4, PHP RFC: DOM HTML5 parsing and serialization (Новые возможности ext-dom и поддержка HTML5)

Введение автора RFC: Расширение DOM PHP позволяет загружать HTML-документы с помощью методов \DOMDocument::loadHTML и \DOMDocument::loadHTMLFile, используя HTML-парсер libxml2 для разбора в дерево документов. Однако этот парсер поддерживает только HTML до версии 4.01, что создает проблемы, так как HTML5 стал стандартом за последнее десятилетие. Внедрение поддержки HTML5 в DOM PHP критически важно для улучшения обработки современного веб-контента. Использование loadHTML(File) для загрузки HTML5 приводит к ошибкам парсинга и неправильным деревьям документов из-за изменений в правилах между HTML4 и HTML5. Текущий парсер не распознает семантические теги HTML5 (например, main, article, section) и сталкивается с проблемами вложенности элементов, что приводит к неправильным деревьям документов.
Введение автора RFC: Поддержку селекторов CSS, заполнении отсутствующих функций и добавлении новых свойств.
Новый DOM API, поддерживает разбор HTML5-документов в соответствии со стандартами, исправляет несколько давних ошибок в поведении DOM и добавляет несколько функций, делающих работу с документами более удобной. Документы, использующие новый DOM API, могут быть созданы с помощью классов Dom\HTMLDocument и Dom\XMLDocument.

```php
$dom = new DOMDocument();
$dom->loadHTML(
   <<<'HTML
       <main>
           <article>PHP 8.4 is a feature-rich release!</article>
           <article class="featured">PHP 8.4 adds new DOM classes that are spec-compliant, keeping the old ones for compatibility.</article>
       </main>
       HTML',
   LIBXML_NOERROR,
);

$xpath = new DOMXPath($dom);
$node = $xpath->query(".//main/article[not(following-sibling::*)]")[0];
$classes = explode(" ", $node->className); // Simplified
var_dump(in_array("featured", $classes)); // bool(true)
```

```php
$dom = Dom\HTMLDocument::createFromString(
   <<<'HTML
       <main>
           <article>PHP 8.4 is a feature-rich release!</article>
           <article class="featured">PHP 8.4 adds new DOM classes that are spec-compliant, keeping the old ones for compatibility.</article>
       </main>
       HTML',
   LIBXML_NOERROR,
);

$node = $dom->querySelector('main > article:last-child');
var_dump($node->classList->contains("featured")); // bool(true)
```

### 5. PHP RFC: Support object type in BCMath (Объектно-ориентированный API для BCMath)

Введение автора RFC: BCMath в настоящее время поддерживает только процедурную функциональность и не имеет поддержки объектно-ориентированного программирования, что делает его несколько устаревшим. В отличие от него, расширение GMP уже поддерживает типы объектов. Хотя можно рассматривать BCMath как объект в пользовательском пространстве, такие функции, как перегрузка операторов, не могут быть реализованы.
Новый объект BcMath\Number позволяет использовать объектно-ориентированный стиль и стандартные математические операторы при работе с числами произвольной точности.
Эти объекты неизменяемы и реализуют интерфейс Stringable, поэтому их можно использовать в строковых контекстах, например, echo $num.

```php
$num1 = '0.12345';
$num2 = 2;
$result = bcadd($num1, $num2, 5);

echo $result; // '2.12345'
var_dump(bccomp($num1, $num2) > 0); // false
```

```php
use BcMath\Number;

$num1 = new Number('0.12345');
$num2 = new Number('2');
$result = $num1 + $num2;

echo $result; // '2.12345'
var_dump($num1 > $num2); // false
```

### 6. PHP RFC: array_find (Новые функции array_*())

Введение автора RFC: Этот RFC предлагает добавить новые функции массива: `array_find`, `array_find_key`, `array_any` и `array_all`, которые являются вспомогательными функциями для общих шаблонов проверки массива на наличие элементов, соответствующих определенному условию. Сейчас есть функции для обработки массивов с использованием обратного вызова, но отсутствуют функции для поиска одного элемента, соответствующего условию, и для проверки наличия таких элементов.

```php
$animal = null;
foreach (['dog', 'cat', 'cow', 'duck', 'goose'] as $value) {
   if (str_starts_with($value, 'c')) {
       $animal = $value;
       break;
   }
}

var_dump($animal); // string(3) "cat
```

```php
$animal = array_find(
   ['dog', 'cat', 'cow', 'duck', 'goose'],
   static fn (string $value): bool => 
       str_starts_with($value, 'c'),
);

var_dump($animal); // string(3) "cat"
```

### 7. PHP RFC: PDO driver specific sub-classes (SQL-парсеры, специфичные для драйверов PDO)

Введение автора RFC: PDO — универсальный класс базы данных, поддерживающий функциональность, специфичную для разных баз данных. Например, при подключении к SQLite доступна функция PDO::sqliteCreateFunction. Наличие методов в классе, зависящих от компиляции и подключаемой базы данных, усложняет код. Было бы проще использовать подклассы PDO для каждой базы данных с определенной функциональностью.
Добавлены дочерние классы Pdo\Dblib, Pdo\Firebird, Pdo\MySql, Pdo\Odbc, Pdo\Pgsql, Pdo\Sqlite драйверов, наследующие PDO.

```php
$connection = new PDO(
   'sqlite:foo.db',
   $username,
   $password,
); // object(PDO)

$connection->sqliteCreateFunction(
   'prepend_php',
   static fn ($string) => "PHP {$string}",
);

$connection->query('SELECT prepend_php(version) FROM php');
```

```php
$connection = PDO::connect(
   'sqlite:foo.db',
   $username,
   $password,
); // object(Pdo\Sqlite)

$connection->createFunction(
   'prepend_php',
   static fn ($string) => "PHP {$string}",
); // Не существует на не соответствующем драйвере.

$connection->query('SELECT prepend_php(version) FROM php');
```

### 8. PHP RFC: new MyClass()->method() without parentheses (new MyClass()->method() без скобок)

Введение автора RFC: Функция «доступа к членам класса при создании экземпляра» была введена в PHP 5.4.0, позволяя получать доступ к константам, свойствам и методам нового экземпляра без промежуточной переменной, если новое выражение заключено в скобки.
Цель этого RFC: сделать кодирование на PHP более удобным и удовлетворить запросы многих пользователей, уменьшить визуальный долг во всех видах конструкторов и конфигураторов (более полумиллиона строк открытого исходного кода PHP можно было бы упростить) упростить переключение между другими языками типа C, не требующими скобок (Java, C#, TypeScript)
К свойствам и методам только что инициализированного объекта теперь можно обращаться, не оборачивая выражение new в круглые скобки.

```php
class PhpVersion
{
   public function getVersion(): string
   {
       return 'PHP 8.3';
   }
}

var_dump((new PhpVersion())->getVersion());
```

```php
class PhpVersion
{
   public function getVersion(): string
   {
       return 'PHP 8.4';
   }
}

var_dump(new PhpVersion()->getVersion());
```

### PHP 8.4 по умолчанию поставляется с увеличенной стоимостью bcrypt (PHP RFC: Increasing the default BCrypt cost)

Введение автора RFC: Значение PHP по умолчанию BCrypt для password_hash не изменилась с момента добавления API хеширования паролей в PHP 5.5, это было 11 лет назад. Коэффициент стоимости переменной Bcrypt предназначен для обеспечения адаптивной защиты от увеличения вычислительной мощности и, следовательно, увеличения скорости взлома. Кажется целесообразным пересмотреть значение по умолчанию по прошествии 11 лет.
В PHP 8.4 стоимость bcrypt по умолчанию увеличена до 12. Чем выше значение или "стоимость", тем сильнее защита. bcrypt — это адаптивная функция: со временем количество итераций может быть увеличено, чтобы сделать функцию медленнее и защищённее к атакам перебора даже при увеличении вычислительной мощности.
Это важно, потому что увеличение стоимости bcrypt делает хеширование паролей медленнее.

### Ленивые объекты и неявные nullable типы устарели

PHP 8.4 добавляет нативную поддержку ленивых объектов, общий паттерн, используемый фреймворками для создания прокси-объектов.

```php
$initializer = static function (MyClass $proxy): MyClass {
   return new MyClass(123);
};
$reflector = new ReflectionClass(MyClass::class);
$object = $reflector->newLazyProxy($initializer);
```


У PHP было странное поведение, когда типизированная переменная со значением по умолчанию null автоматически становилась nullable:

```php
// Устарело: Неявная маркировка параметра $book как nullable устарела,
function save(Book $book = null) {}
// должен использоваться явный nullable тип
function save(?Book $book = null) {}
```
