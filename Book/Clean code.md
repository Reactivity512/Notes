# Чистый код: создание, анализ и рефакторинг.

Книга автором которой является Роберт Мартин, также известный как "Дядюшка Боб" (Uncle Bob)

### Зачем Роберт Мартин написал книгу чистый код?

Была цель помочь разработчикам писать **понятный, поддерживаемый и качественный код.**

Вот несколько ключевых причин написании книги:
* **Борьба с плохим кодом.** Он видел, как плохой код разрушает проекты: замедляет разработку, делает баги неуловимыми и демотивирует команды. Он хотел показать, что **хороший код — это необходимость**.
* **Обучение профессионализму.** У него десятилетия опыта в разработке ПО и он хочет **поделиться проверенными практиками**. Книга учит ответственности за свой код перед коллегами и будущими разработчиками. По его мнению, программист — это не просто человек, пишущий код, а инженер, который должен осознанно создавать поддерживаемые решения.
* **Популяризация культуры чистоты.** Книга — часть его миссии по повышению стандартов в разработке ПО. Он стремится задать нормы для понятного и чистого кода, чтобы команды могли говорить на одном языке. Для него код — это не просто инструмент, а отражение профессионализма.
* Он хочет, чтобы программисты перестали быть "писателями кода" и стали "инженерами-профессионалами".

"Чистый код — это не роскошь. Это необходимость. Даже если сроки горят, именно чистый код сэкономит ваше время в будущем."

> Умение писать чистый код — тяжелая работа. Р. Мартин.

## Содержательные имена

Плохие имена усложняют чтение кода и увеличивают вероятность ошибок.

### Имя должно раскрывать намерение

```java
❌ int d; // Имя d не передает ровным счетом ничего. Оно не ассоциируется ни с временными интервалами, ни с днями.
✅ int daysSinceLastLogin;
```

### Избегай сокращений

```java
❌ genRpt();
✅ generateReport();
```

### Используйте удобопроизносимые имена

```java
❌ String ymdd;
✅ String creationDate;
```

### Избегайте обманчивых имен

```java
Account[] accountList = new Account[10];
❌ accountList // Название говорит "List", но это массив (array)
✅ accounts
```

> Слово *variable* никогда не должно встречаться в именах переменных. Слово *table* никогда не должно встречаться в именах таблиц.

### Имена функций должны содержать глагол

```java
❌ email()
✅ sendEmail()
```

### Классы — существительные

```java
❌ ManageUser // звучит как функция
✅ UserManager // звучит как объект, управляющий пользователями
```

### Имена булевых переменных должны звучать как вопрос с да/нет ответом

```java
✅ isLoggedIn = true или hasPermission = false;
```

### Не добавляйте избыточный контекст в имена

```java

❌ class Address {
    String addressStreet;
    String addressCity;
    String addressZipCode;
}

✅ class Address {
    String street;
    String city;
    String zipCode;
}
```

Здесь слово **address** избыточно, потому что уже понятно, что это **Address**.
Это:

* Уменьшается шум — читается быстрее и проще.
* Повышается ясность — имя короче, но не теряет смысла.
* Снижается дублирование — легче менять и сопровождать код.

Если переменная объявлена вне ясной области (например, в большом методе или в глобальном контексте), иногда небольшой контекст оправдан. **Главное — не дублировать уже очевидное**.

### Имена переменных и функций должны нести достаточно контекста

```java
❌ public void save(Map<String, String> data) {
    if (!data.containsKey("email") || data.get("email").isEmpty() ||
        !data.containsKey("name") || data.get("name").isEmpty()) {
        throw new IllegalArgumentException("Email and name are required");
    }
    
    this.db.create(data);
}
```

* Название функции `save()` слишком общее — что именно сохраняется?
* Параметр `data` неинформативен — какие поля ожидаются?
* Что именно создает `create()`, создание записи происходит в какой таблице users, orders, products

```java
✅ public void saveUserToDatabase(UserDto userData) {
    if (userData.getEmail() == null || userData.getEmail().isEmpty() || 
        userData.getName() == null || userData.getName().isEmpty()) {
        throw new IllegalArgumentException("Email and name are required");
    }
    
    this.userRepository.save(userData);
}
```

* Имя функции (`saveUserToDatabase`) сразу говорит:
    * Что делает? Сохраняет (save).
    * Что сохраняет? Пользователя (User).
    * Куда? В базу данных (ToDatabase).
* Параметр `userData` вместо `data` — ясно, что это данные пользователя.
* Использование репозитория (userRepository) — явное указание слоя работы с БД.

“Избегайте остроумия” и “воздержитесь от каламбуров” при именовании. Каламбуры, шутки, метафоры, отсылки — не универсальны. То, что для одного разработчика «смешно» или «умно», для другого будет непонятно и раздражающе.

### Выбирайте одно слово для каждой концепции

```java
❌ fetchUser(); getUser(); retrieveUser();
```
Все три делают одно и то же. Использование разных слов (fetch, get, retrieve) создаёт ложное различие.

```java
✅ getUser(); getOrder(); getInvoice();
```
Тут все ясно, что все методы одного типа — они получают данные.

**Выбирайте имена, удобные для поиска.**

```java
❌ public static final int M = 7;

✅ public static final int MAX_CLASSES_PER_STUDENT = 7;
```

> Не бойтесь расходовать время на выбор имени. Опробуйте несколько разных имен и посмотрите, как читается код с каждым из вариантов. Поиски хороших имен приводят к полезной реструктуризации кода. (Р. Мартин.)

## Функции

### Блоки и отступы:
* Отступы делают структуру кода понятной.
* Используется 4 пробела на один уровень вложенности (в Java (Oracle Java Code Conventions) и PHP (стандарт PSR-12: Extended Coding Style Guide) используется  по соглашению).
* Логические блоки (например, в if, for, try) отделяются пустой строкой, если это улучшает читаемость.

```java
❌ Плохой стиль:
public void login(String user, String pass){
    if(user == null || pass == null){throw new IllegalArgumentException();}
    authService.authenticate(user, pass);}
```

```java
✅ Хороший стиль:
public void login(String user, String pass) {
    if (user == null || pass == null) {
        throw new IllegalArgumentException();
    }

    authService.authenticate(user, pass);
}
```

> Первое правило: функции должны быть компактными. Второе правило: функции должны быть еще компактнее. (Р. Мартин.)

