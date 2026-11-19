# Shop API

## Workflow

```yaml
plans: plans/
repositories:
  shop-api:
    resource: https://git.example.org/shop/shop-api
    host: github
delivery:
  mode: branch
```

Worktrees live next to the clone, named `shop-api-<branch>`. Commits follow Conventional Commits.
