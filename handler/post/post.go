package post

import (
	"bufio"
	"encoding/json"
	"html/template"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/VolticFroogo/Bernies-Busy-Bees/db"
	"github.com/VolticFroogo/Bernies-Busy-Bees/helpers"
	"github.com/VolticFroogo/Bernies-Busy-Bees/middleware"
	"github.com/VolticFroogo/Bernies-Busy-Bees/models"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/zemirco/uid"
)

type deleteCommentData struct {
	PostID                int
	CommentID, CsrfSecret string
}

// Posts is the handler for the posts page.
func Posts(w http.ResponseWriter, r *http.Request) {
	uuidString := context.Get(r, "uuid").(string)
	uuid, err := strconv.Atoi(uuidString)
	if err != nil {
		helpers.ThrowErr(w, r, "Error converting string to int", err)
		return
	}

	user, err := db.GetUserFromID(uuid)
	if err != nil {
		helpers.ThrowErr(w, r, "Error getting user from ID", err)
		return
	}

	t, err := template.ParseFiles("handler/templates/post/posts.html", "handler/templates/nested.html") // Parse the HTML pages
	if err != nil {
		helpers.ThrowErr(w, r, "Template parsing error", err)
		return
	}

	csrfSecret, err := r.Cookie("csrfSecret")
	if err != nil {
		helpers.ThrowErr(w, r, "Reading cookie error", err)
		return
	}

	vars := mux.Vars(r)
	page, err := strconv.Atoi(vars["page"])
	if err != nil {
		helpers.ThrowErr(w, r, "Page number to int error", err)
		return
	}

	posts, err := db.GetPosts(7, 6, page)
	if err != nil {
		helpers.ThrowErr(w, r, "Getting posts error", err)
		return
	}

	if page < 1 {
		// The user is trying to get an unexpected result; throw an error.
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for i := 0; i < len(posts); i++ {
		post := &posts[i]
		err = json.Unmarshal([]byte(post.ImagesJSON), &post.Images)
		if err != nil {
			helpers.ThrowErr(w, r, "Unmarshalling images error", err)
			return
		}
	}

	nextPage := 0
	if len(posts) == 7 {
		posts = append(posts[:6], posts[7:]...) // Delete the last post.
		nextPage = page + 1
	} else if len(posts) == 0 && page != 1 {
		http.Redirect(w, r, "/panel/posts/1", http.StatusTemporaryRedirect)
	}

	variables := models.TemplateVariables{
		User:       user,
		CsrfSecret: csrfSecret.Value,
		Users:      db.Users,
		Posts:      posts,
		Page: models.Page{
			Next:    nextPage,
			Current: page,
			Last:    page - 1,
		},
	}

	err = t.Execute(w, variables) // Execute temmplate with variables
	if err != nil {
		helpers.ThrowErr(w, r, "Template execution error", err)
	}
}

// Post is the handler for a post's page.
func Post(w http.ResponseWriter, r *http.Request) {
	uuidString := context.Get(r, "uuid").(string)
	uuid, err := strconv.Atoi(uuidString)
	if err != nil {
		helpers.ThrowErr(w, r, "Error converting string to int", err)
		return
	}

	user, err := db.GetUserFromID(uuid)
	if err != nil {
		helpers.ThrowErr(w, r, "Error getting user from ID", err)
		return
	}

	if user.Priv != models.PrivUser && user.Priv != models.PrivAdmin && user.Priv != models.PrivSuperAdmin {
		return
	}

	csrfSecret, err := r.Cookie("csrfSecret")
	if err != nil {
		helpers.ThrowErr(w, r, "Reading cookie error", err)
		return
	}

	vars := mux.Vars(r)
	postID, err := strconv.Atoi(vars["postID"])
	if err != nil {
		helpers.ThrowErr(w, r, "Converting post to int error", err)
		return
	}

	post, exists, err := db.GetPost(postID)
	if err != nil {
		helpers.ThrowErr(w, r, "Getting post error", err)
		return
	}
	if !exists {
		return
	}

	t, err := template.ParseFiles("handler/templates/post/post.html", "handler/templates/nested.html") // Parse the HTML pages
	if err != nil {
		helpers.ThrowErr(w, r, "Template parsing error", err)
		return
	}

	err = json.Unmarshal([]byte(post.ImagesJSON), &post.Images)
	if err != nil {
		helpers.ThrowErr(w, r, "Unmarshalling images error", err)
		return
	}
	err = json.Unmarshal([]byte(post.CommentsJSON), &post.Comments)
	if err != nil {
		helpers.ThrowErr(w, r, "Unmarshalling comments error", err)
		return
	}

	users := make(map[int]models.User)
	for i := range post.Comments {
		userUUID := post.Comments[i].UserUUID

		if user, ok := users[userUUID]; ok {
			post.Comments[i].User = user
		} else {
			user, err := db.GetUserFromID(userUUID)
			if err != nil {
				helpers.ThrowErr(w, r, "Getting user from ID error", err)
			}
			users[userUUID] = user
			post.Comments[i].User = user
		}
	}

	// Reverse the comments (newest first).
	for i, j := 0, len(post.Comments)-1; i < j; i, j = i+1, j-1 {
		post.Comments[i], post.Comments[j] = post.Comments[j], post.Comments[i]
	}

	variables := models.TemplateVariables{
		User:       user,
		CsrfSecret: csrfSecret.Value,
		Users:      db.Users,
		Post:       post,
		UnixTime:   time.Now().Unix(),
	}
	err = t.Execute(w, variables) // Execute temmplate with variables
	if err != nil {
		helpers.ThrowErr(w, r, "Template execution error", err)
	}
}

// NewPage is the handler for the new post page.
func NewPage(w http.ResponseWriter, r *http.Request) {
	uuidString := context.Get(r, "uuid").(string)
	uuid, err := strconv.Atoi(uuidString)
	if err != nil {
		helpers.ThrowErr(w, r, "Error converting string to int", err)
		return
	}

	user, err := db.GetUserFromID(uuid)
	if err != nil {
		helpers.ThrowErr(w, r, "Error getting user from ID", err)
		return
	}

	if user.Priv != models.PrivUser && user.Priv != models.PrivAdmin && user.Priv != models.PrivSuperAdmin {
		return
	}

	if err != nil {
		helpers.ThrowErr(w, r, "Template parsing error", err)
		return
	}

	csrfSecret, err := r.Cookie("csrfSecret")
	if err != nil {
		helpers.ThrowErr(w, r, "Reading cookie error", err)
		return
	}

	t, err := template.ParseFiles("handler/templates/post/post-new.html", "handler/templates/nested.html") // Parse the HTML pages
	if err != nil {
		helpers.ThrowErr(w, r, "Template parsing error", err)
		return
	}

	variables := models.TemplateVariables{
		User:       user,
		CsrfSecret: csrfSecret.Value,
	}
	err = t.Execute(w, variables) // Execute temmplate with variables
	if err != nil {
		helpers.ThrowErr(w, r, "Template execution error", err)
	}
}

// New is the function called when the user sends a new post request.
func New(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 100*1024*1024) // 100MB max request size otherwise decline.
	err := r.ParseMultipartForm(10 * 1024 * 1024)          // Use a total of 10MB RAM and the rest in temporary disk (SSD for my server).
	if err != nil {
		helpers.SuccessResponse(false, w, r)
		helpers.ThrowErr(w, r, "Parsing multipart form error", err)
		return
	}

	form := r.MultipartForm // Declare the multipart form.

	uploader := s3manager.NewUploader(session.Must(session.NewSession(&aws.Config{
		Region: aws.String("eu-west-2"),
	}))) // Create an uploader with default uploader with session

	var imageLocations []string

	// Upload the thumbnail.
	location, err := uploadImage(form.File["thumbnail"][0], uploader)
	if err != nil {
		helpers.SuccessResponse(false, w, r)
		helpers.ThrowErr(w, r, "Uploading image error", err)
		return
	}

	imageLocations = append(imageLocations, location)

	// Upload the images.
	images := form.File["images"]
	var wg sync.WaitGroup // Declare a waitgroup.
	wg.Add(len(images))   // Make the waitgroup wait until every image is done.
	for inc := range images {
		go func(i int) {
			defer wg.Done()

			if images[i].Filename == "" {
				return
			}

			location, err := uploadImage(images[i], uploader)
			if err != nil {
				helpers.SuccessResponse(false, w, r)
				helpers.ThrowErr(w, r, "Uploading image error", err)
				return
			}

			imageLocations = append(imageLocations, location)
		}(inc)
	}
	wg.Wait() // Wait until all of the images have been uploaded.

	err = db.NewPost(form.Value["title"][0], form.Value["description"][0], imageLocations)
	if err != nil {
		helpers.SuccessResponse(false, w, r)
		helpers.ThrowErr(w, r, "Adding Post to DB error", err)
		return
	}

	helpers.SuccessResponse(true, w, r)
}

