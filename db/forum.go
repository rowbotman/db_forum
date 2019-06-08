package db

import (
	"database/sql"
	"errors"
	"strconv"
)

type Forum struct {
	Title   string `json:"title,omitempty"`
	User    string `json:"user,omitempty"`
	Slug    string `json:"slug,omitempty"`
	Posts   int    `json:"posts,omitempty"`
	Threads int    `json:"threads,omitempty"`
}

type DataForNewForum struct {
	Title    string `json:"title"`
	Nickname string `json:"user"`
	Slug     string `json:"slug"`
}

func InsertIntoForum(data DataForNewForum) (DataForNewForum, error) {
	sqlStatement := `SELECT u.uid, u.nickname FROM profile u WHERE u.nickname = $1;`
	row := DB.QueryRow(sqlStatement, data.Nickname)
	authorId := 0
	nickname := ""
	err := row.Scan(&authorId, &nickname)
	if err == sql.ErrNoRows {
		return DataForNewForum{}, errors.New("Can't find user with nickname: " + data.Nickname)
	} else if err != nil {
		return DataForNewForum{}, err
	}
	data.Nickname = nickname
	existForum, err := SelectForumInfo(data.Slug, false)
	if err == nil {
		return DataForNewForum{
			existForum.Title,
			existForum.User,
			existForum.Slug}, errors.New("slug exist")
	}
	sqlStatement = `INSERT INTO forum (title, author_id, slug) VALUES ($1, $2, $3);`
	_, err = DB.Exec(sqlStatement, data.Title, authorId, data.Slug)
	if err != nil {
		return DataForNewForum{}, err
	}

	return data, nil
}

func SelectForumInfo(slug string, isUid bool) (Forum, error) {
	var forum Forum
	sqlStatement1 := `
SELECT f.uid, f.title, f.slug, COUNT(p.uid) FROM forum f 
LEFT JOIN post p ON (p.forum_id = f.uid) WHERE `

	sqlStatement2 := `
SELECT f.uid, p.nickname, COUNT(t.uid) FROM forum f 
LEFT JOIN thread t ON (t.forum_id = f.uid)
LEFT JOIN profile p ON (p.uid = f.author_id) WHERE `
	var row *sql.Row
	if isUid {
		sqlStatement1 += `f.uid = $1 GROUP BY f.uid, f.title;`
		sqlStatement2 += `f.uid = $1 GROUP BY f.uid, p.nickname;`
		id, err := strconv.Atoi(slug)
		if err != nil {
			return Forum{}, err
		}

		row = DB.QueryRow(sqlStatement1, id)
		err = row.Scan(
			&id,
			&forum.Title,
			&forum.Slug,
			&forum.Posts)
		if err != nil {
			return Forum{}, err
		}

		row = DB.QueryRow(sqlStatement2, id)
		err = row.Scan(
			&id,
			&forum.User,
			&forum.Threads)

		if err != nil {
			return Forum{}, err
		}

	} else {
		sqlStatement1 += `LOWER(f.slug) = LOWER($1) GROUP BY f.uid, f.title;`
		sqlStatement2 += `LOWER(f.slug) = LOWER($1) GROUP BY f.uid, p.nickname;`

		id := int64(0)

		row = DB.QueryRow(sqlStatement1, slug)
		err := row.Scan(
			&id,
			&forum.Title,
			&forum.Slug,
			&forum.Posts)
		if err != nil {
			return Forum{Slug: slug}, errors.New("Can't find forum by slug: " + slug)
		}

		row = DB.QueryRow(sqlStatement2, slug)
		err = row.Scan(
			&id,
			&forum.User,
			&forum.Threads)

		if err != nil {
			return Forum{}, err
		}
	}
	return forum, nil
}

