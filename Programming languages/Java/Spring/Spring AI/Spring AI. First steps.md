# Spring AI: первые шаги

## Подготовка

Убедись, что Ollama запущена и модель скачана:
```bash
ollama pull qwen2.5:3b
```

В `application.properties`:
```properties
spring.ai.ollama.chat.options.model=qwen2.5:3b
spring.ai.ollama.base-url=http://localhost:11434

# Опционально: температура (0.0 - детерминировано, 1.0 - креативно)
spring.ai.ollama.chat.temperature=0.7
```

> Можно использовать любую другую модель — имя должно совпадать с выводом ollama list.

## 1. AI Hello World

Создай контроллер:
```java
package com.example.demo;

import org.springframework.ai.chat.client.ChatClient;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class AiController {

    private final ChatClient chatClient;

    public AiController(ChatClient.Builder chatClientBuilder) {
        this.chatClient = chatClientBuilder.build();
    }

    @GetMapping("/ai/hello")
    public String aiHello() {
        // Формируем промпт
        String prompt = "Скажи 'Привет, мир!' и добавь один короткий интересный факт о языке Java.";

        // Вызываем модель и получаем ответ
        return chatClient.prompt()
                .user(prompt)
                .call()
                .content(); // .content() извлекает просто строку ответа из объекта ChatResponse
    }
}
```

Откройте `http://localhost:8080/ai/hello`.

**Заметки:**
* Первый запрос может занять 2–5 секунд — модель загружается в память.
* Ошибка `404 Model not found` — проверь точное имя модели в `ollama list`.
* `.call().content()` — синхронный вызов. Для чата с печатающимся текстом используйте `.stream()` (см. раздел 5).


## 2. Структурированный вывод (JSON → Java Record)

Spring AI умеет возвращать не просто текст, а готовый Java-объект. Создайте record:

```java
package com.example.demo;

public record RoutingIntent(
    String targetService,
    String action,
    String userId,
    int priority
) {}
```

Добавьте эндпоинт в AiController.java:
```java
@GetMapping("/ai/route")
public RoutingIntent routeRequest(@RequestParam String text) {

    // ... твой старый метод /ai/hello ...
    
    String systemPrompt = """
        Ты умный API-шлюз. Проанализируй запрос пользователя
        и извлеки данные для маршрутизации.
        Если поле неизвестно, оставь его пустым (null или 0).
        """;

    return chatClient.prompt()
            .system(systemPrompt)
            .user(text)
            .call()
            .entity(RoutingIntent.class);
}
```

Пример: `http://localhost:8080/ai/route?text=Отключи уведомления для пользователя user-999`

Ответ — готовый JSON, сериализованный Spring. Никакого ручного парсинга.

## 3. Function Calling (LLM вызывает Java-код)

Дайте модели «инструменты» — Java-методы, помеченные `@Tool`. Если пользователь спрашивает "Какой статус сервиса оплаты?", LLM не выдумывает ответ, а говорит Spring AI: "Эй, вызови метод `checkServiceStatus("payment")`, и я использую его результат".

Создай класс `SystemTools.java`. Аннотация `@Tool` делает метод видимым для LLM.

```java
package com.example.demo;

import org.springframework.ai.chat.model.ToolContext;
import org.springframework.stereotype.Service;
import org.springframework.ai.tool.annotation.Tool;
import org.springframework.ai.tool.annotation.ToolParam;

@Service
public class SystemTools {

    @Tool(description = "Проверяет текущий статус и задержку (в мс) указанного микросервиса")
    public String getServiceStatus(@ToolParam(description = "Имя сервиса, например: 'payment', 'auth', 'database'") String serviceName) {
        // Здесь может быть реальный вызов к Prometheus, ClickHouse или твоему прокси-логгеру
        System.out.println(">>> LLM вызвала Java-метод для сервиса: " + serviceName);
        
        return switch (serviceName.toLowerCase()) {
            case "payment" -> "Статус: ОК, средняя задержка 45мс.";
            case "auth" -> "Статус: ДЕГРАДАЦИЯ, задержка 300мс, есть таймауты.";
            default -> "Статус: Неизвестный сервис или нет данных.";
        };
    }
}
```

