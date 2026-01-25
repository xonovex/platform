# advanced-patterns: Advanced Terraform Patterns

**Guideline:** Use data sources, dynamic blocks, and for_each/count appropriately

**Rationale:** Advanced patterns enable flexible, maintainable infrastructure code

**Example:**

**Data Sources:**

```hcl
data "aws_ami" "amazon_linux_2" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn2-ami-hvm-*-x86_64-gp2"]
  }
}

data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_instance" "app" {
  ami           = data.aws_ami.amazon_linux_2.id
  instance_type = "t3.micro"
}
```

**Dynamic Blocks:**

```hcl
resource "aws_security_group" "this" {
  name   = "${var.environment}-sg"
  vpc_id = var.vpc_id

  dynamic "ingress" {
    for_each = var.ingress_rules

    content {
      from_port   = ingress.value.from_port
      to_port     = ingress.value.to_port
      protocol    = ingress.value.protocol
      cidr_blocks = ingress.value.cidr_blocks
    }
  }
}
```

**For_Each vs Count:**

```hcl
# ✅ Use for_each for map-based resources
resource "aws_subnet" "private" {
  for_each = local.subnet_cidrs

  vpc_id            = aws_vpc.this.id
  cidr_block        = each.value
  availability_zone = each.key

  tags = {
    Name = "${var.environment}-private-${each.key}"
  }
}

# ✅ Use count for simple repetition
resource "aws_instance" "worker" {
  count = var.worker_count

  ami           = data.aws_ami.amazon_linux_2.id
  instance_type = "t3.micro"

  tags = {
    Name = "${var.environment}-worker-${count.index + 1}"
  }
}

# ❌ Avoid count with lists (order matters, causes unnecessary recreation)
```

**Techniques:**
- **Data sources:** Query existing AWS resources without importing or creating them
- **Dynamic blocks:** Generate repeated nested blocks from variables or locals
- **For_each with maps:** Maintain consistent resource names when order changes
- **Count for simple repetition:** Use only for stateless resource duplication
- **Avoid count with lists:** Order changes cause unnecessary recreation and state corruption
