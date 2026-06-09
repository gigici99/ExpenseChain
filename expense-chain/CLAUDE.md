# Expense Chain — CLAUDE.md

## Scopo

Blockchain in Go per validazione spese aziendali tramite smart contract.
Flusso: dipendente esegue transazione → smart contract controlla policy aziendali → esito salvato on-chain.

## Struttura

```
expense-chain/
├── go.mod
├── CLAUDE.md                   # questo file
└── src/
    ├── main.go                 # entry point (da implementare)
    ├── model/                  # entità dominio
    │   ├── company.go
    │   ├── employee.go
    │   ├── card.go
    │   ├── policy.go
    │   └── transaction.go
    ├── blockchain/             # da creare — blocchi, catena, hashing
    ├── contract/               # da creare — smart contract di validazione policy
    └── storage/                # da creare — persistenza della chain (file/db)
```

## Modelli

### Company
Dati minimi azienda: `ID`, `Name`, `VatID`, `Address`, `CreatedAt`

### Employee
Dipendente linked a company e policy: `ID`, `CompanyID`, `FirstName`, `LastName`, `Email`, `PolicyID`, `CreatedAt`

### Card
Strumento di pagamento (fisica o virtuale — Google Pay, Apple Pay, ecc.):
- `Type`: `PHYSICAL | VIRTUAL | PREPAID`
- `Last4`: ultime 4 cifre (mai salvare PAN completo)
- `Provider`: "Visa", "Mastercard", "Google Pay", ecc.

### Policy
Regole che bloccano tipi di spesa per company:
- `MaxAmountPerTx` — limite singola transazione (€)
- `MaxAmountPerDay` — cap giornaliero (€)
- `MaxAmountPerMonth` — cap mensile (€)
- `BlockedCategories` — categorie vietate: `FOOD | TRAVEL | ACCOMMODATION | ENTERTAINMENT | ELECTRONICS | OTHER`
- `AllowedCardTypes` — tipi carta ammessi

### Transaction
Spesa effettuata: contiene snapshot di `Company`, `Employee`, `Card`, `Policy` al momento della transazione.
- `Status`: `PENDING → APPROVED | REJECTED`
- `RejectReason`: motivo blocco (settato da smart contract)
- `ValidatedAt`: timestamp validazione on-chain

## Flusso previsto

```
1. Employee tenta pagamento con Card
2. Transaction creata con Status=PENDING
3. Smart contract legge Policy associata a Employee
4. Validazione:
   a. Amount <= MaxAmountPerTx?
   b. Category non in BlockedCategories?
   c. CardType in AllowedCardTypes?
   d. Spesa giornaliera/mensile entro cap?
5. Se OK → Status=APPROVED, Transaction aggiunta a blocco
6. Se KO → Status=REJECTED + RejectReason, registrata comunque on-chain (audit)
```

## Da implementare (next steps)

- [x] `src/db/db.go` — connessione SQLite con log e gestione errori
- [ ] `src/db/migrate.go` — AutoMigrate di tutti i modelli
- [x] `src/repository/` — CRUD per ogni entità (Company, Employee, Card, Policy, Transaction)
- [x] `src/service/` — business logic: CompanyService, EmployeeService, CardService, PolicyService, TransactionService (orchestrazione contract + blockchain)
- [x] `src/handler/` — HTTP handlers + router (Go 1.22+ ServeMux)
- [x] `src/contract/validator.go` — Validator reale: 5 regole, funzione pura (opzione C)
- [x] `src/blockchain/ledger.go` — ledger hash-chained: Append, Verify, AddBlock (sostituisce stub blockchain.go)
- [x] `src/main.go` — wiring completo: DB, migrate, repo, service, handler, server :8080

## Smart Contract (contract/validator.go)

Funzione PURA — niente DB, niente stato. Riceve tutto come input, ritorna APPROVED/REJECTED.
Firma: `Validate(tx, policy, spentToday, spentMonth float64) ValidationResult`

5 regole (in ordine, primo fail → reject):
1. `tx.Amount > policy.MaxAmountPerTx` → reject (se MaxAmountPerTx > 0)
2. `tx.Category` in `policy.BlockedCategories` → reject
3. `policy.AllowedCardTypes` non vuoto E `tx.Card.Type` non incluso → reject
4. `spentToday + tx.Amount > policy.MaxAmountPerDay` → reject (se cap > 0)
5. `spentMonth + tx.Amount > policy.MaxAmountPerMonth` → reject (se cap > 0)

