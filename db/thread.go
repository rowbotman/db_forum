package db

import (
	"errors"
	"github.com/jackc/pgx"
	json "github.com/mailru/easyjson"
	"github.com/rowbotman/db_forum/models"
	"net/http"
	"strconv"
	"time"
)

func isThreadExist(slugOrId string) (models.Thread, bool) {
	reqId, err := strconv.ParseInt(slugOrId, 10, 64)
	sqlStatement := `SELECT uid, title, forum_id, "message", slug, user_id, created FROM thread `
	var row *pgx.Row
	thread := models.Thread{}
	if err != nil {
		sqlStatement += `WHERE LOWER(slug) = LOWER($1);`
		row = DB.QueryRow(sqlStatement, slugOrId)
	} else {
		sqlStatement += `WHERE uid = $1;`
		row = DB.QueryRow(sqlStatement, reqId)
	}
	id := int64(0)
	err = row.Scan(
		&thread.Uid,
		&thread.Title,
		&thread.ForumId,
		&thread.Message,
		&thread.Slug,
		&thread.UserId,
		&thread.Created)
	if err != nil || id < 0 {
		return models.Thread{}, false
	}
	return thread, true
}

func isParentPost(parentId int, thread int64) (ok bool) {
	sqlStatement := `SELECT p.thread_id FROM post p WHERE p.uid = $1;`
	threadId := int64(0)
	err := DB.QueryRow(sqlStatement, parentId).Scan(&threadId)
	if err != nil && parentId != 0{
		return false
	}
	return threadId == thread || threadId == 0
}


func InsertNewPosts(slugOrId string, posts models.Posts) (models.Posts, error) {
	sqlTime := `SELECT current_timestamp(3);`
	curTime := time.Time{}
	err := DB.QueryRow(sqlTime).Scan(&curTime)
	if err != nil {
		return nil, errors.New("error getting current time")
	}

	thread, ok := isThreadExist(slugOrId)
	if !ok {
		return models.Posts{{Uid: -1}},
			errors.New("Can't find post thread by id: " + slugOrId)
	}

	sqlForPostData := `SELECT slug FROM forum WHERE uid = $1;`
	forum := ""
	err = DB.QueryRow(sqlForPostData, thread.ForumId).Scan(&forum)
	if err == pgx.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	if len(posts) == 0 {
		return posts, nil
	}
	sqlStatement := `INSERT INTO post(uid, is_edited, path, forum_id, user_id, thread_id, message, created, parent_id, author) VALUES `
	valuesStr := ``
	retId := ` RETURNING uid;`
	values := []interface{}{}
	itemNum := 7
	postNum := len(posts)
	for i := 0; i < postNum; i++ {
		userData, err := SelectUser(posts[i].Author)
		if err != nil {
			return []models.Post{{Uid: -1}},
				errors.New("Can't find post author by nickname:" + posts[i].Author)
		}
		ok := isParentPost(posts[i].ParentId, thread.Uid)
		if !ok {
			return posts, errors.New("Parent post was created in another thread")
		}
		valuesStr += ` (default, default, array[0], `
		for j := 1; j <= itemNum; j++ {
			valuesStr += `$` + strconv.Itoa(i * itemNum + j)
			if j != itemNum {
				valuesStr += `, `
			}
		}
		posts[i].Forum = forum
		posts[i].Created = curTime
		posts[i].ThreadId = thread.Uid
		valuesStr += ` ) `
		values = append(values, thread.ForumId, userData.Pk, posts[i].ThreadId, posts[i].Message, posts[i].Created, posts[i].ParentId, posts[i].Author)
		if i != postNum - 1 {
			valuesStr += ` , `
		}
	}
	trans, _  := DB.Begin()
	rows, err := trans.Query(sqlStatement + valuesStr + retId, values...)
	if err != nil {
		_ = trans.Rollback()
		return nil, err
	}
	i := 0
	for rows.Next() {
		if err := rows.Scan(&posts[i].Uid); err != nil {
			return nil, err
		}
		i++
	}
	err = trans.Commit()
	if err != nil {
		return nil, err
	}
	sqlStatement = `CALL inc_posts($1, $2);`
	_, err = DB.Exec(sqlStatement, thread.ForumId, postNum)
	if err != nil {
		return nil, err
	}
	return posts, nil
}

