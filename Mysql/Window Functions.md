# Оконные функции (Window Functions)

**Оконные функции (Window Functions)** - это функции, который позволяют выполнять вычисления над набором строк, связанных с текущей строкой, не группируя строки как это делает **GROUP BY**.
С MySQL 8.0 оконные функции поддерживаются нативно.
Для определения окна используется **OVER()**. Синтаксис оконной функции:
```
<функция>() OVER (
   [PARTITION BY <столбец>]
   [ORDER BY <столбец>]
   [ROWS BETWEEN ...]
)
```
**PARTITION BY** — разбивает строки на группы (похожие на **GROUP BY**, но сохраняет строки).
**ORDER BY** — задаёт порядок строк внутри каждой группы.
**ROWS / RANGE** — задаёт рамки окна (необязательно).

## Основные типы оконных функций

### 1. Агрегатные функции как оконные:  **SUM(), AVG(), COUNT(), MIN(), MAX().**

Примеры:
| id | region | month | revenue |
| --- | --- | --- | --- |
| 1 | North | Jan | 1000 |
| 2 | North | Feb | 1200 |
| 3 | North | Mar | 1300 |
| 4 | South | Jan | 800 |
| 5 | South | Feb | 950 |
| 6 | South | Mar | 1100 |

Выполняем запрос:

```
SELECT region, month, revenue,
    SUM(revenue) OVER() AS total_revenue,
    SUM(revenue) OVER(PARTITION BY region) AS total_region_revenue,
    SUM(revenue) OVER(PARTITION BY month) AS total_month_revenue,
    AVG(revenue) OVER(PARTITION BY region) AS avg_region_revenue
FROM sales;
```

Результат:

| region | month | revenue | total_revenue | total_region_revenue |total_month_revenue | avg_region_revenue |
| --- | --- | --- | --- | --- | --- | --- |
| North | Jan | 1000 | 6350 | 3500 | 1800 | 1166.6667 |
| North | Feb | 1200 | 6350 | 3500 | 2150 | 1166.6667 |
| North | Mar | 1300 | 6350 | 3500 | 2400 | 1166.6667 |
| South | Jan | 800 | 6350 | 2850 | 1800 | 950.0000 |
| South | Feb | 950 | 6350 | 2850 | 2150 | 950.0000 |
| South | Mar | 1100 | 6350 | 2850 | 2400 | 950.0000 |

**total_revenue** - общая сумма за все месяцы и все регионы.
**total_region_revenue** - общая сумма за все месяцы по регионам.
**total_month_revenue** - общая сумма за месяц независимо от региона. 
**avg_region_revenue** - средняя сумма по регионам.

Если нужно посчитать сумму revenue по текущей дате и всем предыдущим, можно воспользоваться **RANGE** или **ROWS**

| id | sale_date | revenue |
| --- | --- | --- |
| 1 | 2024-01-01 | 100 |
| 2 | 2024-01-01 | 200 |
| 3 | 2024-01-02 | 150 |
| 4 | 2024-01-02 | 120 |
| 5 | 2024-01-03 | 220 |
| 6 | 2024-01-04 | 140 |

Выполняем запрос с **RANGE**:
```
SELECT sale_date, revenue,
   SUM(revenue) OVER (
       ORDER BY sale_date
       RANGE BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
   ) AS cumulative_sum
FROM sales;
```
Результат:

| sale_date | revenue | cumulative_sum |
| --- | --- | --- |
| 2024-01-01 | 100 | 300 |
| 2024-01-01 | 200 | 300 |
| 2024-01-02 | 150 | 570 |
| 2024-01-02 | 120 | 570 |
| 2024-01-03 | 220 | 790 |
| 2024-01-04 | 140 | 930 |

* **`ORDER BY sale_date`** сортирует строки по дате.
* **`RANGE BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW`** включает все строки с датами ≤ текущей.
* Все строки с **одинаковыми датами** считаются в одном «уровне».

С использованием **ROWS**:
```
SELECT sale_date, revenue,
    SUM(revenue) OVER (
        ORDER BY sale_date
        ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
    ) AS cumulative_sum
FROM sales;
```
Результат:

| sale_date | revenue | cumulative_sum |
| --- | --- | --- |
| 2024-01-01 | 100 | 100 |
| 2024-01-01 | 200 | 300 |
| 2024-01-02 | 150 | 450 |
| 2024-01-02 | 120 | 570 |
| 2024-01-03 | 220 | 790 |
| 2024-01-04 | 140 | 930 |

Здесь каждая **строка** считается по отдельности. Даже строки с одинаковой датой идут последовательно.


### 2. Ранжирующие функции: **ROW_NUMBER() ,RANK(), DENSE_RANK().**

Пример:

| department | name | salary |
| --- | --- | --- |
| HR | Марина | 5000 |
| HR | Екатерина | 6000 |
| HR | Алиса | 6000 |
| IT | Александр | 8000 |
| IT | Николай | 8000 |
| IT | Анна | 7000 |

