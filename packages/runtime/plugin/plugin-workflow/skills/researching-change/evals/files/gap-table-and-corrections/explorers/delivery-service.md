# Observed

The service accepts recipients of type account and device. repos/delivery-service/src/api/Recipient.kt:12

An unknown recipient returns a 404 to the caller. repos/delivery-service/src/api/SendHandler.kt:55

No code deduplicates on an idempotency key; the header is read and dropped. repos/delivery-service/src/api/SendHandler.kt:31

Persistence of notices was removed in commit 5d0e8aa. repos/delivery-service/src/store/README.md:3

# Inference

The design page describes a plan, not the code.

# Unknowns

Whether a branch restores persistence.
