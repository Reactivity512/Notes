# Python 3.8

Python 3.8 был выпущен 14 октября 2019 года

## 1. Оператор "Морж" (Walrus Operator) `:=`

Позволяет присваивать значения переменным внутри выражений, включая условия в `if` и `while`.

Было (до 3.8):
```python
n = len([1, 2, 3])
if n > 2:
    print(f"Длина списка равна {n}")

# Или так, что менее читаемо
if len([1, 2, 3]) > 2:
    print(f"Длина списка равна {len([1, 2, 3])}") # Дважды вычисляем len
```

Стало (в 3.8):
```python
if (n := len([1, 2, 3])) > 2:
    print(f"Длина списка равна {n}")
```

## 2. Позиционные-only аргументы (Positional-only parameters)

Позволяет указать, что некоторые аргументы функции могут передаваться только позиционно (не по имени). Для этого используется символ `/` в списке параметров.

```python
def greet(name, /, greeting="Hello"):
    """
    name - позиционный-only аргумент.
    greeting - обычный аргумент (может быть и позиционным, и ключевым).
    """
    print(f"{greeting}, {name}!")

# Работает
greet("Анна")
greet("Анна", greeting="Привет")

# Вызовет ОШИБКУ!
greet(name="Анна") # TypeError: greet() got some positional-only arguments passed as keyword arguments: 'name'
```

## 3. Спецификатор `f`-строк `=` (Debug Specifier)

Упрощает отладку, позволяя выводить и имя переменной, и ее значение.

Было:
```python
x = 10
y = 25
print(f"x = {x}, y = {y}") # Приходилось дублировать имя
```

Стало:
```python
x = 10
y = 25
print(f"{x=}, {y=}")       # Выведет: x=10, y=25
print(f"{x = }, {y = }")   # Выведет: x = 10, y = 25 (с пробелами)

z = 42
print(f"{z=:.2f}")  # z=42.00
```

## 4. Улучшения модуля `typing`

Появилось несколько новых возможностей для статической типизации.

### Типизированные словари (TypedDict): Позволяет указывать типы для ключей словаря.

```python
from typing import TypedDict

class Movie(TypedDict):
    title: str
    year: int

movie: Movie = {"title": "Matrix", "year": 1999}
```

### Литеральные типы (Literal): Позволяет указывать, что аргумент может быть только одним из конкретных значений.

```python
from typing import Literal

def draw_shape(shape: Literal["circle", "square"]) -> None:
    pass

draw_shape("circle")  # OK
draw_shape("triangle") # Ошибка типа для type checker'а
```

### Final: Позволяет помечать переменные, атрибуты или методы как "финальные" (не должны переопределяться).

```python
from typing import Final

MAX_SIZE: Final = 9000
MAX_SIZE = 10000  # Type checker предупредит об этом
```

## 5. Улучшения для модуля `importlib.metadata`

Позволяет получать метаданные о установленных пакетах (например, их версию) прямо из кода.

```python
from importlib.metadata import version, requires

# Узнать версию установленного пакета
print(version('requests')) # Выведет что-то вроде '2.28.1'

# Посмотреть зависимости пакета
print(list(requires('requests')))
```

## 6. `functools.cached_property`

Декоратор, который превращает метод класса в свойство, значение которого вычисляется один раз при первом обращении, а затем кэшируется на время жизни экземпляра. ((аналог `@property`, но с кэшированием).)

```python
from functools import cached_property
import statistics

class Dataset:
    def __init__(self, sequence_of_numbers):
        self.data = sequence_of_numbers

    @cached_property
    def stdev(self):
        print("Вычисление стандартного отклонения...")
        return statistics.stdev(self.data)

dataset = Dataset([1, 2, 3, 4, 5])
print(dataset.stdev) # Вычисление произойдет здесь
print(dataset.stdev) # А здесь значение будет взято из кэша, вычисления не будет
```

## 7. Предупреждения при использовании `is` и `is not` с литералами

Теперь интерпретатор выдает предупреждение `SyntaxWarning`, если вы используете `is` или `is not` для сравнения с литералами (например, строками, числами). Эти операторы сравнивают идентичность объектов, а не их равенство, что часто приводит к ошибкам.

```python
# Работает, но теперь выдает SyntaxWarning
if x is "hello":
    pass

# Всегда правильный способ
if x == "hello":
    pass
```

## 8. `math.prod()` и `math.isqrt()`

