# Развертывание в Production

Эта инструкция описывает полное развертывание на чистом сервере Ubuntu: подготовку каталога под проект, установку Docker, скачивание проекта через Git, установку выданных администратором TLS-сертификатов, настройку `.env`, запуск контейнеров и базовые smoke-проверки.

Production-стек отдает Vue-фронтенд через nginx по HTTPS и проксирует `/api/*` на Go-бэкенд внутри Docker-сети. Бэкенд не публикуется на хосте. Redis доступен только внутри стека. Grafana, Prometheus и Loki привязаны только к `127.0.0.1`.

Postgres сейчас намеренно не запускается в production. Текущий бэкенд использует `MemoryUserRepository` для временных снимков профиля, Redis для кеша и ITILIUM как источник достоверных данных. Каталог `migrations/` и dev-профиль миграций остаются заготовкой для будущего постоянного репозитория.

## 1. Необходимые исходные данные и место развертывания

Перед началом подготовьте следующие значения:

- Сервер: Ubuntu 22.04/24.04 с доступом root или sudo.
- Домен: DNS-запись `A` указывает на публичный IP сервера.
- Файрвол или security group у провайдера разрешает входящие `80/tcp` и `443/tcp`.
- Доступ к репозиторию на GitHub.
- TLS-сертификат от администратора для production-домена: файл полной цепочки сертификатов и файл приватного ключа.
- Production URL, логин и пароль ITILIUM.
- Токен MAX-бота.
- Случайный секрет сессий бэкенда для `AUTH_ACCESS_TOKEN_SECRET`.

Админ сообщил, что для Docker создан общий каталог `/opt/docker-shared/`, чтобы контейнеры можно было запускать из-под любого пользователя с правами sudo и не завязываться на `/home`. Поэтому, если админ не укажет другой точный путь, разворачивайте проект здесь:

```bash
/opt/docker-shared/maxapp-invest-itilium
```

Дальше в инструкции используется именно этот путь. Если позже админ даст другой путь внутри `/opt/docker-shared/`, замените путь во всех командах.

Важно:

- Не разворачивайте проект в `/home/<user>/...`, чтобы запуск не зависел от домашнего каталога конкретного пользователя.
- Команды Docker ниже выполняются через `sudo docker ...`. Так ими сможет пользоваться любой администратор с sudo-доступом без отдельной настройки Docker-группы.
- Реальные пароли, токены и `.env` нельзя отправлять в чат, коммитить или вставлять в документацию.

## 2. Подключение к серверу

Подключитесь по SSH:

```bash
ssh user@server-ip
```

Проверьте, что вы на нужном сервере:

```bash
hostname
whoami
pwd
```

Обновите базовые пакеты и установите только то, что нужно для Docker-репозитория и скачивания проекта:

```bash
sudo apt-get update
sudo apt-get upgrade -y
sudo apt-get install -y ca-certificates curl gnupg git
```

При необходимости настройте часовой пояс сервера:

```bash
sudo timedatectl set-timezone Europe/Moscow
timedatectl
```

## 3. Настройка файрвола и проверка DNS

Если на сервере используется UFW, установите его и разрешите SSH, HTTP и HTTPS. Сделайте это до включения UFW:

```bash
sudo apt-get install -y ufw
sudo ufw allow OpenSSH
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
sudo ufw status verbose
```

Проверьте, что домен уже смотрит на сервер. В выводе должен быть публичный IP этого сервера:

```bash
getent ahostsv4 your.domain.example
curl -4 ifconfig.me
```

Замените `your.domain.example` на реальный домен, который потом будет указан в `.env` как `CERT_DOMAIN`.

Не открывайте в интернет бэкенд `3000`, Redis `6379`, Prometheus `9090`, Loki `3100` или Grafana `3001`. В `docker-compose.yml` порты observability привязаны к `127.0.0.1`.

## 4. Установка Docker Engine и Compose

Удалите старые пакеты Docker, если они есть:

```bash
for pkg in docker.io docker-doc docker-compose podman-docker containerd runc; do
  sudo apt-get remove -y "$pkg" || true
done
```

Добавьте официальный репозиторий Docker:

```bash
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

. /etc/os-release
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  ${VERSION_CODENAME} stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
```

Установите Docker:

```bash
sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
```

Включите Docker и проверьте установку:

```bash
sudo systemctl enable --now docker
sudo docker version
sudo docker compose version
sudo docker run --rm hello-world
```

