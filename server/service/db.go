package service

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/sergivb01/forums/user"
	"github.com/sergivb01/forums/util"
)

var (
	errUserNotFound = errors.New("user NOT found or does not exists")
	errUserAlreadyExists = errors.New("user already exists")
)

func (s *Server) getPostByID(id int) (*user.Post, error) {
	t := util.Start("getPostByID(%d)", id)
	defer t.Stop()

	post := &user.Post{}

	if err := s.db.QueryRowx("SELECT * FROM posts WHERE id=$1", id).StructScan(post); err != nil {
		return nil, err
	}

	return post, nil
}

func (s *Server) createPost(usr user.User, title, content string) (*user.Post, error) {
	t := util.Start("createPost(%s,%s,%s)", usr.Username, title, content)
	defer t.Stop()

	post := &user.Post{}

	row := s.db.QueryRowx("INSERT INTO posts (title, content, userid) VALUES ($1, $2, $3) RETURNING *", title, content, usr.ID)
	if err := row.StructScan(post); err != nil {
		return nil, fmt.Errorf("couldn't add post into database: %v", err)
	}

	return post, nil
}

func (s *Server) registerUser(username, password string) (*user.User, error) {
	if usr, err := s.findUserByUsername(username); usr != nil || err != errUserNotFound {
		if err == nil {
			return nil, errUserAlreadyExists
		}
		return nil, fmt.Errorf("couldn't check if user already exists: %v", err)
	}

	hashed, err := util.HashFromPassword(password)
	if err != nil {
		return nil, fmt.Errorf("couldn't securly hash your password: %v", err)
	}

	t := util.Start("registerUser(%s)", username)
	defer t.Stop()

	usr := &user.User{}

	err = s.db.QueryRowx(`INSERT INTO users (username, password) VALUES ($1, $2) RETURNING *`, username, hashed).StructScan(usr)
	if err != nil {
		return nil, fmt.Errorf("couldn't add user into database: %v", err)
	}

	return usr, nil
}

func (s *Server) findUserByUsername(username string) (*user.User, error) {
	t := util.Start("findUserByUsername(%s)", username)
	defer t.Stop()

	usr := &user.User{}

	err := s.db.Get(usr, "SELECT * FROM users WHERE username=$1", username);
	if usr == nil || err == sql.ErrNoRows {
		return nil, errUserNotFound
	}

	return usr, err
}
