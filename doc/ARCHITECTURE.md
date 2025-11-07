# Архитектура pm

## Общая схема

```
┌─────────────────┐
│   pm (wrapper)  │  Shell скрипт (pm.sh / pm.ps1)
│   bash/pwsh     │
└────────┬────────┘
         │ вызывает
         ▼
┌─────────────────┐
│    pm-bin       │  Go бинарник
│   (главный)     │  Парсит аргументы, строит план
└────────┬────────┘
         │ генерирует
         ▼
┌─────────────────┐
│  Shell Script   │  Bash/PowerShell команды
│  (output)       │  Готовые к выполнению
└────────┬────────┘
         │ выполняется через eval
         ▼
┌─────────────────┐
│  Actual Work    │  Docker, git, build tools, etc.
└─────────────────┘
```

## Почему двухфазная архитектура?

### Проблема
Если бы pm-bin выполнял команды напрямую:
- `cd` в Go процессе не меняет директорию shell
- Environment переменные Go процесса изолированы
- Сложности с pipes, redirects, job control

### Решение
pm-bin **строит команды**, wrapper **выполняет их**:
1. pm-bin читает конфиг и аргументы
2. Строит абстрактный план (Plan) с операциями
3. Рендерит план в shell script (bash/pwsh)
4. Wrapper выполняет через `eval`

### Преимущества
- ✅ `pushd`/`popd` работают корректно
- ✅ Environment переменные доступны
- ✅ Pipes, redirects, job control работают нативно
- ✅ Один бинарник для Windows и Linux
- ✅ Можно добавить любой shell (fish, zsh, cmd)

## Модули

### 1. cmd/pm-bin/main.go

**Роль**: Entry point, CLI парсинг

**Ответственность**:
- Парсинг флагов (`--dialect`, `--plugins`)
- Обработка top-level команд (`add`, `rm`, `ls`)
- Координация всего процесса
- Вывод результата

**Поток**:
```
Аргументы → Registry commands (add/rm/ls)
         ↓
         Project resolution (по имени или пути)
         ↓
         DSL parsing (:build :test args)
         ↓
         Plan building
         ↓
         Template rendering
         ↓
         Script rendering
         ↓
         Output
```

### 2. internal/config

**Файлы**:
- `types.go` - структуры данных
- `registry.go` - работа с реестром проектов

**Типы**:
```go
ProjectMeta {
    Info { Name, Description, Root }
    Func map[string]FuncDef
    Commands map[string]CommandDef
    Docker DockerDef
}

FuncDef {
    Params map[string]ParamMeta
    Script any  // string или []string
}

CommandDef {
    Description string
    Cmd any  // string или []string
}
```

**Registry**:
- Хранится в `~/.config/pm/registry.yml`
- Содержит список проектов с путями к meta файлам
- API: `RegAdd()`, `RegRm()`, `RegLs()`, `ResolveProject()`

**Global Config**:
- Путь: `~/.config/pm/global.yml`
- Глобальные функции и переменные
- Доступны через `#{global.*}` и `_{global.func()}`

### 3. internal/dsl

**Файл**: `parse.go`

**Роль**: Парсинг command-line DSL

**Синтаксис**:
```
:build -x test :run @{args}
│      └─────┘ │    └──────┘
│      args    │    args
└─ command     └─ command

docker ps -a
└─────────────┘
   RAW mode
```

**Функция**:
```go
SplitColonCommands(args []string) []Chunk

Chunk {
    Name string   // "build" или "__RAW__"
    Args []string // ["-x", "test"]
}
```

**Правила**:
- Если первый токен начинается с `:` → named mode
- Иначе → RAW mode (всё как есть)
- Named mode: каждый `:cmd` → новый Chunk с args до следующего `:`

### 4. internal/templ

**Файл**: `render.go`

**Роль**: Шаблонизатор с подстановкой

**Паттерны**:
```
@{param}        - параметры функции/команды
${ENV_VAR}      - environment переменные
#{config.path}  - значения из конфига
_{func(args)}   - вызов функций
```

**Regex паттерны**:
```go
envRe   = `\$\{([A-Za-z_][A-Za-z0-9_]*)\}`
paramRe = `@\{([A-Za-z0-9_.-]+)\}`
cfgRe   = `#\{([^}]+)\}`
funcRe  = `_\{([A-Za-z0-9_.-]+)\((.*?)\)\}`
```

**Функция**:
```go
RenderString(
    text string,
    params map[string]string,     // @{param}
    proj *ProjectMeta,            // #{info.*}
    global *GlobalConfig,         // #{global.*}
    ctx []map[string]any,         // для относительных путей
) (string, error)
```

**Порядок подстановки**:
1. `${ENV}` → environment
2. `@{param}` → params
3. `#{path}` → config
4. `_{func(...)}` → function call (рекурсивно)

