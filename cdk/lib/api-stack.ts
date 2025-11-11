import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
import * as apigateway from 'aws-cdk-lib/aws-apigateway';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as wafv2 from 'aws-cdk-lib/aws-wafv2';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as logs from 'aws-cdk-lib/aws-logs';
import { WafConstruct } from './constructs/waf';

export class ApiStack extends cdk.Stack {
  public readonly api: apigateway.RestApi;
  public readonly lambdaFunction: lambda.Function;
  public readonly waf: WafConstruct;

  constructor(scope: Construct, id: string, props: cdk.StackProps) {
    super(scope, id, props);

    // Create Lambda function
    this.lambdaFunction = new lambda.Function(this, 'HaikuLambdaFunction', {
      runtime: lambda.Runtime.PROVIDED_AL2023,
      handler: 'bootstrap',
      code: lambda.Code.fromAsset('../lambda-function.zip'),
      timeout: cdk.Duration.seconds(30),
      memorySize: 256,
      architecture: lambda.Architecture.X86_64,
      description: 'Lambda function to generate haiku from commit messages'
    });

    const bedrockModelID = "anthropic.claude-haiku-4-5-20251001-v1:0"

    this.lambdaFunction.addToRolePolicy(new iam.PolicyStatement({
      effect: iam.Effect.ALLOW,
      actions: ['bedrock:InvokeModel'],
      resources: [
        `arn:aws:bedrock:${props.env?.region}:${props.env?.account}:inference-profile/global.${bedrockModelID}`,
        `arn:aws:bedrock:*::foundation-model/${bedrockModelID}`,
      ]
    }));

    const apiGatewayCloudWatchRole = new iam.Role(this, 'ApiGatewayCloudWatchRole', {
      assumedBy: new iam.ServicePrincipal('apigateway.amazonaws.com'),
      managedPolicies: [
        iam.ManagedPolicy.fromAwsManagedPolicyName('service-role/AmazonAPIGatewayPushToCloudWatchLogs')
      ]
    });

    this.api = new apigateway.RestApi(this, 'HaikuApi', {
      restApiName: 'Haiku Generator API',
      description: 'API to generate haiku from commit messages',
      deployOptions: {
        stageName: 'prod',
        throttlingRateLimit: 10,
        throttlingBurstLimit: 50,
        loggingLevel: apigateway.MethodLoggingLevel.INFO,
        dataTraceEnabled: true,
        metricsEnabled: true,
        accessLogDestination: new apigateway.LogGroupLogDestination(new logs.LogGroup(this, 'ApiAccessLogs')),
        accessLogFormat: apigateway.AccessLogFormat.jsonWithStandardFields()
      },
      defaultCorsPreflightOptions: {
        allowOrigins: apigateway.Cors.ALL_ORIGINS,
        allowMethods: ['POST', 'OPTIONS'],
        allowHeaders: ['Content-Type', 'Authorization']
      },
      endpointConfiguration: {
        types: [apigateway.EndpointType.REGIONAL]
      },
      cloudWatchRole: true
    });
    
    const apiGatewayAccount = new apigateway.CfnAccount(this, 'ApiGatewayAccount', {
      cloudWatchRoleArn: apiGatewayCloudWatchRole.roleArn
    });
    
    // Add dependency to ensure the role is created before API Gateway tries to use it
    this.api.node.addDependency(apiGatewayAccount);

    const requestValidator = new apigateway.RequestValidator(this, 'HaikuRequestValidator', {
      restApi: this.api,
      requestValidatorName: 'haiku-request-validator',
      validateRequestBody: true,
      validateRequestParameters: false
    });

    const haikuRequestModel = new apigateway.Model(this, 'HaikuRequestModel', {
      restApi: this.api,
      modelName: 'HaikuRequest',
      contentType: 'application/json',
      schema: {
        type: apigateway.JsonSchemaType.OBJECT,
        properties: {
          commitMessage: {
            type: apigateway.JsonSchemaType.STRING,
            minLength: 1,
            maxLength: 500,
            description: 'The commit message to generate a haiku from'
          },
          mood: {
            type: apigateway.JsonSchemaType.STRING,
            enum: ['humorous', 'reflective', 'technical'],
            description: 'Optional mood for the haiku'
          }
        },
        required: ['commitMessage'],
        additionalProperties: false
      }
    });

    const haikuResponseModel = new apigateway.Model(this, 'HaikuResponseModel', {
      restApi: this.api,
      modelName: 'HaikuResponse',
      contentType: 'application/json',
      schema: {
        type: apigateway.JsonSchemaType.OBJECT,
        properties: {
          haiku: {
            type: apigateway.JsonSchemaType.STRING
          }
        },
        required: ['haiku']
      }
    });

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

    this.waf = new WafConstruct(this, 'HaikuWaf', {
      name: 'HaikuApiWaf',
      rateLimit: 50, // 50 requests per 5-minute window per IP
      description: 'WAF for Haiku API with rate limiting protection'
    });

    const wafAssociation = new wafv2.CfnWebACLAssociation(this, 'HaikuApiWafAssociation', {
      resourceArn: `arn:aws:apigateway:${this.region}::/restapis/${this.api.restApiId}/stages/prod`,
      webAclArn: this.waf.webAcl.attrArn
    });
    
    // Add explicit dependency on the API Gateway deployment stage
    wafAssociation.node.addDependency(this.api.deploymentStage);

    new cdk.CfnOutput(this, 'HaikuApiUrl', {
      value: this.api.url,
      description: 'URL of the Haiku API',
      exportName: 'HaikuApiUrl'
    });

    new cdk.CfnOutput(this, 'HaikuEndpoint', {
      value: `${this.api.url}haiku`,
      description: 'POST endpoint for haiku generation',
      exportName: 'HaikuEndpoint'
    });

    new cdk.CfnOutput(this, 'HaikuApiId', {
      value: this.api.restApiId,
      description: 'ID of the Haiku API',
      exportName: 'HaikuApiId'
    });

    new cdk.CfnOutput(this, 'HaikuLambdaArn', {
      value: this.lambdaFunction.functionArn,
      description: 'ARN of the Haiku Lambda function',
      exportName: 'HaikuLambdaArn'
    });
  }
}