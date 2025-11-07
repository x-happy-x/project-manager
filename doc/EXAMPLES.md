# Примеры использования pm

## Содержание

1. [Основные команды](#основные-команды)
2. [Работа с функциями](#работа-с-функциями)
3. [Docker Compose](#docker-compose)
4. [Шаблоны и переменные](#шаблоны-и-переменные)
5. [Реальные сценарии](#реальные-сценарии)

## Основные команды

### Добавление проекта

```bash
# Создать конфиг
cat > ~/repos/myapp/.pm.meta.yml << 'EOF'
info:
  name: myapp
  description: My Application
  root: ~/repos/myapp

commands:
  start:
    description: Start application
    cmd: "npm start"
EOF

# Добавить в реестр
pm add ~/repos/myapp/.pm.meta.yml

# Проверить
pm ls
```

### Выполнение команд

```bash
# Выполнить одну команду
pm myapp :start

# Выполнить несколько команд подряд
pm myapp :install :build :test

# Передать аргументы команде
pm myapp :build --prod --verbose

# Выполнить произвольную команду в директории проекта
pm myapp npm run dev

# Посмотреть доступные команды
pm myapp :help
```

## Работа с функциями

### Простая функция

```yaml
func:
  hello:
    params:
      name:
        required: true
    script: "echo Hello, @{name}!"

commands:
  greet:
    cmd: "_{hello(name=World)}"
```

```bash
pm myapp :greet
# Output: Hello, World!
```

### Функция с несколькими командами

```yaml
func:
  setup-java:
    params:
      version:
        required: true
        default: "21"
    script:
      - "sdk use java @{version}"
      - "java -version"
      - "echo Java @{version} is ready"

commands:
  build:
    cmd:
      - "_{setup-java(version=17)}"
      - "./gradlew build"
```

### Функция с необязательными параметрами

```yaml
func:
  deploy:
    params:
      env:
        required: true
      region:
        required: false
        default: "us-east-1"
      dry_run:
        required: false
        default: "false"
    script:
      - "echo Deploying to @{env} in @{region}"
      - "echo Dry run: @{dry_run}"

commands:
  deploy-prod:
    cmd: "_{deploy(env=production, region=eu-west-1)}"

  deploy-dev:
    cmd: "_{deploy(env=development)}"
```

## Docker Compose

### Базовая настройка

```yaml
docker:
  compose_file: docker-compose.yml
  groups:
    db: [postgres, redis]
    app: [api, worker]
    monitoring: [prometheus, grafana, loki]
```

### Использование групп

```bash
# Запустить только базы данных
pm myapp :up @db

# Запустить базы и приложение
pm myapp :up @db @app

# Запустить группу + дополнительные сервисы
pm myapp :up @db nginx elasticsearch

# Запустить всё без аргументов
pm myapp :up

# Остановить определённые сервисы
pm myapp docker compose down @db
```

### Команды для работы с Docker

```yaml
commands:
  up:
    # встроенная команда для docker compose up -d
    # использует docker.groups

  logs:
    description: "Show logs"
    cmd: "docker compose logs -f @{args}"

  ps:
    description: "List containers"
    cmd: "docker compose ps @{args}"

  restart:
    description: "Restart services"
    cmd: "docker compose restart @{args}"

  exec:
    description: "Execute command in container"
    cmd: "docker compose exec @{args}"
```

```bash
# Примеры использования
pm myapp :logs api
pm myapp :ps
pm myapp :restart worker
pm myapp :exec api bash
```

## Шаблоны и переменные

### Параметры команд (@{args})

```yaml
commands:
  test:
    cmd: "npm test -- @{args}"

  deploy:
    cmd: "kubectl apply -f @{args}"
```

```bash
pm myapp :test --watch --coverage
pm myapp :deploy deployment.yaml service.yaml
```

### Environment переменные (${VAR})

```yaml
commands:
  show-env:
    cmd:
      - "echo User: ${USER}"
      - "echo Home: ${HOME}"
      - "echo Path: ${PATH}"

  use-token:
    cmd: "curl -H 'Authorization: Bearer ${API_TOKEN}' https://api.example.com"
```

```bash
# Установить переменную и выполнить
API_TOKEN=secret pm myapp :use-token
```

### Конфиг проекта (#{info.*})

```yaml
info:
  name: myapp
  description: My Application
  root: ~/repos/myapp

commands:
  show-info:
    cmd:
      - "echo Project: #{info.name}"
      - "echo Description: #{info.description}"
      - "echo Root: #{info.root}"

  backup:
    cmd: "tar -czf backup-#{info.name}.tar.gz #{info.root}"
```

### Глобальный конфиг (#{global.*})

`~/.config/pm/global.yml`:
```yaml
func:
  slack-notify:
    params:
      message:
        required: true
    script: "curl -X POST ${SLACK_WEBHOOK} -d '{\"text\": \"@{message}\"}'"

vars:
  docker_registry: registry.example.com
  default_region: us-east-1
```

В проекте:
```yaml
commands:
  build-and-notify:
    cmd:
      - "docker build -t #{global.vars.docker_registry}/myapp:latest ."
      - "_{global.slack-notify(message='Build complete')}"
```

## Реальные сценарии

### Сценарий 1: Микросервис на Node.js

```yaml
info:
  name: user-service
  root: ~/repos/user-service

func:
  use-node:
    params:
      version:
        default: "20"
    script: "nvm use @{version}"

  check-env:
    params:
      env:
        required: true
    script:
      - "test -f .env.@{env} || (echo '.env.@{env} not found' && exit 1)"
      - "cp .env.@{env} .env"

commands:
  install:
    description: "Install dependencies"
    cmd:
      - "_{use-node()}"
      - "npm install"

  dev:
    description: "Start development server"
    cmd:
      - "_{check-env(env=dev)}"
      - "_{use-node()}"
      - "npm run dev"

  build:
    description: "Build for production"
    cmd:
      - "_{use-node()}"
      - "npm run build"
      - "npm prune --production"

  test:
    description: "Run tests"
    cmd:
      - "_{use-node()}"
      - "npm test -- @{args}"

  docker-build:
    description: "Build Docker image"
    cmd:
      - "docker build -t user-service:latest ."
      - "docker tag user-service:latest #{global.vars.docker_registry}/user-service:latest"

  deploy:
    description: "Deploy to k8s"
    cmd:
      - "_{check-env(env=prod)}"
      - "kubectl apply -f k8s/"
      - "_{global.slack-notify(message='user-service deployed')}"

docker:
  compose_file: docker-compose.yml
  groups:
    db: [postgres, redis]
    infra: [postgres, redis, rabbitmq]
```

Использование:
```bash
# Разработка
pm user-service :install
pm user-service :up @db
pm user-service :dev

# Тестирование
pm user-service :test --coverage
pm user-service :test --watch

# Деплой
pm user-service :build
pm user-service :docker-build
pm user-service :deploy
```

### Сценарий 2: Monorepo с Frontend и Backend

```yaml
info:
  name: webapp
  root: ~/repos/webapp

func:
  use-node:
    params:
      version:
        default: "20"
    script: "nvm use @{version}"

  workspace-cmd:
    params:
      workspace:
        required: true
      cmd:
        required: true
    script: "npm run @{cmd} --workspace=@{workspace}"

commands:
  install:
    description: "Install all dependencies"
    cmd:
      - "_{use-node()}"
      - "npm install"

  dev-frontend:
    description: "Start frontend dev server"
    cmd:
      - "_{use-node()}"
      - "_{workspace-cmd(workspace=frontend, cmd=dev)}"

  dev-backend:
    description: "Start backend dev server"
    cmd:
      - "_{use-node()}"
      - "_{workspace-cmd(workspace=backend, cmd=dev)}"

  build:
    description: "Build all packages"
    cmd:
      - "_{use-node()}"
      - "npm run build --workspaces"

  test:
    description: "Run all tests"
    cmd:
      - "_{use-node()}"
      - "npm test --workspaces -- @{args}"

  lint:
    description: "Lint all packages"
    cmd:
      - "_{use-node()}"
      - "npm run lint --workspaces"

  clean:
    description: "Clean build artifacts"
    cmd:
      - "rm -rf packages/*/dist"
      - "rm -rf packages/*/node_modules"
      - "rm -rf node_modules"

docker:
  compose_file: docker-compose.yml
  groups:
    db: [postgres, mongodb]
    cache: [redis, memcached]
    queue: [rabbitmq]
    infra: [postgres, mongodb, redis, rabbitmq]
```

Использование:
```bash
# Разработка
pm webapp :install
pm webapp :up @infra
pm webapp :dev-frontend  # В одном терминале
pm webapp :dev-backend   # В другом терминале

# CI/CD
pm webapp :install
pm webapp :lint
pm webapp :test
pm webapp :build
```

### Сценарий 3: Java проект с Gradle

```yaml
info:
  name: backend-api
  root: ~/repos/backend-api

func:
  use-java:
    params:
      version:
        default: "21"
    script:
      - "sdk use java @{version}"
      - "java -version"

  gradle:
    params:
      tasks:
        required: true
      args:
        default: ""
    script: "./gradlew @{tasks} @{args}"

commands:
  clean:
    description: "Clean build"
    cmd: "_{gradle(tasks=clean)}"

  build:
    description: "Build project"
    cmd:
      - "_{use-java()}"
      - "_{gradle(tasks=build, args=-x test)}"

  test:
    description: "Run tests"
    cmd:
      - "_{use-java()}"
      - "_{gradle(tasks=test, args=@{args})}"

  run:
    description: "Run application"
    cmd:
      - "_{use-java()}"
      - "_{gradle(tasks=bootRun)}"

  docker-build:
    description: "Build Docker image"
    cmd:
      - "_{use-java()}"
      - "_{gradle(tasks=bootBuildImage)}"

  deploy:
    description: "Deploy to k8s"
    cmd:
      - "kubectl apply -f k8s/deployment.yaml"
      - "kubectl rollout status deployment/backend-api"

docker:
  compose_file: docker-compose.yml
  groups:
    db: [postgres]
    cache: [redis]
    infra: [postgres, redis, kafka, zookeeper]
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

# CI/CD
pm backend-api :clean
pm backend-api :build
pm backend-api :test
pm backend-api :docker-build
pm backend-api :deploy
```

### Сценарий 4: Множественные команды в одном вызове

```bash
# Полный цикл: очистка → установка → сборка → тест
pm myapp :clean :install :build :test

# Деплой: сборка → docker build → push → deploy
pm myapp :build :docker-build :docker-push :deploy

# Локальная разработка: очистка → установка → запуск БД → dev сервер
pm myapp :clean :install :up @db :dev
```

### Сценарий 5: Использование raw команд

```bash
# Выполнить произвольные команды в директории проекта
pm myapp git status
pm myapp ls -la
pm myapp docker compose ps
pm myapp kubectl get pods

# Комбинирование с именованными командами невозможно!
# Либо :команды, либо raw режим
```

## Советы и best practices

### 1. Используйте функции для переиспользования

❌ Плохо:
```yaml
commands:
  build-dev:
    cmd:
      - "sdk use java 21"
      - "./gradlew build"

  build-prod:
    cmd:
      - "sdk use java 21"
      - "./gradlew build -Pprod"
```

✅ Хорошо:
```yaml
func:
  use-java:
    params:
      version:
        default: "21"
    script: "sdk use java @{version}"

commands:
  build-dev:
    cmd:
      - "_{use-java()}"
      - "./gradlew build"

  build-prod:
    cmd:
      - "_{use-java()}"
      - "./gradlew build -Pprod"
```

### 2. Используйте описания

```yaml
commands:
  build:
    description: "Build project with all optimizations"
    cmd: "./gradlew build"

  test:
    description: "Run unit and integration tests"
    cmd: "./gradlew test integrationTest"
```

### 3. Группируйте Docker сервисы логически

```yaml
docker:
  groups:
    # Минимум для работы
    minimal: [postgres]

    # Локальная разработка
    dev: [postgres, redis]

    # Полное окружение
    full: [postgres, redis, rabbitmq, elasticsearch]

    # Только мониторинг
    monitoring: [prometheus, grafana]
```

### 4. Используйте глобальный конфиг для общих функций

`~/.config/pm/global.yml`:
```yaml
func:
  git-check-clean:
    script: |
      git diff --quiet || (echo "Git working directory is not clean" && exit 1)

  require-env:
    params:
      var:
        required: true
    script: |
      test -n "${@{var}}" || (echo "Required env var @{var} is not set" && exit 1)
```

Использование в проектах:
```yaml
commands:
  deploy:
    cmd:
      - "_{global.git-check-clean()}"
      - "_{global.require-env(var=API_TOKEN)}"
      - "./deploy.sh"
```

### 5. Параметры @{args} для гибкости

```yaml
commands:
  test:
    description: "Run tests with custom args"
    cmd: "npm test -- @{args}"

  docker-logs:
    description: "Show docker logs"
    cmd: "docker compose logs @{args}"
```

```bash
pm myapp :test --watch --coverage
pm myapp :docker-logs -f --tail=100 api
```