Обнови `AiController.java`, добавив зависимость от `SystemTools`:

```java
package com.example.demo;

import org.springframework.ai.chat.client.ChatClient;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class AiController {

    private final ChatClient chatClient;

    // Инжектим наши инструменты
    public AiController(ChatClient.Builder chatClientBuilder, SystemTools systemTools) {
        this.chatClient = chatClientBuilder
                .defaultTools(systemTools) // <--- вот ключевая строка
                .build();
    }

    @GetMapping("/ai/tool")
    public String askSystem(@RequestParam String question) {
        return chatClient.prompt()
                .user(question)
                .call()
                .content();
    }
}
```

Запрос: `http://localhost:8080/ai/tool?question=Какой сейчас статус сервиса payment и auth?`

LLM сама решит вызвать `getServiceStatus("payment")`, Spring AI выполнит метод, и модель сформирует ответ на основе результата.

Что произойдет "под капотом":
1. LLM видит вопрос и понимает, что не знает ответа.
2. Она видит твой `@Tool` и генерирует специальный запрос: "Вызови getServiceStatus с аргументом payment".
3. Spring AI перехватывает это, вызывает твой реальный Java-метод.
4. В консоли ты увидишь: >>> LLM вызвала Java-метод для сервиса: payment
5. LLM получает результат ("Статус: ОК...") и формирует финальный человеческий ответ.

## 4. Prompt Templates (динамические промпты)

Тебе нужно формировать промпты динамически, подставляя данные из контекста (имя пользователя, параметры запроса, результаты из БД). Хардкодить промпты — плохая практика. Spring AI позволяет создавать шаблоны промптов с переменными, как в Thymeleaf или Mustache.

Добавь новый эндпоинт в `AiController.java`:

```java
package com.example.demo;

import org.springframework.ai.chat.client.ChatClient;
import org.springframework.ai.chat.prompt.PromptTemplate;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.util.Map;

@RestController
public class AiController {

    private final ChatClient chatClient;
    private final SystemTools systemTools;

    public AiController(ChatClient.Builder chatClientBuilder, SystemTools systemTools) {
        this.chatClient = chatClientBuilder
                .defaultTools(systemTools)
                .build();
    }

    // ... твои старые методы /ai/hello, /ai/route, /ai/tool ...

    @GetMapping("/ai/template")
    public String generateWithTemplate(
            @RequestParam String userName,
            @RequestParam String serviceName,
            @RequestParam String action
    ) {
        // Создаем шаблон промпта с переменными
        String templateText = """
            Ты системный администратор. Пользователь {userName} хочет выполнить действие "{action}" 
            над сервисом {serviceName}. 
            Предоставь краткую инструкцию на русском языке, как это сделать безопасно.
            """;

        // Заполняем переменные
        PromptTemplate promptTemplate = new PromptTemplate(templateText);
        String finalPrompt = promptTemplate.render(Map.of(
                "userName", userName,
                "serviceName", serviceName,
                "action", action
        ));

        // Отправляем готовый промпт в модель
        return chatClient.prompt()
                .user(finalPrompt)
                .call()
                .content();
    }
}
```

Пример: `http://localhost:8080/ai/template?userName=Иван&serviceName=payment&action=перезапустить`

Ответ:
```
Иван, для безопасной перезапуска сервиса payment выполните следующие шаги:
Убедитесь, что нет активных транзакций...
Выполните команду...
Проверьте логи...
```

Шаблоны можно хранить в отдельных файлах (`.txt`, `.prompt`) и версионировать отдельно от кода.

## 5. Streaming (потоковая отдача ответов)