func SelectFromThread(slugOrId string, isId bool, thread *models.ThreadInfo) error {
	sqlStatement := `SELECT t.uid, t.title, u.nickname, f.slug, t.message, t.votes, t.created, t.slug FROM thread t
	JOIN forum f ON (f.uid = t.forum_id)
	JOIN profile u ON (t.user_id = u.uid) WHERE`
	var row *pgx.Row
	if isId {
		sqlStatement += ` t.uid = $1;`
		id, err := strconv.ParseInt(slugOrId, 10, 64)
		if err != nil {
			return err
		}
		row = DB.QueryRow(sqlStatement, id)
	} else {
		sqlStatement += ` LOWER(t.slug) = LOWER($1);`
		row = DB.QueryRow(sqlStatement, slugOrId)
	}

	err := row.Scan(
		&thread.Uid,
		&thread.Title,
		&thread.Author,
		&thread.Forum,
		&thread.Message,
		&thread.Votes,
		&thread.Created,
		&thread.Slug)
	if err == pgx.ErrNoRows {
		(*thread).Uid = -1
		return errors.New("Can't find thread by slug: " + slugOrId)
	} else if err != nil {
		return err
	}

	return nil
}

func UpdateThread(slugOrId string, thread *models.ThreadInfo) error {
	existThread, ok := isThreadExist(slugOrId)
	if !ok {
		(*thread).Uid = -1
		return errors.New("Can't find thread by slug: " + slugOrId)
	}
	sqlStatement := `UPDATE thread SET `
	status := 1
	if len((*thread).Title) > 0 {
		sqlStatement += `title = $1 `
		if len((*thread).Message) > 0 {
			sqlStatement += `, message = $2  WHERE uid = $3;`
			status = 2
		} else {
			sqlStatement += `WHERE uid = $2;`
		}
	} else if len((*thread).Message) > 0 {
		sqlStatement += `message = $1  WHERE uid = $2; `
		status = 3
	} else {
		status = 0
	}

	if status > 0 {
		var err error
		switch status {
		case 1: {
			_, err = DB.Exec(sqlStatement, thread.Title, existThread.Uid)
			break
		}
		case 2: {
			_, err = DB.Exec(sqlStatement, thread.Title, thread.Message, existThread.Uid)
			break
		}
		case 3: {
			_, err = DB.Exec(sqlStatement, thread.Message, existThread.Uid)
			break
		}
		default:
			//log.Println("something got error in switch")
		}
		if err != nil {
			return err
		}
	}
	threadInfo := models.ThreadInfo{}
	err := SelectFromThread(*existThread.Slug, false, &threadInfo)
	if err != nil {
		return err
	}
	(*thread).Author = threadInfo.Author
	(*thread).Created = threadInfo.Created
	(*thread).Uid = threadInfo.Uid
	(*thread).Forum = threadInfo.Forum
	(*thread).Slug = threadInfo.Slug
	(*thread).Message = threadInfo.Message
	(*thread).Title = threadInfo.Title
	return nil
}


