from nlp_model_training.training import NlpModelTraining

my_nlp_model = NlpModelTraining()

# Данные для обучения (пример)
TRAIN_DATA = [
    ("Иван Иванов из компании ООО Ромашка", {"entities": [(0, 12, "PERSON"), (24, 36, "ORG")]}),
    ("Москва — столица России", {"entities": [(0, 6, "GPE"), (20, 26, "GPE")]}),
    ("Apple был основан в 1976 году в Калифорнии", {"entities": [(0, 5, "ORG"), (20, 29, "DATE"), (32, 42, "LOC")]}),
    # добавьте больше данных
]

my_nlp_model.set_train_data(TRAIN_DATA)

my_nlp_model.training(16, "./my_model_ru")
