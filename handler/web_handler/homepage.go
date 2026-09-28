package webhandler

import (
	"html/template"
	"log"
	"net/http"
)

func Homepage_handler(w http.ResponseWriter, r *http.Request) {
	// 1. ดึงค่าที่ส่งมาจาก Redirect
	success := r.URL.Query().Get("success")
	name := r.URL.Query().Get("name")
	family := r.URL.Query().Get("family")

	// 2. สร้างโครงสร้างข้อมูลสำหรับส่งให้ HTML
	show_data := struct {
		Name         string
		Family_name  string
		Fullname     string
		ErrorMessage string
		IsSuccess    bool
	}{
		IsSuccess: success == "true",
	}

	// 3. เช็คเงื่อนไขการแสดงผล
	if success == "false" {
		// กรณี Login/Save ไม่สำเร็จ
		show_data.ErrorMessage = "Login not success, Please try again"
	} else if name != "" {
		// กรณีสำเร็จ
		show_data.Name = name
		show_data.Family_name = family
		show_data.Fullname = name + " " + family
	} else {
		// กรณีแอบเข้าหน้าตรงๆ โดยไม่มีพารามิเตอร์ ให้ดีดกลับหน้า Login
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// 4. แสดงผลไฟล์ Template
	tmpl, err := template.ParseFiles("web_ui/homepage.html")
	if err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	tmpl.Execute(w, show_data)
}
