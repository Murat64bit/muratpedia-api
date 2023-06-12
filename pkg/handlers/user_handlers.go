package handlers

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/murat64bit/muratpedia-api/pkg/user"
	"go.mongodb.org/mongo-driver/mongo"
)

var ErrorMethodNotAllowed = "method not allowed"

type ErrorBody struct {
	ErrorMsg *string `json:"error,omitempty"`
}

func RegisterUser(ctx context.Context, req events.APIGatewayProxyRequest, mongoColl mongo.Collection) (
	*events.APIGatewayProxyResponse, error,
) {
	result, err := user.RegisterUser(ctx, req, mongoColl)
	if err != nil {
		return apiResponse(http.StatusBadRequest, ErrorBody{
			aws.String(err.Error()),
		})
	}
	return apiResponse(http.StatusCreated, result)
}
