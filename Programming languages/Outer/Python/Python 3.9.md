# Python 3.9

Python 3.9 был выпущен 5 октября 2020 года

## 1. Объединение типов (Union Types) с `|`

Более простой и читаемый синтаксис для объединения типов вместо `Union`.

Было (до 3.9):
```python
from typing import Union

def handle_value(value: Union[int, str]) -> None:
    pass
```

Стало (в 3.9):
```python
def handle_value(value: int | str) -> None:
    pass
```

Работает и с isinstance/issubclass:
```python
# Проверка типа
if isinstance(some_var, int | str):
    pass

# Проверка наследования
print(issubclass(bool, int | str))  # True (bool подкласс int)
```

## 2. Встроенные типы коллекций в аннотациях (Generic Aliases)

Теперь можно использовать стандартные типы `list`, `dict`, `set` и т.д. вместо импортов из `typing`.

Было:
```python
from typing import List, Dict, Tuple, Optional

def process_data(items: List[str]) -> Dict[str, int]:
    pass
```

Стало:
```python
def process_data(items: list[str]) -> dict[str, int]:
    pass

# Также работает
def get_coordinates() -> tuple[float, float]:
    return (1.0, 2.0)

def find_user(id: int) -> str | None:  # Вместо Optional[str]
    return None
```

## 3. Новые строковые методы

Два очень полезных метода для работы со строками: `removeprefix(prefix)` и `removesuffix(suffix)`

Было:
```python
url = "https://example.com"
if url.startswith("https://"):
    clean_url = url[len("https://"):]

filename = "archive.tar.gz"
if filename.endswith(".gz"):
    name = filename[:-3]
```

Стало:
```python
url = "https://example.com"
clean_url = url.removeprefix("https://")

filename = "archive.tar.gz"
name = filename.removesuffix(".gz")
```

Методы безопасны - если префикс/суффикс отсутствует, возвращают исходную строку:
```python
print("test.py".removesuffix(".txt"))  # "test.py" (без ошибки)
```

## 4. Новые и улучшенные типы в `typing`

`Annotated` для метаданных типов

```python
from typing import Annotated

# Добавляем метаданные к типу
UserId = Annotated[int, "Идентификатор пользователя", "positive"]
```

`TypeAlias` для явного объявления псевдонимов

```python
from typing import TypeAlias

Vector: TypeAlias = list[float]
Matrix: TypeAlias = list[Vector]
```

## 5. Работа со словарями

Объединение словарей с `|` и `|=`

Было:
```python
dict1 = {"a": 1, "b": 2}
dict2 = {"b": 3, "c": 4}

# Объединение
merged = {**dict1, **dict2}  # {"a": 1, "b": 3, "c": 4}
```

Стало:
```python
dict1 = {"a": 1, "b": 2}
dict2 = {"b": 3, "c": 4}

# Объединение
merged = dict1 | dict2  # {"a": 1, "b": 3, "c": 4}

# Обновление на месте
dict1 |= dict2  # Теперь dict1 = {"a": 1, "b": 3, "c": 4}
```

## 6. Улучшения модуля `zoneinfo` (новый в 3.9)

Встроенная поддержка часовых поясов через базу данных IANA

```python
from datetime import datetime
from zoneinfo import ZoneInfo

# Создание времени с учетом часового пояса
dt = datetime(2023, 1, 1, 12, 0, tzinfo=ZoneInfo("Europe/Moscow"))
print(dt)  # 2023-01-01 12:00:00+03:00

# Доступные часовые пояса
import zoneinfo
print(list(zoneinfo.available_timezones())[:5])
```

## 7. Улучшения модуля `graphlib` (новый в 3.9)

Поддержка топологической сортировки "из коробки".