func uploadImage(file *multipart.FileHeader, uploader *s3manager.Uploader) (location string, err error) {
	image, err := file.Open()
	defer image.Close()
	if err != nil {
		return
	}

	imageID := uid.New(32)
	fileName := imageID + filepath.Ext(file.Filename)

	// Upload file to S3
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String("s.froogo.co.uk"),                               // Bucket name to upload (not necessarily domain)
		Key:    aws.String("Static/berniesbusybees.co.uk/img/" + fileName), // Directory to upload in S3
		Body:   bufio.NewReader(image),                                     // Body to upload (just bytes)
		ACL:    aws.String("public-read"),                                  // Set to public read (no key required to read)
	})

	url, err := url.Parse(result.Location)
	if err != nil {
		return
	}
	urlSplit := strings.Split(url.Path, "/")

	location = urlSplit[len(urlSplit)-1]
	return
}

// Delete deletes a post and removes all of the relevant images from S3.
func Delete(w http.ResponseWriter, r *http.Request) {
	var data models.PostDelete                   // Create struct to store data.
	err := json.NewDecoder(r.Body).Decode(&data) // Decode response to struct.
	if err != nil {
		helpers.SuccessResponse(false, w, r)
		helpers.ThrowErr(w, r, "JSON decoding error", err)
		return
	}

	if !middleware.AJAX(w, r, models.AJAXData{CsrfSecret: data.CsrfSecret}) {
		// Failed middleware (invalid credentials)
		helpers.SuccessResponse(false, w, r)
		return
	}

	uuidString := context.Get(r, "uuid").(string)
	uuid, err := strconv.Atoi(uuidString)
	if err != nil {
		helpers.SuccessResponse(false, w, r)
		helpers.ThrowErr(w, r, "Error converting string to int", err)
		return
	}

	user, err := db.GetUserFromID(uuid)
	if err != nil {
		helpers.SuccessResponse(false, w, r)
		helpers.ThrowErr(w, r, "Error getting user from ID", err)
		return
	}

	if user.Priv != models.PrivAdmin && user.Priv != models.PrivSuperAdmin {
		helpers.SuccessResponse(false, w, r)
		return
	}

	images, err := db.DeletePost(data.ID)
	if err != nil {
		helpers.SuccessResponse(false, w, r)
		helpers.ThrowErr(w, r, "Deleting post from DB error", err)
		return
	}

	svc := s3.New(session.Must(session.NewSession(&aws.Config{
		Region: aws.String("eu-west-2"),
	})))

	for inc := range images {
		go func(i int) {
			object := &s3.DeleteObjectInput{
				Bucket: aws.String("s.froogo.co.uk"),
				Key:    aws.String("Static/berniesbusybees.co.uk/img/" + images[i]),
			}

			_, err = svc.DeleteObject(object)
			if err != nil {
				helpers.SuccessResponse(false, w, r)
				helpers.ThrowErr(w, r, "Deleting object error", err)
				return
			}
		}(inc)
	}

	helpers.SuccessResponse(true, w, r)
}

