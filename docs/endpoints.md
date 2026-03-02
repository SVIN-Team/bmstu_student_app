### **3.5. Описание REST API**
#### **3.5.1. Общие соглашения**

**Базовый URL:** `https://api.example.com/api/v1`
В локальном окружении используется тот же префикс `/api/v1` (например, `http://localhost:8080/api/v1`).

**Формат данных:** JSON

**Заголовки:**

```
Content-Type: application/json
Authorization: Bearer <access_token>  (для защищённых эндпоинтов)
```

**Формат успешного ответа:**

```json
{
  "success": true,
  "data": { ... }
}
```

**Формат ответа с ошибкой:**

```json
{
  "success": false,
  "error": {
    "code": "QUEUE_FULL",
    "message": "Очередь заполнена"
  }
}
```

**Формат дат:** ISO 8601

* В URL-параметрах: `YYYY-MM-DD` (например, `2026-02-23`)
* В теле запроса/ответа: `YYYY-MM-DDTHH:mm:ssZ` (UTC)

#### **3.5.2. Коды ошибок**

| HTTP Code | Error Code | Описание |
|-----------|------------|----------|
| 400 | `VALIDATION_ERROR` | Ошибка валидации входных данных |
| 401 | `UNAUTHORIZED` | Отсутствует или невалидный токен |
| 401 | `TOKEN_EXPIRED` | Токен истёк |
| 403 | `FORBIDDEN` | Недостаточно прав |
| 404 | `NOT_FOUND` | Ресурс не найден |
| 409 | `ALREADY_EXISTS` | Ресурс уже существует |
| 409 | `QUEUE_FULL` | Очередь заполнена |
| 409 | `QUEUE_CLOSED` | Очередь закрыта для записи |
| 409 | `CONFLICT` | Конфликт конкурентного доступа |
| 409 | `INVALID_STATE_TRANSITION` | Недопустимая смена статуса |
| 429 | `RATE_LIMITED` | Превышен лимит запросов |
| 500 | `INTERNAL_ERROR` | Внутренняя ошибка сервера |

#### **3.5.3. Эндпоинты аутентификации**

| Метод | Эндпоинт | Описание | Доступ |
|-------|----------|----------|--------|
| POST | `/auth/register` | Регистрация | Гость |
| POST | `/auth/login` | Вход в систему | Гость |
| POST | `/auth/refresh` | Обновление токенов | Авторизованный |
| POST | `/auth/logout` | Выход из системы | Авторизованный |
| POST | `/auth/logout-all` | Выход со всех устройств | Авторизованный |

**POST /auth/register**

Request:

