# Семантические задачи NLP (уровень смысла)

Эти задачи направлены на понимание смысла текста.

## Анализ тональности (Sentiment Analysis)

Анализ тональности (Sentiment Analysis): Определение эмоциональной окраски текста (позитивный, негативный, нейтральный) и/или выявления конкретных эмоций.

**Основные уровни анализа:**
*  Бинарная классификация: `Позитивный` vs `Негативный`
`Пример: "Отличный товар!" → Позитивный`

* Трехклассовая классификация: `Позитивный` vs `Негативный` vs `Нейтральный`
`Пример: "Заказ доставлен" → Нейтральный`

* Многоклассовая классификация: Конкретные эмоции: радость, гнев, грусть, удивление и т.д.
`Пример: "Это просто ужасно!" → Гнев`

* Анализ на уровне аспектов: Определение тональности по конкретным аспектам
`Пример: "Камера отличная, но батарея слабая" -> Камера: Позитивный, Батарея: Негативный`

Пример:

```py
from transformers import pipeline, AutoTokenizer, AutoModelForSequenceClassification
import torch

# Загрузка модели для русского языка
sentiment_pipeline = pipeline(
    "sentiment-analysis",
    model="blanchefort/rubert-base-cased-sentiment",
    tokenizer="blanchefort/rubert-base-cased-sentiment"
)

texts = [
    "Это просто прекрасный продукт!",
    "Ужасное качество, никогда больше не куплю.",
    "Товар доставлен вовремя."
]

results = sentiment_pipeline(texts)
for text, result in zip(texts, results):
    print(f"Текст: {text}")
    print(f"Тональность: {result['label']} (score: {result['score']:.3f})")
    print("---")
```

Пример создание собственной модели:

```py
import pandas as pd
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.linear_model import LogisticRegression
from sklearn.model_selection import train_test_split
from sklearn.metrics import classification_report
import joblib

# Пример создания простой модели
def create_sentiment_model():
    # Пример данных для обучения (в реальности нужен большой датасет)
    data = {
        'text': [
            "очень хороший товар", "плохое качество", "нормально", 
            "отлично работает", "ужасный сервис", "доволен покупкой",
            "разочарован", "прекрасный результат", "кошмарная доставка"
        ],
        'sentiment': ['positive', 'negative', 'neutral', 'positive', 
                     'negative', 'positive', 'negative', 'positive', 'negative']
    }
    
    df = pd.DataFrame(data)
    
    # Векторизация текста
    vectorizer = TfidfVectorizer(max_features=1000)
    X = vectorizer.fit_transform(df['text'])
    y = df['sentiment']
    
    # Разделение на train/test
    X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.3, random_state=42)
    
    # Обучение модели
    model = LogisticRegression()
    model.fit(X_train, y_train)
    
    # Оценка качества
    y_pred = model.predict(X_test)
    print(classification_report(y_test, y_pred))
    
    return model, vectorizer

# Создание и использование модели
model, vectorizer = create_sentiment_model()

# Функция для предсказания
def predict_sentiment(text, model, vectorizer):
    text_vectorized = vectorizer.transform([text])
    prediction = model.predict(text_vectorized)[0]
    probability = model.predict_proba(text_vectorized).max()
    return prediction, probability

# Тестирование
test_texts = ["хороший продукт", "плохое качество", "обычный товар"]
for text in test_texts:
    pred, prob = predict_sentiment(text, model, vectorizer)
    print(f"'{text}' → {pred} (вероятность: {prob:.2f})")
```

Пример обучения (train.py):

```py
from transformers import pipeline, AutoTokenizer, AutoModelForSequenceClassification, TrainingArguments, Trainer
import torch
import pandas as pd
import numpy as np
from datasets import Dataset, DatasetDict
from sklearn.metrics import accuracy_score, f1_score
from train_steps import data_preprocessing, tokenize_function, compute_metrics, model_name, tokenizer

dataset = data_preprocessing()
tokenized_datasets = dataset.map(tokenize_function, batched=True)

# Загрузка модели и настройка обучения
model = AutoModelForSequenceClassification.from_pretrained(model_name, num_labels=3)
training_args = TrainingArguments(
    output_dir="./results",
    eval_strategy="epoch",
    save_strategy="epoch",
    learning_rate=2e-5,
    per_device_train_batch_size=16,
    per_device_eval_batch_size=16,
    num_train_epochs=3,
    weight_decay=0.01,
    logging_dir="./logs",
    load_best_model_at_end=True,
    metric_for_best_model="f1",
)
trainer = Trainer(
    model=model,
    args=training_args,
    train_dataset=tokenized_datasets["train"],
    eval_dataset=tokenized_datasets["validation"],
    compute_metrics=compute_metrics,
)

trainer.train()

trainer.save_model("./sentiment-model-rus")
tokenizer.save_pretrained("./sentiment-model-rus")

```

