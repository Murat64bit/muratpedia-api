package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/go-playground/validator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

// MongoDB bağlantısı için küresel değişkenler
var (
	dbClient   *mongo.Client
	userCollection *mongo.Collection
	articleCollection *mongo.Collection
	validate   *validator.Validate
)

type UserData struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type ArticleData struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description" validate:"required"`
	Author      string `json:"author" validate:"required"`
	Date        string `json:"date" validate:""`
}

// POST isteğini yönlendirme işlevi
func handlePostRequestRouting(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch request.Path {
	case "/createUser":
		return handleCreateUserRequest(ctx, request)
	case "/addArticle":
		return handleAddArticleRequest(ctx, request)
	default:
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNotFound,
			Body:       "Not Found",
		}, nil
	}
}

// GET isteğini işlemek için kullanılan işlev
func handleGetRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// MongoDB'den veri alın
	filter := bson.M{}
	cursor, err := userCollection.Find(ctx, filter)
	if err != nil {
		log.Println("Find error:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Internal Server Error",
		}, nil
	}
	defer cursor.Close(ctx)

	var result []bson.M
	if err := cursor.All(ctx, &result); err != nil {
		log.Println("Cursor All error:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Internal Server Error",
		}, nil
	}

	// Elde edilen verileri JSON formatında döndür
	responseBody, err := json.Marshal(result)
	if err != nil {
		log.Println("JSON Marshal error:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Internal Server Error",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(responseBody),
	}, nil
}



// AddArticle isteğini işlemek için kullanılan işlev
func handleAddArticleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	currentTime := time.Now()

	// İstek gövdesinden veriyi çözümle
	var articleData ArticleData
	if err := json.Unmarshal([]byte(request.Body), &articleData); err != nil {
		log.Println("JSON Unmarshal error:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Bad Request",
		}, nil
	}

	// Veri doğrulama işlemini gerçekleştir
	if err := validate.Struct(articleData); err != nil {
		log.Println("Validation error:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Validation Error",
		}, nil
	}

	data := bson.M{
		"title":       articleData.Title,
		"description": articleData.Description,
		"author":      articleData.Author,
		"date":        currentTime.Format("02-01-2006"),
	}

	// Veriyi MongoDB'ye ekle
	_, err := articleCollection.InsertOne(ctx, data)
	if err != nil {
		log.Println("InsertOne error:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Internal Server Error",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusCreated,
		Body:       "Article created",
	}, nil
}

func main() {
	// MongoDB bağlantısı için gerekli parametreleri ayarlayın
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

	// Validator oluştur
	validate = validator.New()

	// MongoDB koleksiyonunu seçin
	userCollection = client.Database(dbName).Collection(userCollName)
	articleCollection = client.Database(dbName).Collection(articleCollName)

	// AWS Lambda işlevini başlatın
	lambda.Start(handlePostRequestRouting)
}
