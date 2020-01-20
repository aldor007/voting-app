package main

import (
	"log"
	"net/http"
	"time"
	"fmt"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

const (
	userkey = "user"
)

// Thanks to otraore for the code example
// https://gist.github.com/otraore/4b3120aa70e1c1aa33ba78e886bb54f3
var db *sql.DB

func main() {
	fmt.Println("Server start")
	p, _ := HashPassword("password")
	fmt.Println(p)
	var err error
	db, err = sql.Open("sqlite3", "db/database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := engine()
	r.Use(gin.Logger())
	if err := engine().Run(":8080"); err != nil {
		log.Fatal("Unable to start:", err)
	}
}

type User struct {
	Login string
	FullName string
	Password string
	Avatar string
	Votes int
}

func engine() *gin.Engine {
	r := gin.New()
	r.Static("/assets", "./assets")
	r.LoadHTMLGlob("templates/*")
	r.Use(sessions.Sessions("session", sessions.NewCookieStore([]byte("secret"))))
	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})
	r.GET("/users", func(c *gin.Context) {
		c.HTML(http.StatusOK, "addUser.html", nil)
	})
	r.POST("/login", login)
	r.POST("/api/users", addUser)

	private := r.Group("")
	private.Use(AuthRequired)
	private.GET("/api/logout", logout)
	private.POST("/api/vote/:subject", vote)
	private.GET("/", func(c *gin.Context) {
		userData, err := db.Query("select login, fullName, avatar from users")
		var users []User
		for userData.Next() {
			user := User{}
			err = userData.Scan(&user.Login, &user.FullName, &user.Avatar)
			if user.Avatar == "" {
				user.Avatar = "https://media.gettyimages.com/photos/persian-picture-id537312023?s=612x612"
			}
			users = append(users, user)
		}
		tn := time.Now().UTC()
		year, week := tn.ISOWeek()
		startWeek := firstDayOfISOWeek(year, week, time.UTC)

		votesData, err := db.Query(fmt.Sprintf("select subject from votes where timestamp > %d", startWeek.Unix()))

		var subject string
		votesMap := make(map[string]int)
		for votesData.Next() {
			err = votesData.Scan(&subject)

			if v, ok := votesMap[subject]; ok {
				votesMap[subject] = v + 1
			} else {
				votesMap[subject] = 1
			}
		}

		lenUsers := len(users)
		for i := 0; i < lenUsers; i++ {
			if v, ok := votesMap[users[i].Login]; ok {
				users[i].Votes = v
			}
		}

		if err != nil {
			log.Fatal(err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "db error"})
			return
		}
		c.HTML(http.StatusOK, "index.html", gin.H{
			"users": users,
		})
	})
	{
		private.GET("/me", me)
		private.GET("/status", status)
	}
	return r
}
