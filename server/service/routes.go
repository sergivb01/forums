package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"strconv"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"

	"github.com/sergivb01/forums/util"
)

func (s *Server) routes() {
	r := mux.NewRouter()

	r.HandleFunc("/", s.isLoggedIn(s.handleIndex())).Methods("GET")

	r.HandleFunc("/post/{id}", s.handlePostGET()).Methods("GET")
	r.HandleFunc("/post", s.isLoggedIn(s.handlePostPOST())).Methods("POST")

	r.HandleFunc("/refresh", s.handleRefreshToken()).Methods("POST")
	r.HandleFunc("/login", s.handleLogin()).Methods("POST")
	r.HandleFunc("/register", s.handleRegister()).Methods("POST")

	s.router = r
}

func (s *Server) handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	}
}

func (s *Server) handlePostGET() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rawID, ok := mux.Vars(r)["id"]
		if !ok {
			http.Error(w, "Missing id param", http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(rawID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to parse id number %s: %v", rawID, err), http.StatusBadRequest)
			return
		}


		post, err := s.getPostByID(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(post)
	}
}

func (s *Server) handlePostPOST() http.HandlerFunc {
	type postParams struct {
		Title string `json:"title"`
		Content string `json:"content"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var rawPost postParams

		if err := json.NewDecoder(r.Body).Decode(&rawPost); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		claims, err := s.getClaims(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		usr := claims.User

		post, err := s.createPost(usr, rawPost.Title, rawPost.Content)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(post)
	}
}

func (s *Server) handleRegister() http.HandlerFunc {
	type credentialPost struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var creds credentialPost

		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		user, err := s.registerUser(creds.Username, creds.Password)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	}
}

func (s *Server) handleLogin() http.HandlerFunc {
	// TODO: implement postgresql login system and remove mockup sys

	type credentialPost struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var creds credentialPost

		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		user, err := s.findUserByUsername(creds.Username)
		if err != nil {
			http.Error(w, "couldn't find user by username: "+err.Error(), http.StatusBadRequest)
			return
		}

		if !util.ComparePassword(creds.Password, user.Password) {
			http.Error(w, "invalid user or password for username: "+creds.Username, http.StatusUnauthorized)
			return
		}

		token, err := s.generateToken(user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    token.TokenString,
			Expires:  token.ExpiresAt,
			HttpOnly: true,
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(token)
	}
}

func (s *Server) handleRefreshToken() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claim, err := s.getClaims(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// do not allow to refresh token if it expires in > 30 seconds
		diff := time.Unix(claim.ExpiresAt, 0).Sub(time.Now())

		if diff > time.Second*30 {
			fmt.Fprintf(w, "token expires in more than 30 seconds. Expires in %s", diff)
			return
		}

		expirationTime := time.Now().Add(s.cfg.JWT.Duration)
		claim.ExpiresAt = expirationTime.Unix()

		token := jwt.NewWithClaims(jwt.SigningMethodHS512, claim)
		tokenString, err := token.SignedString([]byte(s.cfg.JWT.Secret))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    tokenString,
			Expires:  expirationTime,
			HttpOnly: true,
		})
		fmt.Fprintf(w, "Token refreshed! Token=%s, ExpiresAt=%s", tokenString, expirationTime)
	}
}
