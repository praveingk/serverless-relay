# cluster-relay
A Relay maintaining persistent connection between serverless functions and the clusterlink gateways to 

# Steps to run cluster-relay

   make build

   ./bin/cluster_relay start --port 9000 --gw 10.241.64.5 --target 10.241.64.5:9999

   