Пример обучения (train_steps.py):

```py
from transformers import AutoTokenizer
import pandas as pd
import numpy as np
from datasets import Dataset, DatasetDict
from sklearn.metrics import accuracy_score, f1_score

#model_name = "cointegrated/rubert-tiny2" # лёгкая и быстрая модель. Мини-BERT (4 слоя, ~4 млн параметров)
#model_name = "DeepPavlov/rubert-base-cased"  # Более мощная, но тяжелее. Полноценный BERT (12 слоёв, ~110 млн параметров)

model_name = "cointegrated/rubert-tiny2"
tokenizer = AutoTokenizer.from_pretrained(model_name)

def data_preprocessing():
    # Загрузка данных
    df = pd.read_csv("sentiment_dataset.csv")

    # Разделение на train/val
    train_df = df.sample(frac=0.8, random_state=42)
    val_df = df.drop(train_df.index)

    # Преобразование в Dataset
    train_dataset = Dataset.from_pandas(train_df)
    val_dataset = Dataset.from_pandas(val_df)

    dataset = DatasetDict({"train": train_dataset, "validation": val_dataset})

    return dataset


def tokenize_function(examples):
    return tokenizer(examples["text"], truncation=True, padding="max_length", max_length=128)


def compute_metrics(eval_pred):
    predictions, labels = eval_pred
    predictions = np.argmax(predictions, axis=1)
    return {
        "accuracy": accuracy_score(labels, predictions),
        "f1": f1_score(labels, predictions, average="weighted")
    }

```

Пример логов при обучении:

```bash
{'loss': 0.534, 'grad_norm': 7.5901265144348145, 'learning_rate': 2.327342835502307e-06, 'epoch': 2.65}
{'loss': 0.5509, 'grad_norm': 7.354257106781006, 'learning_rate': 2.0978218458077992e-06, 'epoch': 2.69}
{'loss': 0.5394, 'grad_norm': 11.516979217529297, 'learning_rate': 1.8683008561132918e-06, 'epoch': 2.72}
{'loss': 0.5327, 'grad_norm': 14.716002464294434, 'learning_rate': 1.638779866418784e-06, 'epoch': 2.75}
{'loss': 0.535, 'grad_norm': 24.410781860351562, 'learning_rate': 1.4092588767242766e-06, 'epoch': 2.79}
{'loss': 0.5547, 'grad_norm': 14.547464370727539, 'learning_rate': 1.179737887029769e-06, 'epoch': 2.82}
{'loss': 0.5377, 'grad_norm': 10.393383979797363, 'learning_rate': 9.502168973352615e-07, 'epoch': 2.86}
{'loss': 0.5561, 'grad_norm': 12.771370887756348, 'learning_rate': 7.206959076407538e-07, 'epoch': 2.89}
{'loss': 0.5281, 'grad_norm': 14.245198249816895, 'learning_rate': 4.911749179462461e-07, 'epoch': 2.93}
{'loss': 0.5419, 'grad_norm': 4.636935234069824, 'learning_rate': 2.6165392825173864e-07, 'epoch': 2.96}
{'loss': 0.5503, 'grad_norm': 15.734369277954102, 'learning_rate': 3.213293855723106e-08, 'epoch': 3.0}
{'eval_loss': 0.608680248260498, 'eval_accuracy': 0.7336466294842663, 'eval_f1': 0.7342447685287531, 'eval_runtime': 429.1121, 'eval_samples_per_second': 135.377, 'eval_steps_per_second': 8.462, 'epoch': 3.0}
```
* **loss (потери)**: 
    * Это значение функции потерь (например, cross-entropy) на текущем батче.
    * Чем ниже, тем лучше модель предсказывает.
    * В идеале: постепенно уменьшается как на `train`, так и на `validation`.

