# Secure Token Service (STS) & OIDC Demo on Kubernetes

![Architecture Status](https://img.shields.io/badge/Architecture-Microservices-blue?style=flat-square)
![Language](https://img.shields.io/badge/Go-1.23-cyan?style=flat-square&logo=go)
![Language](https://img.shields.io/badge/Python-FastAPI-yellow?style=flat-square&logo=python)
![Infrastructure](https://img.shields.io/badge/Kubernetes-K3s-green?style=flat-square&logo=kubernetes)

Una implementación de referencia de un sistema de autenticación distribuida basada en **Criptografía Asimétrica (RSA)** y **Zero Trust Architecture**.

Este proyecto demuestra cómo desacoplar la **generación de identidades** (STS en Go) del **consumo de recursos** (App en Python), desplegado sobre Kubernetes utilizando contenedores optimizados y gestión segura de secretos.

---

## 🏗 Arquitectura

El sistema se compone de dos microservicios que no comparten contraseñas ni bases de datos, confiando únicamente en criptografía matemática.

### Componentes

1.  **STS (Secure Token Service) - *Golang*:**
    * Actúa como **Identity Provider (IdP)**.
    * Gestiona el registro (`/register`) y login (`/login`) de usuarios (Base de datos en memoria para demo).
    * Firma tokens **JWT (RS256)** usando una **Clave Privada** montada como Kubernetes Secret.
    * Expone su **Clave Pública** en un endpoint (`/public-key`) para validación externa.

2.  **Resource Server (App Cliente) - *Python (FastAPI)*:**
    * Servicio protegido que requiere autenticación.
    * **Stateless:** No tiene base de datos de usuarios.
    * Valida los tokens recibidos consultando dinámicamente la Clave Pública del STS vía DNS interno de Kubernetes (`http://sts-service`).

3.  **Infraestructura - *Kubernetes (K3s)*:**
    * Gestión de secretos (`Secrets`) para inyectar claves `.pem`.
    * Service Discovery para comunicación interna.
    * Imágenes Docker optimizadas (Multi-stage builds) importadas directamente al runtime Containerd.

---

## 📂 Estructura del Proyecto

```text
.
├── k8s/                # Manifiestos de Kubernetes (Deployments, Services)
├── sts-core/           # Código fuente del STS (Go) + Dockerfile Multi-stage
├── app-client/         # Código fuente de la App (Python) + Dockerfile
├── start.sh            # Script de automatización "Zero-Install"
├── .gitignore          # Reglas de seguridad (Ignora claves privadas)
└── README.md           # Documentación

## ⚙️ Cómo usar este proyecto

Este laboratorio sigue la filosofía **"Zero Host Install"**. No necesitas instalar Go, Python ni OpenSSL en tu máquina local. Todo el entorno de construcción y despliegue se gestiona mediante contenedores y el script de automatización.

### Requisitos Previos

* **Sistema Operativo:** Linux (Debian/Ubuntu) o Windows con WSL2.
* **Docker:** Motor de contenedores activo.
* **Kubernetes:** Un clúster funcional (se recomienda **K3s** por su ligereza).
* **Kubectl:** Configurado para conectar con tu clúster.

### Instalación Automática

El script `start.sh` incluido actúa como orquestador de todo el ciclo de vida. Realiza las siguientes tareas secuencialmente:
1.  Genera nuevas claves RSA usando un contenedor efímero de OpenSSL.
2.  Compila las imágenes Docker de los microservicios.
3.  Importa las imágenes al registro interno de K3s (Containerd).
4.  Crea los Secretos y aplica los manifiestos de Kubernetes.

**Pasos para desplegar:**

1.  **Clonar el repositorio:**
    ```bash
    git clone <URL_DEL_REPOSITORIO>
    cd proyecto-sts
    ```

2.  **Lanzar el entorno:**
    ```bash
    chmod +x start.sh
    ./start.sh
    ```

3.  **Verificar el despliegue:**
    Al finalizar el script, asegúrate de que ambos servicios (`sts-deployment` y `app-deployment`) estén en estado `Running`:
    ```bash
    kubectl get pods
    ```