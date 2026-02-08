# Agent Operator Go

Kubernetes operator for managing AI coding agent runs.

## CRDs
- **AgentRun**: Primary workload resource - creates Jobs for agent execution
- **AgentProvider**: Reusable provider configuration with K8s secret management
- **AgentConfig**: Namespace-level defaults
