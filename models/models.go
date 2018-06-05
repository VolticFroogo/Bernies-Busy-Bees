package models

import (
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

// Token lifetimes
const (
	// AuthTokenValidTime is the lifetime of an auth token.
	AuthTokenValidTime = time.Minute * 15
	// RefreshTokenValidTime is the lifetime of a refresh token.
	RefreshTokenValidTime = time.Hour * 72
)

// Privileges
const (
	PrivNone = iota
	PrivUser
	PrivAdmin
	PrivSuperAdmin
)

// Post is the struct used for a post.
type Post struct {
	ID                                                       int
	Title, Description, ImagesJSON, CommentsJSON, CreateTime string
	Images                                                   []string
	Comments                                                 []DisplayComment
}

// Posts is an array of Post.
type Posts []Post

// PostEdit is the struct recieved by an admin when they change a post.
type PostEdit struct {
	ID                             int
	CsrfSecret, Title, Description string
}

// PostDelete is the struct recieved by an admin when they delete a post.
type PostDelete struct {
	ID         int
	CsrfSecret string
}

// NewComment is the struct recieved by a user when they comment on something.
type NewComment struct {
	ID, UserUUID        int
	Timestamp           int64
	CsrfSecret, Comment string
}

// Comment is the struct used to save comments in a DB.
type Comment struct {
	UserUUID    int
	Timestamp   int64
	Comment, ID string
}

// DisplayComment is the struct used to display a comment on the website.
type DisplayComment struct {
	UserUUID    int
	Timestamp   int64
	Comment, ID string
	User        User
}

// Comments is an array of comments to be stored in a DB on a post.
type Comments []Comment

// User is a user retrieved from a Database.
type User struct {
	UUID, Priv                                int
	Email, Password, Fname, Lname, CreateTime string
}

// Users is an array of User for the admin page.
type Users []User

// TokenClaims are the claims in a token.
type TokenClaims struct {
	jwt.StandardClaims
	CSRF string `json:"csrf"`
}

// TemplateVariables is the struct used when executing a template.
type TemplateVariables struct {
	CsrfSecret string
	User       User
	Users      Users
	Posts      Posts
	Post       Post
	UnixTime   int64
	Page       Page
}

// AJAXData is the struct used with the AJAX middleware.
type AJAXData struct {
	CsrfSecret string
}

// JTI is the struct used for JTIs in the DB.
type JTI struct {
	ID     int
	Expiry int64
	JTI    string
}

// Page is a convenience for template execution.
type Page struct {
	Next, Current, Last int
}

// ResponseWithID is a simple struct for responding to an AJAX request.
type ResponseWithID struct {
	Success bool   `json:"success"`
	ID      string `json:"id"`
}

// ResponseWithIDInt is a simple struct for responding to an AJAX request.
type ResponseWithIDInt struct {
	Success bool `json:"success"`
	ID      int  `json:"id"`
}
