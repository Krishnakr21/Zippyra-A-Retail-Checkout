# Runbook: Exit Gate Alarm Spike (SEV1)

## Symptoms
- Alert: `ExitGateAlarmSpike` triggered (exit gate alarms rate > 5/s)
- Physical barrier alarms triggering at retail store exit turnstiles

## Immediate Actions

1. **Verify Redis `exit_token` Cluster**:
   - Confirm Redis `exit_token` cluster (port 6383) is reachable and memory policy is `noeviction`:
   ```bash
   kubectl exec -n monitoring deploy/prometheus-server -- \
     wget -q -O- 'http://redis-exit:6383/info' | grep "maxmemory_policy"
   ```
   - Ensure policy is `noeviction` (exit tokens must NEVER be evicted under memory pressure).

2. **Check Exit Validation Logs**:
   ```bash
   kubectl logs -n zippyra-prod -l app=exit-validation-service --tail=200 | grep "DENIED"
   ```

3. **Bypass Exit Barriers (Emergency Physical Gate Unlock)**:
   - If exit-validation-service is offline, send emergency override command to store turnstiles:
   ```bash
   curl -X POST "https://api.zippyra.com/v1/device/emergency-gate-unlock" \
     -H "Authorization: Bearer ${EMERGENCY_TOKEN}" \
     -d '{"store_id":"ALL","action":"UNLOCK_BARRIERS"}'
   ```
