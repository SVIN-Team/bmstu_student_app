### **3.5. Описание REST API**
#### **3.5.1. Общие соглашения**

**Базовый URL:** `https://api.example.com/api/v1`

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
*   В URL-параметрах: `YYYY-MM-DD` (например, `2026-02-23`)
*   В теле запроса/ответа: `YYYY-MM-DDTHH:mm:ssZ` (UTC)

---

#### **3.5.2. Коды ошибок**

| HTTP Code | Error Code                  | Описание                            |
| --------- | --------------------------- | ----------------------------------- |
| 400       | `VALIDATION_ERROR`          | Ошибка валидации входных данных     |
| 401       | `UNAUTHORIZED`              | Отсутствует или невалидный токен    |
| 401       | `TOKEN_EXPIRED`             | Токен истёк                         |
| 403       | `FORBIDDEN`                 | Недостаточно прав                   |
| 404       | `NOT_FOUND`                 | Ресурс не найден                    |
| 409       | `ALREADY_EXISTS`            | Ресурс уже существует               |
| 409       | `QUEUE_FULL`                | Очередь заполнена                   |
| 409       | `QUEUE_CLOSED`              | Очередь закрыта для записи          |
| 409       | `CONFLICT`                  | Конфликт конкурентного доступа      |
| 409       | `INVALID_STATE_TRANSITION`  | Недопустимая смена статуса          |
| 429       | `RATE_LIMITED`              | Превышен лимит запросов             |
| 500       | `INTERNAL_ERROR`            | Внутренняя ошибка сервера           |

---

#### **3.5.3. Ресурс: Аутентификация (`/auth`)**

| Метод | Эндпоинт          | Описание                 | Доступ         |
| ----- | ----------------- | ------------------------ | -------------- |
| POST  | `/auth/register`  | Регистрация              | Гость          |
| POST  | `/auth/login`     | Вход в систему           | Гость          |
| POST  | `/auth/refresh`   | Обновление токенов       | Авторизованный |
| POST  | `/auth/logout`    | Выход из системы         | Авторизованный |
| POST  | `/auth/logout-all`| Выход со всех устройств  | Авторизованный |

**POST /auth/register**

*Request:*
```json
{
  "email": "student@university.edu",
  "password": "SecurePass123!",
  "first_name": "Иван",
  "last_name": "Петров",
  "group_id": "123e4567-e89b-12d3-a456-426614174000"
}
```
*Response (201 Created):*
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "student@university.edu",
    "first_name": "Иван",
    "last_name": "Петров"
  }
}
```

**POST /auth/login**

*Request:*
```json
{
  "email": "student@university.edu",
  "password": "SecurePass123!"
}
```
*Response (200 OK):*
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
      "group": { "id": "uuid", "name": "ИУ7-81Б" }
    }
  }
}
```
\+ `Set-Cookie: refresh_token=...; HttpOnly; Secure; SameSite=Strict; Path=/api/v1/auth; Max-Age=604800`

---

#### **3.5.4. Ресурс: Занятия (`/lessons`)**

| Метод | Эндпоинт       | Описание                       | Доступ   |
| ----- | -------------- | ------------------------------ | -------- |
| GET   | `/lessons`     | Получить список занятий        | Студент+ |
| GET   | `/lessons/{id}`| Получить детали занятия        | Студент+ |
| POST  | `/lessons/imports` | Импорт расписания из файла | Админ    |

**GET /lessons**

*Query Parameters:*
*   `group_id` (UUID, optional) — фильтр по группе. По умолчанию — группа текущего пользователя.
*   `date_from` (date, required) — начало периода.
*   `date_to` (date, required) — конец периода.

*Response (200 OK):*
```json
{
  "success": true,
  "data": [
    {
      "id": "lesson-uuid-1",
      "subject": { "id": "subject-uuid", "name": "Базы данных" },
      "teacher": { "id": "teacher-uuid", "full_name": "Иванов И.И." },
      "room": { "id": "room-uuid", "name": "ГУК-513" },
      "type": "lecture",
      "starts_at": "2026-02-23T09:00:00Z",
      "ends_at": "2026-02-23T10:30:00Z",
      "queue_id": "queue-uuid-1"
    }
  ]
}
```
> Поле `queue_id` присутствует, только если к занятию привязана очередь.

**POST /lessons/imports**

*Request (multipart/form-data):* файл `schedule.json`

*Структура файла:*
```json
{
  "group_id": "uuid",
  "lessons": [
    {
      "subject_name": "Базы данных",
      "teacher_name": "Иванов И.И.",
      "room_name": "ГУК-513",
      "type": "lecture",
      "starts_at": "2026-02-23T09:00:00Z",
      "ends_at": "2026-02-23T10:30:00Z"
    }
  ]
}
```
*Response (200 OK):*
```json
{
  "success": true,
  "data": {
    "imported_count": 42,
    "created_subjects": ["Базы данных"],
    "created_teachers": ["Иванов И.И."],
    "errors": [
      { "row": 5, "message": "Некорректный формат времени" }
    ]
  }
}
```