```json
{
  "email": "student@university.edu",
  "password": "SecurePass123!",
  "first_name": "Иван",
  "last_name": "Петров",
  "group_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

Response (201 Created):

```json
{
  "success": true,
  "data": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "message": "Регистрация успешна"
  }
}
```

**POST /auth/login**

Request:

```json
{
  "email": "student@university.edu",
  "password": "SecurePass123!"
}
```

Response (200 OK):

```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_in": 900,
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "student@university.edu",
      "role": "student",
      "first_name": "Иван",
      "last_name": "Петров",
      "group_id": "uuid"
    }
  }
}
```

\+ Set-Cookie: `refresh_token=...; HttpOnly; Secure; SameSite=Strict; Path=/api/v1/auth; Max-Age=604800`

#### **3.5.4. Эндпоинты расписания**

| Метод | Эндпоинт | Описание | Доступ |
|-------|----------|----------|--------|
| GET | `/schedule` | Получить расписание | Студент+ |
| GET | `/schedule/lessons/{id}` | Детали занятия | Студент+ |
| POST | `/admin/schedule/import` | Импорт расписания | Админ |

**GET /schedule**

Query Parameters:

* `group_id` (UUID, optional) — если не указан, берётся группа текущего пользователя
* `date_from` (date, required) — начало периода
* `date_to` (date, required) — конец периода

Response (200 OK):

```json
{
  "success": true,
  "data": {
    "lessons": [
      {
        "id": "lesson-uuid-1",
        "subject": {
          "id": "subject-uuid",
          "name": "Базы данных"
        },
        "teacher": {
          "id": "teacher-uuid",
          "full_name": "Иванов И.И."
        },
        "room": "ГУК-513",
        "type": "lecture",
        "starts_at": "2026-02-23T09:00:00Z",
        "ends_at": "2026-02-23T10:30:00Z",
        "has_queue": true,
        "queue_id": "queue-uuid-1"
      }
    ]
  }
}
```

**POST /admin/schedule/import**

Формат: multipart/form-data, файл в формате JSON.

Структура файла:

```json
{
  "group_id": "uuid",
  "lessons": [
    {
      "subject_name": "Базы данных",
      "teacher_name": "Иванов И.И.",
      "room": "ГУК-513",
      "type": "lecture",
      "starts_at": "2026-02-23T09:00:00Z",
      "ends_at": "2026-02-23T10:30:00Z"
    }
  ]
}
```

Если `subject_name` или `teacher_name` не найдены в справочнике — создаются автоматически.

Response (200 OK):

```json
{
  "success": true,
  "data": {
    "imported": 42,
    "created_subjects": ["Базы данных"],
    "created_teachers": ["Иванов И.И."],
    "errors": [
      { "row": 5, "reason": "Некорректный формат времени" }
    ]
  }
}
```

#### **3.5.5. Эндпоинты очередей**

| Метод | Эндпоинт | Описание | Доступ |
|-------|----------|----------|--------|
| GET | `/queues` | Список очередей | Студент+ |
| GET | `/queues/{id}` | Детали очереди | Студент+ |
| POST | `/queues` | Создать очередь | Староста |
| PATCH | `/queues/{id}` | Редактировать очередь | Староста |
| DELETE | `/queues/{id}` | Удалить очередь | Староста |
| POST | `/queues/{id}/signup` | Записаться в очередь | Студент+ |
| DELETE | `/queues/{id}/signup` | Отменить запись | Студент |
| PATCH | `/queues/{id}/slots/{slot_id}` | Изменить статус слота | Староста |
| POST | `/queues/{id}/transfer-failed` | Перенести неуспевших | Староста |

**GET /queues**

Query Parameters:

* `group_id` (UUID, optional) — фильтр по группе (если не указан — очереди своей группы)
* `subject_id` (UUID, optional) — фильтр по предмету
* `status` (string, optional) — фильтр по статусу: `draft`, `open`, `closed`, `archived`
* `page` (int, default: 1)
* `per_page` (int, default: 20, max: 100)

Response (200 OK):

```json
{
  "success": true,
  "data": {
    "queues": [
      {
        "id": "uuid",
        "subject": { "id": "uuid", "name": "Базы данных" },
        "status": "open",
        "opens_at": "2026-02-23T08:00:00Z",
        "closes_at": "2026-02-23T12:00:00Z",
        "slots_count": 15,
        "max_size": 25
      }
    ],
    "pagination": {
      "page": 1,
      "per_page": 20,
      "total": 47
    }
  }
}
```

**GET /queues/{id}**


**Response:**

```json
{
  "success": true,
  "data": {
    "id": "queue-uuid",
    "subject": { "id": "subject-uuid", "name": "Базы данных" },
    "status": "open",
    "opens_at": "2026-02-23T08:00:00Z",
    "closes_at": "2026-02-23T12:00:00Z",
    "max_size": 25,
    "slots_count": 12,
    "my_slot": {
      "slot_id": "uuid",
      "status": "waiting",
      "signed_up_at": "2026-02-23T08:05:10Z"
    },
    "slots": [
        {
            "slot_id": "...",
            "student": { "id": "...", "first_name": "Иван", "last_name": "Петров" },
            "status": "waiting",
            "signed_up_at": "2026-02-23T08:01:00Z"
        }
    ]
  }
}

```

> Поле `my_slot` присутствует, только если пользователь записан.

**POST /queues**

Request:

```json
{
  "subject_id": "subject-uuid",
  "lesson_id": "lesson-uuid",
  "opens_at": "2026-02-23T08:00:00Z",
  "closes_at": "2026-02-23T12:00:00Z",
  "max_size": 25
}
```

Response (201 Created):

```json
{
  "success": true,
  "data": {
    "id": "new-queue-uuid",
    "status": "draft"
  }
}
```

**PATCH /queues/{id}**

Редактирование разрешено только в статусе `draft`. В статусе `open` можно изменить только `closes_at` и `status`.

Request (все поля опциональны):

```json
{
  "status": "open",
  "opens_at": "2026-02-23T09:00:00Z",
  "closes_at": "2026-02-23T13:00:00Z",
  "max_size": 30
}
```

Response (200 OK):

```json
{
  "success": true,
  "data": {
    "id": "queue-uuid",
    "status": "open",
    "opens_at": "2026-02-23T09:00:00Z",
    "closes_at": "2026-02-23T13:00:00Z",
    "max_size": 30
  }
}
```

**POST /queues/{id}/signup**

Request: пустое тело

Response (201 Created):

```json
{
  "success": true,
  "data": {
    "slot": {
      "id": "new-slot-uuid",
      "queue_id": "queue-uuid",
      "student_id": "current-user-uuid",
      "status": "waiting",
      "signed_up_at": "2026-02-23T09:12:34Z"
    }
  }
}

