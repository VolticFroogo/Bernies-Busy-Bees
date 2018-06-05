package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/VolticFroogo/Bernies-Busy-Bees/db/dbCredentials"
	"github.com/VolticFroogo/Bernies-Busy-Bees/helpers"
	"github.com/VolticFroogo/Bernies-Busy-Bees/models"
	_ "github.com/go-sql-driver/mysql" // Necessary for connecting to MySQL.
)

/*
	Structs and variables
*/

var (
	db *sql.DB
	// Users is a struct for the admin Users.
	Users models.Users
	// IndexPosts are the posts for the index page to prevent an attacker flooding our DB.
	IndexPosts models.Posts
)

// InitDB initializes the Database.
func InitDB() (err error) {
	db, err = sql.Open(dbCredentials.Type, dbCredentials.ConnString)
	if err != nil {
		return
	}

	err = UpdateUsers()
	if err != nil {
		return
	}

	err = UpdateIndexPosts()
	if err != nil {
		return
	}

	go jtiGarbageCollector()
	return
}

/*
	Helper functions
*/

func rowExists(query string, args ...interface{}) (exists bool, err error) {
	query = fmt.Sprintf("SELECT exists (%s)", query)
	err = db.QueryRow(query, args...).Scan(&exists)
	return
}

/*
	MySQL DataBase related functions
*/

// StoreRefreshToken generates, stores and then returns a JTI.
func StoreRefreshToken() (jti models.JTI, err error) {
	// No need to duplication check as the JTI takes input from time and are unique.
	jti.JTI, err = helpers.GenerateRandomString(32)
	if err != nil {
		return
	}

	jti.Expiry = time.Now().Add(models.RefreshTokenValidTime).Unix()

	_, err = db.Exec("INSERT INTO jti (jti, expiry) VALUES (?, ?)", jti.JTI, jti.Expiry)
	if err != nil {
		return
	}

	rows, err := db.Query("SELECT id FROM jti WHERE jti=? AND expiry=?", jti.JTI, jti.Expiry)
	if err != nil {
		return
	}

	defer rows.Close()

	rows.Next()
	err = rows.Scan(&jti.ID) // Scan data from query.
	return
}

// GetJTI takes a JTI string and returns the JTI struct.
func GetJTI(jti string) (jtiStruct models.JTI, err error) {
	rows, err := db.Query("SELECT id, expiry FROM jti WHERE jti=?", jti)
	if err != nil {
		return
	}

	defer rows.Close()

	jtiStruct.JTI = jti
	rows.Next()
	err = rows.Scan(&jtiStruct.ID, &jtiStruct.Expiry) // Scan data from query.
	return
}

// CheckJTI returns the validity of a JTI.
func CheckJTI(jti models.JTI) (valid bool, err error) {
	if jti.Expiry > time.Now().Unix() { // Check if token has expired.
		return true, nil // Token is valid.
	}

	_, err = db.Exec("DELETE FROM jti WHERE id=?", jti.ID)
	if err != nil {
		return false, err
	}

	return false, nil // Token is invalid.
}

// DeleteJTI deletes a JTI based on a jti key.
func DeleteJTI(jti string) (err error) {
	_, err = db.Exec("DELETE FROM jti WHERE jti=?", jti)
	return
}

func jtiGarbageCollector() {
	ticker := time.NewTicker(5 * time.Minute) // Tick every five minutes.
	for {
		<-ticker.C
		rows, err := db.Query("SELECT id, jti, expiry FROM jti")
		if err != nil {
			log.Printf("Error querying JTI DB in JTI garbage collector: %v", err)
			return
		}

		defer rows.Close()

		jti := models.JTI{} // Create struct to store a JTI in.
		for rows.Next() {
			err = rows.Scan(&jti.ID, &jti.JTI, &jti.Expiry) // Scan data from query.
			if err != nil {
				log.Printf("Error scanning rows in JTI garbage collector: %v", err)
				return
			}

			_, err := CheckJTI(jti)
			if err != nil {
				log.Printf("Error checking in JTI garbage collector: %v", err)
				return
			}
		}
	}
}

// GetUserFromID retrieves a user from the MySQL database.
func GetUserFromID(uuid int) (user models.User, err error) {
	rows, err := db.Query("SELECT email, password, fname, lname, priv, create_time FROM users WHERE uuid=?", uuid)
	if err != nil {
		return
	}

	defer rows.Close()

	user.UUID = uuid
	for rows.Next() {
		err = rows.Scan(&user.Email, &user.Password, &user.Fname, &user.Lname, &user.Priv, &user.CreateTime) // Scan data from query.
		if err != nil {
			return
		}
	}

	return
}

