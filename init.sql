-- CREATE DATABASE park_forum WITH OWNER = postgres;

CREATE EXTENSION IF NOT EXISTS CITEXT;
DROP INDEX IF EXISTS forum_slug_idx;
DROP INDEX IF EXISTS profile_nick_idx;
DROP INDEX IF EXISTS profile_email_idx;
DROP INDEX IF EXISTS vote_id_thread_idx;
DROP INDEX IF EXISTS thread_userid_idx;
DROP INDEX IF EXISTS thread_forumid_idx;
DROP INDEX IF EXISTS post_path_idx;
DROP INDEX IF EXISTS forum_id_idx;
DROP INDEX IF EXISTS thread_id_idx;
DROP INDEX IF EXISTS post_id_idx;
DROP INDEX IF EXISTS post_forumid_idx;
DROP INDEX IF EXISTS forum_authorid_idx;
DROP INDEX IF EXISTS profile_lownick_idx;
DROP INDEX IF EXISTS profile_id_idx;
DROP INDEX IF EXISTS thread_slug_idx;
DROP INDEX IF EXISTS post_parentid_uid_idx;
DROP INDEX IF EXISTS post_userid_idx;
DROP INDEX IF EXISTS post_many_idx;
DROP INDEX IF EXISTS post_uid_withpath_idx;

--
DROP INDEX IF EXISTS post_parentid_idx;
DROP INDEX IF EXISTS post_threadpath_idx;
DROP INDEX IF EXISTS post_threadid_idx;
DROP INDEX IF EXISTS post_pathcreated_idx;
DROP INDEX IF EXISTS post_idincl_idx;

TRUNCATE TABLE profile, forum, thread, vote, post, forum_meta CASCADE;
DROP TABLE     IF EXISTS vote             CASCADE;
DROP TABLE     IF EXISTS post             CASCADE;
DROP TABLE     IF EXISTS thread           CASCADE;
DROP TABLE     IF EXISTS forum            CASCADE;
DROP TABLE     IF EXISTS profile          CASCADE;
DROP TABLE     IF EXISTS forum_meta;
DROP TRIGGER   IF EXISTS t_vote           ON vote;
DROP TRIGGER   IF EXISTS t_post           ON post;
DROP TRIGGER   IF EXISTS t_forum_meta     ON post;
DROP TRIGGER   IF EXISTS t_add_forum      ON forum;
DROP FUNCTION  IF EXISTS add_to_thread();
DROP FUNCTION  IF EXISTS clear_post_count();
DROP PROCEDURE IF EXISTS inc_posts(fid INT, new_posts BIGINT);
DROP PROCEDURE IF EXISTS inc_threads(fid INT);
DROP FUNCTION  IF EXISTS add_path();

-- DROP TABLE IF EXISTS  CASCADE;

-- CREATE ROLE park_forum;
-- CREATE USER park WITH SUPERUSER PASSWORD 'admin';

CREATE UNLOGGED TABLE IF NOT EXISTS profile
(
  uid       SERIAL                                              PRIMARY KEY,
  nickname  CITEXT       UNIQUE NOT NULL CHECK (nickname <> ''),
  full_name VARCHAR(128)        NOT NULL CHECK (nickname <> ''),
  about     TEXT                                                 DEFAULT '',
--   email     VARCHAR(256) UNIQUE NOT NULL CHECK (email <> '')
  email     CITEXT UNIQUE       NOT NULL CHECK (email <> '')
) WITH (autovacuum_enabled = off);

CREATE UNLOGGED TABLE IF NOT EXISTS forum
(
  uid       SERIAL                                          PRIMARY KEY,
  title     TEXT        NOT NULL CHECK ( title <> '' ),
  author_id INT                 NOT NULL,
  slug      CITEXT       UNIQUE NOT NULL,
--   slug      VARCHAR(256) UNIQUE NOT NULL,
  FOREIGN   KEY (author_id) REFERENCES profile (uid) -- попробовать убрать
) WITH (autovacuum_enabled = off);

CREATE UNLOGGED TABLE IF NOT EXISTS forum_meta
(
    forum_id     INT NOT NULL REFERENCES forum(uid),
    post_count   BIGINT DEFAULT 0,
    user_count   BIGINT DEFAULT 0,
    thread_count BIGINT DEFAULT 0
) WITH (autovacuum_enabled = off);


CREATE UNLOGGED TABLE IF NOT EXISTS thread
(
  uid      SERIAL                                            PRIMARY KEY,
  user_id  INT                         NOT NULL,
  forum_id INT                         NOT NULL,
  title    TEXT                        NOT NULL CHECK ( title <> '' ),
--   slug     CITEXT       UNIQUE         NOT NULL,
  slug     CITEXT UNIQUE         ,
  message  TEXT                        NOT NULL,
  created  TIMESTAMP(3) WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
  votes    INT                         NOT NULL DEFAULT 0,

  FOREIGN  KEY (user_id)  REFERENCES profile (uid),
  FOREIGN  KEY (forum_id) REFERENCES forum   (uid)
) WITH (autovacuum_enabled = off);

CREATE UNLOGGED TABLE IF NOT EXISTS post
(
  uid       SERIAL                                          PRIMARY KEY,
  parent_id INT                          NOT NULL DEFAULT 0,
  path      INTEGER[]                    NOT NULL,
  forum_id  INT                          NOT NULL,
  user_id   INT                          NOT NULL,
  thread_id INT                          NOT NULL,
  message   TEXT                         NOT NULL,
  is_edited BOOLEAN                               DEFAULT FALSE,
  created   TIMESTAMP(3)  WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
  author    CITEXT                                DEFAULT NULL
--   FOREIGN   KEY (forum_id)  REFERENCES forum   (uid),
--   FOREIGN   KEY (user_id)   REFERENCES profile (uid)
--   FOREIGN   KEY (parent_id) REFERENCES post    (uid),
--   FOREIGN   KEY (thread_id) REFERENCES thread  (uid)
) WITH (autovacuum_enabled = off);

