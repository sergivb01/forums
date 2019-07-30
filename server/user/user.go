package user

import "time"

type User struct {
	ID         int       `db:"id" json:"id"`
	Username   string    `db:"username" json:"username"`
	Password   string    `db:"password" json:"-"`
	RegisteredAt time.Time `db:"registeredat" json:"registeredAt"`
}
