# cluster-relay
A Relay maintaining persistent connection between serverless functions and the clusterlink gateways. 
A relay is needed when two clients want to send communicate between each other without having to host a server, and maintaining connections.
This is especially the case when we want one or more serverless functions/jobs to communicate with each other.

# Steps to run cluster-relay

   make build

   ./bin/cluster_relay start --port <portnum> --target <ip:port / Clusterlink service name>

   Refer to [tests/README.md](tests/README.md) for an end-to-end example
