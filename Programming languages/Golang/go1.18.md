# Go 1.18

Go 1.18 был выпущен 15 марта 2022 года. В нем появились одни из самых значимых изменений языка — дженерики (generics).

## Дженерики (Generics)

Новый синтаксис с использованием типовых параметров:

```go
func PrintSlice[T any](s []T) {
    for _, v := range s {
        fmt.Println(v)
    }
}
```

где:

* `[T any]` — параметр типа.
* `any` — псевдоним для `interface{}`.

Ограничения типов (constraints) через интерфейсы:

```go
type Number interface {
    int | float64
}
func Sum[T Number](s []T) T {
    var sum T
    for _, v := range s {
        sum += v
    }
    return sum
}
```

Обобщённые структуры:

```go
type Box[T any] struct {
    Content T
}
```

**Ограничения:**
* Нет специализации методов (все методы должны явно работать с T).
* Некоторые сложные случаи требуют `go:generate`.

Благодаря обобщениям появились два новых пакета:

1. `slices`

```go
import "slices"

nums := []int{3, 1, 4, 1}
slices.Sort(nums)           // [1 1 3 4]
found := slices.Contains(nums, 4)
idx := slices.Index(nums, 1)
```

2. `maps`

```go
import "maps"

m := map[string]int{"a": 1, "b": 2}
keys := maps.Keys(m)        // []string{"a", "b"}
values := maps.Values(m)    // []int{1, 2}
maps.Copy(dst, src)
```

## Fuzzing (Фаззинг-тестирование)

Fuzzing (фаззинг) становится частью `go test`.
Фаззинг — тестирования с рандомными входными данными для поиска багов и паник.
Тесты с префиксом `Fuzz` и `*testing.F`:

```go
func FuzzReverse(f *testing.F) {
    f.Add([]byte("hello")) // Начальные данные
    f.Fuzz(func(t *testing.T, orig []byte) {
        reversed := Reverse(orig)
        if !bytes.Equal(Reverse(reversed), orig) {
            t.Errorf("Ошибка: %q -> %q", orig, reversed)
        }
    })
}
```

Запуск: `go test -fuzz=FuzzReverse`

## Workspace Mode (Рабочие пространства)

Упрощает работу с мульти-модульными проектами.
Создаётся файл `go.work` в корне проекта.
Пример `go.work`:

```go
go 1.18

use (
    ./module1
    ./module2
)
```

Команды:

```bash
go work init
go work use ./module1
```

Локальные зависимости заменяются без правки `go.mod`.
Команды `go build/go test` работают для всех модулей сразу.


## Новые API и изменения в стандартной библиотеке

### Новый пакет `net/netip`

`netip.Addr` — замена net.IP (более быстрая и иммутабельная).

```go
addr, _ := netip.ParseAddr("192.168.1.1")
```

### Новый пакет `strings.Cut`

Удобная замена `Split` для разделения строк:

```go
before, after, found := strings.Cut("hello=world", "=")
// before="hello", after="world", found=true
```

### Новый пакет `debug/buildinfo`

`debug/buildinfo`: Чтение метаданных сборки из бинарников.

### Улучшения `sync/atomic`

Новые типы `Atomic[T]` (для `int32`, `uint64` и др.):

```go
var val atomic.Int32
val.Store(42)
```

### Улучшения `reflect`

`reflect`: Поддержка обобщённых типов (например, `TypeFor[T]()`).

### Улучшения `testing`

`testing`: Новые опции для фаззинга.

### Новые функции в `runtime` и `debug`

Появился `runtime/debug.ReadBuildInfo()` — позволяет читать информацию о сборке (модули, версии, замены). Полезно для логирования, health-check'ов, диагностики:

```go
info, ok := debug.ReadBuildInfo()
if ok {
    fmt.Println("Module:", info.Main.Path)
    for _, dep := range info.Deps {
        fmt.Printf("%s %s\n", dep.Path, dep.Version)
    }
}
```

## Улучшения

* `go vet` теперь понимает `generics`
* `go get` больше не изменяет `go.mod`
* macOS: полная поддержка ARM64 (M1/M2).
* Windows: улучшена работа с символьными ссылками.
* Unicode обновлён до 14.0.
* Добавили экспериментальную поддержку архитектуры RISC-V под Linux
  ```bash
  GOOS=linux GOARCH=riscv64 go build
  ```
* *encoding/json* теперь кэширует метаданные (до 10% быстрее маршалинг).
* `// +build` официально устарел в пользу `//go:build.`
* `strings.Clone` — безопасное клонирование строк (редко нужно, но полезно при работе с `unsafe`).
* `sync.Map` — небольшие оптимизации.

## Улучшения в компиляторе и отладке

* Лучшая поддержка отладки (`debug info`) — теперь отладчики (например, dlv) лучше видят переменные, особенно в функциях с обобщениями.
* Улучшена производительность компиляции для проектов с большим количеством обобщённых функций.
* Поддержка инструкций по памяти (`memory layout`) для `unsafe.Sizeof`, `unsafe.Alignof` и др.

## Улучшения производительности

* Компилятор теперь использует `generics` внутри (например, для `sort`).
* Оптимизация вызовов методов интерфейсов (до 20% быстрее в некоторых случаях).
* Ускорена работа GC для больших хипов.