В этой инструкции дальше используется `sudo docker ...`, поэтому добавлять пользователя в группу `docker` не обязательно.

Опционально, если админ разрешает запуск Docker без `sudo`, можно добавить своего пользователя в группу `docker`:

```bash
sudo usermod -aG docker "$USER"
newgrp docker
docker ps
```

Если `newgrp docker` недостаточно, переподключитесь по SSH. Даже после этого production-команды из этой инструкции можно продолжать выполнять через `sudo docker`.

## 5. Установка Docker logging driver для Loki

Сервис бэкенда использует Docker logging driver `loki`, поэтому установите плагин до запуска production-стека. Для обычного Ubuntu-сервера `x86_64` используйте тег `amd64`:

```bash
sudo docker plugin install grafana/loki-docker-driver:3.7.2-amd64 --alias loki --grant-all-permissions
sudo docker plugin ls
```

Если `uname -m` показывает `aarch64` или `arm64`, используйте тег `3.7.2-arm64`.

Если плагин уже существует, убедитесь, что он включен:

```bash
sudo docker plugin enable loki || true
```

## 6. Подготовка TLS-сертификатов

Certbot и Let's Encrypt в этом развертывании не используются: сертификаты выдает администратор.

Уточните у администратора два файла:

- полная цепочка сертификатов для домена, обычно `fullchain.pem`, `certificate.crt` или похожее имя;
- приватный ключ сертификата, обычно `privkey.pem`, `private.key` или похожее имя.

В `.env` значение `CERT_DOMAIN` должно быть только доменом без схемы и слэша:

```env
CERT_DOMAIN=maxbot.fpkinvest.ru
```

Нельзя писать так:

```env
CERT_DOMAIN=https://maxbot.fpkinvest.ru/
```

Nginx внутри контейнера ожидает файлы по пути `/etc/nginx/certs/fullchain.pem` и `/etc/nginx/certs/privkey.pem`. В `docker-compose.yml` туда монтируется серверный каталог проекта `./ssl`.

Создайте каталог под сертификаты рядом с проектом:

```bash
cd /opt/docker-shared/maxapp-invest-itilium
sudo mkdir -p ssl
```

Скопируйте выданные администратором файлы в этот каталог. Для текущего сервера они уже лежат здесь:

```bash
ls -lah /opt/docker-shared/maxapp-invest-itilium/ssl
```

Минимально нужны:

- `fullchain.pem`
- `privkey.pem`

Выставьте владельца и права:

```bash
sudo chown root:root /opt/docker-shared/maxapp-invest-itilium/ssl/fullchain.pem
sudo chown root:root /opt/docker-shared/maxapp-invest-itilium/ssl/privkey.pem
sudo chmod 644 /opt/docker-shared/maxapp-invest-itilium/ssl/fullchain.pem
sudo chmod 600 /opt/docker-shared/maxapp-invest-itilium/ssl/privkey.pem
```

Проверьте сертификат и ключ:

```bash
sudo openssl x509 -in /opt/docker-shared/maxapp-invest-itilium/ssl/fullchain.pem -noout -subject -issuer -dates
sudo openssl rsa -in /opt/docker-shared/maxapp-invest-itilium/ssl/privkey.pem -check -noout
```

Если `openssl` не установлен, выполните `sudo apt-get install -y openssl`.

## 7. Подготовка каталога и скачивание проекта через Git

Убедитесь, что общий каталог Docker существует. Если админ уже создал `/opt/docker-shared/`, команда просто проверит и не сломает его:

```bash
sudo mkdir -p /opt/docker-shared
sudo chmod 755 /opt/docker-shared
ls -ld /opt/docker-shared
```

Создайте каталог приложения:

```bash
sudo mkdir -p /opt/docker-shared/maxapp-invest-itilium
sudo chown root:root /opt/docker-shared/maxapp-invest-itilium
sudo chmod 755 /opt/docker-shared/maxapp-invest-itilium
```

Если каталог пустой, склонируйте проект из GitHub. Команда скачает весь проект прямо в `/opt/docker-shared/maxapp-invest-itilium`:

```bash
sudo git clone https://github.com/Ckala62rus/maxapp-invest-itilium.git /opt/docker-shared/maxapp-invest-itilium
```

Перейдите в каталог проекта и проверьте, что код скачался:

```bash
cd /opt/docker-shared/maxapp-invest-itilium
pwd
sudo git status
sudo git branch --show-current
ls -la
```

