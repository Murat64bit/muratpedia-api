package handlers

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/murat64bit/muratpedia-api/pkg/article"
	"github.com/murat64bit/muratpedia-api/pkg/user"
	"go.mongodb.org/mongo-driver/mongo"
)

var ErrorMethodNotAllowed = "method not allowed"

type ErrorBody struct {
	ErrorMsg *string `json:"error,omitempty"`
}

func RegisterUser(ctx context.Context, request events.APIGatewayProxyRequest, mongoColl mongo.Collection) (*events.APIGatewayProxyResponse, error) {
	result, err := user.RegisterUser(ctx, request, mongoColl)
	if err != nil {
		return apiResponse(http.StatusBadRequest, ErrorBody{
			aws.String(err.Error()),
		})
	}
	return apiResponse(http.StatusCreated, result)
}

func LoginUser(ctx context.Context, request events.APIGatewayProxyRequest, mongoColl mongo.Collection) (*events.APIGatewayProxyResponse, error) {
	result, err := user.LoginUser(ctx, request, mongoColl)
	if err != nil {
		return apiResponse(http.StatusBadRequest, ErrorBody{
			aws.String(err.Error()),
		})
	}
	return apiResponse(http.StatusCreated, result)
}

func GetUsers(ctx context.Context, request events.APIGatewayProxyRequest, mongoColl mongo.Collection) (*events.APIGatewayProxyResponse, error) {
	result, err := user.GetUsers(ctx, request, mongoColl)
	if err != nil {
		return apiResponse(http.StatusBadRequest, ErrorBody{
			aws.String(err.Error()),
		})
	}
	return apiResponse(http.StatusCreated, result)
}

func GetUserById(ctx context.Context, request events.APIGatewayProxyRequest, mongoColl mongo.Collection) (*events.APIGatewayProxyResponse, error) {
	result, err := user.GetUserById(ctx, request, mongoColl)
	if err != nil {
		return apiResponse(http.StatusBadRequest, ErrorBody{
			aws.String(err.Error()),
		})
	}
	return apiResponse(http.StatusCreated, result)
}

func DeleteUserById(ctx context.Context, request events.APIGatewayProxyRequest, mongoColl mongo.Collection) (*events.APIGatewayProxyResponse, error) {
	result, err := user.DeleteUserById(ctx, request, mongoColl)
	if err != nil {
		return apiResponse(http.StatusBadRequest, ErrorBody{
			aws.String(err.Error()),
		})
	}
	return apiResponse(http.StatusCreated, result)
}

func AddArticle(ctx context.Context, request events.APIGatewayProxyRequest, mongoColl mongo.Collection) (*events.APIGatewayProxyResponse, error) {
	result, err := article.AddArticle(ctx, request, mongoColl)
	if err != nil {
		return apiResponse(http.StatusBadRequest, ErrorBody{
			aws.String(err.Error()),
		})
	}
	return apiResponse(http.StatusCreated, result)
}

func GetArticles(ctx context.Context, request events.APIGatewayProxyRequest, mongoColl mongo.Collection) (*events.APIGatewayProxyResponse, error) {
	result, err := article.GetArticles(ctx, request, mongoColl)
	if err != nil {
		return apiResponse(http.StatusBadRequest, ErrorBody{
			aws.String(err.Error()),
		})
	}
	return apiResponse(http.StatusCreated, result)
}

func GetArticlesByTitle(ctx context.Context, request events.APIGatewayProxyRequest, mongoColl mongo.Collection) (*events.APIGatewayProxyResponse, error) {
	result, err := article.GetArticlesByTitle(ctx, request, mongoColl)
	if err != nil {
		return apiResponse(http.StatusBadRequest, ErrorBody{
			aws.String(err.Error()),
		})
	}
	return apiResponse(http.StatusCreated, result)
}

func DeleteArticleByTitle(ctx context.Context, request events.APIGatewayProxyRequest, mongoColl mongo.Collection) (*events.APIGatewayProxyResponse, error) {
	result, err := article.DeleteArticleByTitle(ctx, request, mongoColl)
	if err != nil {
		return apiResponse(http.StatusBadRequest, ErrorBody{
			aws.String(err.Error()),
		})
	}
	return apiResponse(http.StatusCreated, result)
}

func UnhandledMethod() (*events.APIGatewayProxyResponse, error) {
	return apiResponse(http.StatusMethodNotAllowed, ErrorMethodNotAllowed)
}

func UnauthorizedResponse() (*events.APIGatewayProxyResponse, error) {
	response := events.APIGatewayProxyResponse{
		StatusCode: http.StatusUnauthorized,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       `{"message": "Unauthorized"}`,
	}
	return &response, nil
}
