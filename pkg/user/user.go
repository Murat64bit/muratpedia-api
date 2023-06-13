package user

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"regexp"

	"github.com/aws/aws-lambda-go/events"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrorFailedToUnmarshalRecord = "failed to unmarshal record"
	ErrorFailedToFetchRecord     = "failed to fetch record"
	ErrorInvalidUserData         = "invalid user data"
	ErrorInvalidEmail            = "invalid email"
	ErrorCouldNotMarshalItem     = "could not marshal item"
	ErrorCouldNotDeleteItem      = "could not delete item"
	ErrorCouldNotDynamoPutItem   = "could not dynamo put item"
	ErrorUserAlreadyExists       = "user.User already exists"
	ErrorUserDoesNotExist        = "user.User does not exist"
)

type UserData struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func getUserByEmail(email string, mongoColl mongo.Collection) (*UserData, error) {
	filter := bson.M{"email": email}
	result := mongoColl.FindOne(context.Background(), filter)
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

func validateEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func RegisterUser(ctx context.Context, request events.APIGatewayProxyRequest, mongoColl mongo.Collection) (events.APIGatewayProxyResponse, error) {
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
	existingUser, err := getUserByEmail(newUser.Email, mongoColl)
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

	_, err = mongoColl.InsertOne(ctx, newUser)
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

// JWT tabanlı oturum açma işlemini gerçekleştiren işlev
func LoginUser(ctx context.Context, request events.APIGatewayProxyRequest, mongoColl mongo.Collection) (events.APIGatewayProxyResponse, error) {
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
	result := mongoColl.FindOne(ctx, filter)
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

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       "Succesfuly you are logged in",
	}, nil
}

func GetUsers(ctx context.Context, request events.APIGatewayProxyRequest, mongoColl mongo.Collection) (events.APIGatewayProxyResponse, error) {

	// MongoDB'den veri alın
	filter := bson.M{}
	cursor, err := mongoColl.Find(ctx, filter)
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

func GetUserById(ctx context.Context, request events.APIGatewayProxyRequest, mongoColl mongo.Collection) (events.APIGatewayProxyResponse, error) {

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
	err = mongoColl.FindOne(ctx, filter).Decode(&user)
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

func DeleteUserById(ctx context.Context, request events.APIGatewayProxyRequest, mongoColl mongo.Collection) (events.APIGatewayProxyResponse, error) {

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
	result, err := mongoColl.DeleteOne(ctx, filter)
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
