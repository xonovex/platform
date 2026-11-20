# module-structure: Module Organization

Separate reusable `modules/` (grouped by capability) from per-environment root modules under `environments/`. Every module and env uses the standard `main.tf`, `variables.tf`, `outputs.tf` (+ `backend.tf` for envs); each module carries a `README.md`. Source modules by relative path (`../../modules/network`) for portability.

```
terraform/
├── modules/
│   ├── network/   {main,variables,outputs}.tf + README.md
│   ├── compute/   {main,variables,outputs}.tf + README.md
│   └── database/  {main,variables,outputs}.tf + README.md
└── environments/
    ├── dev/       {main,variables}.tf + terraform.tfvars + backend.tf
    ├── staging/   {main,variables}.tf + terraform.tfvars + backend.tf
    └── prod/      {main,variables}.tf + terraform.tfvars + backend.tf
```