Если твой прокси общается с пользователем в реальном времени (чат-бот, голосовой ассистент), ждать 2-3 секунды полный ответ — плохой UX. Streaming позволяет отдавать ответ по токенам, как только они генерируются. Для чат-интерфейсов используйте `.stream()` и `Flux<String>`. Spring Boot автоматически конвертирует это в Server-Sent Events (SSE).

Добавьте зависимость:
```xml
<dependency>
    <groupId>org.springframework.boot</groupId>
    <artifactId>spring-boot-starter-webflux</artifactId>
</dependency>
```

Добавляем `resources/static/stream-test.html`:
```html
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>AI Stream Test</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 800px;
            margin: 50px auto;
            padding: 20px;
        }
        #question {
            width: 70%;
            padding: 10px;
            font-size: 16px;
        }
        button {
            padding: 10px 20px;
            font-size: 16px;
            cursor: pointer;
        }
        #response {
            margin-top: 20px;
            padding: 15px;
            border: 1px solid #ccc;
            border-radius: 8px;
            min-height: 150px;
            background: #f9f9f9;
            white-space: pre-wrap;
            font-size: 16px;
            line-height: 1.6;
        }
        .cursor {
            display: inline-block;
            width: 2px;
            height: 18px;
            background: black;
            animation: blink 0.7s infinite;
            vertical-align: text-bottom;
        }
        @keyframes blink {
            0%, 100% { opacity: 1; }
            50% { opacity: 0; }
        }
    </style>
</head>
<body>
    <h2>🤖 AI Streaming Chat</h2>
    <div>
        <input type="text" id="question" placeholder="Введите вопрос..." value="Расскажи короткий факт о Java">
        <button onclick="startStream()">Отправить</button>
    </div>
    <div id="response"></div>

    <script>
        function startStream() {
            const question = document.getElementById('question').value;
            const responseDiv = document.getElementById('response');
            
            // Очищаем предыдущий ответ
            responseDiv.innerHTML = '<span class="cursor"></span>';

            // Используем fetch + ReadableStream (надёжнее, чем EventSource)
            fetch('/ai/stream?question=' + encodeURIComponent(question))
                .then(response => {
                    const reader = response.body.getReader();
                    const decoder = new TextDecoder('utf-8');

                    function read() {
                        reader.read().then(({ done, value }) => {
                            if (done) {
                                // Убираем курсор, когда поток закончился
                                const cursor = responseDiv.querySelector('.cursor');
                                if (cursor) cursor.remove();
                                return;
                            }

                            // Декодируем байты в текст
                            const chunk = decoder.decode(value, { stream: true });
                            
                            // Парсим SSE: каждая строка "data:..." 
                            const lines = chunk.split('\n');
                            let text = '';
                            for (const line of lines) {
                                if (line.startsWith('data:')) {
                                    // Убираем префикс "data:" и добавляем к тексту
                                    text += line.substring(5);
                                }
                            }

                            // Вставляем текст перед курсором
                            const cursor = responseDiv.querySelector('.cursor');
                            if (cursor) {
                                cursor.insertAdjacentText('beforebegin', text);
                            } else {
                                responseDiv.textContent += text;
                            }

                            // Читаем следующий чанк
                            read();
                        });
                    }

                    read();
                })
                .catch(error => {
                    responseDiv.textContent = 'Ошибка: ' + error.message;
                });
        }
    </script>
</body>
</html>
```

Добавляем streaming-эндпоинт:
```java
@GetMapping(value = "/ai/stream", produces = MediaType.TEXT_EVENT_STREAM_VALUE + ";charset=UTF-8")
public Flux<String> streamResponse(@RequestParam String question) {
    return chatClient.prompt()
            .user(question)
            .stream()
            .content();
}
```

Открой в браузере: `http://localhost:8080/stream-test.html`

Ответ будет появляться постепенно, слово за словом, как в ChatGPT. Это работает через SSE (Server-Sent Events) — браузер получает поток данных и рендерит их в реальном времени.

