# context

Пакет `context` предназначен для передачи информации о контексте выполнения между функциями и горутинами. Он особенно полезен при управлении временем выполнения, отмене операций и передаче метаданных.

Основные цели и преимущества:

1. Управление отменой операций:
Позволяет отменить длительные или зависимые операции, например, при завершении HTTP-запроса или тайм-ауте. Это помогает избегать утечек ресурсов и обеспечивает более отзывчивое приложение.

2. Передача метаданных:
Можно добавлять значения (например, идентификаторы пользователя, токены) в контекст и получать их в глубине вызовов без необходимости передавать дополнительные параметры.

3. Обеспечение тайм-аутов:
Контекст позволяет задать ограничение по времени выполнения операции с помощью `WithTimeout` или `WithDeadline`, что помогает избегать зависания или слишком долгого выполнения.

4. Передача сигнала о завершении:
Горутины могут слушать канал отмены (`Done()`), чтобы корректно завершать работу при необходимости.

**Создание контекста:**

Базовый контекст:

```go
ctx := context.Background()
```

С тайм-аутом
Задает относительное время — сколько времени ждать до отмены.
Когда важно ограничить выполнение по продолжительности (Дай этой операции максимум 2 секунды, потом отменяй):

```go
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()
```

С дедлайном
Задает абсолютное время (в виде time.Time), после которого контекст будет отменён.
Когда уже есть заранее известный момент времени, до которого нужно уложиться (Вне зависимости от момента начала операции, всё должно закончиться в 10:00:00):

```go
ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(2*time.Second))
defer cancel()
```

Внутренне `WithTimeout` просто вызывает `WithDeadline`, добавляя к текущему времени заданный интервал. То есть `WithTimeout(d)` — это упрощённая обёртка над `WithDeadline(time.Now().Add(d))`.

С отменой вручную:

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
```

Передача значений:

```go
ctx := context.WithValue(context.Background(), "userID", 123)
val := ctx.Value("userID") // вернёт 123

```

Пример с HTTP-запросом с таймаутом:

```go
package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	// Создаем контекст с таймаутом 2 секунды
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel() // Освобождаем ресурсы

	// Создаем HTTP-запрос с контекстом
	req, err := http.NewRequestWithContext(ctx, "GET", "https://example.com", nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// Выполняем запрос
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Request failed:", err)
		return
	}
	defer resp.Body.Close()

	// Читаем тело ответа
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	fmt.Printf("Response length: %d\n", len(body))
}
```

Пример с горутиной и таймаутом:

```go
package main

import (
	"context"
	"fmt"
	"time"
)

func longOperation(ctx context.Context) error {
	select {
	case <-time.After(3 * time.Second): // Имитация долгой операции
		fmt.Println("Operation completed")
		return nil
	case <-ctx.Done(): // Срабатывает при отмене контекста
		fmt.Println("Operation canceled")
		return ctx.Err()
	}
}

