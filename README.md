# pm - Project Manager

Утилита для управления микросервисами и проектами через единый интерфейс.

## Основная идея

pm позволяет:
- Хранить информацию о проектах в YAML конфигах
- Определять custom команды для каждого проекта
- Использовать функции (многократное использование кода)
- Работать с docker-compose группами сервисов
- Поддерживать шаблоны с подстановкой параметров, env переменных и конфига

## Установка

```bash
# Клонировать репозиторий
git clone <repo-url>
cd project-manager

# Собрать бинарник
go build -o pm-bin ./cmd/pm-bin

# Для Linux/WSL
chmod +x pm.sh
sudo ln -s $(pwd)/pm.sh /usr/local/bin/pm

# Для Windows (PowerShell)
# Добавить директорию в PATH или создать alias
```

### Автодополнение (Shell Completion)

#### Для Zsh / Oh My Zsh

```bash
# Скопировать файл автодополнения
sudo cp shell/pm-completion.zsh /usr/local/share/zsh/site-functions/_pm

# Или для Oh My Zsh
mkdir -p ~/.oh-my-zsh/completions
cp shell/pm-completion.zsh ~/.oh-my-zsh/completions/_pm

# Перезагрузить shell
exec zsh
```

Или добавить в `~/.zshrc`:
```bash
source /path/to/project-manager/shell/pm-completion.zsh
```

#### Для Bash

```bash
# Добавить в ~/.bashrc
echo "source /path/to/project-manager/shell/pm-completion.bash" >> ~/.bashrc

# Или скопировать в системную директорию
sudo cp shell/pm-completion.bash /etc/bash_completion.d/pm

# Перезагрузить shell
source ~/.bashrc
```

#### Для Windows PowerShell

PowerShell использует встроенный скрипт автодополнения `shell/pm-completion.ps1`.

После установки автодополнение будет предлагать:
- Зарегистрированные проекты
- Команды проекта (`:build`, `:test`, etc.)
- Docker группы (`@base`, `@app`, etc.)

## Быстрый старт

### 1. Создать конфиг проекта

Создайте `.pm.meta.yml` в корне вашего проекта:

```yaml
info:
  name: myproject
  description: My awesome project
  root: ~/repos/myproject

func:
  use-java:
    params:
      version:
        required: true
        default: "21.0.8-tem"
    script: "sdk use java @{version}"

commands:
  build:
    description: "Build project"
    cmd:
      - "_{use-java(version=21.0.8-tem)}"
      - "./gradlew build @{args}"

  run:
    description: "Start services"
    cmd: "docker compose up -d @{args}"

  test:
    description: "Run tests"
    cmd: "./gradlew test @{args}"

docker:
  compose_file: docker-compose.yml
  groups:
    base: [postgres, redis]
    app: [api, worker]
```

### 2. Добавить проект в реестр

```bash
pm add ~/repos/myproject/.pm.meta.yml
```

### 3. Использовать

```bash
# Просмотр зарегистрированных проектов
pm ls

# Выполнить команду
pm myproject :build -x test

# Выполнить несколько команд
pm myproject :build :test :run

# Запустить docker compose с группой сервисов
pm myproject :up @base @app nginx

# Выполнить произвольную команду в директории проекта
pm myproject docker compose logs -f api

# Помощь по командам проекта
pm myproject :help
```

## Синтаксис конфига

### Секция info

```yaml
info:
  name: project-name        # Имя проекта (обязательно)
  description: Description  # Описание
  root: ~/repos/project     # Корневая директория (обязательно)
```

### Секция func (функции)

Функции позволяют переиспользовать код:

```yaml
func:
  setup-env:
    params:
      env:
        required: true
      debug:
        required: false
        default: "false"
    script:
      - "export ENV=@{env}"
      - "export DEBUG=@{debug}"
```

Вызов функции:
```yaml
cmd: "_{setup-env(env=prod, debug=true)}"
```

### Секция commands

```yaml
commands:
  build:
    description: "Build project"
    cmd: "make build"          # Одна команда (string)

  deploy:
    description: "Deploy"
    cmd:                       # Несколько команд (array)
      - "_{setup-env(env=prod)}"
      - "docker build -t app ."
      - "docker push app"
```

### Секция docker

```yaml
docker:
  compose_file: docker-compose.yml
  groups:
    base: [postgres, redis, kafka]
    app: [api, worker, scheduler]
```

Использование:
```bash
pm myproject :up @base        # Запустит postgres, redis, kafka
pm myproject :up @app api-v2  # Запустит api, worker, scheduler, api-v2
```

## Подстановка переменных