// GetUserFromEmail retrieves a user's ID from the MySQL database.
func GetUserFromEmail(email string) (user models.User, err error) {
	rows, err := db.Query("SELECT uuid, password, fname, lname, priv, create_time FROM users WHERE email=?", email)
	if err != nil {
		return
	}

	defer rows.Close()

	user.Email = email
	for rows.Next() {
		err = rows.Scan(&user.UUID, &user.Password, &user.Fname, &user.Lname, &user.Priv, &user.CreateTime) // Scan data from query.
		if err != nil {
			return
		}
	}

	return
}

// UpdateUsers updates the users by querying the MySQL DataBase.
func UpdateUsers() (err error) {
	rows, err := db.Query("SELECT uuid, email, fname, lname, password, priv, create_time FROM users")
	if err != nil {
		return
	}

	defer rows.Close()

	users := models.Users{} // Create struct to store slides in.
	user := models.User{}   // Create struct to store a slide in.
	for rows.Next() {
		err = rows.Scan(&user.UUID, &user.Email, &user.Fname, &user.Lname, &user.Password, &user.Priv, &user.CreateTime) // Scan data from query.
		if err != nil {
			return
		}

		users = append(users, user) // Append just read slide into the slides.
	}

	Users = users // Replace the old menu with the newly read struct.
	return
}

// UpdateIndexPosts updates the index posts by querying the MySQL DataBase.
func UpdateIndexPosts() (err error) {
	rows, err := db.Query("SELECT id, title, description, images, comments, create_time FROM posts ORDER BY id DESC LIMIT 3")
	if err != nil {
		return
	}

	defer rows.Close()

	posts := models.Posts{} // Create struct to store posts in.
	for rows.Next() {
		post := models.Post{} // Create struct to store a post in.

		err = rows.Scan(&post.ID, &post.Title, &post.Description, &post.ImagesJSON, &post.CommentsJSON, &post.CreateTime) // Scan data from query.
		if err != nil {
			return
		}

		err = json.Unmarshal([]byte(post.ImagesJSON), &post.Images)
		if err != nil {
			return
		}

		err = json.Unmarshal([]byte(post.CommentsJSON), &post.Comments)
		if err != nil {
			return
		}

		posts = append(posts, post) // Append just read post into the posts.
	}

	IndexPosts = posts // Replace the old menu with the newly read struct.
	return
}

// GetPosts returns a specified amount of posts.
func GetPosts(amount, perPage, page int) (posts models.Posts, err error) {
	rows, err := db.Query("SELECT id, title, description, images, comments, create_time FROM posts ORDER BY id DESC LIMIT ?,?", perPage*(page-1), amount)
	if err != nil {
		return
	}

	defer rows.Close()

	post := models.Post{} // Create struct to store a post in.
	for rows.Next() {
		err = rows.Scan(&post.ID, &post.Title, &post.Description, &post.ImagesJSON, &post.CommentsJSON, &post.CreateTime) // Scan data from query.
		if err != nil {
			return
		}

		posts = append(posts, post) // Append just read post into the posts.
	}

	return
}

// GetPost returns a post with a specified ID.
func GetPost(id int) (post models.Post, exists bool, err error) {
	rows, err := db.Query("SELECT title, description, images, comments, create_time FROM posts WHERE id=?", id)
	if err != nil {
		return
	}

	defer rows.Close()

	exists = rows.Next()
	if !exists {
		return
	}

	post.ID = id
	err = rows.Scan(&post.Title, &post.Description, &post.ImagesJSON, &post.CommentsJSON, &post.CreateTime) // Scan data from query.

	exists = true
	return
}

// EditUser updates a user.
func EditUser(ID int, Email, Password, Fname, Lname string, Privileges int) (err error) {
	_, err = db.Exec("UPDATE users SET email=?, password=?, fname=?, lname=?, priv=? WHERE uuid=?", Email, Password, Fname, Lname, Privileges, ID)
	if err != nil {
		return
	}

	err = UpdateUsers()
	return
}

// EditUserNoPassword updates a user without changing the password.
func EditUserNoPassword(ID int, Email, Fname, Lname string, Privileges int) (err error) {
	_, err = db.Exec("UPDATE users SET email=?, fname=?, lname=?, priv=? WHERE uuid=?", Email, Fname, Lname, Privileges, ID)
	if err != nil {
		return
	}

	err = UpdateUsers()
	return
}

