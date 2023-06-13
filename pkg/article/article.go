package article

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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

type ArticleData struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description" validate:"required"`
	Author      string `json:"author" validate:""`
	Date        string `json:"date" validate:""`
}

func AddArticle(ctx context.Context, request events.APIGatewayProxyRequest, mongoColl mongo.Collection, username string) (events.APIGatewayProxyResponse, error) {
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

	data := bson.M{
		"title":       articleData.Title,
		"description": articleData.Description,
		"author":      username,
		"date":        currentTime.Format("02-01-2006"),
	}

	// Veriyi MongoDB'ye ekle
	_, err := mongoColl.InsertOne(ctx, data)
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

func GetArticles(ctx context.Context, request events.APIGatewayProxyRequest, mongoColl mongo.Collection) (events.APIGatewayProxyResponse, error) {
	cursor, err := mongoColl.Find(ctx, bson.M{})
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

func GetArticlesByTitle(ctx context.Context, request events.APIGatewayProxyRequest, mongoColl mongo.Collection) (events.APIGatewayProxyResponse, error) {
	// MongoDB'den veri alın
	title := request.QueryStringParameters["title"]
	filter := bson.M{"title": title}
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

func DeleteArticleByTitle(ctx context.Context, request events.APIGatewayProxyRequest, mongoColl mongo.Collection) (events.APIGatewayProxyResponse, error) {
	// Silinecek Article'ın "title" değerini al
	title := request.QueryStringParameters["title"]

	// MongoDB'den Article'ı sil
	filter := bson.M{"title": title}
	result, err := mongoColl.DeleteOne(ctx, filter)
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
