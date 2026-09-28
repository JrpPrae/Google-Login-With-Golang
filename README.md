# Google Login with Golang

![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)
![OAuth 2.0](https://img.shields.io/badge/OAuth%202.0-OpenID%20Connect-EB5424?logo=openid&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Database-4169E1?logo=postgresql&logoColor=white)
![Google](https://img.shields.io/badge/Google-Sign--In-4285F4?logo=google&logoColor=white)

เว็บแอปพลิเคชันตัวอย่างสำหรับ **เข้าสู่ระบบด้วยบัญชี Google (Google Sign-In)** เขียนด้วยภาษา **Go** ผ่านมาตรฐาน **OAuth 2.0 + OpenID Connect (OIDC)** — เมื่อผู้ใช้ล็อกอินสำเร็จ ระบบจะดึงข้อมูลโปรไฟล์ (ชื่อ, นามสกุล) จาก Google มาบันทึกลงฐานข้อมูล **PostgreSQL** แล้วพาไปยังหน้า Homepage ที่แสดงข้อความต้อนรับด้วยชื่อของผู้ใช้

## สารบัญ

- [ภาพรวม](#ภาพรวม)
- [Workflow](#workflow)
- [Tech Stack](#tech-stack)
- [โครงสร้างโปรเจกต์](#โครงสร้างโปรเจกต์)
- [Routes / Endpoints](#routes--endpoints)
- [การตั้งค่า (config.yaml)](#การตั้งค่า-configyaml)
- [Database Schema](#database-schema)
- [วิธีใช้งาน](#วิธีใช้งาน)
- [ตัวอย่างโค้ดที่น่าสนใจ](#ตัวอย่างโค้ดที่น่าสนใจ)
- [หมายเหตุและสิ่งที่ควรพัฒนาต่อ](#หมายเหตุและสิ่งที่ควรพัฒนาต่อ)

## ภาพรวม

โปรเจกต์นี้แสดงการทำ **Authentication ด้วย Google** แบบ Authorization Code Flow โดยใช้ไลบรารีมาตรฐานของ Go (`net/http`) เป็น web server ร่วมกับ

- [`golang.org/x/oauth2`](https://pkg.go.dev/golang.org/x/oauth2) — สร้าง URL สำหรับ redirect ไปหน้า consent ของ Google และแลก `code` เป็น access token
- [`github.com/coreos/go-oidc/v3`](https://pkg.go.dev/github.com/coreos/go-oidc/v3/oidc) — ค้นหา endpoint ของ Google อัตโนมัติจาก Issuer (OIDC Discovery) และเรียก UserInfo endpoint เพื่อดึงข้อมูลผู้ใช้
- [`github.com/lib/pq`](https://pkg.go.dev/github.com/lib/pq) — driver สำหรับเชื่อมต่อ PostgreSQL ผ่าน `database/sql`
- [`gopkg.in/yaml.v3`](https://pkg.go.dev/gopkg.in/yaml.v3) — อ่านค่า config (Google credentials, database) จากไฟล์ `config.yaml`

โค้ดแบ่งเป็นชั้นชัดเจน: `config` (โหลดค่าตั้งค่า) → `database` (เชื่อมต่อ/บันทึกข้อมูล) → `handler` (จัดการ HTTP request) → `web_ui` (หน้า HTML)

## Workflow

### ภาพรวมการทำงานตั้งแต่เริ่มโปรแกรม

```text
go run main.go
    │
    ├─ 1. config.LoadConfig()          อ่าน config.yaml → เก็บใน config.AppConfig
    ├─ 2. database.SetupDatabase()     เปิด connection PostgreSQL + Ping ตรวจสอบ
    ├─ 3. google_handler.SetupGoogleAuth()
    │        ├─ oidc.NewProvider(Issuer)   ดึง OIDC Discovery ของ Google
    │        └─ สร้าง oauth2.Config (ClientID, Secret, RedirectURL, Scopes)
    ├─ 4. ลงทะเบียน Routes (/ , /google_login , /callback , /homepage)
    └─ 5. http.ListenAndServe(":8080")
```

### ขั้นตอนการ Login (OAuth 2.0 Authorization Code Flow)

```text
 ผู้ใช้ (Browser)            Go Server (:8080)                 Google                  PostgreSQL
      │                            │                             │                          │
      │── GET / ──────────────────▶│                             │                          │
      │◀── login.html ─────────────│                             │                          │
      │                            │                             │                          │
      │── คลิก "Login with Google" │                             │                          │
      │── GET /google_login ──────▶│                             │                          │
      │◀── 302 Redirect ───────────│  AuthCodeURL(state)         │                          │
      │                            │                             │                          │
      │── หน้า Consent / เลือกบัญชี ──────────────────────────────▶│                          │
      │◀── 302 Redirect → /callback?code=...&state=... ──────────│                          │
      │                            │                             │                          │
      │── GET /callback?code=... ─▶│                             │                          │
      │                            │── Exchange(code) ──────────▶│                          │
      │                            │◀── access token + id token ─│                          │
      │                            │── UserInfo(token) ─────────▶│                          │
      │                            │◀── claims (given_name, ...) │                          │
      │                            │                             │                          │
      │                            │── INSERT INTO users ... ───────────────────────────────▶│
      │                            │◀── OK / Error ─────────────────────────────────────────│
      │                            │                             │                          │
      │◀── 303 Redirect ───────────│  สำเร็จ: /homepage?success=true&name=...&family=...     │
      │                            │  ล้มเหลว: /homepage?success=false                        │
      │                            │                             │                          │
      │── GET /homepage?... ──────▶│                             │                          │
      │◀── homepage.html ──────────│  "Welcome, <ชื่อ นามสกุล>!!" หรือข้อความ error             │
```

### รายละเอียดแต่ละขั้น

1. **หน้า Login (`/`)** — แสดง [`web_ui/login.html`](web_ui/login.html) ซึ่งมีลิงก์ `Login with Google` ชี้ไปที่ `/google_login` (ถ้า path ไม่ใช่ `/` พอดี จะตอบ 404 แทน)
2. **Redirect ไป Google (`/google_login`)** — สร้าง URL ของ Google ด้วย `oauthConfig.AuthCodeURL(state)` แล้ว redirect (HTTP 302) ผู้ใช้ไปยังหน้า consent
3. **Callback (`/callback`)** — Google ส่งผู้ใช้กลับมาพร้อม `code`
   - ถ้ามี query `error` (เช่น ผู้ใช้กดยกเลิก) → ตอบ HTTP 500 พร้อมข้อความ error
   - แลก `code` เป็น token ด้วย `oauthConfig.Exchange()`
   - เรียก `provider.UserInfo()` เพื่อดึงข้อมูลผู้ใช้ แล้ว map claims ลง struct `models.GoogleInformationResponse`
   - บันทึกลงฐานข้อมูลด้วย `database.SaveToDB()` (ฟิลด์ที่ว่างจะถูกบันทึกเป็น `NULL`)
4. **Homepage (`/homepage`)** — อ่าน query `success`, `name`, `family` แล้ว render [`web_ui/homepage.html`](web_ui/homepage.html) ด้วย `html/template`
   - `success=true` → แสดง `Welcome, <ชื่อ นามสกุล>!!`
   - `success=false` → แสดง `Login not success, Please try again` (ตัวอักษรสีแดง)
   - ไม่มีพารามิเตอร์ (เข้าหน้าโดยตรง) → redirect กลับไปหน้า Login

## Tech Stack

| หมวด | เทคโนโลยี | หน้าที่ในโปรเจกต์ |
| --- | --- | --- |
| ภาษา | **Go 1.25** | ภาษาหลักของ backend |
| Web Server | `net/http` (Standard Library) | ทำ routing และให้บริการ HTTP ที่พอร์ต `8080` |
| Template Engine | `html/template` (Standard Library) | render หน้า Homepage พร้อม escape ข้อมูลอัตโนมัติ (กัน XSS) |
| Authentication | **OAuth 2.0** — `golang.org/x/oauth2` v0.28.0 | สร้าง Authorization URL และแลก code → token |
| Identity | **OpenID Connect** — `github.com/coreos/go-oidc/v3` v3.17.0 | OIDC Discovery และเรียก UserInfo endpoint |
| Identity Provider | **Google Identity Platform** | ผู้ให้บริการยืนยันตัวตน (Issuer: `https://accounts.google.com`) |
| Database | **PostgreSQL** | เก็บข้อมูลผู้ใช้ที่ล็อกอินผ่าน Google |
| DB Driver | `github.com/lib/pq` v1.11.2 | PostgreSQL driver สำหรับ `database/sql` |
| Config | `gopkg.in/yaml.v3` v3.0.1 | อ่านไฟล์ตั้งค่า `config.yaml` |
| Frontend | HTML (ไม่มี CSS framework / JavaScript) | หน้า Login และหน้า Homepage แบบเรียบง่าย |

> ใน `go.mod` ยังมี `github.com/gofiber/fiber/v2` และ `github.com/joho/godotenv` อยู่ แต่โค้ดปัจจุบันไม่ได้ใช้งานแล้ว (เคยใช้ `.env` ก่อนย้ายมาใช้ `config.yaml`)

## โครงสร้างโปรเจกต์

```text
Google-Login-with-Golang/
├── main.go                          # จุดเริ่มโปรแกรม: โหลด config, ต่อ DB, ตั้งค่า OAuth, ลงทะเบียน routes, start server
├── go.mod / go.sum                  # Go module (ชื่อ module: google_login) และ dependencies
├── config.yaml                      # (ต้องสร้างเอง, ถูก .gitignore) เก็บ Google credentials และค่าเชื่อมต่อ DB
│
├── config/
│   └── config.go                    # struct Config + LoadConfig() อ่าน config.yaml ด้วย yaml.v3
│
├── database/
│   ├── set_db.go                    # SetupDatabase(): สร้าง connection string, sql.Open("postgres"), Ping
│   ├── query.go                     # SaveToDB(): INSERT ข้อมูลผู้ใช้ลงตาราง users (รองรับค่า NULL)
│   └── get_info.go                  # Google_Response(): log ผลลัพธ์การ login ในรูปแบบ status/message/data
│
├── handler/
│   ├── google_handler/
│   │   └── auth.go                  # SetupGoogleAuth(), GoogleLoginHandler(), CallbackHandler()
│   └── web_handler/
│       └── homepage.go              # Homepage_handler(): อ่าน query แล้ว render homepage.html
│
├── models/
│   ├── google.go                    # Google_request — ค่าที่ใช้ตั้งค่า OAuth
│   ├── database.go                  # ConnDB — ค่าที่ใช้เชื่อมต่อ DB
│   └── user.go                      # GoogleInformationResponse — ข้อมูลผู้ใช้ที่ได้จาก Google
│
└── web_ui/
    ├── login.html                   # หน้า Login (ลิงก์ไป /google_login)
    └── homepage.html                # Go template แสดงข้อความต้อนรับ / error
```

## Routes / Endpoints

| Method | Path | Handler | คำอธิบาย |
| --- | --- | --- | --- |
| `GET` | `/` | inline ใน `main.go` | แสดงหน้า Login (`web_ui/login.html`) — path อื่นที่ไม่ตรงจะได้ 404 |
| `GET` | `/google_login` | `google_handler.GoogleLoginHandler` | Redirect ไปหน้า consent ของ Google |
| `GET` | `/callback` | `google_handler.CallbackHandler` | รับ `code` จาก Google → แลก token → ดึง UserInfo → บันทึก DB → redirect ไป `/homepage` |
| `GET` | `/homepage` | `webhandler.Homepage_handler` | แสดงผลการ Login ตาม query `success`, `name`, `family` |

## การตั้งค่า (config.yaml)

สร้างไฟล์ `config.yaml` ไว้ที่ root ของโปรเจกต์ (ไฟล์นี้อยู่ใน `.gitignore` จึงไม่ถูก commit ขึ้น Git):

```yaml
google:
  ClientID: "xxxxxxxxxxxx-xxxxxxxx.apps.googleusercontent.com"
  ClientSecret: "GOCSPX-xxxxxxxxxxxxxxxx"
  ISSUER: "https://accounts.google.com"
  REDIRECT_URL: "http://localhost:8080/callback"

database:
  HOST: "localhost"
  PORT: "5432"
  DATABASENAME: "google_login"
  USER_NAME: "postgres"
  PASSWORD: "your_password"
```

| Key | คำอธิบาย |
| --- | --- |
| `google.ClientID` | OAuth 2.0 Client ID จาก Google Cloud Console |
| `google.ClientSecret` | OAuth 2.0 Client Secret |
| `google.ISSUER` | URL ของ OIDC Issuer — สำหรับ Google ใช้ `https://accounts.google.com` |
| `google.REDIRECT_URL` | URL ที่ Google จะส่งผู้ใช้กลับมา ต้องตรงกับที่ลงทะเบียนไว้ใน Google Cloud Console |
| `database.HOST` / `PORT` | ที่อยู่และพอร์ตของ PostgreSQL |
| `database.DATABASENAME` | ชื่อฐานข้อมูล |
| `database.USER_NAME` / `PASSWORD` | บัญชีผู้ใช้สำหรับเชื่อมต่อฐานข้อมูล |

> การเชื่อมต่อใช้ `sslmode=disable` ซึ่งเหมาะกับการพัฒนาบนเครื่อง (local) เท่านั้น

## Database Schema

โค้ดบันทึกข้อมูลด้วยคำสั่ง

```sql
INSERT INTO users(pid, given_name, family_name, birthdate) VALUES ($1, $2, $3, $4)
```

จึงต้องสร้างตาราง `users` ไว้ก่อน ตัวอย่าง schema ที่ใช้ได้:

```sql
CREATE TABLE users (
    id           SERIAL PRIMARY KEY,
    pid          VARCHAR(255),          -- อนุญาต NULL
    given_name   VARCHAR(255),
    family_name  VARCHAR(255),
    birthdate    VARCHAR(20),           -- อนุญาต NULL
    created_at   TIMESTAMP DEFAULT NOW()
);
```

| คอลัมน์ | ที่มา (claim จาก Google) | หมายเหตุ |
| --- | --- | --- |
| `pid` | `pid` | ถ้าว่างจะบันทึกเป็น `NULL` |
| `given_name` | `given_name` | ชื่อจริง |
| `family_name` | `family_name` | นามสกุล |
| `birthdate` | `birthdate` | ถ้าว่างจะบันทึกเป็น `NULL` |

> ใน `SaveToDB()` ถ้า `pid` **หรือ** `birthdate` ตัวใดตัวหนึ่งว่าง ทั้งสองคอลัมน์จะถูกบันทึกเป็น `NULL` พร้อมกัน

## วิธีใช้งาน

### 1. สิ่งที่ต้องมี

- [Go](https://go.dev/dl/) เวอร์ชัน 1.25 ขึ้นไป
- [PostgreSQL](https://www.postgresql.org/download/) ที่รันอยู่บนเครื่องหรือเข้าถึงได้
- บัญชี [Google Cloud](https://console.cloud.google.com/)

### 2. สร้าง OAuth Client บน Google Cloud Console

1. เข้า **Google Cloud Console** → สร้างหรือเลือก Project
2. ไปที่ **APIs & Services → OAuth consent screen** ตั้งค่าชื่อแอปและเพิ่ม scope `openid`, `profile`, `email` (ระหว่างทดสอบให้เพิ่มอีเมลตัวเองเป็น Test user)
3. ไปที่ **APIs & Services → Credentials → Create Credentials → OAuth client ID**
   - Application type: **Web application**
   - Authorized redirect URIs: `http://localhost:8080/callback`
4. คัดลอก **Client ID** และ **Client Secret** ไปใส่ใน `config.yaml`

### 3. เตรียมฐานข้อมูล

```bash
psql -U postgres -c "CREATE DATABASE google_login;"
psql -U postgres -d google_login -c "CREATE TABLE users (id SERIAL PRIMARY KEY, pid VARCHAR(255), given_name VARCHAR(255), family_name VARCHAR(255), birthdate VARCHAR(20), created_at TIMESTAMP DEFAULT NOW());"
```

### 4. ติดตั้ง dependencies และรันโปรแกรม

```bash
cd Google-Login-with-Golang
go mod tidy
go run main.go
```

ถ้าสำเร็จจะเห็นข้อความใน terminal:

```text
Successfully connected to Database on port 5432
Starting server at :8080...
```

> ต้องรันจาก root ของโปรเจกต์ เพราะโปรแกรมอ่าน `config.yaml` และ `web_ui/*.html` ด้วย relative path

### 5. ทดสอบ

1. เปิดเบราว์เซอร์ไปที่ <http://localhost:8080>
2. คลิก **Login with Google** แล้วเลือกบัญชี Google
3. เมื่อสำเร็จจะถูกพาไปที่ `/homepage` และเห็นข้อความ `Welcome, <ชื่อ นามสกุล>!!`
4. ใน terminal จะมี log ของการบันทึกข้อมูล เช่น

```text
Save user to Database successfully
map[data:{ John Doe } message:Login successful status:success]
Database save success, redirecting to homepage...
```

## ตัวอย่างโค้ดที่น่าสนใจ

**ตั้งค่า OAuth ผ่าน OIDC Discovery** — ไม่ต้อง hardcode URL ของ Google เพราะ `go-oidc` ดึง endpoint ให้จาก Issuer ([handler/google_handler/auth.go](handler/google_handler/auth.go))

```go
provider, err = oidc.NewProvider(ctx, gg_env.Issuer)

oauthConfig = &oauth2.Config{
    ClientID:     gg_env.ClientID,
    ClientSecret: gg_env.ClientSecret,
    Endpoint:     provider.Endpoint(),
    RedirectURL:  gg_env.Redirect_url,
    Scopes:       []string{oidc.ScopeOpenID, "openid profile email"},
}
```

**แลก code เป็น token และดึงข้อมูลผู้ใช้** ([handler/google_handler/auth.go](handler/google_handler/auth.go))

```go
oauth2Token, err := oauthConfig.Exchange(ctx, code)
userInfo, err := provider.UserInfo(ctx, oauth2.StaticTokenSource(oauth2Token))

var info models.GoogleInformationResponse
err = userInfo.Claims(&info)
```

**รองรับค่า NULL ในฐานข้อมูลด้วย `sql.NullString`** ([database/query.go](database/query.go))

```go
if u.BirthDate == "" || u.PID == "" {
    birthDate = sql.NullString{Valid: false} // จะกลายเป็น NULL ใน Database
    pid = sql.NullString{Valid: false}
}
```

**ส่งข้อมูลไปหน้า Homepage อย่างปลอดภัย** — ใช้ `url.QueryEscape` ตอนสร้าง redirect URL และ `html/template` escape ค่าอัตโนมัติตอน render ([handler/web_handler/homepage.go](handler/web_handler/homepage.go))

```html
{{if .IsSuccess}}
    <h1>Welcome, {{.Fullname}}!!</h1>
{{else}}
    <h1 style="color: red;">{{.ErrorMessage}}</h1>
{{end}}
```

## หมายเหตุและสิ่งที่ควรพัฒนาต่อ

โปรเจกต์นี้เป็นตัวอย่างเพื่อการเรียนรู้ ก่อนนำไปใช้งานจริงควรปรับปรุงประเด็นต่อไปนี้:

- **State parameter** — ตอนนี้ใช้ค่าคงที่ `"random_state_string"` และไม่ได้ตรวจสอบใน `/callback` ควรสุ่มค่าใหม่ทุกครั้ง เก็บไว้ใน cookie/session และตรวจสอบตอน callback เพื่อป้องกัน CSRF
- **Session / สถานะการล็อกอิน** — ยังไม่มีการสร้าง session หรือ cookie หลังล็อกอิน หน้า `/homepage` อ่านชื่อจาก query string โดยตรง ผู้ใช้จึงสามารถพิมพ์ URL เช่น `/homepage?success=true&name=abc` เพื่อแสดงหน้าได้เองโดยไม่ต้องล็อกอิน
- **Claims ที่ได้จาก Google** — UserInfo ของ Google ระบุตัวผู้ใช้ด้วย `sub` (ไม่ใช่ `pid`) และไม่ส่ง `birthdate` มาใน scope `profile`/`email` ดังนั้นในทางปฏิบัติ `pid` และ `birthdate` จะเป็น `NULL` เสมอ หากต้องการเก็บ ID ผู้ใช้ ควรเปลี่ยน tag เป็น `json:"sub"` (และอาจเก็บ `email` เพิ่ม)
- **ผู้ใช้ซ้ำ** — ทุกครั้งที่ล็อกอินจะ `INSERT` แถวใหม่ ควรตั้ง unique key (เช่น `sub` หรือ `email`) แล้วใช้ `INSERT ... ON CONFLICT DO UPDATE` แทน
- **Scopes** — ค่า `"openid profile email"` เป็น string เดียวที่มีช่องว่าง ใช้งานได้เพราะ library นำ scopes มาต่อกันด้วยช่องว่าง แต่ `openid` ซ้ำกับ `oidc.ScopeOpenID` ควรเขียนเป็น `[]string{oidc.ScopeOpenID, "profile", "email"}`
- **ID Token verification** — ยังไม่ได้ verify `id_token` ด้วย `provider.Verifier()` ซึ่งเป็นแนวทางที่แนะนำของ OIDC
- **Error handling** — กรณีผู้ใช้กดยกเลิกที่หน้า consent ปัจจุบันตอบ HTTP 500 อาจเปลี่ยนเป็น redirect กลับหน้า Login พร้อมข้อความที่เป็นมิตรกว่า และ `log.Fatal` ใน `config.go` ควรเปลี่ยนเป็น `log.Fatalf` เพื่อให้ format `%v` ทำงาน
- **Dependencies** — ทุกแพ็กเกจใน `go.mod` ถูกทำเครื่องหมาย `// indirect` และมีแพ็กเกจที่ไม่ได้ใช้ (`fiber`, `godotenv`) รัน `go mod tidy` เพื่อจัดระเบียบ
- **ไฟล์ `go`** — ไฟล์ว่างที่ root ของโปรเจกต์ ไม่มีผลต่อการทำงานและลบทิ้งได้
- **Secrets** — อย่า commit `config.yaml` ขึ้น Git (ถูกกันไว้ใน `.gitignore` แล้ว) และใน production ควรใช้ HTTPS กับ `sslmode=require`
