# Python 3.11

Python 3.11 был выпущен 24 октября 2022 года

## 1. Значительное увеличение производительности

Python 3.11 работает на 25-60% быстрее чем Python 3.10 благодаря проекту "Faster CPython".

Адаптивный интерпретатор:
```py
# Пример, где новая оптимизация особенно заметна
def calculate_sum(n):
    result = 0
    for i in range(n):  # Быстрые итерации
        result += i
    return result

# Эта функция выполняется значительно быстрее в 3.11
print(calculate_sum(1_000_000))
```

Оптимизация вызовов функций:
```py
# Рекурсивные вызовы стали значительно быстрее
def factorial(n):
    return 1 if n <= 1 else n * factorial(n-1)

# В 3.11 рекурсия оптимизирована
print(factorial(100))
```

## 2. Улучшенные сообщения об ошибках

Теперь ошибки показывают конкретный фрагмент кода, где произошла ошибка.

Неопределенные переменные:
```py
def example():
    print(undefined_variable)

example()
```

Было в 3.10:
```
NameError: name 'variable_name' is not defined
```
Стало в 3.11:
```
Traceback (most recent call last):
  File "example.py", line 4, in <module>
    example()
  File "example.py", line 2, in example
    print(undefined_variable)
          ^^^^^^^^^^^^^^^^^^
NameError: name 'undefined_variable' is not defined
```

Ошибки в словарях:
```py
data = {
    'name': 'Alice',
    'age': 30,
    'city': 'Moscow'  # ← Забыта запятая
    'country': 'Russia' 
}
```

```
  File "example.py", line 5
    'country': 'Russia'
    ^
SyntaxError: invalid syntax. Perhaps you forgot a comma?
```

## 3. Новый синтаксис для исключений - `ExceptionGroups`

Позволяет группировать несколько исключений вместе.

```py
def validate_user(user_data):
    errors = []
    if not user_data.get('name'):
        errors.append(ValueError("Имя обязательно"))
    if not user_data.get('email'):
        errors.append(ValueError("Email обязателен"))
    if errors:
        raise ExceptionGroup("Ошибки валидации", errors)

try:
    validate_user({'age': 25})
except* ValueError as eg:
    for error in eg.exceptions:
        print(f"Ошибка: {error}")

# Ошибка: Имя обязательно
# Ошибка: Email обязателен
```

Обработка разных типов исключений:
```py
try:
    raise ExceptionGroup("комплексная ошибка", [
        ValueError("неправильное значение"),
        TypeError("неправильный тип"),
        KeyError("отсутствующий ключ")
    ])
except* ValueError as ve:
    print(f"ValueError: {ve.exceptions}")
except* (TypeError, KeyError) as te:
    print(f"TypeError/KeyError: {te.exceptions}")
```

## 4. Новые типы в модуле `typing`

* `Self` тип для возврата экземпляра класса

```py
from typing import Self

class Database:
    def set_host(self, host: str) -> Self:
        self.host = host
        return self  # Больше не нужно писать -> 'Database'
    
    def set_port(self, port: int) -> Self:
        self.port = port
        return self

# Чейнинг методов с правильной типизацией
db = Database().set_host("localhost").set_port(5432)
```

* `LiteralString` для защиты от SQL-инъекций

```py
from typing import LiteralString

def execute_query(query: LiteralString) -> None:
    print(f"Выполняем: {query}")

# Это работает
safe_query: LiteralString = "SELECT * FROM users"
execute_query(safe_query)

# Type checker предупредит об этом
user_input = input()  # str, не LiteralString
# execute_query(user_input)  # Ошибка типов!
```

* `Never` и `assert_never`

```py
from typing import Never, assert_never
from enum import Enum

class Color(Enum):
    RED = "red"
    GREEN = "green"
    BLUE = "blue"

def handle_color(color: Color) -> str:
    match color:
        case Color.RED:
            return "красный"
        case Color.GREEN:
            return "зеленый"
        case Color.BLUE:
            return "синий"
        case _:
            assert_never(color)  # Гарантирует, что все случаи обработаны
```

## 5. Улучшения для работы с модулями

`sys.exception()` вместо `sys.exc_info()`

```py
import sys

try:
    1 / 0
except ZeroDivisionError:
    # Новый способ получить текущее исключение
    current_exception = sys.exception()
    print(f"Исключение: {current_exception}")
```

## 6. Новые возможности для объектов

`object.__getstate__()` для кастомизации pickle

```py
class CustomObject:
    def __init__(self, data):
        self.data = data
        self._cache = {}  # Не хотим сериализовать кэш
    
    def __getstate__(self):
        # Возвращаем только то, что нужно сериализовать
        state = self.__dict__.copy()
        del state['_cache']
        return state
    
    def __setstate__(self, state):
        self.__dict__.update(state)
        self._cache = {}  # Восстанавливаем кэш
```

## 7. Улучшения для асинхронного программирования

`asyncio.TaskGroup` - замена `asyncio.gather`. Более безопасный способ управления группой задач.

