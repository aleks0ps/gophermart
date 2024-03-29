DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS balance;
DROP TABLE IF EXISTS users;

CREATE TABLE IF NOT EXISTS users (
  id BIGSERIAL PRIMARY KEY,
  login TEXT NOT NULL,
  password TEXT NOT NULL
);

CREATE UNIQUE INDEX unique_users ON users (login) NULLS NOT DISTINCT;

CREATE TABLE IF NOT EXISTS orders (
  id BIGSERIAL PRIMARY KEY,
  login TEXT NOT NULL,
  order_number TEXT NOT NULL,
  uploaded_at TEXT NOT NULL,
  CONSTRAINT fk_login
    FOREIGN KEY(login)
      REFERENCES users(login)
);	

CREATE UNIQUE INDEX unique_order ON orders (order_number) NULLS NOT DISTINCT;

CREATE TABLE IF NOT EXISTS balance (
  id BIGSERIAL PRIMARY KEY,
  login TEXT NOT NULL,
  current float4,
  withdrawn float4,
  CONSTRAINT fk_login
    FOREIGN KEY(login)
      REFERENCES users(login)
);

CREATE UNIQUE INDEX unique_users ON balance (login) NULLS NOT DISTINCT;
