# Руководство по CI/CD

Документ описывает запуск GitLab CI/CD для проекта hiTech, настройку self-hosted раннеров и подготовку серверов развертывания.

## 1. Структура пайплайна

Файл `.gitlab-ci.yml` определяет четыре стадии:

| Стадия | Джобы | Назначение |
| --- | --- | --- |
| `build` | `build:go-orchestrator`, `build:drone-service`, `build:admin-panel` | Сборка Docker-образов и загрузка в GitLab Container Registry. На ветке `main` также обновляется тег `:latest`. |
| `test` | `test:go`, `test:drone-service`, `test:admin-panel` | Запуск unit-тестов (Go), pytest для Python и линт для admin-панели. |
| `deploy` | `deploy:dev` (ветка `dev`), `deploy:prod` (ветка `main`) | Подключение по SSH к серверам, `docker compose pull` и `docker compose up -d --remove-orphans` с файлом `deployment/docker/docker-compose.cicd.yml`. |
| `cleanup` | `docker:gc` | Очистка кэша Docker на раннере (опционально). |

### Флаги запуска
- `FORCE_BUILD_ALL=true` — принудительно пересобрать все сервисы.
- `FORCE_DEPLOY=true` — выполнить деплой даже при отсутствии изменений в отслеживаемых файлах.

Флаги можно задать в переменных пайплайна (`Run pipeline → Variables`) или в настройках CI/CD.

## 2. Переменные GitLab

Задайте переменные в `Settings → CI/CD → Variables`. Отметьте секреты как **protected**/**masked**.

### Registry и авторизация
| Переменная | Пример | Комментарий |
| --- | --- | --- |
| `CI_REGISTRY` | `registry.gitlab.com` | По умолчанию задаётся GitLab. |
| `CI_REGISTRY_USER` | `gitlab-ci-token` | Системная, используется в скриптах. |
| `CI_REGISTRY_PASSWORD` | `${CI_JOB_TOKEN}` | Автоматически доступна. |

### SSH и деплой-серверы
| Переменная | Пример | Описание |
| --- | --- | --- |
| `SSH_PRIVATE_KEY` | (текст ключа) | Приватный ключ, добавленный в `~/.ssh/authorized_keys` на серверах. |
| `DEV_SERVER_IP` | `192.168.10.20` | IP dev-сервера. |
| `DEV_SERVER_USER` | `deploy` | SSH-пользователь dev-сервера. |
| `DEV_DEPLOY_PATH` | `/opt/hitech-dev` | (опционально) путь деплоя, по умолчанию `/opt/hitech-dev`. |
| `DEV_ENV_FILE` | `.env.dev` | (опционально) имя env-файла, по умолчанию `.env.dev` (fallback `.env.prod`). |
| `PROD_SERVER_IP` | `203.0.113.10` | Прод-сервер. |
| `PROD_SERVER_USER` | `deploy` | SSH-пользователь прод-сервера. |
| `PROD_DEPLOY_PATH` | `/opt/hitech` | (опционально) путь деплоя, по умолчанию `/opt/hitech`. |
| `PROD_ENV_FILE` | `.env.prod` | (опционально) env-файл для prod. |

### Секреты приложения
Все значения из `.env.prod`/`.env.dev` должны быть заданы на серверах (JWT, Postgres, MinIO, RabbitMQ, SMSAero, Firebase и т.д.). Не коммитьте секреты в репозиторий.

## 3. Подготовка серверов

1. **Установка Docker и Compose**
   ```bash
   sudo apt-get update
   sudo apt-get install -y docker.io docker-compose-plugin
   sudo usermod -aG docker $USER
   ```
   Перезайдите в систему, чтобы применились права.

2. **Клонирование репозитория** в каталог деплоя (например `/opt/hitech`). Пайплайн будет выполнять `git reset --hard`, поэтому не храните в каталоге локальные изменения.

3. **Env-файлы**: создайте `.env.prod`, `.env.dev`, заполните все требуемые переменные.

4. **Firebase**: поместите JSON сервисного аккаунта по пути, указанному в `FIREBASE_CREDENTIALS_PATH` (по умолчанию `./secrets/firebase-service-account.json`). Настройте права доступа (`chmod 600`).

5. **Firewall**: откройте необходимые порты (HTTP/HTTPS, MinIO/Grafana/RabbitMQ при необходимости).

## 4. Установка GitLab Runner

На каждом сервере (dev/prod) зарегистрируйте runner для деплоя.

```bash
curl -L https://packages.gitlab.com/install/repositories/runner/gitlab-runner/script.deb.sh | sudo bash
sudo apt-get install gitlab-runner
sudo gitlab-runner register
```

При регистрации:
- URL: `https://gitlab.com/` (или адрес вашей инсталляции).
- Token: возьмите из `Settings → CI/CD → Runners → Register a runner`.
- Описание: `hitech-dev`, `hitech-prod`.
- Tags: `dev` или `prod` (используются в job'ах).
- Executor: `shell` (рекомендуется, Docker CLI уже на хосте).

Запуск и проверка статуса:
```bash
sudo gitlab-runner start
sudo gitlab-runner verify
```

Убедитесь, что runner виден в проекте и помечен нужными тегами.

## 5. Запуск пайплайнов

- **Автоматически:** пуши в `dev`/`main` запускают build/test и, при изменениях, соответствующий deploy.
- **Ручной запуск:** `CI/CD → Pipelines → Run pipeline`, указать ветку и, при необходимости, флаги `FORCE_*`.
- **Наблюдение:** проверяйте логи деплоя (`docker compose ... ps` в конце). При ошибках используйте раздел «Troubleshooting» ниже.

## 6. Частые проблемы

| Симптом | Возможная причина | Решение |
| --- | --- | --- |
| Ошибка SSH | Неверные IP/пользователь или ключ не добавлен | Обновить переменные, добавить ключ в `authorized_keys`. |
| Не найден `.env` | Файл отсутствует на сервере или неверный `ENV_FILE` | Скопировать env-файл, проверить путь. |
| Образ не найден | Сборка упала или теги не совпадают | Запустить пайплайн с `FORCE_BUILD_ALL=true`. |
| Нет доступа к Firebase | Некорректный путь или права на файл | Проверить `FIREBASE_CREDENTIALS_PATH` и права. |
| Runner offline | Сервис остановлен / неверный токен | `sudo gitlab-runner start`, `gitlab-runner verify`, при необходимости повторно зарегистрировать. |

## 7. Ручной откат

1. Подключиться к серверу.
2. `cd /opt/hitech` (или указанный `DEPLOY_PATH`).
3. `git checkout <commit>` + задать `IMAGE_TAG` на нужный SHA.
4. `docker compose --env-file <env> -f deployment/docker/docker-compose.cicd.yml pull <services>`.
5. `docker compose --env-file <env> -f ... up -d --remove-orphans`.

## 8. Полезные ссылки
- `docs/ru/РАЗВЕРТЫВАНИЕ.md` — детальный деплой вручную.
- `docs/ru/МОНИТОРИНГ И ЛОГИРОВАНИЕ.md` — подключение Prometheus/Grafana/Loki.
- `deployment/docker/docker-compose.*.yml` — конфигурации сервисов для разных окружений.
