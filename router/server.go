package router

import (
	"burnclub/db"
	"burnclub/tools"
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/auth"
	"firebase.google.com/go/messaging"
	"github.com/gin-gonic/gin"
)

type Server struct {
	Router   *gin.Engine
	Database *firestore.Client
	Auth     *auth.Client
	Msg      *messaging.Client
}

type UserToken struct {
	Token       string `header:"Authorization"`
	OnboardType string `header:"OnboardType"`
}

func SetupServer() *Server {
	// Disable Console Color
	// gin.DisableConsoleColor()

	// gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	firestoreClient, authClient, msgClient := db.ConnectDB(ctx)
	// defer firestoreClient.Close()

	burnclubServer := Server{
		Router:   r,
		Database: firestoreClient,
		Auth:     authClient,
		Msg:      msgClient,
	}
	burnclubServer.Router.Use(burnclubServer.TokenAuthMiddleware())

	v1 := burnclubServer.Router.Group("/api/v1")
	burnclubServer.initUserRoutes(v1)
	return &burnclubServer
}

func (s *Server) TokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var bindData UserToken
		if err := c.ShouldBindHeader(&bindData); err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": "Could not bind token data."})
		}
		c.Set("OnboardType", bindData.OnboardType)
		fmt.Println("BIND DATA", bindData)
		if bindData.OnboardType == "Google" {
			fmt.Println("TOKEN", bindData.Token)
			token, err := s.Auth.VerifyIDToken(context.TODO(), bindData.Token)
			if err != nil {
				c.AbortWithStatusJSON(401, gin.H{"error": "Incorrect token data."})
				return
			}
			c.Set("AuthId", token.UID)
			c.Set("Token", bindData.Token)
			c.Next()
		}
		if bindData.OnboardType == "Regular" {
			fmt.Println("CONTEXT", c.Request.URL)
			if c.Request.URL.RequestURI() == "/api/v1/users/sign/in" {
				c.Next()
				return
			}
			fmt.Println("TESTING", bindData.Token, bindData.OnboardType)
			userId, validate := tools.ValidateToken(bindData.Token)
			fmt.Println("VALIDATE", validate)
			if validate != nil {
				c.AbortWithStatusJSON(401, gin.H{"error": "Incorrect token data."})
				return
			}
			fmt.Println("USER ID TEST", userId)
			c.Set("AuthId", bindData.Token)
			c.Set("Token", bindData.Token)
			c.Set("UserId", userId)
			c.Next()
			return
		}
		if bindData.OnboardType == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "Onboard type not specificed."})
		}
		// c.Set("Name", token.Claims["name"])
		// c.Next()
		c.AbortWithStatusJSON(401, gin.H{"error": "Incorrect token data."})
		return
	}
}
