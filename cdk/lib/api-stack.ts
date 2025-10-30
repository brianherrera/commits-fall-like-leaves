import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
import * as apigateway from 'aws-cdk-lib/aws-apigateway';
import * as lambda from 'aws-cdk-lib/aws-lambda';

export class ApiStack extends cdk.Stack {
  public readonly api: apigateway.RestApi;
  public readonly lambdaFunction: lambda.Function;

  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    // Create Lambda function (placeholder for now)
    this.lambdaFunction = new lambda.Function(this, 'HaikuLambdaFunction', {
      runtime: lambda.Runtime.NODEJS_18_X,
      handler: 'index.handler',
      code: lambda.Code.fromInline(`
        exports.handler = async (event) => {
          console.log('Event:', JSON.stringify(event, null, 2));
          
          const body = JSON.parse(event.body || '{}');
          const commitMessage = body.commit_message;
          
          return {
            statusCode: 200,
            headers: {
              'Content-Type': 'application/json',
              'Access-Control-Allow-Origin': '*',
              'Access-Control-Allow-Methods': 'POST, OPTIONS',
              'Access-Control-Allow-Headers': 'Content-Type, Authorization'
            },
            body: JSON.stringify({
              message: 'Haiku generated successfully!',
              commit_message: commitMessage,
              haiku: [
                'Code flows like water',
                'Commits fall like autumn leaves',
                'Beauty in the merge'
              ],
              timestamp: new Date().toISOString(),
              requestId: event.requestContext?.requestId
            })
          };
        };
      `),
      timeout: cdk.Duration.seconds(30),
      memorySize: 256,
      description: 'Lambda function to generate haiku from commit messages'
    });

    // Create REST API Gateway
    this.api = new apigateway.RestApi(this, 'HaikuApi', {
      restApiName: 'Haiku Generator API',
      description: 'API to generate haiku from commit messages',
      deployOptions: {
        stageName: 'prod',
        throttlingRateLimit: 100,
        throttlingBurstLimit: 200,
        loggingLevel: apigateway.MethodLoggingLevel.INFO,
        dataTraceEnabled: true,
        metricsEnabled: true
      },
      defaultCorsPreflightOptions: {
        allowOrigins: apigateway.Cors.ALL_ORIGINS,
        allowMethods: ['POST', 'OPTIONS'],
        allowHeaders: ['Content-Type', 'Authorization']
      },
      endpointConfiguration: {
        types: [apigateway.EndpointType.REGIONAL]
      }
    });

    // Create request validator for method request validation
    const requestValidator = new apigateway.RequestValidator(this, 'HaikuRequestValidator', {
      restApi: this.api,
      requestValidatorName: 'haiku-request-validator',
      validateRequestBody: true,
      validateRequestParameters: false
    });

    // Create request model for POST /haiku validation
    const haikuRequestModel = new apigateway.Model(this, 'HaikuRequestModel', {
      restApi: this.api,
      modelName: 'HaikuRequest',
      contentType: 'application/json',
      schema: {
        type: apigateway.JsonSchemaType.OBJECT,
        properties: {
          commit_message: {
            type: apigateway.JsonSchemaType.STRING,
            minLength: 1,
            maxLength: 500,
            description: 'The commit message to generate a haiku from'
          }
        },
        required: ['commit_message'],
        additionalProperties: false
      }
    });

    // Create success response model
    const haikuResponseModel = new apigateway.Model(this, 'HaikuResponseModel', {
      restApi: this.api,
      modelName: 'HaikuResponse',
      contentType: 'application/json',
      schema: {
        type: apigateway.JsonSchemaType.OBJECT,
        properties: {
          message: {
            type: apigateway.JsonSchemaType.STRING
          },
          commit_message: {
            type: apigateway.JsonSchemaType.STRING
          },
          haiku: {
            type: apigateway.JsonSchemaType.ARRAY,
            items: {
              type: apigateway.JsonSchemaType.STRING
            }
          },
          timestamp: {
            type: apigateway.JsonSchemaType.STRING
          },
          requestId: {
            type: apigateway.JsonSchemaType.STRING
          }
        },
        required: ['message', 'commit_message', 'haiku', 'timestamp']
      }
    });

