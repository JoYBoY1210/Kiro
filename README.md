Kiro Service Mesh 🛡️

Kiro is a secure service mesh simulator built in Go, demonstrating service-to-service communication with mutual TLS (mTLS), sidecar proxies, and HMAC-based request verification. It shows how modern service meshes manage and secure inter-service traffic.

Overview 🌐

A service mesh is an infrastructure layer that handles communication between microservices. Key features include:

Secure communication (mTLS) between services 🔒

Service discovery and routing

Policy enforcement ✅

Observability and logging 📊


In Kiro, each service (Dashboard, Profile, Auth) has a sidecar proxy that handles mTLS connections, request routing, and logging. Services never communicate directly; all requests pass through their proxies.

Features ✨

mTLS between services and proxies

HMAC-based request authentication

Sidecar proxy for secure routing

Detailed logging of requests and client certificates

Mesh policy enforcement


Architecture 🏗️

[dashboardService] <--> [dashboardService Proxy] <--> [profileService Proxy] <--> [profileService]
                            |
                            v
                  Logs & HMAC verification

All inter-service requests go through proxies, ensuring secure, authenticated, and logged communication.

Running Locally 💻

Prerequisites

Go 1.20+ installed


Generating Certificates 🔑

All services use mTLS, so certificates are required for each service and proxy. You can generate them with the provided script:

cd security/certs
chmod +x generate_all_certs.sh
./generate_all_certs.sh

This will generate:

CA certificate (KiroCA)

Service certificates for each service (dashboardService, profileService, authService)

Proxy certificates


Modifying Hosts File for Development 🏠

⚠️ Important for local dev: Since there are no real domains, you need to add all service names with IP 127.0.0.1 (or your localhost IP) in your /etc/hosts file:

127.0.0.1 authService
127.0.0.1 dashboardService
127.0.0.1 profileService

This ensures that mTLS hostname verification works correctly.

Starting the Services 🚀

All services can be started from the main.go file. Simply run:

go run .

This will start all services along with their proxies. By default, services listen on 127.0.0.1 for security.

Notes 📝

Services communicate only via their sidecar proxy

Hostnames must match the SANs in certificates

HMAC authentication is optional but recommended for extra security

The project includes logging for every request and certificate verification


Future Enhancements 🚧

Mesh-wide health checks

Rate limiting and circuit breakers

Dynamic service discovery

Integration with Kubernetes or Docker Compose