### 1. Параметры команд/функций: `@{name}`

```yaml
commands:
  deploy:
    cmd: "kubectl apply -f @{args}"
```

```bash
pm myproject :deploy deployment.yaml
# → kubectl apply -f deployment.yaml
```

### 2. Environment переменные: `${VAR}`

```yaml
commands:
  show:
    cmd: "echo Current user: ${USER}"
```

### 3. Значения из конфига: `#{path.to.value}`

```yaml
commands:
  info:
    cmd: "echo Project: #{info.name} at #{info.root}"
```

### 4. Глобальный конфиг: `#{global.*}`

Создайте `~/.config/pm/global.yml`:

```yaml
func:
  notify:
    params:
      message:
        required: true
    script: "notify-send '@{message}'"

vars:
  default_branch: main
  docker_registry: registry.example.com
```

Использование:
```yaml
commands:
  done:
    cmd:
      - "_{global.notify(message='Build complete')}"
      - "echo Branch: #{global.vars.default_branch}"
```

## Архитектура

```
pm (shell wrapper)
  ↓
pm-bin (Go binary) - строит план команд
  ↓
Shell script - выполняется через eval
```

### Модули

- `internal/config` - работа с конфигами и реестром проектов
- `internal/dsl` - парсинг аргументов командной строки
- `internal/templ` - шаблонизатор с подстановкой переменных
- `internal/plan` - построение плана выполнения
- `internal/render` - генерация bash/pwsh скриптов
- `internal/docker` - работа с docker compose

## Поддержка платформ

- **Linux/WSL**: Используйте `pm.sh` (bash)
- **Windows**: Используйте `pm.ps1` (PowerShell)

Бинарник `pm-bin` работает на обеих платформах и генерирует соответствующие скрипты.

## Тестирование

```bash
# Запустить все тесты
go test ./...

# E2E тесты (полный цикл от конфига до команд)
go test ./internal/e2e -v

# Unit тесты отдельного модуля
go test ./internal/templ -v
```

## Примеры использования

### Пример 1: Java проект с Maven

```yaml
info:
  name: backend-api
  root: ~/repos/backend-api

func:
  use-java:
    params:
      version:
        default: "17"
    script: "sdk use java @{version}"

commands:
  build:
    cmd:
      - "_{use-java()}"
      - "mvn clean install @{args}"

  run:
    cmd:
      - "_{use-java()}"
      - "mvn spring-boot:run"

docker:
  groups:
    infra: [postgres, redis]
```

### Пример 2: Frontend проект

```yaml
info:
  name: frontend
  root: ~/repos/frontend

commands:
  install:
    description: "Install dependencies"
    cmd: "npm install @{args}"

  dev:
    description: "Run dev server"
    cmd: "npm run dev @{args}"

  build:
    description: "Build for production"
    cmd:
      - "npm run build"
      - "echo Build completed at #{info.root}/dist"
```

### Пример 3: Микросервисы с docker

```yaml
info:
  name: microservices
  root: ~/repos/services

commands:
  logs:
    description: "View logs"
    cmd: "docker compose logs -f @{args}"

  restart:
    description: "Restart service"
    cmd: "docker compose restart @{args}"

docker:
  compose_file: docker-compose.yml
  groups:
    db: [postgres, mongodb]
    cache: [redis, memcached]
    app: [api, worker, scheduler]
    monitoring: [prometheus, grafana]
```

Использование:
```bash
# Запустить базы и кеш
pm microservices :up @db @cache

# Запустить приложение
pm microservices :up @app

# Посмотреть логи API
pm microservices :logs api

# Перезапустить worker
pm microservices :restart worker
```

## Расширенные возможности

### Относительные пути в конфиге (. .. ...)

```yaml
func:
  deploy:
    params:
      env:
        required: true
    script:
      - "echo Deploying #{.env}"  # доступ к параметру функции
```

### Внешние плагины рендереров

Можно создать свой рендерер для другой оболочки:

```bash
# ~/.config/pm/plugins/pm-render-fish
#!/usr/bin/env fish
# Читает JSON план из stdin, выводит fish скрипт
```

Использование:
```bash
pm --dialect fish myproject :build
```

## Удаление проекта

```bash
pm rm myproject
```

## Environment переменные

- `PM_CONFIGS` - директория для конфигов (default: `~/.config/pm`)
- `PM_DIALECT` - dialect для рендеринга (bash/pwsh/custom)
- `PM_PLUGIN_DIR` - директория с плагинами (default: `~/.config/pm/plugins`)
- `PM_BIN` - путь к pm-bin бинарнику

## Лицензия

MIT
