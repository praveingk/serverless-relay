## An end-to-end test with cluster-relay acting as the proxy service for two client applications in two clusters.

In this scenario, there are applications running in two clusters (or cloud providers). For each cluster, we provision a Clusterlink gateway VM to enable connectivity with the other cluster. In this specific example, a cluster is a set of serverless functions running in a certain domain. The aim of this exercise is to achieve end-to-end connectivity between certain serverless functions running across two clusters/cloud providers.

## Steps to get started 

### Provision two VMs
Two VMs will need to be provisioned to host the gateways for the two clusters. 

1) Ideally setup an ubuntu VMs (with a single network interface) with default configuration provided by the cloud provider, unless there are any specific requirements.
2) Setup a floating IP/Public IP to access the VMs from public Internet. 
3) Add security rules to allow ports (443 and 9000 for the relay) to be accessible.

### Start the gateway in Cluster 1 VM
Login to VM1, and run the following commands to start the gateway with the required certificates.

    cd clusterlink-tutorial 

    make function

    export VM1_IP=`hostname -I | cut -d' ' -f1`

    sudo setcap CAP_NET_BIND_SERVICE=+eip /usr/local/bin/mbg

    mbg start --id "mbg1" --ip $VM1_IP  --cport 443 --cportLocal 443  --dataplane mtls --rootCa utils/mtls/ca.crt --certificate utils/mtls/mbg1.crt --key utils/mtls/mbg1.key --startPolicyEngine=True --deployment vm &

    mbgctl create --id "mbgctl1" --mbgIP $VM1_IP:443  --dataplane mtls --rootCa utils/mtls/ca.crt --certificate utils/mtls/mbg1.crt --key utils/mtls/mbg1.key


### Start the gateway in Cluster 2 VM
Login to VM2, and run the following commands to start the gateway with the required certificates.

    cd clusterlink-tutorial 

    make function

    export VM2_IP=`hostname -I | cut -d' ' -f1`

    sudo setcap CAP_NET_BIND_SERVICE=+eip /usr/local/bin/mbg

    mbg start --id "mbg2" --ip $VM2_IP  --cport 443 --cportLocal 443  --dataplane mtls --rootCa utils/mtls/ca.crt --certificate utils/mtls/mbg2.crt --key utils/mtls/mbg2.key --startPolicyEngine=True --deployment vm &

    mbgctl create --id "mbgctl2" --mbgIP $VM2_IP:443  --dataplane mtls --rootCa utils/mtls/ca.crt --certificate utils/mtls/mbg2.crt --key utils/mtls/mbg2.key

Now, ensure VM1 and VM2 have a public IP, and the ports are allowed to be reachable. Lets call them $VM1_PUBLIC_IP, and $VM2_PUBLIC_IP

### Connect both the gateways (in gateway 2)
    mbgctl add peer  --id mbg1 --target $VM1_PUBLIC_IP --port 443
    mbgctl hello

### Add Cluster 1 relay service (in gateway 1)
    mbgctl add service --id crelay-func1 --target $VM1_IP --port 9000
    mbgctl expose --service crelay-func1

### Add Cluster 2 relay service (in gateway 2)
    mbgctl add service --id crelay-func2 --target $VM2_IP --port 9000
    mbgctl expose --service crelay-func2

Verify if services are available in either clusters using 
    mbgctl get service

### Start relay (in Cluster 1 gateway VM)
    cluster_relay start --port 9000 --target crelay-func2

### Start relay (in Cluster 2 gateway VM)
    cluster_relay start --port 9000 --target crelay-func1

### Start a normal TCP Client function from a remote cluster  which connects to crelay1
    docker run -e "TARGET=$VM1_IP:9000" -e "MESSAGE=hello_1_func" quay.io/mcnet/client_function:latest_tls


### Start a normal TCP Client function from a remote cluster which connects to crelay2
    docker run -e "TARGET=$VM2_PUBLIC_IP:9000" -e "MESSAGE=hello_2_func" quay.io/mcnet/client_function:latest_tls


### Additionally, you could start a function that uses TLS(client) from a remote cluster which connects to crelay1
    docker run -e "TARGET=$VM1_PUBLIC_IP:9000" -e "MESSAGE=hello_1_func" -e "MODE=tls_client" quay.io/mcnet/client_function:latest_tls


### And, you will need to start a function that uses TLS(server) from a remote cluster which connects to crelay2
    docker run -e "TARGET=$VM2_PUBLIC_IP:9000" -e "MESSAGE=hello_2_func" -e "MODE=tls_server" quay.io/mcnet/client_function:latest_tls
    
You can run the client_function app as a job in IBM Code Engine using *quay.io/mcnet/client_function:latest* ,  ensure the environment variables *TARGET_RELAY*,  *MESSAGE* and *MODE* are set accordingly.