    // Create error response model
    const errorResponseModel = new apigateway.Model(this, 'ErrorResponseModel', {
      restApi: this.api,
      modelName: 'ErrorResponse',
      contentType: 'application/json',
      schema: {
        type: apigateway.JsonSchemaType.OBJECT,
        properties: {
          error: {
            type: apigateway.JsonSchemaType.STRING
          },
          message: {
            type: apigateway.JsonSchemaType.STRING
          },
          timestamp: {
            type: apigateway.JsonSchemaType.STRING
          }
        },
        required: ['error', 'message']
      }
    });

    // Create Lambda integration
    const lambdaIntegration = new apigateway.LambdaIntegration(this.lambdaFunction, {
      requestTemplates: {
        'application/json': JSON.stringify({
          httpMethod: '$context.httpMethod',
          resourcePath: '$context.resourcePath',
          headers: '$input.params().header',
          body: '$input.json(\'$\')',
          requestContext: {
            requestId: '$context.requestId',
            stage: '$context.stage',
            accountId: '$context.accountId',
            apiId: '$context.apiId',
            identity: {
              sourceIp: '$context.identity.sourceIp',
              userAgent: '$context.identity.userAgent'
            }
          }
        })
      },
      integrationResponses: [
        {
          statusCode: '200',
          responseTemplates: {
            'application/json': '$input.json(\'$\')'
          },
          responseParameters: {
            'method.response.header.Access-Control-Allow-Origin': "'*'"
          }
        },
        {
          statusCode: '400',
          selectionPattern: '.*"statusCode":400.*',
          responseTemplates: {
            'application/json': JSON.stringify({
              error: 'Bad Request',
              message: '$input.path(\'$.errorMessage\')',
              timestamp: '$context.requestTime'
            })
          },
          responseParameters: {
            'method.response.header.Access-Control-Allow-Origin': "'*'"
          }
        },
        {
          statusCode: '500',
          selectionPattern: '.*"statusCode":500.*',
          responseTemplates: {
            'application/json': JSON.stringify({
              error: 'Internal Server Error',
              message: 'An error occurred processing your request',
              timestamp: '$context.requestTime'
            })
          },
          responseParameters: {
            'method.response.header.Access-Control-Allow-Origin': "'*'"
          }
        }
      ]
    });

    // Create /haiku resource
    const haikuResource = this.api.root.addResource('haiku');

    // POST /haiku - Generate haiku from commit message with request body validation
    haikuResource.addMethod('POST', lambdaIntegration, {
      requestValidator: requestValidator,
      requestModels: {
        'application/json': haikuRequestModel
      },
      methodResponses: [
        {
          statusCode: '200',
          responseModels: {
            'application/json': haikuResponseModel
          },
          responseParameters: {
            'method.response.header.Access-Control-Allow-Origin': true
          }
        },
        {
          statusCode: '400',
          responseModels: {
            'application/json': errorResponseModel
          },
          responseParameters: {
            'method.response.header.Access-Control-Allow-Origin': true
          }
        },
        {
          statusCode: '500',
          responseModels: {
            'application/json': errorResponseModel
          },
          responseParameters: {
            'method.response.header.Access-Control-Allow-Origin': true
          }
        }
      ]
    });

    // Output the API URL
    new cdk.CfnOutput(this, 'HaikuApiUrl', {
      value: this.api.url,
      description: 'URL of the Haiku API',
      exportName: 'HaikuApiUrl'
    });

    // Output the specific endpoint URL
    new cdk.CfnOutput(this, 'HaikuEndpoint', {
      value: `${this.api.url}haiku`,
      description: 'POST endpoint for haiku generation',
      exportName: 'HaikuEndpoint'
    });

    // Output the API ID
    new cdk.CfnOutput(this, 'HaikuApiId', {
      value: this.api.restApiId,
      description: 'ID of the Haiku API',
      exportName: 'HaikuApiId'
    });

    // Output the Lambda function ARN
    new cdk.CfnOutput(this, 'HaikuLambdaArn', {
      value: this.lambdaFunction.functionArn,
      description: 'ARN of the Haiku Lambda function',
      exportName: 'HaikuLambdaArn'
    });
  }
}