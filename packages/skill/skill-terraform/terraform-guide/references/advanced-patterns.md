# advanced-patterns: Dynamic Blocks, for_each, count

- **Data sources** query existing infra without creating/importing it (`data "aws_ami" "x" { ... }` → `data.aws_ami.x.id`).
- **Dynamic blocks** generate repeated nested blocks from a collection: `dynamic "ingress" { for_each = var.rules; content { ... ingress.value.x } }`.
- **for_each** (maps/sets) keys resources stably — reordering or removing one key doesn't churn the others. Use it for anything with a natural identity.
- **count** only for stateless repetition (`count.index`). Never `count` over a list: inserting/removing an element shifts indices and recreates every subsequent resource.

```hcl
# ✅ for_each — keyed, stable across reorders
resource "aws_subnet" "private" {
  for_each          = local.subnet_cidrs # map az => cidr
  vpc_id            = aws_vpc.this.id
  cidr_block        = each.value
  availability_zone = each.key
}

# ✅ count — simple stateless repetition
resource "aws_instance" "worker" {
  count         = var.worker_count
  ami           = data.aws_ami.amazon_linux_2.id
  instance_type = "t3.micro"
}
```
