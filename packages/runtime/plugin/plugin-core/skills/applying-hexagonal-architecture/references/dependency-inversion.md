# Dependency Inversion and the Dependency Rule

Every source-code dependency points inward toward a port the core owns, never outward at a concrete adapter. Inner code (core, policy, domain) names only inner things and ports; outer code (adapters, frameworks, I/O) names inner code. No adapter type, driver struct, client SDK, or transport name may appear in the core, if it leaked in, the boundary is fictional.

The rule is mechanically checkable: grep the core tree for any adapter symbol; a hit is a violation:

```sh
grep -rn 'PostgresRepository\|sql.Open\|postgres://' core/   # any hit = leak
```

Wiring lives in a [composition root](composition-root.md), never a service locator (which hides the dependency and re-couples the core to a global lookup).

## Instability and the Zone of Pain

A module many others depend on should be abstract (a port). Concrete **and** heavily depended-on is the "zone of pain": stable yet rigid, expensive to change. Keep stability on interfaces; let adapters stay concrete but unstable (few inbound deps) so they are cheap to swap or delete. Binding late at the composition root keeps the abstract center from hardening around one implementation.

Back to [SKILL.md](../SKILL.md).
