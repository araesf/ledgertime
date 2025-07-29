terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

# VPC for Ledgertime infrastructure
resource "aws_vpc" "ledgertime_vpc" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name        = "ledgertime-vpc"
    Environment = var.environment
    Project     = "ledgertime"
  }
}

# Internet Gateway
resource "aws_internet_gateway" "ledgertime_igw" {
  vpc_id = aws_vpc.ledgertime_vpc.id

  tags = {
    Name = "ledgertime-igw"
  }
}

# Public Subnets
resource "aws_subnet" "public_subnets" {
  count             = 2
  vpc_id            = aws_vpc.ledgertime_vpc.id
  cidr_block        = "10.0.${count.index + 1}.0/24"
  availability_zone = data.aws_availability_zones.available.names[count.index]

  map_public_ip_on_launch = true

  tags = {
    Name = "ledgertime-public-${count.index + 1}"
  }
}

# Private Subnets for Database
resource "aws_subnet" "private_subnets" {
  count             = 2
  vpc_id            = aws_vpc.ledgertime_vpc.id
  cidr_block        = "10.0.${count.index + 10}.0/24"
  availability_zone = data.aws_availability_zones.available.names[count.index]

  tags = {
    Name = "ledgertime-private-${count.index + 1}"
  }
}

# RDS PostgreSQL Database
resource "aws_db_subnet_group" "ledgertime_db_subnet_group" {
  name       = "ledgertime-db-subnet-group"
  subnet_ids = aws_subnet.private_subnets[*].id

  tags = {
    Name = "Ledgertime DB subnet group"
  }
}

resource "aws_db_instance" "ledgertime_postgres" {
  identifier = "ledgertime-postgres"
  
  engine         = "postgres"
  engine_version = "15.4"
  instance_class = "db.t3.micro"
  
  allocated_storage     = 20
  max_allocated_storage = 100
  storage_type          = "gp2"
  storage_encrypted     = true
  
  db_name  = "ledgertime"
  username = "ledgertime_user"
  password = var.db_password
  
  vpc_security_group_ids = [aws_security_group.rds_sg.id]
  db_subnet_group_name   = aws_db_subnet_group.ledgertime_db_subnet_group.name
  
  backup_retention_period = 7
  backup_window          = "03:00-04:00"
  maintenance_window     = "sun:04:00-sun:05:00"
  
  skip_final_snapshot = true
  deletion_protection = false

  tags = {
    Name        = "ledgertime-postgres"
    Environment = var.environment
  }
}

# EKS Cluster for Microservices
resource "aws_eks_cluster" "ledgertime_cluster" {
  name     = "ledgertime-cluster"
  role_arn = aws_iam_role.eks_cluster_role.arn
  version  = "1.28"

  vpc_config {
    subnet_ids = aws_subnet.public_subnets[*].id
  }

  depends_on = [
    aws_iam_role_policy_attachment.eks_cluster_policy,
  ]

  tags = {
    Name        = "ledgertime-eks"
    Environment = var.environment
  }
}

# MSK Kafka Cluster
resource "aws_msk_cluster" "ledgertime_kafka" {
  cluster_name           = "ledgertime-kafka"
  kafka_version          = "3.5.1"
  number_of_broker_nodes = 2

  broker_node_group_info {
    instance_type   = "kafka.t3.small"
    client_subnets  = aws_subnet.private_subnets[*].id
    security_groups = [aws_security_group.msk_sg.id]
    
    storage_info {
      ebs_storage_info {
        volume_size = 100
      }
    }
  }

  tags = {
    Name        = "ledgertime-kafka"
    Environment = var.environment
  }
}

# Security Groups
resource "aws_security_group" "rds_sg" {
  name_prefix = "ledgertime-rds-"
  vpc_id      = aws_vpc.ledgertime_vpc.id

  ingress {
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = [aws_vpc.ledgertime_vpc.cidr_block]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "msk_sg" {
  name_prefix = "ledgertime-msk-"
  vpc_id      = aws_vpc.ledgertime_vpc.id

  ingress {
    from_port   = 9092
    to_port     = 9092
    protocol    = "tcp"
    cidr_blocks = [aws_vpc.ledgertime_vpc.cidr_block]
  }
}

# IAM Role for EKS
resource "aws_iam_role" "eks_cluster_role" {
  name = "ledgertime-eks-cluster-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "eks.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "eks_cluster_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.eks_cluster_role.name
}

# Data sources
data "aws_availability_zones" "available" {
  state = "available"
}
