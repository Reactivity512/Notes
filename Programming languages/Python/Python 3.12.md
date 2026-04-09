# Python 3.12

Python 3.12 был выпущен 2 октября 2023 года.

## 1. Новый синтаксис: Форматированные строки (f-strings) становятся еще мощнее

f-strings теперь поддерживают произвольные выражения

Было (до 3.12):
```py
# Многострочные f-strings были ограничены
name = "Alice"
age = 30
# Приходилось делать так:
message = (
    f"Имя: {name}, "
    f"Возраст: {age}, "
    f"Год рождения: {2023 - age}"
)
```

Стало (в 3.12):
```py
name = "Alice"
age = 30

# Многострочные f-strings с комментариями и произвольными выражениями
message = f"""
Имя: {name},
Возраст: {age},
Год рождения: {2023 - age},  # Можно вычислять прямо здесь
Статус: {"совершеннолетний" if age >= 18 else "несовершеннолетний"}
"""
print(message)
```

Вложенные f-strings с кавычками:
```py
# Теперь можно легко вкладывать f-strings
name = "Alice"
template = "приветствие"

message = f"{f'{name}, добро пожаловать!'}"  # Работает!
print(message)  # Alice, добро пожаловать!
```

## 2. Новый синтаксис: Параметры типа (Type Parameter Syntax)

Упрощенный синтаксис для дженериков

Было:
```py
from typing import TypeVar, Generic

T = TypeVar('T')
U = TypeVar('U')

class Container(Generic[T, U]):
    def __init__(self, value1: T, value2: U):
        self.value1 = value1
        self.value2 = value2
```

Стало:
```py
class Container[T, U]:  # Новый синтаксис!
    def __init__(self, value1: T, value2: U):
        self.value1 = value1
        self.value2 = value2

def first_item[V](items: list[V]) -> V:
    return items[0]

# Использование
container = Container[int, str](42, "hello")
number = first_item([1, 2, 3])
```

Функции с параметрами типа:
```py
def process_data[T](data: list[T]) -> T:
    return data[0]

def merge_dicts[K, V](dict1: dict[K, V], dict2: dict[K, V]) -> dict[K, V]:
    return {**dict1, **dict2}

# Автовывод типов работает отлично
result1 = process_data([1, 2, 3])        # result1: int
result2 = merge_dicts({"a": 1}, {"b": 2}) # result2: dict[str, int]
```

## 3. Улучшения производительности

Подпроект "Faster CPython" продолжается, Python 3.12 еще на 10-25% быстрее чем 3.11.

```py
# Пример, демонстрирующий улучшения
def process_large_data():
    data = [i ** 2 for i in range(1000000)]
    
    # Вложенные циклы стали значительно быстрее
    result = []
    for i in data:
        for j in range(10):
            if i % 2 == 0:
                result.append(i + j)
    return sum(result)

# Время выполнения уменьшилось заметно
```

Аннотации типов больше не замедляют выполнение кода:
```py
def calculate(x: int, y: int) -> int:  # Аннотации не влияют на скорость
    return x * y + 1000
```

## 4. Улучшенные сообщения об ошибках

* Еще более точные указания на ошибки

```py
# Более точное указание на проблему в сложных выражениях
data = {
    'users': [
        {'name': 'Alice', 'age': 30},
        {'name': 'Bob', 'age': 25}
    ],
    'settings': {
        'theme': 'dark'
        'language': 'en'  # ← Забыта запятая
    }
}
```

```
  File "example.py", line 8
    'language': 'en'
    ^^^^^^^^^^^^^^^^
SyntaxError: invalid syntax. Perhaps you forgot a comma?
```

* Лучшие сообщения для импортов:

```py
# Если модуль не найден, объяснение стало понятнее
import non_existent_module
```

```
ModuleNotFoundError: No module named 'non_existent_module'
Did you mean: 'existing_module'?
```

## 5. Новые возможности для типизации

```py
# Новый синтаксис для псевдонимов типов
type UserId = int
type UserData = dict[str, str | int]
type StringOrInt = str | int

def get_user(id: UserId) -> UserData:
    return {"name": "Alice", "age": 30, "id": id}
```