Выполняем запрос:
```
SELECT department, name, salary,
   ROW_NUMBER() OVER (PARTITION BY department ORDER BY salary DESC) AS row_num,
   RANK()       OVER (PARTITION BY department ORDER BY salary DESC) AS rank_num,
   DENSE_RANK() OVER (PARTITION BY department ORDER BY salary DESC) AS dense_rank_num
FROM employees;
```

Результат:

| department | name | salary | row_num | rank_num | dense_rank_num |
| --- | --- | --- | --- | --- | --- |
| HR | Екатерина | 6000 | 1 | 1 | 1 |
| HR | Алиса | 6000 | 2 | 1 | 1 |
| HR | Марина | 5000 | 3 | 3 | 2 |
| IT | Александр | 8000 | 1 | 1 | 1 |
| IT | Анна | 7000 | 2 | 2 | 2 |
| IT | Николай | 6000 | 3 | 3 | 3 |



* **`ROW_NUMBER()`** — уникальный номер строки в порядке сортировки. Дубли не учитываются.
* **`RANK()`** — одинаковый ранг для одинаковых значений, но пропускает следующий номер.
* **`DENSE_RANK()`** — как `RANK()`, но **не пропускает** номера после повторяющихся значений.

### 3. Функции смещения: **LAG(), LEAD(), FIRST_VALUE(), LAST_VALUE().**

Пример:

| department | name | salary |
| --- | --- | --- |
| HR | Марина | 5000 |
| HR | Екатерина | 6000 |
| HR | Алиса | 6000 |
| IT | Александр | 8000 |
| IT | Анна | 7000 |
| IT | Николай | 6000 |
| Finance | Петр | 9000 |
| Finance | Александра | 10000 |
| Finance | Владимир | 12000 |
| Finance | Сергей | 10000 |

Выполняем запрос:
```
SELECT department, name, salary,
   LAG(salary)  OVER (PARTITION BY department ORDER BY salary DESC) AS prev_salary,
   LEAD(salary) OVER (PARTITION BY department ORDER BY salary DESC) AS next_salary,
   FIRST_VALUE(name) OVER (PARTITION BY department ORDER BY salary DESC) AS top_earner,
   LAST_VALUE(name)  OVER (
       PARTITION BY department
       ORDER BY salary DESC
       ROWS BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING
   ) AS lowest_earner
FROM employees;
```


В `ROWS BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING` -

* **`ROWS`** — говорит, что мы определяем окно по **строкам**, а не по значениям (`RANGE`) или группам.
* **`UNBOUNDED PRECEDING`** — от самого начала окна (самой первой строки в разделе).
* **`UNBOUNDED FOLLOWING`** — до самого конца окна (последней строки в разделе).

Результат:

| department | name | salary | prev_salary | next_salary | top_earner | lowest_earner |
| --- | --- | --- | --- | --- | --- | --- |
| Finance | Владимир | 12000 | null | 10000 | Владимир | Петр |
| Finance | Александра | 10000 | 12000 | 10000 | Владимир | Петр |
| Finance | Сергей | 10000 | 10000 | 9000 | Владимир | Петр |
| Finance | Петр | 9000 | 10000 | null | Владимир | Петр |
| HR | Екатерина | 6000 | null | 6000 | Екатерина | Марина |
| HR | Алиса | 6000 | 6000 | 5000 | Екатерина | Марина |
| HR | Марина | 5000 | 6000 | null | Екатерина | Марина |
| IT | Александр | 8000 | null | 7000 | Александр | Николай |
| IT | Анна | 7000 | 8000 | 6000 | Александр | Николай |
| IT | Николай | 6000 | 7000 | null | Александр | Николай |

* **`LAG(salary)`** — предыдущая зарплата в пределах отдела.
* **`LEAD(salary)`** — следующая зарплата в пределах отдела.
* **`FIRST_VALUE(name)`** — имя сотрудника с наивысшей зарплатой в отделе.
* **`LAST_VALUE(name)`** — имя сотрудника с самой низкой зарплатой.

Если у сотрудников одинаковая минимальная зарплата, то в **lowest_earner** будет отображаться последний сотрудник с минимальной зарплатой. С **top_earner** ситуация аналогичная, последний сотрудник с топ зарплатой.

### 4. Функции распределения: PERCENT_RANK(), CUME_DIST(), NTILE(n).

Примеры:

| name | department | salary |
| --- | --- | --- |
| Ivan P | IT | 85000 |
| Mariy S | HR | 65000 |
| Alexey I | IT | 95000 |
| Elena K | HR | 70000 |
| Dmitry S | IT | 110000 |
| Olga V | Marketing | 60000 |
| Sergey F | Marketing | 75000 |
| Anna P | IT | 90000 |
| Nikolay S | HR | 68000 |
| Tatina M | Marketing | 80000 |

Выполняем запрос (Для **PERCENT_RANK()**):

