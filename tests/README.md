## An end-to-end test with cluster-relay acting as the proxy service for two client applications in two clusters

### Start Cluster 1 gateway VM
sudo setcap CAP_NET_BIND_SERVICE=+eip ./bin/mbg
mbg start --id "mbg1" --ip 10.241.64.4  --cport 443 --cportLocal 443  --dataplane tcp --rootCa tests/utils/mtls/ca.crt --certificate tests/utils/mtls/mbg1.crt --key tests/utils/mtls/mbg1.key --startPolicyEngine=True --deployment vm &
mbgctl create --mbgIP 10.241.64.4:443  --dataplane tcp --rootCa tests/utils/mtls/ca.crt --certificate tests/utils/mtls/mbg1.crt --key tests/utils/mtls/mbg1.key

### Start Cluster 2 gateway VM
mbg start --id "mbg2" --ip 10.241.64.5  --cport 443 --cportLocal 443  --dataplane tcp --rootCa tests/utils/mtls/ca.crt --certificate tests/utils/mtls/mbg2.crt --key tests/utils/mtls/mbg2.key --startPolicyEngine=True --deployment vm &
mbgctl create --mbgIP 10.241.64.5:443  --dataplane tcp --rootCa tests/utils/mtls/ca.crt --certificate tests/utils/mtls/mbg2.crt --key tests/utils/mtls/mbg2.key

### Connect both the gateways (in gateway 2)
mbgctl add peer  --id mbg1 --target 10.241.64.4 --port 443
mbgctl hello

### Add Cluster 1 relay service (in gateway 1)
mbgctl add service --id crelay-func1 --target 127.0.0.1 --port 9000
mbgctl expose --service crelay-func1


### Add Cluster 2 relay service (in gateway 2)
mbgctl add service --id crelay-func2 --target 127.0.0.1 --port 9000
mbgctl expose --service crelay-func2

### Start relay (in Cluster 1 gateway VM)
mbgctl get service 
./bin/crelay start --port 9000 --target crelay-func2

### Start relay (in Cluster 2 gateway VM)
./bin/crelay start --port 9000 --target crelay-func1