`@override` декоратор теперь в стандартной библиотеке

```py
from typing import override

class Base:
    def process(self) -> str:
        return "base"

class Derived(Base):
    @override  # Теперь встроенный декоратор
    def process(self) -> str:
        return "derived"

    # @override  # Если раскомментировать - будет ошибка типов
    # def wrong_method(self) -> str:  # Метод не существует в базовом классе
    #     return "error"
```

## 6. Улучшения для работы с файлами и путями

`pathlib` теперь быстрее и функциональнее

```py
from pathlib import Path

# Ускоренные операции с путями
current = Path.cwd()
new_file = current / "data.txt"

# Новые методы и улучшенная производительность
if new_file.exists():
    content = new_file.read_text(encoding='utf-8')
    print(f"Размер файла: {new_file.stat().st_size} байт")
```

## 7. Улучшения модуля `asyncio`

Упрощенное создание асинхронных программ

```py
import asyncio

# Новые высокоуровневые API для работы с таймерами
async def main():
    print("Начало")
    
    # Асинхронный sleep с лучшим контролем
    await asyncio.sleep(1.5)
    
    print("Прошла 1.5 секунды")
    
    # Улучшенная работа с таймаутами
    async with asyncio.timeout(5):
        await some_long_operation()

asyncio.run(main())
```

## 8. Новый модуль `tomllib`

Полная поддержка TOML в стандартной библиотеке

```py
import tomllib  # Теперь встроенный модуль!

# Чтение TOML из строки
config_text = """
[server]
host = "localhost"
port = 8080
enabled = true

[database]
url = "postgresql://user:pass@localhost/db"
retry_attempts = 3
"""

config = tomllib.loads(config_text)
print(f"Сервер: {config['server']['host']}:{config['server']['port']}")

# Чтение из файла
with open("config.toml", "rb") as f:
    file_config = tomllib.load(f)
```

## 9. Улучшения для отладки

Более информативные tracebacks

```py
def deep_function():
    raise ValueError("Глубокая ошибка")

def middle_function():
    deep_function()

def top_function():
    middle_function()

try:
    top_function()
except ValueError:
    import traceback
    traceback.print_exc()  # Более понятный вывод
```

```
Traceback (most recent call last):
  File "/Users/sergey/Documents/py/main2.py", line 11, in <module>
    top_function()
    ~~~~~~~~~~~~^^
  File "/Users/sergey/Documents/py/main2.py", line 8, in top_function
    middle_function()
    ~~~~~~~~~~~~~~~^^
  File "/Users/sergey/Documents/py/main2.py", line 5, in middle_function
    deep_function()
    ~~~~~~~~~~~~~^^
  File "/Users/sergey/Documents/py/main2.py", line 2, in deep_function
    raise ValueError("Глубокая ошибка")
ValueError: Глубокая ошибка
```

## 10. Улучшенная валидация URL

```py
from urllib.parse import urlparse

# Более строгая проверка URL
def safe_url_check(url):
    parsed = urlparse(url)
    if parsed.scheme not in ('http', 'https'):
        raise ValueError("Разрешены только HTTP и HTTPS URL")
    return True

safe_url_check("https://example.com")  # OK
# safe_url_check("javascript:alert('xss')")  # ValueError
```

## 11. Улучшения для научных вычислений

```py
import math

# Быстрые математические операции стали еще быстрее
result = sum(math.log(x) for x in data if x > 0)
print(f"Сумма логарифмов: {result}")
```

## 12. Устаревания и удаления

* Устаревшие функции из `asyncio` удалены
* Старые форматы конфигураций больше не поддерживаются

```py
import warnings

# Многие устаревшие функции теперь выдают DeprecationWarning
warnings.warn("Это устаревшая функция", DeprecationWarning)
```

## 13. Улучшения модуля `subprocess`

Новые высокоуровневые API для подпроцессов:

```py
import subprocess

# Новый удобный API для запуска команд
result = subprocess.run(
    ["python", "--version"],
    capture_output=True,
    text=True,
    check=True  # Автоматически вызывает исключение при ошибке
)

print(f"Версия Python: {result.stdout.strip()}")

# Асинхронная поддержка улучшена
async def run_async_command():
    process = await asyncio.create_subprocess_exec(
        "ls", "-la",
        stdout=asyncio.subprocess.PIPE,
        stderr=asyncio.subprocess.PIPE
    )
    stdout, stderr = await process.communicate()
    return stdout.decode()
```

## 14. Улучшения модуля `socket`

Поддержка быстрого переиспользования портов

```py
import socket
import contextlib

@contextlib.contextmanager
def create_server_socket(host='localhost', port=0):
    """Создание серверного сокета с улучшенными настройками"""
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    
    # Улучшенные опции для production использования
    sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    
    # Новая опция для быстрого переиспользования портов
    if hasattr(socket, 'SO_REUSEPORT'):
        sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEPORT, 1)
    
    try:
        sock.bind((host, port))
        sock.listen()
        yield sock
    finally:
        sock.close()

# Использование
with create_server_socket(port=8080) as server_socket:
    print(f"Сервер запущен на порту: {server_socket.getsockname()[1]}")
```

## 15. Улучшения модуля `threading`

Более эффективные примитивы синхронизации

```py
import threading
from concurrent.futures import ThreadPoolExecutor

# Улучшенная производительность блокировок
class ThreadSafeCounter:
    def __init__(self):
        self._value = 0
        self._lock = threading.Lock()
    
    def increment(self):
        # Более быстрая работа с блокировками
        with self._lock:
            self._value += 1
            return self._value

# Тестирование производительности
counter = ThreadSafeCounter()

def worker():
    for _ in range(1000):
        counter.increment()

# Запуск в пуле потоков
with ThreadPoolExecutor(max_workers=10) as executor:
    futures = [executor.submit(worker) for _ in range(10)]
    
    # Ожидание завершения
    for future in futures:
        future.result()

print(f"Финальное значение: {counter._value}")
```

## 16. Улучшения модуля `multiprocessing`

Более эффективная межпроцессная коммуникация

```py
import multiprocessing as mp
import time

def worker_function(shared_value, results_queue, process_id):
    """Функция-воркер с улучшенной IPC"""
    for _ in range(100):
        with shared_value.get_lock():
            shared_value.value += 1
            current_value = shared_value.value
        
        # Улучшенная передача данных через очередь
        results_queue.put({
            'process_id': process_id,
            'value': current_value,
            'timestamp': time.time()
        })
        time.sleep(0.001)

def main():
    # Улучшенные shared values
    shared_value = mp.Value('i', 0)
    results_queue = mp.Queue()
    
    processes = []
    for i in range(4):
        p = mp.Process(
            target=worker_function,
            args=(shared_value, results_queue, i)
        )
        processes.append(p)
        p.start()
    
    # Сбор результатов
    results = []
    for p in processes:
        p.join()
    
    # Обработка результатов из очереди
    while not results_queue.empty():
        results.append(results_queue.get())
    
    print(f"Финальное значение: {shared_value.value}")
    print(f"Собрано результатов: {len(results)}")

if __name__ == "__main__":
    main()
```

## 17. Улучшения модуля `inspect`

Расширенная интроспекция кода

```py
import inspect
from typing import get_type_hints

class DataProcessor:
    def __init__(self, data: list[int]):
        self.data = data
    
    def process(self, multiplier: int = 2) -> list[int]:
        """Обрабатывает данные и возвращает результат"""
        return [x * multiplier for x in self.data]
    
    async def process_async(self) -> list[int]:
        """Асинхронная обработка"""
        return self.process()

# Новая функциональность интроспекции
def analyze_class(cls):
    print(f"Анализ класса: {cls.__name__}")
    
    # Получение аннотаций с улучшенной поддержкой
    annotations = get_type_hints(cls)
    print(f"Аннотации класса: {annotations}")
    
    # Анализ методов
    for name, method in inspect.getmembers(cls, predicate=inspect.isfunction):
        print(f"\nМетод: {name}")
        print(f"Сигнатура: {inspect.signature(method)}")
        print(f"Аннотации: {get_type_hints(method)}")
        
        # Проверка на асинхронность
        if inspect.iscoroutinefunction(method):
            print("⚡ Асинхронный метод")

analyze_class(DataProcessor)
```

