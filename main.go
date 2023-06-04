package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-playground/validator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	jwtKey = []byte("my_secret_key")
)

// E-posta adresinin doğrulama için kullanılacak regex deseni
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

type UserData struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

type ArticleData struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description" validate:"required"`
	Author      string `json:"author" validate:""`
	Date        string `json:"date" validate:""`
}

// POST isteğini yönlendirme işlevi
func handlePostRequestRouting(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch request.Path {
	case "/login":
		return handleLoginRequest(ctx, request)
	case "/register":
		return handleRegisterRequest(ctx, request)
	case "/addArticle":
		return handleAddArticleRequest(ctx, request)
	case "/getUsers":
		return handleGetUsersRequest(ctx, request)
	case "/getArticles":
		return handleGetArticlesRequest(ctx, request)
	case "/getUserById":
		return handleGetUserByID(ctx, request)
	case "/getArticlesByTitle":
		return handleGetArticlesByTitle(ctx, request)
	case "/deleteUserById":
		return handleDeleteUserByID(ctx, request)
	case "/deleteArticleByTitle":
		return handleDeleteArticleByTitle(ctx, request)
	default:
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNotFound,
			Body:       "Not Found",
		}, nil
	}
}

func handleGetArticlesRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	cursor, err := articleCollection.Find(ctx, bson.M{})
	if err != nil {
		log.Println("Find error:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Internal Server Error",
		}, nil
	}
	defer cursor.Close(ctx)

	var articles []ArticleData
	if err := cursor.All(ctx, &articles); err != nil {
		log.Println("Cursor All error:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Internal Server Error",
		}, nil
	}

	responseBody, err := json.Marshal(articles)
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

// "title" değerine göre Article'ı silen DELETE işlevi
func handleDeleteArticleByTitle(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	valid := validateJWT(request)
	if !valid {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusUnauthorized,
			Body:       "Unauthorized",
		}, nil
	}

	// Silinecek Article'ın "title" değerini al
	title := request.QueryStringParameters["title"]

	// MongoDB'den Article'ı sil
	filter := bson.M{"title": title}
	result, err := articleCollection.DeleteOne(ctx, filter)
	if err != nil {
		log.Println("DeleteOne error:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Internal Server Error",
		}, nil
	}

	// Silinen Article sayısını kontrol et
	if result.DeletedCount == 0 {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNotFound,
			Body:       "Article not found",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       "Article deleted successfully",
	}, nil
}


// MongoDB _id değerine göre kullanıcıyı silen DELETE işlevi
func handleDeleteUserByID(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	valid := validateJWT(request)
	if !valid {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusUnauthorized,
			Body:       "Unauthorized",
		}, nil
	}
	// Silinecek kullanıcının _id'sini al
	id := request.QueryStringParameters["_id"]

	// _id'yi MongoDB ObjectID formatına dönüştür
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Println("Invalid ObjectID:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Invalid ObjectID",
		}, nil
	}

	// MongoDB'den veri silin
	filter := bson.M{"_id": objectID}
	result, err := userCollection.DeleteOne(ctx, filter)
	if err != nil {
		log.Println("DeleteOne error:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Internal Server Error",
		}, nil
	}

	// Silinen kullanıcı sayısını kontrol et
	if result.DeletedCount == 0 {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNotFound,
			Body:       "User not found",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       "User deleted successfully",
	}, nil
}


// MongoDB _id değerine göre kullanıcı getiren GET işlevi
func handleGetUserByID(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Getirilecek kullanıcının _id'sini al
	id := request.QueryStringParameters["_id"]

	// _id'yi MongoDB ObjectID formatına dönüştür
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Println("Invalid ObjectID:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Invalid ObjectID",
		}, nil
	}

	// MongoDB'den kullanıcıyı getir
	filter := bson.M{"_id": objectID}
	var user UserData
	err = userCollection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		log.Println("FindOne error:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNotFound,
			Body:       "User not found",
		}, nil
	}

	// Kullanıcıyı JSON formatına dönüştür
	responseBody, err := json.Marshal(user)
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