func main() {
	// Контекст с таймаутом 1 секунда
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Запускаем операцию
	err := longOperation(ctx)
	if err != nil {
		fmt.Println("Error:", err) // Выведет "context deadline exceeded"
	}
}
```

## Context

Контекст (Context) передаёт дедлайн, сигнал отмены и другие значения через API.
Методы контекста могут вызываться одновременно из нескольких goroutine.

Основные методы:

* **`Background() Context`** - возвращает непустой (non-nil), но пустой по содержанию `Context`. Он никогда не отменяется, не содержит значений и не имеет дедлайна. Обычно используется в main-функции, при инициализации, в тестах, а также как корневой `Context` для входящих запросов.
* **`TODO() Context`** - возвращает непустой (non-nil), но пустой `Context`. Следует использовать `context.TODO` в тех случаях, когда ещё не ясно, какой `Context` должен быть использован, или он пока недоступен (например, если окружающая функция ещё не была изменена так, чтобы принимать `Context` в качестве параметра).
* ***`WithoutCancel(parent Context) Context`** - возвращает производный контекст, который ссылается на родительский контекст, но не отменяется, когда отменяется родитель. Возвращённый контекст не имеет дедлайна, его Done-канал равен `nil`, и `Err` также отсутствует. Вызов `Cause` для этого контекста возвращает `nil`. (Полезно, когда ты хочешь унаследовать данные (`Value`) из контекста, но не хочешь, чтобы отмена родителя повлияла на твою операцию (например, при запуске фоновой задачи).)
* **`WithValue(parent Context, key, val any) Context`** - возвращает производный контекст, ссылающийся на родительский `Context`. В этом новом контексте ключу `key` сопоставлено значение `val`.
Используйте значения в `context` только для данных, относящихся к текущему запросу и передающихся между процессами и API.
Не используйте `context` для передачи необязательных параметров функциям.
Переданный ключ должен быть сравнимым типом (comparable) и не должен быть строкой (string) или другим встроенным типом,
чтобы избежать конфликтов между разными пакетами, использующими `context`.
В WithValue рекомендуется определять собственные типы ключей.
Чтобы избежать дополнительных аллокаций при приведении к `interface{}`, ключи часто имеют конкретный тип `struct{}`.
Также допустимы экспортируемые переменные ключей с типом-указателем или интерфейсом.
Пример:
    ```go
    import (
        "context"
        "fmt"
    )

    func main() {
        type favContextKey string

        f := func(ctx context.Context, k favContextKey) {
            if v := ctx.Value(k); v != nil {
                fmt.Println("found value:", v)
                return
            }
            fmt.Println("key not found:", k)
        }

        k := favContextKey("language")
        ctx := context.WithValue(context.Background(), k, "Go")

        f(ctx, k)
        f(ctx, favContextKey("color"))
    }
    ```

## CancelFunc

**`type CancelFunc func()`** - сообщает операции, что нужно прекратить выполнение. `CancelFunc` не ожидает, пока операция завершится. Эту функцию могут вызывать одновременно несколько `goroutine`. После первого вызова все последующие вызовы `CancelFunc` не имеют эффекта.

Пример:
```go
package main

import (
    "context"
    "fmt"
    "time"
)

func main() {
    // Создаём контекст с возможностью отмены
    ctx, cancel := context.WithCancel(context.Background())

    // Запускаем горутину, которая работает, пока контекст не отменён
    go func(ctx context.Context) {
        for {
            select {
            case <-ctx.Done():
                fmt.Println("Операция отменена")
                return
            default:
                fmt.Println("Работаю...")
                time.Sleep(500 * time.Millisecond)
            }
        }
    }(ctx)

    // Даём горутине поработать 2 секунды
    time.Sleep(2 * time.Second)

    // Отменяем контекст, вызывая CancelFunc
    cancel()

    // Ждём немного, чтобы увидеть завершение горутины
    time.Sleep(1 * time.Second)
    fmt.Println("Главная функция завершена")
}
```

## CancelCauseFunc

Работает как `CancelFunc`, но дополнительно устанавливает причину отмены. Эту причину можно получить, вызвав `Cause` у отменённого контекста или любого из его производных контекстов.
Если контекст уже был отменён, `CancelCauseFunc` не устанавливает причину.
Пример: если `childContext` создан на основе `parentContext`:
* Если сначала отменяется `parentContext` с причиной `cause1`, а затем `childContext` — с `cause2`,
то `Cause(parentContext) == Cause(childContext) == cause1`
* Если сначала отменяется `childContext` с причиной `cause2`, а затем `parentContext` — с `cause1`,
то `Cause(parentContext) == cause1`, а `Cause(childContext) == cause2`

Пример:
```go
package main

import (
    "context"
    "errors"
    "fmt"
    "time"
)

func main() {
    // Создаем контекст с возможностью отмены с причиной
    ctx, cancel := context.WithCancelCause(context.Background())

    go func(ctx context.Context) {
        // Ждем отмены контекста
        <-ctx.Done()
        fmt.Println("Контекст отменен")
        fmt.Println("Причина отмены:", context.Cause(ctx))
    }(ctx)

    time.Sleep(1 * time.Second)

    // Отменяем контекст с пользовательской причиной
    cancel(errors.New("операция больше не требуется"))

    // Даем горутине время завершиться
    time.Sleep(500 * time.Millisecond)
}
```


## func AfterFunc(ctx Context, f func()) (stop func() bool)

Планирует вызов функции `f` в отдельной `goroutine` после отмены контекста `ctx`. Если `ctx` уже отменён — `f` вызывается немедленно, также в отдельной `goroutine`.
Множественные вызовы `AfterFunc` для одного контекста работают независимо — один не заменяет другой.
Вызов возвращаемой функции `stop` отменяет связь между `ctx` и `f`. Она возвращает `true`, если удалось предотвратить запуск `f`.
Если возвращается `false`, это значит, что:
* контекст уже был отменён, и `f` уже запущена в `goroutine`, или
* выполнение `f` уже было остановлено ранее.

Функция `stop` не ожидает завершения `f`.
Если вызывающему нужно знать, что `f` завершена — это нужно координировать вручную (например, через `sync.WaitGroup` или канал).
Если у `ctx` есть метод `AfterFunc(func()) func() bool`, то будет использоваться он для планирования вызова.

Пример: Отмена `f` с помощью `stop()`:

```go
package main

