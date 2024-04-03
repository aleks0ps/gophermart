DROP TABLE IF EXISTS orders CASCADE;
DROP TABLE IF EXISTS balance CASCADE;
DROP TABLE IF EXISTS users CASCADE;

CREATE TABLE IF NOT EXISTS users (
  id BIGSERIAL PRIMARY KEY,
  login TEXT NOT NULL,
  password TEXT NOT NULL
);

CREATE UNIQUE INDEX users_uniq_login ON users (login) NULLS NOT DISTINCT;

CREATE TABLE IF NOT EXISTS orders (
  id BIGSERIAL PRIMARY KEY,
  login TEXT NOT NULL,
  order_number TEXT NOT NULL,
  withdrawn float4 DEFAULT 0,
  uploaded_at TEXT NOT NULL,
  CONSTRAINT fk_login
    FOREIGN KEY(login)
      REFERENCES users(login)
      ON DELETE CASCADE
);	

CREATE UNIQUE INDEX orders_uniq_order ON orders (order_number) NULLS NOT DISTINCT;

CREATE TABLE IF NOT EXISTS balance (
  id BIGSERIAL PRIMARY KEY,
  login TEXT NOT NULL,
  current float4 DEFAULT 0,
  CONSTRAINT fk_login
    FOREIGN KEY(login)
      REFERENCES users(login)
      ON DELETE CASCADE 
);

CREATE UNIQUE INDEX balance_uniq_login ON balance (login) NULLS NOT DISTINCT;
