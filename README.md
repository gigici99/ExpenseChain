# ExpenseChain

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