import (
    "context"
    "fmt"
    "time"
)

func main() {
    ctx, cancel := context.WithCancel(context.Background())

    // Зарегистрируем функцию, которая должна выполниться при отмене контекста
    stop := context.AfterFunc(ctx, func() {
        fmt.Println("Функция f вызвана (контекст отменён)")
    })

    // Отменим контекст через 2 секунды
    go func() {
        time.Sleep(2 * time.Second)
        cancel()
    }()

    // Попробуем остановить выполнение f до отмены контекста
    time.Sleep(1 * time.Second)
    cancelled := stop()
    fmt.Println("Функция f была отменена до запуска:", cancelled)

    // Подождём, чтобы увидеть, была ли вызвана f
    time.Sleep(3 * time.Second)
}
```

* `AfterFunc` регистрирует `f`, которая будет вызвана при отмене `ctx`.
* Через 1 секунду вызывается `stop()`, и если `ctx` ещё не был отменён, `f` не будет вызвана.
* Вывод покажет `true`, если `f` была успешно отменена до запуска.

Пример: Ожидание завершения `f` с помощью `sync.WaitGroup`

```go
package main

import (
    "context"
    "fmt"
    "sync"
    "time"
)

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    var wg sync.WaitGroup

    wg.Add(1)
    context.AfterFunc(ctx, func() {
        defer wg.Done()
        fmt.Println("f начала выполняться")
        time.Sleep(2 * time.Second)
        fmt.Println("f завершена")
    })

    // Отменим контекст (запустит f)
    time.Sleep(1 * time.Second)
    cancel()

    // Ожидаем завершения f
    wg.Wait()
    fmt.Println("Главная функция завершена")
}
```

* `AfterFunc` запускает `f`, как только `ctx` отменяется.
* Мы используем `sync.WaitGroup`, чтобы дождаться завершения `f`.
* Подходит, если тебе важно дождаться окончания логики в `f`, прежде чем завершить основную программу.

## func Cause(c Context) error

Возвращает ненулевую ошибку (`error`), объясняющую, почему контекст c был отменён. Причина устанавливается при первой отмене самого c или одного из его родительских контекстов.
Если отмена произошла через вызов `CancelCauseFunc(err)`, то `Cause` вернёт переданную ошибку `err`.
В остальных случаях `Cause(c)` возвращает то же, что и `c.Err()`.
Если контекст ещё не был отменён, `Cause` возвращает `nil`.

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "time"
)

func main() {
    // ✅ Пример 1: отмена без причины (обычный CancelFunc)
    ctx1, cancel1 := context.WithCancel(context.Background())
    cancel1()

    fmt.Println("Пример 1:")
    fmt.Println("Err():   ", ctx1.Err())           // context canceled
    fmt.Println("Cause():", context.Cause(ctx1))   // context canceled
    fmt.Println()

    // ✅ Пример 2: отмена с причиной через CancelCauseFunc
    ctx2, cancel2 := context.WithCancelCause(context.Background())
    cancel2(errors.New("запрос больше не нужен"))

    fmt.Println("Пример 2:")
    fmt.Println("Err():   ", ctx2.Err())           // context canceled
    fmt.Println("Cause():", context.Cause(ctx2))   // запрос больше не нужен
}

/*
    Вывод: 
    Пример 1:
    Err():   context canceled
    Cause(): context canceled

    Пример 2:
    Err():   context canceled
    Cause(): запрос больше не нужен
*/
```

