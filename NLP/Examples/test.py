import spacy
import pytest

# Загрузите модель (замените на вашу модель)
MODEL_PATH = "./model_ru_custom"

@pytest.fixture
def nlp():
    return spacy.load(MODEL_PATH)

def test_model_loaded(nlp):
    # Проверяем, что модель содержит компонент NER
    assert "ner" in nlp.pipe_names

def test_prediction_entities(nlp):
    # Проверяем, что модель распознает сущности в тестовом тексте
    text = "Apple был основан в 1976 году в Калифорнии."
    doc = nlp(text)
    entities = [(ent.text, ent.label_) for ent in doc.ents]
    # Проверяем наличие конкретной сущности
    assert any(ent[1] == "ORG" and "Apple" in ent[0] for ent in entities)
    assert any(ent[1] == "LOC" and "Калифорнии" in ent[0] for ent in entities)
    assert any(ent[1] == "DATE" and "1976" in ent[0] for ent in entities)

def test_no_entities_for_empty_text(nlp):
    # Тестирование на пустом или неинформативном тексте
    doc = nlp("")
    assert len(doc.ents) == 0

def test_model_output_types(nlp):
    text = "Этот тест помогает проверить модель."
    doc = nlp(text)
    for ent in doc.ents:
        assert isinstance(ent.text, str)
        assert isinstance(ent.label_, str)