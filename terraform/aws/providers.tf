terraform {
    required_version = ">= 1.15.0"

    required_providers {
        aws = {
            source = "hashicorp/aws"
            version = "~> 6.56"
        }
        helm = {
            source = "hashicorp/helm"
            version = "~> 3.2"
        }
        kubernetes = {
            source = "hashicorp/kubernetes"
            version = "~> 3.2"
        }
    }
}

provider "aws" {
    region = var.aws_region

    default_tags {
        tags = {
            Project = "devops_refresher"
            ManagedBy = "Terraform"
            Environment = var.environment
        }
    }
}

data "aws_eks_cluster_auth" "cluster" {
    name = module.eks.cluster_name
}

provider "kubernetes" {
    host = module.eks.cluster_endpoint
    cluster_ca_certificate = base64decode(module.eks.cluster_certificate_authority_data)
    token = data.aws_eks_cluster_auth.cluster.token  
}

provider "helm" {
    kubernetes = {
        host = module.eks.cluster_endpoint
        cluster_ca_certificate = base64decode(module.eks.cluster_certificate_authority_data)
        token = data.aws_eks_cluster_auth.cluster.token  
    }
}