### 5. internal/plan

**Файл**: `plan.go`

**Роль**: Абстрактный план выполнения

**Операции**:
```go
type Op interface{ isOp() }

OpPushd { Dir string }  // cd в директорию
OpPopd  {}              // вернуться назад
OpRun   { Line string } // выполнить команду
OpEcho  { Line string } // вывести сообщение
```

**Plan**:
```go
type Plan struct {
    Ops []Op
}

// API
func New() *Plan
func (p *Plan) Pushd(dir string)
func (p *Plan) Popd()
func (p *Plan) Run(line string)
func (p *Plan) Echo(line string)
```

**Использование**:
```go
pl := plan.New()
pl.Pushd("/path/to/project")
pl.Run("npm install")
pl.Run("npm build")
pl.Popd()
```

### 6. internal/docker

**Файл**: `up.go` (логика в plan.go)

**Роль**: Построение docker compose команд

**Функция**:
```go
DockerUp(meta *ProjectMeta, args []string) []string
```

**Логика**:
- Разворачивает `@group` в список сервисов
- Строит команду `docker compose -f FILE up -d SERVICES...`
- Неизвестные группы передаются как есть (безопасно)

**Пример**:
```go
// groups: {base: [db, redis], app: [api]}
DockerUp(meta, []string{"@base", "nginx"})
// → ["docker compose -f file.yml up -d db redis nginx"]
```

### 7. internal/render

**Файлы**:
- `plugin.go` - общий интерфейс и выбор рендерера
- `bash.go` - bash рендерер
- `pwsh.go` - PowerShell рендерер

**Интерфейс**:
```go
type Renderer interface {
    Name() string
    Begin(root string) []string
    RenderOp(op Op) []string
    End() []string
}
```

**Bash рендерер**:
```bash
# Begin
pushd /path >/dev/null

# OpRun
command here

# OpEcho
echo 'message'

# End
popd >/dev/null
```

**PowerShell рендерер**:
```powershell
# Begin
Push-Location 'C:\path'

# OpRun
command here

# OpEcho
Write-Host 'message'

# End
Pop-Location
```

**External plugins**:
- Исполняемые файлы в `~/.config/pm/plugins/`
- Имя: `pm-render-DIALECT`
- Принимают JSON план через stdin
- Выводят shell script в stdout

## Поток данных

### Пример: `pm myproject :build -x test`

#### 1. Парсинг (dsl)
```
[":build", "-x", "test"] → [Chunk{Name: "build", Args: ["-x", "test"]}]
```

#### 2. Резолвинг (config)
```
"myproject" → ProjectMeta + root path
```

#### 3. Plan building
```go
pl := plan.New()
pl.Pushd("/path/to/myproject")

// Для :build команды
cmd := meta.Commands["build"]
for _, line := range cmd.AsLines() {
    rendered := templ.RenderString(line,
        map[string]string{"args": "-x test"},
        meta, global, nil)
    pl.Run(rendered)
}

pl.Popd()
```

#### 4. Template rendering
```yaml
# Конфиг
commands:
  build:
    cmd:
      - "_{use-java(version=21)}"
      - "mvn clean install @{args}"

# После рендеринга
pl.Run("sdk use java 21")
pl.Run("mvn clean install -x test")
```

#### 5. Script rendering
```bash
# pm begin
pushd /path/to/myproject >/dev/null
sdk use java 21
mvn clean install -x test
popd >/dev/null
# pm end
```

#### 6. Execution (wrapper)
```bash
eval "$script"
```

## Тестирование

### Unit тесты

**internal/config/types_test.go**:
- Тесты структур данных
- YAML unmarshaling

**internal/dsl/parse_test.go**:
- Парсинг команд
- RAW vs Named mode

**internal/templ/render_test.go**:
- Подстановка переменных
- Функции

**internal/plan/plan_test.go**:
- Построение плана
- Docker команды

**internal/render/render_test.go**:
- Bash рендеринг
- PowerShell рендеринг

**internal/docker/up_test.go**:
- Docker compose группы

### E2E тесты

**internal/e2e/e2e_test.go**:

Полный цикл:
1. Создать temp директорию
2. Написать `.pm.meta.yml`
3. Зарегистрировать через `RegAdd()`
4. Запустить весь пайплайн
5. Проверить финальный скрипт

