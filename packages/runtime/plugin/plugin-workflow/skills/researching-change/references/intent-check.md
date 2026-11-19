# Checking the intent

Check the intent once, before reading anything. The research can only be as sharp as the request, and an explorer sent after a vague subject returns a long answer to the wrong question.

## What the request must say

- **What is changing:** the behavior, feature, quality problem or product to be built.
- **What it affects:** the systems, repositories, teams or users the change touches.
- **Where to look:** the repositories, documents, environments or web pages that hold the evidence.

If all three are present, go on. If one is missing, ask one question that names it and offers the likeliest reading, and wait. Do not ask a list of questions, and do not read the repositories first to guess.

## Examples by kind

| Kind | Complete enough | One question to ask |
| --- | --- | --- |
| migration | Move the order publisher from the legacy queue to the new broker; read orders-api, broker-service and the two apps. | Which repositories hold the publisher and the broker? |
| feature | Add saved searches to the shop app; read shop-app and search-api and the product page on search. | Should saved searches live on the device or in the account, or is that part of the research? |
| quality | Find why the checkout module takes minutes to build; read the build files of checkout and the shared library. | Is build time the only quality concern, or are test flakiness and size part of it? |
| investigation | Orders are sometimes duplicated; find where; read orders-api and its message consumer; the incident page is attached. | Since which release does it happen? |
| project | A new loyalty service for the shop; no code yet; compare what the retail team and the finance team need. | None needed, if the research is the context, constraints and comparable solutions. Say so in the file. |

## A new project

A new project has no current system to read. Do not invent a producer, a target or a consumer. Say in the research file that no existing system exists, and set the roles from the concerns instead: the users and their needs, the constraints (regulation, budget, platforms), the neighboring systems the new one must work with, and comparable solutions. The facts then come from documents, interviews recorded in documents, neighboring repositories and web pages, and the file still separates observed from inferred.

## When the answer changes the subject

If the answer shows the request was about a different change, restate the subject in one sentence and continue with that. The `subject` of the research file is the restated sentence.