Старый способ:
```py
import asyncio

async def main():
    task1 = asyncio.create_task(coroutine1())
    task2 = asyncio.create_task(coroutine2())
    await asyncio.gather(task1, task2)
```

Новый способ:
```py
import asyncio

async def main():
    async with asyncio.TaskGroup() as tg:
        task1 = tg.create_task(coroutine1())
        task2 = tg.create_task(coroutine2())
    # Все задачи завершены здесь
```

Преимущества TaskGroup:
* Если одна задача падает, отменяются все остальные
* Более чистый синтаксис
* Лучшая обработка ошибок

## 8. Улучшения математических функций

Новые функции в модуле `math`

```py
import math

# Сумма произведений (часто используется в машинном обучении)
a = [1, 2, 3]
b = [4, 5, 6]
print(math.sumprod(a, b))  # (1*4 + 2*5 + 3*6) = 32

# Ближайшее снизу/сверху целое с заданным шагом
print(math.nextafter(1.0, 2.0))  # 1.0000000000000002
```

## 9. Улучшения для отладки

`tomllib` - встроенная поддержка TOML

```py
import tomllib

# Чтение TOML файлов (аналог json для конфигураций)
config = """
[server]
host = "localhost"
port = 8080

[database]
url = "postgresql://user:pass@localhost/db"
"""

data = tomllib.loads(config)
print(data['server']['host'])  # localhost
```

## 10. Улучшения для типизации

`@typing.override` декоратор

```py
from typing import override

class Base:
    def process(self) -> str:
        return "base"

class Derived(Base):
    @override  # Проверяет, что метод действительно переопределен
    def process(self) -> str:
        return "derived"

# Если бы мы опечатались в имени метода, type checker бы предупредил
```

## 11. Устаревания и удаления

Удален модуль `distutils` в пользу `setuptools`.

Устаревшие API:
* Некоторые старые API в `asyncio`
* Устаревшие методы в `email` модуле
* Старые форматы в `datetime`

## 12. Улучшения низкоуровневого программирования

* Новый C API для управления GIL (Global Interpreter Lock)
* Улучшенная поддержка подпроцессов
```py
import subprocess

# Более эффективное создание подпроцессов
result = subprocess.run(['ls', '-la'], capture_output=True, text=True)
print(result.stdout)
```

## 13. Улучшения модуля `asyncio`

* `asyncio.Runner` для управления event loop

```py
import asyncio

async def main():
    print("Hello")
    await asyncio.sleep(1)
    print("World")

# Новый способ запуска асинхронного кода
with asyncio.Runner() as runner:
    runner.run(main())
```

* Улучшения `asyncio.Timeout`

```py
import asyncio

async def slow_operation():
    await asyncio.sleep(10)
    return "Готово"

async def main():
    try:
        # Более точный контроль таймаутов
        async with asyncio.timeout(5):
            result = await slow_operation()
            print(result)
    except TimeoutError:
        print("Операция заняла слишком много времени!")

asyncio.run(main())
```

## 14. Улучшения модуля `datetime`

Поддержка любых лет в `datetime`

```py
from datetime import datetime, timezone

# Теперь можно работать с датами до 1 года
ancient_date = datetime(100, 1, 1)
future_date = datetime(10000, 1, 1)

print(ancient_date)  # 0100-01-01 00:00:00
print(future_date)   # 10000-01-01 00:00:00
```

## 15. Улучшения модуля `inspect`

`inspect.get_annotations()`
```py
import inspect
from typing import Optional

class Example:
    name: str
    age: Optional[int] = None

# Новый способ получения аннотаций
annotations = inspect.get_annotations(Example)
print(annotations)  # {'name': <class 'str'>, 'age': typing.Optional[int]}
```

Улучшенная интроспекция генераторов:
```py
import inspect

def generator_function():
    yield 1
    yield 2

gen = generator_function()
print(inspect.getgeneratorstate(gen))  # Более точная информация о состоянии
```

## 16. Улучшения модуля `socket`

Поддержка быстрого повторного использования портов
```py
import socket

# Улучшенная работа с сокетами, особенно для серверов
server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
server_socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
```

## 17. Улучшения модуля `hashlib`

Поддержка новых алгоритмов хеширования
```py
import hashlib

# Новые алгоритмы и улучшенная производительность
data = b"hello world"

# Blake3 - новый быстрый алгоритм
try:
    print(hashlib.blake3(data).hexdigest())
except AttributeError:
    print("Blake3 не доступен")

# Улучшенная производительность для существующих алгоритмов
print(hashlib.sha256(data).hexdigest())
print(hashlib.sha3_512(data).hexdigest())
```

## 18. Улучшения модуля `secrets`

Генерация более безопасных токенов

```py
import secrets

# Улучшенная генерация криптографически безопасных данных
token = secrets.token_urlsafe(32)
print(f"Безопасный токен: {token}")

# Более эффективная генерация случайных чисел
random_number = secrets.randbelow(1000)
print(f"Случайное число: {random_number}")
```

## 19. Улучшения модуля `contextvars`

Более эффективная работа с контекстными переменными

