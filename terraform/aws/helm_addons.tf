# 1. Fetch official AWS Load Balancer Controller IAM Policy document
data "http" "alb_controller_iam_policy" {
  url = "https://raw.githubusercontent.com/kubernetes-sigs/aws-load-balancer-controller/main/docs/install/iam_policy.json"
}

resource "aws_iam_policy" "alb_controller" {
  name        = "${var.cluster_name}-alb-controller-policy"
  description = "IAM policy for AWS Load Balancer Controller"
  policy      = data.http.alb_controller_iam_policy.response_body
}

# 2. IAM Role with Pod Identity Trust Policy (pods.eks.amazonaws.com)
data "aws_iam_policy_document" "alb_pod_identity_trust" {
  statement {
    effect  = "Allow"
    actions = [
      "sts:AssumeRole",
      "sts:TagSession"
    ]

    principals {
      type        = "Service"
      identifiers = ["pods.eks.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "alb_controller" {
  name               = "${var.cluster_name}-alb-controller-role"
  assume_role_policy = data.aws_iam_policy_document.alb_pod_identity_trust.json
}

resource "aws_iam_role_policy_attachment" "alb_controller" {
  policy_arn = aws_iam_policy.alb_controller.arn
  role       = aws_iam_role.alb_controller.name
}

# 3. EKS Pod Identity Association linking Role -> Namespace + ServiceAccount
resource "aws_eks_pod_identity_association" "alb_controller" {
  cluster_name    = module.eks.cluster_name
  namespace       = "kube-system"
  service_account = "aws-load-balancer-controller"
  role_arn        = aws_iam_role.alb_controller.arn
}
resource "helm_release" "aws_load_balancer_controller" {
    name = "aws-load-balancer-controller"
    repository = "https://aws.github.io/eks-charts"
    chart = "aws-load-balancer-controller"
    version = "3.4.2"
    namespace = "kube-system"

    set = [
        {
            name = "clusterName"
            value = module.eks.cluster_name
        },
        {
            name = "serviceAccount.create"
            value = "true"
        },
        {
            name = "serviceAccount.name"
            value = "aws-load-balancer-controller"
        },
        {
            name = "serviceAccount.annotations.eks\\.amazonaws\\.com/role-arn"
        },
        {
            name = "region"
            value = var.aws_region
        },
        {
            name = "vpcId"
            value = module.vpc.vpc_id
        }
    ]
    depends_on = [
        module.eks,
        aws_eks_pod_identity_association.alb_controller
    ]
}

resource "helm_release" "argocd" {
    name = "argocd"
    repository = "https://argoproj.github.io/argo-helm"
    chart = "argo-cd"
    namespace = "argocd"
    create_namespace = true
    version = "10.2.1"

    set = [
      {
        name = "server.service.type"
        value = "ClusterIP"
      }
    ]
    depends_on = [module.eks]

}