* `math.prod(iterable)` — аналог `sum()`, но для умножения.
* `math.isqrt(n)` — целочисленный квадратный корень (возвращает `int`, а не `float`).

```python
import math

print(math.prod([2, 3, 4]))  # 24
print(math.isqrt(10))        # 3
```

## 9. Улучшения обработки исключений

* Подробные сообщения об ошибках в выражениях присваивания: Интерпретатор стал лучше указывать место ошибки, особенно в сложных выражениях, например, при использовании морж-оператора.
* Улучшенная диагностика для `SyntaxError`: Например, при забытой запятой в словаре теперь будет указываться на правильное место.
Было (в 3.7):
`SyntaxError: invalid syntax`
Стало (в 3.8):
`SyntaxError: invalid syntax. Perhaps you forgot a comma?`

## 10. Многопроцессорность: общая память в `multiprocessing`

В модуль `multiprocessing` добавлена поддержка общей памяти (`shared_memory`), позволяющая создавать регионы памяти, доступные для прямого чтения и записи из разных процессов Python, без необходимости сериализации через очередь.

```python
from multiprocessing import shared_memory
import numpy as np

# Создаем блок общей памяти
shm = shared_memory.SharedMemory(create=True, size=100)
# Представляем его как numpy array
buffer = np.ndarray((25,), dtype=np.int32, buffer=shm.buf)
buffer[:] = np.arange(25)  # Записываем данные

# В другом процессе можно подключиться к этому блоку по имени
existing_shm = shared_memory.SharedMemory(name=shm.name)
```

## 11. Новый протокол импорта Python (Python Import Protocol)

Была добавлена поддержка нового, более гибкого, но и более сложного протокола импорта модулей (`importlib`). Это изменение "под капотом" в первую очередь интересно разработчикам фреймворков и систем сборки, которые хотят полностью кастомизировать процесс загрузки модулей.

## 12. `as_integer_ratio()` для `bool`

Метод `as_integer_ratio()`, ранее доступный для `float` и `decimal.Decimal`, теперь добавлен и для типа `bool`. Он возвращает пару `(numerator, denominator)`, где для `True` это `(1, 1)`, а для `False` — `(0, 1)`.

```python
print(True.as_integer_ratio()) # (1, 1)
print(False.as_integer_ratio()) # (0, 1)
```

Хотя это может показаться странным, это помогает унифицировать интерфейсы разных типов.

## 13. `continue` в `finally`

Начиная с Python 3.8, внутри блока `finally` теперь можно использовать операторы `continue`. Раньше это вызывало `SyntaxError`.

```python
for i in range(5):
    try:
        # Что-то делаем
        pass
    finally:
        if i == 2:
            continue  # Теперь это разрешено!
        print(f"Завершение итерации {i}")
```

## 14. `__future__`аннотации и отложенные аннотации типов (PEP 563)

Хотя PEP 563 («Postponed Evaluation of Annotations») был введён в Python 3.7 как опциональный, в 3.8 он стал рекомендованным, а в 3.10 — включён по умолчанию.

В Python 3.8 можно включить отложенную оценку аннотаций:

```python
from __future__ import annotations

def greet(name: str) -> Greeting:  # Greeting ещё не определён — но это нормально
    return Greeting(f"Привет, {name}!")

class Greeting:
    def __init__(self, msg: str):
        self.msg = msg
```

## 15. Производительность и оптимизации

* Оптимизация вызовов методов: Повторные вызовы методов (например, `obj.method()`) теперь работают быстрее (до 20% в некоторых бенчмарках), так как интерпретатор кэширует поиск метода.
* Оптимизация `pickle`: Модуль `pickle` теперь использует Protocol 5 (внедренный в Python 3.8) для более эффективной сериализации больших объектов с поддержкой out-of-band данных.
* Быстрее создание экземпляров класса: Процесс создания экземпляра класса (`__new__` и `__init__`) был немного оптимизирован.

## 16. Устаревания и предупреждения

* `collections` абстрактные базовые классы: Использование абстрактных базовых классов из модуля `collections` (например, `collections.Mapping`) официально объявлено устаревшим. Вместо них следует использовать классы из `collections.abc`.
* Итераторы, не генерирующие `StopIteration`: Генератор, который просто завершает выполнение, должен использовать `return`, а не вызывать `StopIteration` вручную. Такое поведение станет ошибкой в будущих версиях.
* Модуль `py_compile` и `compileall`: По умолчанию теперь генерируют байт-код с оптимизацией `-O` (удаляются операторы `assert`), если скрипт запущен с флагом `-O`.
