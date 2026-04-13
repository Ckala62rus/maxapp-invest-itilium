# План по MAX identity: просто и по шагам

## В чем проблема сейчас

Сейчас проект живет на временной схеме:

- frontend передает `userId`
- backend принимает `X-User-ID` или `userId`
- этот id потом уходит дальше в ITILIUM

Это удобно для разработки, но это не настоящая MAX-аутентификация.

## Как должно быть правильно

Правильная схема такая:

1. во frontend подключается реальная JS library MAX
2. Mini App получает init data / token от MAX
3. frontend отправляет эти данные на backend
4. backend проверяет и расшифровывает payload
5. backend получает реальный MAX user id
6. именно этот MAX user id используется дальше в проекте
7. именно этот MAX user id уходит в ITILIUM вместо Telegram id

## Что у нас уже готово

- frontend и backend уже умеют работать с единым внешним `userId`
- profile / registration / tickets flow уже построены
- архитектура frontend уже приведена в порядок:
  - экраны вынесены
  - `ticket` flow вынесен в composable
  - `auth` flow вынесен в composable

То есть основа уже есть. Не хватает именно настоящего identity-слоя.

## Что еще не сделано

- подключение MAX JS library в HTML
- получение init data / token на фронте
- backend endpoint для валидации MAX payload
- логика validate/decrypt на backend
- trusted identity в middleware/context
- отказ от временного `X-User-ID` / `userId`

## План работ

### Шаг 1. Подключить MAX JS library во frontend

Что сделать:

- добавить библиотеку MAX в HTML entrypoint
- получить из нее init data / token
- сделать маленький frontend adapter/composable под это

Результат:

- frontend реально видит данные пользователя MAX

### Шаг 2. Добавить backend endpoint валидации

Что сделать:

- добавить endpoint вроде `POST /api/v1/auth/max/validate`
- принять init data / token
- проверить подпись или расшифровать payload
- вернуть нормализованные данные пользователя

Результат:

- backend умеет подтвердить, что пользователь настоящий

### Шаг 3. Перевести backend на trusted identity

Что сделать:

- положить MAX user id в request context
- использовать его в profile и ticket handlers
- перестать доверять клиентскому `userId` как основному источнику

Результат:

- все сценарии работают от доверенного пользователя

### Шаг 4. Передавать MAX user id в ITILIUM

Что сделать:

- вместо Telegram id использовать MAX user id
- оставить это в одном backend adapter слое, чтобы потом легко менять mapping

Результат:

- ITILIUM работает уже с identity из MAX

### Шаг 5. Убрать временные костыли

Что сделать:

- убрать или жестко ограничить `X-User-ID` / `userId` fallback
- если нужен debug-режим, оставить его только под явным флагом

Результат:

- test/prod режимы становятся безопаснее и ближе к боевому поведению

## Что делаем следующим

Самый правильный следующий шаг:

- внедрить backend endpoint `POST /api/v1/auth/max/validate`

Почему именно он:

- без него нельзя построить нормальную доверенную identity-схему
- после него станет понятен точный контракт для frontend MAX JS integration
- после него можно уже честно переводить весь backend с временного `userId` на реальный MAX user id

## Что проверять после внедрения

После реализации MAX identity нужно проверить:

1. MAX library реально загружается
2. frontend получает init data
3. backend валидирует / расшифровывает payload
4. в логах видно правильный MAX user id
5. профиль грузится для реального пользователя
6. этот же id уходит дальше в ITILIUM
7. подмена `X-User-ID` больше не влияет на пользователя в защищенных сценариях
