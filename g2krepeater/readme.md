# g2krepeater

Consumes kafka messages published by g2krelay and sends a POST request to specified endpoint.
The body is guaranteed to be the same as the original webhook event.
The headers are all preserved as well, one additional header is added
`g2krepeater=true`
Make sure to set `REPLAY_ENDPOINT` to the endpoint where the webhooks should be replayed to.
Currently only one endpoint is supported. I do plan to add support for multiple endpoints.

| **Variable**              | **Value**                             | **Description**                 | **Required** |
|---------------------------|---------------------------------------|---------------------------------|--------------|
| `KAFKA_BOOTSTRAP_SERVERS` | `redpanda:9092`                                | The Kafka cluster's bootstrap servers used to establish initial connection.                        | Yes          |
| `KAFKA_SECURITY_PROTOCOL` | `SASL_SSL`                                     | The security protocol for Kafka connections (e.g., SASL_SSL, PLAINTEXT).                           | No          |
| `KAFKA_SASL_MECHANISMS`   | `PLAIN`                                        | The SASL mechanism to use for authentication (e.g., PLAIN, SCRAM-SHA-256).                          | No          |
| `KAFKA_SASL_USERNAME`     | ""                                           | The username for SASL authentication with the Kafka cluster.                                       | No          |
| `KAFKA_SASL_PASSWORD`     | ""                                         | The password for SASL authentication with the Kafka cluster. **_Keep this value secure!_**          | No          |
| `REPLAY_ENDPOINT`     | "<http://host.docker.internal:5001/api/webhooks/github>"                                         | The endpoint to replay the webhooks to          | Yes          |