// Update is an AJAX request response.
func Update(w http.ResponseWriter, r *http.Request) {
	var data models.PostEdit                     // Create struct to store data.
	err := json.NewDecoder(r.Body).Decode(&data) // Decode response to struct.
	if err != nil {
		helpers.SuccessResponse(false, w, r)
		helpers.ThrowErr(w, r, "JSON decoding error", err)
		return
	}

	if !middleware.AJAX(w, r, models.AJAXData{CsrfSecret: data.CsrfSecret}) {
		// Failed middleware (invalid credentials)
		helpers.SuccessResponse(false, w, r)
		return
	}

	uuidString := context.Get(r, "uuid").(string)
	uuid, err := strconv.Atoi(uuidString)
	if err != nil {
		helpers.ThrowErr(w, r, "Error converting string to int", err)
		return
	}

	user, err := db.GetUserFromID(uuid)
	if err != nil {
		helpers.ThrowErr(w, r, "Error getting user from ID", err)
		return
	}

	if user.Priv != models.PrivAdmin && user.Priv != models.PrivSuperAdmin {
		return
	}

	err = db.EditPost(data.ID, data.Title, data.Description)
	if err != nil {
		helpers.SuccessResponse(false, w, r)
		helpers.ThrowErr(w, r, "Editing menu item error", err)
		return
	}

	helpers.SuccessResponse(true, w, r)
}

