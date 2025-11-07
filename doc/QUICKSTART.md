# Quick Start Guide

## Установка (5 минут)

### Linux/WSL

```bash
# 1. Клонировать репозиторий
git clone <repo-url>
cd project-manager

# 2. Собрать бинарник
go build -o pm-bin ./cmd/pm-bin

# 3. Установить wrapper
chmod +x pm.sh
sudo ln -s $(pwd)/pm.sh /usr/local/bin/pm

# 4. Проверить
pm ls
```

### Windows (PowerShell)

```powershell
# 1. Клонировать репозиторий
git clone <repo-url>
cd project-manager

# 2. Собрать бинарник
go build -o pm-bin.exe ./cmd/pm-bin

# 3. Добавить в PATH или создать alias
# Вариант 1: Добавить директорию в System PATH через System Properties

# Вариант 2: Создать alias в PowerShell профиле
notepad $PROFILE
# Добавить строку:
# Set-Alias pm C:\path\to\project-manager\pm.ps1

# 4. Проверить
.\pm.ps1 ls
# или просто: pm ls (если добавили alias)
```

## Первый проект (2 минуты)

### 1. Создать конфиг

Создайте файл `.pm.meta.yml` в корне вашего проекта:

```yaml
info:
  name: myapp
  description: My awesome application
  root: ~/repos/myapp  # Windows: C:\repos\myapp

commands:
  install:
    description: "Install dependencies"
    cmd: "npm install"

  dev:
    description: "Start development server"
    cmd: "npm run dev"

  build:
    description: "Build for production"
    cmd: "npm run build"

  test:
    description: "Run tests"
    cmd: "npm test -- @{args}"
```

### 2. Добавить проект

```bash
pm add ~/repos/myapp/.pm.meta.yml
```

### 3. Использовать

```bash
# Посмотреть список проектов
pm ls

# Посмотреть доступные команды
pm myapp :help

# Установить зависимости
pm myapp :install

# Запустить dev сервер
pm myapp :dev

# Собрать
pm myapp :build

# Запустить тесты
pm myapp :test --watch
```

## Добавить Docker (1 минута)

Обновите `.pm.meta.yml`:

```yaml
# ... предыдущий конфиг ...

docker:
  compose_file: docker-compose.yml
  groups:
    db: [postgres, redis]
    app: [api, worker]
```

Использование:

```bash
# Запустить только базы
pm myapp :up @db

# Запустить всё
pm myapp :up @db @app

# Логи
pm myapp docker compose logs -f api
```

## Добавить функции (2 минуты)

Для переиспользования кода:

```yaml
info:
  name: backend
  root: ~/repos/backend

func:
  use-java:
    params:
      version:
        default: "21"
    script: "sdk use java @{version}"

  check-env:
    params:
      file:
        required: true
    script: |
      test -f @{file} || (echo "@{file} not found" && exit 1)

commands:
  build:
    description: "Build project"
    cmd:
      - "_{use-java(version=17)}"
      - "./gradlew build"

  deploy:
    description: "Deploy to production"
    cmd:
      - "_{check-env(file=.env.prod)}"
      - "_{use-java()}"
      - "./deploy.sh"
```

## Глобальные функции (опционально)

Создайте `~/.config/pm/global.yml`:

```yaml
func:
  notify:
    params:
      message:
        required: true
    script: "echo '📢 @{message}'"

  git-check-clean:
    script: |
      git diff --quiet || (echo "Git working directory is not clean" && exit 1)

vars:
  docker_registry: registry.example.com
  default_region: us-east-1
```

Используйте в проектах:

```yaml
commands:
  deploy:
    cmd:
      - "_{global.git-check-clean()}"
      - "docker build -t #{global.vars.docker_registry}/myapp ."
      - "_{global.notify(message='Deployed!')}"
```

## Полный пример (Java проект)

`.pm.meta.yml`:

```yaml
info:
  name: backend-api
  description: Spring Boot REST API
  root: ~/repos/backend-api

func:
  use-java:
    params:
      version:
        default: "21.0.8-tem"
    script: "sdk use java @{version}"

commands:
  clean:
    description: "Clean build"
    cmd:
      - "_{use-java()}"
      - "./gradlew clean"

  build:
    description: "Build project"
    cmd:
      - "_{use-java()}"
      - "./gradlew build -x test @{args}"

  test:
    description: "Run tests"
    cmd:
      - "_{use-java()}"
      - "./gradlew test @{args}"

  run:
    description: "Run application"
    cmd:
      - "_{use-java()}"
      - "./gradlew bootRun"

  docker-build:
    description: "Build Docker image"
    cmd:
      - "./gradlew bootBuildImage"

  deploy:
    description: "Deploy to k8s"
    cmd:
      - "kubectl apply -f k8s/"
      - "kubectl rollout status deployment/backend-api"

docker:
  compose_file: docker-compose.yml
  groups:
    db: [postgres]
    cache: [redis]
    infra: [postgres, redis, kafka]
```