```
SELECT name, department, salary,
   PERCENT_RANK() OVER(ORDER BY salary DESC) AS percent_rank_overall,
   PERCENT_RANK() OVER(PARTITION BY department ORDER BY salary DESC) AS percent_rank_dept
FROM employees
ORDER BY salary DESC;
```

Результат:

| name | department | salary | percent_rank_overall | percent_rank_dept |
| --- | --- | --- | --- | --- |
| Dmitry S | IT | 110000 | 0 | 0 |
| Alexey I | IT | 95000 | 0.1111111111111111 | 0.3333333333333333 |
| Anna P | IT | 90000 | 0.2222222222222222 | 0.6666666666666666 |
| Ivan P | IT | 85000 | 0.3333333333333333 | 1 |
| Sergey F | Marketing | 80000 | 0.4444444444444444 | 0 |
| Tatina M | Marketing | 80000 | 0.4444444444444444 | 0 |
| Elena K | HR | 70000 | 0.6666666666666666 | 0 |
| Nikolay S | HR | 68000 | 0.7777777777777778 | 0.5 |
| Mariy S | HR | 65000 | 0.8888888888888888 | 1 |
| Olga V | Marketing | 60000 | 1 | 1 |

* **`PERCENT_RANK() OVER(ORDER BY salary DESC)`** - показывает позицию зарплаты среди всех сотрудников (0.0 - самая высокая, 1.0 - самая низкая)
* **`PERCENT_RANK() OVER(PARTITION BY department ORDER BY salary DESC)`** - то же самое, но внутри каждого отдела

Выполняем запрос (Для **CUME_DIST()**):

```
SELECT emp_name, department, salary,
   CUME_DIST() OVER(ORDER BY salary) AS cume_dist_overall,
   CUME_DIST() OVER(PARTITION BY department ORDER BY salary) AS cume_dist_dept
FROM employees
ORDER BY salary;
```

Результат:

| name | department | salary | cume_dist_overall | cume_dist_dept |
| --- | --- | --- | --- | --- |
| Olga V | Marketing | 60000 | 0.1 | 0.3333333333333333 |
| Mariy S | HR | 65000 | 0.2 | 0.3333333333333333 |
| Nikolay S | HR | 68000 | 0.3 | 0.6666666666666666 |
| Elena K | HR | 70000 | 0.4 | 1 |
| Sergey F | Marketing | 80000 | 0.6 | 1 |
| Tatina M | Marketing | 80000 | 0.6 | 1 |
| Ivan P | IT | 85000 | 0.7 | 0.25 |
| Anna P | IT | 90000 | 0.8 | 0.5 |
| Alexey I | IT | 95000 | 0.9 | 0.75 |
| Dmitry S | IT | 110000 | 1 | 1 |

* **`CUME_DIST()`** показывает кумулятивное распределение
* Кумулятивное распределение (Cumulative Distribution Function, CDF) — это статистическая функция, которая показывает долю наблюдений, значения которых меньше или равны текущему значению. Вычисляется: **CUME_DIST()** = (количество строк со значениями ≤ текущему значению) / (общее количество строк)

Выполняем запрос (Для **NTILE(n)**):

```
SELECT name, department, salary,
   NTILE(3) OVER(ORDER BY salary DESC) AS salary_tier_overall,
   NTILE(2) OVER(PARTITION BY department ORDER BY salary DESC) AS salary_tier_dept
FROM employees
ORDER BY salary DESC;
```

Результат:

| name | department | salary | salary_tier_overall | salary_tier_dept |
| --- | --- | --- | --- | --- |
| Dmitry S | IT | 110000 | 1 | 1 |
| Alexey I | IT | 95000 | 1 | 1 |
| Anna P | IT | 90000 | 1 | 2 |
| Ivan P | IT | 85000 | 1 | 2 |
| Sergey F | Marketing | 80000 | 2 | 1 |
| Tatina M | Marketing | 80000 | 2 | 1 |
| Elena K | HR | 70000 | 2 | 1 |
| Nikolay S | HR | 68000 | 3 | 1 |
| Mariy S | HR | 65000 | 3 | 2 |
| Olga V | Marketing | 60000 | 3 | 2 |

* **`NTILE(3) OVER(ORDER BY salary DESC)`** - делит всех сотрудников на 3 группы по зарплате (1 - высокая, 2 - средняя, 3 - низкая)
* **`NTILE(2) OVER(PARTITION BY department ORDER BY salary DESC)`** - делит сотрудников каждого отдела на 2 группы по зарплате

---

### Преимущества оконных функций 
* Устраняют необходимость в сложных подзапросах
* Повышают читаемость SQL-кода
* Позволяют выполнять сложные аналитические расчеты
* Более эффективны, чем аналогичные решения с JOIN или подзапросами

### Недостатки оконных функций 
* Доступны только в MySQL 8.0 и выше
* Не могут использоваться в WHERE, GROUP BY или HAVING
* Могут влиять на производительность при неправильном использовании
