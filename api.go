package main

import (
	"net/http"
	"strings"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"fmt"
	"time"
)


// AuthRequired is a simple middleware to check the session
func AuthRequired(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(userkey)
	if user == nil {
		// Abort the request with the appropriate error code
		c.Redirect(307, "/login")
		return
	}
	// Continue down the chain to handler etc
	c.Next()
}

//
func vote(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(userkey)
	stmt, err := db.Prepare("insert into votes(timestamp, voter, subject) values(?, ?, ?)")
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	subject := c.Param("subject")
	if subject == "" {
		c.JSON(400, gin.H{"error": "wrong path"})
		return
	}

	_, err = stmt.Exec(time.Now().Unix(), user, subject)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
}

func addUser(c *gin.Context)  {
	username := c.PostForm("fullName")
	login := c.PostForm("login")
	password := c.PostForm("password")
	avatar := c.PostForm("avatar")
	rows, err := db.Query(fmt.Sprintf("select * from users where login = '%s'", login))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if rows.Next() {
		c.JSON(400, gin.H{"error": "Login " + login + " in use"})
		return
	}

	stmt, err := db.Prepare("insert into users(login, fullName, password, avatar) values(?, ?,  ?, ?)")
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	hashPwd, _ := HashPassword(password)
	_, err = stmt.Exec(login, username, hashPwd, avatar)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status":"ok", "message": "user added"})
}

// login is a handler that parses a form and checks for specific data
func login(c *gin.Context) {
	session := sessions.Default(c)
	username := c.PostForm("username")
	password := c.PostForm("password")

	// Validate form input
	if strings.Trim(username, " ") == "" || strings.Trim(password, " ") == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parameters can't be empty"})
		return
	}

	hashPwd, _ := HashPassword(password)
	rows, err := db.Query(fmt.Sprintf("select * from users where login = '%s' and password = '%s'", username, hashPwd))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if  rows.Next() {
		session.Set(userkey, username) // In real world usage you'd set this to the users ID
		if err := session.Save(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
			return
		}
		c.Redirect(302, "/")
		return
	}

	c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
	// Save the username in the session
}

func logout(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(userkey)
	if user == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session token"})
		return
	}
	session.Delete(userkey)
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}

func me(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(userkey)
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func status(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "You are logged in"})
}