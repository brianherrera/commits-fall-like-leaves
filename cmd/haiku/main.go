package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/brianherrera/commits-fall-like-leaves/internal/api"
	"github.com/gin-gonic/gin"
)

var ginLambda *ginadapter.GinLambda

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("failed to load aws config")
	}

	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	haikuAPI := api.NewDefaultHaikuAPI(cfg)
	haikuAPI.SetupMiddleware(router)
	haikuAPI.SetupRoutes(router)

	// Lambda adapter
	ginLambda = ginadapter.New(router)
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(Handler)
}
