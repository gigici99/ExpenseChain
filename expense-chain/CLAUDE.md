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
- [x] `src/contract/validator.go` — stub Validator (approva tutto, TODO logica reale)
- [x] `src/blockchain/blockchain.go` — stub Chain (log only, TODO SHA-256 + chain)
- [x] `src/main.go` — wiring completo: DB, migrate, repo, service, handler, server :8080
- [ ] `src/contract/validator.go` — implementare logica reale: limiti €, categorie bloccate, card type
- [ ] `src/blockchain/blockchain.go` — implementare Block, SHA-256, append-only chain

## Tech Stack

- **Language**: Go 1.26.2
- **ORM**: GORM v1.31.1
- **DB**: SQLite via `github.com/glebarez/sqlite` (no CGO, no GCC)
- **Hashing**: SHA-256 (stdlib `crypto/sha256`)
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
