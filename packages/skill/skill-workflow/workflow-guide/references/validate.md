# Validate

1. Resolve the exact subject, optional subject revision, supplied context, binding
   criteria, advisory perspectives, requested method, and applicable capabilities.
2. Block when no binding criterion can be resolved.
3. When independence is requested or required, pin the subject and evaluate it without
   implementation context or prior findings. Preserve the first-pass results before
   resolving supplied context; block when mandatory isolation is unavailable.
4. Resolve context in a second pass, use it to understand intent, constraints, and
   known tradeoffs, and report any changed or additional result. Context alone cannot
   make a criterion pass. Without independence, resolve context before the single
   validation pass.
5. Return `pass`, `fail`, or `blocked` for every binding criterion.
6. Report advisory checks separately from the binding outcome and preserve each
   evidence entry's source, subject revision, freshness, and limitation.

Do not revise, accept, publish, or mutate the subject.
