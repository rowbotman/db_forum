package db

import (
	"../models"
	"errors"
	"github.com/jackc/pgx"
	json "github.com/mailru/easyjson"
	"net/http"
	"strconv"
)


func UpdatePost(data models.DataForUpdPost) (models.Post, error) {
	sqlStatement := `
  SELECT p.uid, p.parent_id, p.message, p.is_edited, u.nickname, f.slug, p.thread_id, p.created 
  FROM post p JOIN profile u ON (p.user_id = u.uid) JOIN forum f ON (p.forum_id = f.uid)
  WHERE p.uid = $1 GROUP BY p.uid, u.nickname, f.slug;`
	row := DB.QueryRow(sqlStatement, data.Id)
	post := models.Post{}
	isEdited := false
	err := row.Scan(
		&post.Uid,
		&post.ParentId,
		&post.Message,
		&isEdited,
		&post.Author,
		&post.Forum,
		&post.ThreadId,
		&post.Created)
	if err == pgx.ErrNoRows {
		ret := models.Post{}
		ret.Uid = -1
		return ret, errors.New("Can't find post with id: " + strconv.FormatInt(data.Id, 10))
	} else if err != nil {
		return models.Post{}, err
	}
	if len(data.Message) > 0 {
		if data.Message != post.Message {
			sqlStatement = `UPDATE post SET message = $1, is_edited = TRUE WHERE uid = $2;`
			_, err = DB.Exec(sqlStatement, data.Message, post.Uid)
			if err != nil {
				return models.Post{}, err
			}
			post.Message = data.Message
			post.IsEdited = true
		}
	}
	return post, nil
}



func GetPostInfo(postId int64, strArray []string, w http.ResponseWriter) (map[string]interface{}, error) {
	sqlStatement := `
  SELECT p.uid, p.parent_id, u.nickname, p.message, p.is_edited, f.slug, f.uid, p.thread_id, p.created 
  FROM post p JOIN profile u ON (p.user_id = u.uid) JOIN forum f ON (p.forum_id = f.uid)
  WHERE p.uid = $1 GROUP BY p.uid, p.parent_id, u.nickname, f.slug, f.uid;`
	row := DB.QueryRow(sqlStatement, postId)
	post := models.Post{}
	forumId := int(0)
	err := row.Scan(
		&post.Uid,
		&post.ParentId,
		&post.Author,
		&post.Message,
		&post.IsEdited,
		&post.Forum,
		&forumId,
		&post.ThreadId,
		&post.Created)
	fullInfo := models.FullThreadInfo{}
	if err == pgx.ErrNoRows {
		fullInfo["err"] = true
		return fullInfo, errors.New("Can't find post with id: " + strconv.FormatInt(postId, 10))
	} else if err != nil {
		return models.FullThreadInfo{}, err
	}
	fullInfo["post"] = post
	if len(strArray) > 0 && len(strArray[0]) > 0 {
		for _, obj := range strArray {
			switch obj {
			case "user": {
				userData, err := SelectUser(post.Author)
				if err != nil {
					return models.FullThreadInfo{}, err
				}
				fullInfo["author"] = userData
			}
			case "forum": {
				id := strconv.Itoa(int(forumId))
				forumData, err := SelectForumInfo(id, true)
				if err != nil {
					return models.FullThreadInfo{}, err
				}
				fullInfo["forum"] = forumData
			}
			case "thread": {
				id := strconv.FormatInt(post.ThreadId, 10)
				threadData := models.ThreadInfo{}
				err := SelectFromThread(id, true, &threadData)
				if err != nil {
					return models.FullThreadInfo{}, err
				}
				fullInfo["thread"] = threadData
			}
			}
		}
	}
	output, err := json.Marshal(fullInfo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return models.FullThreadInfo{}, nil
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(output)
	return models.FullThreadInfo{}, nil
}
