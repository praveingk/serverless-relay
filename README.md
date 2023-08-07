# cluster-relay
A Relay maintaining persistent connection between serverless functions and the clusterlink gateways to 

# Steps to run cluster-relay

   make build

   ./bin/cluster_relay start --port <portnum> --target <ip:port / Clusterlink service name>

   Refer to [tests/README.md](tests/README.md) for an end-to-end example
