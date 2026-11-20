# Shared Core Go

Use this shared Go library for common terminal, environment, logging, and shell utilities, and for the process and scratch directory ports.

## Packages

Import the package that owns the required utility.

| Package | Description |
| --- | --- |
| `pkg/colors` | Terminal color output utilities |
| `pkg/envutil` | Environment variable helpers |
| `pkg/logging` | Structured logging |
| `pkg/shell` | Shell argument safety helpers |
| `pkg/process` | Port that runs a program, streams its output and captures it |
| `pkg/process/osexec` | Process adapter that starts real programs |
| `pkg/process/processtest` | Process test double |
| `pkg/scratch` | Port for private working directories |
| `pkg/scratch/osscratch` | Scratch adapter that makes real temporary directories |
| `pkg/scratch/scratchtest` | Scratch test double |

Only `pkg/process/osexec` and `pkg/scratch/osscratch` touch the host, and only a composition root imports them. `internal/arch` enforces both.
