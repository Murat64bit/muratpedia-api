package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
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

// POST isteğini yönlendirme işlevi
func handlePostRequestRouting(ctx context.Context, request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	switch request.Path {
	case "/login":
		return handlers.LoginUser(ctx, request, *userCollection)
	case "/register":
		return handlers.RegisterUser(ctx, request, *userCollection)
	// case "/addArticle":
	// 	return handleAddArticleRequest(ctx, request)
	// case "/getUsers":
	// 	return handleGetUsersRequest(ctx, request)
	// case "/getArticles":
	// 	return handleGetArticlesRequest(ctx, request)
	// case "/getUserById":
	// 	return handleGetUserByID(ctx, request)
	// case "/getArticlesByTitle":
	// 	return handleGetArticlesByTitle(ctx, request)
	// case "/deleteUserById":
	// 	return handleDeleteUserByID(ctx, request)
	// case "/deleteArticleByTitle":
	// 	return handleDeleteArticleByTitle(ctx, request)
	default:
		return handlers.UnhandledMethod()
	}
}
