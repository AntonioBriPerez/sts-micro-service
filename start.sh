#!/bin/bash

# Colores para logs
GREEN='\033[0;32m'
NC='\033[0m' # No Color

echo -e "${GREEN}>>> 1. Preparando entorno...${NC}"
mkdir -p keys

echo -e "${GREEN}>>> 2. Generando claves RSA nuevas (Sin OpenSSL local)...${NC}"
# Usamos Docker para generar la clave, igual que hicimos en la clase
docker run --rm -v "$(pwd)/keys:/keys" -w /keys alpine/openssl \
  genrsa -out sts_privada.pem 2048

chmod 600 keys/sts_privada.pem
echo -e "${GREEN}>>> 3. Construyendo Imágenes Docker...${NC}"
docker build -t mi-sts:latest ./sts-core
docker build -t mi-app-python:latest ./app-client

echo -e "${GREEN}>>> 4. Importando imágenes a K3s (Containerd)...${NC}"
# Exportamos de Docker -> Importamos en K3s
docker save mi-sts:latest | sudo k3s ctr images import -
docker save mi-app-python:latest | sudo k3s ctr images import -

echo -e "${GREEN}>>> 5. Configurando Secretos en Kubernetes...${NC}"
# Borramos el secreto si existe para recrearlo con la clave nueva
kubectl delete secret sts-secret --ignore-not-found
kubectl create secret generic sts-secret --from-file=sts_privada.pem=./keys/sts_privada.pem

echo -e "${GREEN}>>> 6. Desplegando en Kubernetes...${NC}"
# Asegúrate de que tus YAMLs usen la etiqueta :latest o la que definas aquí
# Vamos a usar sed para forzar que usen 'latest' dinámicamente si quieres, 
# o simplemente asegúrate de que tus yamls dicen image: mi-sts:latest
kubectl apply -f k8s/sts.yaml
kubectl apply -f k8s/app.yaml

echo -e "${GREEN}>>> 7. Reiniciando Pods para aplicar cambios...${NC}"
kubectl rollout restart deployment sts-deployment
kubectl rollout restart deployment app-deployment

echo -e "${GREEN}>>> ¡LISTO! Esperando a que los pods arranquen...${NC}"
sleep 5
kubectl get pods

echo -e "${GREEN}---------------------------------------------------${NC}"
echo -e "Para probar, abre estos túneles en otras terminales:"
echo -e "  kubectl port-forward svc/sts-service 8080:80"
echo -e "  kubectl port-forward svc/app-service 3000:80"
echo -e "${GREEN}---------------------------------------------------${NC}"