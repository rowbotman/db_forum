package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx"
	"../models"
	"net/http"
	"strconv"
)

func InsertIntoForum(data models.DataForNewForum) (models.DataForNewForum, error) {
	sqlStatement := `SELECT u.uid, u.nickname FROM profile u WHERE u.nickname = $1;`
	row := DB.QueryRow(sqlStatement, data.Nickname)
	authorId := 0
	nickname := ""
	err := row.Scan(&authorId, &nickname)
	if err == pgx.ErrNoRows {
		return models.DataForNewForum{}, errors.New("Can't find user with nickname: " + data.Nickname)
	} else if err != nil {
		return models.DataForNewForum{}, err
	}
	data.Nickname = nickname
	existForum, err := SelectForumInfo(data.Slug, false)
	if err == nil {
		return models.DataForNewForum{
			existForum.Title,
			existForum.User,
			existForum.Slug}, errors.New("slug exist")
	}
	sqlStatement = `INSERT INTO forum (title, author_id, slug) VALUES ($1, $2, $3);`
	_, err = DB.Exec(sqlStatement, data.Title, authorId, data.Slug)
	if err != nil {
		return models.DataForNewForum{}, err
	}

	return data, nil
}

func SelectForumInfo(slug string, isUid bool) (models.Forum, error) {
	var forum models.Forum
	sqlStatement1 := `
SELECT f.uid, f.title, f.slug, m.post_count FROM forum f 
JOIN forum_meta m ON (m.forum_id = f.uid) WHERE  `

	sqlStatement2 := `
SELECT f.uid, p.nickname, m.thread_count FROM forum f 
JOIN forum_meta m ON (m.forum_id = f.uid)
LEFT JOIN profile p ON (p.uid = f.author_id) WHERE `
	var row *pgx.Row
	if isUid {
		sqlStatement1 += `f.uid = $1;	`
		sqlStatement2 += `f.uid = $1;`
		id, err := strconv.Atoi(slug)
		if err != nil {
			return models.Forum{}, err
		}

		row = DB.QueryRow(sqlStatement1, id)
		err = row.Scan(
			&id,
			&forum.Title,
			&forum.Slug,
			&forum.Posts)
		if err != nil {
			return models.Forum{}, err
		}

		row = DB.QueryRow(sqlStatement2, id)
		err = row.Scan(
			&id,
			&forum.User,
			&forum.Threads)

		if err != nil {
			return models.Forum{}, err
		}
	} else {
		sqlStatement1 += `LOWER(f.slug) = LOWER($1);`
		sqlStatement2 += `LOWER(f.slug) = LOWER($1);`

		id := int64(0)
		row = DB.QueryRow(sqlStatement1, slug)
		fmt.Println(sqlStatement1, slug)
		err := row.Scan(
			&id,
			&forum.Title,
			&forum.Slug,
			&forum.Posts)
		if err != nil {
			return models.Forum{Slug: slug}, errors.New("Can't find forum by slug: " + slug)
		}

		row = DB.QueryRow(sqlStatement2, slug)
		fmt.Println(sqlStatement1, slug)
		err = row.Scan(
			&id,
			&forum.User,
			&forum.Threads)

		if err != nil {
			return models.Forum{}, err
		}
	}
	fmt.Println("finish")
	return forum, nil
}

