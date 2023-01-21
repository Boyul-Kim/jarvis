package tools

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	// "github.com/golang-jwt/jwt/v4"
	jwt "github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) []byte {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil
	}
	return hashedPassword
}

func VerifyPassword(password string, passwordHash string) error {
	comparison := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	fmt.Println("COMPARISON", comparison, []byte(passwordHash), []byte(password))
	return comparison
}

// for regular onboardtype
func AuthIdGenerator() string {
	var alphaNumerics = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z", "a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}
	var id []string
	for i := 0; i < 29; i++ {
		rand.Seed(time.Now().UnixNano())
		randNum := rand.Intn((len(alphaNumerics) - 0) + 0)
		id = append(id, alphaNumerics[randNum])
	}
	return strings.Join(id, "")
}

func GenerateToken(userId string) (string, error) {
	tokenLifespan, err := strconv.Atoi(goDotEnvVariable("TOKEN_EXPIRATION"))
	if err != nil {
		return "", err
	}
	claims := jwt.MapClaims{}
	claims["authorized"] = true
	claims["userId"] = userId
	claims["exp"] = time.Now().Add(time.Minute * time.Duration(tokenLifespan)).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("TOKEN_SECRET")))
}

func ValidateToken(token string) (string, error) {
	// extractedToken := ExtractBearerToken(token) //might need this if the client includes bearer in front of token
	extractedToken := token
	validatedToken, err := jwt.Parse(extractedToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(goDotEnvVariable("TOKEN_SECRET")), nil
	})
	if err != nil {
		return "", err
	}
	claims, ok := validatedToken.Claims.(jwt.MapClaims)
	fmt.Println("CLAIMS", claims["userId"])
	if ok && validatedToken.Valid {
		uid := claims["userId"]
		if uid == nil || uid == "" {
			return "", err
		}
		return uid.(string), nil
	}
	return "", nil
}

func ExtractBearerToken(bearerToken string) string {
	if len(strings.Split(bearerToken, " ")) == 2 {
		return strings.Split(bearerToken, " ")[1]
	}
	return ""
}
