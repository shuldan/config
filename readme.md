Конечно! Вот README для пакета `config`, написанный в том же стиле:

---

# `config` — Универсальный и безопасный загрузчик конфигурации для Go-приложений

[![Go CI](https://github.com/shuldan/config/workflows/Go%20CI/badge.svg)](https://github.com/shuldan/config/actions)
[![codecov](https://codecov.io/gh/shuldan/config/branch/main/graph/badge.svg)](https://codecov.io/gh/shuldan/config)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

Пакет `config` предоставляет простой, типобезопасный и расширяемый способ загрузки, объединения и доступа к конфигурации из различных источников — файлов (JSON/YAML), переменных окружения и произвольных структур Go.

---

## 🚀 Основные возможности

- **Мультиисточниковость**: загружайте конфиг из `.json`, `.yaml`, переменных окружения и `map[string]any`.
- **Гибкий доступ к данным**: поддержка вложенных ключей через точку (`parent.child.key`).
- **Автоматическое приведение типов**: получайте значения как `string`, `int`, `bool`, `float64`, `[]string` и др. — даже если в источнике они в другом формате.
- **Шаблонизация значений**: используйте Go-шаблоны внутри строк (`{{env "PORT"}}`, `{{default "8080" .PORT}}` и т.д.).
- **Безопасность**: предотвращение path traversal при загрузке файлов.
- **Полное покрытие тестами**: строгие unit-тесты на все граничные случаи.
- **Цепочка загрузчиков**: несколько источников объединяются в один конфиг с приоритетом (первый — самый низкий).

---

## 📦 Установка

Убедитесь, что у вас установлен **Go 1.24+**.

```sh
go get github.com/shuldan/config
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

Выполняет:
- проверку форматирования,
- статический анализ,
- security-сканирование,
- запуск тестов.

### CI-проверка

```sh
make ci
```

Аналогично тому, что запускается в GitHub Actions.

---

## 🧱 Архитектура

### `Config`

Основной объект для доступа к конфигурации:

```go
cfg, err := config.New(
    config.FromYaml("config.yaml"),
    config.FromEnv("APP_"),
)
```

Поддерживает методы:

- `Has(key string) bool`
- `Get(key string) any`
- `GetString(key string, defaultVal ...string) string`
- `GetInt(key string, defaultVal ...int) int`
- `GetBool(key string, defaultVal ...bool) bool`
- `GetStringSlice(key string, separator ...string) []string`
- `GetSub(key string) (*Config, bool)`
- `All() map[string]any`

### `Loader`

Интерфейс для подключения новых источников:

```go
type Loader interface {
    Load() (map[string]any, error)
}
```

Встроенные загрузчики:

- `FromYaml("app.yaml", "local.yaml")`
- `FromJSON("config.json")`
- `FromEnv("APP_")`
- `FromMap(map[string]any{"key": "value"})` (часто используется в тестах)

---

## 🧪 Пример использования

### Загрузка из YAML и переменных окружения

```yaml
# config.yaml
server:
  host: localhost
  port: "{{env \"PORT\" | default \"8080\"}}"
database:
  url: "postgres://user:pass@localhost:5432/mydb"
debug: false
```

```go
package main

import (
	"log/slog"
	"github.com/shuldan/config"
)

func main() {
	cfg, err := config.New(
		config.FromYaml("config.yaml"),
		config.FromEnv("APP_"),
	)
	if err != nil {
		panic(err)
	}

	host := cfg.GetString("server.host")
	port := cfg.GetInt("server.port")
	dbURL := cfg.GetString("database.url")
	debug := cfg.GetBool("debug")

	slog.Info("Config loaded",
		"host", host,
		"port", port,
		"db", dbURL,
		"debug", debug,
	)
}
```

### Использование только переменных окружения

```sh
export APP_SERVER__HOST=0.0.0.0
export APP_SERVER__PORT=3000
export APP_DEBUG=true
```

```go
cfg, _ := config.New(config.FromEnv("APP_"))
port := cfg.GetInt("server.port") // 3000
debug := cfg.GetBool("debug")     // true
```

> Двойное подчёркивание (`__`) в имени переменной преобразуется в точку (`.`).

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