func SelectForumUsers(slug string, limit int32, since string, desc bool, w http.ResponseWriter) error {
	sqlStatement := `SELECT uid FROM forum WHERE LOWER(slug) = LOWER($1);`
	forumId := int64(0)
	err := DB.QueryRow(sqlStatement, slug).Scan(&forumId)
	if err != nil {
		return errors.New("Can't find forum by slug: " + slug)
	}
	sqlStatement = `
SELECT * FROM (
    SELECT u.uid, u.nickname, u.full_name, u.about, u.email FROM profile u
        JOIN thread t ON (t.user_id = u.uid) WHERE t.forum_id = $1
	UNION 
	SELECT u.uid, u.nickname, u.full_name, u.about, u.email FROM profile u
	    JOIN post p ON (p.user_id = u.uid) WHERE p.forum_id = $1
) _ `
	if len(since) > 0 {
		if desc {
			sqlStatement += `WHERE lower(nickname)::bytea < lower($2)::bytea `
		} else {
			sqlStatement += `WHERE lower(nickname)::bytea > lower($2)::bytea `
		}
	}
	sqlStatement += `ORDER BY lower(nickname)::bytea`
	if desc {
		sqlStatement += ` DESC`
	} else {
		sqlStatement += ` ASC`
	}
	if len(since) > 0 {
		sqlStatement += ` LIMIT $3;`
	} else {
		sqlStatement += ` LIMIT $2;`
	}
	var rows *pgx.Rows
	if len(since) > 0 {
		rows, err = DB.Query(sqlStatement, forumId, since, limit)
		if err != nil {
			return err
		}
	} else {
		rows, err = DB.Query(sqlStatement, forumId, limit)
		if err != nil {
			return err
		}
	}
	defer rows.Close()
	users := models.Users{}
	for rows.Next() {
		newUser := models.User{}
		err = rows.Scan(
			&newUser.Pk,
			&newUser.Nickname,
			&newUser.Name,
			&newUser.About,
			&newUser.Email)
		if err != nil {
			return err
		}
		users = append(users, newUser)
	}

	err = rows.Err()
	if err != nil {
		return err
	}

	output, err := json.Marshal(users)
	if err != nil {
		//http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	w.Header().Set("content-type", "application/json")
	_, _ = w.Write(output)
	return nil
}

func SelectForumThreads(slug string, limit int32, since string, desc bool) (models.Threads, error) {
	sqlStatement := `SELECT title FROM forum WHERE LOWER(slug) = LOWER($1);`
	row := DB.QueryRow(sqlStatement, slug)
	forum := ""
	err := row.Scan(&forum)
	if err == pgx.ErrNoRows {
		return models.Threads{{Uid : -1}}, errors.New("Can't find forum by slug: " + slug)
	} else if err != nil {
		return nil, err
	}

	sqlStatement = `
  SELECT t.uid, t.title, p.nickname, f.slug, t.message, t.votes, t.slug, date_trunc('microseconds', t.created)
  FROM forum f
  JOIN thread  t ON (t.forum_id = f.uid)
  JOIN profile p ON (t.user_id  = p.uid)
  WHERE LOWER(f.slug) = LOWER($1) `

	var rows *pgx.Rows
	if len(since) > 0 {
		if desc {
			sqlStatement += ` AND t.created <= $2 ORDER BY t.created DESC LIMIT $3;`
		} else {
			sqlStatement += ` AND t.created >= $2 ORDER BY t.created ASC  LIMIT $3;`
		}
		rows, err = DB.Query(sqlStatement, slug, since, limit)
	} else {
		if desc {
			sqlStatement += ` ORDER BY t.created DESC LIMIT $2;`
		} else {
			sqlStatement += ` ORDER BY t.created ASC  LIMIT $2;`
		}
		rows, err = DB.Query(sqlStatement, slug, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	threads := models.Threads{}
	for rows.Next() {
		thread := models.ThreadInfo{}
		err = rows.Scan(
			&thread.Uid,
			&thread.Title,
			&thread.Author,
			&thread.Forum,
			&thread.Message,
			&thread.Votes,
			&thread.Slug,
			&thread.Created)
		if err != nil {
			return nil, err
		}
		threads = append(threads, thread)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return threads, nil
}

func InsertIntoThread(slug string, threadData models.ThreadInfo) (models.ThreadInfo, error) {
	sqlStatement := `SELECT p.uid FROM profile p WHERE p.nickname = $1;`
	row := DB.QueryRow(sqlStatement, threadData.Author)
	authorId := int64(0)
	err := row.Scan(&authorId)
	if err == pgx.ErrNoRows {
		return models.ThreadInfo{Uid: -1}, errors.New("Can't find thread author by nickname: " + threadData.Author)
	} else if err != nil {
		return models.ThreadInfo{}, err
	}

	sqlStatement = `SELECT f.uid, f.slug FROM forum f WHERE LOWER(f.slug) = LOWER($1);`
	row = DB.QueryRow(sqlStatement, slug)
	forum := int64(0)
	err = row.Scan(&forum, &threadData.Forum)
	if err == pgx.ErrNoRows {
		return models.ThreadInfo{Uid: -1}, errors.New("Can't find thread forum by slug: " + slug)
	} else if err != nil {
		return models.ThreadInfo{}, err
	}

	if threadData.Slug != nil {
		sqlStatement = `INSERT INTO thread VALUES(default, $1, $2, $3, $4, $5, $6, default) RETURNING uid;`
		err = DB.QueryRow(
			sqlStatement,
			authorId,
			forum,
			threadData.Title,
			threadData.Slug,
			threadData.Message,
			threadData.Created).Scan(&threadData.Uid)
	} else {
		sqlStatement = `INSERT INTO thread(uid, user_id, forum_id, title, message, created, votes) VALUES(default, $1, $2, $3, $4, $5, default) RETURNING uid;`
		err = DB.QueryRow(
			sqlStatement,
			authorId,
			forum,
			threadData.Title,
			threadData.Message,
			threadData.Created).Scan(&threadData.Uid)
	}

	if err == nil {
		sqlStatement = `CALL inc_threads($1);`
		_, err = DB.Exec(sqlStatement, forum)
		if err != nil {
			return models.ThreadInfo{}, err
		}
		return threadData, nil
	} else {
		existThread, ok := isThreadExist(*threadData.Slug)
		if ok {
			threadData.Title = existThread.Title
			threadData.Slug = existThread.Slug
			threadData.Message = existThread.Message
			threadData.Created = existThread.Created
			threadData.Uid = existThread.Uid
			sqlStatement = `
WITH get_name AS (
    SELECT nickname FROM profile WHERE uid = $1
) SELECT slug, nickname FROM forum, get_name WHERE uid = $2`
			err := DB.QueryRow(
				sqlStatement,
				existThread.UserId,
				existThread.ForumId).Scan(
				&threadData.Forum,
				&threadData.Author)
			if err != nil {
				return threadData, nil
			}
			return threadData, errors.New("thread exist")
		}
		return models.ThreadInfo{Uid: -1}, err
	}
}
