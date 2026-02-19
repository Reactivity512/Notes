# Python 3.10

Python 3.10 был выпущен 4 октября 2021 года.

## 1. Сопоставление с образцом (Pattern Matching) - оператор `match`/`case`

Самое ожидаемое нововведение, добавляющее функциональность, похожую на `switch` из других языков, но гораздо более мощную.

Простое сопоставление:
```python
def http_status(status):
    match status:
        case 200:
            return "OK"
        case 404:
            return "Not Found"
        case 500:
            return "Internal Server Error"
        case _:  # wildcard - любой другой случай
            return "Unknown status"

print(http_status(200))  # OK
```

Сопоставление с образцом (structural pattern matching):

```python
def handle_command(command):
    match command.split():
        case ["quit"]:
            print("Выход из программы")
        case ["load", filename]:
            print(f"Загрузка файла {filename}")
        case ["save", filename]:
            print(f"Сохранение в файл {filename}")
        case ["delete", *files]:
            print(f"Удаление файлов: {files}")
        case _:
            print("Неизвестная команда")

handle_command("load document.txt")  # Загрузка файла document.txt
```

Сопоставление с классами:

```python
class Point:
    def __init__(self, x, y):
        self.x = x
        self.y = y

def check_point(point):
    match point:
        case Point(x=0, y=0):
            print("Точка в начале координат")
        case Point(x=x, y=y) if x == y:
            print(f"Точка на диагонали: ({x}, {y})")
        case Point(x, y):
            print(f"Обычная точка: ({x}, {y})")

check_point(Point(0, 0))    # Точка в начале координат
check_point(Point(5, 5))    # Точка на диагонали: (5, 5)
```

## 2. Уточняющие сообщения об ошибках (Better Error Messages)

Теперь интерпретатор дает гораздо более понятные и точные сообщения об ошибках.

```python
# File "example.py", line 1
#     data = [1, 2, 3
#            ^
# SyntaxError: '[' was never closed
```

```python
#   File "example.py", line 3
#     print("Неверный отступ")
#     ^
# IndentationError: expected an indented block after 'if' statement on line 1
```

```python
# Было: SyntaxError: invalid syntax
# Стало: SyntaxError: cannot assign to literal here. Maybe you meant '==' instead of '='?
if x = 5:  # Ошибка!
    pass
```

## 3. Уточняющие типы (Precise Types)

Параметр `strict` для `zip`. Теперь `zip` может проверять, что все итерируемые имеют одинаковую длину.

Было:
```python
names = ["Alice", "Bob", "Charlie"]
ages = [25, 30]

# Молча обрезало до минимальной длины
for name, age in zip(names, ages):
    print(f"{name}: {age}")
# Alice: 25
# Bob: 30
```

Стало:
```python
names = ["Alice", "Bob", "Charlie"]
ages = [25, 30]

# Выбрасывает ошибку если длины не совпадают
for name, age in zip(names, ages, strict=True):
    print(f"{name}: {age}")
# ValueError: zip() argument 2 is shorter than argument 1
```

Тип `TypeAlias`: Явное объявление псевдонимов типов.

```python
from typing import TypeAlias

# Явное объявление псевдонима типа
UserId: TypeAlias = int
UserName: TypeAlias = str

def get_user(id: UserId) -> UserName:
    return f"User{id}"
```

Оператор `|` для типов стал стандартом. Теперь не нужно импортировать `Union`:

```python
# Вместо from typing import Union
def process_data(data: int | str | None) -> list | dict:
    pass
```

## 4. Улучшения для словарей

Метод `mapping` для `dict.keys()`, `dict.values()`, `dict.items()`.
Возвращает объект представления, который отражает изменения в исходном словаре.

```python
data = {"a": 1, "b": 2, "c": 3}
keys = data.keys()
values = data.values()

print(list(keys))    # ['a', 'b', 'c']
print(list(values))  # [1, 2, 3]

data["d"] = 4
print(list(keys))    # ['a', 'b', 'c', 'd'] - автоматически обновилось!
print(list(values))  # [1, 2, 3, 4]
```

## 5. Синтаксические улучшения

* Разрешен пробел вокруг `=` в параметрах функций

```python
def greet(
    name: str = "мир",   # Теперь можно ставить пробелы вокруг =
    punctuation: str = "!"
) -> str:
    return f"Привет, {name}{punctuation}"
```

* Разрешен оператор `with` в нескольких строках без обратного слеша

```python
# Теперь работает без обратных слешей
with (
    open("file1.txt") as f1,
    open("file2.txt") as f2,
    open("file3.txt") as f3
):
    content = f1.read() + f2.read() + f3.read()
```

## 6. Улучшения модуля `dataclasses`

Поддержка `slots` в `dataclass`

```python
from dataclasses import dataclass

@dataclass(slots=True)  # Автоматически создает __slots__
class Point:
    x: float
    y: float

p = Point(1.0, 2.0)
print(p.__slots__)  # ('x', 'y')
```