```java
❌ 
public void processUserData(User user) {
    if (user != null) {
        // Валидация
        if (user.getName() == null || user.getEmail() == null) {
            throw new IllegalArgumentException("Invalid data");
        }
        // Сохранение
        database.save(user);
        // Отправка уведомления
        emailService.sendWelcomeEmail(user.getEmail());
    }
}
```

```java
✅ 
public void processUserData(User user) {
    validate(user);
    save(user);
    notify(user);
}
private void notify(User user) {
    emailService.sendWelcomeEmail(user.getEmail());
}
…
```

* Функция должна выполнять только одну операцию, выполнять ее хорошо, и ничего другого она делать не должна.
* Короткие функции легче читать, тестировать и переиспользовать.
* Идеальная длина — от 3 до 15 строк (в зависимости от задачи).

### Все строки внутри функции должны находиться на одном уровне абстракции
Функция не должна одновременно оперировать высокоуровневыми действиями (например, "обработать заказ") и низкоуровневыми деталями (например, "открыть соединение с базой").

```java
❌
public void processOrder(Order order) throws SQLException, EmailException {
    // Высокоуровневая логика
    if (order.isValid()) {
        // Низкоуровневая логика работы с БД
        try (Connection conn = DriverManager.getConnection(DB_URL, DB_USER, DB_PASS);
            PreparedStatement stmt = conn.prepareStatement(
                 "INSERT INTO orders (id, customer_email, amount) VALUES (?, ?, ?)")) {
            stmt.setInt(1, order.getId());
            stmt.setString(2, order.getEmail());
            stmt.setBigDecimal(3, order.getAmount());
            stmt.executeUpdate();
        }
        
        // Высокоуровневая логика отправки email
        sendConfirmationEmail(order.getEmail());
    }
}
```

```java
✅
public void processOrder(Order order) throws SQLException, EmailException {
    if (!order.isValid()) {
        return;
    }

    saveOrder(order);
    sendConfirmationEmail(order.getEmail());
}
private void saveOrder(Order order) throws SQLException {
    try (Connection conn = DriverManager.getConnection(DB_URL, DB_USER, DB_PASSWORD);
        PreparedStatement stmt = conn.prepareStatement("INSERT INTO orders (id, customer_email, amount) VALUES (?, ?, ?)")) {
            
        stmt.setInt(1, order.getId());
        stmt.setString(2, order.getEmail());
        stmt.setBigDecimal(3, order.getAmount());
        stmt.executeUpdate();
    }
}
...
```

`processOrder` оперирует **бизнес-логикой**
`saveOrder` оперирует **технической реализацией**

### switch

Использование `switch` в контексте чистого кода имеет свои важные правила и антипаттерны. `switch` — это признак потенциальной проблемной архитектуры, особенно при нарушении принципа открытости/закрытости.

Что не так с ``switch``
* ❌ Ломает принцип OCP (Open/Closed Principle) - Каждый раз, когда нужно добавить новый case, приходится модифицировать существующий код. Это нарушает идею, что код должен быть открыт для расширения, но закрыт для изменения.
* ❌ Часто сигнализирует о плохом дизайне - Если ты видишь `switch` по типу объекта — это *запах кода*. Значит, возможно, пора использовать полиморфизм.

*Запах кода* - Это структурный симптом, указывающий, что в коде может быть скрытая проблема — например, плохая читаемость, сложность поддержки, дублирование или нарушение принципов SOLID. "Запахи" не всегда означают, что код работает неправильно — он может быть функциональным, но некачественным.

* ✅  Альтернатива: Полиморфизм вместо `switch`
* ✅ Но иногда `switch` оправдан (например, в роутинге, трансформациях, CLI-командах)

### Чем меньше аргументов в функции — тем лучше
Идеально: 0–2 аргумента, максимум — 3. Более 3 аргументов — запах кода, который почти всегда можно упростить.

### Избегайте булевых флагов
Булевый аргумент (`true/false`) означает, что функция делает две вещи, а не одну.

```java
❌
public String render(Map<String, Object> data, boolean isAdmin) { ... }

✅
public String renderForAdmin(Map<String, Object> data) { ... }
public String renderForUser(Map<String, Object> data) { ... }
```

### Группируйте связанные аргументы в объекты
Если в функцию передается много аргументов, попробуйте объединить их в объект или ассоциативный массив.

```java
❌
public User createUser(String name, String email, String password, boolean isAdmin, String status) {
    // ... валидация и логика создания
}

✅
public User createUser(UserData userData) {
    // ... валидация и логика создания
}
или
public User createUser(Map<String, Object> userData) {
    // ... валидация и логика создания
}
```

### Используй именованные параметры

```java
❌ sendEmail('user@example.com', 'Welcome', true);
```

```java
// На kotlin
✅ sendEmail(
    to = "user@example.com",
    subject = "Welcome",
    isHtml = true
)

// На java
✅ Email email = new Email.Builder()
    .to("user@example.com")
    .subject("Welcome")
    .isHtml(true)
    .build();

sendEmail(email);
```

### Избавьтесь от побочных эффектов
Побочный эффект — это любое изменение состояния вне текущей функции, которое не отражено напрямую в возвращаемом значении.

```java
❌
public class CounterExample {
    private static int counter = 0;

    public static void incrementCounter() {
        counter++; // Побочный эффект: изменяет внешнюю переменную
    }

    public static void main(String[] args) {
        incrementCounter();
        System.out.println(counter); // 1
    }
}
```

*Проблема:*
Функция `incrementCounter()` изменяет внешнее состояние (counter), что усложняет тестирование и может привести к неожиданным ошибкам в многопоточной среде.

```java
✅
public class CounterExample {
    public static int increment(int value) {
        // Чистая функция: не изменяет внешнее состояние, только возвращает результат
        return value + 1;
    }

    public static void main(String[] args) {
        int counter = 0;
        counter = increment(counter); // Передаем значение и получаем новое
        System.out.println(counter); // 1
    }
}
```
### Плохо если ваша функция обещает делать что-то одно, но делает что-то другое, скрытое от пользователя

```java
❌
public class UserValidator {
   private Cryptographer cryptographer;

   public boolean checkPassword(String userName, String password)
    {
        User user = UserGateway.findByName(userName);
        if (user != User.NULL) {
             String codedPhrase = user.getPhraseEncodedByPassword();
             String phrase = cryptographer.decrypt(codedPhrase, password);
             if ("Valid Password".equals(phrase)) {
                Session.initialize(); // Побочный эффект
                return true;
             }
        }
        return false;
    }
}
```

