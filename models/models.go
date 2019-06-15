package models

import "time"

//easyjson:json
type Forum struct {
	Title   string `json:"title,omitempty"`
	User    string `json:"user,omitempty"`
	Slug    string `json:"slug,omitempty"`
	Posts   int    `json:"posts,omitempty"`
	Threads int    `json:"threads,omitempty"`
}

//easyjson:json
type DataForNewForum struct {
	Title    string `json:"title"`
	Nickname string `json:"user"`
	Slug     string `json:"slug"`
}


//easyjson:json
type Post struct {
	Uid      int64     `json:"id,omitempty"`
	ParentId int       `json:"parent,omitempty"`
	Author   string    `json:"author,omitempty"`
	Message  string    `json:"message,omitempty"`
	Forum    string    `json:"forum,omitempty"`
	ThreadId int64     `json:"thread,omitempty"`
	IsEdited bool      `json:"isEdited,omitempty"`
	Created  time.Time `json:"created,omitempty"`
}

//easyjson:json
type DataForUpdPost struct {
	Id      int64  `json:"id,omitempty"`
	Message string `json:"message,omitempty"`
}


//easyjson:json
type ServiceInfo struct {
	User   int64 `json:"user"`
	Forum  int64 `json:"forum"`
	Thread int64 `json:"thread"`
	Post   int64 `json:"post"`
}


type Thread struct {
	Uid     int64     `json:"uid, omitempty"`
	Title   string    `json:"title, omitempty"`
	UserId  int       `json:"userId, omitempty"`
	ForumId int       `json:"forumId, omitempty"`
	Forum   string    `json:"forum, omitempty"`
	Message string    `json:"message, omitempty"`
	Votes   int       `json:"votes, omitempty"`
	Slug    *string   `json:"slug, omitempty"`
	Created time.Time `json:"created, omitempty"`
}

//easyjson:json
type ThreadInfo struct {
	Uid     int64     `json:"id,omitempty"`
	Title   string    `json:"title,omitempty"`
	Author  string    `json:"author,omitempty"`
	Forum   string    `json:"forum,omitempty"`
	Message string    `json:"message,omitempty"`
	Votes   int       `json:"votes,omitempty"`
	Slug    *string   `json:"slug,omitempty"`
	Created time.Time `json:"created,omitempty"`
}

type ThreadInfoMin struct {
	Uid     int64     `json:"id, omitempty"`
	Title   string    `json:"title, omitempty"`
	Author  string    `json:"author, omitempty"`
	Forum   string    `json:"forum, omitempty"`
	Message string    `json:"message, omitempty"`
	Created time.Time `json:"created, omitempty"`
}

//easyjson:json
type VoteInfo struct {
	Nickname string `json:"nickname,omitempty"`
	Voice    int    `json:"voice, omitempty"`
}

//easyjson:json
type User struct {
	Pk       int64      `json:"-"`         // why we used '-' here?
	Nickname string     `json:"nickname,omitempty"`
	Name     string     `json:"fullname,omitempty"`
	About    string     `json:"about,omitempty"`
	Email    string     `json:"email,omitempty"`
}

//easyjson:json
type Vote struct {
	Nickname string `json:"nickname,omitempty"`
	SlugOrId string `json:"slug_or_id,omitempty"`
	Value    int    `json:"voice, omitempty"`
}

//easyjson:json
type NotFoundPage struct {
	Message string `json:"message"`
}

func (us *User)IsEmpty() bool {
	if len(us.Email) == 0 &&
		len(us.Name) == 0 && len(us.About) == 0 {
		return true
	}
	return false
}

//easyjson:json
type Users []User

//easyjson:json
type Posts []Post

//easyjson:json
type Threads []ThreadInfo

//easyjson:json
type FullThreadInfo map[string]interface{}