`kw_only` для `dataclasses`. Все параметры становятся `keyword-only`.

```python
from dataclasses import dataclass

@dataclass(kw_only=True)
class Person:
    name: str
    age: int = 0

# Теперь обязательно использовать имена параметров
person = Person(name="Alice", age=25)
# person = Person("Alice", 25)  # TypeError!
```

## 7. Улучшения модуля `contextlib`

Декоратор `contextlib.aclosing()`. Для корректного закрытия асинхронных контекстов.

```python
import contextlib

async def async_process():
    async with contextlib.aclosing(some_async_resource()) as resource:
        result = await resource.process()
        return result
```

## 8. Улучшения безопасности

Улучшенная система аудита для мониторинга выполнения кода.

```python
import sys
import urllib.request

# Можно отслеживать определенные события
def audit_hook(event, args):
    if event == 'urllib.Request':
        print(f"URL запрос: {args}")

sys.addaudithook(audit_hook)
urllib.request.urlopen('https://python.org')
```

## 9. Улучшения для отладки

Атрибут `__builtins__` в модулях

```python
# В любом модуле можно посмотреть встроенные функции
print(__builtins__.len)
print(__builtins__.type)
```

## 10. Улучшения аннотаций типов

`ParamSpec` и `Concatenate`. Более точной аннотации декораторов.

```python
from typing import TypeVar, ParamSpec, Callable, Concatenate

P = ParamSpec('P')
T = TypeVar('T')

def debug_decorator(func: Callable[P, T]) -> Callable[P, T]:
    def wrapper(*args: P.args, **kwargs: P.kwargs) -> T:
        print(f"Вызов {func.__name__}")
        return func(*args, **kwargs)
    return wrapper
```

Улучшена работа type checker'ов (mypy, pyright).

## 11. Улучшения производительности

* На 15% ускорено создание классов
* Ускорены некоторые математические операции
* Оптимизирована работа с атрибутами

## 12. Устаревания и предупреждения

Устарел `distutils` модуль

```python
# Было:
from distutils.core import setup

# Рекомендуется:
from setuptools import setup
```

Многие устаревшие API теперь выдают `DeprecationWarning`.

## 13. Улучшения строковых методов

`str.removeprefix()` и `str.removesuffix()` стали быстрее.

Методы, представленные в 3.9, были значительно оптимизированы:
```python
# Теперь работают еще быстрее
filename = "archive.tar.gz"
clean_name = filename.removesuffix(".gz").removeprefix("archive.")
print(clean_name)  # "tar"
```

## 14. Улучшения работы с числами

```python
# Теперь можно использовать произвольное количество цифр
# в числовых литералах с подчеркиваниями
large_number = 123_456_789_123_456_789_123_456_789
hex_value = 0xDEAD_BEEF_CAFE_BABE
binary_value = 0b1101_1110_1010_1101
```

Улучшения в `int.bit_count()`
```python
x = 255
print(x.bit_count())  # 8 (быстрее чем bin(x).count('1'))
```

## 15. Улучшения модуля `functools`

`functools.singledispatch` для методов класса

Теперь `singledispatch` можно использовать с методами классов:
```python
from functools import singledispatch

class Processor:
    @singledispatch
    def process(self, data):
        return f"Обработка неизвестного типа: {type(data)}"
    
    @process.register
    def _(self, data: int):
        return f"Обработка целого числа: {data}"
    
    @process.register
    def _(self, data: str):
        return f"Обработка строки: {data}"

processor = Processor()
print(processor.process(42))   # Обработка целого числа: 42
print(processor.process("test")) # Обработка строки: test
```

## 16. Улучшения модуля `itertools`

`itertools.pairwise()`:

Было:
```python
from itertools import tee

def pairwise(iterable):
    a, b = tee(iterable)
    next(b, None)
    return zip(a, b)
```

Стало:
```python
from itertools import pairwise

data = [1, 2, 3, 4, 5]
for a, b in pairwise(data):
    print(f"({a}, {b})")
# (1, 2)
# (2, 3)
# (3, 4)
# (4, 5)
```

## 17. Улучшения модуля `statistics`

Новые функции для статистики

```python
import statistics

data = [1, 2, 3, 4, 5, 6, 7, 8, 9]

# Ковариация
x = [1, 2, 3, 4, 5]
y = [2, 4, 6, 8, 10]
print(statistics.covariance(x, y))  # 5.0

# Корреляция Пирсона
print(statistics.correlation(x, y))  # 1.0

# Линейная регрессия
slope, intercept = statistics.linear_regression(x, y)
print(f"y = {slope:.1f}x + {intercept:.1f}")  # y = 2.0x + 0.0
```

## 18. Улучшения асинхронного программирования

* Улучшенная поддержка асинхронных генераторов:

```python
import asyncio

async def async_counter(limit):
    for i in range(limit):
        yield i
        await asyncio.sleep(0.1)

async def main():
    async for number in async_counter(5):
        print(number)

asyncio.run(main())
```

* Улучшения `asyncio.Semaphore` и `asyncio.Lock`

```python
import asyncio

async def worker(semaphore, name):
    async with semaphore:
        print(f"{name} начал работу")
        await asyncio.sleep(1)
        print(f"{name} завершил работу")

async def main():
    semaphore = asyncio.Semaphore(2)  # Не более 2 одновременных работ
    await asyncio.gather(
        worker(semaphore, "Worker 1"),
        worker(semaphore, "Worker 2"),
        worker(semaphore, "Worker 3")
    )

asyncio.run(main())
```

## 19. Улучшения модуля `pathlib`

`pathlib.Path.walk()`. Аналог `os.walk()` для `pathlib`:

```python
from pathlib import Path

# Рекурсивный обход директорий
for root, dirs, files in Path('.').walk():
    print(f"Директория: {root}")
    print(f"Поддиректории: {dirs}")
    print(f"Файлы: {files}")
    print("---")
```

## 20. Улучшения модуля `json`

`json.JSONDecodeError` с улучшенной информацией.

```python
import json

try:
    data = json.loads('{"invalid": json}')
except json.JSONDecodeError as e:
    print(f"Ошибка в позиции {e.pos}: {e.msg}")
    print(f"Строка {e.lineno}, столбец {e.colno}")
```

## 21. Улучшения модуля `logging`

`logging.getLevelNamesMapping()` - Новый метод для получения mapping'а уровней логирования:
```python
import logging

level_map = logging.getLevelNamesMapping()
print(level_map)
# {'CRITICAL': 50, 'ERROR': 40, 'WARNING': 30, 'INFO': 20, 'DEBUG': 10, 'NOTSET': 0}

# Удобно для конфигурации из строк
level_name = "INFO"
level = logging.getLevelNamesMapping().get(level_name, logging.INFO)
```

## 22. Улучшения модуля `csv`

Лучшая обработка ошибок в CSV
```python
import csv
from io import StringIO

data = "name,age\nAlice,25\nBob,invalid_age\nCharlie,30"

reader = csv.DictReader(StringIO(data))
for row in reader:
    try:
        age = int(row['age'])
        print(f"{row['name']}: {age}")
    except ValueError:
        print(f"Некорректный возраст для {row['name']}")
```

## 23. Улучшения для отладки и профилирования

Улучшенный `sys._current_frames()`
```python
import sys
import threading
import time

def worker():
    time.sleep(10)

thread = threading.Thread(target=worker)
thread.start()

# Получение фреймов всех потоков
frames = sys._current_frames()
for thread_id, frame in frames.items():
    print(f"Thread {thread_id}: {frame.f_code.co_name}")
```

Улучшения в `traceback`
```python
import traceback

def deep_function():
    raise ValueError("Тестовая ошибка")

try:
    deep_function()
except Exception:
    # Более информативный вывод
    traceback.print_exc()
```

## 24. Улучшения безопасности

* Улучшенная валидация URL в `urllib.parse`

```python
from urllib.parse import urlparse

# Более строгая проверка URL
result = urlparse("javascript:alert('xss')")
print(result.scheme)  # 'javascript' - теперь лучше обрабатываются опасные схемы
```

* Улучшения в `hashlib`

```python
import hashlib

# Новые алгоритмы и улучшенная производительность
data = b"hello world"
print(hashlib.sha3_256(data).hexdigest())
```

## 25. Улучшенный `platform` модуль и улучшения в `shutil`

```python
import platform

# Более точная информация о системе
print(f"Система: {platform.system()}")
print(f"Релиз: {platform.release()}")
print(f"Версия: {platform.version()}")
print(f"Архитектура: {platform.architecture()}")
print(f"Процессор: {platform.processor()}")
```

```python
import shutil

# Более надежные операции с файлами
shutil.copy2("source.txt", "destination.txt")
shutil.copytree("src_dir", "dst_dir", dirs_exist_ok=True)  # Новый параметр
```

## 26. Улучшения для тестирования

* Улучшения в `unittest.mock`

```python
from unittest.mock import Mock, AsyncMock

# Лучшая поддержка асинхронных mock'ов
async def async_function():
    return "result"

mock_async = AsyncMock(return_value="mocked result")
result = await mock_async()
print(result)  # "mocked result"
```

* Новые assertion методы

```python
import unittest

class TestExamples(unittest.TestCase):
    def test_new_assertions(self):
        # Более информативные сообщения об ошибках
        self.assertRegex("hello world", r"hello")
        self.assertCountEqual([1, 2, 2], [2, 1, 2])
```

## 27. Улучшения для международization (i18n)

```python
import gettext

# Более удобная работа с переводами
ru = gettext.translation('base', localedir='locales', languages=['ru'])
ru.install()
print(_("Hello, World!"))  # "Привет, Мир!"
```
