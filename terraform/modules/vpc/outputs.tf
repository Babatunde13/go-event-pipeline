output "vpc_id" {
  description = "VPC ID"
  value       = aws_vpc.main.id
}

output "public_subnet_ids" {
  description = "List of public subnet IDs"
  value       = aws_subnet.public[*].id
}

output "private_subnet_ids" {
  description = "List of private subnet IDs"
  value       = aws_subnet.private[*].id
}

output "lambda_sg_id" {
  description = "Security Group for Lambda functions"
  value       = aws_security_group.lambda_sg.id
}

output "kafka_sg_id" {
  description = "Security Group for Kafka brokers"
  value       = aws_security_group.kafka_sg.id
}

output "redis_sg_id" {
  description = "Security Group for Redis"
  value       = aws_security_group.redis_sg.id
}