Limiti a 0 = disabilitati (nessun controllo).

**Pattern opzione C — chi fa cosa:**
- `TransactionService.Submit` calcola `spentToday`/`spentMonth` via `txRepo.SumApprovedBetween()` (range giorno/mese), poi chiama il contract
- Contract = deterministico, testabile senza DB
- `Transaction` ora ha campi piatti `EmployeeID`/`CompanyID` (gorm:index) per le query SUM, oltre agli snapshot JSON immutabili
- `SumApprovedBetween(employeeID, from, to)` somma solo transazioni APPROVED nel range

**Fix bug:** `FindByEmployeeID` usava colonna inesistente `employee__i_d` (Employee era solo JSON blob) → ora usa campo piatto `employee_id`.

## Ledger (audit trail + blockchain unificati)

Un'unica chain hash-concatenata in tabella SQLite `ledger_entries` (append-only).
Registra TUTTI i movimenti applicativi + transazioni validate:

- `model/ledger_entry.go` — LedgerEntry: Sequence, EntityType, EntityID, Action, Payload (JSON snapshot), PrevHash, Hash
- `repository/ledger_repository.go` — solo Append + letture, NO update/delete
- `blockchain/ledger.go` — core:
  - `Append(entityType, entityID, action, entity)` — serializza entità, calcola `Hash = SHA-256(Seq|EntityType|EntityID|Action|Payload|PrevHash|CreatedAt)`, concatena. Mutex per serializzare append.
  - `AddBlock(tx)` — implementa `service.BlockchainWriter`, transazioni APPROVED → action VALIDATED
  - `Verify()` — ricalcola tutti gli hash + check linkage → tamper detection
- `service/audit.go` — interfaccia `AuditLogger`, ogni service la chiama dopo Create/Update/Delete
- Transazioni REJECTED → entrano comunque nel ledger (action REJECTED) per audit

Actions: CREATE, UPDATE, DELETE, VALIDATED, REJECTED
EntityTypes: COMPANY, EMPLOYEE, CARD, POLICY, TRANSACTION

**Endpoint ledger:**
- `GET /api/ledger` — chain completa
- `GET /api/ledger/verify` — verifica integrità (demo tamper: modifica riga a mano in SQLite → verify fallisce)
- `GET /api/ledger/entity/{entity_id}` — storia di una singola entità

**Nota demo esame:** audit append fallito NON blocca l'operazione business (solo WARNING log) — scelta deliberata, discutibile: in sistema reale audit-critical si farebbe transazione atomica.

## Autenticazione (JWT + ruoli)

JWT con access token (15 min) + refresh token (7 giorni). Password hashate bcrypt.
Secret da env `JWT_SECRET` (fallback dev insicuro se assente).

- `model/user.go` — User: ID, Username, PasswordHash (json:"-"), Role, CompanyID, EmployeeID
- `repository/user_repository.go` — Create, FindByUsername, FindByID
- `service/auth_service.go` — Register (bcrypt), Login, Refresh, ParseToken. Claims include Role+CompanyID+EmployeeID+TokenType (access/refresh). HS256.
- `handler/auth_handler.go` — POST /api/auth/{register,login,refresh} (pubblici)
- `handler/middleware.go` — `AuthMiddleware.Protect(handler, roles...)`: valida Bearer token, controlla tipo=access, controlla ruolo. ADMIN passa sempre. Lista vuota = qualsiasi utente autenticato. Claims iniettati nel context (`ClaimsFrom(r)`).

**Ruoli:**
- `ADMIN` — accesso a tutto (override in `roleAllowed`)
- `COMPANY` — crea/gestisce policy, employee, card; vede ledger/blockchain
- `EMPLOYEE` — inserisce spese (POST /api/transactions, /api/payments/incoming), vede le proprie

**Matrice ruoli→endpoint** (oltre ad ADMIN sempre ammesso): vedi `handler/router.go`.
- Company CRUD: solo ADMIN (GET visibile a COMPANY)
- Employee/Card/Policy: COMPANY
- Ledger: COMPANY (read)
- POST transactions/payments: EMPLOYEE
- GET transactions: COMPANY + EMPLOYEE (le proprie)