Побочным эффектом является вызов `Session.initialize();`
Имя `checkPassword` сообщает, что функция проверяет пароль. Оно ничего не говорит о том, что функция инициализирует сеанс.
В таких случаях функцию лучше назвать `checkPasswordAndInitializeSession`, что хотя бы расскажет о действие, которое было скрыто от пользователя.

### Разделение команд и запросов (Command–Query Separation. CQS)
Функция должна что-то делать или отвечать на какой-то вопрос, но не одновременно. Либо функция изменяет состояние объекта, либо возвращает информацию об этом объекте.
**Команда (Command)** — изменяет состояние, но ничего не возвращает.
**Запрос (Query)** — возвращает данные, но не изменяет состояние.

```java
❌
public Invoice createInvoice(InvoiceData data) {
    int invoiceId = saveInvoiceToDb(data); // Команда (изменяет состояние)
    return getInvoiceById(invoiceId);      // Запрос (возвращает данные)
}
```

```java
✅
// Команда (сохраняет инвойс, ничего не возвращает)
public void saveInvoice(InvoiceData data) {
    saveInvoiceToDb(data);
}

// Запрос (получает инвойс по ID, не изменяет состояние)
public Invoice getInvoiceById(int invoiceId) {
    return fetchInvoiceFromDb(invoiceId);
}

// Использование:
InvoiceData data = new InvoiceData(...);
saveInvoice(data);                     // Команда (сохраняем)
Invoice invoice = getInvoiceById(1);   // Запрос (получаем)

или

// Команда возвращает ID, но это не строгий CQS
public int createInvoiceAndReturnId(InvoiceData data) {
    return saveInvoiceToDb(data);
}

// Использование:
int newInvoiceId = createInvoiceAndReturnId(data);
Invoice invoice = getInvoiceById(newInvoiceId);

или

// Функциональный стиль (если нужна иммутабельность)

// Возвращает новый объект вместо изменения БД (если работаем с памятью)
public Invoice createInvoice(InvoiceData data) {
    return new Invoice(data); // Например, для тестов или in-memory реализации
}
```

### Используйте исключения вместо возвращения кодов ошибок
Возвращение кодов ошибок (например, false, -1, null) — это устаревший подход, который усложняет обработку ошибок и делает код менее читаемым. Вместо этого следует использовать исключения.

```java
❌ 
public class PaymentProcessor {
    public static final int ERROR_INVALID_AMOUNT = 1;
    public static final int ERROR_INVALID_METHOD = 2;
    public static final int ERROR_PROCESSING_FAILED = 3;

    public static Integer processPayment(double amount, PaymentMethod paymentMethod) {
        if (amount <= 0) {
            return ERROR_INVALID_AMOUNT;
        }

        if (!paymentMethod.isValid()) {
            return ERROR_INVALID_METHOD;
        }

        if (!paymentMethod.process(amount)) {
            return ERROR_PROCESSING_FAILED;
        }

        return null;
    }
}

int result = PaymentProcessor.processPayment(-10, paymentMethod);

if(result == PaymentProcessor.ERROR_INVALID_AMOUNT) {
    System.out.println("Ошибка: неверная сумма.");
} else if (result == PaymentProcessor.ERROR_INVALID_METHOD){
    System.out.println("Ошибка: неверный способ оплаты.");
} else if (result == PaymentProcessor.ERROR_PROCESSING_FAILED){
    System.out.println("Ошибка: не удалось провести платеж.");
}

System.out.println("Платеж успешно обработан.");
```

**Проблемы:**
1. Путаница в типах – функция может вернуть int или null.
2. Неочевидная обработка – нужно вручную проверять if (result == 1).
3. Легко пропустить ошибку – если забыть проверить, код продолжит работу с некорректными данными.

```java
✅
class PaymentException extends Exception {
    public PaymentException(String message) {
        super(message);
    }
}

public class PaymentProcessor {
    public void processPayment(double amount, PaymentMethod paymentMethod) throws PaymentException {
        
        if (amount <= 0) {
            throw new PaymentException("Неверная сумма платежа");
        }

        if (!paymentMethod.isValid()) {
            throw new PaymentException("Неверный способ оплаты");
        }

        if (!paymentMethod.process(amount)) {
            throw new PaymentException("Ошибка при обработке платежа");
        }
    }
}

public class Main {
    public static void main(String[] args) {
        PaymentProcessor processor = new PaymentProcessor();
        PaymentMethod method = new CreditCardPayment();
        
        try {
            processor.processPayment(-10, method);
            System.out.println("Платеж успешно обработан");
        } catch (InvalidAmountException | InvalidPaymentMethodException | PaymentProcessingException e) {
            System.out.println("Ошибка: " + e.getMessage());
        }
    }
}
```

### Принцип DRY (Don't Repeat Yourself)
Принцип DRY подразумевает, что мы не должны повторять одни и те же фрагменты кода, чтобы повысить его читаемость, поддержку и избежать ошибок.

```java
❌ // Плохо: одна и та же проверка в двух местах

// Использование:
User user = new User();

if (user.isAdmin() || user.isModerator()) {
    // Дать доступ
}

// В другом месте кода
if (user.isAdmin() || user.isModerator()) {
    // Разрешить действие
}
```

```java

✅
class User {    
    public boolean canManageContent() {
        return this.isAdmin() || this.isModerator();
    }
}

// Теперь везде используем этот метод:
if (user.canManageContent()) {
   // Дать доступ
}

```

**Плюсы:** 
* Устранение дублирования - проверка в одном месте
* Легче поддерживать - при изменении правил нужно править только один метод
* Более читаемо - название метода ясно выражает намерение
* Гибкость - можно легко изменить логику проверки

Про важность принципов DRY также говорил известный разработчик и автор книги "The Art of Unit Testing" Джеффри Фридл (Jeffrey Friedl).
> "Повторяющийся код — это мина замедленного действия. Рано или поздно вы забудете обновить его в одном из мест, и это приведет к багам, которые будет сложно найти и исправить."  (Jeffrey Friedl)

Соблюдение DRY:
* Дублирование логики → Выносите в методы/классы.
* Дублирование данных → Используйте константы/конфиги.
* Дублирование запросов → Применяйте репозитории/ORM.
* Дублирование валидации → Создавайте валидаторы/DTO.