CREATE UNLOGGED TABLE IF NOT EXISTS vote
(
  user_id   INT           NOT NULL,
  thread_id INT           NOT NULL,
  value     SMALLINT      NOT NULL   DEFAULT 0,
  is_edited BOOLEAN       NOT NULL   DEFAULT FALSE,

  FOREIGN KEY (user_id)   REFERENCES profile (uid) ON DELETE CASCADE,
  FOREIGN KEY (thread_id) REFERENCES thread  (uid) ON DELETE CASCADE
) WITH (autovacuum_enabled = off);


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

CREATE OR REPLACE PROCEDURE inc_posts(fid INT, new_posts BIGINT)
LANGUAGE plpgsql AS $$
BEGIN
    -- subtracting the amount from the sender's account
    UPDATE forum_meta
    SET post_count = post_count + new_posts
    WHERE forum_id = fid;
    COMMIT;
END;
$$;

CREATE OR REPLACE PROCEDURE inc_threads(fid INT)
    LANGUAGE plpgsql AS $$
BEGIN
    -- subtracting the amount from the sender's account
    UPDATE forum_meta
    SET thread_count = thread_count + 1
    WHERE forum_id = fid;
    COMMIT;
END;
$$;

CREATE OR REPLACE FUNCTION clear_post_count() RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'DELETE') THEN
        UPDATE forum_meta
        SET post_count = post_count - 1
        WHERE forum_id = OLD.uid;
        RETURN NULL;
    END IF;
    RETURN NULL; -- возвращаемое значение для триггера AFTER игнорируется
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION add_forum() RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'INSERT') THEN
        INSERT INTO forum_meta VALUES(NEW.uid, DEFAULT, DEFAULT);
        RETURN NULL;
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

CREATE TRIGGER t_forum_meta
    AFTER DELETE ON post
    FOR EACH ROW EXECUTE PROCEDURE clear_post_count ();

CREATE TRIGGER t_add_forum
    AFTER INSERT ON forum
    FOR EACH ROW EXECUTE FUNCTION add_forum();

-- forum uid = +2
-- forum lower(slug) =  +5
-- thread forum_id = +2
-- thread lower(slug) = +2
-- thread uid = +4
-- profile nickname = +5
-- post uid +5
-- thread uid and post path >/>= +1
-- thread uid = and post uid (mb and post uid >/< path[1])
-- thread uid = and post >/< +2
-- vote user_id = and thread_id = +1
-- profile lower(nickname) = or lower(email) = +1
-- profile lower(email) = +1


CREATE INDEX IF NOT EXISTS forum_slug_idx        ON forum(LOWER(slug));
CREATE INDEX IF NOT EXISTS forum_id_idx          ON forum(uid);
CREATE INDEX IF NOT EXISTS forum_authorid_idx    ON forum(author_id);
CREATE INDEX IF NOT EXISTS profile_nick_idx      ON profile(nickname);
CREATE INDEX IF NOT EXISTS profile_email_idx     ON profile(LOWER(email));
CREATE INDEX IF NOT EXISTS profile_lownick_idx   ON profile(LOWER(nickname));
CREATE INDEX IF NOT EXISTS profile_id_idx        ON profile(uid); -- ?
CREATE INDEX IF NOT EXISTS thread_userid_idx     ON thread(user_id);
CREATE INDEX IF NOT EXISTS thread_forumid_idx    ON thread(forum_id);
CREATE INDEX IF NOT EXISTS thread_slug_idx       ON thread USING HASH(LOWER(slug));
CREATE INDEX IF NOT EXISTS thread_id_idx         ON thread(uid);
CREATE INDEX IF NOT EXISTS post_forumid_idx      ON post(forum_id);
-- CREATE INDEX IF NOT EXISTS post_parentid_idx     ON post(parent_id);
CREATE INDEX IF NOT EXISTS post_parentid_uid_idx ON post(parent_id, uid);
-- CREATE INDEX IF NOT EXISTS post_threadpath_idx   ON post(thread_id, path);
CREATE INDEX IF NOT EXISTS post_userid_idx       ON post(user_id);
-- CREATE INDEX IF NOT EXISTS post_threadid_idx     ON post(thread_id);
CREATE INDEX IF NOT EXISTS post_id_idx           ON post(uid);
-- CREATE INDEX IF NOT EXISTS post_pathcreated_idx  ON post(path, created);
CREATE INDEX IF NOT EXISTS post_many_idx         ON post(thread_id, parent_id, uid);
CREATE INDEX IF NOT EXISTS post_uid_withpath_idx ON post(uid) INCLUDE (path);
CREATE INDEX IF NOT EXISTS post_path_idx         ON post(path);
-- CREATE INDEX IF NOT EXISTS post_idincl_idx ON post(uid) INCLUDE (path, thread_id, forum_id);

CREATE INDEX IF NOT EXISTS vote_id_thread_idx    ON vote(user_id, thread_id);

CLUSTER post_uid_withpath_idx ON post;

GRANT ALL PRIVILEGES ON DATABASE park_forum TO park_forum;--why we granted privileges to park_forum if
                                                          --db owner is postgres?
GRANT USAGE ON SCHEMA public TO park_forum;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO park_forum;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO park_forum;
