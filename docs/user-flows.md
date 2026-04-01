# User Flows - Сценарии использования системы

**Версия**: 2.1
**Дата**: 19 января 2026
**Статус**: RESTful API (все endpoints обновлены к стандартам REST)

> Документ содержит детальное сопоставление пользовательских сценариев с командами Telegram-бота и RESTful backend endpoints.

## Changelog

- **v2.1** (19.01.2026):
  - ✅ Все endpoints обновлены к RESTful стандартам
  - ✅ Добавлена поддержка сущности "Введение" в user flows

- **v1.0** (28.12.2025):
  - Базовая спецификация user flows

## Содержание

- [Сценарий 1: Регистрация нового пользователя](#сценарий-1-регистрация-нового-пользователя)
- [Сценарий 2: Изучение темы и прохождение теста](#сценарий-2-изучение-темы-и-прохождение-теста)
- [Сценарий 3: Активация промокода преподавателем](#сценарий-3-активация-промокода-преподавателем)
- [Сценарий 4: Присоединение студента по промокоду](#сценарий-4-присоединение-студента-по-промокоду)
- [Сценарий 5: Активация личной подписки через оплату](#сценарий-5-активация-личной-подписки-через-оплату)
- [Сценарий 6: Просмотр личного прогресса студента](#сценарий-6-просмотр-личного-прогресса-студента)
- [Сценарий 7: Просмотр прогресса студентов преподавателем](#сценарий-7-просмотр-прогресса-студентов-преподавателем)
- [Сценарий 8: Переход к следующей теме/модулю](#сценарий-8-переход-к-следующей-темемодулю)

---

## Сценарий 1: Регистрация нового пользователя

**Описание**: Первое взаимодействие пользователя с ботом, создание профиля и выбор роли.

| Шаг | Действие пользователя | Команда/UI событие | Backend endpoint | Изменяемые сущности | Возвращаемый результат |
|-----|----------------------|-------------------|-----------------|---------------------|----------------------|
| 1 | Пользователь впервые запускает бота | `/start` | ✅ `POST /api/v1/users` | `users` (INSERT: telegram_id, role=student) | 201 Created: User ID, роль, статус подписки |
| 2 | Пользователь выбирает роль | Inline keyboard: "Я студент" / "Я преподаватель" | ✅ `PATCH /api/v1/users/{telegram_id}` | `users` (UPDATE role) | 200 OK: Подтверждение роли + переход к меню |
| 3 | Бот показывает приветственное сообщение и основные команды | - | - | - | Список команд для выбранной роли (student/teacher) |

### Обнаруженные пробелы

- ❌ **Критический**: Отсутствует endpoint `POST /api/v1/users/register` для создания нового пользователя
- ❌ **Критический**: Отсутствует endpoint `PATCH /api/v1/users/{telegram_id}/role` для обновления роли
- ❌ **Критический**: Отсутствует команда `/start` в спецификации команд бота
- ⚠️ Нет обработки случая, когда пользователь уже зарегистрирован (проверка существования)
- ⚠️ Нет валидации telegram_id

### Предложения по улучшению

1. Добавить endpoint `POST /api/v1/users/register`:
   - Параметры: `telegram_id`, `first_name`, `last_name`, `username`
   - Логика: проверка существования → создание записи → возврат user_id

2. Добавить endpoint `PATCH /api/v1/users/{telegram_id}/role`:
   - Параметры: `role` (enum: student, teacher)
   - Логика: валидация роли → обновление → возврат подтверждения

3. Добавить команду `/start` с интерактивным выбором роли

---

## Сценарий 2: Изучение темы и прохождение теста

**Описание**: Основной учебный сценарий - студент выбирает модуль, изучает тему, проходит тест.

| Шаг | Действие пользователя | Команда/UI событие | Backend endpoint | Изменяемые сущности | Возвращаемый результат |
|-----|----------------------|-------------------|-----------------|---------------------|----------------------|
| 1 | Пользователь запрашивает список модулей | `/modules` или `/start` | ✅ `GET /api/v1/content/modules` | - | Список модулей с названиями, описанием, прогрессом |
| 2 | Пользователь выбирает модуль | Inline keyboard: "Модуль 1" | ✅ `GET /api/v1/content/modules/{id}/themes` | - | Список тем (первая - "Введение") с is_introduction, доступностью |
| 3 | Пользователь выбирает тему | Inline button: "Тема X" | ✅ `POST /api/v1/users/{user_id}/study-sessions` | `user_progress` (UPDATE status=started, started_at) | 201 Created: Контент темы (мнемоники, is_introduction) |
| 4 | Backend проверяет доступ к теме | - | Внутри `POST /user/study-theme`: вызов `check_access(user_id)` | - | access_granted / access_denied (с указанием причины) |
| 5 | Пользователь изучает мнемоники | - (чтение контента) | - | - | - |
| 6 | Пользователь готов к тесту | `/test_ready theme=2` или Inline button: "Начать тест" | ✅ `POST /user/start-test` (user_id, theme_id) | - | Список вопросов теста (загружается из tests.questions_json) |
| 7 | Пользователь отвечает на вопросы | Inline keyboard с вариантами ответов (multiple choice) | - | - | - |
| 8 | Пользователь отправляет ответы | `/submit_answers answers=[...]` или final callback | ✅ `POST /user/submit-test` (user_id, theme_id, answers) | `user_progress` (обновление: score, attempts++, status=completed/failed, completed_at) | Результат: score, passed/failed, мотивационное сообщение, рекомендация следующей темы |
| 9 | При успехе: открытие следующей темы (для пользователей без подписки) | - (автоматически) | Внутри `POST /user/submit-test` | `user_progress` (создание записи для следующей темы с возможностью доступа) | - |

### Обнаруженные пробелы

- ❌ **Критический**: Отсутствует endpoint `GET /api/v1/content/modules/{id}/themes` для получения списка тем модуля
- ❌ **Высокий**: Отсутствует команда `/modules` в спецификации
- ❌ **Высокий**: Нет обработки случая истечения подписки МЕЖДУ началом изучения темы и прохождением теста
- ❌ **Высокий**: Нет ограничения количества попыток прохождения теста (отсутствует поле `max_attempts` в таблице `tests`)
- ❌ **Средний**: Нет таймаута для прохождения теста (отсутствует поле `time_limit_minutes`)
- ❌ **Средний**: Нет валидации полноты ответов (пользователь может отправить неполный набор ответов)
- ⚠️ В ответе `access_denied` не указывается, какую тему нужно пройти для разблокировки текущей
- ⚠️ Нет явной команды `/retry_test` для повторной попытки

### Предложения по улучшению

1. **Endpoint `GET /api/v1/content/modules/{id}/themes`**:
   - Параметры: `module_id`, `user_id` (для проверки доступа)
   - Возвращает: список тем с полями `id`, `name`, `order_num`, `is_accessible` (boolean)

2. **Ограничение попыток**:
   - Добавить поле `max_attempts` в таблицу `tests`
   - Добавить поле `current_attempt` в таблицу `user_progress`
   - Проверка в `POST /user/start-test`: если `current_attempt >= max_attempts` → возврат ошибки

3. **Таймаут теста**:
   - Добавить поле `time_limit_minutes` в таблицу `tests`
   - Добавить поле `test_started_at` в таблицу `user_progress`
   - Проверка в `POST /user/submit-test`: если `NOW() - test_started_at > time_limit` → тест не засчитывается

4. **Улучшенная проверка доступа**:
   - В ответе `access_denied` добавить поле `required_theme_id` (какую тему нужно пройти)
   - Добавить endpoint `GET /api/v1/users/{user_id}/theme/{theme_id}/access` для явной проверки доступа

5. **Команда `/retry_test`**:
   - Добавить команду `/retry_test theme_id` для явной пересдачи теста

---

## Сценарий 3: Активация промокода преподавателем

**Описание**: Преподаватель активирует промокод, полученный от администратора, для предоставления студентам доступа.

| Шаг | Действие пользователя | Команда/UI событие | Backend endpoint | Изменяемые сущности | Возвращаемый результат |
|-----|----------------------|-------------------|-----------------|---------------------|----------------------|
| 1 | Преподаватель запрашивает активацию кода | `/activate_code ABC123` | ⚠️ `POST /api/v1/teacher/activate-promo` (teacher_id, code) | `promo_codes` (обновление: teacher_id, activated_at=NOW(), remaining--) | Success: "Код активирован, передайте студентам: ABC123"; Error: "Код неверный/исчерпан" |
| 2 | Backend проверяет валидность кода | - | Внутри endpoint: `SELECT FROM promo_codes WHERE code="ABC123"` | - | promo_data (activations, status, expires_at) |
| 3 | Backend обновляет запись промокода | - | `UPDATE promo_codes SET teacher_id=?, activated_at=NOW(), remaining=remaining-1 WHERE code=?` | `promo_codes` | - |
| 4 | Бот возвращает подтверждение или ошибку | - | - | - | Сообщение с кодом для передачи студентам или описание ошибки |

### Обнаруженные пробелы

- ⚠️ **Критический**: Endpoint упомянут в container.puml как `POST /api/v1/subscriptions/validate-promo`, но логика не разделена на активацию преподавателем vs присоединение студентом
- ❌ **Высокий**: Отсутствует поле `expires_at` в таблице `promo_codes` для установки срока действия
- ❌ **Высокий**: Отсутствует поле `status` (ENUM: pending, active, expired, deactivated) в таблице `promo_codes`
- ❌ **Средний**: Нет проверки на повторную активацию преподавателем (если `teacher_id` уже заполнен)
- ⚠️ Нет endpoint для получения списка промокодов преподавателя (`GET /api/v1/teacher/promo-codes`)

### Предложения по улучшению

1. **Разделить endpoint на два**:
   - `POST /api/v1/teacher/activate-promo` - для преподавателей
   - `POST /api/v1/users/join-promo` - для студентов

2. **Добавить поля в `promo_codes`**:
   ```sql
   ALTER TABLE promo_codes ADD COLUMN expires_at TIMESTAMP;
   ALTER TABLE promo_codes ADD COLUMN status ENUM('pending', 'active', 'expired', 'deactivated');
   ALTER TABLE promo_codes ADD COLUMN created_by_admin_id BIGINT;
   ```

3. **Проверка повторной активации**:
   ```go
   if promo.TeacherID != nil {
       return errors.New("Промокод уже активирован преподавателем")
   }
   ```

4. **Endpoint для списка промокодов**:
   - `GET /api/v1/teacher/promo-codes` (teacher_id)
   - Возвращает: список промокодов с информацией об активациях

---

## Сценарий 4: Присоединение студента по промокоду

**Описание**: Студент вводит промокод, полученный от преподавателя, и получает доступ к материалам.

| Шаг | Действие пользователя | Команда/UI событие | Backend endpoint | Изменяемые сущности | Возвращаемый результат |
|-----|----------------------|-------------------|-----------------|---------------------|----------------------|
| 1 | Студент получает код от преподавателя | - (внешнее взаимодействие) | - | - | - |
| 2 | Студент вводит код в бота | `/join_university ABC123` или `/promo ABC123` | ❌ `POST /api/v1/users/join-promo` (student_id, code) | `subscriptions` (INSERT/UPDATE), `teacher_promo_students` (INSERT связи teacher-student), `users` (subscription_status=active) | Success: "Доступ открыт"; Error: "Код недействителен" |
| 3 | Backend проверяет код | - | `SELECT FROM promo_codes WHERE code="ABC123" AND status="ACTIVE"` | - | promo_data (teacher_id, remaining, expires_at) |
| 4 | Backend создает связь студент-преподаватель | - | `INSERT INTO teacher_promo_students (teacher_id, student_id, university_code, joined_at)` | `teacher_promo_students` | - |
| 5 | Backend открывает доступ студенту | - | `INSERT/UPDATE subscriptions (user_id, type=university, status=active, expires_at)` | `subscriptions` | - |
| 6 | Backend уменьшает счетчик промокода | - | `UPDATE promo_codes SET remaining=remaining-1 WHERE code=?` | `promo_codes` | - |
| 7 | Бот отправляет подтверждение | - | - | - | "Добро пожаловать! Доступ к материалам преподавателя открыт" |

### Обнаруженные пробелы

- ❌ **Критический**: Endpoint `POST /api/v1/users/join-promo` не упомянут в container.puml
- ❌ **Высокий**: Нет проверки на дубликаты (студент уже активировал этот промокод)
- ❌ **Высокий**: Нет UNIQUE constraint на `(teacher_id, student_id)` в таблице `teacher_promo_students`
- ❌ **Высокий**: Нет проверки срока действия промокода (отсутствует поле `expires_at`)
- ❌ **Средний**: Нет обработки случая `remaining=0` (промокод исчерпан)
- ⚠️ Нет информативного сообщения об ошибке для пользователя

### Предложения по улучшению

1. **Endpoint `POST /api/v1/users/join-promo`**:
   - Параметры: `user_id`, `code`
   - Логика:
     ```go
     1. SELECT promo_code WHERE code=? AND status='active' AND expires_at > NOW() AND remaining > 0
     2. CHECK EXISTS teacher_promo_students WHERE teacher_id=? AND student_id=?
     3. INSERT teacher_promo_students
     4. INSERT/UPDATE subscriptions
     5. UPDATE promo_codes SET remaining=remaining-1
     ```

2. **Добавить UNIQUE constraint**:
   ```sql
   ALTER TABLE teacher_promo_students
   ADD CONSTRAINT unique_teacher_student UNIQUE (teacher_id, student_id);
   ```

3. **Обработка ошибок**:
   - `remaining=0` → "Промокод исчерпан. Обратитесь к преподавателю"
   - `expires_at < NOW()` → "Промокод истек"
   - `duplicate` → "Вы уже активировали этот промокод"

4. **Добавить поле в `teacher_promo_students`**:
   ```sql
   ALTER TABLE teacher_promo_students ADD COLUMN promo_code VARCHAR(255);
   ```

---

## Сценарий 5: Активация личной подписки через оплату

**Описание**: Студент оформляет персональную подписку через платежный шлюз.

| Шаг | Действие пользователя | Команда/UI событие | Backend endpoint | Изменяемые сущности | Возвращаемый результат |
|-----|----------------------|-------------------|-----------------|---------------------|----------------------|
| 1 | Студент запрашивает подписку | `/subscribe` или Inline button: "Оформить подписку" | ❌ `POST /api/v1/subscriptions/create-invoice` (user_id, plan) | - | Invoice URL (ссылка на платежный шлюз) |
| 2 | Backend создает счет в Payment Gateway | - | Вызов Payment Gateway API: `create_invoice(amount, user_id, plan)` | - | payment_id, payment_url |
| 3 | Backend сохраняет pending payment | - | `UPDATE users SET pending_payment_id=? WHERE telegram_id=?` | `users` (pending_payment_id) | - |
| 4 | Пользователь переходит по ссылке и оплачивает | Переход по ссылке, оплата в платежной системе | - | - | - |
| 5 | Payment Gateway отправляет webhook о успешной оплате | - (внешний вызов) | ❌ `POST /api/v1/webhooks/payment-callback` (payment_id, status, user_id, metadata) | `subscriptions` (INSERT: payment_id, user_id, status=active, type=personal, expires_at), `users` (subscription_status=active, pending_payment_id=NULL) | HTTP 200 OK (для Payment Gateway) |
| 6 | Backend проверяет идемпотентность | - | `SELECT FROM subscriptions WHERE payment_id=?` | - | Если существует → возврат 200 без изменений |
| 7 | Бот отправляет подтверждение пользователю | - | Telegram API: `send_message(user_id, "Подписка активирована!")` | - | "Подписка активирована! Доступ ко всем материалам открыт" |

### Дополнительный сценарий: Проверка статуса платежа

| Шаг | Действие пользователя | Команда/UI событие | Backend endpoint | Изменяемые сущности | Возвращаемый результат |
|-----|----------------------|-------------------|-----------------|---------------------|----------------------|
| 1 | Студент проверяет статус платежа | `/check_payment` или Inline button: "Проверить статус оплаты" | ❌ `GET /api/v1/subscriptions/{user_id}/check-payment` | - | Статус: "Оплата получена" / "Ожидание оплаты" / "Оплата отклонена" |

### Обнаруженные пробелы

- ❌ **Критический**: Отсутствует endpoint `POST /api/v1/subscriptions/create-invoice` для создания счета
- ❌ **Критический**: Отсутствует endpoint `POST /api/v1/webhooks/payment-callback` для обработки webhook
- ❌ **Критический**: Отсутствует поле `pending_payment_id` в таблице `users`
- ❌ **Высокий**: Нет обработки webhook с `status=failed` (платеж отклонен)
- ❌ **Высокий**: Нет идемпотентности webhook (дублирование вызовов)
- ❌ **Средний**: Нет endpoint для проверки статуса платежа
- ❌ **Средний**: Нет обработки случая, когда пользователь закрыл страницу оплаты
- ⚠️ Нет проверки активной подписки перед созданием нового счета

### Предложения по улучшению

1. **Endpoint `POST /api/v1/subscriptions/create-invoice`**:
   ```go
   // Параметры: user_id, plan (monthly/yearly)
   1. SELECT subscriptions WHERE user_id AND status='active' AND expires_at > NOW()
   2. IF active_subscription EXISTS THEN return error
   3. Call PaymentGateway.CreateInvoice(amount, user_id, plan)
   4. UPDATE users SET pending_payment_id=payment_id
   5. RETURN payment_url
   ```

2. **Endpoint `POST /api/v1/webhooks/payment-callback`**:
   ```go
   // Параметры: payment_id, status, signature
   1. Verify webhook signature
   2. SELECT subscriptions WHERE payment_id
   3. IF EXISTS THEN return 200 (идемпотентность)
   4. IF status='success':
      INSERT subscriptions (user_id, payment_id, status='active', expires_at=NOW()+30days)
      UPDATE users SET subscription_status='active', pending_payment_id=NULL
      Send Telegram notification
   5. IF status='failed':
      UPDATE users SET pending_payment_id=NULL
      Send Telegram notification "Оплата отклонена"
   6. RETURN 200
   ```

3. **Endpoint `GET /api/v1/subscriptions/{user_id}/check-payment`**:
   ```go
   1. SELECT users WHERE telegram_id=user_id
   2. IF pending_payment_id IS NULL THEN return "Нет ожидаемых платежей"
   3. SELECT subscriptions WHERE payment_id=pending_payment_id
   4. IF EXISTS THEN return "Оплата получена"
   5. Call PaymentGateway.GetPaymentStatus(pending_payment_id)
   6. RETURN status
   ```

4. **Добавить поле в таблицу `users`**:
   ```sql
   ALTER TABLE users ADD COLUMN pending_payment_id VARCHAR(255);
   ```

5. **Добавить поля в таблицу `subscriptions`**:
   ```sql
   ALTER TABLE subscriptions ADD COLUMN auto_renew BOOLEAN DEFAULT false;
   ALTER TABLE subscriptions ADD COLUMN cancelled_at TIMESTAMP;
   ```

---

## Сценарий 6: Просмотр личного прогресса студента

**Описание**: Студент просматривает свой прогресс обучения.

| Шаг | Действие пользователя | Команда/UI событие | Backend endpoint | Изменяемые сущности | Возвращаемый результат |
|-----|----------------------|-------------------|-----------------|---------------------|----------------------|
| 1 | Студент запрашивает свой прогресс | `/progress` или `/my_progress` | ✅ `GET /api/v1/progress/stats` (user_id) | - | Процент прохождения модулей/тем, список завершенных тем, средний балл, количество попыток |
| 2 | Студент выбирает детализацию по модулю | Inline keyboard: "Модуль 1" | ❌ `GET /api/v1/progress/module/{module_id}` (user_id) | - | Список тем модуля с прогрессом: статус (пройдена/не пройдена/начата), баллы, количество попыток |

### Обнаруженные пробелы

- ✅ Endpoint `GET /api/v1/progress/stats` ОПРЕДЕЛЕН в container.puml
- ❌ **Высокий**: Отсутствует endpoint `GET /api/v1/progress/module/{module_id}` для детализации по модулю
- ⚠️ Нет визуализации прогресса (например, progress bar)
- ⚠️ Нет сравнения с другими студентами (средний балл по курсу)

### Предложения по улучшению

1. **Endpoint `GET /api/v1/progress/module/{module_id}`**:
   ```go
   // Параметры: user_id, module_id
   1. SELECT themes WHERE module_id=? ORDER BY order_num
   2. LEFT JOIN user_progress ON theme_id
   3. RETURN [
        {theme_id, name, status, score, attempts, completed_at}
      ]
   ```

2. **Дополнительные метрики в `GET /api/v1/progress/stats`**:
   - Добавить: общее время обучения, streak (дни подряд), рейтинг среди студентов

3. **Endpoint для экспорта прогресса** (низкий приоритет):
   - `GET /api/v1/users/{user_id}/progress/export?format=pdf`

---

## Сценарий 7: Просмотр прогресса студентов преподавателем

**Описание**: Преподаватель просматривает список своих студентов и их прогресс.

| Шаг | Действие пользователя | Команда/UI событие | Backend endpoint | Изменяемые сущности | Возвращаемый результат |
|-----|----------------------|-------------------|-----------------|---------------------|----------------------|
| 1 | Преподаватель запрашивает список своих студентов | `/my_students` | ❌ `GET /api/v1/teacher/students` (teacher_id) | - | Список студентов: имя, дата присоединения, общий прогресс (% завершения), средний балл |
| 2 | Преподаватель выбирает студента | Inline keyboard: "Студент Иван" | ❌ `GET /api/v1/teacher/students/{student_id}/progress` (teacher_id, student_id) | - | Детальный прогресс студента: модули, темы, баллы, попытки, даты завершения |
| 3 | Преподаватель просматривает статистику по группе | `/group_stats` | ❌ `GET /api/v1/teacher/group-stats` (teacher_id) | - | Средние показатели группы, топ-5 студентов, проблемные темы (низкий средний балл) |

### Обнаруженные пробелы

- ❌ **Критический**: Все endpoints для функционала преподавателя НЕ ОПРЕДЕЛЕНЫ
- ❌ **Критический**: Нет проверки связи teacher-student при просмотре прогресса (уязвимость безопасности)
- ❌ **Высокий**: Нет команды `/my_students`
- ❌ **Средний**: Нет команды `/group_stats`
- ⚠️ Нет обработки случая, когда у преподавателя нет студентов
- ⚠️ Нет обработки случая, когда студент еще не начал обучение (пустой прогресс)

### Предложения по улучшению

1. **Endpoint `GET /api/v1/teacher/students`**:
   ```go
   // Параметры: teacher_id
   1. SELECT users JOIN teacher_promo_students ON student_id WHERE teacher_id=?
   2. LEFT JOIN user_progress для подсчета прогресса
   3. RETURN [
        {student_id, name, joined_at, total_themes, completed_themes, avg_score}
      ]
   ```

2. **Endpoint `GET /api/v1/teacher/students/{student_id}/progress`**:
   ```go
   // Параметры: teacher_id, student_id
   1. Verify teacher-student relationship:
      SELECT FROM teacher_promo_students WHERE teacher_id=? AND student_id=?
   2. IF NOT EXISTS THEN return 403 Forbidden
   3. SELECT user_progress JOIN themes JOIN modules WHERE user_id=?
   4. IF empty THEN return {message: "Студент еще не начал обучение"}
   5. RETURN detailed_progress
   ```

3. **Endpoint `GET /api/v1/teacher/group-stats`**:
   ```go
   // Параметры: teacher_id
   1. SELECT all students of teacher
   2. Aggregate statistics:
      - avg_score по всем студентам
      - completion_rate
      - most_difficult_themes (WHERE avg_score < 70)
      - top_students (ORDER BY avg_score DESC LIMIT 5)
   3. RETURN group_statistics
   ```

4. **Добавить команды бота**:
   - `/my_students` → показать список студентов
   - `/group_stats` → показать статистику группы

---

## Сценарий 8: Переход к следующей теме/модулю

**Описание**: После успешного прохождения теста студент переходит к следующей теме или модулю.

| Шаг | Действие пользователя | Команда/UI событие | Backend endpoint | Изменяемые сущности | Возвращаемый результат |
|-----|----------------------|-------------------|-----------------|---------------------|----------------------|
| 1 | Студент завершает тест текущей темы | - (часть сценария 2, шаг 8) | ✅ `POST /user/submit-test` | `user_progress` (status, score, attempts) | Результат + рекомендация следующей темы |
| 2 | Backend проверяет: это была последняя тема модуля? | - (внутренняя логика) | `SELECT themes WHERE module_id=? AND order_num > current_order` | - | next_theme_id OR module_completed=true |
| 3 | Если последняя тема: открытие следующего модуля | - | `UPDATE` для разблокировки первой темы следующего модуля | `user_progress` (создание записи для первой темы следующего модуля) | Мотивационное сообщение о завершении модуля |
| 4 | Если не последняя: открытие следующей темы | - | `UPDATE` для разблокировки следующей темы | `user_progress` | - |
| 5 | Бот показывает кнопку перехода | Inline button: "Перейти к теме X" или "Начать модуль Y" | - | - | - |
| 6 | Студент переходит к следующей теме | Нажатие на кнопку | ✅ `POST /user/study-theme` (новая тема) | `user_progress` (status=started для новой темы) | Контент следующей темы |

### Обнаруженные пробелы

- ✅ Основная логика описана в architecture.md и sequence диаграммах
- ❌ **Средний**: Нет отдельного endpoint для явной проверки доступа к следующей теме
- ⚠️ Нет обработки случая, когда пользователь хочет перейти к теме вне последовательности (для пользователей с подпиской)
- ⚠️ Нет визуализации "дерева" прогресса (какие темы доступны, какие заблокированы)

### Предложения по улучшению

1. **Endpoint для проверки доступа** (опционально):
   ```go
   GET /api/v1/users/{user_id}/theme/{theme_id}/access
   // RETURN: {accessible: true/false, reason: "subscription_required" | "previous_theme_required", required_theme_id: 123}
   ```

2. **Улучшенный ответ от `POST /user/submit-test`**:
   ```json
   {
     "score": 85,
     "passed": true,
     "next_action": {
       "type": "next_theme" | "next_module" | "course_completed",
       "theme_id": 5,
       "module_id": 2,
       "message": "Отличный результат! Переходите к следующей теме."
     }
   }
   ```

3. **Добавить поля `is_locked` в таблицы**:
   ```sql
   ALTER TABLE modules ADD COLUMN is_locked BOOLEAN DEFAULT false;
   ALTER TABLE themes ADD COLUMN is_locked BOOLEAN DEFAULT false;
   ```

---

## Общие выводы по сценариям

### Полностью определенные сценарии
- ✅ Изучение темы (частично - основные endpoints существуют)
- ✅ Прохождение теста (частично)

### Частично определенные сценарии
- ⚠️ Активация промокода (endpoint существует, но логика не разделена)
- ⚠️ Просмотр прогресса (базовый endpoint существует)

### Полностью неопределенные сценарии
- ❌ Регистрация пользователя
- ❌ Присоединение по промокоду
- ❌ Оплата подписки
- ❌ Функционал преподавателя

### Критические приоритеты

1. **Фаза 1** (критические endpoints для работоспособности):
   - Регистрация пользователей
   - Навигация по контенту (список тем модуля)
   - Система промокодов (разделение на teacher/student)
   - Оплата подписки (invoice + webhook)
   - Функционал преподавателя (список студентов, просмотр прогресса)

2. **Фаза 2** (улучшение UX и надежности):
   - Проверка доступа к темам
   - Ограничение попыток тестов
   - Детализация прогресса
   - Проверка статуса платежа

3. **Фаза 3** (дополнительный функционал):
   - Персонализация (settings)
   - Аналитика (group stats)
   - Экспорт данных

---

**Дата создания**: 2025-12-28
**Версия**: 1.0
**Статус**: Требует реализации