DRY — это не только про код, но и про данные, конфиги и даже документацию.

## Комментарии

Комментарий — признак неудачи. Хороший код должен быть самодокументированным, то есть его логика и поведение должны быть настолько понятными и очевидными, что не требуется дополнительных комментариев. Вместо того чтобы помогать понять код, комментарии могут путать, усложнять его восприятие и маскировать настоящие проблемы.
> «Не комментируйте плохой код — перепишите его.» — Роберт Мартин

### Избыточные комментарии
Комментарий дублирует код, но не добавляет смысла.

```java
❌
// Создаем новый объект пользователя
user = new User();

или

long price = product.getPrice() + tax; // Добавляем налог к цене
```

Почему это плохо?
* Код сам по себе ясен. Комментарий не добавляет никакой дополнительной информации.
* Такой комментарий только увеличивает размер файла без пользы.

### Закомментированный код
Мёртвый код, который «на всякий случай» оставили в кодовой базе.

```java
❌
// public double oldCalculateTax(double price) {
//    return price * 0.18; // Старая версия
// }
```

Почему это плохо?
* Захламляет код.
* Если он вдруг понадобится, его можно найти в истории Git.

### Комментарии-оправдания
Разработчик пытается объяснить, почему код плохой, вместо того чтобы его исправить.

```java
❌
// Пришлось сделать так из-за бага в API (TODO: починить, когда API обновят)
public static JSONObject getData() {
   String data = getFileContents("http://buggy-api.com/data");

   return jsonDecode(data . "}"); // Костыль: API возвращает битый JSON
}
```

Почему это плохо?
* TODO редко исправляют.
* Лучше сразу написать обходное решение или добавить валидацию.

### Слишком детальные комментарии
Объяснение очевидных вещей.

```java
❌
// Увеличиваем счетчик на 1
counter++;
```

Почему это плохо?
* Очевидно из кода, что мы увеличиваем переменную `counter` на 1.
* Такой комментарий не добавляет ничего нового и просто делает код длиннее, не помогая читателю.

### Комментарии, которые не объясняют "почему", а только описывают "что".

```java
❌
// Устанавливаем значение по умолчанию
int timeout = 30;
```

Почему это плохо?
* Этот комментарий не объясняет, почему устанавливается именно 30, что это означает или почему именно такой подход выбран.

### Комментарий, который не соответствует действительности
Этот вид комментариев бывает очень опасным, потому что приводит к путанице и ошибочному пониманию кода. Если комментарий неправдоподобен или устарел, это может вызвать серьезные проблемы в будущем.

```java
❌
// Возвращает список пользователей
public List<User> getUserList() {

    List<User> users = userRepository.findAll();

    ...

    updateUsersStatus(users); // Не только возвращает список пользователей, но и обновляет их статус

    return users;
}
```

Почему это плохо?
* Функция не делает то, что написано в комментарии. Это запутывает читателя, так как комментарий не отражает реальное поведение кода.
* Такая ошибка может привести к недопониманию кода или даже неправильной его эксплуатации.

### Комментарий, который просто говорит "Тут код, не меняйте"

```java
❌
// Не трогайте этот код
```

Почему это плохо?
* Этот комментарий не помогает понять, что делает код.
* Подразумевается, что код работает, но это не объясняет, почему он работает или почему его не стоит изменять.
* На самом деле, такие комментарии могут скрывать проблемы в коде, которые нуждаются в улучшении или изменении.


Если ты начинаешь писать комментарии, это может быть сигналом, что код нужно рефакторить и улучшить его читаемость. Но комментарии не всегда зло. Хороший комментарий должен быть полезным, лаконичным и давать информацию, которую нельзя выразить или понять из самого кода

### Объяснение намерений
Комментарий должен отвечать на вопрос «почему так?», а не «как это работает?».

```java
✅
// Используем бинарный поиск вместо линейного, потому что массив отсортирован и большой.
public Optional<User> findUserById(List<User> users, int id) {
   // ... бинарный поиск ...
}
```

В примере объясняет выбор алгоритма, который не очевиден из кода.

### Документация публичного API
(PHPDoc для методов). Помогает IDE показывать подсказки. Объясняет граничные условия (например исключения).

```java
✅
/**
 * @param dates   список дат для фильтрации (не может быть {@code null})
 * @param records список записей, которые нужно обработать (не может быть {@code null})
 * @param types   допустимые типы записей (если {@code null}, используются все типы)
 * @return {@code Map<String, List<Record>>} — карта, где ключ — это дата в формате "yyyy-MM-dd",
 *         а значение — список отфильтрованных записей для этой даты
 * @throws IllegalArgumentException если {@code dates} или {@code records} равны {@code null}
 */
```

### Предупреждение о неочевидных последствиях

```java
✅
// Внимание: кешируем на 1 час, потому что API лимитирует запросы.
Data data = cache.get("api_data", key -> fetchApiData(), 3600);
```

### Юридические комментарии

```java
✅
// Copyright (C) 2003,2004,2005 by Object Mentor, Inc. All rights reserved.
// Публикуется на условиях лицензии GNU General Public License версии 2 и выше.
```

## Форматирование

### Вертикальное форматирование
100–200 строк кода на файл, максимум: 500 строк (для сложных случаев).

```java
❌ // Все слитно
public class Calculator { public int add(int a, int b) { return a + b; }}
```

```java
✅ // Логические блоки разделены
public class Calculator {
     public int add(int a, int b) {
           return a + b;
     }
}
```

### Горизонтальное форматирование
80–120 символов.

```java
❌ // Неровные отступы
public void saveUser(String name,
int age,
boolean isActive) {...}
```

```java
✅ // Хорошо
public void saveUser(
     String name,
     int age,
     boolean isActive
) {...}
```

### Группировка кода, расположение методов в классе:

```java
❌ // Плохо:
class Order {
       public function setOrderId($orderId) {...}
       public function getItems() {...}
       public function __construct($orderId) {...}
       protected function validateOrder() {...}
       public function getTotal() {...}
       private function calculateDiscount() {...}
}
```

Правильное расположения методов в классе:
1. Конструкторы и деструкторы
2. Публичные методы
3. Приватные методы
4. Защищённые методы
5. Геттеры и сеттеры

