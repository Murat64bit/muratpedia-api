package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/dgrijalva/jwt-go"
	"github.com/murat64bit/muratpedia-api/pkg/handlers"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	dbClient          *mongo.Client
	userCollection    *mongo.Collection
	articleCollection *mongo.Collection
	jwtKey            = []byte("my_secret_key")
)

func main() {
	mongoURI := "mongodb+srv://murat64bit:muratturk123@cluster1.afbta3i.mongodb.net/"
	dbName := "muratpedia"
	userCollName := "user"
	articleCollName := "article"

	// MongoDB istemcisini oluşturun
	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal("MongoDB connection error:", err)
	}
	defer client.Disconnect(context.Background())

	// MongoDB'ye başarılı bir şekilde bağlandığınızı kontrol edin
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal("MongoDB ping error:", err)
	}

	// MongoDB koleksiyonunu seçin
	userCollection = client.Database(dbName).Collection(userCollName)
	articleCollection = client.Database(dbName).Collection(articleCollName)

	// AWS Lambda işlevini başlatın
	lambda.Start(handlePostRequestRouting)
}

func handlePostRequestRouting(ctx context.Context, request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	// JWT tokenını al
	tokenString := request.Headers["Authorization"]
	if tokenString == "" {
		return handlers.UnauthorizedResponse()
	}

	// JWT tokenını doğrula
	valid, _ := validateJWT(tokenString)
	if !valid {
		return handlers.UnauthorizedResponse()
	}

	// İstek yolu (Path) üzerindeki işlemi belirle ve ilgili işlevi çağır
	switch request.Path {
	case "/login":
		return handlers.LoginUser(ctx, request, *userCollection)
	case "/register":
		return handlers.RegisterUser(ctx, request, *userCollection)
	case "/addArticle":
		return handlers.AddArticle(ctx, request, *userCollection)
	case "/getUsers":
		return handlers.GetUsers(ctx, request, *userCollection)
	case "/getArticles":
		return handlers.GetArticles(ctx, request, *userCollection)
	case "/getUserById":
		return handlers.GetUserById(ctx, request, *userCollection)
	case "/getArticlesByTitle":
		return handlers.GetArticlesByTitle(ctx, request, *userCollection)
	case "/deleteUserById":
		return handlers.DeleteUserById(ctx, request, *userCollection)
	case "/deleteArticleByTitle":
		return handlers.DeleteArticleByTitle(ctx, request, *userCollection)
	default:
		return handlers.UnhandledMethod()
	}
}

// JWT doğrulama işlemi
func validateJWT(tokenString string) (bool, string) {
	// JWT tokenini doğrula
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// JWT'nin doğrulama anahtarını kullanarak doğrula
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Invalid token")
		}
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		return false, ""
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false, ""
	}

	username, ok := claims["username"].(string)
	if !ok {
		return false, ""
	}

	return true, username
}

// JWT oluşturma işlemi
func generateJWT(userID string) (string, error) {
	// JWT tokeni oluşturmak için bir dizi içeriği hazırlayın
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour * 24).Unix(), // Tokenin geçerlilik süresi (örnekte 1 gün)
	}

	// Token oluşturucu yapılandırması
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Tokeni imzalayarak string olarak döndürün
	signedToken, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}
