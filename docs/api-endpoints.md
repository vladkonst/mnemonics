# API Endpoints - Спецификация Backend API

**Версия**: 2.2
**Дата**: 20 января 2026
**Статус**: ✅ Полная спецификация (все 29 endpoints документированы)

> Полный список RESTful API endpoints системы подготовки к экзаменам по анатомии

## Changelog

- **v2.2** (20.01.2026):
  - ✅ Реализованы все недостающие endpoints (GET /api/v1/users/{user_id}/theme/{theme_id}/access)
  - ✅ Удалены дублирующие endpoints (POST /api/v1/commands/*)
  - ✅ Удален устаревший endpoint (POST /api/v1/subscriptions/validate-promo)
  - ✅ Полная спецификация admin endpoints (9 endpoints)
  - ✅ Обновлена сводная таблица с правильными RESTful путями
  - ✅ Итого: 29 endpoints полностью документированы

- **v2.1** (19.01.2026):
  - ✅ Все endpoints рефакторены к RESTful стандартам (убраны глаголы из URL)
  - ✅ Добавлена поддержка сущности "Введение" (is_introduction) в ответах
  - ✅ Унифицирован endpoint создания подписки (promo + payment)
  - ✅ Идемпотентность: PUT для submit-test, POST 200 для webhooks
  - ✅ Правильные HTTP status codes и Location headers

- **v1.0** (28.12.2025):
  - Базовая спецификация API endpoints

## Обозначения

- ✅ **Реализован** - endpoint полностью документирован с примерами запросов/ответов
- 🔴 **Критический** - блокирует запуск MVP
- 🟡 **Высокий** - снижает качество UX
- 🟠 **Средний** - дополнительные улучшения
- 🟢 **Низкий** - административные функции (реализованы через GoAdmin)

## Содержание

- [1. Управление пользователями](#1-управление-пользователями)
- [2. Контент (модули, темы, мнемоники, тесты)](#2-контент-модули-темы-мнемоники-тесты)
- [3. Прогресс пользователей](#3-прогресс-пользователей)
- [4. Подписки и промокоды](#4-подписки-и-промокоды)
- [5. Платежи](#5-платежи)
- [6. Функционал преподавателя](#6-функционал-преподавателя)
- [7. Администрирование](#7-администрирование)

---

## 1. Управление пользователями

### ✅ POST /api/v1/users
**Описание**: Регистрация нового пользователя при первом запуске бота (RESTful)

**Приоритет**: 🔴 КРИТИЧЕСКИЙ

**Параметры запроса**:
```json
{
  "telegram_id": 123456789,
  "first_name": "Иван",
  "last_name": "Иванов",
  "username": "ivan_ivanov"
}
```

**Ответ**:
```json
{
  "user_id": "123456789",
  "role": "student",
  "subscription_status": "inactive",
  "created_at": "2025-12-28T10:00:00Z"
}
```

**HTTP Status**: 201 Created

**Headers**: `Location: /api/v1/users/123456789`

**Изменяемые сущности**: `users` (INSERT)

**Обработка ошибок**:
- 409 Conflict - пользователь уже существует
- 400 Bad Request - невалидный telegram_id

---

### ✅ PATCH /api/v1/users/{telegram_id}
**Описание**: Обновление данных пользователя (роль, настройки) (RESTful)

**Приоритет**: 🔴 ВЫСОКИЙ

**Параметры запроса**:
```json
{
  "role": "student" | "teacher",
  "language": "ru" | "en",
  "notifications_enabled": true
}
```

**Ответ**:
```json
{
  "user_id": "123456789",
  "role": "teacher",
  "language": "ru",
  "updated_at": "2025-12-28T10:05:00Z"
}
```

**HTTP Status**: 200 OK

**Изменяемые сущности**: `users` (UPDATE)

**Обработка ошибок**:
- 404 Not Found - пользователь не найден
- 400 Bad Request - невалидные данные

---

### ✅ GET /api/v1/users/{user_id}/subscription
**Описание**: Получение информации о подписке пользователя (RESTful)

**Приоритет**: 🔴 ВЫСОКИЙ

**Ответ**:
```json
{
  "has_subscription": true,
  "type": "personal" | "university",
  "status": "active" | "expired" | "cancelled",
  "expires_at": "2026-01-28T10:00:00Z",
  "plan": "monthly",
  "university_name": "МГУ",
  "teacher_name": "Иван Иванов"
}
```

**HTTP Status**:
- 200 OK - подписка найдена
- 404 Not Found - нет активной подписки

**Изменяемые сущности**: - (только SELECT)

---

## 2. Контент (модули, темы, мнемоники, тесты)

### ✅ GET /api/v1/content/modules
**Описание**: Получение списка всех модулей с учетом прогресса пользователя

**Приоритет**: 🔴 ВЫСОКИЙ

**Параметры запроса**:
```
?user_id=123456789
```

**Ответ**:
```json
{
  "modules": [
    {
      "id": 1,
      "name": "Анатомия костей",
      "description": "Изучение костной системы человека",
      "order_num": 1,
      "total_themes": 10,
      "completed_themes": 3,
      "is_accessible": true
    }
  ]
}
```

**Изменяемые сущности**: - (только SELECT)

**Источник**: container.puml

---

### ✅ GET /api/v1/content/modules/{module_id}/themes
**Описание**: Получение списка тем конкретного модуля с иконками доступности (RESTful)

**Приоритет**: 🔴 КРИТИЧЕСКИЙ

**Параметры запроса**:
```
?user_id=123456789
```

**Ответ**:
```json
{
  "module_id": 1,
  "module_name": "Анатомия костей",
  "themes": [
    {
      "id": 1,
      "name": "Введение",
      "order_num": 1,
      "is_introduction": true,
      "is_accessible": true,
      "is_completed": true,
      "score": 90
    },
    {
      "id": 2,
      "name": "Строение черепа",
      "order_num": 2,
      "is_introduction": false,
      "is_accessible": true,
      "is_completed": true,
      "score": 85
    },
    {
      "id": 3,
      "name": "Позвоночник",
      "order_num": 3,
      "is_introduction": false,
      "is_accessible": true,
      "is_completed": false,
      "score": null
    },
    {
      "id": 4,
      "name": "Грудная клетка",
      "order_num": 4,
      "is_introduction": false,
      "is_accessible": false,
      "is_completed": false,
      "score": null,
      "locked_reason": "Пройдите тему 'Позвоночник'"
    }
  ]
}
```

**HTTP Status**: 200 OK

**Изменяемые сущности**: - (только SELECT)

**Логика доступности**:
- Если подписка активна → все темы доступны (введение не обязательно)
- Если подписки нет → последовательный доступ начиная с введения (order_num по порядку)

---

### ✅ POST /api/v1/users/{user_id}/study-sessions
**Описание**: Создание сессии изучения темы + получение контента (RESTful, Hybrid)

**Приоритет**: 🔴 КРИТИЧЕСКИЙ

**Параметры запроса**:
```json
{
  "theme_id": 2
}
```

**Ответ**:
```json
{
  "session_id": "uuid-abc-123",
  "theme": {
    "id": 2,
    "name": "Позвоночник",
    "description": "Изучение строения позвоночного столба",
    "is_introduction": false,
    "mnemonics": [
      {
        "id": 1,
        "type": "text",
        "content_text": "Мнемоника для запоминания...",
        "order_num": 1
      },
      {
        "id": 2,
        "type": "image",
        "s3_image_url": "https://s3.amazonaws.com/mnemo/images/spine_01.jpg",
        "order_num": 2
      }
    ]
  },
  "test_available": true,
  "next_step": {
    "type": "test",
    "test_id": 5
  }
}
```

**HTTP Status**: 201 Created

**Side Effect**: `user_progress` (UPDATE status=started, started_at=NOW())

**Обработка ошибок**:
- 403 Forbidden - нет доступа к теме (требуется подписка или прохождение предыдущей темы)
- 404 Not Found - тема не найдена

---

### ✅ POST /api/v1/users/{user_id}/test-attempts
**Описание**: Создание попытки прохождения теста + получение вопросов (RESTful)

**Приоритет**: 🔴 КРИТИЧЕСКИЙ

**Параметры запроса**:
```json
{
  "theme_id": 2
}
```

**Ответ**:
```json
{
  "attempt_id": "uuid-test-attempt-456",
  "test": {
    "id": 5,
    "theme_id": 2,
    "difficulty": 3,
    "passing_score": 70,
    "questions": [
      {
        "id": 1,
        "question": "Сколько позвонков в шейном отделе?",
        "type": "multiple_choice",
        "options": ["5", "7", "9", "12"],
        "order_num": 1
      }
    ]
  },
  "attempt_number": 1,
  "started_at": "2025-12-28T10:30:00Z"
}
```

**HTTP Status**: 201 Created

**Headers**: `Location: /api/v1/users/{user_id}/test-attempts/{attempt_id}`

**Side Effect**: `user_progress` (UPDATE test_started_at, current_attempt++)

**Обработка ошибок**:
- 404 Not Found - тест не найден
- 403 Forbidden - нет доступа к теме

---

### ✅ PUT /api/v1/users/{user_id}/test-attempts/{attempt_id}
**Описание**: Завершение попытки теста (отправка ответов) - идемпотентный (RESTful)

**Приоритет**: 🔴 КРИТИЧЕСКИЙ

**Параметры запроса**:
```json
{
  "answers": [
    {"question_id": 1, "answer": "7"},
    {"question_id": 2, "answer": "B"}
  ]
}
```

**Ответ**:
```json
{
  "result": {
    "score": 85,
    "passing_score": 70,
    "passed": true,
    "correct_answers": 17,
    "total_questions": 20,
    "attempt_number": 1
  },
  "next_action": {
    "type": "next_theme",
    "theme_id": 3,
    "theme_name": "Грудная клетка",
    "is_introduction": false,
    "message": "Отличный результат! Переходите к следующей теме."
  },
  "motivation_message": "Поздравляем! Вы успешно освоили тему 'Позвоночник'."
}
```

**HTTP Status**: 200 OK (идемпотентный - повторная отправка тех же ответов возвращает тот же результат)

**Side Effects**:
- `user_progress` (UPDATE score, status=completed/failed, completed_at)
- `test_attempts` (INSERT attempt record) - только при первом вызове
- Разблокировка next_theme (если passed и нет подписки)

**Обработка ошибок**:
- 400 Bad Request - неполный набор ответов
- 404 Not Found - attempt_id не найден
- 409 Conflict - попытка уже завершена с другими ответами

**Источник**: study_test_sequence.puml

---

### ✅ GET /api/v1/users/{user_id}/theme/{theme_id}/access
**Описание**: Проверка доступа к конкретной теме

**Приоритет**: 🟡 ВЫСОКИЙ

**Ответ** (доступ запрещен):
```json
{
  "accessible": false,
  "reason": "previous_theme_required",
  "required_theme_id": 2,
  "required_theme_name": "Позвоночник",
  "required_action": "complete_test"
}
```

**Ответ** (доступ разрешен):
```json
{
  "accessible": true,
  "access_type": "subscription" | "sequential"
}
```

**HTTP Status**:
- 200 OK - всегда возвращает информацию о доступе

**Изменяемые сущности**: - (только SELECT)

**Логика**:
1. Проверка подписки пользователя
2. Если подписка активна → accessible=true, access_type="subscription"
3. Если подписки нет:
   - Проверка order_num предыдущей темы
   - Если предыдущая тема пройдена → accessible=true, access_type="sequential"
   - Если нет → accessible=false, указать required_theme_id

**Обработка ошибок**:
- 404 Not Found - тема не найдена

---

## 3. Прогресс пользователей

### ✅ GET /api/v1/users/{user_id}/progress
**Описание**: Получение общей статистики прогресса пользователя (RESTful)

**Приоритет**: 🔴 ВЫСОКИЙ

**Ответ**:
```json
{
  "user_id": 123456789,
  "overall_progress": {
    "total_modules": 5,
    "completed_modules": 1,
    "total_themes": 50,
    "completed_themes": 12,
    "completion_percentage": 24,
    "average_score": 82.5,
    "total_attempts": 15,
    "study_days": 7
  },
  "recent_activity": [
    {
      "theme_id": 2,
      "theme_name": "Позвоночник",
      "score": 85,
      "completed_at": "2025-12-28T09:00:00Z"
    }
  ]
}
```

**Изменяемые сущности**: - (только SELECT)

**Источник**: container.puml

---

### ✅ GET /api/v1/users/{user_id}/progress/modules/{module_id}
**Описание**: Детальный прогресс пользователя по конкретному модулю (RESTful)

**Приоритет**: 🟡 ВЫСОКИЙ

**Ответ**:
```json
{
  "module_id": 1,
  "module_name": "Анатомия костей",
  "completion_percentage": 30,
  "themes": [
    {
      "theme_id": 1,
      "theme_name": "Введение",
      "is_introduction": true,
      "status": "completed",
      "score": 95,
      "attempts": 1,
      "completed_at": "2025-12-26T12:00:00Z"
    },
    {
      "theme_id": 2,
      "theme_name": "Строение черепа",
      "is_introduction": false,
      "status": "completed",
      "score": 90,
      "attempts": 1,
      "completed_at": "2025-12-27T15:00:00Z"
    },
    {
      "theme_id": 3,
      "theme_name": "Позвоночник",
      "is_introduction": false,
      "status": "completed",
      "score": 85,
      "attempts": 2,
      "completed_at": "2025-12-28T09:00:00Z"
    },
    {
      "theme_id": 4,
      "theme_name": "Грудная клетка",
      "is_introduction": false,
      "status": "started",
      "score": null,
      "attempts": 0,
      "completed_at": null
    }
  ]
}
```

**HTTP Status**: 200 OK

**Изменяемые сущности**: - (только SELECT)

---

## 4. Подписки и промокоды

### ✅ POST /api/v1/teachers/{teacher_id}/promo-codes
**Описание**: Активация промокода преподавателем (RESTful)

**Приоритет**: 🔴 КРИТИЧЕСКИЙ

**Параметры запроса**:
```json
{
  "code": "ABC123"
}
```

**Ответ**:
```json
{
  "code": "ABC123",
  "university_name": "МГУ",
  "remaining_activations": 49,
  "max_activations": 50,
  "expires_at": "2026-06-30T23:59:59Z",
  "message": "Промокод активирован. Передайте студентам: ABC123"
}
```

**Изменяемые сущности**: `promo_codes` (UPDATE teacher_id, activated_at, remaining)

**Обработка ошибок**:
- 403 Forbidden - пользователь не преподаватель
- 404 Not Found - промокод не найден
- 400 Bad Request - промокод уже активирован, исчерпан, истек

---

### ✅ POST /api/v1/users/{user_id}/subscriptions
**Описание**: Создание подписки (RESTful) - унифицированный endpoint для promo и payment

**Приоритет**: 🔴 КРИТИЧЕСКИЙ

**Параметры запроса** (promo):
```json
{
  "promo_code": "ABC123"
}
```

**Параметры запроса** (payment):
```json
{
  "payment_id": "pay_abc123xyz",
  "plan": "monthly"
}
```

**Ответ**:
```json
{
  "subscription_id": "sub_xyz789",
  "type": "university",
  "status": "active",
  "plan": "ABC123",
  "expires_at": null,
  "university_name": "МГУ",
  "teacher_name": "Профессор Иванов",
  "message": "Добро пожаловать! Доступ к материалам преподавателя открыт."
}
```

**HTTP Status**: 201 Created

**Side Effects** (promo):
- `teacher_promo_students` (INSERT связь teacher-student)
- `users` (UPDATE subscription_status=active, university_code=promo)
- `promo_codes` (UPDATE remaining--)

**Side Effects** (payment):
- `subscriptions` (INSERT подписка type=personal)
- `users` (UPDATE subscription_status=active)

**Обработка ошибок**:
- 404 Not Found - промокод/платеж не найден
- 400 Bad Request - промокод исчерпан, истек
- 409 Conflict - подписка уже существует

---

### ✅ GET /api/v1/teachers/{teacher_id}/promo-codes
**Описание**: Список промокодов, активированных преподавателем (RESTful)

**Приоритет**: 🟡 ВЫСОКИЙ

**Ответ**:
```json
{
  "promo_codes": [
    {
      "code": "ABC123",
      "university_name": "МГУ",
      "max_activations": 50,
      "remaining": 35,
      "used": 15,
      "activated_at": "2025-09-01T10:00:00Z",
      "expires_at": "2026-06-30T23:59:59Z",
      "status": "active"
    }
  ],
  "total": 1
}
```

**HTTP Status**: 200 OK

**Изменяемые сущности**: - (только SELECT)

**Обработка ошибок**:
- 403 Forbidden - пользователь не является преподавателем
- 404 Not Found - преподаватель не найден

**SQL логика**:
```sql
SELECT pc.*, COUNT(tps.student_id) as used
FROM promo_codes pc
LEFT JOIN teacher_promo_students tps ON pc.code = tps.promo_code
WHERE pc.teacher_id = ?
GROUP BY pc.code
```

---

## 5. Платежи

### ✅ POST /api/v1/users/{user_id}/payment-invoices
**Описание**: Создание счета для оплаты персональной подписки (RESTful)

**Приоритет**: 🔴 КРИТИЧЕСКИЙ

**Параметры запроса**:
```json
{
  "plan": "monthly" | "yearly"
}
```

**Ответ**:
```json
{
  "invoice_id": "inv_abc123xyz",
  "payment_url": "https://payment-gateway.com/checkout/inv_abc123xyz",
  "amount": 500,
  "currency": "RUB",
  "plan": "monthly",
  "expires_at": "2025-12-28T11:00:00Z"
}
```

**HTTP Status**: 201 Created

**Headers**: `Location: /api/v1/users/{user_id}/payment-invoices/{invoice_id}`

**Side Effect**: `users` (UPDATE pending_payment_id='inv_abc123xyz')

**Обработка ошибок**:
- 409 Conflict - у пользователя уже есть активная подписка
- 500 Internal Server Error - ошибка Payment Gateway

---

### ✅ POST /api/v1/webhooks/payment-gateway
**Описание**: Обработка webhook от платежного шлюза о статусе платежа (RESTful)

**Приоритет**: 🔴 КРИТИЧЕСКИЙ

**Параметры запроса** (от Payment Gateway):
```json
{
  "payment_id": "pay_abc123xyz",
  "status": "succeeded" | "failed" | "pending",
  "user_id": 123456789,
  "amount": 500,
  "currency": "RUB",
  "metadata": {
    "plan": "monthly"
  },
  "signature": "sha256_hash"
}
```

**Ответ**:
```json
{
  "received": true
}
```

**HTTP Status**: 200 OK (всегда, даже при ошибках - для идемпотентности)

**Логика**:
1. Верификация подписи webhook
2. Проверка идемпотентности (SELECT subscriptions WHERE payment_id)
3. Если status='succeeded':
   - INSERT subscriptions (user_id, payment_id, status=active, expires_at)
   - UPDATE users SET subscription_status=active, pending_payment_id=NULL
   - Отправка уведомления в Telegram
4. Если status='failed':
   - UPDATE users SET pending_payment_id=NULL
   - Отправка уведомления в Telegram об ошибке
5. RETURN HTTP 200

**Side Effects**:
- `subscriptions` (INSERT) - только при succeeded
- `users` (UPDATE subscription_status, pending_payment_id)

**Обработка ошибок**:
- 400 Bad Request - невалидная подпись
- 200 OK - даже если платеж уже обработан (идемпотентность)

---

### ✅ GET /api/v1/users/{user_id}/payment-invoices/pending
**Описание**: Получение информации о pending invoice (RESTful filtering pattern)

**Приоритет**: 🟡 СРЕДНИЙ

**Ответ**:
```json
{
  "invoice_id": "inv_abc123xyz",
  "status": "pending" | "succeeded" | "failed",
  "payment_url": "https://payment-gateway.com/checkout/inv_abc123xyz",
  "amount": 500,
  "created_at": "2025-12-28T10:30:00Z"
}
```

**HTTP Status**:
- 200 OK - pending invoice найден
- 404 Not Found - нет pending invoice

**Изменяемые сущности**: - (SELECT + возможен вызов Payment Gateway API)

---

## 6. Функционал преподавателя

### ✅ GET /api/v1/teachers/{teacher_id}/students
**Описание**: Список студентов преподавателя (RESTful)

**Приоритет**: 🔴 КРИТИЧЕСКИЙ

**Параметры запроса**:
```
?teacher_id=987654321
```

**Ответ**:
```json
{
  "students": [
    {
      "student_id": 123456789,
      "name": "Иван Иванов",
      "username": "ivan_ivanov",
      "joined_at": "2025-09-15T10:00:00Z",
      "university_code": "MGU",
      "progress": {
        "total_themes": 50,
        "completed_themes": 12,
        "completion_percentage": 24,
        "average_score": 82.5
      },
      "last_activity": "2025-12-28T09:00:00Z"
    }
  ],
  "total_students": 15,
  "active_students": 12
}
```

**Изменяемые сущности**: - (только SELECT)

**SQL логика**:
```sql
SELECT u.*, tp.joined_at, COUNT(up.theme_id) as completed_themes, AVG(up.score) as avg_score
FROM users u
JOIN teacher_promo_students tp ON u.telegram_id = tp.student_id
LEFT JOIN user_progress up ON u.telegram_id = up.user_id AND up.status = 'completed'
WHERE tp.teacher_id = ?
GROUP BY u.telegram_id
```

---

### ✅ GET /api/v1/teachers/{teacher_id}/students/{student_id}/progress
**Описание**: Детальный прогресс конкретного студента (RESTful)

**Приоритет**: 🔴 КРИТИЧЕСКИЙ

**Ответ**:
```json
{
  "student": {
    "id": 123456789,
    "name": "Иван Иванов",
    "joined_at": "2025-09-15T10:00:00Z"
  },
  "modules": [
    {
      "module_id": 1,
      "module_name": "Анатомия костей",
      "completion_percentage": 40,
      "themes": [
        {
          "theme_id": 1,
          "theme_name": "Введение",
          "is_introduction": true,
          "status": "completed",
          "score": 95,
          "attempts": 1,
          "completed_at": "2025-12-26T10:00:00Z"
        },
        {
          "theme_id": 2,
          "theme_name": "Строение черепа",
          "is_introduction": false,
          "status": "completed",
          "score": 90,
          "attempts": 1,
          "completed_at": "2025-12-27T15:00:00Z"
        }
      ]
    }
  ]
}
```

**HTTP Status**: 200 OK

**Изменяемые сущности**: - (только SELECT)

**Обработка ошибок**:
- 403 Forbidden - студент не принадлежит этому преподавателю
- 404 Not Found - студент не найден

**SQL проверка доступа**:
```sql
SELECT * FROM teacher_promo_students
WHERE teacher_id = ? AND student_id = ?
```

---

### ✅ GET /api/v1/teachers/{teacher_id}/statistics
**Описание**: Статистика по группе студентов преподавателя (RESTful)

**Приоритет**: 🟡 СРЕДНИЙ

**Параметры запроса**:
```
?teacher_id=987654321
```

**Ответ**:
```json
{
  "total_students": 15,
  "active_students": 12,
  "average_completion": 35.5,
  "average_score": 78.2,
  "top_students": [
    {
      "student_id": 111,
      "name": "Петр Петров",
      "completion": 85,
      "avg_score": 92
    }
  ],
  "difficult_themes": [
    {
      "theme_id": 5,
      "theme_name": "Мышцы таза",
      "avg_score": 65,
      "attempts_avg": 2.3
    }
  ]
}
```

**Изменяемые сущности**: - (только SELECT с агрегацией)

---

## 7. Администрирование

**Примечание**: Административные функции реализуются через GoAdmin панель. API endpoints для администрирования имеют низкий приоритет, так как CRUD операции выполняются через веб-интерфейс.

### ✅ POST /api/v1/admin/promo-codes
**Описание**: Создание промокода администратором (RESTful)

**Приоритет**: 🟢 НИЗКИЙ

**Параметры запроса**:
```json
{
  "code": "ABC123",
  "university_name": "МГУ",
  "max_activations": 50,
  "expires_at": "2026-06-30T23:59:59Z"
}
```

**Ответ**:
```json
{
  "code": "ABC123",
  "university_name": "МГУ",
  "max_activations": 50,
  "remaining": 50,
  "status": "pending",
  "expires_at": "2026-06-30T23:59:59Z",
  "created_at": "2026-01-20T10:00:00Z"
}
```

**HTTP Status**: 201 Created

**Headers**: `Location: /api/v1/admin/promo-codes/ABC123`

**Изменяемые сущности**: `promo_codes` (INSERT)

**Обработка ошибок**:
- 409 Conflict - промокод с таким кодом уже существует
- 400 Bad Request - невалидные данные

---

### ✅ DELETE /api/v1/admin/promo-codes/{code}
**Описание**: Деактивация промокода (RESTful)

**Приоритет**: 🟢 НИЗКИЙ

**Ответ**:
```json
{
  "code": "ABC123",
  "status": "deactivated",
  "deactivated_at": "2026-01-20T10:30:00Z"
}
```

**HTTP Status**: 200 OK

**Изменяемые сущности**: `promo_codes` (UPDATE status='deactivated')

**Обработка ошибок**:
- 404 Not Found - промокод не найден

---

### ✅ POST /api/v1/admin/content/modules
**Описание**: Создание модуля (RESTful)

**Приоритет**: 🟢 НИЗКИЙ

**Параметры запроса**:
```json
{
  "name": "Анатомия мышц",
  "description": "Изучение мышечной системы человека",
  "order_num": 2,
  "is_locked": false
}
```

**Ответ**:
```json
{
  "id": 2,
  "name": "Анатомия мышц",
  "description": "Изучение мышечной системы человека",
  "order_num": 2,
  "is_locked": false,
  "created_at": "2026-01-20T11:00:00Z"
}
```

**HTTP Status**: 201 Created

**Headers**: `Location: /api/v1/admin/content/modules/2`

**Изменяемые сущности**: `modules` (INSERT)

---

### ✅ PUT /api/v1/admin/content/modules/{id}
**Описание**: Обновление модуля (RESTful, идемпотентный)

**Приоритет**: 🟢 НИЗКИЙ

**Параметры запроса**:
```json
{
  "name": "Анатомия мышц (обновлено)",
  "description": "Полное изучение мышечной системы",
  "order_num": 2,
  "is_locked": false
}
```

**Ответ**:
```json
{
  "id": 2,
  "name": "Анатомия мышц (обновлено)",
  "description": "Полное изучение мышечной системы",
  "order_num": 2,
  "is_locked": false,
  "updated_at": "2026-01-20T11:15:00Z"
}
```

**HTTP Status**: 200 OK

**Изменяемые сущности**: `modules` (UPDATE)

**Обработка ошибок**:
- 404 Not Found - модуль не найден

---

### ✅ POST /api/v1/admin/content/themes
**Описание**: Создание темы в модуле (RESTful)

**Приоритет**: 🟢 НИЗКИЙ

**Параметры запроса**:
```json
{
  "module_id": 1,
  "name": "Позвоночник",
  "description": "Изучение позвоночного столба",
  "order_num": 3,
  "is_introduction": false,
  "is_locked": false
}
```

**Ответ**:
```json
{
  "id": 3,
  "module_id": 1,
  "name": "Позвоночник",
  "description": "Изучение позвоночного столба",
  "order_num": 3,
  "is_introduction": false,
  "is_locked": false,
  "created_at": "2026-01-20T11:30:00Z"
}
```

**HTTP Status**: 201 Created

**Headers**: `Location: /api/v1/admin/content/themes/3`

**Изменяемые сущности**: `themes` (INSERT)

**Обработка ошибок**:
- 404 Not Found - модуль не найден
- 400 Bad Request - order_num=1 и is_introduction=false (первая тема должна быть введением)

---

### ✅ POST /api/v1/admin/content/mnemonics
**Описание**: Создание мнемоники для темы (RESTful)

**Приоритет**: 🟢 НИЗКИЙ

**Параметры запроса**:
```json
{
  "theme_id": 3,
  "type": "text" | "image",
  "content_text": "Запоминалка для позвоночника...",
  "s3_image_key": null,
  "order_num": 1
}
```

**Ответ**:
```json
{
  "id": 5,
  "theme_id": 3,
  "type": "text",
  "content_text": "Запоминалка для позвоночника...",
  "s3_image_url": null,
  "order_num": 1,
  "created_at": "2026-01-20T12:00:00Z"
}
```

**HTTP Status**: 201 Created

**Headers**: `Location: /api/v1/admin/content/mnemonics/5`

**Изменяемые сущности**: `mnemonics` (INSERT)

**Обработка ошибок**:
- 404 Not Found - тема не найдена
- 400 Bad Request - type=image, но s3_image_key отсутствует

---

### ✅ POST /api/v1/admin/content/tests
**Описание**: Создание теста для темы (RESTful)

**Приоритет**: 🟢 НИЗКИЙ

**Параметры запроса**:
```json
{
  "theme_id": 3,
  "difficulty": 3,
  "passing_score": 70,
  "questions": [
    {
      "question": "Сколько позвонков в шейном отделе?",
      "type": "multiple_choice",
      "options": ["5", "7", "9", "12"],
      "correct_answer": "7",
      "order_num": 1
    }
  ]
}
```

**Ответ**:
```json
{
  "id": 10,
  "theme_id": 3,
  "difficulty": 3,
  "passing_score": 70,
  "total_questions": 1,
  "created_at": "2026-01-20T12:30:00Z"
}
```

**HTTP Status**: 201 Created

**Headers**: `Location: /api/v1/admin/content/tests/10`

**Изменяемые сущности**: `tests` (INSERT)

**Обработка ошибок**:
- 404 Not Found - тема не найдена
- 400 Bad Request - questions пустой массив

---

### ✅ GET /api/v1/admin/users
**Описание**: Список всех пользователей с фильтрацией (RESTful)

**Приоритет**: 🟢 НИЗКИЙ

**Параметры запроса**:
```
?role=student|teacher
&subscription_status=active|inactive
&limit=50
&offset=0
```

**Ответ**:
```json
{
  "users": [
    {
      "telegram_id": 123456789,
      "first_name": "Иван",
      "last_name": "Иванов",
      "username": "ivan_ivanov",
      "role": "student",
      "subscription_status": "active",
      "university_code": "MGU",
      "created_at": "2025-09-15T10:00:00Z",
      "last_activity": "2026-01-20T09:00:00Z"
    }
  ],
  "total": 150,
  "limit": 50,
  "offset": 0
}
```

**HTTP Status**: 200 OK

**Изменяемые сущности**: - (только SELECT)

---

### ✅ GET /api/v1/admin/analytics/overview
**Описание**: Общая аналитика системы (RESTful)

**Приоритет**: 🟢 НИЗКИЙ

**Ответ**:
```json
{
  "total_users": 150,
  "active_subscriptions": 85,
  "total_students": 130,
  "total_teachers": 20,
  "total_modules": 5,
  "total_themes": 50,
  "total_tests_completed": 1250,
  "average_completion_rate": 35.5,
  "average_score": 78.2,
  "top_modules": [
    {
      "module_id": 1,
      "module_name": "Анатомия костей",
      "completions": 120,
      "avg_score": 82
    }
  ],
  "recent_activity": [
    {
      "user_id": 123456789,
      "activity_type": "test_completed",
      "theme_name": "Позвоночник",
      "score": 85,
      "timestamp": "2026-01-20T09:00:00Z"
    }
  ]
}
```

**HTTP Status**: 200 OK

**Изменяемые сущности**: - (только SELECT с агрегацией)

---

## Сводная таблица endpoints

### 1. Управление пользователями

| Endpoint | Метод | Статус | Приоритет | Сервис |
|----------|-------|--------|-----------|--------|
| `/api/v1/users` | POST | ✅ | 🔴 КРИТИЧЕСКИЙ | Users Service |
| `/api/v1/users/{telegram_id}` | PATCH | ✅ | 🔴 ВЫСОКИЙ | Users Service |
| `/api/v1/users/{user_id}/subscription` | GET | ✅ | 🔴 ВЫСОКИЙ | Users Service |

### 2. Контент

| Endpoint | Метод | Статус | Приоритет | Сервис |
|----------|-------|--------|-----------|--------|
| `/api/v1/content/modules` | GET | ✅ | 🔴 ВЫСОКИЙ | Content Service |
| `/api/v1/content/modules/{id}/themes` | GET | ✅ | 🔴 КРИТИЧЕСКИЙ | Content Service |
| `/api/v1/users/{user_id}/study-sessions` | POST | ✅ | 🔴 КРИТИЧЕСКИЙ | Content + Progress Service |
| `/api/v1/users/{user_id}/test-attempts` | POST | ✅ | 🔴 КРИТИЧЕСКИЙ | Content Service |
| `/api/v1/users/{user_id}/test-attempts/{attempt_id}` | PUT | ✅ | 🔴 КРИТИЧЕСКИЙ | Progress Service |
| `/api/v1/users/{user_id}/theme/{theme_id}/access` | GET | ✅ | 🟡 ВЫСОКИЙ | Progress Service |

### 3. Прогресс пользователей

| Endpoint | Метод | Статус | Приоритет | Сервис |
|----------|-------|--------|-----------|--------|
| `/api/v1/users/{user_id}/progress` | GET | ✅ | 🔴 ВЫСОКИЙ | Progress Service |
| `/api/v1/users/{user_id}/progress/modules/{module_id}` | GET | ✅ | 🟡 ВЫСОКИЙ | Progress Service |

### 4. Подписки и промокоды

| Endpoint | Метод | Статус | Приоритет | Сервис |
|----------|-------|--------|-----------|--------|
| `/api/v1/teachers/{teacher_id}/promo-codes` | POST | ✅ | 🔴 КРИТИЧЕСКИЙ | Subscriptions Service |
| `/api/v1/users/{user_id}/subscriptions` | POST | ✅ | 🔴 КРИТИЧЕСКИЙ | Subscriptions Service |
| `/api/v1/teachers/{teacher_id}/promo-codes` | GET | ✅ | 🟡 ВЫСОКИЙ | Subscriptions Service |

### 5. Платежи

| Endpoint | Метод | Статус | Приоритет | Сервис |
|----------|-------|--------|-----------|--------|
| `/api/v1/users/{user_id}/payment-invoices` | POST | ✅ | 🔴 КРИТИЧЕСКИЙ | Payments Service |
| `/api/v1/webhooks/payment-gateway` | POST | ✅ | 🔴 КРИТИЧЕСКИЙ | Payments Service |
| `/api/v1/users/{user_id}/payment-invoices/pending` | GET | ✅ | 🟠 СРЕДНИЙ | Payments Service |

### 6. Функционал преподавателя

| Endpoint | Метод | Статус | Приоритет | Сервис |
|----------|-------|--------|-----------|--------|
| `/api/v1/teachers/{teacher_id}/students` | GET | ✅ | 🔴 КРИТИЧЕСКИЙ | Teachers Service |
| `/api/v1/teachers/{teacher_id}/students/{student_id}/progress` | GET | ✅ | 🔴 КРИТИЧЕСКИЙ | Teachers Service |
| `/api/v1/teachers/{teacher_id}/statistics` | GET | ✅ | 🟠 СРЕДНИЙ | Teachers Service |

### 7. Администрирование

| Endpoint | Метод | Статус | Приоритет | Сервис |
|----------|-------|--------|-----------|--------|
| `/api/v1/admin/promo-codes` | POST | ✅ | 🟢 НИЗКИЙ | Admin Service |
| `/api/v1/admin/promo-codes/{code}` | DELETE | ✅ | 🟢 НИЗКИЙ | Admin Service |
| `/api/v1/admin/content/modules` | POST | ✅ | 🟢 НИЗКИЙ | Admin Service |
| `/api/v1/admin/content/modules/{id}` | PUT | ✅ | 🟢 НИЗКИЙ | Admin Service |
| `/api/v1/admin/content/themes` | POST | ✅ | 🟢 НИЗКИЙ | Admin Service |
| `/api/v1/admin/content/mnemonics` | POST | ✅ | 🟢 НИЗКИЙ | Admin Service |
| `/api/v1/admin/content/tests` | POST | ✅ | 🟢 НИЗКИЙ | Admin Service |
| `/api/v1/admin/users` | GET | ✅ | 🟢 НИЗКИЙ | Admin Service |
| `/api/v1/admin/analytics/overview` | GET | ✅ | 🟢 НИЗКИЙ | Admin Service |

---

**Итого**:
- ✅ Определено и реализовано: **29** endpoints
- 🔴 Критических: **11** endpoints
- 🟡 Высоких: **7** endpoints
- 🟠 Средних: **2** endpoints
- 🟢 Низких: **9** endpoints (admin)

---

**Дата обновления**: 2026-01-20
**Версия**: 2.2
**Статус**: ✅ Все endpoints определены и документированы
