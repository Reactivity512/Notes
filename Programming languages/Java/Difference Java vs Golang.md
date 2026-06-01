# Разница ArrayList, HashMap из java и slice, map из Go

## ArrayList vs slice

### Структура данных

`Slice` в Go — это триединая структура (24 байта на 64-bit системе):

```go
type slice struct {
    array unsafe.Pointer  // указатель на массив в heap (8 байт)
    len   int              // длина (8 байт)
    cap   int              // ёмкость (8 байт)
}
```

`ArrayList` в Java — это один объект, содержащий:

```java
class ArrayList<E> {
    private transient Object[] elementData;  // ссылка на массив
    private int size;                        // длина
    // capacity = elementData.length
    // Нет отдельного поля cap — оно вычисляется из elementData.length
}
```

Если `new ArrayList<>(5)` то
* Создаётся внутренний массив `Object[5]`
* `size = 0` (сколько элементов реально добавлено)
* `capacity = 5` (максимум до следующего расширения)

### Передача в функции

```java
// Java — передача ссылки на объект (8 байт на 64-bit)
void modify(ArrayList<Integer> list) {
    list.set(0, 100);              // изменит оригинал
    list.add(200);                 // изменит оригинал
}
```

```go
// Go — копирование slice header (быстро, 24 байта)
func modify(s []int) {
    s[0] = 100          // изменит оригинал (shared array)
    s = append(s, 200)  // НЕ изменит len/cap оригинала
}
```

* В Go передача slice — дешёвая (copy 24 байт)
* Но `append` может не отразиться в вызывающей функции
* В Java передача — дешёвая (1 ссылка), и вы всегда работаете с оригиналом

### Расширение (reallocation)

Java (ArrayList):
```java
ArrayList<Integer> list = new ArrayList<>(5); // capacity=5
for(int i=0; i<5; i++) list.add(i);           // без reallocation
list.add(5);  // capacity = 5 + (5 >> 1) = 7 (increase 50%)
// Старый массив становится мусором
```

Go (slice):
```go
s := make([]int, 0, 5)  // cap=5
s = append(s, 1,2,3,4,5) // cap=5, без reallocation
s = append(s, 6)         // cap=10 (обычно *2 для <1024, потом +25%)
// Старый массив (cap 5) становится мусором (GC соберёт)
```

Алгоритмы роста:
* Go: грубо cap = cap * 2 для малых, cap = cap + cap/4 для больших
* Java: newCapacity = oldCapacity + (oldCapacity >> 1) (1.5x)

### Сброс в ноль vs сохранение ссылок

Java (при перевыделении):
```java
ArrayList<String> list = new ArrayList<>(3);
list.add("a"); list.add("b"); list.add("c");
list.add("d");  // копирует ссылки в новый массив
// Старые объекты НЕ зануляются, но старый массив GC-rod
```

Go (при перевыделении):
```go
s := make([]int, 3, 3)  // [0,0,0]
s = append(s, 1)        // новый массив, старый заполнен 0 (zeroed)
```

### Интересный нюансы:

**Go:**
```go
func smallSlice() []int {
    s := make([]int, 3, 3)   // Может остаться на стеке, если не escape
    return s                 // Escape в heap
}
```

**Java:**
* Объект `ArrayList` всегда в **heap**
* Но JVM может делать `scalar replacement` (Stack Allocation) для small objects

**Go:**
```go
// Go: нужно помнить о return
func addToSlice(s []int, val int) []int {
    return append(s, val)  // обязательно возвращаем новый slice
}

// Иначе:
func brokenAdd(s []int, val int) {
    s = append(s, val)  // BUG: не изменит оригинал
}
```

**Java:**
```java
// Java: корректно работает всегда
void addToList(ArrayList<Integer> list, int value) {
    list.add(value);  // всегда меняет оригинал
}
```

**Java, `subList()` — возвращает view (не копию):**
```java
ArrayList<Integer> list = new ArrayList<>(List.of(1,2,3,4,5));
List<Integer> sub = list.subList(1, 4);  // [2,3,4]
sub.set(0, 100);   // меняет оригинал: list = [1,100,3,4,5]
sub.add(999);      // тоже меняет list

// ОПАСНО: при структурной модификации оригинала sub становится invalid
list.add(6);       // sub теперь бросает ConcurrentModificationException
```

