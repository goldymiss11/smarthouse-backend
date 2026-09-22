-- SmartHome MAX — полная схема базы данных
-- Запускать в Supabase SQL Editor

-- 1. Пользователи
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    vk_id VARCHAR(64) UNIQUE NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'resident', -- resident | uk
    created_at TIMESTAMP DEFAULT NOW()
);

-- 2. Адреса домов (ЖК)
CREATE TABLE IF NOT EXISTS addresses (
    id SERIAL PRIMARY KEY,
    full_address TEXT NOT NULL,
    lat DOUBLE PRECISION DEFAULT 0,
    lng DOUBLE PRECISION DEFAULT 0
);

-- 3. Привязка пользователей к адресам
CREATE TABLE IF NOT EXISTS user_addresses (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    address_id INTEGER REFERENCES addresses(id) ON DELETE CASCADE,
    apartment VARCHAR(20),
    account_number VARCHAR(30),
    UNIQUE(user_id, address_id)
);

-- 4. Заявки от жителей
CREATE TABLE IF NOT EXISTS requests (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    address_id INTEGER REFERENCES addresses(id) ON DELETE SET NULL,
    type VARCHAR(30) DEFAULT 'other', -- water | electricity | other | critical
    title TEXT,
    description TEXT,
    start_date VARCHAR(30),
    end_date VARCHAR(30),
    status VARCHAR(20) DEFAULT 'pending', -- pending | approved | rejected | resolved
    photo_url TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 5. Счета ЖКУ
CREATE TABLE IF NOT EXISTS bills (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    month VARCHAR(20) NOT NULL,
    amount NUMERIC(10,2) NOT NULL DEFAULT 0,
    is_paid BOOLEAN DEFAULT FALSE
);

-- 6. Показания счётчиков
CREATE TABLE IF NOT EXISTS meters (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    water NUMERIC(10,2) DEFAULT 0,
    electricity NUMERIC(10,2) DEFAULT 0,
    submitted_at TIMESTAMP DEFAULT NOW()
);

-- 7. Лента новостей / уведомлений
CREATE TABLE IF NOT EXISTS feed_items (
    id SERIAL PRIMARY KEY,
    address_id INTEGER REFERENCES addresses(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    body TEXT,
    category VARCHAR(30) DEFAULT 'info', -- info | water | electricity | emergency
    created_at TIMESTAMP DEFAULT NOW()
);

-- 8. Партнёры
CREATE TABLE IF NOT EXISTS partners (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL
);

-- 9. Офферы (привилегии)
CREATE TABLE IF NOT EXISTS offers (
    id SERIAL PRIMARY KEY,
    partner_id INTEGER REFERENCES partners(id) ON DELETE CASCADE,
    badge_text VARCHAR(50),
    offer_text TEXT,
    icon_url TEXT DEFAULT '',
    action_type VARCHAR(30) DEFAULT 'link',
    action_data TEXT DEFAULT '',
    requires_zero_debt BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE
);

-- Индексы для производительности
CREATE INDEX IF NOT EXISTS idx_requests_status ON requests(status);
CREATE INDEX IF NOT EXISTS idx_requests_address ON requests(address_id);
CREATE INDEX IF NOT EXISTS idx_bills_user ON bills(user_id);
CREATE INDEX IF NOT EXISTS idx_user_addresses_user ON user_addresses(user_id);
CREATE INDEX IF NOT EXISTS idx_user_addresses_address ON user_addresses(address_id);
CREATE INDEX IF NOT EXISTS idx_feed_items_address ON feed_items(address_id);
