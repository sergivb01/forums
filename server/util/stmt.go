package util

var (
	CreateUsersTable = `CREATE TABLE IF NOT EXISTS users(
	    id serial unique primary key,
	    username varchar(25) not null,
	    password varchar(255) not null,
	    registeredAt timestamptz default now()
	);`

	CreatePostsTable = `CREATE TABLE IF NOT EXISTS posts(
		id serial unique primary key,
		title varchar(50) not null,
		content text,
		createdAt timestamptz default now(),
		userID int not null
	);`
)
