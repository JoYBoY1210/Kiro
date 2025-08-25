# Kiro Service Mesh 🛡️

Kiro is a secure service mesh simulator built in Go, showcasing **service-to-service communication** with **mutual TLS (mTLS)**, sidecar proxies, and **HMAC-based request verification**. Experience how modern service meshes manage and secure inter-service traffic!

---

## 🌐 Overview

A **service mesh** is an infrastructure layer that manages communication between microservices.
**Key features:**

* 🔒 Secure communication (mTLS) between services
* 📍 Service discovery & routing
* ✅ Policy enforcement
* 📊 Observability & logging

In Kiro, each service (**Dashboard, Profile, Auth**) has a **sidecar proxy** responsible for mTLS connections, request routing, and logging.
**No service communicates directly:** all requests must pass through corresponding proxies.

---

## ✨ Features

* **mTLS** between services and proxies
* **HMAC-based request authentication**
* **Sidecar proxy** for secure routing
* **Detailed logging** of requests and client certificates
* **Mesh policy enforcement**

---

## 🏗️ Architecture

All inter-service requests flow through proxies, ensuring secure, authenticated, and logged communication.

```plaintext
[Dashboard Service] <-> [Dashboard Proxy] <-> [Profile Proxy] <-> [Profile Service]
     |
     v
  Logs & HMAC verification
```

Workflow:

* Every service has its own sidecar proxy
* All outgoing requests go to its proxy
* The proxy securely transmits the request (with mTLS and optional HMAC) to the next service's proxy
* The destination proxy receives, authenticates, logs, and forwards to the target service
* No direct service-to-service calls, everything flows through proxies

---

## 💻 Running Locally

### Prerequisites

* Go 1.20+ installed

### 🔑 Generating Certificates (mTLS)

All services use mTLS, so certificates must be generated for each service and proxy.

```bash
cd security/certs
chmod +x generate_all_certs.sh
./generate_all_certs.sh
```

This will generate:

* CA certificate (KiroCA)
* Service certificates for Dashboard, Profile, Auth
* Proxy certificates

### 🚀 Starting the Services

All services and proxies can be started from the main file:

```bash
go run main.go
```

All services listen securely on 127.0.0.1 so that no external entity can access them.

### 🛠️ Development Workflow

For a dev environment (no real domain names), update your /etc/hosts file as follows:

```text
127.0.0.1       profileService
127.0.0.1       dashboardService
127.0.0.1       authService
127.0.0.1       auth
127.0.0.1       dashboard
127.0.0.1       profile
```

This allows proxies/services to resolve each other as named endpoints (for certificates) even without proper DNS.

### 📝 Notes

* All services communicate only via their sidecar proxy
* Hostnames in certificates must match Subject Alternative Names (SAN)
* HMAC authentication is optional but strongly recommended for extra security
* Project logs every request and certificate verification

### 🚧 Future Enhancements

* Mesh-wide health checks
* Rate limiting & circuit breakers
* Dynamic service discovery
* Kubernetes/Docker Compose integration
