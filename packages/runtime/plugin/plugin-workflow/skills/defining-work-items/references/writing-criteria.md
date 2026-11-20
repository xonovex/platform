# Writing criteria

## What a criterion is

A criterion is a condition about the delivered result that a check can fail. It has an `id` (`c1`, `c2`, unique in the file), a `text` that states the condition, and a `verification` that says how the condition is checked. The criteria together say when the item is done; the objective says why.

## What a verification is

The verification names the check precisely enough that someone who did not write the item can run it: a command and the result it must show, a request and the response it must return, a test and what it asserts, a screenshot and what it must show, or a question a named kind of reviewer answers.

## Examples

| Weak | Observable | Verification |
| --- | --- | --- |
| The export works as expected | An invoice issued in the test environment appears on the feed within 60 seconds | Issue a test invoice and read the feed. |
| The code is clean | The reporting module has no reference to the legacy export option | Search the module for the option name and expect no match. |
| The page is fast | The product page responds in under 300 ms at the 95th percentile on the test data set | Run the load script against the test environment and read the 95th percentile. |
| Error handling is added | A request with an unknown account returns status 404 and an error body that names the account | Send the request to the local service and check the status and the body. |

## Rules

- One condition for each criterion. "Returns 404 and logs the account" is two criteria.
- State the observable result, never the implementation. A class, a library or a pattern is not a criterion unless the result is that a dependency is gone.
- A criterion the item alone can satisfy. If it holds only when another item is done, that item is a `depends_on`, and the criterion is checked after it.
- No criterion restates the objective, and none is a task ("add the endpoint").
- A criterion that cannot fail is removed, not softened.
- Cover every outcome the objective names. Include the negative case when the outcome has one (the old path no longer sends, the unauthorized caller is refused).
- Keep the ids of cited criteria stable. A verdict or a file elsewhere may cite `c2`; add new criteria at the end and never renumber.
- An outline criterion that is still `draft` and that nothing cites may be narrowed or split while defining the item. Keep its id on the part that continues it and give the other parts new ids at the end. Once a file elsewhere cites an id, that id keeps its meaning.
- A vague verification in an outline row ("test it", "check it works") is rewritten to a precise check. A row that waits on an open decision is left as written until the row is defined.
- Do not copy a decision's text into a criterion. Link the decision in `decisions`; a criterion states the observable consequence of it.
- No secret, token or customer data in a criterion or a verification.

## Criteria that need a person

Some checks only a person can do: a design review, a judgment on wording, a check on a physical device. Write them as criteria anyway, and make the verification name who answers which question ("The design owner confirms that the empty state matches the approved design"). Do not hide them in the objective, so a reader sees which criteria a machine can check and which wait for a person.
