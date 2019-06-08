-- CREATE DATABASE park_forum WITH OWNER = postgres;

CREATE EXTENSION IF NOT EXISTS CITEXT;

DROP TABLE    IF EXISTS vote             CASCADE;
DROP TABLE    IF EXISTS post             CASCADE;
DROP TABLE    IF EXISTS thread           CASCADE;
DROP TABLE    IF EXISTS forum            CASCADE;
DROP TABLE    IF EXISTS profile          CASCADE;
DROP TRIGGER  IF EXISTS t_vote           ON vote;
DROP TRIGGER  IF EXISTS t_post           ON post;
DROP FUNCTION IF EXISTS add_to_thread();
DROP FUNCTION IF EXISTS add_path();
-- DROP TABLE IF EXISTS  CASCADE;

-- CREATE ROLE park_forum;

CREATE TABLE IF NOT EXISTS profile
(
  uid       SERIAL                                              PRIMARY KEY,
  nickname  CITEXT       UNIQUE NOT NULL CHECK (nickname <> ''),
  full_name VARCHAR(128)        NOT NULL CHECK (nickname <> ''),
  about     VARCHAR(512)                                         DEFAULT '',
  email     VARCHAR(256) UNIQUE NOT NULL CHECK (email <> '')
--   email     CITEXT UNIQUE       NOT NULL CHECK (email <> '')
);

CREATE TABLE IF NOT EXISTS forum
(
  uid       SERIAL                                          PRIMARY KEY,
  title     VARCHAR(128)        NOT NULL CHECK ( title <> '' ),
  author_id INT                 NOT NULL,
--   slug      CITEXT       UNIQUE NOT NULL,
  slug      VARCHAR(256) UNIQUE NOT NULL,
  FOREIGN   KEY (author_id) REFERENCES profile (uid)
);

CREATE TABLE IF NOT EXISTS thread
(
  uid      SERIAL                                            PRIMARY KEY,
  user_id  INT                         NOT NULL,
  forum_id INT                         NOT NULL,
  title    VARCHAR(128)                NOT NULL CHECK ( title <> '' ),
--   slug     CITEXT       UNIQUE         NOT NULL,
  slug     VARCHAR(128) UNIQUE         NOT NULL,
  message  VARCHAR(2048)               NOT NULL,
  created  TIMESTAMP(3) WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
  votes    INT                         NOT NULL DEFAULT 0,

  FOREIGN  KEY (user_id)  REFERENCES profile (uid),
  FOREIGN  KEY (forum_id) REFERENCES forum   (uid)
);

CREATE TABLE IF NOT EXISTS post
(
  uid       SERIAL                                          PRIMARY KEY,
  parent_id INT                          NOT NULL DEFAULT 0,
  path      INTEGER[]                    NOT NULL,
  forum_id  INT                          NOT NULL,
  user_id   INT                          NOT NULL,
  thread_id INT                          NOT NULL,
  message   VARCHAR(2048)                NOT NULL,
  is_edited BOOLEAN                               DEFAULT FALSE,
  created   TIMESTAMP(3)  WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
  FOREIGN   KEY (forum_id)  REFERENCES forum   (uid),
  FOREIGN   KEY (user_id)   REFERENCES profile (uid)
--   FOREIGN   KEY (parent_id) REFERENCES post    (uid),
--   FOREIGN   KEY (thread_id) REFERENCES thread  (uid)
);

CREATE TABLE IF NOT EXISTS vote
(
  user_id   INT           NOT NULL,
  thread_id INT           NOT NULL,
  value     SMALLINT      NOT NULL   DEFAULT 0,
  is_edited BOOLEAN       NOT NULL   DEFAULT FALSE,

  FOREIGN KEY (user_id)   REFERENCES profile (uid) ON DELETE CASCADE,
  FOREIGN KEY (thread_id) REFERENCES thread  (uid) ON DELETE CASCADE
);


CREATE OR REPLACE FUNCTION add_to_thread() RETURNS TRIGGER AS $$
DECLARE
  old_value  INTEGER;
BEGIN
  --
  -- Добавление строки в emp_audit, которая отражает операцию, выполняемую в emp;
  -- для определения типа операции применяется специальная переменная TG_OP.
  --
  SELECT t.votes INTO old_value FROM thread AS t WHERE t.uid = NEW.thread_id;
  IF (TG_OP = 'UPDATE') THEN
    UPDATE thread
    SET votes = old_value + NEW.value - OLD.value
    WHERE thread.uid = NEW.thread_id;
    RETURN NEW;
  ELSIF (TG_OP = 'INSERT') THEN
    UPDATE thread
    SET votes = old_value + NEW.value
    WHERE thread.uid = NEW.thread_id;
    RETURN NEW;
  END IF;
  RETURN NULL; -- возвращаемое значение для триггера AFTER игнорируется
END;
$$ LANGUAGE plpgsql;


CREATE OR REPLACE FUNCTION add_path() RETURNS TRIGGER AS $$
    DECLARE
        new_path INTEGER[];
    BEGIN
        SELECT "path" INTO new_path FROM post p WHERE p.uid = NEW.parent_id;
        IF (TG_OP = 'INSERT') THEN
            UPDATE post
            SET path = array_append(new_path, NEW.uid)
            WHERE uid = NEW.uid;
            RETURN NEW;

        END IF;
    RETURN NULL; -- возвращаемое значение для триггера AFTER игнорируется
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER t_vote
  AFTER INSERT OR UPDATE OR DELETE ON vote
  FOR EACH ROW EXECUTE PROCEDURE add_to_thread ();

CREATE TRIGGER t_post
  AFTER INSERT ON post
  FOR EACH ROW EXECUTE PROCEDURE add_path ();



GRANT ALL PRIVILEGES ON DATABASE park_forum TO park_forum;--why we granted privileges to park_forum if
                                                          --db owner is postgres?
GRANT USAGE ON SCHEMA public TO park_forum;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO park_forum;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO park_forum;
