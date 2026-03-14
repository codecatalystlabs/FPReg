# HMIS MCH 007 – Integrated Family Planning Register

A production-ready, API-first digital register system for the Uganda Ministry of Health HMIS Form 007 — Integrated Family Planning Register (Print Version July 2024).

## Tech Stack

| Layer       | Technology                         |
|-------------|-------------------------------------|
| Backend     | Go 1.22+ with Gin                  |
| ORM         | GORM                               |
| Database    | PostgreSQL 14+                     |
| Frontend    | HTML, Bootstrap 5, vanilla JS      |
| Auth        | JWT (access + refresh tokens)      |
| API Docs    | Swagger / OpenAPI via swag         |
| Audit       | Append-only audit log table        |

## Project Structure

```
FPReg/
├── cmd/server/main.go          # Application entry point
├── internal/
│   ├── config/                  # Environment configuration
│   ├── database/                # DB connection, migrations, seed data
│   ├── handler/                 # HTTP handlers (controllers)
│   ├── middleware/              # Auth, CORS, audit middleware
│   ├── models/                  # GORM models
│   ├── repository/              # Data access layer
│   ├── routes/                  # Route registration
│   ├── service/                 # Business logic
│   └── utils/                   # Response helpers, validators
├── migrations/                  # Raw SQL migrations (reference)
├── web/
│   ├── static/css/              # Stylesheets
│   ├── static/js/               # Client-side JavaScript
│   └── templates/               # HTML templates
├── docs/                        # Swagger docs
├── .env.example                 # Environment template
├── go.mod
└── README.md
```

## Quick Start

### Prerequisites

- Go 1.22+
- PostgreSQL 14+
- Git

### 1. Clone & configure

```bash
git clone <repo-url> FPReg
cd FPReg
cp .env.example .env
# Edit .env with your database credentials
```

### 2. Create database

```sql
CREATE DATABASE fpreg;
```

### 3. Install dependencies

```bash
go mod tidy
```

### 4. Run the server

```bash
go run cmd/server/main.go
```

The server starts on `http://localhost:8080`. On first run it will:
- Auto-migrate all database tables
- Seed option sets from the register
- Create a demo facility
- Create a superadmin user (credentials from `.env`)

### 5. Login

Open `http://localhost:8080` and sign in with the seed admin credentials from your `.env` file.

Default: `admin@moh.go.ug` / `ChangeMe@2026!`

## API Endpoints

### Authentication
| Method | Endpoint              | Description            |
|--------|-----------------------|------------------------|
| POST   | /api/v1/auth/login    | Login                  |
| POST   | /api/v1/auth/refresh  | Refresh tokens         |
| POST   | /api/v1/auth/logout   | Revoke refresh token   |
| GET    | /api/v1/auth/me       | Current user profile   |

### Registrations (FP Register Entries)
| Method | Endpoint                      | Description            |
|--------|-------------------------------|------------------------|
| GET    | /api/v1/registrations         | List (paginated, filtered) |
| POST   | /api/v1/registrations         | Create new entry       |
| GET    | /api/v1/registrations/:id     | Get by ID              |
| PUT    | /api/v1/registrations/:id     | Update                 |
| DELETE | /api/v1/registrations/:id     | Soft delete (admin)    |

### Option Sets
| Method | Endpoint                          | Description            |
|--------|-----------------------------------|------------------------|
| GET    | /api/v1/option-sets               | All sets grouped       |
| GET    | /api/v1/option-sets/categories    | List category names    |
| GET    | /api/v1/option-sets/:category     | Sets by category       |

### Facilities
| Method | Endpoint                  | Description            |
|--------|---------------------------|------------------------|
| GET    | /api/v1/facilities        | List all               |
| POST   | /api/v1/facilities        | Create (superadmin)    |
| GET    | /api/v1/facilities/:id    | Get by ID              |
| PUT    | /api/v1/facilities/:id    | Update (superadmin)    |
| DELETE | /api/v1/facilities/:id    | Delete (superadmin)    |

### Users
| Method | Endpoint                          | Description            |
|--------|-----------------------------------|------------------------|
| GET    | /api/v1/users                     | List users             |
| POST   | /api/v1/users                     | Create user            |
| GET    | /api/v1/users/:id                 | Get by ID              |
| PUT    | /api/v1/users/:id                 | Update user            |
| PATCH  | /api/v1/users/:id/deactivate      | Deactivate user        |

### Audit Logs
| Method | Endpoint              | Description            |
|--------|-----------------------|------------------------|
| GET    | /api/v1/audit-logs    | List (admin only)      |

## Client Number Format

Format: `{prefix}{YYMMDD}{NNN}`

- **prefix**: Facility's client code prefix (2-5 chars, e.g., `DHC`)
- **YYMMDD**: Date of visit
- **NNN**: Daily sequence starting at 001

Example: `DHC260314001` — Demo Health Centre, March 14 2026, first client.

Sequence resets per facility per day. Row-level DB locking prevents duplicates under concurrency.

## User Roles

| Role            | Access                                          |
|-----------------|-------------------------------------------------|
| superadmin      | Full access to all facilities and features       |
| facility_admin  | Manage users and registrations for own facility  |
| facility_user   | Create and view registrations for own facility   |
| reviewer        | Read-only access to own facility's data          |

## Register Fields (HMIS MCH 007)

The system captures all 27 columns from the paper register:

1. Serial Number (auto-generated)
2. Client Number (auto-generated)
3. NIN
4. Client Name & Contact
5. Physical Address
6. Sex
7. Age
8. New User
9. Revisit
10. Previous Method Used
11. HTS Code
12. FP Counseling (Individual/Couple, OM/SE/WD/MS)
13. Switching Method & Reason
14. Oral Pills (CoC/POP/ECP)
15. Condoms (Male/Female)
16. Injectables (DMPA-IM, DMPA-SC PA/SI)
17. Implants (3yr/5yr)
18. IUDs (Copper-T, Hormonal 3yr/5yr)
19. Sterilization (Tubal/Vasectomy)
20. FAM (Standard Days/LAM/Two Day)
21. Post-Pregnancy FP (Postpartum/Post-Abortion timing)
22. LARC Removal (Implant/IUD reason & timing)
23. Side Effects
24. Cancer Screening (Cervical & Breast)
25. STI Screening
26. Referral
27. Remarks

## Mobile App Readiness

The system is designed API-first. All data operations are available via JSON REST endpoints. A mobile client can:

- Authenticate and receive JWT tokens
- Fetch option sets for local caching
- Submit registrations
- Query and filter submissions
- All without needing the web UI

**Offline sync recommendation**: Cache option sets locally, queue submissions offline, sync via POST when online.

## License

Ministry of Health — Uganda. Internal use.