func SelectForumUsers(slug string, limit int32, since string, desc bool) ([]User, error) {
	sqlStatement := `SELECT uid FROM forum WHERE LOWER(slug) = LOWER($1);`
	forumId := int64(0)
	err := DB.QueryRow(sqlStatement, slug).Scan(&forumId)
	if err != nil {
		return nil, errors.New("Can't find forum by slug: " + slug)
	}
	sqlSelect := `SELECT u.uid, u.nickname, u.full_name, u.about, u.email FROM profile u` // TODO: убрать лишний sqlStatement
	sqlStatement = `
SELECT * FROM (
` + sqlSelect + `JOIN thread t ON (t.user_id = u.uid) WHERE t.forum_id = $1 
UNION
` + sqlSelect + `JOIN post p   ON (p.user_id = u.uid) WHERE p.forum_id = $1
) _ ORDER BY nickname COLLATE "C"`
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
	var rows *sql.Rows
	if len(since) > 0 {
		rows, err = DB.Query(sqlStatement, forumId, since, limit)
		if err != nil {
			return nil, err
		}
	} else {
		rows, err = DB.Query(sqlStatement, forumId, limit)
		if err != nil {
			return nil, err
		}
	}
	defer rows.Close()
	users := []User{}
	for rows.Next() {
		newUser := User{}
		err = rows.Scan(
			&newUser.Pk,
			&newUser.Nickname,
			&newUser.Name,
			&newUser.About,
			&newUser.Email)
		if err != nil {
			return nil, err
		}
		users = append(users, newUser)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return users, nil
}

func SelectForumThreads(slug string, limit int32, since string, desc bool) ([]ThreadInfo, error) {
	sqlStatement := `SELECT title FROM forum WHERE LOWER(slug) = LOWER($1)`
	row := DB.QueryRow(sqlStatement, slug)
	forum := ""
	err := row.Scan(&forum)
	if err == sql.ErrNoRows {
		return []ThreadInfo{{Slug: slug}}, errors.New("Can't find forum by slug: " + slug)
	} else if err != nil {
		return nil, err
	}

	sqlStatement = `
  SELECT t.uid, t.title, p.nickname, f.slug, t.message, t.votes, t.slug, date_trunc('microseconds', t.created)
  FROM forum f
  JOIN thread  t ON (t.forum_id = f.uid)
  JOIN profile p ON (t.user_id  = p.uid)
  WHERE LOWER(f.slug) = LOWER($1) `

	var rows *sql.Rows
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
	threads := []ThreadInfo{}
	for rows.Next() {
		thread := ThreadInfo{}
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

func InsertIntoThread(slug string, threadData ThreadInfo) (ThreadInfo, error) {
	sqlStatement := `SELECT p.uid FROM profile p WHERE p.nickname = $1;`
	row := DB.QueryRow(sqlStatement, threadData.Author)
	authorId := int64(0)
	err := row.Scan(&authorId)
	if err == sql.ErrNoRows {
		return ThreadInfo{Uid: -1}, errors.New("Can't find thread author by nickname: " + threadData.Author)
	} else if err != nil {
		return ThreadInfo{}, err
	}

	sqlStatement = `SELECT f.uid, f.slug FROM forum f WHERE LOWER(f.slug) = LOWER($1);`
	row = DB.QueryRow(sqlStatement, slug)
	forum := int64(0)
	err = row.Scan(&forum, &threadData.Forum)
	if err == sql.ErrNoRows {
		return ThreadInfo{Uid: -1}, errors.New("Can't find thread forum by slug: " + slug)
	} else if err != nil {
		return ThreadInfo{}, err
	}

	existThread, ok := isThreadExist(threadData.Slug)
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
	sqlStatement = `INSERT INTO thread VALUES(default, $1, $2, $3, $4, $5, $6, default) RETURNING uid;`
	err = DB.QueryRow(
		sqlStatement,
		authorId,
		forum,
		threadData.Title,
		threadData.Slug,
		threadData.Message,
		threadData.Created).Scan(&threadData.Uid)
	if err != nil {
		return ThreadInfo{}, err
	}
	return threadData, nil
}
