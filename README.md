# reCached

reCached - это легкая библиотека для Go, которая предоставляет автоматически обновляемый кеш с поддержкой дженериков. Библиотека позволяет кешировать любые типы данных и автоматически обновлять их через заданные интервалы времени.

## Особенности

- Поддержка дженериков (Go 1.18+)
- Автоматическое обновление кеша через заданные интервалы
- Потокобезопасность (thread-safe)
- Возможность ручного обновления кеша
- Глобальное обновление всех экземпляров кеша одной командой

## Установка

```bash
go get github.com/petar/recached
```

## Использование

### Базовый пример

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/petar/recached"
)

func main() {
	// Создаем контекст, который можно отменить
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Функция обновления, которая будет вызываться периодически
	updateFunc := func() (string, error) {
		return time.Now().Format(time.RFC3339), nil
	}

	// Создаем кеш, который будет обновляться каждую секунду
	cache := recached.New(ctx, time.Second, updateFunc)

	// Получаем текущее значение из кеша
	fmt.Println("Current value:", cache.Get())

	// Ждем некоторое время, чтобы кеш обновился
	time.Sleep(3 * time.Second)

	// Получаем обновленное значение
	fmt.Println("Updated value:", cache.Get())

	// Принудительно обновляем кеш
	cache.Update()
	fmt.Println("Manually updated value:", cache.Get())

	// Отменяем контекст, чтобы остановить автоматическое обновление
	cancel()
}
```

### Пример с обработкой ошибок

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/petar/recached"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Функция обновления, которая иногда возвращает ошибку
	updateFunc := func() (int, error) {
		// Симулируем случайную ошибку
		if rand.Intn(3) == 0 {
			return 0, errors.New("random error occurred")
		}
		return rand.Intn(100), nil
	}

	// Создаем кеш с периодом обновления 500 мс
	cache := recached.New(ctx, 500*time.Millisecond, updateFunc)

	// Мониторим значение кеша в течение некоторого времени
	for i := 0; i < 10; i++ {
		fmt.Printf("Value %d: %d\n", i+1, cache.Get())
		time.Sleep(300 * time.Millisecond)
	}
}
```

### Пример глобального обновления кешей

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/petar/recached"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Создаем несколько кешей разных типов
	intCache := recached.New(ctx, time.Hour, func() (int, error) {
		return 42, nil
	})

	stringCache := recached.New(ctx, time.Hour, func() (string, error) {
		return "hello", nil
	})

	// Выводим начальные значения
	fmt.Println("Initial int value:", intCache.Get())
	fmt.Println("Initial string value:", stringCache.Get())

	// Обновляем все кеши одной командой
	recached.GlobalCacheUpdate()

	// Выводим обновленные значения
	fmt.Println("Updated int value:", intCache.Get())
	fmt.Println("Updated string value:", stringCache.Get())
}
```

## Документация API

### Создание нового кеша

```go
func New[T any](ctx context.Context, period time.Duration, updateFunc func() (T, error)) ReCached[T]
```

- `ctx` - контекст для управления жизненным циклом кеша
- `period` - интервал между автоматическими обновлениями
- `updateFunc` - функция, которая возвращает новое значение для кеша

### Глобальное обновление кешей

```go
func GlobalCacheUpdate()
```

Эта функция обновляет все экземпляры кеша, созданные через `New()`. Обновление происходит параллельно для всех кешей.

### Интерфейс ReCached

```go
type ReCached[T any] interface {
	Get() T
	Update()
}
```

- `Get()` - возвращает текущее значение из кеша
- `Update()` - принудительно обновляет значение в кеше

## Тестирование

Библиотека имеет полный набор тестов, которые проверяют все аспекты её функциональности:

```bash
go test -v
```

## Лицензия

GNU General Public License v3.0
