variable "aws_region" {
    type = string
    default = "us-east-1"
    description = "AWS region used"
}

variable "environment" {
    type = string
    default = "dev"
    description = "Deployment environment name"
}

variable "cluster_name" {
    type = string
    default = "devops-refresher"
    description = "Name of the Kubernetes cluster"
}

variable "vpc_cidr" {
    type = string
    default = "10.0.0.0/16"
    description = "VPC CIDR block"
}