Использование:

```bash
# Разработка
pm backend-api :up @infra
pm backend-api :build
pm backend-api :run

# Тестирование
pm backend-api :test --tests UserServiceTest
pm backend-api :test --continuous

# Деплой
pm backend-api :clean :build :test :docker-build :deploy

# Просмотр логов Docker
pm backend-api docker compose logs -f postgres
```

## Продвинутые возможности

### Множественные команды

```bash
# Выполнить несколько команд подряд
pm myapp :clean :install :build :test

# С разными аргументами
pm myapp :clean :build --prod :test --coverage
```

### RAW команды

```bash
# Выполнить любую команду в директории проекта
pm myapp git status
pm myapp docker ps
pm myapp ls -la
pm myapp npm run custom-script
```

### Environment переменные

```yaml
commands:
  deploy:
    cmd: "kubectl apply -f deployment.yaml --namespace=${K8S_NAMESPACE}"
```

```bash
K8S_NAMESPACE=production pm myapp :deploy
```

### Подстановка из конфига

```yaml
info:
  name: myapp
  root: ~/repos/myapp

commands:
  backup:
    cmd: "tar -czf backup-#{info.name}-$(date +%Y%m%d).tar.gz #{info.root}"
```

## Советы

### 1. Используйте :help

```bash
pm myapp :help
```

### 2. Создавайте группы Docker сервисов

```yaml
docker:
  groups:
    minimal: [postgres]              # Минимум
    dev: [postgres, redis]           # Для разработки
    full: [postgres, redis, kafka]   # Всё
```

### 3. Выносите часто используемое в функции

❌ Плохо:
```yaml
commands:
  build-dev:
    cmd:
      - "sdk use java 21"
      - "gradlew build"
  build-prod:
    cmd:
      - "sdk use java 21"
      - "gradlew build -Pprod"
```

✅ Хорошо:
```yaml
func:
  use-java:
    script: "sdk use java 21"

commands:
  build-dev:
    cmd:
      - "_{use-java()}"
      - "gradlew build"
  build-prod:
    cmd:
      - "_{use-java()}"
      - "gradlew build -Pprod"
```

### 4. Используйте @{args} для гибкости

```yaml
commands:
  test:
    cmd: "npm test -- @{args}"

  logs:
    cmd: "docker compose logs @{args}"
```

```bash
pm myapp :test --watch --coverage
pm myapp :logs -f --tail=100 api
```

## Troubleshooting

### pm: command not found (Linux)

```bash
# Проверить установку
which pm

# Если не установлено
cd /path/to/project-manager
sudo ln -s $(pwd)/pm.sh /usr/local/bin/pm
```

### pm-bin.exe not found (Windows)

```powershell
# Собрать бинарник
cd C:\path\to\project-manager
go build -o pm-bin.exe ./cmd/pm-bin

# Проверить
.\pm-bin.exe -h
```

### Project not found

```bash
# Проверить список проектов
pm ls

# Добавить проект
pm add ~/repos/myapp/.pm.meta.yml

# Или использовать напрямую путь к конфигу
pm ~/repos/myapp/.pm.meta.yml :build
```

### Unknown command

```bash
# Проверить доступные команды
pm myapp :help

# Проверить конфиг
cat ~/repos/myapp/.pm.meta.yml
```

## Что дальше?

- Прочитать [README.md](../README.md) для полной документации
- Посмотреть [EXAMPLES.md](EXAMPLES.md) для реальных сценариев
- Изучить [samples/](samples/) для примеров конфигов
- Прочитать [ARCHITECTURE.md](ARCHITECTURE.md) чтобы понять как всё устроено

## Получить помощь

1. Посмотреть доступные команды: `pm myapp :help`
2. Проверить конфиг: `cat path/to/.pm.meta.yml`
3. Проверить генерируемый скрипт: `pm-bin --dialect bash myapp :build`
4. Создать issue на GitHub

Удачи! 🚀