// Title'a göre Article'ları getiren GET işlevi
func handleGetArticlesByTitle(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// MongoDB'den veri alın
	title := request.QueryStringParameters["title"]
	filter := bson.M{"title": title}
	cursor, err := articleCollection.Find(ctx, filter)
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


// JWT tabanlı oturum açma işlemini gerçekleştiren işlev
func handleLoginRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var loginData struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	err := json.Unmarshal([]byte(request.Body), &loginData)
	if err != nil {
		log.Println("JSON Unmarshal error:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Bad Request",
		}, nil
	}

	// Kullanıcıyı doğrula
	filter := bson.M{"email": loginData.Email}
	result := userCollection.FindOne(ctx, filter)
	if result.Err() != nil {
		log.Println("FindOne error:", result.Err())
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusUnauthorized,
			Body:       "Unauthorized",
		}, nil
	}

	var user UserData
	err = result.Decode(&user)
	if err != nil {
		log.Println("Decode error:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Internal Server Error",
		}, nil
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginData.Password))
	if err != nil {
		log.Println("Password comparison error:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusUnauthorized,
			Body:       "Unauthorized",
		}, nil
	}

	// JWT talep üzerine oluşturulur
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
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Internal Server Error",
		}, nil
	}

	// JWT'yi yanıtla birlikte döndür
	responseBody := map[string]string{"token": tokenString}
	responseJSON, err := json.Marshal(responseBody)
	if err != nil {
		log.Println("JSON Marshal error:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Internal Server Error",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(responseJSON),
	}, nil
}


// Kullanıcı oluşturma isteğini işleyen işlev
func handleRegisterRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var newUser UserData
	err := json.Unmarshal([]byte(request.Body), &newUser)
	if err != nil {
		log.Println("JSON Unmarshal error:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Bad Request",
		}, nil
	}

	// E-posta adresinin doğruluğunu kontrol et
	if !validateEmail(newUser.Email) {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Invalid email address",
		}, nil
	}

	// E-posta adresinin benzersizliğini kontrol et
	existingUser, err := getUserByEmail(newUser.Email)
	if err != nil {
		log.Println("Database query error:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Internal Server Error",
		}, nil
	}
	if existingUser != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Email address is already taken",
		}, nil
	}

	// Kullanıcıyı veritabanına kaydet
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Password hashing error:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Internal Server Error",
		}, nil
	}
	newUser.Password = string(hashedPassword)

	_, err = userCollection.InsertOne(ctx, newUser)
	if err != nil {
		log.Println("InsertOne error:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Internal Server Error",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       "User created successfully",
	}, nil
}

// E-posta adresinin doğruluğunu kontrol eden işlev
func validateEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// Veritabanında belirli bir e-posta adresine sahip kullanıcıyı bulan işlev
func getUserByEmail(email string) (*UserData, error) {
	filter := bson.M{"email": email}
	result := userCollection.FindOne(context.Background(), filter)
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			return nil, nil // Kullanıcı bulunamadı
		}
		return nil, result.Err()
	}

	var user UserData
	err := result.Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}


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

func validateJWT(request events.APIGatewayProxyRequest) (bool) {
	tokenString := request.Headers["Authorization"]
	if tokenString == "" {
		return  false
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		return  false
	}

	return true
}



// GET isteğini işlemek için kullanılan işlev
func handleGetUsersRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// POST işleminden önce JWT doğrulaması yapılmalı
	valid := validateJWT(request)
	if !valid {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusUnauthorized,
			Body:       "Unauthorized",
		}, nil
	}

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
	
	// POST işleminden önce JWT doğrulaması yapılmalı
	username, valid := validateJWTandGetUserName(request)
	if !valid {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusUnauthorized,
			Body:       "Unauthorized",
		}, nil
	}

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
		"author":      username,
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
