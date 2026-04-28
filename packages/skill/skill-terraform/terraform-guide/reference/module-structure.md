# module-structure: Module Organization

**Guideline:** Organize modules with separate directories for reusable modules and environment-specific configurations

**Rationale:** Clear separation between modules and environments enables reusability and environment isolation

**Example:**

```
terraform/
├── modules/
│   ├── network/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   ├── outputs.tf
│   │   └── README.md
│   ├── compute/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   ├── outputs.tf
│   │   └── README.md
│   └── database/
│       ├── main.tf
│       ├── variables.tf
│       ├── outputs.tf
│       └── README.md
└── environments/
    ├── dev/
    │   ├── main.tf
    │   ├── variables.tf
    │   ├── terraform.tfvars
    │   └── backend.tf
    ├── staging/
    │   ├── main.tf
    │   ├── variables.tf
    │   ├── terraform.tfvars
    │   └── backend.tf
    └── prod/
        ├── main.tf
        ├── variables.tf
        ├── terraform.tfvars
        └── backend.tf
```

**Techniques:**

- **Modules directory:** Group reusable infrastructure components by capability
- **Environments directory:** Create separate root modules for dev, staging, prod
- **Standard files:** Use main.tf, variables.tf, outputs.tf, backend.tf consistently
- **Module README:** Document each module's purpose, variables, and outputs
- **Relative sourcing:** Use relative paths in module sources for portability