**Покрытие**:
- RAW команды
- Named команды с аргументами
- Функции (локальные и глобальные)
- Шаблоны (@{}, ${}, #{})
- Docker compose с группами
- Multiline функции
- PowerShell dialect
- Error cases

## Расширяемость

### Добавление нового рендерера

#### Встроенный (Go)
```go
// internal/render/fish.go
type fishRenderer struct{}

func (f fishRenderer) Name() string { return "fish" }

func (f fishRenderer) Begin(root string) []string {
    return []string{
        "pushd " + quote(root),
    }
}
// ...

// internal/render/plugin.go
var builtins = map[string]Renderer{
    "bash": bashRenderer{},
    "pwsh": pwshRenderer{},
    "fish": fishRenderer{},  // добавить
}
```

#### External plugin
```bash
#!/usr/bin/env python3
# ~/.config/pm/plugins/pm-render-fish

import json
import sys

plan = json.load(sys.stdin)

print(f"pushd {plan['root']}")

for op in plan['ops']:
    if op['kind'] == 'run':
        print(op['line'])
    elif op['kind'] == 'echo':
        print(f"echo {op['msg']}")
    elif op['kind'] == 'pushd':
        print(f"pushd {op['dir']}")
    elif op['kind'] == 'popd':
        print("popd")

print("popd")
```

Использование:
```bash
pm --dialect fish myproject :build
```

### Добавление новых операций

```go
// internal/plan/plan.go
type OpCd struct{ Dir string }
func (OpCd) isOp() {}

func (p *Plan) Cd(dir string) {
    p.Ops = append(p.Ops, OpCd{Dir: dir})
}

// internal/render/bash.go
func (b bashRenderer) RenderOp(op plan.Op) []string {
    switch v := op.(type) {
    case plan.OpCd:
        return []string{fmt.Sprintf("cd %s", sh(v.Dir))}
    // ...
    }
}
```

## Производительность

### Оптимизации

1. **Lazy loading**: Глобальный конфиг загружается только при необходимости
2. **Minimal I/O**: Только один read для registry и meta файлов
3. **No external deps**: Только `gopkg.in/yaml.v3`, всё остальное stdlib
4. **Fast compilation**: Маленький бинарник, быстрая компиляция

### Бенчмарки

```bash
# Холодный старт (с чтением файлов)
time pm myproject :build
# ~20-50ms

# Registry операции
time pm ls
# ~10-20ms

# Raw команды (минимальный overhead)
time pm myproject docker ps
# ~15-30ms
```

## Безопасность

### Что НЕ делаем

- ❌ Не используем `sh -c` с пользовательским вводом
- ❌ Не eval'им пользовательские строки в Go
- ❌ Не модифицируем файловую систему без явного запроса

### Что делаем

- ✅ Экранируем спецсимволы при генерации shell команд
- ✅ Проверяем существование файлов перед операциями
- ✅ Валидируем структуру YAML конфигов
- ✅ Изолируем функции через параметры (не глобальный scope)

### Escape rules

**Bash**:
```go
func sh(s string) string {
    if strings.ContainsAny(s, " \t\"'") {
        return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
    }
    return s
}
```

**PowerShell**:
```go
func pwshQuote(s string) string {
    return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}
```

## Известные ограничения

1. **Относительные пути в #{.}**
   - Реализованы, но сложны в использовании
   - Лучше использовать абсолютные пути

2. **Вложенные вызовы функций**
   - `_{func1(x=_{func2()})}` - не поддерживается
   - Нужно разбивать на отдельные команды

3. **Условная логика**
   - Нет встроенных `if/else`
   - Можно эмулировать через shell команды

4. **Циклы**
   - Нет встроенной поддержки
   - Можно использовать shell loops в script

5. **Интерполяция в середине функции**
   - `_{func(x=prefix@{var}suffix)}` может не работать
   - Лучше: `_{func(x=@{fullvar})}`

## Будущие улучшения

### Возможные фичи

1. **Интерактивный режим**
   - `pm` без аргументов → TUI с выбором проектов/команд

2. **Completion**
   - Bash/Zsh/Fish completion scripts

3. **Validation**
   - `pm validate` - проверка корректности конфигов

4. **Watch mode**
   - `pm watch :build` - пересборка при изменениях

5. **Dependency graph**
   - Команды зависящие друг от друга

6. **Hooks**
   - Pre/post hooks для команд

7. **Secrets management**
   - Интеграция с vault/pass/keychain