## 18. Улучшения модуля `contextlib`

Новые декораторы и утилиты

```py
import contextlib
import time
from typing import Iterator

@contextlib.contextmanager
def timed_operation(operation_name: str) -> Iterator[None]:
    """Контекстный менеджер для измерения времени"""
    start_time = time.perf_counter()
    try:
        print(f"Начало: {operation_name}")
        yield
    finally:
        end_time = time.perf_counter()
        duration = end_time - start_time
        print(f"Завершено: {operation_name} за {duration:.3f} секунд")

# Использование улучшенных контекстных менеджеров
with timed_operation("Обработка данных"):
    data = [i ** 2 for i in range(100000)]
    result = sum(data)

# Цепочка контекстных менеджеров
@contextlib.contextmanager
def debug_context():
    print("Начало отладки")
    try:
        yield
    except Exception as e:
        print(f"Ошибка в контексте: {e}")
        raise
    finally:
        print("Конец отладки")

# Комбинирование контекстов
with debug_context(), timed_operation("Сложная операция"):
    complex_data = [x for x in range(1000) if x % 2 == 0]
```

## 19. Улучшения модуля `functools`

Новые декораторы и оптимизации

```py
import functools
from typing import TypeVar, ParamSpec

P = ParamSpec('P')
T = TypeVar('T')

# Улучшенный singledispatch с поддержкой типов
@functools.singledispatch
def process_data(data):
    """Обработка данных по умолчанию"""
    return f"Обработка неизвестного типа: {type(data)}"

@process_data.register
def _(data: int) -> str:
    return f"Обработка целого числа: {data}"

@process_data.register
def _(data: str) -> str:
    return f"Обработка строки: {data}"

@process_data.register
def _(data: list) -> str:
    return f"Обработка списка длиной {len(data)}"

# Тестирование
print(process_data(42))        # Обработка целого числа: 42
print(process_data("hello"))   # Обработка строки: hello
print(process_data([1, 2, 3])) # Обработка списка длиной 3
```

## 20. Улучшения модуля `statistics`

Расширенная статистическая функциональность

```py
import statistics
import random

# Генерация тестовых данных
data = [random.gauss(100, 15) for _ in range(1000)]

# Новые статистические функции и улучшения
def comprehensive_analysis(dataset):
    """Всесторонний статистический анализ"""
    
    # Базовая статистика
    mean = statistics.mean(dataset)
    median = statistics.median(dataset)
    stdev = statistics.stdev(dataset)
    
    # Новые метрики
    try:
        mode = statistics.mode(dataset)
    except statistics.StatisticsError:
        mode = "Нет моды"
    
    # Квантили и перцентили
    quantiles = statistics.quantiles(dataset, n=4)  # Квартили
    deciles = statistics.quantiles(dataset, n=10)   # Децили
    
    print(f"Статистический анализ:")
    print(f"  Среднее: {mean:.2f}")
    print(f"  Медиана: {median:.2f}")
    print(f"  Стандартное отклонение: {stdev:.2f}")
    print(f"  Мода: {mode}")
    print(f"  Квартили: {[f'{q:.2f}' for q in quantiles]}")
    print(f"  Децили: {[f'{d:.2f}' for d in deciles[:3]]}...")

comprehensive_analysis(data)
```

## 21. Улучшения для отладки и профилирования

```py
import cProfile
import pstats
import io
from functools import wraps

def profile_function(func):
    """Декоратор для профилирования функций"""
    @wraps(func)
    def wrapper(*args, **kwargs):
        profiler = cProfile.Profile()
        profiler.enable()
        
        try:
            result = func(*args, **kwargs)
        finally:
            profiler.disable()
            
            # Улучшенный вывод статистики
            s = io.StringIO()
            ps = pstats.Stats(profiler, stream=s).sort_stats('cumulative')
            ps.print_stats(20)  # Топ-20 функций
            
            print(f"📊 Профилирование {func.__name__}:")
            print(s.getvalue())
        
        return result
    return wrapper

@profile_function
def expensive_operation():
    """Дорогая операция для профилирования"""
    data = []
    for i in range(10000):
        data.append(i ** 2)
    
    # Симуляция сложных вычислений
    result = sum(x for x in data if x % 2 == 0)
    return result

# Запуск профилирования
expensive_operation()
```