```py
import contextvars
import asyncio

# Контекстные переменные стали работать быстрее
user_id = contextvars.ContextVar('user_id')

async def process_request():
    user_id.set(123)
    print(f"Обработка для пользователя {user_id.get()}")
    
    # Улучшенная производительность для вложенных контекстов
    async with some_context():
        print(f"Вложенный контекст: {user_id.get()}")
```

## 20. Улучшения для отладки и профилирования

* Улучшенный `faulthandler`

```py
import faulthandler
import sys

# Лучшая диагностика сбоев
faulthandler.enable()

# Запись дампа при определенных сигналах
if sys.platform != "win32":
    faulthandler.register(signal.SIGUSR1, all_threads=True)
```

* Улучшения в `tracemalloc`

```py
import tracemalloc

# Более точное отслеживание использования памяти
tracemalloc.start()

# Создаем некоторые данные
data = [bytearray(1000) for _ in range(100)]

# Получаем снимок памяти
snapshot = tracemalloc.take_snapshot()
top_stats = snapshot.statistics('lineno')

print("[ Top 10 ]")
for stat in top_stats[:10]:
    print(stat)
```

## 21. Улучшения модуля `pickle`

Более эффективная сериализация

```py
import pickle

class ComplexObject:
    def __init__(self, data):
        self.data = data
        self.metadata = {"created": "2023", "version": 1.0}

obj = ComplexObject(list(range(1000)))

# Улучшенная производительность сериализации
serialized = pickle.dumps(obj, protocol=5)
deserialized = pickle.loads(serialized)

print(f"Размер сериализованных данных: {len(serialized)} байт")
```

## 22. Улучшения модуля `json`

Более быстрый парсинг JSON
```py
import json
import time

large_json = '{"data": [' + ','.join(['{"id": ' + str(i) + '}' for i in range(10000)]) + ']}'

start = time.time()
data = json.loads(large_json)
end = time.time()

print(f"Парсинг занял: {(end - start) * 1000:.2f} мс")
```

## 23. Улучшения для системных вызовов

Более эффективная работа с файлами

```py
import os

# Улучшенная работа с путями и файлами
path = "/some/directory"

# Более быстрая проверка существования
if os.path.exists(path):
    print("Путь существует")
    
# Улучшенная работа с разрешениями
stat_info = os.stat(path)
print(f"Размер: {stat_info.st_size} байт")
```

## 24. Улучшения модуля `threading`

Более эффективные блокировки

```py
import threading
import time

shared_data = []
lock = threading.Lock()

def worker(thread_id):
    for i in range(100):
        # Улучшенная производительность блокировок
        with lock:
            shared_data.append(f"Thread {thread_id}: {i}")
        time.sleep(0.001)

threads = []
for i in range(5):
    t = threading.Thread(target=worker, args=(i,))
    threads.append(t)
    t.start()

for t in threads:
    t.join()

print(f"Обработано {len(shared_data)} элементов")
```

## 25. Улучшения для научных вычислений

Более быстрые математические операции

```py
import math

# Ускоренные математические функции
numbers = [math.sin(i * 0.1) for i in range(1000)]

# Более быстрые агрегатные операции
total = sum(numbers)
mean = total / len(numbers)

print(f"Сумма: {total:.4f}, Среднее: {mean:.4f}")
```

## 26. Улучшения модуля `re` (регулярные выражения)

Оптимизация компиляции регулярных выражений

```py
import re
import time

# Регулярные выражения компилируются быстрее
pattern = re.compile(r'\b\w{4,}\b')

text = "Эта строка содержит несколько слов разной длины"

start = time.time()
matches = pattern.findall(text)
end = time.time()

print(f"Найдены слова: {matches}")
print(f"Поиск занял: {(end - start) * 1000:.4f} мс")
```

## 27. Улучшения для работы с байтами

Более эффективные операции с bytes и bytearray

```py
# Ускоренные операции с байтами
data = bytearray(1000)

# Быстрее заполнение и модификация
for i in range(len(data)):
    data[i] = i % 256

# Ускоренные преобразования
as_bytes = bytes(data)
back_to_array = bytearray(as_bytes)

print(f"Размер данных: {len(as_bytes)} байт")
```

## 28. Улучшения GC (Garbage Collector)

Более эффективная сборка мусора

```py
import gc
import sys

# Создаем циклические ссылки для демонстрации
class Node:
    def __init__(self, name):
        self.name = name
        self.next = None

# Создаем цикл
a = Node("A")
b = Node("B")
a.next = b
b.next = a  # Циклическая ссылка

# Улучшенный GC лучше справляется с такими случаями
collected = gc.collect()
print(f"Собрано объектов: {collected}")

# Более точная статистика
print(f"Текущие объекты: {len(gc.get_objects())}")
```

## 29. Улучшения для кроссплатформенной разработки

Улучшенная поддержка Windows

```py
import os
import sys

# Лучшая работа с путями на разных платформах
if sys.platform == "win32":
    # Улучшенная обработка путей Windows
    path = r"C:\Users\Username\Documents"
else:
    path = "/home/username/documents"

# Универсальная работа с путями
normalized_path = os.path.normpath(path)
print(f"Нормализованный путь: {normalized_path}")
```
