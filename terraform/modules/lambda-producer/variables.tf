variable "lambda_role_arn" {
  description = "IAM role ARN for the Lambda function"
  type        = string
}

variable "function_name" {
  description = "Name of the Lambda function"
  type        = string
  default     = "kafka-producer-lambda"
}

variable "handler" {
  description = "Lambda function handler"
  type        = string
  default     = "index.handler"
}

variable "runtime" {
  description = "Lambda runtime environment"
  type        = string
  default     = "nodejs18.x"
}

variable "source_path" {
  description = "Path to Lambda source code folder (zip or directory)"
  type        = string
  default     = "../../lambda-producer-code"
}

variable "subnet_ids" {
  description = "List of subnet IDs to place Lambda in VPC"
  type        = list(string)
  default     = []
}

variable "security_group_ids" {
  description = "Security group IDs for the Lambda function"
  type        = list(string)
  default     = []
}

variable "environment_variables" {
  description = "Environment variables map for Lambda"
  type        = map(string)
  default     = {}
}