```python
from graphlib import TopologicalSorter

graph = {
    "курслы": {"носки", "ботинки"},
    "брюки": {"ботинки"},
    "рубашка": {"брюки", "галстук"},
    "галстук": {"пиджак"},
    "пиджак": {},
    "носки": {},
    "ботинки": {}
}

ts = TopologicalSorter(graph)
print(list(ts.static_order()))
# Правильный порядок одевания
# ['ботинки', 'носки', 'пиджак', 'брюки', 'курслы', 'галстук', 'рубашка']
```

## 8. Улучшения декораторов

Теперь декораторы могут сохранять метаданные оригинальной функции.

```python
import functools

def decorator(func):
    @functools.wraps(func)
    def wrapper(*args, **kwargs):
        return func(*args, **kwargs)
    return wrapper

# Теперь __annotations__, __doc__ и другие атрибуты
# сохраняются правильно без дополнительных усилий
```

## 9. Отмена контекстных переменных

Появилась возможность откатывать изменения в контекстных переменных.

```python
import contextvars

var = contextvars.ContextVar('var')
var.set('начальное значение')

token = var.set('новое значение')
print(var.get())  # 'новое значение'

# Откат к предыдущему значению
var.reset(token)
print(var.get())  # 'начальное значение'
```

## 10. Обновление парсеров

Python перешел с LL(1) парсера на более мощный PEG-парсер. Это внутреннее изменение, но оно:

* Позволяет использовать более сложные грамматики
* Упрощает дальнейшее развитие языка
* Убирает некоторые старые ограничения синтаксиса

## 11. Улучшения производительности

Оптимизация вызовов методов: Ускорены повторные вызовы методов
Улучшен `str.replace()`: Работает до 2x быстрее в некоторых случаях
Оптимизация `math` функций: Некоторые математические операции ускорены

## 12. Устаревания

Удалены устаревшие функции: Убраны некоторые давно устаревшие API
`collections.abc`: Окончательный переход от прямого импорта из `collections`

## 13. Улучшения работы с модулями и атрибутами

* `__file__` для замороженных (frozen) модулей
Теперь у модулей, встроенных в бинарник (через `freeze`), тоже есть атрибут `__file__`, что упрощает отладку.
    ```python
    import collections

    print(collections.__file__)  # Теперь работает даже для встроенных модулей
    ```

* `find_spec()` вместо `find_module()`

* Интерфейс импорта окончательно перешел на современные методы:
Было:
    ```python
    import importlib
    loader = importlib.find_loader('json')
    ```
    Стало:
    ```python
    import importlib
    spec = importlib.util.find_spec('json')
    ```

## 14. Улучшения словарей

Метод `update()` теперь принимает любые итерируемые объекты, содержащие пары ключ-значение.

Было:
```python
d = {}
pairs = [('a', 1), ('b', 2)]
d.update(dict(pairs))
```

Стало:
```python
d = {}
pairs = [('a', 1), ('b', 2)]
d.update(pairs)  # Напрямую!
```

## 15. Улучшения декораторов и интроспекции

`functools.cache` - упрощенный кэшировщик. Простой декоратор для мемоизации, аналог `lru_cache(maxsize=None)`.

```python
import functools

@functools.cache
def factorial(n):
    print(f"Вычисляю факториал {n}")
    return 1 if n <= 1 else n * factorial(n-1)

print(factorial(5))
print(factorial(5))  # Результат берется из кэша
```

Сохранение `__wrapped__` для декораторов с параметрами. Теперь `functools.wraps` правильно сохраняет ссылку на оригинальную функцию даже для параметризованных декораторов.

## 16. Улучшения работы с датами и временем

`astimezone()` для наивных datetime:
Метод `astimezone()` теперь может конвертировать наивные объекты datetime.

```python
from datetime import datetime
from zoneinfo import ZoneInfo

naive_dt = datetime(2023, 1, 1, 12, 0)
# Раньше: ошибка! Теперь: работает!
moscow_dt = naive_dt.astimezone(ZoneInfo("Europe/Moscow"))
```

