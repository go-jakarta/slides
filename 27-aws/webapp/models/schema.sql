DROP TABLE IF EXISTS books CASCADE;
DROP TYPE IF EXISTS book_type CASCADE;
DROP TABLE IF EXISTS authors CASCADE;

CREATE TABLE authors (
  author_id SERIAL PRIMARY KEY,
  name text NOT NULL DEFAULT ''
);

CREATE INDEX authors_name_idx ON authors(name);

CREATE TYPE book_type AS ENUM (
  'FICTION',
  'NONFICTION'
);

--

CREATE TABLE books (
  book_id SERIAL PRIMARY KEY,
  author_id integer NOT NULL REFERENCES authors(author_id),
  isbn text NOT NULL DEFAULT '' UNIQUE,
  booktype book_type NOT NULL DEFAULT 'FICTION',
  title text NOT NULL DEFAULT '',
  year integer NOT NULL DEFAULT 2000,
  available timestamp with time zone NOT NULL DEFAULT 'NOW()',
  tags varchar[] NOT NULL DEFAULT '{}'
);

CREATE INDEX books_title_year_idx ON books(title, year);