// EditSelf updates a user from settings.
func EditSelf(ID int, Password, Fname, Lname string) (err error) {
	_, err = db.Exec("UPDATE users SET password=?, fname=?, lname=? WHERE uuid=?", Password, Fname, Lname, ID)
	if err != nil {
		return
	}

	err = UpdateUsers()
	return
}

// EditSelfNoPassword updates a user from settings without changing the password.
func EditSelfNoPassword(ID int, Fname, Lname string) (err error) {
	_, err = db.Exec("UPDATE users SET fname=?, lname=? WHERE uuid=?", Fname, Lname, ID)
	if err != nil {
		return
	}

	err = UpdateUsers()
	return
}

// NewUser creates a new user.
func NewUser(Email, Password, Fname, Lname string, Privileges int) (id int, err error) {
	_, err = db.Exec("INSERT INTO users (email, password, fname, lname, priv) VALUES (?, ?, ?, ?, ?)", Email, Password, Fname, Lname, Privileges)
	if err != nil {
		return
	}

	rows, err := db.Query("SELECT uuid FROM users WHERE email=? AND password=? AND fname=? AND lname=? AND priv=? ORDER BY uuid DESC", Email, Password, Fname, Lname, Privileges)
	if err != nil {
		return
	}

	defer rows.Close()

	rows.Next()
	err = rows.Scan(&id)
	if err != nil {
		return
	}

	err = UpdateUsers()
	return
}

// DeleteUser deletes a user.
func DeleteUser(ID int) (err error) {
	_, err = db.Exec("DELETE FROM users WHERE uuid=?", ID)
	if err != nil {
		return
	}

	err = UpdateUsers()
	return
}

// NewPost creates a new post.
func NewPost(title, description string, fileLocations []string) (err error) {
	fileLocationBytes, err := json.Marshal(fileLocations)
	if err != nil {
		return
	}

	_, err = db.Exec("INSERT INTO posts (title, description, images, comments) VALUES (?, ?, ?, ?)", title, description, string(fileLocationBytes[:]), "[]")
	if err != nil {
		return
	}

	err = UpdateIndexPosts()
	return
}

// EditPost updates a post.
func EditPost(ID int, Title, Description string) (err error) {
	_, err = db.Exec("UPDATE posts SET title=?, description=? WHERE id=?", Title, Description, ID)
	if err != nil {
		return
	}

	err = UpdateIndexPosts()
	return
}

// DeletePost deletes a post and returns all of the images.
func DeletePost(ID int) (images []string, err error) {
	rows, err := db.Query("SELECT images FROM posts WHERE id=?", ID)
	if err != nil {
		return
	}

	defer rows.Close()

	rows.Next()
	var imagesJSON string
	err = rows.Scan(&imagesJSON)
	if err != nil {
		return
	}

	err = json.Unmarshal([]byte(imagesJSON), &images)
	if err != nil {
		return
	}

	_, err = db.Exec("DELETE FROM posts WHERE id=?", ID)
	if err != nil {
		return
	}

	err = UpdateIndexPosts()
	return
}

// AddCommentPost adds a comment to a post.
func AddCommentPost(comment models.NewComment) (id string, err error) {
	rows, err := db.Query("SELECT comments FROM posts WHERE id=?", comment.ID)
	if err != nil {
		return
	}

	defer rows.Close()

	rows.Next()
	var commentsJSONOld string
	err = rows.Scan(&commentsJSONOld)
	if err != nil {
		return
	}

	var comments models.Comments
	err = json.Unmarshal([]byte(commentsJSONOld), &comments)
	if err != nil {
		return
	}

	unique := false
	for !unique {
		id, err = helpers.GenerateRandomString(32)
		if err != nil {
			return
		}

		uniqueTemp := true
		for _, a := range comments {
			if a.ID == id {
				uniqueTemp = false
				break
			}
		}

		if uniqueTemp {
			unique = true
		}
	}

	comments = append(comments, models.Comment{
		ID:        id,
		UserUUID:  comment.UserUUID,
		Timestamp: comment.Timestamp,
		Comment:   comment.Comment,
	})

	commentsBytes, err := json.Marshal(comments)
	if err != nil {
		return
	}

	_, err = db.Exec("UPDATE posts SET comments=? WHERE id=?", string(commentsBytes[:]), comment.ID)
	return
}

