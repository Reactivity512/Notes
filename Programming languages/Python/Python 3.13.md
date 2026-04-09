# Python 3.13

Python 3.13 был выпущен 7 октября 2024 года

## 1. Экспериментальный режим без GIL

Добавлена сборочная опция `--disable-gil`, позволяющая компилировать Python без глобальной блокировки интерпретатора.
**Пока экспериментально: часть расширений C ещё не совместима.**

## 2. Экспериментальный JIT-компилятор

Добавлен экспериментальный JIT-компилятор, который может быть включен с флагом `--enable-experimental-jit`:

```bash
# Компиляция Python с JIT
./configure --enable-experimental-jit
make
```

```py
# Код автоматически получает выгоду от JIT
def calculate_sum(n):
    result = 0
    for i in range(n):  # JIT оптимизирует этот цикл
        result += i * i
    return result
```

Дальнейшие оптимизации "Faster CPython"

* Ускорение вызовов функций на 10-20%
* Более быстрая работа с атрибутами
* Оптимизация создания объектов


## 3. Новый синтаксис и возможности

* Улучшенные f-strings

```py
# Многострочные f-strings с комментариями и отступами
message = f"""
    Пользователь: {user.name}
    Возраст: {user.age}
    Статус: {"активен" if user.active else "неактивен"}  # inline условие
    Баланс: ${user.balance:,.2f}
"""
```

* Упрощенный паттерн-матчинг

```py
# Более краткий синтаксис для простых случаев
match value:
    case int(x) if x > 0:
        print(f"Положительное число: {x}")
    case [x, y, *rest]:
        print(f"Список начинается с {x}, {y}")
```

## 3. Улучшения системы типов

* Новые типы в typing:

```py
from typing import TypeVar, reveal_type

# Улучшенный TypeVar с синтаксисом границ
T = TypeVar('T', bound=Union[int, str])
U = TypeVar('U', default=int)

class Container[T]:
    def get_first(self) -> T: ...

# Функция для отладки типов
def process_data(data: list[int]) -> None:
    reveal_type(data)  # Показывает тип во время проверки
```

* Статические типы в рантайме

```py
import typing

# Аннотации теперь более доступны в runtime
def process(user_id: int, name: str) -> bool:
    pass

annotations = typing.get_type_hints(process)
print(annotations)  # {'user_id': <class 'int'>, 'name': <class 'str'>, 'return': <class 'bool'>}
```

## 4. Улучшения стандартной библиотеки

* Новый модуль `pathlib2`:

```py
from pathlib2 import Path  # Улучшенная версия pathlib

# Новые методы и улучшенная производительность
path = Path("data.txt")
if path.exists():
    content = path.read_text(encoding='utf-8')
    new_path = path.with_stem("backup")  # Новый метод
```

* Улучшения `asyncio`:

```py
import asyncio

# Упрощенное создание асинхронных приложений
async def main():
    async with asyncio.TaskGroup() as tg:
        task1 = tg.create_task(fetch_data(url1))
        task2 = tg.create_task(process_data(data))
    
    # Автоматическая отмена при ошибках
    results = await asyncio.gather(*tasks, return_exceptions=True)
```

* Новый модуль `statistics2`

```py
import statistics2

# Расширенная статистическая функциональность
data = [1, 2, 3, 4, 5, 6, 7, 8, 9]

# Новые функции
rolling_mean = statistics2.rolling_mean(data, window=3)
outliers = statistics2.detect_outliers(data)
```

## 5. Улучшения для разработчиков

* Улучшенные сообщения об ошибках:

```py
# Еще более точные указания на ошибки
def example():
    data = {
        'name': 'Alice',
        'age': 30
        'city': 'Moscow'  # ← Python точно укажет на забытую запятую
    }
```

```
  File "example.py", line 4
    'city': 'Moscow'
    ^
SyntaxError: missing comma before 'city'
```

* Улучшенная интроспекция:

```py
import inspect

# Более детальная информация о функциях и классах
def sample_function(x: int, y: str = "hello") -> bool:
    """Пример функции"""
    return True

# Новая функциональность
print(inspect.get_annotations(sample_function, eval_str=True))
print(inspect.get_source(sample_function))
```

## 6. Улучшения безопасности

* Улучшенная валидация входных данных

```py
import urllib.parse

# Более строгая проверка URL
def safe_url_parse(url):
    parsed = urllib.parse.urlparse(url)
    if parsed.scheme not in ('http', 'https', 'ftp'):
        raise ValueError(f"Недопустимая схема: {parsed.scheme}")
    return parsed
```

* Новые криптографические функции

```py
import hashlib
import secrets

# Новые алгоритмы и улучшенная безопасность
data = b"sensitive data"

# Улучшенные хеш-функции
hash_obj = hashlib.sha3_512(data)
secure_token = secrets.token_hex(32)
```

## 7. Улучшения для научных вычислений

* Новые математические функции

```py
import math

# Дополнительные математические функции
x = 2.5
print(math.log2(x))      # Логарифм по основанию 2
print(math.exp2(x))      # 2 в степени x
print(math.cbrt(27))     # Кубический корень
```

* Улучшения для работы с массивами

```py
import array

# Более эффективные операции с массивами
arr = array.array('i', [1, 2, 3, 4, 5])

# Новые методы
squared = arr.apply(lambda x: x * x)  # Поэлементное применение
filtered = arr.filter(lambda x: x > 2)  # Фильтрация
```

## 8. Улучшения межплатформенной совместимости

* Улучшенная поддержка Windows

```py
import os
import sys

# Лучшая работа с путями на разных платформах
if sys.platform == "win32":
    # Улучшенная обработка путей Windows
    config_path = os.path.expanduser("~/AppData/Local/MyApp")
else:
    config_path = os.path.expanduser("~/.config/myapp")

# Универсальное создание директорий
os.makedirs(config_path, exist_ok=True, mode=0o755)
```

## 9. Улучшения для пакетного менеджмента

* Улучшения `pip` и виртуальных окружений

```py
# Встроенные улучшения для управления зависимостями
import importlib.util
import sys

# Более надежная загрузка модулей
def load_module_from_path(path):
    spec = importlib.util.spec_from_file_location("module.name", path)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module
```

## 10. Удаления и устаревания

Удалены устаревшие API, включая:
* некоторые функции из `collections`
* устаревшие методы в `datetime`
* старые форматы в `email` модуле

## 11. Улучшения для отладки и профилирования

* Новые инструменты отладки

```py
import sys
import traceback

# Улучшенная трассировка
def debug_function():
    try:
        # Код, который может вызвать ошибку
        result = 1 / 0
    except Exception:
        # Более информативный вывод
        exc_type, exc_value, exc_tb = sys.exc_info()
        tb_list = traceback.format_exception(exc_type, exc_value, exc_tb)
        print("Детальная трассировка:")
        for line in tb_list:
            print(line.strip())
```

