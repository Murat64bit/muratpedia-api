package main

import (
	"context"
	"log"

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

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

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

// POST isteğini yönlendirme işlevi
func handlePostRequestRouting(ctx context.Context, request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	switch request.Path {

	case "/login":
		return handlers.LoginUser(ctx, request, *userCollection)
	default:

		// JWT tokenını doğrula
		username, valid := jwt_auth.validateJWTandGetUserName(request)
		if !valid {
			return handlers.UnauthorizedResponse()
		}

		switch request.Path {
		case "/register":
			return handlers.RegisterUser(ctx, request, *userCollection)
		case "/addArticle":
			return handlers.AddArticle(ctx, request, *userCollection, username)
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
}
