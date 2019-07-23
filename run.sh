#!/usr/bin/env bash

consul_agent_host='local-consul-agent'
service_check_host="${SERVICE_CHECK_HOST:-localhost}"

echo 'Registering with consul'

service_id="prebid-server-${instance}"

curl -s http://${consul_agent_host}:8500/v1/agent/service/deregister/${service_id}
curl -s -H 'Content-Type: application/json' -X PUT -d '{
    "id": "'"${service_id}"'",
    "name": "prebid-server",
    "port": 8003,
    "checks": [
        {
            "name": "TCP port 8003",
            "http": "http://'"${service_check_host}"':8003/status",
            "interval": "10s"
        }
    ]
}' http://${consul_agent_host}:8500/v1/agent/service/register

echo "Registered prebid server with consul"

export HOST=0.0.0.0

echo "Starting Prebid"
exec  /usr/local/bin/prebid-server -v 1 -logtostderr