module "eks" {
    source = "terraform-aws-modules/eks/aws"
    version = "~> 21.24"

    name = var.cluster_name
    kubernetes_version = "1.36"

    endpoint_public_access = true

    # Required: Without both of these, the terraform user won't have acess to the cluster.
    authentication_mode = "API_AND_CONFIG_MAP"
    enable_cluster_creator_admin_permissions = true

    vpc_id = module.vpc.vpc_id
    subnet_ids = module.vpc.private_subnets

    eks_managed_node_groups = {
        general = {
            name = "refresher-nodes"
            # Save costs with this, good for dev.
            capacity_type  = "SPOT"
            instance_types = ["t3.medium", "t3a.medium", "t2.medium", "m5.large", "m4.large"]
            min_size = 0
            max_size = 5
            desired_size = 3

            subnet_ids = module.vpc.private_subnets
            iam_role_additional_policies = {
                ssm = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
            }

            timeouts = {
                create = "10m" 
                update = "10m" 
                delete = "10m" 
            }
        }
    }

    addons = {
        coredns = { 
            most_recent = true 
            # do not use before_compute, it requires the nodes to be ready.
        }
        kube-proxy = { 
            most_recent = true 
            before_compute = true
        }
        vpc-cni = { 
            most_recent = true 
            before_compute = true
        }
        eks-pod-identity-agent = { 
            most_recent = true 
            resolve_conflicts_on_create = "OVERWRITE"
        }
    }
}