---

#### **3.5.5. Ресурс: Очереди (`/queues`)**

| Метод  | Эндпоинт          | Описание                | Доступ   |
| ------ | ----------------- | ----------------------- | -------- |
| GET    | `/queues`         | Список очередей         | Студент+ |
| POST   | `/queues`         | Создать очередь         | Староста |
| GET    | `/queues/{id}`    | Детали очереди          | Студент+ |
| PATCH  | `/queues/{id}`    | Редактировать очередь   | Староста |
| DELETE | `/queues/{id}`    | Удалить очередь         | Староста |

**GET /queues**

*Query Parameters:*
*   `group_id` (UUID, optional)
*   `subject_id` (UUID, optional)
*   `status` (string, optional): `draft`, `open`, `closed`, `archived`
*   `page` (int, default: 1)
*   `per_page` (int, default: 20, max: 100)

*Response (200 OK):*
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "subject": { "id": "uuid", "name": "Базы данных" },
      "group": { "id": "uuid", "name": "ИУ7-81Б" },
      "status": "open",
      "opens_at": "2026-02-23T08:00:00Z",
      "closes_at": "2026-02-23T12:00:00Z",
      "slots_count": 15,
      "max_size": 25
    }
  ],
  "pagination": { "page": 1, "per_page": 20, "total": 47 }
}
```

**POST /queues**

*Request:*
```json
{
  "subject_id": "subject-uuid",
  "lesson_id": "lesson-uuid",
  "opens_at": "2026-02-23T08:00:00Z",
  "closes_at": "2026-02-23T12:00:00Z",
  "max_size": 25
}
```
*Response (201 Created):*
```json
{
  "success": true,
  "data": {
    "id": "new-queue-uuid",
    "group": { "id": "uuid", "name": "ИУ7-81Б" },
    "subject": { "id": "subject-uuid", "name": "Базы данных" },
    "status": "draft",
    "opens_at": "2026-02-23T08:00:00Z",
    "closes_at": "2026-02-23T12:00:00Z",
    "max_size": 25
  }
}
```

**GET /queues/{id}**

*Response (200 OK):*
```json
{
  "success": true,
  "data": {
    "id": "queue-uuid",
    "subject": { "id": "subject-uuid", "name": "Базы данных" },
    "group": { "id": "group-uuid", "name": "ИУ7-81Б" },
    "lesson_id": "lesson-uuid",
    "status": "open",
    "opens_at": "2026-02-23T08:00:00Z",
    "closes_at": "2026-02-23T12:00:00Z",
    "max_size": 25,
    "created_by": { "id": "user-uuid", "first_name": "Пётр", "last_name": "Сидоров" },
    "created_at": "2026-02-20T10:00:00Z"
  }
}
```
> Для получения списка слотов используйте отдельный эндпоинт `GET /queues/{id}/slots`.

**PATCH /queues/{id}**

*Request (все поля опциональны):*
```json
{
  "status": "open",
  "closes_at": "2026-02-23T13:00:00Z",
  "max_size": 30
}
```
*Response (200 OK):* Полный объект очереди (как в `GET /queues/{id}`).

---

#### **3.5.6. Ресурс: Слоты в очереди (`/queues/{queue_id}/slots`)**

Слоты — это вложенный ресурс, принадлежащий очереди.

| Метод  | Эндпоинт                          | Описание                     | Доступ   |
| ------ | --------------------------------- | ---------------------------- | -------- |
| GET    | `/queues/{queue_id}/slots`        | Список слотов в очереди      | Студент+ |
| POST   | `/queues/{queue_id}/slots`        | Записаться в очередь         | Студент+ |
| GET    | `/queues/{queue_id}/slots/{id}`   | Получить слот                | Студент+ |
| PATCH  | `/queues/{queue_id}/slots/{id}`   | Изменить статус слота        | Староста |
| DELETE | `/queues/{queue_id}/slots/{id}`   | Отменить запись (свой слот)  | Студент  |

**GET /queues/{queue_id}/slots**

*Response (200 OK):*
Список отсортирован по `signed_up_at`.
```json
{
  "success": true,
  "data": [
    {
      "id": "slot-uuid-1",
      "student": { "id": "...", "first_name": "Иван", "last_name": "Петров" },
      "status": "waiting",
      "signed_up_at": "2026-02-23T08:01:00Z"
    },
    {
      "id": "slot-uuid-2",
      "student": { "id": "...", "first_name": "Мария", "last_name": "Сидорова" },
      "status": "waiting",
      "signed_up_at": "2026-02-23T08:02:30Z"
    }
  ]
}
```

**POST /queues/{queue_id}/slots**

*Request:* Пустое тело `{}`

*Response (201 Created):*
```json
{
  "success": true,
  "data": {
    "id": "new-slot-uuid",
    "queue_id": "queue-uuid",
    "student": { "id": "current-user-uuid", "first_name": "Иван", "last_name": "Петров" },
    "status": "waiting",
    "signed_up_at": "2026-02-23T09:12:34Z"
  }
}
```

**PATCH /queues/{queue_id}/slots/{id}**

*Request:*
```json
{ "status": "passed" }
```
*Response (200 OK):* Полный объект слота.

**DELETE /queues/{queue_id}/slots/{id}**

*Response (204 No Content):* Пустое тело.

---

#### **3.5.7. Ресурс: Переносы слотов (`/queues/{queue_id}/transfers`)**

Это ресурс-контроллер для действия "перенос неуспевших".

| Метод | Эндпоинт                         | Описание             | Доступ   |
| ----- | -------------------------------- | -------------------- | -------- |
| POST  | `/queues/{queue_id}/transfers`   | Перенести неуспевших | Староста |

**POST /queues/{queue_id}/transfers**

Переносит студентов со статусом `failed` или `no_show` из указанной очереди в текущую.

*Требования:*
*   Текущая очередь (`{queue_id}`) должна быть в статусе `draft`.
*   Исходная очередь (`source_queue_id`) должна быть `closed` или `archived`.
*   Обе очереди должны принадлежать одной группе и одному предмету.

*Request:*
```json
{ "source_queue_id": "previous-queue-uuid" }
```
*Response (201 Created):*
```json
{
  "success": true,
  "data": {
    "transferred_count": 2,
    "slots": [
      {
        "id": "new-slot-uuid-1",
        "student": { "id": "student-uuid-1", "first_name": "Иван", "last_name": "Петров" },
        "original_status": "failed",
        "signed_up_at": "2026-02-24T10:00:00Z"
      }
    ]
  }
}
```

---

#### **3.5.8. Ресурс: Текущий пользователь (`/users/me`)**

| Метод | Эндпоинт               | Описание                     | Доступ         |
| ----- | ---------------------- | ---------------------------- | -------------- |
| GET   | `/users/me`            | Профиль текущего пользователя| Авторизованный |
| PATCH | `/users/me`            | Редактировать профиль        | Авторизованный |
| GET   | `/users/me/slots`      | Мои записи в очередях        | Студент+       |
| PUT   | `/users/me/headman-role` | Передать роль старосты     | Староста       |

**GET /users/me**

*Response (200 OK):*
```json
{
  "success": true,
  "data": {
    "id": "user-uuid",
    "email": "student@university.edu",
    "first_name": "Иван",
    "last_name": "Петров",
    "role": "student",
    "group": { "id": "uuid", "name": "ИУ7-81Б" },
    "created_at": "2026-01-15T12:00:00Z"
  }
}
```

**PATCH /users/me**

*Request (все поля опциональны):*
```json
{
  "first_name": "Иван",
  "last_name": "Петров",
  "group_id": "new-group-uuid"
}
```
*Response (200 OK):* Полный объект пользователя.

**GET /users/me/slots**

*Query Parameters:*
*   `status` (string, optional): `waiting`, `passed`, `failed`, `no_show`
*   `queue_status` (string, optional): `open`, `closed` — фильтр по статусу очереди.

*Response (200 OK):*
```json
{
  "success": true,
  "data": [
    {
      "id": "slot-uuid",
      "status": "waiting",
      "signed_up_at": "2026-02-23T08:05:10Z",
      "queue": {
        "id": "queue-uuid",
        "subject": { "id": "subj-uuid", "name": "Базы данных" },
        "status": "open"
      }
    }
  ]
}
```

**PUT /users/me/headman-role**

Передача роли старосты. Используем `PUT`, так как мы "заменяем" владельца роли.

*Request:*
```json
{ "to_user_id": "550e8400-e29b-41d4-a716-446655440000" }
```
*Response (200 OK):*
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

---

#### **3.5.9. Ресурсы: Администрирование (`/admin/...`)**

**Пользователи (`/admin/users`)**

| Метод  | Эндпоинт                    | Описание           |
| ------ | --------------------------- | ------------------ |
| GET    | `/admin/users`              | Список пользователей |
| GET    | `/admin/users/{id}`         | Получить пользователя |
| PATCH  | `/admin/users/{id}`         | Редактировать (роль, блокировка) |
| DELETE | `/admin/users/{id}`         | Удалить пользователя |

*PATCH /admin/users/{id} — Request:*
```json
{
  "role": "headman",
  "is_blocked": false
}
```

**Справочники**

Стандартный CRUD для всех справочников:

| Ресурс       | Эндпоинты                                              |
| ------------ | ------------------------------------------------------ |
| Группы       | `GET`, `POST` `/admin/groups`; `GET`, `PATCH`, `DELETE` `/admin/groups/{id}` |
| Предметы     | `GET`, `POST` `/admin/subjects`; `GET`, `PATCH`, `DELETE` `/admin/subjects/{id}` |
| Преподаватели| `GET`, `POST` `/admin/teachers`; `GET`, `PATCH`, `DELETE` `/admin/teachers/{id}` |
| Аудитории    | `GET`, `POST` `/admin/rooms`; `GET`, `PATCH`, `DELETE` `/admin/rooms/{id}` |