**`removeRange()` — защищённый метод**
```java
public class MyArrayList<T> extends ArrayList<T> {
    public void removeRangePublic(int from, int to) {
        super.removeRange(from, to);  // удаляет без создания мусора
    }
}

// Было: size=10, capacity=20
list.removeRange(2, 5);  // удалили 3 элемента

// Стало: size=7, capacity=20 (без изменений)
```
* `size` уменьшается на количество удалённых элементов.
* `capacity` (вместимость) не меняется.

Что происходит внутри:
```java
protected void removeRange(int fromIndex, int toIndex) {
    int numMoved = size - toIndex;
    System.arraycopy(elementData, toIndex, elementData, fromIndex, numMoved);
    size = fromIndex + numMoved;  // только size уменьшается
    // elementData.length (capacity) остаётся прежним
}
```

**`trimToSize()` — уменьшает capacity до size**
```java
ArrayList<String> list = new ArrayList<>(1000);
// ... добавили только 10 элементов
list.trimToSize();  // elementData = new Object[10], экономия памяти
```


**`ArrayList.get(index)` работает за **O(1)**, как и slice в Go.**
```java
// ArrayList.get() - прямой доступ по индексу
public E get(int index) {
    return (E) elementData[index];  // константное время
}

// А вот LinkedList.get() - O(n)
LinkedList<String> linked = new LinkedList<>();
linked.get(1000);  // нужно пройти 1000 узлов
```

**Synchronized вариант**

**В Go:**
```go
type SafeSlice struct {
    mu sync.RWMutex
    slice []int
}

func (s *SafeSlice) Add(val int) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.slice = append(s.slice, val)
}

func (s *SafeSlice) Get(i int) int {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.slice[i]
}
```

**Java:**

```java
// Обычный ArrayList - не thread-safe
List<String> list = new ArrayList<>();
list.add("a");  // несинхронизирован

// Синхронизированная обёртка
List<String> syncList = Collections.synchronizedList(new ArrayList<>());

// ВСЕГДА синхронизироваться при итерации
synchronized (syncList) {
    for (String s : syncList) {  // manual sync нужен
        // ...
    }
}

// ConcurrentModificationException ВОЗМОЖНА даже в synchronizedList
```

**Memory visibility (volatile vs sync/atomic)**

**Go:**
```go
type Data struct {
    mu sync.Mutex
    slice []int
}

func (d *Data) Add(val int) {
    d.mu.Lock()
    defer d.mu.Unlock()
    d.slice = append(d.slice, val)  // mutex гарантирует visibility
}

// Без mutex - data race:
func (d *Data) BrokenAdd(val int) {
    d.slice = append(d.slice, val)  // гонка данных!
}
```

**Java:**
```java
class SharedData {
    private volatile ArrayList<String> list;  // volatile для ССЫЛКИ
    
    public void update() {
        list = new ArrayList<>();  // изменение ссылки видно всем
    }
    
    public void addElement(String s) {
        // Плохо: volatile не защищает элементы!
        list.add(s);  // может быть не видно другим потокам
    }
    
    // Правильно:
    public synchronized void addSafe(String s) {
        list.add(s);  // synchronized гарантирует видимость
    }
}
```

## HashMap vs map

### Типизация

```java
// Java - типобезопасен, но через generics
HashMap<String, Integer> map = new HashMap<>();
map.put("key", 123);
```

```go
// Go - строгая типизация на уровне компилятора
m := make(map[string]int)
m["key"] = 123
```

### nil vs null

```java
// Java
HashMap<String, String> map = null;  // можно, но NullPointerException при использовании
map.put("key", "value");  // NullPointerException
// Нельзя вызывать методы на null ссылке

map = new HashMap<>();
map.put(null, "value");  // null ключ разрешён
```

```go
// Go
var m map[string]string   // nil map
m["key"] = "value"        // panic: assignment to entry in nil map

m = make(map[string]string)  // нужно инициализировать
m["key"] = "value"           // ok
// Go не разрешает nil в качестве ключа
```

### Удаление элементов
```java
// Java
map.remove("key");      // возвращает удалённое значение (или null)
```

```go
// Go
delete(m, "key")        // не возвращает значение
value, ok := m["key"]   // нужно проверять отдельно
```

### Iteration гарантии

