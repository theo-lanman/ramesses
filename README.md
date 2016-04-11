## Sidecar

(better name TBD)

Sidecar is a kafka message forwarder which listens on HTTP, and (someday) forwards messages (batched, compressed) to Kafka brokers. Messages are stored to disk before success is returned for safety.

### Features

- A successful response guarantees that a message has been stored for delivery.
- At-least-once forwarding.

### TODO

- tests
- configurability
- support multiple topics
- actually forward messages
- benchmarks
