package user

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

var(
	ErrorFailedToUnmarshalRecord = "failed to unmarshal record"
	ErrorFailedToFetchRecord = "failed to fetch record"
	ErrorInvalidUserData = "invalid user data"
	ErrorInvalidEmail = "invalid email"
	ErrorCouldNotMarshalItem = "could not marshal item"
	ErrorCouldNotDeleteItem = "could not delete item"
	ErrorCouldNotDynamoPutItem = "could not dynamo put item"
	ErrorUserAlreadyExists = "user.User already exists"
	ErrorUserDoesNotExist = "user.User does not exist"
)

type User struct{
	Email 		string	`json:"email"`
	FirstName	string 	`json:"firstName"`
	LastName	string 	`json:"lastName"`
}

// POST isteğini işlemek için kullanılan işlev
func handleCreateUserRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// İstek gövdesinden veriyi çözümle
	var userData UserData
	if err := json.Unmarshal([]byte(request.Body), &userData); err != nil {
		log.Println("JSON Unmarshal error:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Bad Request",
		}, nil
	}

	// Veri doğrulama işlemini gerçekleştir
	if err := validate.Struct(userData); err != nil {
		log.Println("Validation error:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Validation Error",
		}, nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userData.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Password hashing error:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Internal Server Error",
		}, nil
	}

	data := bson.M{
		"username": userData.Username,
		"email":    userData.Email,
		"password": string(hashedPassword),
	}

	// Veriyi MongoDB'ye ekle
	_, err = userCollection.InsertOne(ctx, data)
	if err != nil {
		log.Println("InsertOne error:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Internal Server Error",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusCreated,
		Body:       "Data created",
	}, nil
}