Ожидается, что внутри будут файлы `docker-compose.yml`, `.env.example`, `Dockerfile`, каталоги `frontend/`, `internal/`, `deploy/`.

Если репозиторий уже был склонирован раньше, не клонируйте его повторно. Просто обновите код:

```bash
cd /opt/docker-shared/maxapp-invest-itilium
sudo git pull --ff-only
```

Если `git pull --ff-only` сообщает о локальных изменениях, остановитесь и сначала разберитесь, кто и зачем менял файлы на сервере:

```bash
sudo git status
```

## 8. Создание production `.env`

Скопируйте пример. Если `.env` уже существует, не затирайте его без бэкапа:

```bash
if [ ! -f .env ]; then
  sudo cp .env.example .env
fi
sudo vi .env
```

Заполните значения по такому шаблону:

```bash
WEB_HTTP_PORT=80
WEB_HTTPS_PORT=443
CERT_DOMAIN=your.domain.example

GF_ADMIN_USER=admin
GF_ADMIN_PASSWORD=replace-with-strong-password
GRAFANA_HTTP_PORT=3001
PROMETHEUS_HTTP_PORT=9090
LOKI_HTTP_PORT=3100

LOKI_URL=http://127.0.0.1:3100/loki/api/v1/push

ITILIUM_BASE_URL=https://itilium.example/itilium/hs/Max
ITILIUM_HOST=https://itilium.example
ITILIUM_LOGIN=replace-me
ITILIUM_PASSWORD=replace-me
ITILIUM_INSECURE_SKIP_VERIFY=false

MAX_BOT_TOKEN=replace-me
AUTH_ACCESS_TOKEN_SECRET=replace-with-long-random-secret
AUTH_ACCESS_TOKEN_TTL=12h
AUTH_MAX_INIT_DATA_TTL=10m
AUTH_ALLOW_DEBUG_IDENTITY_HEADERS=false

LOG_LEVEL=info
```

Сгенерируйте надежный `AUTH_ACCESS_TOKEN_SECRET`. Если `openssl` не установлен, сначала выполните `sudo apt-get install -y openssl`:

```bash
openssl rand -hex 32
```

Ограничьте доступ к файлу:

```bash
sudo chown root:root .env
sudo chmod 600 .env
sudo ls -la .env
```

Никогда не коммитьте `.env` и не вставляйте реальные учетные данные в документацию.

## 9. Проверка Compose-файла

Проверьте, что compose может разрешить переменные и YAML:

```bash
sudo docker compose --env-file .env config > /tmp/maxapp-compose.yml
```

Если проверка завершилась ошибкой, исправьте `.env` или compose-файлы перед продолжением.

## 10. Проверка файлов сертификата перед запуском

Убедитесь, что файлы лежат в каталоге `ssl` рядом с `docker-compose.yml`:

```bash
sudo ls -la /opt/docker-shared/maxapp-invest-itilium/ssl/
```

В каталоге должны быть минимум два файла:

- `fullchain.pem`
- `privkey.pem`

## 11. Запуск production-стека

Соберите и запустите все production-сервисы:

```bash
sudo docker compose --env-file .env up --build -d
sudo docker compose --env-file .env ps
```

Ожидаемые сервисы:

- `maxapp-itilium-web`
- `maxapp-itilium-backend`
- `maxapp-itilium-redis`
- `maxapp-itilium-loki`
- `maxapp-itilium-prometheus`
- `maxapp-itilium-grafana`

Посмотрите логи первого запуска:

```bash
sudo docker compose --env-file .env logs -f web backend
```

## 12. Smoke-проверки

Проверьте HTTPS:

```bash
curl -I https://your.domain.example
```

Проверьте health endpoints через nginx:

```bash
curl -fsS https://your.domain.example/healthz
curl -fsS https://your.domain.example/readyz
```

Проверьте, что бэкенд не открыт напрямую:

```bash
curl -I http://127.0.0.1:3000/healthz
```

Этот прямой запрос должен завершиться ошибкой, если вы намеренно не меняли порты в compose.

Проверьте статус контейнеров:

```bash
sudo docker compose --env-file .env ps
sudo docker compose --env-file .env logs --tail=100 backend
```

Проверьте локальные порты observability с сервера:

```bash
curl -I http://127.0.0.1:3001
curl -I http://127.0.0.1:9090
curl -I http://127.0.0.1:3100/ready
```

Grafana намеренно доступна только с сервера. Используйте SSH port forwarding с рабочей станции:

```bash
ssh -L 3001:127.0.0.1:3001 user@server-ip
```

Затем откройте `http://127.0.0.1:3001` локально.

## 13. Подключение MAX Mini App

В настройках MAX-бота укажите URL mini app:

```text
https://your.domain.example/
```

Затем проверьте реальный сценарий из MAX:

1. Откройте mini app из MAX, а не напрямую из браузера.
2. Проверьте авторизацию: в логах бэкенда должен появиться `POST /api/v1/auth/max/validate`.
3. Проверьте профиль: `GET /api/v1/users/me`.
4. Проверьте один сценарий ITILIUM: список заявок, открытие деталей заявки или создание тестовой заявки.
5. Если что-то не работает, фильтруйте логи по `request_id` из ответа.

Полезные логи:

```bash
sudo docker compose --env-file .env logs -f backend
sudo docker compose --env-file .env logs -f web
```

### 13.1. Проверка кнопки «Открыть заявку» (deep link) на production

Сценарий: после деплоя кода с поддержкой `start_param` отправить себе личное сообщение из Postman с кнопкой `open_app` и убедиться, что mini app открывается сразу на карточке заявки.

#### Шаг 1. Выгрузить обновлённый код на сервер

```bash
cd /opt/docker-shared/maxapp-invest-itilium   # или ваш каталог на сервере
sudo git pull --ff-only
sudo docker compose --env-file .env up --build -d web backend
sudo docker compose --env-file .env ps
```

Нужны контейнеры **`web`** (frontend + nginx) и **`backend`**. Без пересборки `web` на prod останется старый JS без deep link.

Проверка, что prod отвечает:

```bash
curl -sS -o /dev/null -w "%{http_code}" https://your.domain.example/
curl -sS -o /dev/null -w "%{http_code}" https://your.domain.example/healthz
```

Оба запроса — **200** (или `healthz` по вашему маршруту).

#### Шаг 2. URL mini app в настройках MAX-бота

В кабинете MAX у бота URL mini app = **тот же HTTPS**, что и production, например:

```text
https://your.domain.example/
```

Если URL указывает на старый стенд, localhost или tuna — кнопка `open_app` откроет **не тот** frontend.

#### Шаг 3. Postman — параметры

Коллекция: `tools/max-notify/max-notify.postman_collection.json`.

| Переменная | Значение |
|------------|----------|
| `maxBotToken` | токен бота (как в `.env` на сервере) |
| `maxNotifyUserId` | **ваш** MAX user id (получатель лички) |
| `maxBotContactId` | `user_id` бота из **GET /me** |
| `ticketNumber` | номер заявки, напр. `0000024311` |

1. **GET /me** — скопируйте `user_id` → `maxBotContactId`, `username` → `maxBotWebApp` (если нужен).
2. Убедитесь, что вы **хотя бы раз** открывали чат с ботом в MAX (иначе личка не доставится).

#### Шаг 4. Отправить сообщение с кнопкой

Запрос **POST /messages — open_app (contact_id)** (надёжнее, чем `web_app` с чужим username).

Тело (подставьте свои значения):

```json
{
  "text": "Тест: открыть заявку 0000024311",
  "notify": true,
  "attachments": [
    {
      "type": "inline_keyboard",
      "payload": {
        "buttons": [
          [
            {
              "type": "open_app",
              "text": "Открыть 0000024311",
              "contact_id": 270966433,
              "payload": "ticket_0000024311"
            }
          ]
        ]
      }
    }
  ]
}
```

- **`contact_id`** — из GET /me (не из примеров README).
- **`payload`** — строго `ticket_{номер}`; двоеточие MAX не принимает.
- Заявка должна **существовать в ITILIUM** и быть доступна **вашему** пользователю.

Успех Postman: **HTTP 200**, в теле объект `message`.

Альтернатива с сервера:

```bash
cd tools/max-notify
python notify.py --template assigned --ticket 0000024311
```

#### Шаг 5. Нажать кнопку в MAX

1. Откройте **MAX** (телефон или desktop).
2. Личка с ботом → новое сообщение → **«Открыть …»**.
3. Должно открыться **mini app на production** (не браузер с localhost).
4. После «Проверяем MAX-сессию…» — экран **«Карточка заявки»** с номером `0000024311`, не главная.

#### Шаг 6. Что смотреть, если что-то не так

