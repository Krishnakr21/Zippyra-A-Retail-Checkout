# Zippyra On-Premise ERP Connector Installation Guide

This guide describes how store IT administrators can set up and run the Zippyra ERP Connector on a Windows PC hosting **Tally ERP** or **Busy ERP**.

---

## Step 1: File Preparation

1. Create a dedicated folder on your store PC (e.g. `C:\ZippyraConnector\`).
2. Copy `zippyra-connector.exe` and `connector.example.yaml` into this folder.
3. Rename `connector.example.yaml` to `connector.yaml`.

---

## Step 2: Configure `connector.yaml`

Open `connector.yaml` in Notepad and update the fields using the credentials provided in the Zippyra Admin Platform:

```yaml
connection_id: "conn_live_store_101"
agent_api_key: "zip_agent_sec_xxxxxxxxx"
webhook_secret: "whsec_super_secret_xxxx"
zippyra_api_base_url: "https://api.zippyra.com"
erp_type: "TALLY" # Change to "BUSY" if using Busy ERP
erp_local_endpoint: "http://127.0.0.1:9000" # Tally default: 9000, Busy default: 8080/api
poll_interval_seconds: 60
status_server_port: 8085
```

---

## Step 3: Test Running Manual Execution

Double-click `zippyra-connector.exe` or run via Command Prompt:

```cmd
C:\ZippyraConnector\zippyra-connector.exe --config C:\ZippyraConnector\connector.yaml
```

Open a web browser and visit:
[http://127.0.0.1:8085/status](http://127.0.0.1:8085/status)

You should see:
```json
{
  "erp_health": "OK",
  "erp_type": "TALLY",
  "status": "UP",
  "pending_jobs_count": 0
}
```

---

## Step 4: Register as an Automatic Windows Service

To ensure the connector starts automatically whenever the PC reboots, register it as a Windows Service using an Administrator Command Prompt:

```cmd
sc create ZippyraConnector binPath= "C:\ZippyraConnector\zippyra-connector.exe --config C:\ZippyraConnector\connector.yaml" start= auto
sc start ZippyraConnector
```

To stop or check status:
```cmd
sc query ZippyraConnector
sc stop ZippyraConnector
```

---

## Troubleshooting & Logs

- Check `zippyra-connector.log` in `C:\ZippyraConnector\` for diagnostic logs.
- API keys and Webhook secrets are automatically masked in log files.
- If you see `DEGRADED` on `http://127.0.0.1:8085/status`, verify that Tally/Busy local ODBC/HTTP server is running and accessible on port 9000 (Tally) or 8080 (Busy).