## func WithCancel(parent Context) (ctx Context, cancel CancelFunc)

Возвращает производный контекст, который ссылается на родительский, но имеет собственный канал `Done`. Канал `Done` нового контекста закрывается при вызове возвращённой функции `cancel`, или если сначала будет отменён родительский контекст — в зависимости от того, что произойдёт первым.
Отмена этого контекста освобождает связанные с ним ресурсы, поэтому функцию `cancel` следует вызывать как можно скорее после завершения операций, выполняемых в данном контексте.

Пример:

```go
package main

import (
    "context"
    "fmt"
    "time"
)

func main() {
    // Создаём родительский контекст с возможностью отмены
    parentCtx, parentCancel := context.WithCancel(context.Background())

    // Создаём дочерний контекст на основе родительского
    childCtx, childCancel := context.WithCancel(parentCtx)

    // Горутина, следящая за отменой дочернего контекста
    go func() {
        <-childCtx.Done()
        fmt.Println("Дочерний контекст отменён. Причина:", childCtx.Err())
    }()

    // Через 1 секунду отменим родительский контекст
    time.Sleep(1 * time.Second)
    fmt.Println("Отменяем родительский контекст")
    parentCancel()

    // Подождём немного, чтобы увидеть результат
    time.Sleep(500 * time.Millisecond)

    // Вызов cancel на дочернем контексте всё ещё допустим, но уже ничего не изменит
    childCancel()

    // Подождём немного перед завершением
    time.Sleep(500 * time.Millisecond)
}
```

* `childCtx` создаётся на основе `parentCtx`, но имеет свой собственный Done-канал.
* Горутина ждёт отмены `childCtx` через `<-childCtx.Done()`.
* Когда мы отменяем родительский контекст, дочерний тоже отменяется, так как они связаны.
* Отмена дочернего контекста освобождает ресурсы, и даже если вызвать `childCancel()` после этого — это уже не повлияет на `Done()` (он уже закрыт).

## func WithCancelCause(parent Context) (ctx Context, cancel CancelCauseFunc)

Ведёт себя так же, как `WithCancel`, но возвращает `CancelCauseFunc` вместо `CancelFunc`.
Вызов `cancel` с ненулевой ошибкой (так называемой "причиной") сохраняет эту ошибку в контексте `ctx`; её затем можно получить с помощью `Cause(ctx)`.
Если вызвать `cancel` с nil, то причиной устанавливается стандартное значение `context.Canceled`.

```go
ctx, cancel := context.WithCancelCause(parent)
cancel(myError)
ctx.Err() // returns context.Canceled
context.Cause(ctx) // returns myError
```

## func WithDeadline(parent Context, d time.Time) (Context, CancelFunc)

Возвращает производный контекст, который ссылается на родительский,
но имеет собственный дедлайн, не позднее указанного времени `d`.
Если у родительского контекста дедлайн уже раньше, чем `d`, то `WithDeadline(parent, d)` семантически эквивалентен `parent`.
Канал `Done` возвращаемого контекста закрывается при наступлении дедлайна, при вызове возвращённой функции `cancel`, или при отмене родительского контекста — в зависимости от того, что произойдёт первым.
Отмена этого контекста освобождает связанные с ним ресурсы, поэтому функцию `cancel` следует вызывать как можно скорее после завершения всех операций, использующих данный контекст.

## func WithDeadlineCause(parent Context, d time.Time, cause error) (Context, CancelFunc)

Работает так же, как `WithDeadline`, но дополнительно устанавливает причину отмены контекста, если дедлайн истекает.
Возвращаемая функция `CancelFunc` не устанавливает причину.

## func WithTimeout(parent Context, timeout time.Duration) (Context, CancelFunc)

Это сокращённая форма вызова `WithDeadline(parent, time.Now().Add(timeout))`.
Отмена этого контекста освобождает связанные с ним ресурсы,
поэтому функцию `cancel` следует вызывать как можно скорее после завершения всех операций, использующих данный контекст.

## func WithTimeoutCause(parent Context, timeout time.Duration, cause error) (Context, CancelFunc)

Работает так же, как `WithTimeout`, но дополнительно устанавливает причину отмены возвращаемого контекста, если истекает тайм-аут.
Возвращаемая функция `CancelFunc` не устанавливает причину.
