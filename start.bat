@echo off
echo ========================================
echo   ExpenseChain - Avvio completo
echo ========================================
echo.

echo [1/2] Avvio backend Go :8080 ...
start "ExpenseChain Backend" cmd /k "cd /d %~dp0expense-chain && go run ./src/"

echo Attendo 3 secondi per il backend...
timeout /t 3 /nobreak >nul

echo [2/2] Avvio frontend Vue :5173 ...
start "ExpenseChain Frontend" cmd /k "cd /d %~dp0expense-chain-fe && npm run dev"

echo.
echo ========================================
echo   Backend:  http://localhost:8080
echo   Frontend: http://localhost:5173
echo ========================================
echo.
echo Chiudi le finestre CMD per stoppare.
