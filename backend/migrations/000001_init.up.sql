-- Initial database schema
CREATE TABLE IF NOT EXISTS users (id UUID PRIMARY KEY, phone VARCHAR(20) UNIQUE);
