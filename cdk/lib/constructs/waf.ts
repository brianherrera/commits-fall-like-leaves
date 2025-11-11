import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
import * as wafv2 from 'aws-cdk-lib/aws-wafv2';

export interface WafConstructProps {
  /**
   * The name of the WAF Web ACL
   */
  name: string;
  
  /**
   * Rate limit per 5-minute window
   */
  rateLimit: number;
  
  /**
   * Description for the WAF Web ACL
   */
  description?: string;
}

export class WafConstruct extends Construct {
  public readonly webAcl: wafv2.CfnWebACL;

  constructor(scope: Construct, id: string, props: WafConstructProps) {
    super(scope, id);

    // Create WAF Web ACL with rate limiting rule
    this.webAcl = new wafv2.CfnWebACL(this, 'WebACL', {
      name: props.name,
      description: props.description || `WAF Web ACL for ${props.name} with rate limiting`,
      scope: 'REGIONAL', // For API Gateway, ALB, AppSync
      defaultAction: {
        allow: {}
      },
      rules: [
        {
          name: 'RateLimitRule',
          priority: 1,
          statement: {
            rateBasedStatement: {
              limit: props.rateLimit,
              aggregateKeyType: 'IP'
            }
          },
          action: {
            block: {}
          },
          visibilityConfig: {
            sampledRequestsEnabled: true,
            cloudWatchMetricsEnabled: true,
            metricName: 'RateLimitRule'
          }
        }
      ],
      visibilityConfig: {
        sampledRequestsEnabled: true,
        cloudWatchMetricsEnabled: true,
        metricName: props.name.replace(/[^a-zA-Z0-9]/g, '')
      }
    });

    // Output the Web ACL ARN for reference
    new cdk.CfnOutput(this, 'WebAclArn', {
      value: this.webAcl.attrArn,
      description: `ARN of the ${props.name} WAF Web ACL`,
      exportName: `${props.name}WebAclArn`
    });

    // Output the Web ACL ID for reference
    new cdk.CfnOutput(this, 'WebAclId', {
      value: this.webAcl.attrId,
      description: `ID of the ${props.name} WAF Web ACL`,
      exportName: `${props.name}WebAclId`
    });
  }
}