// DeleteCommentPost deletes a comment to a post.
func DeleteCommentPost(commentID string, postID int) (err error) {
	rows, err := db.Query("SELECT comments FROM posts WHERE id=?", postID)
	if err != nil {
		return
	}

	defer rows.Close()

	rows.Next()
	var commentsJSONOld string
	err = rows.Scan(&commentsJSONOld)
	if err != nil {
		return
	}

	var comments models.Comments
	err = json.Unmarshal([]byte(commentsJSONOld), &comments)
	if err != nil {
		return
	}

	for i, a := range comments {
		if a.ID == commentID {
			comments = append(comments[:i], comments[i+1:]...) // Delete comment.
			break
		}
	}

	commentsBytes, err := json.Marshal(comments)
	if err != nil {
		return
	}

	_, err = db.Exec("UPDATE posts SET comments=? WHERE id=?", string(commentsBytes[:]), postID)
	return
}

// DeleteCommentPostIfOwner deletes a comment to a post if the comment owner matches a specified UUID.
func DeleteCommentPostIfOwner(commentID string, postID, userUUID int) (owner bool, err error) {
	rows, err := db.Query("SELECT comments FROM posts WHERE id=?", postID)
	if err != nil {
		return
	}

	defer rows.Close()

	rows.Next()
	var commentsJSONOld string
	err = rows.Scan(&commentsJSONOld)
	if err != nil {
		return
	}

	var comments models.Comments
	err = json.Unmarshal([]byte(commentsJSONOld), &comments)
	if err != nil {
		return
	}

	for i, a := range comments {
		if a.ID == commentID {
			if a.UserUUID == userUUID {
				comments = append(comments[:i], comments[i+1:]...) // Delete comment.
				owner = true
				break
			} else {
				return
			}
		}
	}

	commentsBytes, err := json.Marshal(comments)
	if err != nil {
		return
	}

	_, err = db.Exec("UPDATE posts SET comments=? WHERE id=?", string(commentsBytes[:]), postID)
	return
}

// AddEmailVerification adds an email verification code to the DB.
func AddEmailVerification(id string, userUUID int, email string) (err error) {
	exists, err := rowExists("SELECT id FROM email WHERE useruuid=?", userUUID)
	if err != nil {
		return
	}
	if exists {
		_, err = db.Exec("DELETE FROM email WHERE useruuid=?", userUUID)
		if err != nil {
			return
		}
	}

	_, err = db.Exec("INSERT INTO email (uuid, useruuid, email) VALUES (?, ?, ?)", id, userUUID, email)
	return
}

// GetEmailVerification retrieves an email verification information.
func GetEmailVerification(id string) (userUUID int, email string, err error) {
	rows, err := db.Query("SELECT useruuid, email FROM email WHERE uuid=?", id)
	if err != nil {
		return
	}

	defer rows.Close()

	rows.Next()
	err = rows.Scan(&userUUID, &email)
	if err != nil {
		return
	}

	if userUUID != 0 && email != "" {
		_, err = db.Exec("DELETE FROM email WHERE uuid=?", id)
	}

	return
}

// EditSelfEmail updates a user's email after verification.
func EditSelfEmail(uuid int, email string) (err error) {
	_, err = db.Exec("UPDATE users SET email=? WHERE uuid=?", email, uuid)
	if err != nil {
		return
	}

	err = UpdateUsers()
	return
}

// AddRecovery adds a password recovery code to the DB.
func AddRecovery(id string, userUUID int, email string) (err error) {
	exists, err := rowExists("SELECT id FROM recovery WHERE useruuid=?", userUUID)
	if err != nil {
		return
	}
	if exists {
		_, err = db.Exec("DELETE FROM recovery WHERE useruuid=?", userUUID)
		if err != nil {
			return
		}
	}

	_, err = db.Exec("INSERT INTO recovery (uuid, useruuid, email) VALUES (?, ?, ?)", id, userUUID, email)
	return
}

// GetRecovery retrieves a password recovery code from the DB.
func GetRecovery(id string) (userUUID int, email string, err error) {
	rows, err := db.Query("SELECT useruuid, email FROM recovery WHERE uuid=?", id)
	if err != nil {
		return
	}

	defer rows.Close()

	rows.Next()
	err = rows.Scan(&userUUID, &email)
	if err != nil {
		return
	}

	if userUUID != 0 && email != "" {
		_, err = db.Exec("DELETE FROM recovery WHERE uuid=?", id)
	}

	return
}

// EditPassword updates a user's password after password recovery.
func EditPassword(uuid int, password string) (err error) {
	_, err = db.Exec("UPDATE users SET password=? WHERE uuid=?", password, uuid)
	if err != nil {
		return
	}

	err = UpdateUsers()
	return
}