```java
✅ // Хорошо:
class Order {
       public function __construct($orderId) {...}   // 1. Конструктор
       public function getTotal() {...}              // 2. Публичный метод
       private function calculateDiscount() {...}    // 3. Приватный метод
       protected function validateOrder() {...}      // 4. Защищенный метод
       public function setOrderId($orderId) {...}    // 5. Сеттер
       public function getOrderId() {...}            // 5. Геттер
}
```

Именование переменных и методов:
* Для переменных и методов используется **camelCase**.
* Для классов используется **PascalCase**.
* Для констант используется **UPPER_SNAKE_CASE**.

## Модульные тесты

### Три закона TTD

**Первый закон.** Не пишите код продукта, пока не напишете отказной модульный тест.
**Второй закон.** Не пишите модульный тест в объеме большем, чем необходимо для отказа. Невозможность компиляции является отказом.
**Третий закон.** Не пишите код продукта в объем большем, чем необходимо для прохождения текущего отказного теста.

> «Не пишите ни строчки кода, пока не напишете тест, который проверяет нужную функциональность и падает.» *Роберт Мартин*


**TDD (Test-Driven Development)** — это методология разработки программного обеспечения, при которой тесты пишутся до реализации функционала.
1. **Пишем тест** — создаем тест, который будет изначально неудачным. Этот закон заставляет разработчика четко осознавать, что именно должен делать код, прежде чем он начнется его писать. Программирование по тестам помогает избежать неопределенности в требованиях и дает четкие критерии успеха для каждой маленькой части системы.
2. **Пишем минимальный код** — реализуем минимальное решение, чтобы тест прошел. Код должен быть минимальным, т.е. делать ровно то, что требуется для прохождения теста. Важно не добавлять функциональности или фич, которые пока не нужны. Это предотвращает написание лишнего кода, который усложняет проект и увеличивает количество ошибок.
3. **Рефакторим** — улучшаем код, соблюдая его работоспособность, при этом тесты помогают удостовериться, что рефакторинг не нарушил поведение. На этом этапе добавляется дополнительная логика, можно улучшить названия переменных, улучшить производительность и т.д. благодаря существующим тестам, вы можете быть уверены, что рефакторинг не нарушит работоспособность системы.

Почему важны эти законы?
* **Документация через тесты**— код сразу покрыт проверками.
* **Простота дизайна** — вы пишете только то, что нужно.
* **Безопасные изменения** — рефакторинг без страха сломать функциональность.

> «TDD — это не про тестирование, это про проектирование» — Кент Бек (Автор книги “Экстремальное программирование: разработка через тестирование”)

**Каждый тест должен проверять только одну вещь.** Такой подход способствует созданию чистых и поддерживаемых тестов, которые можно легко изменить или адаптировать при необходимости. Это делает тесты:
* Более читаемыми (ясно, что тестируется).
* Более стабильными (если падает, сразу понятно, что сломалось).
* Лучше изолированными (нет зависимостей между проверками).

**Тестовый код не менее важен, чем код продукта.**

> Тест — это не просто проверка, это документация - *Роберт Мартин*

```java
❌ // Плохо: Множественные проверки в одном тесте
import org.junit.jupiter.api.Test;
import java.time.LocalDateTime;
import static org.junit.jupiter.api.Assertions.*;

class UserRegistrationTest {

    @Test
    void testUserCreation() {
        User user = new User("Alex");
        
        // Проверка 1: Имя
        assertEquals("Alex", user.getName(), "Имя пользователя должно быть 'Alex'");
        
        // Проверка 2: Статус (по умолчанию false)
        assertFalse(user.isActive(), "Новый пользователь должен быть неактивным");
        
        // Проверка 3: Дата создания (не null и в прошлом)
        LocalDateTime createdAt = user.getCreatedAt();
        assertNotNull(createdAt, "Дата создания не должна быть null");
        assertTrue(createdAt.isBefore(LocalDateTime.now()), 
                 "Дата создания должна быть в прошлом");
    }
}
```

Проблемы:
* Если первая проверка упадёт, остальные не выполнятся.
* Сложно понять, что именно сломалось.
* Нарушает принцип «один тест — одна ответственность».

```java
✅ // Хорошо: Разделение на отдельные тесты

class UserTest {

    // Тест 1: Проверка имени пользователя
    @Test
    void userNameShouldMatchInitialValue() {
        // Arrange
        String expectedName = "Alex";
        User user = new User(expectedName);
        
        // Act & Assert
        assertEquals(expectedName, user.getName(), 
            "Имя пользователя должно соответствовать переданному в конструкторе");
    }

    // Тест 2: Проверка статуса нового пользователя
    @Test
    void newUserShouldBeInactive() {
        // Arrange
        User user = new User("Alex");
        
        // Act & Assert
        assertFalse(user.isActive(), 
            "Новый пользователь должен быть неактивным по умолчанию");
    }

    // Тест 3: Проверка даты создания
    @Test
    void userShouldHaveCreationDate() {
        // Arrange
        User user = new User("Alex");
        
        // Act
        LocalDateTime createdAt = user.getCreatedAt();
        
        // Assert
        assertNotNull(createdAt, "Дата создания не должна быть null");
        assertTrue(activatedAt.isBefore(LocalDateTime.now()) || 
                 activatedAt.isEqual(LocalDateTime.now()),
                 "Дата активации должна быть текущей или прошлой");
    }
}
```

Преимущества:
* Ясность: Каждый тест проверяет одну конкретную вещь.
* Стабильность: Если упадёт один тест, остальные выполнятся.
* Лёгкость отладки: По имени теста сразу понятно, что сломалось.
* Структура AAA: Чёткое разделение на блоки Arrange-Act-Assert.

Есть случаи когда можно нарушить это правило. Например когда несколько проверок логически связаны и имеют общий контекст.

```java
✅
class UserActivationTest {

    @Test
    void activateUser_ShouldSetActiveStatusAndActivationDate() {
        // Arrange
        User user = new User("Alex");
        
        // Act
        user.activate();
        
        // Assert
        // Проверка 1: Статус
        assertTrue(user.isActive(), 
                 "Пользователь должен быть активным после активации");
        
        // Проверка 2: Дата активации
        LocalDateTime activatedAt = user.getActivatedAt();
        assertNotNull(activatedAt, 
                     "Дата активации не должна быть null");
        assertTrue(activatedAt.isBefore(LocalDateTime.now()) || 
                 activatedAt.isEqual(LocalDateTime.now()),
                 "Дата активации должна быть текущей или прошлой");
    }
}
```