Добавлены `removefold()` и `fold` атрибут: Для работы с неоднозначным временем при переходе на летнее время.

## 17. Улучшения математических функций

`math.lcm()` - наименьшее общее кратное

```python
import math

print(math.lcm(12, 18))  # 36
print(math.lcm(7, 13))   # 91
```

`math.nextafter()` и `math.ulp()`: Для точной работы с числами с плавающей точкой.

```python
import math

# Следующее представимое число после x в направлении y
print(math.nextafter(1.0, 2.0))  # 1.0000000000000002

# Единица в последнем разряде (Unit in the Last Place)
print(math.ulp(1.0))  # 2.220446049250313e-16
```

## 18. Улучшения модуля ast (Abstract Syntax Trees)

`ast.unparse()` - обратное преобразование AST в код
Можно получить исходный код из AST-дерева.

```python
import ast

code = "x = 1 + 2 * 3"
tree = ast.parse(code)

# Обратно в код!
regenerated = ast.unparse(tree)
print(regenerated)  # (x := (1 + (2 * 3))) - может немного отличаться
```

## 19. Улучшения модуля `asyncio`

`asyncio.to_thread()` - Упрощенный запуск синхронных функций в отдельных потоках.

```python
import asyncio
import time

def blocking_operation():
    time.sleep(2)
    return "Готово!"

async def main():
    # Запускаем блокирующую операцию в отдельном потоке
    result = await asyncio.to_thread(blocking_operation)
    print(result)

asyncio.run(main())
```

## 20. Улучшения модуля `multiprocessing`

`multiprocessing.shared_memory` теперь стабилен

```python
from multiprocessing import shared_memory
import array

# Создаем общую память
shm = shared_memory.SharedMemory(create=True, size=100)

# Используем как массив
arr = array.array('i', [1, 2, 3, 4, 5])
another_arr = array.array('i', [0] * 5)
another_arr.buffer_info = (shm.buf, 5)  # Пример использования
```

## 21. Улучшения модуля `json`

Поддержка `json.JSONEncoder` для большего количества типов. Улучшена сериализация различных типов данных

```python
import json
from decimal import Decimal
from fractions import Fraction

# Лучшая поддержка пользовательских типов
data = {
    'decimal': Decimal('3.14'),
    'fraction': Fraction(2, 3)
}

# Легче написать свой энкодер
class ExtendedEncoder(json.JSONEncoder):
    def default(self, obj):
        if isinstance(obj, Decimal):
            return float(obj)
        return super().default(obj)
```

## 22. Улучшения модуля `http`

`http.HTTPStatus` теперь имеет флаги. Можно проверять категории HTTP статусов.

```python
from http import HTTPStatus

status = HTTPStatus.OK
print(status.is_success)    # True
print(status.is_client_error) # False
print(status.is_server_error) # False
print(status.is_informational) # False
```

## 23. Улучшения безопасности

Улучшенная валидация для `ipaddress`. Более строгая проверка IP-адресов и сетей.

```python
import ipaddress

# Более строгая валидация
try:
    ipaddress.IPv4Address('999.999.999.999')
except ipaddress.AddressValueError as e:
    print(f"Некорректный адрес: {e}")
```

## 24. Более точные сообщения об ошибках

```python
# Более понятные сообщения SyntaxError
# Было: SyntaxError: invalid syntax
# Стало: SyntaxError: did you forget parentheses around the comprehension target?

[x for x in 1, 2, 3]  # Теперь яснее, что не так
```

## 25. Улучшения тестирования

Новые assertion в `unittest`

```python
import unittest

class TestExamples(unittest.TestCase):
    def test_comparisons(self):
        # Новые методы сравнения
        self.assertGreater(5, 3)
        self.assertLessEqual(2, 2)
        self.assertCountEqual([1, 2, 2], [2, 1, 2])  # Проверяет мультимножества
```