В примере `0.5377 → 0.5561 → 0.5281 → 0.5419 → 0.5503` → Нет чёткого тренда вниз. Колеблется около 0.53–0.5.

Это нормально на поздних этапах обучения, особенно если:

* модель почти сошлась,
* `learning rate` уже очень мал,
* данные шумные или задача сложная.

Но если `loss` растёт стабильно — это тревожный сигнал (переобучение или нестабильность).


* **grad_norm (норма градиента)**
    * Это норма (длина) вектора градиентов всех параметров модели.
    * Показывает, насколько сильно модель «обновляется» на этом шаге.
    * Очень высокая → возможен градиентный взрыв.
    * Очень низкая → обучение почти остановилось.

В примере `10.4 → 12.8 → 14.2 → 4.6 → 15.7` → Сильные колебания, особенно скачок до 15.7 на последнем шаге.

В BERT-подобных моделях норма градиента часто колеблется, особенно если:

* `batch` маленький,
* данные разнородные (например, короткие и длинные отзывы),
* используется `gradient clipping`

Но если `grad_norm` постоянно > 20–30 — стоит уменьшить `learning rate` или включить `clipping`.

* **learning_rate (скорость обучения)**
    * Это стандартная практика: сначала быстро учимся, потом «тонко настраиваем».

В примере `9.5e-7 → 7.2e-7 → 4.9e-7 → 2.6e-7 → 3.2e-8` → LR стремится к нулю к концу последней эпохи.

Это нормально. К 3-й эпохе LR почти нулевой — модель перестаёт сильно меняться, «доводит» последние детали.

* **epoch**
    * Показывает, сколько эпох пройдено.
    * от 2.86 до 3.0 → завершилась 3-я эпоха.

Итог:
| Метрика | Состояние | Комментарий |
| --- | --- | --- |
| Loss | Стабилен (~0.54) | Не уменьшается, но и не растёт → модель, скорее всего, сошлась |
| Grad norm | Колеблется | Нормально для последних шагов; нет взрыва >30 |
| Learning rate | -> 0 | Ожидаемо при `decay` |
| Эпоха | Завершилась 3-я | Обучение закончено num_train_epochs=3 |

Вывод: обучение прошло штатно, модель, вероятно, достигла оптимума для заданных условий.

Проверка качества модели:

```py
metrics = trainer.evaluate()
print(metrics)
```

Вывод:
```bash
{'eval_loss': 0.608680248260498, 'eval_model_preparation_time': 0.0016, 'eval_accuracy': 0.7336466294842663, 'eval_f1': 0.7342447685287531, 'eval_runtime': 588.0224, 'eval_samples_per_second': 98.792, 'eval_steps_per_second': 6.175}
```

**Accuracy** ≈ 73.4%
**F1** ≈ 73.4%
**Validation loss** ≈ 0.61

Важные метрики:

* **Precision** (Точность) = TP / (TP + FP)
* **Accuracy** (Общая точность) = (TP + TN) / (TP + TN + FP + FN)
* **Recall** (Полнота) = TP / (TP + FN)
* **F1** = 2 × (Precision × Recall) / (Precision + Recall)

**TP (True Positive)**: Модель правильно предсказала положительный класс
**FP (False Positive)**: Модель неправильно предсказала положительный класс (ложное срабатывание)
**FN (False Negative)**: Модель неправильно предсказала отрицательный класс (пропуск)
**TN (True Negative)**: Модель правильно предсказала отрицательный класс

Какой F1 считается хорошим на проде?
| Контекст | Реалистичный F1 | Комментарий |
| --- | --- | --- |
| Общие отзывы (маркетплейсы, рестораны) | 80–88% | Стандарт для большинства продуктов |
| Узкая тематика (например, только отзывы на ноутбуки) | 85–92% | Меньше шума → выше точность |
| Бинарная классификация (pos/neg, без neutral) | 88–93% | Проще задача → выше метрики |
| Лабораторные условия / research (SOTA) | до 94–95% | На идеальных датасетах |

## Классификация текстов (Text Classification)

Классификация текстов — это задача автоматического присвоения текстовым документам одной или нескольких предопределенных категорий (меток, классов).

