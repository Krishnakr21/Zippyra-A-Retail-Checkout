variable "environment" { type = string }
variable "vpc_id" { type = string }
variable "private_subnet_ids" { type = list(string) }
variable "instance_types" { type = list(string); default = ["t3.medium"] }
variable "desired_size" { type = number; default = 3 }
variable "min_size" { type = number; default = 2 }
variable "max_size" { type = number; default = 10 }
variable "spot_percentage" { type = number; default = 80 }

# EKS Cluster IAM Role
resource "aws_iam_role" "eks_cluster_role" {
  name = "zippyra-${var.environment}-eks-cluster-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = { Service = "eks.amazonaws.com" }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "eks_cluster_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.eks_cluster_role.name
}

resource "aws_iam_role_policy_attachment" "eks_vpc_resource_controller" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSVPCResourceController"
  role       = aws_iam_role.eks_cluster_role.name
}

# EKS Cluster
resource "aws_eks_cluster" "main" {
  name     = "zippyra-${var.environment}-eks"
  role_arn = aws_iam_role.eks_cluster_role.arn
  version  = "1.29"

  vpc_config {
    subnet_ids              = var.private_subnet_ids
    endpoint_private_access = true
    endpoint_public_access  = true # Restricted by CIDR in production
  }

  tags = {
    Name        = "zippyra-${var.environment}-eks"
    Environment = var.environment
  }
}

# Node Group IAM Role
resource "aws_iam_role" "eks_node_role" {
  name = "zippyra-${var.environment}-eks-node-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = { Service = "ec2.amazonaws.com" }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "eks_worker_node" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy"
  role       = aws_iam_role.eks_node_role.name
}

resource "aws_iam_role_policy_attachment" "eks_cni" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"
  role       = aws_iam_role.eks_node_role.name
}

resource "aws_iam_role_policy_attachment" "eks_ecr_readonly" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
  role       = aws_iam_role.eks_node_role.name
}

# On-Demand Node Group (20% of capacity — critical workloads)
resource "aws_eks_node_group" "on_demand" {
  cluster_name    = aws_eks_cluster.main.name
  node_group_name = "zippyra-${var.environment}-on-demand"
  node_role_arn   = aws_iam_role.eks_node_role.arn
  subnet_ids      = var.private_subnet_ids
  instance_types  = var.instance_types
  capacity_type   = "ON_DEMAND"

  scaling_config {
    desired_size = max(1, floor(var.desired_size * (100 - var.spot_percentage) / 100))
    min_size     = 1
    max_size     = max(2, floor(var.max_size * (100 - var.spot_percentage) / 100))
  }

  labels = {
    "workload-type" = "critical"
    "capacity-type" = "on-demand"
  }

  taint {
    key    = "workload-type"
    value  = "critical"
    effect = "PREFER_NO_SCHEDULE"
  }

  tags = {
    Name = "zippyra-${var.environment}-on-demand-nodes"
  }
}

# Spot Node Group (~80% of capacity — cost optimization per platform docs)
resource "aws_eks_node_group" "spot" {
  cluster_name    = aws_eks_cluster.main.name
  node_group_name = "zippyra-${var.environment}-spot"
  node_role_arn   = aws_iam_role.eks_node_role.arn
  subnet_ids      = var.private_subnet_ids
  instance_types  = var.instance_types
  capacity_type   = "SPOT"

  scaling_config {
    desired_size = max(1, ceil(var.desired_size * var.spot_percentage / 100))
    min_size     = 1
    max_size     = max(2, ceil(var.max_size * var.spot_percentage / 100))
  }

  labels = {
    "workload-type" = "general"
    "capacity-type" = "spot"
  }

  tags = {
    Name = "zippyra-${var.environment}-spot-nodes"
  }
}

# IRSA (IAM Roles for Service Accounts) — OIDC Provider
data "tls_certificate" "eks" {
  url = aws_eks_cluster.main.identity[0].oidc[0].issuer
}

resource "aws_iam_openid_connect_provider" "eks" {
  client_id_list  = ["sts.amazonaws.com"]
  thumbprint_list = [data.tls_certificate.eks.certificates[0].sha1_fingerprint]
  url             = aws_eks_cluster.main.identity[0].oidc[0].issuer
}

output "cluster_name" {
  value = aws_eks_cluster.main.name
}

output "cluster_endpoint" {
  value = aws_eks_cluster.main.endpoint
}

output "cluster_certificate_authority" {
  value = aws_eks_cluster.main.certificate_authority[0].data
}

output "oidc_provider_arn" {
  value = aws_iam_openid_connect_provider.eks.arn
}

output "oidc_provider_url" {
  value = replace(aws_eks_cluster.main.identity[0].oidc[0].issuer, "https://", "")
}