## 22. Улучшенная криптография и безопасность

```py
import secrets
import hashlib
import hmac

def secure_password_hash(password: str, salt: bytes = None) -> tuple[bytes, bytes]:
    """Безопасное хеширование пароля с улучшенными алгоритмами"""
    
    if salt is None:
        # Генерация криптографически безопасной соли
        salt = secrets.token_bytes(32)
    
    # Улучшенное хеширование с большим количеством итераций
    password_bytes = password.encode('utf-8')
    
    # Использование современных алгоритмов
    hash_result = hashlib.pbkdf2_hmac(
        'sha256',
        password_bytes,
        salt,
        100000,  # Количество итераций увеличено для безопасности
        dklen=128
    )
    
    return hash_result, salt

def verify_password(password: str, stored_hash: bytes, salt: bytes) -> bool:
    """Проверка пароля"""
    new_hash, _ = secure_password_hash(password, salt)
    return hmac.compare_digest(new_hash, stored_hash)

# Пример использования
password = "my_secure_password"
hash_result, salt = secure_password_hash(password)

print(f"Соль: {salt.hex()[:32]}...")
print(f"Хеш: {hash_result.hex()[:32]}...")

# Проверка
is_valid = verify_password("my_secure_password", hash_result, salt)
print(f"Пароль верный: {is_valid}")
```

## 23. Улучшенная поддержка Unicode и локалей

```py
import locale
import unicodedata

def unicode_analysis(text: str):
    """Анализ Unicode строк с улучшенной поддержкой"""
    
    print(f"Исходный текст: {text}")
    print(f"Длина: {len(text)} символов")
    print(f"Длина в байтах (UTF-8): {len(text.encode('utf-8'))} байт")
    
    # Анализ каждого символа
    for i, char in enumerate(text):
        char_info = {
            'char': char,
            'name': unicodedata.name(char, 'UNKNOWN'),
            'category': unicodedata.category(char),
            'numeric': unicodedata.numeric(char, None)
        }
        print(f"  {i:2d}: {char_info}")

# Тестирование с различными символами
test_text = "Hello 世界 🎉 Café"
unicode_analysis(test_text)

# Улучшенная работа с локалями
def locale_info():
    """Информация о текущей локали"""
    current_locale = locale.getlocale()
    print(f"Текущая локаль: {current_locale}")
    
    # Попытка установки локали
    try:
        locale.setlocale(locale.LC_ALL, 'en_US.UTF-8')
        print("Локаль установлена: en_US.UTF-8")
    except locale.Error as e:
        print(f"Ошибка установки локали: {e}")

locale_info()
```

## 24. Более эффективные математические операции

```py
import math
import time

def benchmark_math_operations():
    """Бенчмарк математических операций"""
    
    operations = [
        ("sin", lambda x: math.sin(x)),
        ("cos", lambda x: math.cos(x)), 
        ("exp", lambda x: math.exp(x)),
        ("log", lambda x: math.log(x + 1)),
        ("sqrt", lambda x: math.sqrt(x + 1))
    ]
    
    # Тестовые данные
    test_data = [i * 0.1 for i in range(1000)]
    
    results = {}
    for op_name, op_func in operations:
        start_time = time.perf_counter()
        
        # Выполнение операции
        result = [op_func(x) for x in test_data]
        
        end_time = time.perf_counter()
        duration = end_time - start_time
        
        results[op_name] = {
            'duration': duration,
            'result_sample': result[:3]  # Первые 3 результата
        }
    
    # Вывод результатов
    print("Бенчмарк математических операций:")
    for op_name, info in sorted(results.items(), key=lambda x: x[1]['duration']):
        print(f"  {op_name:4}: {info['duration']:.4f} сек")

benchmark_math_operations()
```
