# Go Event Pipeline – Real-Time EDA with Kafka and EventBridge

This project is the engineering implementation of a Bachelor's thesis comparing **Apache Kafka (AWS MSK)** and **AWS EventBridge** as real-time event backbones for analytics pipelines.

---

## 🧠 Project Goal

To design, implement, and evaluate two parallel event-driven pipelines — one powered by **Kafka (AWS MSK)** and the other by **AWS EventBridge** — using a synthetic e-commerce use case. Both pipelines will:

- Ingest and process simulated user interaction events
- Write results to a Redis sink
- Be monitored for latency, throughput, and system health
- Be provisioned with Terraform for reproducibility

---

## 📦 Project Structure
```tree
go-event-pipeline/
    ├── cmd/ # Entry points for all services
    │ ├── kafka-producer/
    │ ├── kafka-consumer/
    │ ├── eventbridge-producer/
    │ ├── lambda-consumer/
    │ ├── load-generator/
    │ └── monitoring/
    │
    ├── internal/ # Shared Go packages
    │ ├── event/ # Event models and schema
    │ ├── kafka/ # Kafka utilities
    │ ├── eventbridge/ # EventBridge utilities
    │ ├── database/ # Database(dynamoDB) utilities
    │ ├── telemetry/ # Prometheus, logging, etc.
    │ └── config/ # Configuration loader
    │
    ├── go.mod
    ├── Makefile
    └── README.md
```
---

## 🚀 Pipelines Overview

### Kafka-Based Pipeline
- Producer sends events to Kafka topic
- Consumer reads and processes events
- Dynamo is the final data sink

### EventBridge-Based Pipeline
- Producer pushes events to EventBridge bus
- EventBridge routes to Lambda or Go consumer
- DynamoDB stores the processed events(dynamo is chosen over redis because of the intuitive UI for fetching data)

---

## 🧪 Key Evaluation Metrics

- ⏱️ End-to-end Latency
- 🚀 Throughput and Scalability
- 💸 Cost of Ownership (AWS Billing)
- 🧑‍💻 Developer Effort (Code, Setup Time)
- ⚙️ Architectural Complexity

---

## 📊 Monitoring & Observability

- **Prometheus** collects service metrics
- **Grafana** visualizes latency, throughput, etc.
- Exporters are integrated in each service via the `telemetry` package

---

## 🛠️ Tools & Technologies

- Language: **Go**
- Cloud: **AWS** (MSK, EventBridge, Lambda, Redis, IAM, Event Sourcing, VPC)
- Monitoring: **Prometheus + Grafana**
- Load Simulation: **Custom Go generator**

---

## 📚 Related Resources

- [Apache Kafka](https://kafka.apache.org/)
- [AWS EventBridge](https://docs.aws.amazon.com/eventbridge/)
- [DynamoDB](https://docs.aws.amazon.com/dynamodb)
- [Prometheus](https://prometheus.io/)
- [Grafana](https://grafana.com/)

---

## 📖 Author

**Babatunde Koiki**
Final Year B.Sc. Computer Science – IU International University  
Thesis Supervisor: Prof. Dr. Stefan Remhof  
