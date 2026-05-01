# ML Monitoring

Полная карта метрик для ML-инженера

```
┌─────────────────────────────────────────────────────────────┐
│                    ML MONITORING STACK                      │
├─────────────────────────────────────────────────────────────┤
│  🔹 Level 1: Infrastructure (что вы назвали + дополнения)   │
│  🔹 Level 2: Data Quality (дрейфы + валидация)              │
│  🔹 Level 3: Model Performance (самое важное!)              │
│  🔹 Level 4: Business Impact (часто забывают)               │
└─────────────────────────────────────────────────────────────┘
```

## Level 1: Infrastructure Metrics

| Метрика | Что мониторить | Почему важно | Порог тревоги
|--|--|--|--
| Latency: p50 | Медианное время ответа | Базовый пользовательский опыт | > 2× baseline
| Latency: p95 | 95-й перцентиль | "Хвостовые" задержки для большинства | > 3× baseline
| Latency: p99 | 99-й перцентиль | Критичные случаи, деградация | > 5× baseline
| Memory (RSS/VMS) | Потребление памяти | Утечки, OOM-краши | > 80% лимита
| CPU Usage | Загрузка процессора | Боттленеки, масштабирование | > 70% постоянно
| Throughput (RPS/QPS) | Запросов в секунду | Нагрузка, автоскейлинг | Вне 20 от нормы
| Error Rate | % 5xx / таймаутов | Стабильность сервиса | > 1%
| GPU Memory/Util | Если модель на GPU | Дорогой ресурс, узкое место | > 90%

## Level 2: Data Quality Metrics

| Метод | Когда использовать | Пример порога
|--|--|--
| PSI (Population Stability Index) | Категориальные + бинированные числовые | PSI > 0.1 → warning, > 0.25 → critical
| KL Divergence | Распределения вероятностей | KL > 0.5 → alert
| Wasserstein Distance | Непрерывные признаки | > 20 от baseline
| Chi-Square Test | Категориальные, тест гипотез | p-value < 0.01 → drift

Дрейф предсказаний (Prediction Drift):
* Мониторьте распределение `y_pred` (для регрессии) или `P(class=1)` (для классификации)
* **Важно**: Дрейф предсказаний ≠ ухудшение качества! Модель может стабильно предсказывать, но данные изменились.

### Что вы могли упустить:

| Метрика | Зачем нужна
|--|--
| Missing Values Rate | Внезапные пропуски в фичах = поломка пайплайна
| Schema Validation | Новые/исчезнувшие колонки, изменение типов
| Out-of-Range Values | Признаки за пределами тренировочного диапазона
| Cardinality Changes | Для категорий: появление новых значений (OOV)

## Level 3: Model Performance Metrics (важное)

Это то, что напрямую отвечает за качество модели. Без этих метрик вы "летите вслепую".

**Если есть ground truth (размеченные данные приходят с задержкой)**

| Задача | Метрики | Частота проверки
|--|--|--
| Классификация | Accuracy, Precision, Recall, F1, ROC-AUC, PR-AUC | Ежедневно / при накоплении 1000+ новых лейблов
| Регрессия | MAE, RMSE, MAPE, R² | Ежедневно
| Ранжирование | NDCG@K, MAP@K, MRR | По запросу / еженедельно
| Детекция аномалий | Precision@K, Recall@K | Еженедельно

**Если ground truth нет (онлайн-инференс) — proxy-метрики**

| Метрика | Как считать | Что показывает
|--|--|--
| Prediction Entropy | `-sum(p * log(p))` для вероятностей | Уверенность модели: рост = модель "теряется"
| Confidence Score Distribution | Гистограмма `max(P(class))` | Сдвиг в сторону низких вероятностей = риск
| Disagreement Rate (если есть ансамбль) | % случаев, когда модели не согласны | Рост = нестабильность предсказаний
| Rule-Based Checks | Бизнес-правила: "если доход < 0 → ошибка" | Ловит логические аномалии

**Концептуальный дрейф (Concept Drift)**

Когда зависимость `P(y|X)` меняется, даже если распределение `X` стабильно.

Как детектить:
* **Performance Drop**: падение метрик на отложенных лейблах
* **Residual Analysis**: рост ошибок в определённых сегментах
* **Window Comparison**: сравнение метрик на скользящих окнах