Почему это допустимо:
* Обе проверки относятся к одному сценарию (активация пользователя).
* Разделение не даст возможности протестировать сценарий активации пользователя полностью.

### F.I.R.S.T.

Чистые тесты должны обладать еще пятью характеристиками, названия которых образуют сокращение F.I.R.S.T.

* **Fast (Быстрота).** Тесты должны выполняться быстро. Тесты не должны требовать тяжелой работы с базой данных или сети, если это не требуется для проверки функциональности. Вместо этого можно использовать заглушки (моки), стабы или фейки для имитации работы этих сервисов. Тесты, которые выполняются слишком долго, теряют свою ценность. Автоматические тесты должны быть настолько быстрыми, чтобы их можно было запускать часто.
* **Independent (Независимость).** Тесты не должны зависеть друг от друга или от глобального состояния. Каждый тест должен создавать свои собственные объекты и данные, чтобы не зависеть от предыдущих тестов. Если один тест создает пользователя в базе данных, он должен очистить эти данные после завершения, чтобы следующий тест не столкнулся с проблемой отсутствия данных. Каждый тест должен создавать своё изолированное окружение.
* **Repeatable (Повторяемость).** Тесты должны давать одинаковый результат при любых условиях. Тесты не должны зависеть от внешних факторов, таких как время, дата, случайные значения или состояние системы, которое может измениться. Для теста, который использует случайные числа, можно использовать фиксированные значения или заглушки для того, чтобы результат был предсказуемым и воспроизводимым.
* **Self-Validating (Очевидность).** Результатом выполнения теста должен быть логический признак. Тест должен однозначно показывать, прошёл он или упал, без ручных проверок.
* **Timely (Своевременность).** Тесты должны создаваться своевременно. Тесты пишутся до или во время разработки кода, а не после. Тесты, написанные поздно, могут быть неэффективными или даже не покрывать важные сценарии. Чем раньше вы пишете тесты, тем больше их ценность для процесса разработки.

Почему важно соблюдать F.I.R.S.T.?

Применение этих принципов помогает создать **качественные тесты**, которые можно использовать в долгосрочной перспективе.
Преимущества:
* **Увеличение надежности:** Тесты, соответствующие этим принципам, минимизируют вероятность ошибок и помогут быстрее находить дефекты.
* **Лучшая поддерживаемость:** Легко добавлять, изменять и удалять тесты, так как они независимы и просты в понимании.
* **Ускорение разработки:** Быстрые и надежные тесты дают возможность разработчикам быстро получать обратную связь о своей работе.

> «Сначала напишите тест, который падает, затем заставьте его работать, а после улучшите код». - Кент Бек

## Классы

### SOLID

**SOLID** — это акроним, который обозначает пять принципов объектно-ориентированного проектирования, помогающих создавать чистый, поддерживаемый и масштабируемый код. Принципы были сформулированы Робертом Мартином, и впервые представлены в его статьях и книгах как **SRP**, **OCP**, **LSP**, **ISP**, **DIP**. Майкл Фэзерс — систематизировал в виде акронима SOLID, объединив принципы.
Расшифровка **SOLID**:
* **S** — Single Responsibility Principle (Принцип единственной ответственности)
* **O** — Open/Closed Principle (Принцип открытости/закрытости)
* **L** — Liskov Substitution Principle (Принцип подстановки Лисков)
* **I** — Interface Segregation Principle (Принцип разделения интерфейсов)
* **D** — Dependency Inversion Principle (Принцип инверсии зависимостей)

Каждый принцип SOLID фокусируется на том, как правильно организовать код, чтобы он был гибким, легко модифицируемым и понятным.

> "SOLID — это не правила, а принципы. Нарушать можно, но только осознанно!" — *Роберт Мартин*

### Single Responsibility Principle (SRP) — Принцип единственной ответственности

**Класс или модуль должен иметь только одну причину для изменений**, то есть он должен отвечать за одноконкретное действие или задачу. Это позволяет сделать код более модульным и удобным для тестирования.

```java
❌ // Нарушение SRP
public class User {
    private String name;
    private String email;

    public User(String name, String email) {
        this.name = name;
        this.email = email;
    }

    // Метод сохранения в БД (нарушение SRP)
    public void saveToDatabase() {
        // Логика сохранения в БД
    }

    // Метод отправки email (нарушение SRP)
    public void sendEmail(String subject, String message) {
        // Логика отправки email
    }
}
```

В данном примере класс `User` выполняет две задачи: сохранение данных в БД и отправку email. Каждая из этих задач — это отдельная ответственность, и если потребуется изменить способ отправки email, это повлияет на класс `User`.


```java
✅ // Соблюдение SRP
public class User {
    private String name;
    private String email;

    public User(String name, String email) {
        this.name = name;
        this.email = email;
    }

    public String getName() { return name; }
    public String getEmail() { return email; }
}

// Отдельный класс для работы с БД
public class UserRepository {
    public void save(User user) {
        // Логика сохранения в БД
    }
}

// Отдельный класс для отправки email
public class EmailService {
    public void sendEmail(User user, String subject, String message) {
        // Логика отправки email
    }
}
```

У каждого класса есть своя ответственность.

### Open/Closed Principle (OCP) — Принцип открытости/закрытости

**Классы должны быть открыты для расширения, но закрыты для модификации.** Это означает, что вы должны иметь возможность добавлять новую функциональность в код без изменения уже существующего кода. Когда вам нужно добавить новую функциональность, вы должны делать это через расширение существующих классов, а не через их модификацию.

```java
❌ // Нарушение OCP: нужно менять класс при добавлении новой фигуры
class AreaCalculator {
    public double calculateArea(Object shape) {
        if (shape instanceof Circle) {
            Circle circle = (Circle) shape;
            return Math.PI * circle.radius * circle.radius;
        } else if (shape instanceof Square) {
            Square square = (Square) shape;
            return square.side * square.side;
        }
        // Добавление новой фигуры (например, Triangle) потребует изменения этого класса!
        throw new IllegalArgumentException("Unknown shape");
    }
}

class Circle {
    public double radius;
    public Circle(double radius) { this.radius = radius; }
}

class Square {
    public double side;
    public Square(double side) { this.side = side; }
}
```