// Comment is an AJAX request response.
func Comment(w http.ResponseWriter, r *http.Request) {
	var data models.NewComment                   // Create struct to store data.
	err := json.NewDecoder(r.Body).Decode(&data) // Decode response to struct.
	if err != nil {
		helpers.SuccessResponse(false, w, r)
		helpers.ThrowErr(w, r, "JSON decoding error", err)
		return
	}

	if !middleware.AJAX(w, r, models.AJAXData{CsrfSecret: data.CsrfSecret}) {
		// Failed middleware (invalid credentials)
		helpers.SuccessResponse(false, w, r)
		return
	}

	uuidString := context.Get(r, "uuid").(string)
	uuid, err := strconv.Atoi(uuidString)
	if err != nil {
		helpers.ThrowErr(w, r, "Error converting string to int", err)
		return
	}

	user, err := db.GetUserFromID(uuid)
	if err != nil {
		helpers.ThrowErr(w, r, "Error getting user from ID", err)
		return
	}

	if user.Priv != models.PrivUser && user.Priv != models.PrivAdmin && user.Priv != models.PrivSuperAdmin {
		return
	}

	data.Timestamp = time.Now().Unix()
	data.UserUUID = user.UUID

	id, err := db.AddCommentPost(data)
	if err != nil {
		helpers.SuccessResponse(false, w, r)
		helpers.ThrowErr(w, r, "Adding comment error", err)
		return
	}

	err = helpers.JSONResponse(models.ResponseWithID{
		Success: true,
		ID:      id,
	}, w)
	if err != nil {
		helpers.ThrowErr(w, r, "Sending JSON response error", err)
		return
	}
}

// CommentDelete is an AJAX request response.
func CommentDelete(w http.ResponseWriter, r *http.Request) {
	var data deleteCommentData                   // Create struct to store data.
	err := json.NewDecoder(r.Body).Decode(&data) // Decode response to struct.
	if err != nil {
		helpers.SuccessResponse(false, w, r)
		helpers.ThrowErr(w, r, "JSON decoding error", err)
		return
	}

	if !middleware.AJAX(w, r, models.AJAXData{CsrfSecret: data.CsrfSecret}) {
		// Failed middleware (invalid credentials)
		helpers.SuccessResponse(false, w, r)
		return
	}

	uuidString := context.Get(r, "uuid").(string)
	uuid, err := strconv.Atoi(uuidString)
	if err != nil {
		helpers.ThrowErr(w, r, "Error converting string to int", err)
		return
	}

	user, err := db.GetUserFromID(uuid)
	if err != nil {
		helpers.ThrowErr(w, r, "Error getting user from ID", err)
		return
	}

	if user.Priv != models.PrivAdmin && user.Priv != models.PrivSuperAdmin {
		owner, err := db.DeleteCommentPostIfOwner(data.CommentID, data.PostID, user.UUID)
		if err != nil {
			helpers.SuccessResponse(false, w, r)
			helpers.ThrowErr(w, r, "Adding comment error", err)
			return
		}

		if owner {
			helpers.SuccessResponse(true, w, r)
		} else {
			helpers.SuccessResponse(false, w, r) // The user didn't have valid permission.
		}

		return
	}

	err = db.DeleteCommentPost(data.CommentID, data.PostID)
	if err != nil {
		helpers.SuccessResponse(false, w, r)
		helpers.ThrowErr(w, r, "Adding comment error", err)
		return
	}

	helpers.SuccessResponse(true, w, r)
}
