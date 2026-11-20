# Observed

All sends go through NoticeSender.send. repos/order-publisher/src/NoticeSender.kt:41

A nightly batch job calls the relay directly and bypasses NoticeSender. repos/order-publisher/src/jobs/NightlyBatch.kt:18

The publisher retries three times with a one second delay. repos/order-publisher/src/RelayClient.kt:77

The payload carries recipient, title, body and a route field. repos/order-publisher/src/Notice.kt:9

# Inference

The route field is probably read by both clients.

# Unknowns

Where the relay credentials are rotated.
