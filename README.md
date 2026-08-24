# ExpenseChain 📒

[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.26.2-success.svg)](https://golang.org/dl/)
[![Node](https://img.shields.io/badge/Node-22.0.0-success.svg)](https://nodejs.org/)

**ExpenseChain** is a university project that demonstrates a blockchain‑backed audit trail for expense‑management operations. It includes a Go backend and a Vue 3 front‑end.

---

## Architecture

```
┌──────────────┐     REST/JSON      ┌──────────────────────────────────────┐
│   Frontend   │ ◄────────────────► │           Go Backend (:8080)         │
│  Vue 3 SPA   │                    │                                      │
│  Vite :5173  │                    │  ┌─────────┐  ┌──────────────────┐  │
└──────────────┘                    │  │ Handler  │→ │     Service      │  │
                                     │  │ (REST)   │  │  (business logic)│  │
                                     │  └─────────┘  └────────┬─────────┘  │
                                     │                        │             │
                                     │              ┌─────────▼──────────┐  │
                                     │              │    Repository      │  │
                                     │              │  (GORM → SQLite)   │  │
                                     │              └────────────────────┘  │
                                     │                                      │
                                     │  ┌─────────────────┐ ┌───────────┐  │
                                     │  │  Smart Contract  │ │ Blockchain│  │
                                     │  │  (Validator)     │ │ (Ledger)  │  │
                                     │  └─────────────────┘ └───────────┘  │
                                     └──────────────────────────────────────┘
```

## Technology Stack

| Component | Technology |
|-----------|------------|
| Backend   | Go 1.26, net/http (ServeMux) |
| Database  | SQLite via GORM |
| Authentication | JWT (access 15 min + refresh 7 d), bcrypt |
| Frontend  | Vue 3 (Composition API), Vite |
| Blockchain | SHA‑256 hash‑chain, append‑only ledger |
| Smart Contract | Pure deterministic validator (5 rules) |

## Key Concepts

### Simulated Blockchain
Instead of a full Hyperledger Fabric deployment, the system implements a **hash‑chain append‑only ledger** that provides the same integrity guarantees:
- Every event (CRUD on entities + transaction validation) creates a `LedgerEntry`.
- Each entry stores `prev_hash` (hash of the previous entry) and `hash` (SHA‑256 over all fields + `prev_hash`).
- The first entry uses `"GENESIS"` as `prev_hash`.
- The endpoint `GET /api/ledger/verify` recomputes every hash and validates the chain – any tampering is detected.

**Properties guaranteed:**
- **Immutability**: altering an entry invalidates all subsequent hashes.
- **Append‑only**: no UPDATE/DELETE on the ledger table.
- **Full audit**: every operation is captured with a JSON snapshot of the entity.

### Smart Contract
`Validator` (`contract/validator.go`) is a **pure function** – it receives all inputs and deterministically returns `APPROVED` or `REJECTED`. No database access, no side effects.

**5 enforced rules:**
1. Transaction amount must not exceed `MaxAmountPerTx` (if the limit is non‑zero).
2. Transaction category must not belong to `BlockedCategories`.
3. Card type must be in `AllowedCardTypes` (if defined).
4. Daily spending (`spentToday + amount`) must stay under `MaxAmountPerDay`.
5. Monthly spending (`spentMonth + amount`) must stay under `MaxAmountPerMonth`.

Limits set to `0` are disabled.

### Data Isolation (RBAC)
Three roles with strict data isolation:
| Role | Visibility |
|------|------------|
| **ADMIN** | Full system – can manage companies and users |
| **COMPANY** | Only its own company – employees, cards, policies, transactions |
| **EMPLOYEE** | Only personal data – own cards and transactions |

Isolation is enforced **server‑side** in handlers: each request extracts JWT claims and filters results by `company_id` or `employee_id`. A COMPANY user can never see data belonging to another company.

## Data Model

```
Company 1──N Employee 1──N Card
    │              │
    │              └──── PolicyID → Policy
    │
    └── 1──N Policy

Employee ──submit──► Transaction ──validate──► Smart Contract
                           │                        │
                           │                   APPROVED/REJECTED
                           ▼
                     LedgerEntry (hash‑chained)
```

**Transaction** stores immutable snapshots of Company, Employee, Card and Policy at the moment of creation – guaranteeing that audit data never changes retroactively.

## Setup & Run

### Prerequisites
- Go 1.22+ installed
- Node 18+ installed

### Backend
```bash
cd expense-chain
go run ./src/
```
The server starts on `http://localhost:8080`. On first launch it:
- Creates the SQLite database `expense.db`
- Auto‑migrates tables via GORM
- Seeds an admin user (`admin` / `admin123`)

**Optional environment variables:**
- `JWT_SECRET` – secret for signing JWTs (default: insecure dev fallback)
- `ADMIN_PASSWORD` – admin seed password (default: `admin123`)

### Frontend
```bash
cd expense-chain-fe
npm install   # or npm ci
npm run dev   # dev server at http://localhost:5173 (proxy to backend)
```
The Vite dev server proxies `/api` calls to the backend automatically.

### Manual Test Flow
1. Log in as `admin` / `admin123`.
2. Create a company (which also creates a COMPANY user).
3. Log in as the COMPANY user → create a spending policy.
4. Create an employee (which also creates an EMPLOYEE user).
5. Create a card for the employee.
6. Log in as the EMPLOYEE → submit an expense.
7. The smart contract validates the expense → transaction is `APPROVED` or `REJECTED`.
8. Verify the ledger via the Dashboard or the Ledger page.

## API Endpoints

### Authentication
| Method | Path | Roles | Description |
|--------|------|-------|-------------|
| POST   | `/api/auth/login`    | Public | Returns access + refresh JWT |
| POST   | `/api/auth/refresh`  | Public | Refreshes access token |
| POST   | `/api/auth/register` | ADMIN  | Create a new user |

### Companies
| Method | Path | Roles | Description |
|--------|------|-------|-------------|
| POST   | `/api/companies` | ADMIN | Create a company (+ optional COMPANY user) |
| GET    | `/api/companies` | ADMIN, COMPANY | List (COMPANY sees only its own) |
| GET    | `/api/companies/{id}` | ADMIN, COMPANY | Detail (COMPANY only its own) |
| PUT    | `/api/companies/{id}` | ADMIN | Update |
| DELETE | `/api/companies/{id}` | ADMIN | Delete |

### Employees
| Method | Path | Roles | Description |
|--------|------|-------|-------------|
| POST   | `/api/employees` | COMPANY | Create (+ optional EMPLOYEE user) |
| GET    | `/api/employees` | COMPANY | List (filtered by company_id) |
| GET    | `/api/employees/{id}` | COMPANY, EMPLOYEE | Detail (ownership check) |
| PUT    | `/api/employees/{id}` | COMPANY | Update (ownership check) |
| DELETE | `/api/employees/{id}` | COMPANY | Delete (ownership check) |

### Cards
| Method | Path | Roles | Description |
|--------|------|-------|-------------|
| POST   | `/api/cards` | COMPANY | Create card for employee |
| GET    | `/api/cards/{id}` | COMPANY, EMPLOYEE | Detail (ownership check) |
| GET    | `/api/cards/employee/{employee_id}` | COMPANY, EMPLOYEE | Cards of an employee |
| DELETE | `/api/cards/{id}` | COMPANY | Delete (ownership check) |

### Policies
| Method | Path | Roles | Description |
|--------|------|-------|-------------|
| POST   | `/api/policies` | COMPANY | Create policy (restricted to own company) |
| GET    | `/api/policies` | COMPANY | List (filtered by company_id) |
| GET    | `/api/policies/{id}` | COMPANY | Detail (ownership check) |
| PUT    | `/api/policies/{id}` | COMPANY | Update (ownership check) |
| DELETE | `/api/policies/{id}` | COMPANY | Delete (ownership check) |

### Transactions
| Method | Path | Roles | Description |
|--------|------|-------|-------------|
| POST   | `/api/transactions` | EMPLOYEE | Submit expense (employee_id taken from JWT) |
| GET    | `/api/transactions` | ADMIN, COMPANY, EMPLOYEE | List (role‑based filter) |
| GET    | `/api/transactions/{id}` | COMPANY, EMPLOYEE | Detail (ownership check) |
| POST   | `/api/payments/incoming` | EMPLOYEE | Simulated card payment |

### Ledger
| Method | Path | Roles | Description |
|--------|------|-------|-------------|
| GET    | `/api/ledger` | ADMIN, COMPANY | Full ledger |
| GET    | `/api/ledger/verify` | ADMIN, COMPANY | Verify chain integrity |
| GET    | `/api/ledger/entity/{entity_id}` | ADMIN, COMPANY | History of a single entity |

## Security
- **Passwords** are bcrypt‑hashed.
- **JWT** uses HMAC‑SHA256; access token expires in 15 min, refresh token in 7 days.
- **RBAC** middleware checks role on every route; ADMIN bypasses all checks.
- **Data isolation**: handlers filter results by `company_id`/`employee_id` from JWT.
- **Cards** never store full PAN; only the last 4 digits are kept.
- **Audit**: every operation is recorded in an immutable hash‑chained ledger.
- **Smart contract** is pure, deterministic, and fully testable.

## Project Structure

```
ExpenseChain/
├── expense-chain/              # Go backend
│   ├── src/
│   │   ├── main.go             # Entry point, wiring
│   │   ├── blockchain/
│   │   │   └── ledger.go       # Hash‑chain append‑only ledger
│   │   ├── contract/
│   │   │   └── validator.go    # Smart contract (5 rules)
│   │   ├── db/
│   │   │   └── db.go           # SQLite connection via GORM
│   │   ├── handler/            # REST handlers + middleware + router
│   │   ├── model/              # Domain entities (Company, Employee, Card, Policy, Transaction, LedgerEntry, User)
│   │   ├── repository/         # Data‑access layer
│   │   └── service/            # Business logic
│   ├── go.mod
│   └── expense.db              # SQLite DB (generated)
│
└── expense-chain-fe/           # Vue 3 frontend
    ├── src/
    │   ├── api/client.js       # HTTP client with auto‑refresh JWT
    │   ├── store/auth.js       # Reactive auth state
    │   ├── views/              # Login, Dashboard, Companies, Employees, Cards, Policies, Transactions, Ledger
    │   ├── App.vue             # Shell with role‑based sidebar
    │   └── router.js           # Router with auth guard
    ├── package.json
    └── vite.config.js          # Dev proxy to backend :8080
```

---

Feel free to open an issue, star the repository, or submit a pull request if you have ideas for improvement!
 📒

[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.26.2-success.svg)](https://golang.org/dl/)
[![Node](https://img.shields.io/badge/Node-22.0.0-success.svg)](https://nodejs.org/)

**ExpenseChain** is a university project that demonstrates a blockchain‑backed audit trail for expense‑management operations. It includes a Go backend and a Vue 3 front‑end.


Sistema di gestione spese aziendali con **blockchain simulata** e **smart contract** per la validazione automatica delle transazioni.

Progetto per l'esame di **Sicurezza dei Dati** — Università degli Studi di Salerno.

---

## Architettura

```
┌──────────────┐     REST/JSON      ┌──────────────────────────────────────┐
│   Frontend   │ ◄────────────────► │           Go Backend (:8080)         │
│  Vue 3 SPA   │                    │                                      │
│  Vite :5173  │                    │  ┌─────────┐  ┌──────────────────┐  │
└──────────────┘                    │  │ Handler  │→ │     Service      │  │
                                    │  │ (REST)   │  │  (business logic)│  │
                                    │  └─────────┘  └────────┬─────────┘  │
                                    │                        │             │
                                    │              ┌─────────▼──────────┐  │
                                    │              │    Repository      │  │
                                    │              │  (GORM → SQLite)   │  │
                                    │              └────────────────────┘  │
                                    │                                      │
                                    │  ┌─────────────────┐ ┌───────────┐  │
                                    │  │  Smart Contract  │ │ Blockchain│  │
                                    │  │  (Validator)     │ │ (Ledger)  │  │
                                    │  └─────────────────┘ └───────────┘  │
                                    └──────────────────────────────────────┘
```

## Stack tecnologico

| Componente | Tecnologia |
|------------|-----------|
| Backend | Go 1.26, net/http (ServeMux Go 1.22+) |
| Database | SQLite via GORM |
| Autenticazione | JWT (access 15min + refresh 7d), bcrypt |
| Frontend | Vue 3 (Composition API), Vite |
| Blockchain | Hash-chain SHA-256, append-only ledger |
| Smart Contract | Validatore puro deterministico (5 regole) |

## Concetti chiave

### Blockchain simulata

Anziché usare Hyperledger Fabric (complessità operativa elevata), il sistema implementa una **hash-chain append-only** che fornisce le stesse garanzie di integrità:

- Ogni evento (CRUD su entità + validazione transazioni) genera un `LedgerEntry`
- Ogni entry contiene `prev_hash` (hash dell'entry precedente) e `hash` (SHA-256 su tutti i campi + prev_hash)
- La prima entry usa `"GENESIS"` come prev_hash
- L'endpoint `GET /api/ledger/verify` ricalcola ogni hash e verifica la catena — rileva qualsiasi manomissione

**Proprietà garantite:**
- **Immutabilità**: modificare un'entry invalida tutti gli hash successivi
- **Append-only**: nessuna operazione di UPDATE/DELETE sulla tabella ledger
- **Audit completo**: ogni operazione è tracciata con snapshot JSON dell'entità

### Smart Contract

Il `Validator` (`contract/validator.go`) è una **funzione pura** — riceve tutti gli input necessari e restituisce deterministicamente APPROVED o REJECTED. Zero accessi al database, zero side effects.

**5 regole applicate:**
1. **Limite per transazione** — importo massimo per singola operazione
2. **Categorie bloccate** — es. ENTERTAINMENT non consentito
3. **Whitelist tipo carta** — es. solo carte fisiche, no virtuali
4. **Cap giornaliero** — spesa massima giornaliera
5. **Cap mensile** — spesa massima mensile

### Data Isolation (RBAC)

Tre ruoli con isolamento dati rigoroso:

| Ruolo | Visibilità |
|-------|-----------|
| **ADMIN** | Tutto il sistema — gestisce aziende e utenti |
| **COMPANY** | Solo propria azienda — dipendenti, carte, policy, transazioni della propria company |
| **EMPLOYEE** | Solo propri dati — carte personali, transazioni personali |

L'isolamento è enforced **lato server** nei handler: ogni richiesta estrae i claims JWT e filtra i risultati per `company_id` o `employee_id`. Un utente COMPANY non può mai vedere dati di un'altra azienda.

## Modello dati

```
Company 1──N Employee 1──N Card
    │              │
    │              └──── PolicyID → Policy
    │
    └── 1──N Policy

Employee ──submit──► Transaction ──validate──► Smart Contract
                          │                        │
                          │                   APPROVED/REJECTED
                          ▼
                    LedgerEntry (hash-chained)
```

**Transaction** contiene snapshot immutabili di Company, Employee, Card e Policy al momento della creazione — garantisce che i dati di audit non cambino retroattivamente.

## Setup e avvio

### Prerequisiti
- Go 1.22+ installato
- Node.js 18+ installato

### Backend

```bash
cd expense-chain
go run ./src/
```

Il server parte su `http://localhost:8080`. Al primo avvio:
- Crea il database SQLite `expense.db`
- Crea le tabelle via AutoMigrate
- Crea l'utente admin seed (`admin` / `admin123`)

**Variabili d'ambiente opzionali:**
- `JWT_SECRET` — segreto per firma JWT (default: dev fallback insicuro)
- `ADMIN_PASSWORD` — password admin seed (default: `admin123`)

### Frontend

```bash
cd expense-chain-fe
npm install
npm run dev
```

Il dev server Vite parte su `http://localhost:5173` con proxy automatico verso il backend.

### Flusso di test manuale

1. Login come `admin` / `admin123`
2. Creare un'azienda (con username/password per l'utente COMPANY)
3. Login come COMPANY → creare una policy di spesa
4. Creare un dipendente (con username/password per l'utente EMPLOYEE)
5. Creare una carta per il dipendente
6. Login come EMPLOYEE → inviare una spesa
7. Lo smart contract valida → transazione APPROVED o REJECTED
8. Verificare il ledger dalla dashboard o dalla pagina Ledger

## API Endpoints

### Autenticazione
| Metodo | Path | Ruoli | Descrizione |
|--------|------|-------|-------------|
| POST | `/api/auth/login` | Pubblico | Login, ritorna access + refresh token |
| POST | `/api/auth/refresh` | Pubblico | Rinnova token con refresh token |
| POST | `/api/auth/register` | ADMIN | Crea nuovo utente |

### Aziende
| Metodo | Path | Ruoli | Descrizione |
|--------|------|-------|-------------|
| POST | `/api/companies` | ADMIN | Crea azienda (+ utente COMPANY opzionale) |
| GET | `/api/companies` | ADMIN, COMPANY | Lista (COMPANY vede solo la propria) |
| GET | `/api/companies/{id}` | ADMIN, COMPANY | Dettaglio (COMPANY solo la propria) |
| PUT | `/api/companies/{id}` | ADMIN | Modifica |
| DELETE | `/api/companies/{id}` | ADMIN | Elimina |

### Dipendenti
| Metodo | Path | Ruoli | Descrizione |
|--------|------|-------|-------------|
| POST | `/api/employees` | COMPANY | Crea (+ utente EMPLOYEE opzionale) |
| GET | `/api/employees` | COMPANY | Lista (filtrata per company_id) |
| GET | `/api/employees/{id}` | COMPANY, EMPLOYEE | Dettaglio (con ownership check) |
| PUT | `/api/employees/{id}` | COMPANY | Modifica (con ownership check) |
| DELETE | `/api/employees/{id}` | COMPANY | Elimina (con ownership check) |

### Carte
| Metodo | Path | Ruoli | Descrizione |
|--------|------|-------|-------------|
| POST | `/api/cards` | COMPANY | Crea carta per dipendente |
| GET | `/api/cards/{id}` | COMPANY, EMPLOYEE | Dettaglio (con ownership check) |
| GET | `/api/cards/employee/{employee_id}` | COMPANY, EMPLOYEE | Carte di un dipendente |
| DELETE | `/api/cards/{id}` | COMPANY | Elimina (con ownership check) |

### Policy
| Metodo | Path | Ruoli | Descrizione |
|--------|------|-------|-------------|
| POST | `/api/policies` | COMPANY | Crea policy (forzata su propria company) |
| GET | `/api/policies` | COMPANY | Lista (filtrata per company_id) |
| GET | `/api/policies/{id}` | COMPANY | Dettaglio (con ownership check) |
| PUT | `/api/policies/{id}` | COMPANY | Modifica (con ownership check) |
| DELETE | `/api/policies/{id}` | COMPANY | Elimina (con ownership check) |

### Transazioni
| Metodo | Path | Ruoli | Descrizione |
|--------|------|-------|-------------|
| POST | `/api/transactions` | EMPLOYEE | Submit spesa (employee_id da JWT) |
| GET | `/api/transactions` | ADMIN, COMPANY, EMPLOYEE | Lista (filtrata per ruolo) |
| GET | `/api/transactions/{id}` | COMPANY, EMPLOYEE | Dettaglio (con ownership check) |
| POST | `/api/payments/incoming` | EMPLOYEE | Pagamento carta simulato |

### Ledger
| Metodo | Path | Ruoli | Descrizione |
|--------|------|-------|-------------|
| GET | `/api/ledger` | ADMIN, COMPANY | Ledger completo |
| GET | `/api/ledger/verify` | ADMIN, COMPANY | Verifica integrità catena |
| GET | `/api/ledger/entity/{entity_id}` | ADMIN, COMPANY | Storico di un'entità |

## Sicurezza

- **Password**: hash bcrypt (cost default)
- **JWT**: HMAC-SHA256, access token 15 min, refresh token 7 giorni
- **RBAC**: middleware enforced su ogni route, ADMIN bypassa sempre
- **Data isolation**: handler filtrano per company_id/employee_id dal JWT
- **Carte**: mai PAN completo, solo last 4 digits
- **Audit**: ogni operazione registrata in ledger immutabile hash-chained
- **Smart contract puro**: nessun side effect, deterministico, testabile

## Struttura progetto

```
ExpenseChain/
├── expense-chain/              # Backend Go
│   ├── src/
│   │   ├── main.go             # Entry point, wiring
│   │   ├── blockchain/
│   │   │   └── ledger.go       # Hash-chain append-only ledger
│   │   ├── contract/
│   │   │   └── validator.go    # Smart contract (5 regole)
│   │   ├── db/
│   │   │   └── db.go           # SQLite connection via GORM
│   │   ├── handler/            # REST handlers + middleware + router
│   │   ├── model/              # Entità (Company, Employee, Card, Policy, Transaction, LedgerEntry, User)
│   │   ├── repository/         # Data access layer
│   │   └── service/            # Business logic
│   ├── go.mod
│   └── expense.db              # SQLite database (generato)
│
└── expense-chain-fe/           # Frontend Vue 3
    ├── src/
    │   ├── api/client.js       # HTTP client con auto-refresh JWT
    │   ├── store/auth.js       # Stato autenticazione reattivo
    │   ├── views/              # Login, Dashboard, Companies, Employees, Cards, Policies, Transactions, Ledger
    │   ├── App.vue             # Shell con sidebar RBAC-condizionale
    │   └── router.js           # Route con guard autenticazione
    ├── package.json
    └── vite.config.js          # Proxy dev verso backend :8080
```
