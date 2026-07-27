# Runbook: Kafka Consumer Lag Spike (SEV1 / SEV2)

## Symptoms
- Alert: `KafkaConsumerLagHighP1` (lag > 1000 on payment/order/exit groups)
- Alert: `KafkaConsumerLagP2` (lag > 10000 on analytics groups)

## Immediate Actions

1. **Identify High-Lag Consumer Group**:
   ```bash
   kubectl exec -n kafka msk-broker-0 -- \
     kafka-consumer-groups.sh --bootstrap-server localhost:9092 --describe --all-groups
   ```

2. **Scale Consumer Service Replicas**:
   - Increase replica count for lagging consumer service (e.g. order-service or inventory-service):
   ```bash
   kubectl scale deployment/inventory-service --replicas=6 -n zippyra-prod
   ```

3. **Check for Dead Letter Queue (DLQ) Messages**:
   ```bash
   kubectl exec -n kafka msk-broker-0 -- \
     kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic dlq.order --from-beginning --max-messages 10
   ```
