# Identify and Prioritize

For each scenario and boundary, ask:

1. What are we protecting and from whom or what?
2. What can go wrong through misuse, abuse, failure, compromise, or unsafe assumption?
3. Which preconditions and attacker capabilities make the path credible?
4. Which existing controls prevent, limit, detect, or recover from it?
5. What consequence reaches users, data, operations, safety, or another trust domain?

Use one declared enumeration method when it improves coverage:

- STRIDE across relevant model elements;
- abuse or misuse cases from attacker goals;
- attack trees from a material adverse outcome;
- privacy-threat analysis for linkability, identifiability, disclosure, or unfair use;
- domain-specific failure and fraud cases.

Prioritize using explicit evidence, not an unexplained number. Record impact,
likelihood or feasibility factors, reachability, privilege, detectability, affected
population, existing controls, uncertainty, and why action is or is not urgent.

Deduplicate threats that share the same cause and mitigation, but preserve materially
different consequences or trust boundaries.