## Level 4: Business Metrics

Модель может быть технически идеальной, но бесполезной для бизнеса.

| Метрика | Пример для кредитного скоринга | Как считать
|--|--|--
| Approval Rate | % одобренных заявок | `sum(pred=1) / total`
| Default Rate | % невозвратов среди одобренных | Требует задержки 3-12 мес
| Profit per Prediction | `(доход от хороших) - (потери от плохих)` | Бизнес-логика + модель
| Conversion Rate | % клиентов, завершивших целевое действие | A/B-тест: модель vs контрольная группа
| Cost of False Positive/Negative | Финансовая цена ошибок | `FP * cost_FP + FN * cost_FN`

### Почему бизнес-метрики важнее технических

```
Сценарий: Кредитный скоринг

Технически: 
- AUC упал с 0.85 → 0.83 (незначительно)
- Инженер: "всё ок, в пределах шума"

Бизнес-реальность:
- Из-за сдвига в данных модель стала чаще одобрять "рисковых"
- Default Rate вырос с 5% → 8%
- Потери: +$2M/месяц

Вывод: мониторьте бизнес-метрики даже с задержкой
```

## Критические алерты (что настраивать в первую очередь)

```yaml
# priority: CRITICAL ( PagerDuty / SMS )
- name: "Model Error Rate > 5%"
  condition: error_rate_5xx > 0.05 for 5m
  
- name: "Latency p99 > 5s"
  condition: latency_p99 > 5.0 for 10m
  
- name: "Prediction Drift (PSI) > 0.25"
  condition: feature_psi_max > 0.25

- name: "AUC Drop > 0.05"
  condition: current_auc < baseline_auc - 0.05  # при наличии лейблов

# priority: WARNING ( Slack / Email )
- name: "Memory Usage > 80%"
  condition: memory_percent > 80 for 15m
  
- name: "Missing Values Spike"
  condition: missing_rate > baseline + 3*std
  
- name: "New Categorical Values Detected"
  condition: oov_rate > 0.01
```

## Чеклист мониторинга для ML-инженера

Обязательный минимум (Day 1)
* Latency: p50, p95, p99
* Error rate (5xx, таймауты)
* Memory / CPU usage
* Prediction distribution drift (PSI / KS-test)
* Missing values rate

Продвинутый уровень (Week 1)
* Feature drift по ключевым признакам
* Schema validation (новые/исчезнувшие фичи)
* Model confidence / entropy monitoring
* Business metric proxy (approval rate, etc.)

Production-ready (Month 1)
* Ground-truth based metrics (AUC, RMSE) с задержкой
* Concept drift detection (window comparison)
* Cost-based alerting (FP/FN стоимость)
* Автоматический rollback при деградации
* Dashboard (Grafana / Kibana / custom)

## Инструменты для реализации

| Задача | Инструменты
|--|--
| Metrics Collection | Prometheus, StatsD, OpenTelemetry
| Drift Detection | Evidently AI, NannyML, WhyLabs, Alibi Detect
| Dashboards | Grafana, Kibana, Streamlit, Dash
| Alerting | Alertmanager, PagerDuty, Slack webhooks
| ML-specific | MLflow, Weights & Biases, Arize, Fiddler
| Data Validation | Great Expectations, Pydantic, TensorFlow Data Validation

## Итоговая стратегия

```
Правило 80/20 для мониторинга:

1. Начните с 5 критических алертов:
   - p99 latency > threshold
   - error_rate > 1%
   - PSI > 0.25 для 3 ключевых фич
   - missing_rate spike
   - prediction distribution shift

2. Добавьте бизнес-метрики с задержкой:
   - Даже если лейблы приходят через месяц — считайте AUC постфактум

3. Автоматизируйте реакцию:
   - Падение качества → переключиться на бейзлайн-модель
   - Дрейф данных → триггерить переобучение
   - Технические сбои → autoscaling / restart

4. Документируйте baseline:
   - "Нормальные" значения для каждой метрики
   - Кто отвечает за каждый алерт
   - Процедура эскалации
```

**Главный принцип**: Мониторьте не то, что легко измерить, а то, что больно пропустить.