Основные типы классификации:

1. Бинарная классификация. (2 класса)
Примеры: `Спам/Не спам`

2. Многоклассовая классификация (Несколько классов, один вариант)
Примеры: `Тема новости (спорт, политика, экономика), Язык текста`

3. Многометочная классификация (Несколько классов, несколько вариантов)
Примеры: `Теги статей (машинное обучение, python, nlp), Жанры фильмов`

Примеры:

* Спам-фильтр: Письмо "Выиграй миллион сейчас!" → Спам.
* Тематическая классификация: Новость "Курс акций Apple вырос" → Категория: "Финансы".

Выбор метода: мало данных → традиционные ML, много данных → нейросети

Пример с традиционным методом ML (Подход с мешком слов (Bag-of-Words))

```py
from sklearn.feature_extraction.text import CountVectorizer
from sklearn.naive_bayes import MultinomialNB
from sklearn.pipeline import Pipeline

# Создание пайплайна
model = Pipeline([
    ('vectorizer', CountVectorizer()),  # Преобразование текста в числа
    ('classifier', MultinomialNB())     # Наивный Байес
])

# Пример обучения
texts = ["отличный товар", "ужасное качество", "нормально"]
labels = ["positive", "negative", "neutral"]

model.fit(texts, labels)
```

## Семантический поиск (Semantic Search)

**Семантический поиск** — это поиск информации на основе смысла запроса, а не просто совпадения ключевых слов.

**Традиционный поиск** -> Совпадение слов. `"яблоко"` → документы со словом `"яблоко"`.
**Семантический поиск** -> Совпадение смыслов. `"яблоко"` → документы про фрукты, компанию Apple, здоровое питание.

```py
# Векторное представление текстов
query = "способы приготовления курицы"
document = "рецепты блюд из птицы"

# Традиционный поиск: совпадений слов нет → не найдет
# Семантический поиск: векторы близки → найдет!
```

### Векторные представления (Embeddings):

**Статические эмбеддинги (Word2Vec, GloVe)**
```py
import gensim
from gensim.models import Word2Vec

# Обучение Word2Vec модели
sentences = [
    ["кот", "ловит", "мышь"],
    ["собака", "бежит", "за", "кошкой"], 
    ["птица", "летит", "в", "небе"]
]

model = Word2Vec(sentences, vector_size=100, window=5, min_count=1)

# Получение векторов слов
vector_cat = model.wv["кот"]
vector_dog = model.wv["собака"]
```

**Контекстуальные эмбеддинги (BERT, ELMo)**
```py
from sentence_transformers import SentenceTransformer

# Загрузка предобученной модели для русского
model = SentenceTransformer('sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2')

# Создание эмбеддингов для предложений
sentences = [
    "Кот ловит мышь",
    "Собака бежит за кошкой",
    "Программист пишет код"
]

embeddings = model.encode(sentences)
print(f"Размерность эмбеддингов: {embeddings.shape}")
```

### Методы семантического поиска

**Dense Retrieval (Плотный поиск)**
```py
import numpy as np
from sklearn.metrics.pairwise import cosine_similarity

class DenseSemanticSearch:
    def __init__(self, model_name='sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2'):
        self.model = SentenceTransformer(model_name)
        self.documents = []
        self.embeddings = None
        
    def index_documents(self, documents):
        """Индексация документов"""
        self.documents = documents
        self.embeddings = self.model.encode(documents)
        
    def search(self, query, top_k=5):
        """Поиск по запросу"""
        query_embedding = self.model.encode([query])
        
        # Вычисление косинусной близости
        similarities = cosine_similarity(query_embedding, self.embeddings)[0]
        
        # Сортировка по релевантности
        results = []
        for idx in np.argsort(similarities)[::-1][:top_k]:
            results.append({
                'document': self.documents[idx],
                'score': similarities[idx],
                'rank': len(results) + 1
            })
            
        return results

# Пример использования
documents = [
    "Рецепт приготовления курицы в духовке",
    "Как варить куриный суп", 
    "Программирование на Python для начинающих",
    "Блюда из птицы: индейка и курица",
    "Машинное обучение и искусственный интеллект"
]

searcher = DenseSemanticSearch()
searcher.index_documents(documents)

results = searcher.search("способы готовки мяса птицы")
for result in results:
    print(f"Рейтинг: {result['rank']}, Сходство: {result['score']:.3f}")
    print(f"Документ: {result['document']}\n")
```

