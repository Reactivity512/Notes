import spacy
from collections import Counter
from wordcloud import WordCloud
import matplotlib.pyplot as plt

class NlpModel:
    def __init__(self, path_model:str):
        if (path_model == None):
            self.nlp = spacy.load("ru_core_news_sm") # default
        else:
            self.nlp = spacy.load(path_model)

    def set_text(self, text:str):
        self.doc = self.nlp(text)
        self.__tokenization()

    def __tokenization(self):
        # Фильтруем токены: только слова (не пунктуация, не пробелы, не стоп-слова, не числа)
        self.words = [
            token.lemma_.lower()   # приводит слово к нормальной форме (лемматизация) и нижнему регистру. Если нужно учитывать форму слова как есть (без лемматизации), замени token.lemma_ на token.text.
            for token in self.doc
            if not token.is_stop   # флаг стоп-слова (включая кастомные).
            and not token.is_punct # исключает пунктуацию (точки, запятые и т.д.).
            and not token.is_space #  исключает пробелы и переносы строк.
            and token.is_alpha     # только буквы (исключаем числа и символы)
        ]

    def print_pos_tagg(self):
        print("=== POS Tagging ===")
        for token in self.doc:
            print(f"{token.text:<12} {token.pos_:<8} {spacy.explain(token.pos_)}")

    def print_all_entities(self):
        print("Именованные сущности (NER):")
        for ent in self.doc.ents:
            print(f"Текст: '{ent.text}' | Метка: {ent.label_} | Описание: {spacy.explain(ent.label_)}")

    def print_most_frequent_words(self, number_of_top:int = 5):
        # Подсчёт частот
        word_freq = Counter(self.words)

        most_common = word_freq.most_common(number_of_top)
        print(f"Самые частые слова: {most_common}")

    def print_verbs(self):
        print("Verbs (Глаголы):", [token.lemma_ for token in self.doc if token.pos_ == "VERB"])

    def print_tokens_and_lemmas(self):
        for token in self.doc:
            print(f"{token.text} -> {token.lemma_}")

    def print_dependency_parsing(self):
        print("=== Базовый синтаксический разбор ===")
        for token in self.doc:
            print(f"{token.text:<10} {token.dep_:<15} {token.head.text:<10} {spacy.explain(token.dep_)}")

    def print_dependency_parsing_tree(self):
        for token in self.doc:
            if token.dep_ == "ROOT":
                print(f"ROOT: {token.text}")
                self._print_children(token, "")

    def _print_children(self, token, prefix):
        children = list(token.children)
        for i, child in enumerate(children):
            is_last = i == len(children) - 1
            connector = "└── " if is_last else "├── "
            print(f"{prefix}{connector}{child.text} [{child.dep_}]")
            extension = "    " if is_last else "│   "
            self._print_children(child, prefix + extension)

    def print_verb_arguments(self):
        verb_arguments = {}
        
        for token in self.doc:
            if token.pos_ == "VERB":
                arguments = {
                    'subject': [],
                    'objects': [],
                    'modifiers': []
                }
                
                for child in token.children:
                    if child.dep_ == "nsubj":
                        arguments['subject'].append(child.text)
                    elif child.dep_ == "obj":
                        arguments['objects'].append(child.text)
                    elif child.dep_ in ["advmod", "obl"]:
                        arguments['modifiers'].append((child.text, child.dep_))
                
                verb_arguments[token.text] = arguments
        
        print(f"\n=== Аргументы глаголов ===")
        for verb, args in verb_arguments.items():
            print(f"Глагол: {verb}")
            print(f"  Подлежащее: {args['subject']}")
            print(f"  Дополнения: {args['objects']}")
            print(f"  Обстоятельства: {args['modifiers']}")

    def print_find_action_target(self, action_word):
        target = self._find_action_target(action_word)
        if target is None:
            print(f"action_word не найден")
        else:
            print(f"Ответ: {target}")

    def _find_action_target(self, action_word):
        """Найти объект действия в предложении"""
        for token in self.doc:
            print(token.text.lower())
            if token.text.lower() == action_word.lower() and token.pos_ == "VERB":
                # Ищем прямое дополнение у глагола
                for child in token.children:
                    if child.dep_ == "obj":
                        return child.text
        return None

    def visualization_word_сloud(self, path_save:str = None):
        # Преобразуем Counter в словарь {слово: частота}
        freq_dict = dict(Counter(self.words))
        try:
            wordcloud = WordCloud(
                width=800,
                height=400,
                background_color='white',
                colormap='viridis',
                max_words=100,
                relative_scaling=0.5,
                # font_path='/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf'  # раскомментируй при ошибке с кириллицей
                # или Linux (Debian/Ubuntu): font_path='/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf'
                # или macOS: font_path='/System/Library/Fonts/Supplemental/Arial.ttf'
                # Windows (часто работает без указания): font_path='C:/Windows/Fonts/arial.ttf'
            ).generate_from_frequencies(freq_dict)
        except UnicodeDecodeError:
            print("Ошибка с кодировкой шрифта. Укажите font_path с поддержкой кириллицы.")

        # Отображение
        plt.figure(figsize=(12, 6))
        plt.imshow(wordcloud, interpolation='bilinear')
        plt.axis("off")
        plt.title("Облако слов", fontsize=20)
        plt.tight_layout()
        plt.show()

        # Для сохранения
        if path_save != None:
            wordcloud.to_file(path_save + "wordcloud.png")

    def visualization_matplotlib_bar_chart(self, number_of_top:int = 5):
        # Подсчёт частот
        word_freq = Counter(self.words)

        most_common = word_freq.most_common(number_of_top)

        # Разделяем слова и частоты для графика
        words_top, freqs_top = zip(*most_common)  # распаковка списка кортежей

        # === Визуализация с matplotlib ===
        plt.figure(figsize=(10, 6))
        bars = plt.barh(words_top, freqs_top, color='steelblue')

        # Добавляем значения на концах столбцов (опционально)
        for bar, freq in zip(bars, freqs_top):
            plt.text(bar.get_width() + 0.05, bar.get_y() + bar.get_height()/2,
                     str(freq), va='center', fontsize=10)

        # Настройки графика
        plt.xlabel("Частота", fontsize=12)
        plt.title("Самые частые слова", fontsize=14)
        plt.gca().invert_yaxis()  # чтобы самые частые были наверху
        plt.tight_layout()
        plt.show()