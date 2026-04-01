-- Development seed data
-- Run after migrations: goose -dir database/migrations sqlite3 mnemo_dev.db up

-- Modules
INSERT INTO modules (id, name, description, order_num, is_locked, icon_emoji) VALUES
    (1, 'Остеология', 'Изучение костной системы человека', 1, 0, '🦴'),
    (2, 'Миология', 'Изучение мышечной системы человека', 2, 1, '💪'),
    (3, 'Спланхнология', 'Изучение внутренних органов', 3, 1, '🫁');

-- Themes for Module 1 (Остеология)
INSERT INTO themes (id, module_id, name, description, order_num, is_introduction, estimated_time_minutes) VALUES
    (1, 1, 'Введение в остеологию', 'Ключевые термины и основные понятия', 1, 1, 10),
    (2, 1, 'Строение черепа', 'Кости черепа и их соединения', 2, 0, 20),
    (3, 1, 'Позвоночный столб', 'Позвонки, отделы позвоночника', 3, 0, 25),
    (4, 1, 'Грудная клетка', 'Рёбра, грудина, грудной отдел', 4, 0, 20);

-- Themes for Module 2 (Миология) - first theme is introduction
INSERT INTO themes (id, module_id, name, description, order_num, is_introduction) VALUES
    (5, 2, 'Введение в миологию', 'Основные группы мышц', 1, 1),
    (6, 2, 'Мышцы головы и шеи', 'Жевательные и мимические мышцы', 2, 0);

-- Mnemonics for Theme 1 (Introduction)
INSERT INTO mnemonics (theme_id, type, content_text, order_num) VALUES
    (1, 'text', 'Костей в теле взрослого человека — 206. Запомни: 2+0+6 = 8 — столько отверстий в человеческом черепе!', 1),
    (1, 'text', 'Кость (os) → остеология → наука об ossibus (костях). Латинский корень os/ossis — ключ ко всем терминам.', 2);

-- Mnemonics for Theme 2 (Череп)
INSERT INTO mnemonics (theme_id, type, content_text, order_num) VALUES
    (2, 'text', 'Кости мозгового черепа (8): «Лобная Теменная пара, Затылочная, пара Височных, Клиновидная, Решётчатая» — Л-Т-З-В-К-Р', 1),
    (2, 'text', 'Кости лицевого черепа (14): парные — верхняя челюсть, скуловая, носовая, слёзная, нёбная, нижняя носовая раковина; непарные — нижняя челюсть, сошник, подъязычная.', 2);

-- Test for Theme 1 (Introduction)
INSERT INTO tests (id, theme_id, questions_json, difficulty, passing_score) VALUES
(1, 1, '[
  {"id":1,"text":"Сколько костей в скелете взрослого человека?","type":"multiple_choice","options":["156","206","256","306"],"correct_answer":"206","order_num":1},
  {"id":2,"text":"Что изучает наука остеология?","type":"multiple_choice","options":["Мышцы","Кости","Суставы","Связки"],"correct_answer":"Кости","order_num":2},
  {"id":3,"text":"Губчатое вещество кости содержит костный мозг","type":"true_false","options":["true","false"],"correct_answer":"true","order_num":3},
  {"id":4,"text":"Надкостница (periosteum) покрывает суставные поверхности","type":"true_false","options":["true","false"],"correct_answer":"false","order_num":4},
  {"id":5,"text":"Какая кость является самой длинной в теле человека?","type":"multiple_choice","options":["Плечевая","Бедренная","Большеберцовая","Малоберцовая"],"correct_answer":"Бедренная","order_num":5}
]', 1, 70);

-- Test for Theme 2 (Череп)
INSERT INTO tests (id, theme_id, questions_json, difficulty, passing_score) VALUES
(2, 2, '[
  {"id":1,"text":"Сколько костей образуют мозговой череп?","type":"multiple_choice","options":["6","8","10","12"],"correct_answer":"8","order_num":1},
  {"id":2,"text":"Какая кость черепа является непарной?","type":"multiple_choice","options":["Теменная","Височная","Лобная","Носовая"],"correct_answer":"Лобная","order_num":2},
  {"id":3,"text":"Нижняя челюсть — единственная подвижная кость черепа","type":"true_false","options":["true","false"],"correct_answer":"true","order_num":3}
]', 2, 70);

-- Promo codes (admin-created, pending teacher activation)
INSERT INTO promo_codes (code, university_name, max_activations, remaining, status, expires_at) VALUES
    ('TEST2025', 'Тестовый университет', 50, 50, 'pending', '2026-12-31 23:59:59'),
    ('DEMO2025', 'Демо-университет', 30, 30, 'pending', '2026-06-30 23:59:59');

-- Demo users
INSERT INTO users (telegram_id, username, role, subscription_status) VALUES
    (100000001, 'ivan_student', 'student', 'inactive'),
    (100000002, 'maria_teacher', 'teacher', 'inactive'),
    (100000003, 'alex_subscribed', 'student', 'active');
