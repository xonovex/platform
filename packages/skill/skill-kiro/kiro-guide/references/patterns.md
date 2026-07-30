## Kiro mappings

| Pattern                               | Native mapping                                                                  | Caveat                                                     |
| ------------------------------------- | ------------------------------------------------------------------------------- | ---------------------------------------------------------- |
| Protected path / secret / tool policy | `PreToolUse` command action                                                     | Exit `2` is the documented block; validate tool categories |
| Prompt policy                         | `UserPromptSubmit` command action                                               | Prompt is sensitive input                                  |
| Formatting                            | `PostFileSave` or `PostToolUse` command                                         | Keep file pattern narrow and action idempotent             |
| Validation before spec task           | `PreTaskExec` command                                                           | Covers native spec-task execution only                     |
| Audit                                 | Post-event command                                                              | Minimize prompt/tool/file content                          |
| Context injection                     | Successful command stdout on documented events                                  | Treat as model context                                     |
| Model evaluator                       | Agent action or explicit command runner                                         | Bound credits, purpose, output schema, and failure         |
| Specialist agent                      | Explicit bounded command runner                                                 | No hidden recursion or ambient authority                   |
| Organization-managed                  | Version-controlled workspace hook plus independent repository/endpoint controls | Native managed precedence is not documented here           |

A shared repository hook is executable project content. Establish repository trust and review before enabling it for contributors.
