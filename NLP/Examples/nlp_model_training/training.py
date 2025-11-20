import spacy
from spacy.training import Example

class NlpModelTraining:
    def __init__(self, path_model:str = None):
        if (path_model is None):
            self.nlp = spacy.load("ru_core_news_sm") # default
        else:
            self.nlp = spacy.load(path_model)

    def set_train_data(self, train_data):
        self.TRAIN_DATA = train_data

    def training(self, n_iter:int, path_save_model:str = "./model_ru_custom"):
        if self.TRAIN_DATA == None :
            print("Нет данных для обучения")
            return

        # Обучение
        optimizer = self.nlp.resume_training()
        #n_iter = 16  # число итераций

        for itn in range(n_iter):
            losses = {}
            for text, annotations in self.TRAIN_DATA:
                doc = self.nlp.make_doc(text)
                example = Example.from_dict(doc, annotations)
                self.nlp.update([example], drop=0.2, losses=losses)
            print(f"Итерация {itn + 1}, потери: {losses}")

        # Сохранение дообученной модели
        self.nlp.to_disk(path_save_model)





# Добавление нового типа сущности (если нужно)
#if "NEW_ENTITY" not in nlp.get_pipe("ner").labels:
#    nlp.get_pipe("ner").add_label("NEW_ENTITY")