Проблема:
* Чтобы добавить `Triangle`, нужно изменять метод `calculateArea()` в `AreaCalculator`.


```java
✅ // Соблюдение OCP: новые фигуры добавляются без изменения AreaCalculator

// Класс для вычисления площади (не требует изменений при добавлении новых фигур)
class AreaCalculator {
    public double calculateArea(Shape shape) {
        return shape.calculateArea();  // Делегируем вычисление самой фигуре
    }
}

interface Shape {
    double calculateArea();  // Каждая фигура сама знает, как считать свою площадь
}

class Circle implements Shape {
    private double radius;
    public Circle(double radius) { this.radius = radius; }
    @Override
    public double calculateArea() {
        return Math.PI * radius * radius;
    }
}

class Square implements Shape {
    private double side;
    public Square(double side) { this.side = side; }
    @Override
    public double calculateArea() {
        return side * side;
    }
}

// Пример добавления новой фигуры (без изменения AreaCalculator)
class Triangle implements Shape {
    private double base, height;
    public Triangle(double base, double height) {
        this.base = base;
        this.height = height;
    }
    @Override
    public double calculateArea() {
        return 0.5 * base * height;
    }
}
```

Преимущества:

* `AreaCalculator` не нужно менять при добавлении `Triangle`, `Pentagon` и т.д.
* Каждая фигура инкапсулирует свою логику.
* Код легко расширяется.

OCP достигается через:
✔ Абстракции (интерфейсы/абстрактные классы).
✔ Делегирование логики самим классам.
✔ Отказ от if-else/switch в пользу полиморфизма.

### Liskov Substitution Principle (LSP) — Принцип подстановки Лисков

Подклассы должны быть заменяемыми на свои базовые классы без изменения ожидаемого поведения программы. Подклассы не должны нарушать контракт, заданный в базовом классе.

```java
❌ // Нарушение LSP:
class Bird {
    public void fly() {
        System.out.println("Я лечу!");
    }
}

class Ostrich extends Bird {
    @Override
    public void fly() {
        throw new UnsupportedOperationException("Страусы не летают!");
    }
}

public class Main {
    public static void main(String[] args) {
        Bird bird = new Ostrich();
        bird.fly(); // Ошибка во время выполнения
    }
}
```

Мы ожидаем, что `Bird` может летать, но `Ostrich`, будучи подтипом `Bird`, не поддерживает `fly()`. Это нарушает **LSP**, потому что подтип `Ostrich` не может полностью заменить суперкласс (`Bird`) без изменения поведения.

```java
✅ // Соблюдение LSP
interface Bird {
    void eat();
}

interface FlyingBird extends Bird {
    void fly();
}

class Sparrow implements FlyingBird {
    @Override
    public void eat() {
        System.out.println("Воробей ест.");
    }

    @Override
    public void fly() {
        System.out.println("Воробей летает.");
    }
}

class Ostrich implements Bird {
    @Override
    public void eat() {
        System.out.println("Страус ест.");
    }
}

public class Main {
    public static void main(String[] args) {
        Bird bird1 = new Sparrow();
        Bird bird2 = new Ostrich();

        bird1.eat();
        bird2.eat();

        // Только летающие птицы могут летать
        FlyingBird flyingBird = new Sparrow();
        flyingBird.fly();
    }
}
```

Каждый подтип корректно реализует только те интерфейсы, поведение которых он поддерживает. `Ostrich` не пытается реализовать `fly()`, потому что он не должен уметь летать.

### Interface Segregation Principle (ISP) — Принцип разделения интерфейсов

Клиенты не должны зависеть от интерфейсов, которые они не используют. Интерфейсы должны быть маленькими и специфичными для задачи. Не стоит заставлять класс реализовывать методы, которые он не использует.

```java
❌ // Нарушение ISP
interface Worker {
    void work();
    void eat();
}

class Robot implements Worker {
    @Override
    public void work() {
        // реализация работы
    }
    
    @Override
    public void eat() {
        throw new UnsupportedOperationException("Роботы не едят!");
    }
}
```

`Robot` вынужден реализовывать ненужный метод.

```java
✅ // Соблюдение ISP
interface Workable {
    void work();
}

interface Eatable {
    void eat();
}

class Robot implements Workable {
    @Override
    public void work() {
        // реализация работы робота
    }
}

class Human implements Workable, Eatable {
    @Override
    public void work() {
        // реализация работы человека
    }

    @Override
    public void eat() {
        // реализация приема пищи
    }
}
```

Классы реализуют только нужные интерфейсы.

### Dependency Inversion Principle (DIP) — Принцип инверсии зависимостей

DIP гласит, что классы системы должны зависеть от абстракций, а не от конкретных подробностей
* Модули верхнего уровня не должны зависеть от модулей нижнего уровня. Оба должны зависеть от абстракций.
* Абстракции не должны зависеть от деталей. Детали должны зависеть от абстракций.

```java
❌ // Нарушение DIP
class OrderService {
    private MySqlDatabase database;

    public OrderService() {
        this.database = new MySqlDatabase(); // Жёсткая зависимость
    }

    public void processOrder(Order order) {
        database.save(order);
    }
}

public class Main {
    public static void main(String[] args) {
        OrderService orderService = new OrderService();
        
        orderService.processOrder(new Order());
    }
}
```

Проблемы:
* Класс `OrderService` жёстко зависит от конкретной реализации *MySqlDatabase*.
* Трудно тестировать (например, нельзя подменить *MySqlDatabase* на mock-объект).
* При смене БД (на PostgreSQL, MongoDB и т. д.) придётся изменять код `OrderService`.

```java
✅ // Соблюдение DIP
class OrderService {
    private Database database;

    // Внедрение зависимости через конструктор (Dependency Injection)
    public OrderService(Database database) {
        this.database = database;
    }

    public void processOrder(Order order) {
        database.save(order);
    }
}

interface Database {
    void save(Order order);
}

class MySqlDatabase implements Database {
    @Override
    public void save(Order order) {
        System.out.println("Сохранение заказа в MySQL...");
    }
}

public class Main {
    public static void main(String[] args) {
        Database mySqlDb = new MySqlDatabase();
        OrderService orderService = new OrderService(mySqlDb); // Можно легко заменить на другую БД
        
        orderService.processOrder(new Order());
    }
}
```

Преимущества:

* Интерфейс ``Database`` — абстракция, от которой зависит `OrderService`.
* Инъекция зависимости (DI) — `OrderService` получает `Database` извне (через конструктор).
* Гибкость — можно подменить `MySqlDatabase` на `PostgresDatabase` или mock-объект без изменения `OrderService`.
* Класс `OrderService` закрыт для изменений, но открыт для расширений (**Open/Closed Principle**)

## Многопоточность

Многопоточный код сложен для тестирования из-за недетерминированности, состояния гонки (race conditions) и проблем с изоляцией потоков. Основный проблемы с многопоточным кодом:
1. **Недетерминированность** - Потоки могут выполняться в произвольном порядке, и, следовательно, сложно предсказать, как именно будет выполняться код. Например, один поток может завершить свою работу до того, как начнется выполнение другого.
2. **Состояние гонки (Race Conditions)** - Когда два потока пытаются одновременно изменить один и тот же ресурс (например, переменную или объект), не синхронизируя свои действия, это может привести к ошибкам. 
3. **Взаимоблокировки (Deadlocks)** - Это ситуация когда два или более потока бесконечно блокируют друг друга, ожидая освобождения ресурсов, которые удерживаются самими этими потоками.
4. **Взаимоблокировки (Livelock)** — Это ситуация когда потоки не блокируются полностью (как при deadlock), но вместо этого бесконечно выполняют бесполезную работу, реагируя на действия друг друга, но не продвигаясь к решению задачи.
Фактически, потоки "живы" (не заблокированы), но система не прогрессирует.
5. **Голодание (starvation)** - Один или несколько потоков могут никогда не получить доступ к ресурсам, если другие потоки постоянно захватывают эти ресурсы.
6. **Ложная синхронизация** - Иногда программисты могут использовать синхронизацию избыточно, блокируя код там, где этого не нужно, что может снижать производительность.

Ошибки в многопоточном коде часто возникают из-за случайных факторов, например, из-за того, в каком порядке потоки выполняются. Это делает такие ошибки трудными для нахождения в тестах. Например, ошибка может появиться только при определенной комбинации задержек между потоками, которые трудно воспроизвести.
Порядок, в котором потоки получают доступ к ресурсам, может быть непредсказуемым, и каждый запуск программы может привести к разным результатам. Это усложняет тестирование, потому что тесты могут давать нестабильные или непредсказуемые результаты. Программа, которая в одном случае работает корректно, может дать сбой при другом запуске, если порядок обработки потоков отличается.

### Deadlock

Условия возникновения **deadlock** (Необходимые условия по Коффману):

* Взаимное исключение (Mutual Exclusion)
* Ресурс может использоваться только одним потоком в данный момент.
* Удержание и ожидание (Hold and Wait)
* Поток удерживает один ресурс и ждёт освобождения другого.
* Отсутствие вытеснения (No Preemption)
* Ресурс нельзя принудительно забрать у потока — только добровольное освобождение.
* Кольцевое ожидание (Circular Wait)
* Потоки образуют замкнутый цикл, где каждый ждёт ресурс, удерживаемый следующим.

Пример **deadlock**:

```java
public class DeadlockExample {
    private static final Object lock1 = new Object();
    private static final Object lock2 = new Object();

    public static void main(String[] args) {
        Thread thread1 = new Thread(() -> {
            synchronized (lock1) {
                System.out.println("Thread 1: Holding lock1...");
                try { Thread.sleep(100); } catch (InterruptedException e) {}
                synchronized (lock2) {
                    System.out.println("Thread 1: Acquired lock2!");
                }
            }
        });

        Thread thread2 = new Thread(() -> {
            synchronized (lock2) {
                System.out.println("Thread 2: Holding lock2...");
                try { Thread.sleep(100); } catch (InterruptedException e) {}
                synchronized (lock1) {
                    System.out.println("Thread 2: Acquired lock1!");
                }
            }
        });

        thread1.start();
        thread2.start(); // Оба потока заблокируются навсегда!
    }
}
```

Как избежать **deadlock**?

* Уничтожение одного из условий Коффмана: например, устранить «кольцевое ожидание», запрашивая ресурсы в строгом порядке.
* Использование таймаутов (например, tryLock() в ReentrantLock).
* Алгоритмы детектирования (например, граф ожидания ресурсов).

### Livelock

Условия возникновения **livelock**:

* Активная реакция потоков
* Потоки постоянно меняют своё состояние в ответ на действия других.
* Отсутствие прогресса
* Несмотря на активность, полезная работа не выполняется.
* Кооперативная логика
* Часто возникает из-за "вежливых" алгоритмов, где потоки пытаются уступить ресурсы друг другу.

```java
public class PoliteLivelock {
    static class Person {
        private String name;
        private boolean isPolite;

        public Person(String name, boolean isPolite) {
            this.name = name;
            this.isPolite = isPolite;
        }

        public void passDoorway(Person other, Lock lock) {
            while (true) {
                if (isPolite) {
                    // Вежливый человек уступает дорогу, если другой тоже вежливый
                    if (!other.isPolite) {
                        lock.lock();
                        try {
                            System.out.println(name + ": Прохожу дверь!");
                            break; // Прошли успешно
                        } finally {
                            lock.unlock();
                        }
                    } else {
                        System.out.println(name + ": После вас!");
                        try {
                            Thread.sleep(100); // Имитация задержки
                        } catch (InterruptedException e) {
                            Thread.currentThread().interrupt();
                        }
                        continue; // Продолжаем "уступать"
                    }
                } else {
                    // Грубый человек просто проходит
                    lock.lock();
                    try {
                        System.out.println(name + ": Прохожу дверь!");
                        break;
                    } finally {
                        lock.unlock();
                    }
                }
            }
        }
    }

    public static void main(String[] args) {
        final Person alice = new Person("Алиса", true);
        final Person bob = new Person("Боб", true);
        final Lock doorLock = new ReentrantLock();

        // Запускаем два потока, которые пытаются пройти через дверь
        new Thread(() -> alice.passDoorway(bob, doorLock)).start();
        new Thread(() -> bob.passDoorway(alice, doorLock)).start();
    }
}
```

Как избежать livelock?

* Рандомизация задержек
* Добавить случайные задержки.
* Отказ от излишней вежливости
* Ограничить количество попыток уступить ресурс.
* Приоритизация потоков
* Чётко определить порядок захвата ресурсов.