| Симптом | Вероятная причина |
|---------|-------------------|
| Postman **404** `Link not found ... invest_it_bot` | Неверный `web_app`; используйте **contact_id** из GET /me |
| Сообщение не приходит | Не открывали чат с ботом; неверный `maxNotifyUserId` |
| Открылась **главная**, не заявка | На prod старый frontend; deep link не задеployен |
| Открылась главная, auth OK | Старый кэш WebView — закройте mini app и откройте снова с кнопки |
| Карточка с ошибкой | Неверный номер или заявка недоступна вашему user id в 1С |
| Mini app не открывается | URL mini app в настройках бота не совпадает с prod |

Логи backend при успехе (фильтр по времени нажатия кнопки):

```text
POST /api/v1/auth/max/validate
GET  /api/v1/users/me
GET  /api/v1/tickets/0000024311
```

Локальная отладка без MAX: `documentation/local_development.md` → раздел «Проверка deep link».

## 14. Замена сертификатов

Когда администратор выдаст новые файлы сертификата, замените `fullchain.pem` и `privkey.pem` в `/opt/docker-shared/maxapp-invest-itilium/ssl/`.

После замены проверьте сертификат:

```bash
sudo openssl x509 -in /opt/docker-shared/maxapp-invest-itilium/ssl/fullchain.pem -noout -subject -issuer -dates
```

Затем перезагрузите nginx внутри контейнера:

```bash
cd /opt/docker-shared/maxapp-invest-itilium
sudo docker compose --env-file .env exec web nginx -s reload
```

## 15. Обновление приложения

Разверните новую версию:

```bash
cd /opt/docker-shared/maxapp-invest-itilium
sudo git pull --ff-only
sudo docker compose --env-file .env up --build -d
sudo docker compose --env-file .env ps
```

Периодически очищайте неиспользуемый build cache и образы:

```bash
sudo docker system prune -f
```

## 16. Заметки по резервному копированию

Сейчас в production нет базы данных Postgres. Сохраняйте эти ресурсы на уровне хоста:

- `/opt/docker-shared/maxapp-invest-itilium/.env`
- `/opt/docker-shared/maxapp-invest-itilium/ssl/fullchain.pem`
- `/opt/docker-shared/maxapp-invest-itilium/ssl/privkey.pem`
- Том Grafana, если дашборды настраивались вручную:

```bash
sudo docker volume ls | grep grafana
```

Если позже будут добавлены постоянные репозитории, настройте резервное копирование базы данных до включения production-записей.

## 17. Устранение неполадок

Если HTTPS nginx не запускается:

- Проверьте файлы сертификата: `sudo ls -la /opt/docker-shared/maxapp-invest-itilium/ssl/`.
- Проверьте, что `CERT_DOMAIN` в `.env` содержит только домен.
- Проверьте, что в `CERT_DOMAIN` нет `https://` и завершающего `/`.
- Проверьте сертификат: `sudo openssl x509 -in /opt/docker-shared/maxapp-invest-itilium/ssl/fullchain.pem -noout -subject -issuer -dates`.
- Прочитайте логи nginx: `sudo docker compose --env-file .env logs web`.

Если бэкенд завершается:

- Проверьте обязательные значения окружения: `ITILIUM_BASE_URL`, `ITILIUM_LOGIN`, `ITILIUM_PASSWORD`, `MAX_BOT_TOKEN`, `AUTH_ACCESS_TOKEN_SECRET`.
- Прочитайте логи: `sudo docker compose --env-file .env logs --tail=200 backend`.

Если нужно включить подробные debug-логи, установите в `.env`:

```env
LOG_LEVEL=debug
```

Затем пересоздайте backend:

```bash
sudo docker compose --env-file .env up -d --force-recreate --no-deps backend
sudo docker compose --env-file .env logs --tail=50 backend
```

Если Docker сообщает о неизвестном logging driver `loki`:

```bash
sudo docker plugin install grafana/loki-docker-driver:3.7.2-amd64 --alias loki --grant-all-permissions
sudo docker plugin enable loki
sudo docker compose --env-file .env up -d
```

Если вызовы ITILIUM завершаются ошибкой:

- Проверьте, что сервер может достучаться до ITILIUM по URL из `.env`: `curl -vk "https://itilium.example/itilium/hs/Max"`.
- Проверьте, не использует ли production ITILIUM частную сеть или VPN-маршрут.
- Оставляйте `ITILIUM_INSECURE_SKIP_VERIFY=false` для корректных HTTPS-сертификатов; используйте `true` только как временное исключение для IP/self-signed тестовых endpoints.