## Подключаем встроенный логгер

Добавляем:
``` java
import org.springframework.ai.chat.client.advisor.SimpleLoggerAdvisor;

...

public AiController(ChatClient.Builder chatClientBuilder, com.example.demo.SystemTools systemTools) {
        this.chatClient = chatClientBuilder
                .defaultTools(systemTools)
                .defaultAdvisors(new SimpleLoggerAdvisor()) // <--- Подключаем встроенный логгер
                .build();
    }

...
```

## Полный AiController:
```java
package com.example.demo;

import org.springframework.ai.chat.client.ChatClient;
import org.springframework.ai.chat.prompt.PromptTemplate;
import org.springframework.http.MediaType;
import org.springframework.web.bind.annotation.CrossOrigin;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.ai.chat.client.advisor.SimpleLoggerAdvisor;
import reactor.core.publisher.Flux;

import java.util.*;

@RestController
@CrossOrigin(origins = "*") // <--- ВАЖНО: разрешаем CORS на уровне контроллера
public class AiController {

    private final ChatClient chatClient;

    // Spring автоматически предоставит нам Builder, настроенный на основе application.properties
    // Инжектим наши инструменты
    public AiController(ChatClient.Builder chatClientBuilder, com.example.demo.SystemTools systemTools) {
        this.chatClient = chatClientBuilder
                .defaultTools(systemTools) // <--- ВОТ КЛЮЧЕВАЯ СТРОКА
                .defaultAdvisors(new SimpleLoggerAdvisor()) // <--- Подключаем встроенный логгер
                .build();
    }

    @GetMapping("/ai/hello")
    public String aiHello() {
        // Формируем промпт
        String prompt = "Скажи 'Привет, мир!' и добавь один короткий интересный факт о языке Java.";

        // Вызываем модель и получаем ответ
        return chatClient.prompt()
                .user(prompt)
                .call()
                .content(); // .content() извлекает просто строку ответа из объекта ChatResponse
    }

    @GetMapping("/ai/route")
    public com.example.demo.model.RoutingIntent routeRequest(@RequestParam String text) {
        // Системный промпт задает строгие правила поведения
        String systemPrompt = """
            Ты умный API-шлюз. Твоя задача - проанализировать запрос пользователя 
            и извлечь из него данные для маршрутизации. 
            Если поле неизвестно, оставь его пустым (null или 0).
            """;

        return chatClient.prompt()
                .system(systemPrompt)
                .user(text)
                .call()
                .entity(com.example.demo.model.RoutingIntent.class); // МАГИЯ ЗДЕСЬ: Spring AI сам парсит JSON в Record!
    }

    @GetMapping("/ai/tool")
    public String askSystem(@RequestParam String question) {
        return chatClient.prompt()
                .user(question)
                .call()
                .content();
    }

    @GetMapping("/ai/template")
    public String generateWithTemplate(
            @RequestParam String userName,
            @RequestParam String serviceName,
            @RequestParam String action
    ) {
        // Создаем шаблон промпта с переменными
        String templateText = """
            Ты системный администратор. Пользователь {userName} хочет выполнить действие "{action}" 
            над сервисом {serviceName}. 
            Предоставь краткую инструкцию на русском языке, как это сделать безопасно.
            """;

        // Заполняем переменные
        PromptTemplate promptTemplate = new PromptTemplate(templateText);
        String finalPrompt = promptTemplate.render(Map.of(
                "userName", userName,
                "serviceName", serviceName,
                "action", action
        ));

        // Отправляем готовый промпт в модель
        return chatClient.prompt()
                .user(finalPrompt)
                .call()
                .content();
    }

    @GetMapping(value = "/ai/stream", produces = MediaType.TEXT_EVENT_STREAM_VALUE + ";charset=UTF-8")
    public Flux<String> streamResponse(@RequestParam String question) {
        return chatClient.prompt()
                .user(question)
                .stream()
                .content();
    }
}
```
