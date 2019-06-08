package db

type Vote struct {
	Nickname string `json:"nickname,omitempty"`
	SlugOrId string `json:"slug_or_id,omitempty"`
	Value    int    `json:"voice, omitempty"`
}
