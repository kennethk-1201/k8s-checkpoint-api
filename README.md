# k8s-checkpoint-api
A Kubernetes node-level API to handle checkpoint and restoration of Pods.

## Set up
```
go run main.go
```

## Useful Commands
sudo curl "https://localhost:10250/pods" \
  --cacert /etc/kubernetes/pki/ca.crt

sudo curl -k  "https://10.0.0.12:10250/pods" \
  --key /etc/kubernetes/pki/apiserver-kubelet-client.key \
  --cacert /etc/kubernetes/pki/ca.crt \
  --cert /etc/kubernetes/pki/apiserver-kubelet-client.crt

curl -k --header "Authorization: Bearer $TOKEN"  https://10.0.0.10/api/v1/nodes/worker-node01/proxy/logs/