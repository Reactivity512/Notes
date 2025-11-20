# Go 1.20

Go 1.20 был выпущен в феврале 2023 года.

## Новые API и изменения в стандартной библиотеке

1. Новые функции для работы со срезами

```go
// Проверка, содержит ли срез значение (только для comparable-типов)
strings.Contains([]string{"a", "b"}, "b")  // true

// Клонирование среза с новым underlying array
nums := []int{1, 2, 3}
numsClone := slices.Clone(nums)  
```

2. Улучшения в `sync`
Добавлен `OnceFunc` и `OnceValue` для однократного выполнения

```go
var init = sync.OnceFunc(func() {
    fmt.Println("Инициализация!")
})
```

3. Новые методы в `errors`

```go
err := errors.New("ошибка")
wrapped := fmt.Errorf("контекст: %w", err)
errors.Join(err1, err2)  // Объединение ошибок
```

4. Улучшена вывод типов — меньше нужно явно указывать типы при вызове generic-функций.

```go
func Map[T, U any](slice []T, f func(T) U) []U { ... }

// Раньше:
result := Map[int, string](ints, func(x int) string { return fmt.Sprint(x) })

// Теперь (вывод типов работает лучше):
result := Map(ints, func(x int) string { return fmt.Sprint(x) }) // T и U выведены
```

5. Добавлен в стандартную библиотеку: `slices` — аналог `strings`, но для слайсов.

```go
slices.Contains(slice, item)
slices.Index(slice, item)
slices.Equal(a, b)
slices.Compact(slice)        // удаляет подряд идущие дубликаты
slices.Sort(slice)
slices.SortFunc(slice, less)
slices.Clip(slice)           // освобождает память, обрезая capacity
```

6. Улучшения в `maps`, пакет получил новые функции:

```go
maps.Clone(m)     // копирование мапы
maps.Copy(dst, src)
maps.DeleteFunc(m, func(key, value) bool) // удаление по условию
```

7. Улучшения в `net/http`
* `ServeMux` с поддержкой методов. Теперь можно регистрировать обработчики по HTTP-методу:

```go
mux := http.NewServeMux()
mux.HandleFunc("GET /users", getUsers)
mux.HandleFunc("POST /users", createUser)
mux.HandleFunc("GET /users/{id}", getUser)
```
Поддерживаются:
Методы: `GET`, `POST`, `PUT`, `DELETE` и т.д.
Параметры в путях: `{name}`, `/{id}`

* `http.StripPrefix` и `http.FileServer` устарели
`http.FileServer` и `http.Dir` не устарели, но рекомендуется использовать:
```go
http.FileServerFS(fs) // с поддержкой `fs.FS`
```

8. Новый API для работы с временем: `Time.UnixMicro` и `UnixMilli`
Добавлены методы для работы с микросекундами и миллисекундами:

```go
t.UnixMicro()   // время в микросекундах
t.UnixMilli()   // в миллисекундах
time.UnixMicro(n) // из микросекунд
time.UnixMilli(n) // из миллисекунд
```

9. Улучшения в `crypto/tls`
Поддержка TLS 1.3 0-RTT (Zero Round Trip Time Resumption) — ускорение повторных подключений. Но включается с осторожностью из-за рисков replay-атак.

10. Улучшения в `runtime` и `debug`

* Новый метод `runtime.MemStats` — `HeapAlloc` и другие поля стали точнее.
* Добавлен `debug.ReadGCStats` — более детальная статистика по GC.
* Finalizers теперь вызываются раньше и надёжнее.

## Изменения в инструментах

1. `go vet` теперь проверяет некорректные сравнения `time.Time`
Ловит случаи вроде `t1 == t2` (нужно использовать `t1.Equal(t2)`).
2. Новая команда `go doc` с улучшенным форматированием
Поддержка цветного вывода и Markdown.
3. `cover` теперь поддерживает ветвления (branch coverage)
Показывает, какие `if/switch` не были протестированы:
    ```bash
    go test -cover -covermode=atomic ./...
    ```

## Улучшения производительности

1. Компилятор теперь использует PGO (Profile-Guided Optimization)
Можно подавать CPU-профиль (`default.pgo`) для оптимизации горячих участков кода.
Пример использования:
    ```bash
    go build -pgo=auto ./...  # Автоматически ищет default.pgo
    ```
    Дает прирост 3-15% для критических путей.
2. Ускорена работа GC. Уменьшены задержки при сборке мусора для больших хипов.
3. Оптимизированы `encoding/json` и `encoding/xml`. Скорость работы увеличена на 10-20%.

## Безопасность

* Запрещены рекурсивные типы в `unsafe`. Теперь нельзя создать бесконечную структуру через `unsafe.Pointer`.
* Усилены проверки в `cgo`. Предупреждения о потенциально опасных преобразованиях.

## Поддержка платформ

* Windows 7/8 больше не поддерживаются. Минимальная версия — Windows 10.
* Улучшена работа на ARM64 (Apple Silicon, Linux).

## Прочие изменения

* Новый алгоритм хеширования для `map`
* Уменьшены коллизии для сложных ключей.
* `defer` стал немного быстрее
* Оптимизация на 5-10% для частых вызовов.
* Обновлен Unicode до 15.0
* Добавлены новые эмодзи и символы.