```

Response (409 Conflict — очередь заполнена):

```json
{
  "success": false,
  "error": {
    "code": "QUEUE_FULL",
    "message": "Очередь заполнена. Максимальный размер: 25"
  }
}
```

**DELETE /queues/{id}/signup**

Response (200 OK):

```json
{
  "success": true,
  "data": {
    "message": "Запись отменена"
  }
}
```

**PATCH /queues/{id}/slots/{slot_id}**

Request:

```json
{
  "status": "passed"
}
```

Response (200 OK):

```json
{
  "success": true,
  "data": {
    "slot_id": "slot-uuid",
    "status": "passed"
  }
}
```

**POST /queues/{id}/transfer-failed**

Переносит студентов со статусом `failed` или `no_show` из указанной очереди в начало текущей.

Требования:

* Текущая очередь должна быть в статусе `draft`
* Исходная очередь должна быть в статусе `closed` или `archived`
* Обе очереди должны принадлежать одной группе и одному предмету

Request:

```json
{
  "source_queue_id": "previous-queue-uuid"
}
```

Response (200 OK):

```json
{
  "success": true,
  "data": {
    "transferred_count": 2,
    "transferred_students": [
      {
        "slot_id": "new-slot-uuid-1",
        "student": { "id": "student-uuid-1", "first_name": "Иван", "last_name": "Петров" },
        "original_status": "failed",
        "signed_up_at": "2026-02-24T10:00:00Z"
      }
    ]
  }
}
```

#### **3.5.6. Эндпоинты пользователей**

| Метод | Эндпоинт | Описание | Доступ |
|-------|----------|----------|--------|
| GET | `/users/me` | Профиль текущего пользователя | Авторизованный |
| PATCH | `/users/me` | Редактировать профиль | Авторизованный |
| GET | `/users/me/queues` | Мои записи в очередях | Студент+ |
| POST | `/headman/transfer` | Передать роль старосты | Староста |
| GET | `/admin/users` | Список пользователей | Админ |
| PATCH | `/admin/users/{id}/role` | Изменить роль | Админ |
| POST | `/admin/users/{id}/block` | Заблокировать | Админ |
| DELETE | `/admin/users/{id}/block` | Разблокировать | Админ |

**GET /users/me/queues**

Query Parameters:

* `status` (string, optional) — `waiting`, `passed`, `failed`, `no_show`

Response (200 OK):

```json
{
  "success": true,
  "data": {
    "slots": [
      {
        "slot_id": "slot-uuid",
        "status": "waiting",
        "signed_up_at": "2026-02-23T08:05:10Z",
        "queue": {
          "id": "queue-uuid",
          "subject": { "id": "subj-uuid", "name": "Базы данных" },
          "status": "open",
          "opens_at": "2026-02-23T08:00:00Z"
        }
      }
    ]
  }
}
```

**PATCH /users/me**

Request (все поля опциональны):

```json
{
  "first_name": "Иван",
  "last_name": "Петров",
  "group_id": "uuid"
}
```

Response (200 OK):

```json
{
  "success": true,
  "data": {
    "id": "user-uuid",
    "first_name": "Иван",
    "last_name": "Петров",
    "group_id": "uuid"
  }
}
```

**POST /headman/transfer**

Request:

```json
{
  "to_user_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

Response (200 OK):

```json
{
  "success": true,
  "data": {
    "previous_headman_id": "old-uuid",
    "new_headman_id": "550e8400-e29b-41d4-a716-446655440000",
    "group_id": "group-uuid"
  }
}
```

#### **3.5.7. Эндпоинты справочников (Admin)**

| Метод | Эндпоинт | Описание |
|-------|----------|----------|
| GET | `/admin/groups` | Список групп |
| POST | `/admin/groups` | Создать группу |
| PATCH | `/admin/groups/{id}` | Редактировать группу |
| DELETE | `/admin/groups/{id}` | Удалить группу |
| GET | `/admin/subjects` | Список предметов |
| POST | `/admin/subjects` | Создать предмет |
| PATCH | `/admin/subjects/{id}` | Редактировать предмет |
| DELETE | `/admin/subjects/{id}` | Удалить предмет |
| GET | `/admin/teachers` | Список преподавателей |
| POST | `/admin/teachers` | Создать преподавателя |
| PATCH | `/admin/teachers/{id}` | Редактировать |
| DELETE | `/admin/teachers/{id}` | Удалить |
| GET | `/admin/rooms` | Список аудиторий |
| POST | `/admin/rooms` | Создать аудиторию |