```java
// Java
// Нет гарантии порядка (может меняться)
for (String key : map.keySet()) {
    // порядок не определён
}
```

```go
// Go
// Намеренно рандомизирован (начиная с Go 1.0)
for k, v := range m {
    // порядок случайный, чтобы разработчики не полагались на него
}
```

### Thread-safety

```java
// Java
HashMap<String, String> map = new HashMap<>();  // не thread-safe
ConcurrentHashMap<String, String> cmap = new ConcurrentHashMap<>();  // thread-safe
```

```go
// Go
m := make(map[string]string)  // не thread-safe (data race)
var mu sync.RWMutex           // нужен mutex
// при работе с map нужно lock() unlock() мьютекса
```

### Методы vs встроенные функции

```java
// Java
map.put("key", "value");
map.get("key");
map.containsKey("key");
map.size();
```

```go
// Go
m["key"] = "value"    // встроенный синтаксис
value := m["key"]     // если нет ключа - zero value
_, ok := m["key"]     // проверка существования
len(m)                // встроенная функция
```

### Внутренняя реализация

**Java**:
* **HashMap** - массив корзин с цепочками (linked list)
* Начальное количество корзин (buckets): 16
* При коллизии > 8 элементов в одной корзине + общий размер > 64 → список превращается в красно-чёрное дерево (Treeify)
* load factor 0.75
* Можно задать начальную capacity (ближайшая степень двойки)
```java
new HashMap<>();        // 16 корзин
new HashMap<>(100);     // 128 корзин (ближайшая степень 2)
```

Когда происходит rehash в Java HashMap
```java
HashMap<String, String> map = new HashMap<>(16);  // capacity=16, threshold=12 (16*0.75)

// Добавляем 12 элементов — OK
for(int i=0; i<12; i++) map.put("key"+i, "value");
// Размер (size) = 12, порог (threshold) = 12

map.put("key12", "value");  // 13-й элемент
// Rehash: capacity = 32, threshold = 24 (32*0.75)
// Все 13 элементов перераспределяются
```
Условия rehash:
* `size` >= threshold (порог = capacity * load_factor)
* Только после `put()` (вставки)

**Go**:
* **map** - хеш-таблица с "ведрами" (buckets) по 8 элементов
* Каждый бакет: 8 ячеек (bucket с 8 элементами)
* При заполнении бакета: создаётся overflow bucket (не новый бакет, а доп. цепочка)
* При load factor > ~6.5: происходит growing (удвоение бакетов) с постепенной эвакуацией (incremental rehashing)
```go
make(map[string]int)        // маленький начальный размер
make(map[string]int, 100)   // ~14-16 бакетов (100/8 * ~1.2)
```
В Go нет rehash в Java-стиле. Вместо этого:
* Постепенная эвакуация (incremental)
* При росте создаются новые бакеты, но элементы переносятся постепенно (при каждом доступе по 1-2 элемента)
* Нет долгой блокировки на перестроение всей таблицы

### Размер при создании

```java
// Java
new HashMap<>();        // default capacity 16
new HashMap<>(100);     // capacity = ближайшая степень двойки (128)

// Правило: возвращается ближайшая степень двойки ≥ cap
/*
    cap = 100  → 128  (2^7)
    cap = 120  → 128  (2^7)
    cap = 128  → 128  (степень двойки)
    cap = 129  → 256  (2^8)
    cap = 200  → 256
*/
```

```go
// Go
make(map[string]int)         // hint = 0 (обычный рост)
make(map[string]int, 100)    // подсказка для выделения (не точная capacity)
```

### Значения по умолчанию

```java
// Java
map.get("missing")  // возвращает null
```

```go
// Go
m["missing"]  // возвращает zero value (0 для int, "" для string)
             // чтобы отличить от реального "": value, ok := m["key"]
```

### Хранение структур

```java
// Java - можно изменять через get, так как хранятся ссылки
HashMap<String, User> map = new HashMap<>();
map.put("key", new User("Bob"));
map.get("key").setName("John");  // работает
```

```go
// Go
type User struct { Name string }

m := make(map[string]User)

u := m["key"]
u.Name = "John"  // не скомпилируется (нельзя изменить значение в map)

user := m["key"]
user.Name = "John"  // изменить через временную переменную
m["key"] = user     // записать обратно
```
