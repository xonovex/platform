# typed-variables: Variable Types and Validation

**Guideline:** Use type constraints and validation blocks for all variables to catch configuration errors early

**Rationale:** Typed variables with validation provide immediate feedback and prevent invalid deployments

**Example:**

```hcl
# String with validation
variable "environment" {
  description = "Environment name"
  type        = string

  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be dev, staging, or prod"
  }
}

# Number with constraints
variable "instance_count" {
  description = "Number of instances"
  type        = number
  default     = 2

  validation {
    condition     = var.instance_count >= 1 && var.instance_count <= 10
    error_message = "Instance count must be between 1 and 10"
  }
}

# Boolean
variable "enable_monitoring" {
  description = "Enable detailed monitoring"
  type        = bool
  default     = true
}

# List
variable "allowed_cidr_blocks" {
  description = "CIDR blocks allowed to access"
  type        = list(string)
  default     = []
}

# Map
variable "tags" {
  description = "Resource tags"
  type        = map(string)
  default     = {}
}

# Object with structure
variable "database_config" {
  description = "Database configuration"
  type = object({
    instance_class    = string
    allocated_storage = number
    engine_version    = string
    backup_retention  = number
  })
  default = {
    instance_class    = "db.t3.micro"
    allocated_storage = 20
    engine_version    = "15.4"
    backup_retention  = 7
  }
}

# Optional values (Terraform 1.3+)
variable "optional_config" {
  description = "Optional configuration"
  type = object({
    required_field = string
    optional_field = optional(string, "default")
    optional_number = optional(number)
  })
}
```

**Techniques:**

- **Basic types:** Use string, number, bool, list, map for simple variables
- **Validation blocks:** Add business logic constraints like allowed values
- **Range validation:** Check numeric ranges with >= and <= operators
- **Complex types:** Use object() for structured configurations with nested fields
- **Optional fields:** Use optional(type, default) for flexible object schemas
- **Sensible defaults:** Provide practical defaults for common configurations
- **Early feedback:** Type and validation errors appear before plan phase
