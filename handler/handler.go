package handler

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/VolticFroogo/Bernies-Busy-Bees/db"
	"github.com/VolticFroogo/Bernies-Busy-Bees/handler/post"
	"github.com/VolticFroogo/Bernies-Busy-Bees/handler/recovery"
	"github.com/VolticFroogo/Bernies-Busy-Bees/handler/users"
	"github.com/VolticFroogo/Bernies-Busy-Bees/helpers"
	"github.com/VolticFroogo/Bernies-Busy-Bees/middleware"
	"github.com/VolticFroogo/Bernies-Busy-Bees/middleware/myJWT"
	"github.com/VolticFroogo/Bernies-Busy-Bees/models"
	"github.com/go-recaptcha/recaptcha"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
)

var (
	captchaSecret = os.Getenv("CAPTCHA_SECRET")
	captcha       = recaptcha.New(captchaSecret)
)

type loginData struct {
	Email, Password, Captcha string
}

// Start the server by handling the web server.
func Start() {
	r := mux.NewRouter()
	r.StrictSlash(true)

	r.Handle("/", http.HandlerFunc(index))

	r.Handle("/login", http.HandlerFunc(login)).Methods(http.MethodPost)

	r.Handle("/logout", negroni.New(
		negroni.HandlerFunc(middleware.Form),
		negroni.Wrap(http.HandlerFunc(logout)),
	))

	r.Handle("/panel", negroni.New(
		negroni.HandlerFunc(middleware.Panel),
		negroni.Wrap(http.HandlerFunc(panel)),
	))

	r.Handle("/panel/posts/{page}", negroni.New(
		negroni.HandlerFunc(middleware.Panel),
		negroni.Wrap(http.HandlerFunc(post.Posts)),
	))

	r.Handle("/panel/post/new", negroni.New(
		negroni.HandlerFunc(middleware.Panel),
		negroni.Wrap(http.HandlerFunc(post.NewPage)),
	)).Methods(http.MethodGet)

	r.Handle("/panel/post/new", negroni.New(
		negroni.HandlerFunc(middleware.Panel),
		negroni.Wrap(http.HandlerFunc(post.New)),
	)).Methods(http.MethodPost)

	r.Handle("/panel/settings/update", http.HandlerFunc(users.Settings))

	r.Handle("/panel/user/new", http.HandlerFunc(users.New))
	r.Handle("/panel/user/update", http.HandlerFunc(users.Update))
	r.Handle("/panel/user/delete", http.HandlerFunc(users.Delete))

	r.Handle("/panel/post/update", http.HandlerFunc(post.Update))
	r.Handle("/panel/post/delete", http.HandlerFunc(post.Delete))

	r.Handle("/panel/post/comment", http.HandlerFunc(post.Comment))
	r.Handle("/panel/post/comment/delete", http.HandlerFunc(post.CommentDelete))

	r.Handle("/panel/post/{postID}", negroni.New(
		negroni.HandlerFunc(middleware.Panel),
		negroni.Wrap(http.HandlerFunc(post.Post)),
	))

	r.Handle("/verify-email/{code}", http.HandlerFunc(users.VerifyEmail))
	r.Handle("/forgot-password", http.HandlerFunc(recovery.Begin)).Methods(http.MethodPost)
	r.Handle("/password-recovery", http.HandlerFunc(recovery.End)).Methods(http.MethodPost)

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	log.Printf("Server started...")
	http.ListenAndServe(":81", r)
}

func index(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("handler/templates/index.html", "handler/templates/nested.html") // Parse the HTML pages
	if err != nil {
		helpers.ThrowErr(w, r, "Template parsing error", err)
		return
	}

	variables := models.TemplateVariables{
		Posts: db.IndexPosts,
	}
	err = t.Execute(w, variables) // Execute temmplate with variables
	if err != nil {
		helpers.ThrowErr(w, r, "Template execution error", err)
	}
}

func panel(w http.ResponseWriter, r *http.Request) {
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

	switch user.Priv {
	case models.PrivUser, models.PrivAdmin, models.PrivSuperAdmin:
		execPanel(w, r, user, "panel")

	default:
		execPanel(w, r, user, "no-priv")
	}
}

func execPanel(w http.ResponseWriter, r *http.Request, user models.User, templateName string) {
	t, err := template.ParseFiles("handler/templates/panel/"+templateName+".html", "handler/templates/nested.html") // Parse the HTML pages
	if err != nil {
		helpers.ThrowErr(w, r, "Template parsing error", err)
		return
	}

	csrfSecret, err := r.Cookie("csrfSecret")
	if err != nil {
		helpers.ThrowErr(w, r, "Reading cookie error", err)
		return
	}

	posts, err := db.GetPosts(6, 6, 1)
	if err != nil {
		helpers.ThrowErr(w, r, "Getting posts error", err)
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

	variables := models.TemplateVariables{
		User:       user,
		CsrfSecret: csrfSecret.Value,
		Users:      db.Users,
		Posts:      posts,
	}
	err = t.Execute(w, variables) // Execute temmplate with variables
	if err != nil {
		helpers.ThrowErr(w, r, "Template execution error", err)
	}
}

func logout(w http.ResponseWriter, r *http.Request) {
	refreshTokenString, err := r.Cookie("refreshToken")
	if err != nil {
		helpers.ThrowErr(w, r, "Reading cookie error", err)
		return
	}

	myJWT.DeleteJTI(refreshTokenString.Value) // Remove their old Refresh Token.

	middleware.WriteNewAuth(w, r, "", "", "")

	middleware.RedirectToLogin(w, r)
}

func login(w http.ResponseWriter, r *http.Request) {
	var credentials loginData                           // Create struct to store data.
	err := json.NewDecoder(r.Body).Decode(&credentials) // Decode response to struct.
	if err != nil {
		helpers.SuccessResponse(false, w, r)
		helpers.ThrowErr(w, r, "JSON decoding error", err)
		return
	}

	if credentials.Captcha == "" {
		helpers.SuccessResponse(false, w, r)
		return // There is no captcha response.
	}
	captchaSuccess, err := captcha.Verify(credentials.Captcha, r.Header.Get("CF-Connecting-IP")) // Check the captcha.
	if err != nil {
		helpers.SuccessResponse(false, w, r)
		helpers.ThrowErr(w, r, "Recaptcha error", err)
		return
	}
	if !captchaSuccess {
		helpers.SuccessResponse(false, w, r)
		return // Unsuccessful captcha.
	}

	user, err := db.GetUserFromEmail(credentials.Email)
	if err != nil {
		helpers.SuccessResponse(false, w, r)
		helpers.ThrowErr(w, r, "Getting user from DB error", err)
		return
	}

	valid := helpers.CheckPassword(credentials.Password, user.Password)

	if valid {
		authTokenString, refreshTokenString, csrfSecret, err := myJWT.CreateNewTokens(strconv.Itoa(user.UUID))
		if err != nil {
			helpers.SuccessResponse(false, w, r)
			helpers.ThrowErr(w, r, "Creating tokens error", err)
			return
		}

		middleware.WriteNewAuth(w, r, authTokenString, refreshTokenString, csrfSecret)

		helpers.SuccessResponse(true, w, r)
		return
	}

	helpers.SuccessResponse(false, w, r)
}
