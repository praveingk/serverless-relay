## An end-to-end test with cluster-relay acting as the proxy service for two client applications in two clusters

make sure controlplane & dataplane are set in "/etc/hosts"
127.0.0.1 controlplane
127.0.0.1 dataplane
### Start Cluster 1 gateway VM
    sudo setcap CAP_NET_BIND_SERVICE=+eip ./bin/dataplane

    ./bin/controlplane start --id "mbg1" --ip 10.241.64.7 --cport 8443 --cportLocal 8443  --dataplane mtls --certca demos/utils/mtls/ca.crt --cert demos/utils/mtls/mbg1.crt --key demos/utils/mtls/mbg1.key --startPolicyEngine=True --observe=True --logFile=True --zeroTrust=False --rtenv=vm --profilePort=8000

    ./bin/dataplane --id mbg1 --dataplane mtls --certca demos/utils/mtls/ca.crt --cert demos/utils/mtls/mbg1.crt --key demos/utils/mtls/mbg1.key --controlplane controlplane:8443 --profilePort=8001

    ./bin/gwctl init --id gwctl1 --gwIP 10.241.64.7  --gwPort 443 --dataplane mtls --certca demos/utils/mtls/ca.crt --cert demos/utils/mtls/mbg1.crt --key demos/utils/mtls/mbg1.key

### Start Cluster 2 gateway VM

    sudo setcap CAP_NET_BIND_SERVICE=+eip ./bin/dataplane

    ./bin/controlplane start --id "mbg2" --ip 10.241.64.12 --cport 8443 --cportLocal 8443  --dataplane mtls --certca demos/utils/mtls/ca.crt --cert demos/utils/mtls/mbg2.crt --key demos/utils/mtls/mbg2.key --startPolicyEngine=True --observe=True --logFile=True --zeroTrust=False --rtenv=vm --profilePort=8000

    ./bin/dataplane --id mbg2 --dataplane mtls --certca demos/utils/mtls/ca.crt --cert demos/utils/mtls/mbg2.crt --key demos/utils/mtls/mbg2.key --controlplane controlplane:8443 --profilePort=8001

    ./bin/gwctl init --id gwctl2 --gwIP 10.241.64.12  --gwPort 443 --dataplane mtls --certca demos/utils/mtls/ca.crt --cert demos/utils/mtls/mbg2.crt --key demos/utils/mtls/mbg2.key

### Connect both the gateways
In gateway 1: 
    ./bin/gwctl create peer --myid gwctl1 --name mbg2 --host 10.241.64.12 --port 443 

In gateway 2:
    ./bin/gwctl create peer --myid gwctl2 --name mbg1 --host 10.241.64.7 --port 443

### Add Cluster 1 relay service (in gateway 1)
    ./bin/gwctl create export --myid gwctl1 --name iperf3-client --host iperf3-client --port 5000 

### Add Cluster 2 relay service (in gateway 2)
    ./bin/gwctl --myid gwctl2 create export --name iperf3-server --host iperf3-server --port 5000 


    ./bin/gwctl --myid gwctl1 create import --name iperf3-server --host iperf3-server --port 5000
    ./bin/gwctl --myid gwctl1 create binding --import iperf3-server --peer mbg2 

Verify if services are available in either clusters using 
    ./bin/gwctl get import --myid gwctl1
### Start relay (in Cluster 1 gateway VM)
    ./bin/cluster_relay start --port 9000 --target crelay-func2

### Start relay (in Cluster 2 gateway VM)
    ./bin/cluster_relay start --port 9000 --target crelay-func1

### Start Client function (in Cluster 1) which connects to crelay1
    export TARGET_RELAY=10.241.64.4:9000
    export MESSAGE=hello
    ./bin/client_function

### Start Client function (in Cluster 2) which connects to crelay2
    export TARGET_RELAY=10.241.64.5:9000
    export MESSAGE=hi
    ./bin/client_function

Now we must observe that each clientfunction receives the message from the other clientfunction, which ensures the relay is working.

You can run the client_function app as a job in IBM Code Engine using quay.io/mcnet/client_function

TODO : Generate scripts to automate launch in IBM Code Engine.


openssl s_server -accept 3000  -CAfile demos/utils/mtls/ca.crt -cert demos/utils/mtls/mbg2.crt -key demos/utils/mtls/mbg2.key -state -www 