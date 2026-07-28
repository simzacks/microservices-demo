output "cluster_endpoint" {
    description = "Endpoint for EKS control plane"
    value = module.eks.cluster_endpoint
}

output "cluster_name" {
    description = "Name of the EKS cluster"
    value = module.eks.cluster_name
}

output "vpc_id" {
    description = "ID of the VPC used by the EKS cluster"
    value = module.vpc.vpc_id
}

output "private_subnets" {
    description = "List of private subnets for the EKS cluster"
    value = module.vpc.private_subnets
}

output "valket_primary_endpoint" {
    description = "Elasticache Valkey endpoint to pass to cart service env vars"
    value = aws_elasticache_replication_group.valkey.primary_endpoint_address
}