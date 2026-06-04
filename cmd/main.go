package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"go-blog/internal/handler"
	"go-blog/internal/middleware"
	"go-blog/internal/model"

	"github.com/gorilla/mux"
)

func main() {
	// 初始化数据库
	db, err := model.InitDB("data/blog.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// 创建路由器
	router := mux.NewRouter()

	// 静态文件服务
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// 路由处理器
	h := handler.NewHandler(db)

	// 页面路由
	router.HandleFunc("/", h.Index).Methods("GET")
	router.HandleFunc("/post/{id:[0-9]+}", h.GetPost).Methods("GET")
	router.HandleFunc("/about", h.About).Methods("GET")

	// API路由
	router.HandleFunc("/api/posts", h.GetPosts).Methods("GET")
	router.HandleFunc("/api/posts", h.CreatePost).Methods("POST")
	router.HandleFunc("/api/posts/{id:[0-9]+}", h.UpdatePost).Methods("PUT")
	router.HandleFunc("/api/posts/{id:[0-9]+}", h.DeletePost).Methods("DELETE")

	// 后台管理路由
	router.HandleFunc("/admin", h.Admin).Methods("GET")
	router.HandleFunc("/admin/login", h.Login).Methods("GET", "POST")
	router.HandleFunc("/admin/logout", h.Logout).Methods("GET")
	router.HandleFunc("/admin/posts/new", h.NewPost).Methods("GET", "POST")
	router.HandleFunc("/admin/posts/{id:[0-9]+}/edit", h.EditPost).Methods("GET", "POST")

	// 应用中间件
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	// 获取端口
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server starting on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