**Flusso token:**
1. `POST /api/auth/login {username, password}` → `{access_token, refresh_token}`
2. Richieste protette: header `Authorization: Bearer <access_token>`
3. Access scaduto → `POST /api/auth/refresh {refresh_token}` → nuova coppia

**Seed admin + provisioning utenti:**
- All'avvio `authSvc.EnsureAdmin("admin", ADMIN_PASSWORD)` crea utente ADMIN se non esiste (idempotente). Password da env `ADMIN_PASSWORD`, fallback dev `admin123`.
- `POST /api/auth/register` ora è ADMIN-only (non più pubblico).
- `POST /api/companies` (ADMIN): se body include `username`+`password`, crea anche utente COMPANY collegato (company_id = nuova azienda). Orchestrazione in `CompanyHandler` (ha sia companySvc che authSvc).
- `POST /api/employees` (COMPANY): se body include `username`+`password`, crea anche utente EMPLOYEE collegato (company_id + employee_id). Orchestrazione in `EmployeeHandler`.
- Provisioning utente fallito NON fa rollback dell'entità (azienda/dipendente resta) — ritorna errore con messaggio. Scelta demo, no transazione atomica.

**Catena di onboarding completa:**
1. login `admin` / `admin123`
2. admin crea azienda + utente COMPANY (username/password nel form)
3. login come azienda → crea dipendenti + utenti EMPLOYEE
4. login come dipendente → inserisce spese

**TODO sicurezza (non fatto, valutare per esame):**
- Scoping per CompanyID: ora COMPANY può vedere/modificare dati di QUALSIASI company, non solo la propria. Claims hanno CompanyID ma gli handler non lo usano ancora per filtrare.
- EMPLOYEE può leggere transazioni di altri employee (no ownership check su employee_id).
- Refresh token non revocabile (no blacklist/rotation) — refresh valido fino a scadenza naturale.

## Frontend — expense-chain-fe

Path: `C:\Users\LuigiCirillo\dev\unisa-sd\sicurezza-dei-dati\pj2\ExpenseChain\expense-chain-fe`
Stack: Vue 3 (Composition API, `<script setup>`) + Vite 6 + vue-router 4. CSS plain (no Tailwind).
Run: `npm run dev` (porta 5173, proxy `/api` → localhost:8080). Build: `npm run build`.

- `src/api/client.js` — fetch wrapper: Bearer header, auto-refresh su 401 (1 retry), tutti gli endpoint. Costanti CATEGORIES/CARD_TYPES/ROLES. `decodeToken` per gating UI.
- `src/store/auth.js` — stato reattivo (composable `useAuth`): login/logout, `can(...roles)` (ADMIN sempre true), claims da JWT.
- `src/router.js` — guard: redirect a /login se non autenticato. Token in localStorage.
- `src/App.vue` — shell con sidebar, link filtrati per ruolo via `can()`.
- Views: `Login`, `Dashboard` (stats + integrità ledger), `Companies`, `Employees`, `Cards`, `Policies`, `Transactions` (manuale + payment simulato + esito validazione), `Ledger` (lista entry + verifica integrità).

Sidebar mostra link in base al ruolo: COMPANY vede aziende/dipendenti/carte/policy/ledger, EMPLOYEE vede solo transazioni. Gating UI = solo UX, sicurezza vera è nel BE.

## Tech Stack

- **Language**: Go 1.26.2
- **ORM**: GORM v1.31.1
- **DB**: SQLite via `github.com/glebarez/sqlite` (no CGO, no GCC)
- **Hashing**: SHA-256 (ledger), bcrypt (password)
- **Auth**: JWT `github.com/golang-jwt/jwt/v5` (HS256), access+refresh
- **Serialization**: `encoding/json`

## DB — note

`src/db/db.go` espone:
- `Connect(dsn string)` — apre DB, configura logger GORM, max 1 connessione (SQLite single-writer)
- `MustConnect(dsn string)` — chiama Connect, `log.Fatal` se fallisce — usa all'avvio
- `Close(db)` — chiude pool connessioni
- `var DB *gorm.DB` — istanza globale (opzionale, puoi passarla in giro)

Logger: stampa query lente (>200ms) su stdout, prefisso `[DB]`. Errori `not found` silenziati (GORM li tratta come errori ma sono casi normali).

## Note sicurezza

- Mai salvare PAN completo — solo `Last4`
- Policy snapshot in Transaction: immutabile on-chain, no riferimento live modificabile
- Ogni blocco referenzia hash precedente → tamper-evident
