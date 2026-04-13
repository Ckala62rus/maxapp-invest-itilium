# План интеграции identity из MAX

## Зачем нужен этот план

Сейчас проект все еще использует временную схему identity:

- backend читает `X-User-ID` или `userId`
- frontend передает `userId` в запросах
- этот же `userId` потом уходит дальше в ITILIUM

Это допустимо только как временный migration stub.

Чтобы действительно понимать, что клиентом является реальный пользователь MAX, Mini App должен:

- подключать реальную JavaScript-библиотеку MAX в HTML
- получать signed/encrypted init data из среды MAX
- отправлять init data на backend
- валидировать и расшифровывать payload на backend с использованием токена/секрета приложения
- извлекать реальный MAX user identifier
- использовать этот идентификатор как acting user id для profile flow и ITILIUM

## Текущее состояние

Что уже есть:

- frontend и backend сценарии уже готовы работать с единым внешним user id
- profile, registration и ticket flow уже протаскивают `userId` через бизнес-логику
- в документации boot flow уже фигурирует шаг `ValidateMaxInitData`

Что еще не реализовано:

- реальное подключение MAX JS SDK / MAX JS library во frontend HTML
- backend endpoint для проверки MAX init data
- логика проверки подписи или расшифровки токена
- trusted user context, построенный из MAX payload
- замена временной схемы `X-User-ID` / `userId`

## Целевая архитектура

### Frontend

Frontend должен:

1. подключить реальную MAX Mini App JavaScript library в HTML entrypoint
2. получить init data из MAX runtime
3. отправить raw init data или token payload в backend validation endpoint
4. дождаться валидированного identity/bootstrap response
5. перестать использовать произвольный client-provided `userId` как источник истины

Допустимое временное поведение:

- frontend пока еще может передавать `userId` в payload
- но этот `userId` должен происходить из validated MAX data, а не из захардкоженного значения или ручной подстановки

### Backend

Backend должен:

1. добавить отдельный endpoint, например `POST /api/v1/auth/max/validate`
2. принимать MAX init data / token payload от frontend
3. валидировать подпись или расшифровывать payload с использованием настроенного MAX token/secret
4. извлекать реальный MAX user identifier
5. класть этот identifier в request context для дальнейших handler/service вызовов
6. использовать этот identifier для profile resolution и ITILIUM integration

### ITILIUM

В legacy aiogram-боте использовался:

- Telegram user id

Целевая миграция:

- использовать реальный MAX user id вместо Telegram user id
- позже, если понадобится, ремапить его на другой атрибут ITILIUM в одном backend adapter слое

## Конкретный план работ

### Срез 1. Frontend MAX bootstrap

- подключить реальную MAX JS library в frontend HTML
- добавить тонкий frontend adapter/composable для чтения MAX init data
- определить, какой именно payload отправляется на backend validation endpoint
- встроить bootstrap validation до обычных profile/ticket API вызовов

Результат:

- frontend получает init data из реальной MAX runtime среды, а не из заглушки

### Срез 2. Backend validation endpoint

- добавить route вроде `POST /api/v1/auth/max/validate`
- описать request/response models для MAX validation
- реализовать validate/decrypt с использованием MAX secret/token
- вернуть нормализованный user identity payload

Результат:

- backend умеет подтверждать, что caller является реальным пользователем MAX

### Срез 3. Trusted identity middleware

- убрать доверие к `X-User-ID` / `userId` для защищенных сценариев
- хранить validated MAX user id в request context
- использовать именно context user id в profile и ticket handlers

Результат:

- все user-sensitive routes работают только от trusted identity, полученного после backend validation

### Срез 4. Миграция ITILIUM identity

- передавать validated MAX user id в profile и ticket services
- использовать MAX user id вместо Telegram id при обращении в ITILIUM
- сохранить backend adapter boundary, чтобы потом можно было локально поменять strategy mapping

Результат:

- outbound ITILIUM calls работают уже с MAX-derived identity

### Срез 5. Cleanup переходного режима

- убрать или ограничить fallback на `X-User-ID` / query `userId`
- если оставить debug-bypass, то только за явным флагом конфигурации
- обновить документацию и ручной test checklist

Результат:

- в test/prod режиме не остается случайного insecure identity bypass

## Что потребуется в конфигурации

Для этой интеграции понадобятся MAX-специфичные настройки проекта, например:

- публичный идентификатор MAX Mini App
- MAX secret/token для validate/decrypt
- флаг включения debug bypass для локальной разработки

Точные имена переменных лучше зафиксировать в момент подключения реального MAX SDK контракта.

## Что проверять после внедрения

После реализации MAX identity integration нужно проверить:

1. Mini App открывается внутри реальной MAX среды и JS library действительно загружается
2. frontend получает init data из MAX
3. backend успешно валидирует или расшифровывает token payload
4. извлеченный MAX user id попадает в logs/context как acting user
5. profile lookup работает для реального MAX пользователя
6. этот же user id дальше уходит в ITILIUM calls
7. подмена `X-User-ID` или manual `userId` больше не меняет acting user в защищенных сценариях

## Важное переходное правило

Пока этот план не реализован:

- любой `userId`, который уходит в ITILIUM, нужно считать временным stand-in значением
- по смыслу это уже должен быть "MAX user id вместо Telegram id"
- позже источник этого id нужно заменить централизованно, а не править каждый экран отдельно
