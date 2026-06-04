package handler

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/gorilla/mux"
)

type Handler struct {
	db       *sql.DB
	template *template.Template
}

type PageData struct {
	Title   string
	Posts   []PostData
	Post    *PostData
	User    *UserData
	IsAdmin bool
}

type PostData struct {
	ID        int64
	Title     string
	Content   string
	Summary   string
	Author    string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserData struct {
	Username string
	Email    string
}

func NewHandler(db *sql.DB) *Handler {
	funcMap := template.FuncMap{
		"safe": func(s string) template.HTML {
			return template.HTML(s)
		},
	}
	tmpl := template.Must(template.New("").Funcs(funcMap).ParseGlob("internal/template/*.html"))
	return &Handler{db: db, template: tmpl}
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	posts, err := h.getPublishedPosts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := PageData{
		Title:   "首页",
		Posts:   posts,
		IsAdmin: false,
	}

	h.template.ExecuteTemplate(w, "index.html", data)
}

func (h *Handler) GetPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	post, err := h.getPostByID(id)
	if err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	// 转换Markdown为HTML
	post.Content = markdownToHTML(post.Content)

	data := PageData{
		Title: post.Title,
		Post:  post,
	}

	h.template.ExecuteTemplate(w, "post.html", data)
}

func (h *Handler) About(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title: "关于",
	}

	h.template.ExecuteTemplate(w, "about.html", data)
}

func (h *Handler) Admin(w http.ResponseWriter, r *http.Request) {
	posts, err := h.getAllPosts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := PageData{
		Title:   "后台管理",
		Posts:   posts,
		IsAdmin: true,
	}

	h.template.ExecuteTemplate(w, "admin.html", data)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		h.template.ExecuteTemplate(w, "login.html", nil)
		return
	}

	// 处理登录逻辑
	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "admin" && password == "admin123" {
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    "authenticated",
			Path:     "/",
			HttpOnly: true,
		})
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	http.Error(w, "Invalid credentials", http.StatusUnauthorized)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) NewPost(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		h.template.ExecuteTemplate(w, "post_form.html", nil)
		return
	}

	// 创建新文章
	title := r.FormValue("title")
	content := r.FormValue("content")
	summary := r.FormValue("summary")
	status := r.FormValue("status")

	if status == "" {
		status = "draft"
	}

	_, err := h.db.Exec(
		"INSERT INTO posts (title, content, summary, status) VALUES (?, ?, ?, ?)",
		title, content, summary, status,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *Handler) EditPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	if r.Method == "GET" {
		post, err := h.getPostByID(id)
		if err != nil {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}
		data := PageData{Post: post}
		h.template.ExecuteTemplate(w, "post_form.html", data)
		return
	}

	// 更新文章
	title := r.FormValue("title")
	content := r.FormValue("content")
	summary := r.FormValue("summary")
	status := r.FormValue("status")

	_, err = h.db.Exec(
		"UPDATE posts SET title=?, content=?, summary=?, status=?, updated_at=CURRENT_TIMESTAMP WHERE id=?",
		title, content, summary, status, id,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *Handler) GetPosts(w http.ResponseWriter, r *http.Request) {
	posts, err := h.getPublishedPosts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	var post PostData
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.db.Exec(
		"INSERT INTO posts (title, content, summary, status) VALUES (?, ?, ?, ?)",
		post.Title, post.Content, post.Summary, post.Status,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	post.ID, _ = result.LastInsertId()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)
}

func (h *Handler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	var post PostData
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = h.db.Exec(
		"UPDATE posts SET title=?, content=?, summary=?, status=?, updated_at=CURRENT_TIMESTAMP WHERE id=?",
		post.Title, post.Content, post.Summary, post.Status, id,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	post.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

func (h *Handler) DeletePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	_, err = h.db.Exec("DELETE FROM posts WHERE id=?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) getPublishedPosts() ([]PostData, error) {
	rows, err := h.db.Query(
		"SELECT id, title, content, summary, author, status, created_at, updated_at FROM posts WHERE status='published' ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []PostData
	for rows.Next() {
		var post PostData
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Summary, &post.Author, &post.Status, &post.CreatedAt, &post.UpdatedAt)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, nil
}

func (h *Handler) getAllPosts() ([]PostData, error) {
	rows, err := h.db.Query(
		"SELECT id, title, content, summary, author, status, created_at, updated_at FROM posts ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []PostData
	for rows.Next() {
		var post PostData
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Summary, &post.Author, &post.Status, &post.CreatedAt, &post.UpdatedAt)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, nil
}

func (h *Handler) getPostByID(id int64) (*PostData, error) {
	var post PostData
	err := h.db.QueryRow(
		"SELECT id, title, content, summary, author, status, created_at, updated_at FROM posts WHERE id=?",
		id,
	).Scan(&post.ID, &post.Title, &post.Content, &post.Summary, &post.Author, &post.Status, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func markdownToHTML(md string) string {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)

	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return string(markdown.ToHTML([]byte(md), p, renderer))
}
