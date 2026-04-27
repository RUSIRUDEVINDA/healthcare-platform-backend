# Deployment Guide: Healthcare Platform on Azure (AKS + ACR)

This document outlines the steps to containerize the healthcare microservices, push them to Azure Container Registry (ACR), and deploy them to Azure Kubernetes Service (AKS).

---

## 1. Prerequisites
Before starting, ensure you have the following tools installed and configured:
- **Azure CLI**: [Install Azure CLI](https://learn.microsoft.com/en-us/cli/azure/install-azure-cli)
- **Docker**: For building images locally.
- **kubectl**: For interacting with the Kubernetes cluster.
- **Go**: 1.25.x (if building locally).

---

## 2. Infrastructure Setup (One-time)

### Login to Azure
```bash
az login
```

### Create a Resource Group
```bash
az group create --name HealthcareResourceGroup --location eastus
```

### Create Azure Container Registry (ACR)
```bash
az acr create --resource-group HealthcareResourceGroup --name healthcareacr --sku Basic
```

### Create Azure Kubernetes Service (AKS)
```bash
az aks create \
    --resource-group HealthcareResourceGroup \
    --name HealthcareCluster \
    --node-count 2 \
    --generate-ssh-keys \
    --attach-acr healthcareacr
```

### Connect to the Cluster
```bash
az aks get-credentials --resource-group HealthcareResourceGroup --name HealthcareCluster
```

---

## 3. Containerizing & Pushing to ACR

Repeat these steps for each microservice (e.g., auth-service, appointment-service).

### Login to ACR
```bash
az acr login --name healthcareacr
```

### Build and Tag the Image
From the project root:
```bash
docker build -t healthcareacr.azurecr.io/auth-service:latest -f ./services/auth-service/Dockerfile .
```

### Push to ACR
```bash
docker push healthcareacr.azurecr.io/auth-service:latest
```

---

## 4. Deploying to Kubernetes

### Create Namespace
```bash
kubectl apply -f k8s/namespace.yaml
```

### Deploy Infrastructure (Postgres, RabbitMQ, etc.)
```bash
kubectl apply -f k8s/infrastructure/
```

### Update Manifests
Ensure the `image:` field in each service's deployment YAML (e.g., `k8s/auth-service/deployment.yaml`) points to your ACR:
`image: healthcareacr.azurecr.io/auth-service:latest`

### Apply Service Manifests
```bash
kubectl apply -f k8s/auth-service/
kubectl apply -f k8s/appointment-service/
# ... and so on
```

### Apply Ingress
```bash
kubectl apply -f k8s/ingress.yaml
```

---

## 5. CI/CD Integration (GitHub Actions)
To automate this, update your `.github/workflows/` files:

1. **Service Principal**: Create a service principal for GitHub to access Azure.
2. **Secrets**: Add `AZURE_CREDENTIALS` and `ACR_URL` to your GitHub repo secrets.
3. **Workflow Update**:
   - Use `azure/docker-login@v1` to log in to ACR.
   - Use `azure/aks-set-context@v3` to set the K8s context.
   - Update the `build-and-push` and `deploy` jobs to use ACR image tags.

---

## 6. Verification Commands
- **Check Pods**: `kubectl get pods -n healthcare`
- **Check Services**: `kubectl get svc -n healthcare`
- **Check Ingress**: `kubectl get ingress -n healthcare`
- **View Logs**: `kubectl logs -f <pod-name> -n healthcare`
