---
type: Workflow Plan
schema: workflow/plan/1
title: Retry policy for the payment client
description: Payments are retried safely after a transient failure.
generated: {by: writing-plan/0.1.0, at: 2026-10-20T09:00:00+02:00}
status: draft
---

# Objective

Payments that fail with a transient error are retried by the client within a limit that protects the gateway.

<!-- FIXME: should a payment be retried on a different gateway region after the third failure? -->

# Scope

In scope:

- The payment client retry behavior

Out of scope:

- Changes to the gateway

# Success criteria

- A payment that fails twice with a transient error succeeds on the third attempt in the staging environment

<!-- generated:begin -->

## Decisions

- [Retry limit](/decisions/retry-limit-a2b3.md)

## Work

- [Retry transient payment failures](/work/retry-transient-failures-c4d5.md)

<!-- generated:end -->
