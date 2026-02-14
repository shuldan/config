# `config` — Универсальный и безопасный загрузчик конфигурации для Go-приложений

[![Go CI](https://github.com/shuldan/config/workflows/Go%20CI/badge.svg)](https://github.com/shuldan/config/actions)
[![codecov](https://codecov.io/gh/shuldan/config/branch/main/graph/badge.svg)](https://codecov.io/gh/shuldan/config)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

Пакет `config` предоставляет простой, типобезопасный и расширяемый способ загрузки, объединения, валидации и доступа к конфигурации из различных источников — файлов (JSON/YAML), переменных окружения и произвольных структур Go.

Спроектирован для использования в проектах, построенных по принципам **DDD**: предоставляет интерфейс `ConfigProvider` для инверсии зависимостей и лёгкого мокирования в тестах.

---

## 🚀 Основные возможности

- **Мультиисточниковость** — загрузка из `.json`, `.yaml`, переменных окружения и `map[string]any`
- **Глубокое слияние** — несколько источников объединяются в один конфиг с приоритетом (последний загрузчик побеждает)
- **Вложенные ключи** — доступ через точку: `database.host`, `server.timeouts.read`
- **Типизированные геттеры** — `string`, `int`, `int64`, `uint64`, `float64`, `bool`, `time.Duration`, `time.Time`, слайсы, map-ы
- **Привязка к структурам** — `Unmarshal` с поддержкой тегов `cfg`, `default`, `layout`
- **Валидация** — декларативные правила: обязательные ключи, диапазоны, допустимые значения, регулярные выражения, пользовательские проверки
- **Шаблонизация** — Go-шаблоны внутри значений: `{{ env "PORT" | default "8080" }}`
- **Профили окружений** — автоматическая загрузка `config.production.yaml` поверх `config.yaml`
- **Иммутабельность** — `Config` не изменяется после создания; `WithOverrides` возвращает новую копию
- **Интерфейс `ConfigProvider`** — для инверсии зависимостей в domain/application слоях
- **Безопасность** — защита от path traversal при загрузке файлов
- **Наблюдаемость** — опциональный логгер для диагностики загрузки
- **Авто-парсинг типов из ENV** — опциональное преобразование строковых значений в `bool`, `int`, `float64`

---

## 📦 Установка

**Go 1.24+**

```sh
go get github.com/shuldan/config
```

---

## 🧱 Архитектура

### Структура пакета

```
config/
├── config.go        # ConfigProvider, Config, New, FromMap, типизированные геттеры, WithOverrides
├── option.go        # Option, builder, WithLogger, WithLoader, WithProfile, WithProfileFromEnv
├── loader.go        # Loader interface
├── yaml_loader.go   # FromYAML, WithBasePath, Optional
├── json_loader.go   # FromJSON, WithBasePath, Optional
├── env_loader.go    # FromEnv, WithAutoTypeParse
├── errors.go        # LoadError, ValidationError, sentinel-ошибки
├── logger.go        # Logger interface, nopLogger
├── template.go      # processValue, render, функции шаблонов
├── unmarshal.go     # Unmarshal + конвертация типов
├── validation.go    # Validate, Required, InRange, OneOf, MatchRegex, Custom
└── utils.go         # deepCopy, mergeMaps, normalize, resolveSecurePath, autoParseString
```

### Интерфейс `ConfigProvider`

Принимайте этот интерфейс в domain- и application-слоях для инверсии зависимостей:

```go
type ConfigProvider interface {
    Has(key string) bool
    Get(key string) any
    GetString(key string, defaultVal ...string) string
    GetInt(key string, defaultVal ...int) int
    GetInt64(key string, defaultVal ...int64) int64
    GetUint64(key string, defaultVal ...uint64) uint64
    GetFloat64(key string, defaultVal ...float64) float64
    GetBool(key string, defaultVal ...bool) bool
    GetDuration(key string, defaultVal ...time.Duration) time.Duration
    GetTime(key string, layout string, defaultVal ...time.Time) time.Time
    GetStringSlice(key string, separator ...string) []string
    GetIntSlice(key string) []int
    GetFloat64Slice(key string) []float64
    GetMap(key string) (map[string]any, bool)
    GetSub(key string) (ConfigProvider, bool)
    Unmarshal(key string, target any) error
    All() map[string]any
}
```

Пример использования в сервисе:

```go
type OrderService struct {
    cfg config.ConfigProvider
}

func NewOrderService(cfg config.ConfigProvider) *OrderService {
    return &OrderService{cfg: cfg}
}

func (s *OrderService) MaxItems() int {
    return s.cfg.GetInt("orders.max_items", 100)
}
```

### Интерфейс `Loader`

Для подключения новых источников конфигурации:

```go
type Loader interface {
    Load() (map[string]any, error)
}
```

---

## 📖 Загрузчики

### `FromYAML` — загрузка из YAML-файлов

Принимает список путей-кандидатов. Загружает **первый найденный** файл (fallback-семантика). Чтобы объединить несколько файлов, используйте несколько загрузчиков.

```go
cfg, err := config.New(
    config.FromYAML("config.yaml"),
)
```

С указанием базовой директории:

```go
config.FromYAML("config.yaml").WithBasePath("/etc/myapp")
```

Опциональный файл (отсутствие — не ошибка):

```go
config.FromYAML("config.local.yaml").Optional()
```

Fallback-цепочка — загрузится первый существующий:

```go
config.FromYAML("config.yaml", "config.default.yaml")
```

### `FromJSON` — загрузка из JSON-файлов

API идентичен `FromYAML`:

```go
cfg, err := config.New(
    config.FromJSON("config.json").WithBasePath("/etc/myapp"),
)
```

### `FromEnv` — загрузка из переменных окружения

Читает переменные с заданным префиксом. Префикс удаляется из имени ключа. Двойное подчёркивание (`__`) используется как разделитель вложенности.

```sh
export APP_DATABASE__HOST=localhost
export APP_DATABASE__PORT=5432
export APP_DEBUG=true
```

```go
cfg, err := config.New(
    config.FromEnv("APP_"),
)

cfg.GetString("database.host") // "localhost"
cfg.GetString("database.port") // "5432" (строка!)
cfg.GetBool("debug")           // false (строка "true" без авто-парсинга → default)
```

#### Авто-парсинг типов

По умолчанию все значения из переменных окружения — строки. Включите авто-парсинг для автоматического определения типов:

```go
config.FromEnv("APP_").WithAutoTypeParse()
```

Приоритет распознавания: `bool` → `int` → `float64` → `string`

```sh
export APP_PORT=8080        # → int(8080)
export APP_DEBUG=true       # → bool(true)
export APP_RATE=0.75        # → float64(0.75)
export APP_NAME=myapp       # → string("myapp")
```

### `FromMap` — создание из `map[string]any`

Создаёт конфигурацию напрямую из Go-map. Map копируется глубоко. Удобно для тестов:

```go
cfg := config.FromMap(map[string]any{
    "server": map[string]any{
        "host": "localhost",
        "port": 8080,
    },
})
```

### Пользовательский загрузчик

Реализуйте интерфейс `Loader` и передайте его через `WithLoader`:

```go
type consulLoader struct {
    addr string
}

func (l *consulLoader) Load() (map[string]any, error) {
    // загрузка из Consul...
    return data, nil
}

cfg, err := config.New(
    config.FromYAML("config.yaml"),
    config.WithLoader(&consulLoader{addr: "localhost:8500"}),
)
```

---

## 📖 Слияние конфигураций

Загрузчики выполняются последовательно. Каждый следующий мержится поверх предыдущего. Вложенные map-ы объединяются рекурсивно:

```go
cfg, err := config.New(
    config.FromYAML("config.defaults.yaml"),  // база
    config.FromYAML("config.yaml"),           // переопределения
    config.FromYAML("config.local.yaml").Optional(), // локальные переопределения
    config.FromEnv("APP_"),                   // env переопределяет всё
)
```

**Приоритет**: последний загрузчик — высший приоритет.

---

## 📖 Типизированные геттеры

Каждый геттер принимает опциональное значение по умолчанию. Если ключ не найден или значение не конвертируется — возвращается default (или zero value типа).

### Примитивные типы

```go
// Строка
host := cfg.GetString("server.host", "0.0.0.0")

// Целое число
port := cfg.GetInt("server.port", 8080)

// int64
maxSize := cfg.GetInt64("storage.max_size", 1073741824)

// uint64
fileLimit := cfg.GetUint64("upload.max_bytes", 10485760)

// Число с плавающей точкой
rate := cfg.GetFloat64("billing.tax_rate", 0.2)

// Булево значение
debug := cfg.GetBool("debug", false)
```

#### Преобразование булевых значений

`GetBool` распознаёт: `true`, `1`, `on`, `yes`, `y` → `true`; `false`, `0`, `off`, `no`, `n` → `false`.

### Время и длительность

```go
// Длительность — парсит строки вида "5s", "100ms", "2h30m"
timeout := cfg.GetDuration("server.timeout", 30*time.Second)

// Также поддерживает числа (интерпретируются как миллисекунды)
// "timeout: 5000" → 5s

// Время — с указанием layout
startedAt := cfg.GetTime("job.started_at", time.RFC3339, time.Now())
```

### Слайсы

```go
// Строковый слайс — из YAML-массива или строки с разделителем
tags := cfg.GetStringSlice("app.tags")                  // ["web", "api"]
tags = cfg.GetStringSlice("app.tags_csv")               // "web,api" → ["web", "api"]
tags = cfg.GetStringSlice("app.tags_pipe", "|")          // "web|api" → ["web", "api"]

// Числовые слайсы
ports := cfg.GetIntSlice("server.ports")                // [8080, 8081]
thresholds := cfg.GetFloat64Slice("alert.thresholds")   // [0.5, 0.8, 0.95]
```

### Map и подконфигурация

```go
// Map — возвращает глубокую копию
headers, ok := cfg.GetMap("proxy.headers")

// Подконфигурация — изолированный ConfigProvider
redisCfg, ok := cfg.GetSub("redis")
if ok {
    host := redisCfg.GetString("host")
    port := redisCfg.GetInt("port")
}
```

### Прямой доступ

```go
// Проверка существования ключа
if cfg.Has("feature.enabled") { ... }

// Получение «сырого» значения
raw := cfg.Get("some.key") // any

// Снимок всей конфигурации (глубокая копия)
all := cfg.All()
```

---

## 📖 Привязка к структурам (`Unmarshal`)

Маппинг конфигурации на типизированные Go-структуры. Поля привязываются по тегу `cfg`; если тег не задан, используется имя поля в нижнем регистре.

### Базовый пример

```yaml
# config.yaml
database:
  host: localhost
  port: 5432
  name: myapp
  timeout: 5s
  read_only: false
```

```go
type DatabaseConfig struct {
    Host     string        `cfg:"host"`
    Port     int           `cfg:"port"`
    Name     string        `cfg:"name"`
    Timeout  time.Duration `cfg:"timeout"`
    ReadOnly bool          `cfg:"read_only"`
}

var dbCfg DatabaseConfig
err := cfg.Unmarshal("database", &dbCfg)
// dbCfg.Host     = "localhost"
// dbCfg.Port     = 5432
// dbCfg.Timeout  = 5 * time.Second
// dbCfg.ReadOnly = false
```

### Вложенные структуры

```yaml
app:
  server:
    host: 0.0.0.0
    port: 8080
  database:
    host: db.example.com
    port: 5432
```

```go
type AppConfig struct {
    Server   ServerConfig   `cfg:"server"`
    Database DatabaseConfig `cfg:"database"`
}

type ServerConfig struct {
    Host string `cfg:"host"`
    Port int    `cfg:"port"`
}

var appCfg AppConfig
err := cfg.Unmarshal("app", &appCfg)
```

### Маппинг от корня

```go
err := cfg.Unmarshal("", &appCfg)
```

### Значения по умолчанию (`default`)

Если ключ отсутствует в конфигурации, используется значение из тега `default`:

```go
type ServerConfig struct {
    Host string `cfg:"host" default:"0.0.0.0"`
    Port int    `cfg:"port" default:"8080"`
}
```

### Формат даты (`layout`)

```go
type JobConfig struct {
    ScheduledAt time.Time `cfg:"scheduled_at" layout:"2006-01-02"`
}
```

Если `layout` не указан, используется `time.RFC3339`.

### Слайсы

```go
type Config struct {
    Tags  []string `cfg:"tags"`
    Ports []int    `cfg:"ports"`
}
```

Строковые значения с разделителем можно настроить через тег `separator`:

```go
type Config struct {
    AllowedIPs []string `cfg:"allowed_ips" separator:";"`
}
// "192.168.1.1;10.0.0.1" → ["192.168.1.1", "10.0.0.1"]
```

### Пропуск поля

```go
type Config struct {
    Internal string `cfg:"-"` // поле не будет заполняться
}
```

### Указатели

```go
type Config struct {
    MaxRetries *int          `cfg:"max_retries"`
    Timeout    *time.Duration `cfg:"timeout"`
}
// nil если ключ отсутствует и нет default
```

---

## 📖 Валидация

Метод `Validate` принимает набор правил и возвращает `*ValidationError`, содержащий **все** нарушения (не только первое).

### Встроенные правила

```go
err := cfg.Validate(
    // Обязательные ключи
    config.Required("database.host"),
    config.Required("database.port"),

    // Числовой диапазон [min, max]
    config.InRange("database.port", 1, 65535),
    config.InRange("server.workers", 1, 256),

    // Допустимые строковые значения
    config.OneOf("log.level", "debug", "info", "warn", "error"),

    // Регулярное выражение
    config.MatchRegex("database.host", `^[a-zA-Z0-9.\-]+$`),

    // Пользовательская проверка
    config.Custom("server.timeout", func(v any) error {
        s, ok := v.(string)
        if !ok {
            return fmt.Errorf("expected string")
        }
        d, err := time.ParseDuration(s)
        if err != nil {
            return fmt.Errorf("invalid duration: %w", err)
        }
        if d > 60*time.Second {
            return fmt.Errorf("timeout too large: %s", d)
        }
        return nil
    }),
)

if err != nil {
    // config: validation failed:
    //   - "database.host": required key is missing
    //   - "database.port": value 70000 is out of range [1, 65535]
    log.Fatal(err)
}
```

### Программная обработка ошибок валидации

```go
var valErr *config.ValidationError
if errors.As(err, &valErr) {
    for _, violation := range valErr.Violations {
        fmt.Println("⚠", violation)
    }
}
```

> **Примечание**: `InRange`, `OneOf`, `MatchRegex` не требуют наличия ключа — если ключ отсутствует, правило пропускается. Используйте `Required` отдельно для проверки обязательности.

---

## 📖 Шаблонизация значений

Строковые значения конфигурации могут содержать Go-шаблоны (`{{ ... }}`), которые разрешаются при загрузке.

### Доступные функции

| Функция                          | Описание                                 |
| -------------------------------- | ---------------------------------------- |
| `{{ env "VAR" }}`               | Значение переменной окружения            |
| `{{ default "val" (env "X") }}` | Значение по умолчанию, если пусто        |
| `{{ upper "text" }}`            | Преобразование в верхний регистр         |
| `{{ lower "TEXT" }}`            | Преобразование в нижний регистр          |
| `{{ trimSpace " text " }}`     | Удаление пробелов по краям              |

### Пример

```yaml
server:
  host: "{{ env \"SERVER_HOST\" | default \"0.0.0.0\" }}"
  port: "{{ env \"PORT\" | default \"8080\" }}"

database:
  dsn: "postgres://{{ env \"DB_USER\" }}:{{ env \"DB_PASS\" }}@{{ env \"DB_HOST\" | default \"localhost\" }}:5432/mydb"

app:
  name: "{{ env \"APP_NAME\" | upper }}"
```

### Обработка ошибок шаблонов

Ошибки шаблонизации не замалчиваются — `New` вернёт ошибку с указанием проблемного ключа:

```
config: template rendering failed: key "database.dsn": template parse: ...
```

---

## 📖 Профили окружений

Автоматическая загрузка базового файла и переопределений для конкретного окружения.

### Явное указание профиля

```go
cfg, err := config.New(
    config.WithProfile("config.yaml", "production"),
)
// Загружает: config.yaml → config.production.yaml (если существует, мержится поверх)
```

### Определение профиля из переменной окружения

```sh
export APP_ENV=staging
```

```go
cfg, err := config.New(
    config.WithProfileFromEnv("config.yaml", "APP_ENV"),
)
// Загружает: config.yaml → config.staging.yaml (если существует)
```

Если переменная пуста — загружается только базовый файл.

### Формат определяется автоматически по расширению

```go
config.WithProfile("config.json", "production")
// Загружает: config.json → config.production.json
```

---

## 📖 Иммутабельность и `WithOverrides`

`Config` неизменяем после создания. Все методы безопасны для конкурентного использования. `All()`, `GetSub()`, `GetMap()` возвращают глубокие копии.

Для создания изменённой копии используйте `WithOverrides`:

```go
baseCfg, _ := config.New(
    config.FromYAML("config.yaml"),
)

// Новый Config с переопределёнными значениями
testCfg := baseCfg.WithOverrides(map[string]any{
    "database.host": "localhost",
    "database.port": 5433,
    "log.level":     "debug",
})

baseCfg.GetString("database.host") // "prod-db.example.com" — не изменился
testCfg.GetString("database.host") // "localhost"
```

Ключи с точками раскрываются в вложенные map-ы:

```go
// "database.host" → {"database": {"host": "localhost"}}
```

**Идиоматично для тестов:**

```go
func TestOrderService(t *testing.T) {
    cfg := baseConfig.WithOverrides(map[string]any{
        "orders.max_items": 5,
        "orders.tax_rate":  0.1,
    })
    svc := NewOrderService(cfg)
    // ...
}
```

---

## 📖 Наблюдаемость

Подключите логгер для диагностики процесса загрузки:

```go
type Logger interface {
    Debug(msg string, args ...any)
}
```

```go
cfg, err := config.New(
    config.WithLogger(myLogger),
    config.FromYAML("config.yaml", "config.local.yaml"),
    config.FromEnv("APP_"),
)
```

Примеры сообщений:

```
[config] loader succeeded (keys=23)
[config] loader failed (error=config: no valid YAML configuration source found: ...)
[config] ready (total_keys=28)
```

Если логгер не задан, используется `nopLogger` (ничего не пишет).

---

## 📖 Обработка ошибок

### Sentinel-ошибки

```go
var (
    config.ErrNoConfigSource  // ни один файл не подошёл
    config.ErrParseYAML       // ошибка разбора YAML
    config.ErrParseJSON       // ошибка разбора JSON
)
```

### `LoadError` — детали загрузки

Когда ни один файл из списка не подошёл, ошибка содержит причину пропуска каждого пути:

```go
cfg, err := config.New(
    config.FromJSON("a.json", "b.json"),
)

var loadErr *config.LoadError
if errors.As(err, &loadErr) {
    for _, detail := range loadErr.Details {
        fmt.Printf("  %s: %s\n", detail.Path, detail.Reason)
    }
}
// Вывод:
//   "a.json": path is outside allowed base "/app"
//   "b.json": file not found
```

### `ValidationError` — список нарушений

```go
var valErr *config.ValidationError
if errors.As(err, &valErr) {
    fmt.Println("Violations:", len(valErr.Violations))
}
```

---

## 📖 Безопасность

Файловые загрузчики (`FromYAML`, `FromJSON`) защищают от path traversal:

- Пути разрешаются в абсолютные (`filepath.Abs` + `filepath.Clean`)
- Проверяется, что путь находится внутри разрешённой базовой директории
- По умолчанию база — текущая рабочая директория
- Для изменения базы используйте `WithBasePath`

```go
// Загрузка из /etc/myapp/ — явное разрешение
config.FromYAML("config.yaml").WithBasePath("/etc/myapp")
```

---

## 🧪 Полный пример

```yaml
# config.yaml
server:
  host: 0.0.0.0
  port: 8080
  timeout: 30s

database:
  host: "{{ env \"DB_HOST\" | default \"localhost\" }}"
  port: 5432
  name: myapp
  max_conns: 10

log:
  level: info

features:
  - auth
  - billing
```

```yaml
# config.production.yaml
server:
  port: 443

database:
  host: "{{ env \"DB_HOST\" }}"
  max_conns: 100

log:
  level: warn
```

```go
package main

import (
    "fmt"
    "log"
    "log/slog"
    "time"

    "github.com/shuldan/config"
)

type AppConfig struct {
    Server   ServerConfig   `cfg:"server"`
    Database DatabaseConfig `cfg:"database"`
    Log      LogConfig      `cfg:"log"`
    Features []string       `cfg:"features"`
}

type ServerConfig struct {
    Host    string        `cfg:"host" default:"0.0.0.0"`
    Port    int           `cfg:"port" default:"8080"`
    Timeout time.Duration `cfg:"timeout" default:"30s"`
}

type DatabaseConfig struct {
    Host     string `cfg:"host"`
    Port     int    `cfg:"port"`
    Name     string `cfg:"name"`
    MaxConns int    `cfg:"max_conns" default:"10"`
}

type LogConfig struct {
    Level string `cfg:"level" default:"info"`
}

func main() {
    cfg, err := config.New(
        config.WithLogger(slog.Default()),
        config.WithProfileFromEnv("config.yaml", "APP_ENV"),
        config.FromEnv("APP_").WithAutoTypeParse(),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Валидация
    if err := cfg.Validate(
        config.Required("database.host"),
        config.Required("database.port"),
        config.InRange("database.port", 1, 65535),
        config.InRange("database.max_conns", 1, 1000),
        config.OneOf("log.level", "debug", "info", "warn", "error"),
    ); err != nil {
        log.Fatal(err)
    }

    // Привязка к структуре
    var appCfg AppConfig
    if err := cfg.Unmarshal("", &appCfg); err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Server: %s:%d (timeout: %s)\n",
        appCfg.Server.Host, appCfg.Server.Port, appCfg.Server.Timeout)
    fmt.Printf("Database: %s:%d/%s (max_conns: %d)\n",
        appCfg.Database.Host, appCfg.Database.Port,
        appCfg.Database.Name, appCfg.Database.MaxConns)
    fmt.Printf("Log level: %s\n", appCfg.Log.Level)
    fmt.Printf("Features: %v\n", appCfg.Features)

    // Прямой доступ (альтернатива)
    timeout := cfg.GetDuration("server.timeout", 30*time.Second)
    features := cfg.GetStringSlice("features")

    // Подконфигурация
    if dbCfg, ok := cfg.GetSub("database"); ok {
        fmt.Println("DB host:", dbCfg.GetString("host"))
    }

    _ = timeout
    _ = features
}
```

---

## 🛠️ Работа с проектом

### Установка инструментов

```sh
make install-tools
```

Устанавливает:
- `golangci-lint` (v2.4.0)
- `goimports`
- `gosec`

### Локальная проверка

```sh
make all
```

Выполняет: проверку форматирования, статический анализ, security-сканирование, запуск тестов.

### CI-проверка

```sh
make ci
```

---

## 📄 Лицензия

Распространяется под лицензией [MIT](LICENSE).

---

## 🤝 Вклад в проект

PR и issue приветствуются! Обязательно соблюдайте стиль кода и покрывайте новый функционал тестами.

---

> **Автор**: MSeytumerov  
> **Репозиторий**: `github.com/shuldan/config`  
> **Go**: `1.24.2`
