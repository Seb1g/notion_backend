-- Anemone Holding
-- Создание таблицы пользователей
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Создание таблицы групп закладок
CREATE TABLE IF NOT EXISTS notes_folder (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Создание таблицы страниц
CREATE TABLE IF NOT EXISTS pages (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    content TEXT,
    is_deleted BOOLEAN DEFAULT false,
    folder_id INT REFERENCES notes_folder (id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Создание таблицы токенов
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Индекс для быстрого поиска страниц по пользователю
CREATE INDEX IF NOT EXISTS idx_pages_user_id ON pages (user_id);

-- Anemone Mail
-- Таблица для хранения сгенерированных временных адресов
CREATE TABLE temp_addresses (
    id SERIAL PRIMARY KEY,
    address VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ
);

-- Таблица для хранения писем, связанная с временным адресом
CREATE TABLE emails (
    id SERIAL PRIMARY KEY,
    address_id INTEGER NOT NULL REFERENCES temp_addresses (id) ON DELETE CASCADE,
    sender VARCHAR(255) NOT NULL,
    recipients TEXT [] NOT NULL,
    subject TEXT,
    body TEXT,
    raw_data BYTEA,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Индекс для быстрого поиска писем по адресу
CREATE INDEX idx_emails_address_id ON emails (address_id);

--Anemone Trello
-- Таблица для досок (boards)
CREATE TABLE boards (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(), 
    title      VARCHAR(255) NOT NULL,
    user_id    INT NOT NULL REFERENCES users (id) ON DELETE CASCADE, 
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE
);

-- Индекс для быстрого поиска досок по пользователю
CREATE INDEX idx_boards_user_id ON boards (user_id);

-- Таблица для колонок (columns)
CREATE TABLE columns (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    column_title  VARCHAR(255) NOT NULL,
    board_id      UUID NOT NULL,
    position      INTEGER NOT NULL,

FOREIGN KEY (board_id) REFERENCES boards (id) ON DELETE CASCADE,

UNIQUE (board_id, position) );

-- Индекс для быстрого поиска колонок по доске
CREATE INDEX idx_columns_board_id ON columns (board_id);

-- Таблица для карточек (cards)
CREATE TABLE cards (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content       TEXT NOT NULL,
    column_id     UUID NOT NULL,
    position      INTEGER NOT NULL,

FOREIGN KEY (column_id) REFERENCES columns (id) ON DELETE CASCADE,

UNIQUE (column_id, position) );

-- Индекс для быстрого поиска карточек по колонке
CREATE INDEX idx_cards_column_id ON cards (column_id);