CREATE TABLE users (
  id bigserial not null primary key,
  login text not null unique,
  name text not null,
  encrypted_password text not null
);
-- ID встроенного админа 1
ALTER SEQUENCE users_id_seq RESTART WITH 2;