package authing

import (
	"log"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/dgrijalva/jwt-go"
	"github.com/murat64bit/muratpedia-api/pkg/user"
)

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

var (
	jwtKey = []byte("my_secret_key")
)

func validateJWTandGetUserName(request events.APIGatewayProxyRequest) (string, bool) {
	tokenString := request.Headers["Authorization"]
	if tokenString == "" {
		return "", false
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		return "", false
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return "", false
	}

	return claims.Username, true
}

func generateJWT(user user.UserData) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: user.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		log.Println("JWT Signing error:", err)
		return "", err
	}

	return tokenString, nil
}