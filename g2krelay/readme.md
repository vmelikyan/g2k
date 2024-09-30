g2krelay is a simple services that ingests github webhooks, verifies by webhook secret and publishes messages to kafka.
Messages are published to `github.events` topic, if the topic does not exist it will be created.
Messages are published by the below rules
Key = <github-org>.<repository>.<event-type>
example `"vmelikyan.lc-test-4.pull_request"`
Value = unchanged github webhook payload
Headers = unchanged github webhook headers

*** Hardcoded to use
Partitions:     1,
ReplicationFactor: 1,

| **Variable**              | **Value**                             | **Description**                 | **Required** |
|---------------------------|---------------------------------------|---------------------------------|--------------|
| `KAFKA_BOOTSTRAP_SERVERS` | `redpanda:9092`                                | The Kafka cluster's bootstrap servers used to establish initial connection.                        | Yes          |
| `KAFKA_SECURITY_PROTOCOL` | `SASL_SSL`                                     | The security protocol for Kafka connections (e.g., SASL_SSL, PLAINTEXT).                           | No          |
| `KAFKA_SASL_MECHANISMS`   | `PLAIN`                                        | The SASL mechanism to use for authentication (e.g., PLAIN, SCRAM-SHA-256).                          | Yes          |
| `KAFKA_SASL_USERNAME`     | ""                                           | The username for SASL authentication with the Kafka cluster.                                       | Yes          |
| `KAFKA_SASL_PASSWORD`     | ""                                         | The password for SASL authentication with the Kafka cluster. **_Keep this value secure!_**          | Yes          |
| `WEBHOOK_SECRET`          | ""                                         | A secret key used to authenticate and verify incoming webhook requests. | Yes          |

---
