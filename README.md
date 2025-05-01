# hpc-sshot-net-operator
Slingshot Network operator is a Kubernetes operator that automates configuration of VNI partitions and VLANs to support multi-tenancy.

# Build
For building the operator, follow the following steps.
1) build the operator
1.1) build the operator
```
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o sshot-net-operator cmd/main.go
```

1.2) build docker image
```
docker build --platform=linux/amd64 -t arti.hpc.amslabs.hpecorp.net/slingshot-internal-docker-unstable-local/sshot-net-operator:1.0.0 .
```

1.3) upload docker image to artifactory
```
docker push arti.hpc.amslabs.hpecorp.net/slingshot-internal-docker-unstable-local/sshot-net-operator:1.0.0
```

2) Update helm chart
2.1) Everytime operator updates, update the chart version and app version in chart.yaml

2.2) check the helm chart
```
cd kubernetes/sshot-net-operator-helm-charts
helm lint
```
2.3) package the helm chart
```
helm package .
Successfully packaged chart and saved it to: <path of packaged heml chart>-1.0.0.tgz
```
2.4) upload helm chart to artifactory
```
TBD
```

# Deploy
To deploy Slingshot Network Operator using helm charts, follow the below steps.
```
helm install <app-name> <chart-name> --namespace="sshot-net-operator" --create-namespace 
```

