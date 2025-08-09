output "lambda_producer_role_arn" {
  value = aws_iam_role.lambda_producer_role.arn
}

output "lambda_consumer_role_arn" {
  value = aws_iam_role.lambda_consumer_role.arn
}
