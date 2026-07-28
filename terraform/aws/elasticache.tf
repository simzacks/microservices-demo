# Valkey 7.2 is OSS compatible with Redis and should be compatible across all clouds. 
# The newer versions introduce proprietary functionality

resource "aws_security_group" "valkey" {
    name = "${var.cluster_name}-valkey-sg"
    description = "Allow inbound Redis traffic from EKS worker nodes"
    vpc_id = module.vpc.vpc_id

    ingress {
        description = "Valkey port from EKS nodes"
        from_port = 6379
        to_port = 6379
        protocol = "tcp"
        security_groups = [module.eks.node_security_group_id]
    }

    egress {
        from_port = 0
        to_port = 0
        protocol = "-1"
        cidr_blocks = ["0.0.0.0/0"]
    }

    tags = {
        Name = "${var.cluster_name}-valkey-sg"
    }
}

resource "aws_elasticache_subnet_group" "valkey" {
    name = "${var.cluster_name}-valkey-subnet-group"
    subnet_ids = module.vpc.private_subnets
}

resource "aws_elasticache_replication_group" "valkey" {
    replication_group_id = "${var.cluster_name}-cart-db"
    description = "Valkey cache for online boutique cart service"

    engine = "valkey"
    engine_version = "7.2" 
    node_type = "cache.t4g.micro"
    num_cache_clusters = 1
    port = 6379

    subnet_group_name = aws_elasticache_subnet_group.valkey.name
    security_group_ids = [aws_security_group.valkey.id]

    at_rest_encryption_enabled = true
    transit_encryption_enabled = false

    tags = {
        Name = "${var.cluster_name}-valkey"
    }
}