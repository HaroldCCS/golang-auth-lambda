package handler

import (
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

func JSONResponse(statusCode int, body interface{}) (events.APIGatewayProxyResponse, error) {
	respBody, err := json.Marshal(body)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       `{"error": "Internal server error"}`,
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(respBody),
	}, nil
}

func ErrorResponse(statusCode int, message string) (events.APIGatewayProxyResponse, error) {
	return JSONResponse(statusCode, map[string]string{"error": message})
}
