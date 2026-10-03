# advanced-patterns: Data Sources and Dynamic Blocks

Query existing infrastructure with data sources and generate repeated nested configuration with dynamic blocks.

- **Data sources** query existing infra without creating/importing it (`data "aws_ami" "x" { ... }` → `data.aws_ami.x.id`).
- **Dynamic blocks** generate repeated nested blocks from a collection: `dynamic "ingress" { for_each = var.rules; content { ... ingress.value.x } }`.