Современные метрики:
* nDCG (Normalized Discounted Cumulative Gain)
* MAP (Mean Average Precision)
* MRR (Mean Reciprocal Rank)

Для русского языка: `sentence-transformers/paraphrase-multilingual-mpnet-base-v2`

## Определение схожести текстов (Text Similarity)

**Text Similarity** — это количественная мера того, насколько два текста семантически близки друг к другу по смыслу, а не просто по совпадению слов.

Ключевые аспекты схожести:
* Лексическая схожесть - совпадение слов
* Семантическая схожесть - совпадение смысла
* Синтаксическая схожесть - совпадение структуры
* Стилистическая схожесть - совпадение стиля

Пример: "Как приготовить пиццу?" и "Рецепт неаполитанской пиццы" → Высокая схожесть.


Уровни схожести текстов:
* Символьный уровень (Character-level)
```py
# Сравнение последовательностей символов
text1 = "кот"
text2 = "код"
# Совпадение: "ко" (2 из 3 символов)
```

* Лексический уровень (Word-level)
```py
# Сравнение слов и их совпадений
text1 = "кот ловит мышь"
text2 = "кошка ловит мышку" 
# Совпадение: "ловит" (1 из 3 слов)
```

* Семантический уровень (Semantic-level)
```py
# Сравнение смысла
text1 = "кот охотится на грызуна"
text2 = "кошка ловит мышь"
# Высокая семантическая схожесть despite разными словами
```

Методы вычисления схожести:
1. Строковые методы (String-based)
    Расстояние Левенштейна (Levenshtein Distance)
```py
def levenshtein_distance(s1, s2):
    """Расстояние Левенштейна - минимальное количество операций для преобразования строк"""
    if len(s1) < len(s2):
        return levenshtein_distance(s2, s1)
    
    if len(s2) == 0:
        return len(s1)
    
    previous_row = range(len(s2) + 1)
    for i, c1 in enumerate(s1):
        current_row = [i + 1]
        for j, c2 in enumerate(s2):
            insertions = previous_row[j + 1] + 1
            deletions = current_row[j] + 1
            substitutions = previous_row[j] + (c1 != c2)
            current_row.append(min(insertions, deletions, substitutions))
        previous_row = current_row
    
    return previous_row[-1]

def levenshtein_similarity(s1, s2):
    """Нормализованная схожесть на основе Левенштейна"""
    distance = levenshtein_distance(s1, s2)
    max_len = max(len(s1), len(s2))
    if max_len == 0:
        return 1.0
    return 1 - (distance / max_len)

# Пример
text1 = "кот"
text2 = "кит"
print(f"Расстояние Левенштейна: {levenshtein_distance(text1, text2)}")
print(f"Схожесть: {levenshtein_similarity(text1, text2):.3f}")
```

Jaccard Similarity
```py
def jaccard_similarity(text1, text2):
    """Коэффициент Жаккара для множеств слов"""
    words1 = set(text1.lower().split())
    words2 = set(text2.lower().split())
    
    intersection = len(words1.intersection(words2))
    union = len(words1.union(words2))
    
    if union == 0:
        return 1.0
    return intersection / union

# Пример
text1 = "кот ловит мышь"
text2 = "кошка ловит мышку"
print(f"Схожесть Жаккара: {jaccard_similarity(text1, text2):.3f}")
```

2. Статистические методы
    TF-IDF + Косинусная схожесть

3. Нейросетевые методы (современные)
    Sentence Transformers
    BERT-based Similarity

**Выбирайте строковые методы когда:**
* Тексты короткие и структурно похожие
* Важны точные совпадения слов
* Вычислительные ресурсы ограничены

**Выбирайте статистические методы когда:**
* Тексты средней длины
* Есть достаточное количество данных для TF-IDF
* Нужен баланс между качеством и скоростью

**Выбирайте нейросетевые методы когда:**
* Тексты сложные и разнообразные
* Важен семантический смысл, а не лексика
* Вычислительные ресурсы позволяют
* Требуется высокая точность