func SelectThreadPosts(slugOrid string, limit int32, since int64,
	sort string, desc bool, w http.ResponseWriter) (models.Posts, error) {
	_, err := strconv.ParseInt(slugOrid, 10, 64)
	isId := true
	if err != nil {
		isId = false
	}
	thread := models.ThreadInfo{}
	//log.Println("start")
	err = SelectFromThread(slugOrid, isId, &thread)
	//log.Println("finish")
	if err != nil && thread.Uid == -1 {
		return models.Posts{{Uid: -1}}, err
	}
	sqlStatement := `SELECT p.uid, p.parent_id, p.author,
       p.message, p.is_edited,
       p.thread_id, p.created FROM post p `
	/*
	SELECT p.uid, p.path, p.parent_id, p.author, p.is_edited,
	       p.thread_id, p.created FROM post p WHERE p.thread_id = 376
			AND p.path[1] IN (
				SELECT p.uid FROM post p WHERE array_length(p.path, 1) = 1
					AND p.thread_id = 376
					AND p.uid < (SELECT path[1] FROM post WHERE uid = 2974) ORDER BY p.path[1] LIMIT 3
			) ORDER BY p.path[1] DESC, p.path;
	 */
	var rows *pgx.Rows
	switch sort {
	case "tree": {
		sqlStatement += ` WHERE p.thread_id = $1 AND p.path `
		if desc {
			if since == 0 {
				sqlStatement += ` >= ARRAY(SELECT uid FROM post ORDER BY uid ASC LIMIT 1)`
			} else {
				sqlStatement += ` < (SELECT path FROM post WHERE uid = $2) `
			}
		} else {
			if since == 0 {
				sqlStatement += ` >= ARRAY(SELECT uid FROM post ORDER BY uid ASC LIMIT 1)`
			} else {
				sqlStatement += ` > (SELECT path FROM post WHERE uid = $2) `

			}
		}
		if desc {
			sqlStatement += `ORDER BY p.path DESC, p.created DESC`
		} else {
			sqlStatement += `ORDER BY p.path, p.created ASC`
		}
		//fmt.Println(sqlStatement, thread.Uid, since)
		if since > 0 {
			sqlStatement += ` LIMIT $3;`
			rows, err = DB.Query(sqlStatement, thread.Uid, since, limit)
		} else {
			sqlStatement += ` LIMIT $2;`
			rows, err = DB.Query(sqlStatement, thread.Uid, limit)
		}
//		fmt.Println(sqlStatement, thread.Uid, since, limit)
	}
	case "parent_tree": {
		strLimit := strconv.FormatInt(int64(limit), 10)
		sqlStatement += `WHERE p.thread_id = $1 AND p.path[1] IN ( 
							SELECT uid FROM post WHERE thread_id = $1 `
		if since > 0 {
			if desc {
				sqlStatement += ` AND uid < `
			} else {
				sqlStatement += ` AND uid > `
			}
			sqlStatement += ` (SELECT path[1] FROM post WHERE uid = $2) `
		}
		sqlStatement += ` AND array_length(path, 1) = 1 `
		if desc {
			sqlStatement += `ORDER BY path[1] DESC LIMIT ` + strLimit + `) ORDER BY p.path[1] DESC, p.path;`
		} else {
			sqlStatement += `ORDER BY path[1] LIMIT ` + strLimit + `) ORDER BY p.path;`
		}
		//fmt.Println(sqlStatement, thread.Uid, since)
		if since > 0 {
			rows, err = DB.Query(sqlStatement, thread.Uid, since)
		} else {
			rows, err = DB.Query(sqlStatement, thread.Uid)
		}
	}
	default: {
		if desc {
			sqlStatement += ` WHERE p.thread_id = $1 `
			if since > 0 {
				sqlStatement += ` AND p.uid < $2 ORDER BY p.created DESC, p.uid DESC LIMIT $3;`
				rows, err = DB.Query(sqlStatement, thread.Uid, since, limit)
			} else {
				sqlStatement += `ORDER BY p.created DESC, p.uid DESC LIMIT $2;`
				rows, err = DB.Query(sqlStatement, thread.Uid, limit)
			}
		} else {
			sqlStatement += ` WHERE p.thread_id = $1 AND p.uid > $2 ORDER BY p.created, p.uid ASC LIMIT $3;`
			rows, err = DB.Query(sqlStatement, thread.Uid, since, limit)
		}
	}
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	posts := models.Posts{}
	for rows.Next() {
		newPost := models.Post{}
		newPost.Forum = thread.Forum
		err = rows.Scan(
			&newPost.Uid,
			&newPost.ParentId,
			&newPost.Author,
			&newPost.Message,
			&newPost.IsEdited,
			&newPost.ThreadId,
			&newPost.Created)
		if err != nil {
			return nil, err
		}
		posts = append(posts, newPost)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	//fmt.Println(posts)
	////log.Println(posts)
	output, err := json.Marshal(posts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, err
	}

	w.Header().Set("content-type", "application/json")
	_, _ = w.Write(output)
	//return posts, nil
	return nil, nil
}

func UpdateVote(slugOrId string, vote models.VoteInfo) (models.ThreadInfo, error) {
	_, err := strconv.ParseInt(slugOrId, 10, 64)
	thread := models.ThreadInfo{}
	if err != nil {
		err = SelectFromThread(slugOrId, false, &thread)
	} else {
		err = SelectFromThread(slugOrId, true, &thread)
	}
	if err != nil {
		return models.ThreadInfo{Uid: -1},
		errors.New("Can't find thread by slug: " + slugOrId)
	}

	userId := int64(0)
	sqlGetId := `SELECT p.uid FROM profile p WHERE p.nickname = $1;`
	err = DB.QueryRow(sqlGetId, vote.Nickname).Scan(&userId)
	if err != nil {
		return models.ThreadInfo{Uid: -1},
			errors.New("Can't find user by nickname: " + vote.Nickname)
	}

	sqlGetVotes := `SELECT "value", is_edited FROM vote WHERE user_id = $1 AND thread_id = $2;`
	value := int(0)
	isEdited := false
	err = DB.QueryRow(sqlGetVotes, userId, thread.Uid).Scan(&value, &isEdited)
	if err == pgx.ErrNoRows {
		sqlStatement := `
	INSERT INTO vote (user_id, thread_id, "value") VALUES ($1, $2, $3)`
		_, err  = DB.Exec(sqlStatement, userId, thread.Uid, vote.Voice)
		if err != nil {
			return models.ThreadInfo{}, err
		}
		thread.Votes += vote.Voice
		return thread, nil
	} else if err != nil {
		return models.ThreadInfo{}, err
	}
	if (value > 0 && vote.Voice > 0) ||
		(value < 0 && vote.Voice < 0) {
		return thread, nil
	}

	sqlVote := `UPDATE vote SET "value" = $1, is_edited = true WHERE user_id = $2 AND thread_id = $3;`
	_, err  = DB.Exec(sqlVote, vote.Voice, userId, thread.Uid)
	if err != nil {
		return models.ThreadInfo{}, err
	}
	thread.Votes += vote.Voice * 2
	return thread, nil
}

