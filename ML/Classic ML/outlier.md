# Выброс (outlier)

**Выброс (outlier)** — это наблюдение, которое сильно отличается от остальных. Но важно не статистическое определение, а влияние на модель:

* Для линейных моделей (MSE) — катастрофа (см. ниже)
* Для деревьев/бустинга — почти не страшно (они рубят по порогам)
* Для метрик — может полностью исказить оценку качества

Пример
```py
# Нормальные данные: цены квартир 3-8 млн
prices = [3.2, 4.1, 5.5, 6.0, 7.2, 4.8, 5.9]  # млн

# Добавляем один выброс
prices_with_outlier = [3.2, 4.1, 5.5, 6.0, 7.2, 4.8, 5.9, 500.0]  # 500 млн

# Среднее без выброса: 5.2 млн
# Среднее с выбросом: 67 млн (ошибка в 13 раз!)
```

## Как выбросы убивают разные модели

### Linear Regression / MSE — САМЫЙ УЯЗВИМЫЙ

```py
from sklearn.linear_model import LinearRegression
import numpy as np

# Нормальные данные
X_good = np.array([1,2,3,4,5,6,7,8]).reshape(-1,1)
y_good = np.array([2,4,6,8,10,12,14,16])

# Добавляем выброс
X_bad = np.array([1,2,3,4,5,6,7,8,100]).reshape(-1,1)
y_bad = np.array([2,4,6,8,10,12,14,16,1000])

model_good = LinearRegression().fit(X_good, y_good)
model_bad = LinearRegression().fit(X_bad, y_bad)

print(f"Без выброса: вес = {model_good.coef_[0]:.1f}")  # 2.0
print(f"С выбросом: вес = {model_bad.coef_[0]:.1f}")   # ~9.8 (почти в 5 раз больше!)
```

**MSE** возводит ошибку в квадрат. Ошибка 1000 → штраф 1 000 000. Модель "съедет" в сторону выброса, чтобы уменьшить этот огромный штраф, испортив предсказания для нормальных данных.

### Logistic Regression — ТОЖЕ УЯЗВИМА (но слабее)

Выброс может "перетянуть" разделяющую линию:
* Без выброса: линия разделяет синие и красные нормально
* С выбросом: линия "дергается", чтобы попытаться классифицировать аномалию

### Деревья / XGBoost / LightGBM — УСТОЙЧИВЫ

Деревья решений делят пространство по порогам. Выброс просто уйдет в отдельный лист и не испортит общую картину.

* Дерево просто создаст условие: if x > 50: предскажи 1000
* Остальные данные (x < 50) будут предсказываться нормально

Вывод для инженера: если в данных гарантированно есть выбросы, а модель линейная — нужно их обрабатывать до обучения.

## Как детектить выбросы (в коде для pipeline)

### 1. IQR (межквартильный размах) — самый простой
```py
def detect_outliers_iqr(df, column, multiplier=1.5):
    Q1 = df[column].quantile(0.25)
    Q3 = df[column].quantile(0.75)
    IQR = Q3 - Q1
    lower = Q1 - multiplier * IQR
    upper = Q3 + multiplier * IQR
    return (df[column] < lower) | (df[column] > upper)

# Использование
outliers = detect_outliers_iqr(df, 'income')
print(f"Найдено выбросов: {outliers.sum()}")
```

* `1.5` — умеренные выбросы (стандарт)
* `3.0` — экстремальные выбросы (почти аномалии)

### 2. Z-score (для нормального распределения)

```py
from scipy import stats

def detect_outliers_zscore(df, column, threshold=3):
    z_scores = np.abs(stats.zscore(df[column]))
    return z_scores > threshold

# threshold=3 означает: отклонение больше 3 сигм
```

### 3. Изолирующий лес (для многомерных выбросов)

```py
from sklearn.ensemble import IsolationForest

def detect_outliers_iforest(df, contamination=0.1):
    iso_forest = IsolationForest(contamination=contamination, random_state=42)
    outliers = iso_forest.fit_predict(df)
    return outliers == -1  # -1 означает выброс
```

### Итог

* **IQR** — быстрый, простой, для каждого признака отдельно
* **Z-score** — если данные нормально распределены (редкость)
* **Isolation Forest** — для сложных зависимостей между признаками

## Что делать с выбросами (стратегии для продакшена)

### 1. Удаление (если их мало)

```py
# Удаляем строки с выбросами
df_clean = df[~outliers_mask]
```

Когда: выбросов < 1-2%, и они точно ошибочные (баг в сборе данных).

### 2. Каппинг (Clipping/Capping) — Самый популярный

```py
def cap_outliers(df, column, lower_percentile=0.01, upper_percentile=0.99):
    lower = df[column].quantile(lower_percentile)
    upper = df[column].quantile(upper_percentile)
    df[column] = df[column].clip(lower, upper)
    return df

# Заменяем всё, что ниже 1-го персентиля, на 1-й персентиль
# Всё, что выше 99-го, на 99-й персентиль
```

**Пример**: доход 1 000 000 000 → заменяем на 10 000 000 (99-й персентиль).

### 3. Трансформация (логарифмирование)

```py
# Если распределение скошено (например, доход)
df['income_log'] = np.log1p(df['income'])  # log(1 + x)
```

Логарифм "сжимает" выбросы: 1000 → 6.9, 1 000 000 → 13.8. Разница уже не в 1000 раз, а в 2 раза.

### 4. Робастные модели (если чинить не хочется)

* Использовать `RobustScaler` вместо `StandardScaler`
* Использовать `HuberRegressor` вместо `LinearRegression` (MSE + MAE гибрид)
* Использовать деревья вместо линейных моделей

```py
from sklearn.linear_model import HuberRegressor

model = HuberRegressor(epsilon=1.35)  # epsilon: чем меньше, тем устойчивее к выбросам
model.fit(X, y)
```

## Мониторинг выбросов в продакшене

Это критично! Модель могла обучиться без выбросов, но в реальности они появились.

```py
# В сервисе FastAPI
def validate_no_outliers(features: dict):
    for col, value in features.items():
        # Проверяем по заранее сохраненным границам
        if value < lower_bounds[col] or value > upper_bounds[col]:
            # Логируем алерт!
            logger.warning(f"Outlier detected: {col}={value}")
            # Вариант 1: отвергаем запрос
            # raise HTTPException(400, "Invalid feature value")
            # Вариант 2: каппируем
            features[col] = np.clip(value, lower_bounds[col], upper_bounds[col])
    return features
```

## Итог

| Модель | Чувствительность к выбросам | Что делать
|--|--|--
| LinearRegression | Очень высокая | Каппинг / удаление / HuberRegressor
| LogisticRegression | Средняя | Каппинг / RobustScaler
| Ridge/Lasso | Высокая (чуть ниже, чем LinearRegression) | Каппинг
| Деревья / Random Forest | Низкая | Можно не чистить (но лучше проверить)
| XGBoost / LightGBM | Низкая | Можно не чистить
| KNN / SVM | Высокая | Обязательно масштабировать + каппинг

**Главное правило**: Для линейных моделей выбросы — это яд. Для деревьев — просто редкие значения.

Что обязательно знать инженеру:
* MSE-модели (линейные) страдают от выбросов сильно, деревья — почти нет
* Каппинг (clip по персентилям) — самый практичный метод обработки
* Нужно сохранять границы выбросов из трейна и применять их в продакшене
* Мониторить появление новых выбросов в реальных данных
