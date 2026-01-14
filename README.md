## 🛡️ RBAC Authentication & Organization Service

This repository contains a **Role-Based Access Control (RBAC)** system designed for multi-tenant applications. It provides authentication, authorization, and organization management with a clean, extensible architecture built in Go.

The system supports **users, admins, and organizations**, enabling fine-grained permission control across resources while remaining flexible enough for SaaS and enterprise use cases.

---

## ✨ Features

- **RBAC system**
  - Role and permission definitions
  - Fine-grained access control per resource
- **Authentication**
  - Secure login flow
  - JWT generation and validation
- **Multi-tenant architecture**
  - User, admin, and organization models
  - Organization-scoped users and roles
- **User onboarding**
  - Email-based user invitations
  - Secure invite tokens with expiration
- **Persistence**
  - PostgreSQL database
  - GORM-based store layer
  - Database migrations
- **Developer experience**
  - Centralized JSON response and error handling
  - Live reload with `air`
  - Dockerized PostgreSQL setup
- **Infrastructure**
  - CI build configuration
  - Environment-based configuration

---

## 🧱 Architecture Overview

```text
cmd/
internal/
  ├── auth        # Authentication & JWT logic
  ├── handlers    # HTTP handlers
  ├── models      # Database models
  ├── store       # Database access layer
  ├── rbac        # Roles & permissions
  ├── mailer      # Email invitations
  └── utils       # JSON & error helpers
