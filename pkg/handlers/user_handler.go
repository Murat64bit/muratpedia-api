package handlers

import (
	"net/http"

	"github.com/Murat64bit/muratpedia-api/pkg/user"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
)

var ErrorMethodNotAllowed = "method not allowed"

type ErrorBody struct{
	ErrorMsg *string `json:"error,omitempty"`
}

func CreateUser(req events.APIGatewayProxyRequest)(
	*events.APIGatewayProxyResponse, error,
){
	result, err := user.CreateUser(req)
	if err!=nil {
		return apiResponse(http.StatusBadRequest, ErrorBody{
			aws.String(err.Error()),
		})
	}
	return apiResponse(http.StatusCreated, result)
}
