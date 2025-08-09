module "iam" {
  source = "./modules/iam"
}

module "vpc" {
  source = "./modules/vpc"
}

module "msk" {
  source           = "./modules/msk"
  subnet_ids       = module.vpc.private_subnet_ids
  security_group_id = module.vpc.kafka_sg_id
  cluster_name     = "event-kafka-cluster"
}

module "redis" {
  source           = "./modules/redis"
  subnet_ids       = module.vpc.private_subnet_ids
  security_group_id = module.vpc.redis_sg_id
}

module "eventbridge" {
  source = "./modules/eventbridge"
  lambda_role_arn = module.iam.lambda_producer_role_arn
}

module "lambda-producer" {
  source = "./modules/lambda-producer"
  lambda_role_arn = module.iam.lambda_producer_role_arn
}
