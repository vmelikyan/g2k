# g2krepeater

Consumes kafka messages published by g2krelay and sends a POST request to specified endpoint.
The body is guaranteed to be the same as the original webhook event.
The headers are all preserved as well, one additional header is added
`g2krepeater=true`
Make sure to set `REPLAY_ENDPOINT` to the endpoint where the webhooks should be replayed to.
Currently only one endpoint is supported. I do plan to add support for multiple endpoints.

## Repository Filtering

g2krepeater supports two types of repository filtering:

- **Inclusion filtering (`REPO_FILTERS`)**: When set, only events from specified repositories are processed.
- **Exclusion filtering (`REPO_EXCLUDE`)**: When set (and `REPO_FILTERS` is empty), events from specified repositories are excluded.

**Note**: If both are set, `REPO_FILTERS` takes precedence and `REPO_EXCLUDE` is ignored.

| **Variable**              | **Value**                             | **Description**                 | **Required** |
|---------------------------|---------------------------------------|---------------------------------|--------------|
| `KAFKA_BOOTSTRAP_SERVERS` | `redpanda:9092`                                | The Kafka cluster's bootstrap servers used to establish initial connection.                        | Yes         |
| `KAFKA_SECURITY_PROTOCOL` | `SASL_SSL`                                     | The security protocol for Kafka connections (e.g., SASL_SSL, PLAINTEXT).                           | No          |
| `KAFKA_SASL_MECHANISMS`   | `PLAIN`                                        | The SASL mechanism to use for authentication (e.g., PLAIN, SCRAM-SHA-256).                         | No          |
| `KAFKA_SASL_USERNAME`     | ""                                             | The username for SASL authentication with the Kafka cluster.                                       | No          |
| `KAFKA_SASL_PASSWORD`     | ""                                             | The password for SASL authentication with the Kafka cluster. **_Keep this value secure!_**         | No          |
| `KAFKA_GROUP_ID`          | `g2krepeater-default`                          | Consumer group id                                  | No |
| `REPO_FILTERS`            | ""                                             | List of comma separated repositories to process the events for. If empty or not specified will process all.  (e.g `org/repo`)        | No |
| `REPO_EXCLUDE`            | ""                                             | List of comma separated repositories to exclude from processing. Only used when REPO_FILTERS is empty or not specified. (e.g `org/repo1,org/repo2`)        | No |
| `REPLAY_ENDPOINTS`     | "<http://host.docker.internal:5001/api/webhooks/github>"                                         | Comma separated endpoints to replay the webhooks to.          | Yes          |
