package main

import (
	"fmt"
	"google_login/config"
	"google_login/database"
	"google_login/handler/google_handler"
	webhandler "google_login/handler/web_handler"
	"log"
	"net/http"
)

func main() {

	// load env and setup db
	// _ = godotenv.Load()
	config.LoadConfig()
	database.SetupDatabase()
	google_handler.SetupGoogleAuth()

	// 1. Route: หน้าแรกสำหรับปุ่ม Login
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r) // ถ้าหา path ไม่เจอให้ขึ้น 404 แทนที่จะโชว์หน้า login
			return
		}
		http.ServeFile(w, r, "web_ui/login.html")
	})

	// 2. Route ส่งไปหา Google
	http.HandleFunc("/google_login", google_handler.GoogleLoginHandler)

	// 3. Route: /callback (รับกลับจาก Google และ Save DB )
	http.HandleFunc("/callback", google_handler.CallbackHandler)

	// 4. แสดงหน้า Homepage เมื่อ login success
	http.HandleFunc("/homepage", webhandler.Homepage_handler)

	fmt.Println("Starting server at :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
