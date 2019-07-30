package user

import "time"

type Post struct {
	ID int `db:"id" json:"id"`
	UserID int `db:"userid" json:"userID"`

	Title string `db:"title" json:"title"`
	Content string `db:"content" json:"content"`

	CreatedAt time.Time `db:"createdat" json:"createdAt"`
}
