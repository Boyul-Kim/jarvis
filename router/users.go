package router

import (
	"burnclub/tools"
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"google.golang.org/api/iterator"
)

type NewUser struct {
	Username string `json:"username"`
}

type GetUser struct {
	UserId string `form:"userId"`
	Name   string `form:"name"` //not sure if this field is necessary but will include it for now
}

type EditUser struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Image    string `json:"image"`
}

type LoginInfo struct {
	Uid      string `json:"uid"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
	Username string `json:"username"`
}

type SignUpInfo struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
	Username string `json:"username"`
}

func (s *Server) initUserRoutes(rg *gin.RouterGroup) {
	users := rg.Group("/users")
	users.POST("/sign/in", s.UserSignIn)
	users.GET("/details/self", s.FetchUserDetailSelf)
	users.PUT("/details/self/edit", s.EditUserData)
	users.GET("/details/other", s.FetchUser)
	users.GET("/search/all", s.SearchAllUsers)
}

func (s *Server) UserSignIn(c *gin.Context) {
	var bindData LoginInfo

	if err := c.ShouldBindBodyWith(&bindData, binding.JSON); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "Could not bind token data."})
		return
	}

	switch onboard := c.MustGet("OnboardType"); onboard {
	case "Google":
		query, _ := s.Database.Collection("users").Where("authId", "==", bindData.Uid).Documents(context.TODO()).GetAll()
		if query == nil {
			fmt.Println("User does not exist yet. Creating.", query)
			doc, _, err := s.Database.Collection("users").Add(context.TODO(), map[string]interface{}{
				"authId":         bindData.Uid,
				"email":          bindData.Email,
				"name":           bindData.Name,
				"onboardingType": onboard,
			})
			if err != nil {
				c.AbortWithStatusJSON(400, gin.H{"error": "Error trying to create new user."})
				return
			}
			c.JSON(200, gin.H{
				"data":   "Account successfully created!",
				"result": doc,
			})
			return
		}
		c.JSON(200, gin.H{"data": "Successfully signed in!"})
	case "Regular":
		query, err := s.Database.Collection("users").Where("authId", "==", bindData.Uid).Documents(context.TODO()).GetAll()
		if err != nil {
			c.AbortWithStatusJSON(404, gin.H{"error": "Error trying to find user."})
			return
		}
		verified := tools.VerifyPassword(bindData.Password, query[0].Data()["hashedPassword"].(string))
		if verified != nil {
			c.AbortWithStatusJSON(404, gin.H{"error": "Incorrect password."})
			return
		}
		token, tokenErr := tools.GenerateToken(query[0].Ref.ID)
		fmt.Println("TOKEN", token)
		if tokenErr != nil {
			c.AbortWithStatusJSON(404, gin.H{"error": "Error while trying to generate token."})
			return
		}
		c.JSON(200, gin.H{
			"message": "succesfully logged in!",
			"result":  token,
		})
	default:
		fmt.Println("ONBOARDING TYPE NOT SPECIFIED")
	}

}

func (s *Server) UserSignUp(c *gin.Context) {
	var bindData SignUpInfo
	if err := c.ShouldBindBodyWith(&bindData, binding.JSON); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "Could not bind token data."})
		return
	}
	switch onboard := c.MustGet("OnboardType"); onboard {
	case "Regular":
		query, _ := s.Database.Collection("users").Where("username", "==", bindData.Username).Documents(context.TODO()).GetAll()
		if query == nil {
			hashPassword := tools.HashPassword(bindData.Password)
			fmt.Println("HASH PASSWORD", hashPassword)
			authId := tools.AuthIdGenerator()
			fmt.Println("AUTH ID", authId)
			//need to create auth id generator for normal onboarding
			doc, _, err := s.Database.Collection("users").Add(context.TODO(), map[string]interface{}{
				"email":          bindData.Email,
				"name":           bindData.Name,
				"hashedPassword": string(hashPassword),
				"onboardType":    onboard,
				"authId":         authId,
			})
			if err != nil {
				c.AbortWithStatusJSON(400, gin.H{"error": "Error trying to create new user."})
				return
			}
			c.JSON(200, gin.H{
				"data":   "Account successfully created!",
				"result": doc,
			})
			return
		}
		c.JSON(404, gin.H{
			"data": "Account not found!",
		})
		return
	default:
		fmt.Println("ONBOARDING TYPE NOT SPECIFIED")
	}
}

func (s *Server) FetchUserDetailSelf(c *gin.Context) {
	userData := c.MustGet("AuthId").(string)
	query, _ := s.Database.Collection("users").Where("authId", "==", userData).Documents(context.TODO()).GetAll()
	c.JSON(200, gin.H{"data": query[0].Ref.ID})
}

func (s *Server) EditUserData(c *gin.Context) {
	var bindData EditUser
	if err := c.ShouldBindBodyWith(&bindData, binding.JSON); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "Could not bind token data."})
		return
	}
	ch := make(chan []*firestore.DocumentSnapshot)
	go func() {
		result, _ := s.Database.Collection("users").Where("authId", "==", c.MustGet("AuthId").(string)).Documents(context.TODO()).GetAll()
		ch <- result
	}()
	user := <-ch
	userDoc := user[0].Data()
	if userDoc == nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "Error fetching user."})
	}
	if len(bindData.Name) > 0 {
		userDoc["name"] = bindData.Name
	}
	if len(bindData.Username) > 0 {
		userDoc["username"] = bindData.Username
	}
	if len(bindData.Image) > 0 {
		userDoc["image"] = bindData.Image
	}
	result, err := s.Database.Collection("users").Doc(user[0].Ref.ID).Set(context.TODO(), userDoc)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "Error adding user into club.", "details": err})
		return
	}
	c.JSON(200, gin.H{"data": result})
}

func (s *Server) FetchUser(c *gin.Context) {
	var bindData GetUser
	if err := c.ShouldBindQuery(&bindData); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "Could not bind token data."})
	}
	query, err := s.Database.Collection("users").Doc(bindData.UserId).Get(context.TODO())
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "Error trying to fetch user."})
		return
	}
	if query == nil {
		c.JSON(200, gin.H{"data": "User does not exist"})
		return
	}
	c.JSON(200, gin.H{"data": query.Data()})
}

func (s *Server) SearchAllUsers(c *gin.Context) {
	iter := s.Database.Collection("users").Documents(context.TODO())
	var users []interface{}
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": "Error trying to fetch users."})
			return
		}
		formattedDoc := make(map[string]interface{})
		formattedDoc["name"] = doc.Data()["name"]
		formattedDoc["id"] = doc.Ref.ID
		users = append(users, formattedDoc)
	}
	c.JSON(200, gin.H{"data": users})
}
