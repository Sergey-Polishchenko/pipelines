# Pipelines — Лёгковесная библиотека конвейеров потоковой обработки данных (Go)

`pipelines` предоставляет удобный способ описать и запустить потоковую обработку данных в Go, где каждый узел (node) может иметь 0..N входных каналов и 0..M выходных каналов. Узлы связываются каналами Go, что обеспечивает простую и гибкую модель параллелизма.

## Содержание

* [Задача (TASK.md)](#задача-taskmd)
* [Основные возможности](#основные-возможности)
* [Структура репозитория](#структура-репозитория)
* [Установка](#установка)
* [Использование](#использование)
  * [Быстрый старт](#быстрый-старт)
  * [Примеры (папка `examples`)](#примеры-папка-examples)

* [API библиотеки](#api-библиотеки)
  * [Интерфейс Node](#интерфейс-node)
  * [Основные узлы (`nodes`)](#основные-узлы-nodes)
  * [Утилиты соединения узлов](#утилиты-соединения-узлов)
  * [Pipeline и запуск](#pipeline-и-запуск)

* [Конфигурация узлов](#конфигурация-узлов)
* [Запуск тестов](#запуск-тестов)
* [TODO](#todo)
* [Лицензия](#лицензия)

---

## Задача (TASK.md)

Полную формулировку тестового задания можно найти в файле [TASK.md](TASK.md). Вкратце:

> Требуется реализовать легковесную библиотеку для конструирования пайплайна потоковой обработки данных на языке Go.
>   * Узлы (nodes) могут иметь 0..N входов и 0..M выходов (каналы `chan T`).
>   * Возможность соединять узлы в произвольный ориентированный граф.
>   * Параллельное выполнение независимых операций (через `context.Context` и горутины).
>   * Возможность корректного запуска и остановки конвейера.
>
> **Демонстрационная задача:**
>   * Рекурсивно обойти директорию, найти файлы.
>   * Параллельно вычислить MD5-хеш для каждого файла.
>   * Степень параллелизма настраивается (по умолчанию — 10).
>   * Вывести на консоль результат: `<md5>  <path>`.

---

## Основные возможности

* **Построение произвольного графа** из узлов (DAG).
* **Параллелизм на уровне узлов** и на уровне воркеров (`workerPool`).
* **Гибкие соединения** через: `Connect`, `ConnectToMany`, `ConnectFromMany`.
* **Генераторы и агрегаторы** (узлы-производители и узлы-приёмники данных).
* **Контекст (context.Context)** для остановки конвейера при ошибке или по таймауту.
* **Типобезопасность** благодаря использованию generics (Go 1.18+).

---

## Структура репозитория

```
├── README.md
├── TASK.md
├── LICENSE
├── go.mod
├── pipeline.go
├── node.go
├── connect.go
├── start.go
├── pkg
│   └── utils
│       ├── broadcast.go
│       ├── close_channels.go
│       └── fan_in.go
│
├── nodes
│   ├── aggregator.go
│   ├── config.go
│   ├── node.go
│   ├── errors.go
│   ├── generator.go
│   ├── id.go
│   ├── processor.go
│   ├── worker_pool.go
│   └── zip.go
│
└── examples
    ├── demo
    │   ├── file_generator.go
    │   ├── go.mod
    │   ├── main_test.go
    │   └── main.go
    │
    └── test
        └── main.go
```

---

## Установка

```bash
go get github.com/Sergey-Polishchenko/pipelines@v1.0.0
```

---

## Использование

### Быстрый старт

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/Sergey-Polishchenko/pipelines"
    "github.com/Sergey-Polishchenko/pipelines/nodes"
)

func main() {
    // 1. Генератор чисел 1..5
    gen := nodes.NewGenerator(func(ctx context.Context) (<-chan int, error) {
        out := make(chan int)
        go func() {
            defer close(out)
            for i := 1; i <= 5; i++ {
                select {
                case out <- i:
                case <-ctx.Done():
                    return
                }
            }
        }()
        return out, nil
    })

    // 2. Нода: умножает на 10
    mult := nodes.NewNode(func(x int) (int, error) {
        return x * 10, nil
    }, nodes.Config{Buffer: 5})

    // 3. Агрегатор: печатает результат
    agg := nodes.NewResultAggregator(func(x int) error {
        fmt.Println("Result:", x)
        return nil
    })

    // 4. Собираем Pipeline: gen -> mult -> agg
    p := pipelines.New()
    p.Add(gen, mult, agg)

    // 5. Коннектим узлы
    if err := pipelines.Connect(gen, mult); err != nil {
        panic(err)
    }
    if err := pipelines.Connect(mult, agg); err != nil {
        panic(err)
    }

    // 6. Запускаем с таймаутом 5 секунд
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := p.Run(ctx); err != nil {
        fmt.Println("Pipeline error:", err)
    }
}
```

Ожидаемый вывод (в произвольном порядке, поскольку узлы работают параллельно):

```
Result: 10
Result: 20
Result: 30
Result: 40
Result: 50
```

### Примеры (папка `examples`)

* **MD5-хеширование файлов** (демонстрационная задача).

  * Путь: [`examples/demo`](examples/demo)
  * Файлы:
    * `file_generator.go` — рекурсивный обход директории, генерация имен файлов.
    * `main.go`            — сборка пайплайна: генератор → pool вычисления MD5 → агрегатор.
    * `main_test.go`       — юнит-тест для проверки корректности работы.

  Чтобы запустить пример:

  ```bash
  cd examples/demo
  go run main.go
  ```

  Для тестирования:

  ```bash
  go test ./...
  ```

---

## API библиотеки

### Интерфейс Node

```go
// Node представляет узел конвейера, обрабатывающий элементы из 0..N входных каналов
// и отправляющий результаты в 0..M выходных каналов.
type Node[In, Out any] interface {
    // ID возвращает уникальный идентификатор ноды (для логирования/отладки).
    ID() string

    // SetInput устанавливает один или несколько входных каналов (<-chan In).
    // Если нода не принимает входы, возвращает ErrHasNoInput.
    SetInput(in ...<-chan In) error

    // Output создаёт и возвращает выходной канал (chan Out).
    // Если нода не имеет выходов, возвращает ErrHasNoOutput.
    Output() (chan Out, error)

    // Run запускает обработку ноды:
    // - читает данные из всех входных каналов (FanIn);
    // - применяет пользовательскую логику (Processor/ZipProcessor/Generator/Sink);
    // - отправляет результаты в выходные каналы или в sink-функцию.
    // Если контекст отменяется или при error, нода завершает работу.
    Run(ctx context.Context) error
}
```

### Основные узлы (`nodes`)

#### `NewNode`

```go
func NewNode[In, Out any](proc Processor[In, Out], cfg ...Config) Node[In, Out]
```

* Базовая «простая» нода:

  * Может принимать 0..N входов (через `SetInput`).
  * `Processor[In,Out]` — функция `func(In) (Out, error)`.
  * Читает все входящие данные (с помощью `utils.FanIn`), применяет `proc`, и «broadcast`ит» результат во все выходы (по каналам, созданным через `Output\`).
  * Параметры `Config`:
    * `InBuffer` — размер буфера внутреннего канала для `FanIn`. По умолчанию `0`.
    * `Buffer` — размер буфера для выходных каналов. По умолчанию `10`.

#### `NewWorkerPool`

```go
func NewWorkerPool[In, Out any](proc Processor[In, Out], cfg ...Config) Node[In, Out]
```

* Нода-пул воркеров:
  * **Принимает ровно 1 вход** (если `SetInput` вызван более одного раза, возвращает ошибку).
  * Запускает `cfg.Workers` горутин, каждая читает из входного канала и применяет `proc`.
  * Результаты отправляются в **один** выходной канал (размер буфера равен `cfg.Workers`).

#### `NewResultAggregator`

```go
func NewResultAggregator[In any](sink Sink[In], cfg ...Config) Node[In, any]
```

* Агрегатор:
  * Принимает 0..N входов (через `SetInput`).
  * Читает данные (с помощью `utils.FanIn`) и вызывает `sink(In) error` для каждого элемента.
  * **Не имеет выходов** (вызывает `ErrHasNoOutput` при `Output`).

#### `NewGenerator`

```go
func NewGenerator[Out any](gen Generator[Out]) Node[any, Out]
```

* Генератор:
  * **Не принимает входов** (`SetInput` возвращает `ErrHasNoInput`).
  * При вызове `Run(ctx)`, вызывает `gen(ctx) (chan Out, error)`, получает поток данных и распростра\`саняет их во все выходы (созданные через `Output`).
  * Закрывает выходные каналы после окончания генерации.

#### `NewZip`

```go
func NewZip[In, Out any](proc ZipProcessor[In, Out], cfg ...Config) Node[In, Out]
```

* Zip-нода:
  * Принимает **N входных** каналов.
  * Ждёт по одному элементу от каждого входного потока, собирает их в `[]In`, вызывает `proc([]In) (Out, error)`, и «broadcast\`ит» результат.
  * Если один из каналов закрыт, возвращает `ErrZipNodeClosedInput`.

### Утилиты соединения узлов

```go
func Connect[In, Mid, Out any](from Node[In, Mid], to Node[Mid, Out]) error
func ConnectToMany[In, Mid, Out any](from Node[In, Mid], targets ...Node[Mid, Out]) error
func ConnectFromMany[In, Mid, Out any](sources []Node[In, Mid], to Node[Mid, Out]) error
```

* `Connect`: связывает **1 выход** узла `from` с **1 входом** узла `to`.
* `ConnectToMany`: связывает **один выход** узла `from` с несколькими целями.
  > **Учтите:** в текущей версии `workerPool` поддерживает **только один** выход и поэтому `ConnectToMany` для него заведомо неверна. Либо используйте ее для `node`/`zip`, либо дождитесь исправления.

* `ConnectFromMany`: сливает **множество выходов** разных узлов в **один вход** узла `to` (через `FanIn`).

### Pipeline и запуск

```go
type Pipeline interface {
    Run(ctx context.Context) error
    Add(...Runnable)
}

func New() Pipeline
```

* `Pipeline` хранит список `Runnable` (в основном — нод) и при `Run(ctx)` запускает каждый из них в отдельной горутине.
* Если любая нода возвращает ошибку, контекст отменяется, и через `cancel()` все остальные ноды начинают завершаться.
* `Run(ctx)` возвращает первую ненулевую ошибку или `nil`, если всё прошло успешно.

Пример использования:

```go
p := pipelines.New()
p.Add(node1, node2, node3)
Connect(node1, node2)
Connect(node2, node3)
err := p.Run(ctx)
```

---

## Конфигурация узлов

```go
// Config задаёт параметры буферов и число воркеров (для workerPool)
type Config struct {
    InBuffer int // Размер буфера при сливе входов (FanIn)
    Buffer   int // Размер буфера для выходных каналов
    Workers  int // Число параллельных горутин (для workerPool)
}

// DefaultConfig возвращает Config{InBuffer:0, Buffer:10, Workers:10}
func DefaultConfig() Config
```

* **InBuffer:** используется в `FanIn` при чтении из нескольких входов.
* **Buffer:** размер буфера создаваемых выходных каналов.
* **Workers:** количество горутин-воркеров (только для `NewWorkerPool`).

Рекомендуется явно задавать `Config`, если вы хотите изменить степень параллелизма или размеры буферов. Например:

```go
myPool := nodes.NewWorkerPool(proc, nodes.Config{Buffer: 20, Workers: 15})
```

---

## Запуск тестов

В корне репозитория выполните:

```bash
go test ./...    # Запустит все юнит-тесты библиотеки и примеров
```

В папке [`examples/demo`](examples/demo) есть тест `main_test.go`, проверяющий корректность MD5-конвейера.

---

## TODO

* **Исправить и расширить `ConnectToMany`:**
  * Сделать так, чтобы `workerPool` мог рассылать результаты сразу нескольким подписчикам.
  * Либо выдавать ошибку при попытке `ConnectToMany` к `workerPool`.

* **PipelineBuilder / AutoConnect:**
  * Упростить синтаксис сборки линейных конвейеров, избавившись от ручных вызовов `Connect`.
  * Пример API:
    ```go
    builder := pipelines.NewBuilder().
        Add(gen).
        Then(worker1).
        Then(worker2).
        Then(aggregator).
        Run(ctx)
    ```
* **Middleware-ноды:**
  * Добавить «middleware»-ноду, которая перехватывала бы каждый элемент и выполняла произвольную дополнительную логику (логгирование, фильтрацию).

* **SyncNode (синхронизация):**
  * Узел-барьер, дожидающийся поступления элементов от нескольких потоков, а затем выпускающий единичное уведомление дальше.

* **Улучшения конфигурации:**
  * Разделить `Config` на более узкоспециализированные структуры (`NodeConfig`, `PoolConfig`).
  * Задать оптимальные дефолты для `InBuffer`, `Buffer`, `Workers`.

* **Обработка ошибок и трассировка:**
  * Расширить перечень возвращаемых ошибок (например, `ErrZipNoInput`, `ErrOnlyOneInput` и т. д.) и добавить рекомендации по их логированию.
  * Рассмотреть возможность «dead-letter queue» (канал ошибок), куда можно отправлять «необработанные» элементы без остановки всего пайплайна.

* **Тесты узлов:**
  * Добавить юнит-тесты для каждого базового узла (`node`, `zip`, `aggregator`, `generator`) в отдельности.

---

## Лицензия

Проект распространяется под лицензией [MIT](LICENSE).
