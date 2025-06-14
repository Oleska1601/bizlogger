# Bizlogger - Структурированный логгер для бизнес логов

**Вывод логов** осуществляется в консоль в формате JSON и в postgresql. В случае отсутствия доступа к postgresql логи записываются в файл и в определенное заданное пользователем время будут пытаться осуществить запись в БД. В случае возникновения ошибки в процессе записи в файл логи добавляются в локальную очередь, из которой через заданное пользователем время будут пытаться осуществить запись в файл. 
При успешной вставке в файл логи из очереди очищаются.
При возобновлении доступа к postgresql значения переносятся из файла и также очищаются.

## Формат

1. Временная метка в удобочитаемом формате ("ISO8601")
2. Тип события (CREATE, UPDATE, DELETE)
3. Сущность (user, context)
4. Username/login пользователя, кто совершает действие
5. Роль пользователя (SA, CA - только они могут менять сущности.)
6. Контекст пользователя (если SA, то nil)
7. Само действие с описанием и указанием данных:
- CREATE: ID сущности (username или contextID)
- UPDATE: ID сущности (username или contextID) + old value + new value(кроме паролей. )
- DELETE: ID сущности (username или contextID)

## Установка

```bash
go get github.com/Oleska1601/bizlogger
```

## Структура в БД
```sql
CREATE TABLE business_logs (
id SERIAL PRIMARY KEY,
timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
event_type VARCHAR(10) NOT NULL, -- CREATE, UPDATE, DELETE
entity VARCHAR(20) NOT NULL, -- user, context
username VARCHAR(36) NOT NULL,
user_role VARCHAR(2) NOT NULL, -- SA, CA
context VARCHAR(50),
entity_id VARCHAR(50) NOT NULL,
old_value JSONB,
new_value JSONB,
description TEXT
);
```

## Использование

### Базовая инициализация

```go
ctx := context.Background()
processFileTime := time.Duration(time.Second * 5)
processQueueTime := time.Duration(time.Second * 3)
bizlog, close, err := bizlogger.New(ctx, pgRepoInterface, processFileTime, processQueueTime)
```

### Примеры логирования

Создание пользователя (SA):
```go
description := "Created new CA user for context test_ctx"
bizlogger.LogCreate(bizlogger.EntityUser, "admin", bizlogger.UserRoleSA, nil, "new_user", map[string]string{
	"username": "new_user",
	"role":     "CA",
	"email":    "new@example.com",
	"context":  "test_ctx",
}, &description)
```

Обновление контекста (CA):
```go
context := "test_ctx"
description := "Updated context description"
bizlog.LogUpdate(bizlogger.EntityContext, "user_ca_text_ctx", bizlogger.UserRoleCA, &context, "entity_id", map[string]string{
	"description": "Old description",
	"name":        "Old name",
}, map[string]string{"description": "New updated description",
	"name": "Text context"}, &description)
```

Удаление пользователя (SA):
```go
description := "Deleted user old_user with role CA"
bizlog.LogDelete(bizlogger.EntityUser, "admin22", bizlogger.UserRoleSA, nil, "test_user", &description)
```

## Формат вывода

Пример вывода логов:

Создание пользователя (SA):
```json
{
  "timestamp": "2023-11-15T14:30:45.678Z",
  "event_type": "CREATE",
  "entity": "user",
  "username": "admin",
  "user_role": "SA",
  "context": null,
  "entity_id": "new_user",
  "new_value": {
    "username": "new_user",
    "role": "CA",
    "email": "new@example.com",
    "context": "test_ctx"
  },
  "description": "Created new CA user for context test_ctx"
}
```

Обновление контекста (CA):
```json
{
  "timestamp": "2023-11-15T14:35:50.901Z",
  "event_type": "UPDATE",
  "entity": "context",
  "username": "user_ca_text_ctx",
  "user_role": "CA",
  "context": "test_ctx",
  "entity_id": "test_ctx",
  "old_value": {
    "description": "Old description",
    "name": "Old name"
  },
  "new_value": {
    "description": "New updated description",
    "name": "Text context"
  },
  "description": "Updated context description"
}
```

Удаление пользователя (SA):
```json
{
  "timestamp": "2023-11-15T14:40:55.234Z",
  "event_type": "DELETE",
  "entity": "user",
  "username": "admin",
  "user_role": "SA",
  "context": null,
  "entity_id": "test_user",
  "description": "Deleted user old_user with role CA"
}
```
