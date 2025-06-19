@echo off
echo 🚀 Execution de tous les tests K6 sur Staging
echo ================================================

echo.
echo 📊 Test 1/6: Ping
k6 run -e ENVIRONMENT=staging ping.js

echo.
echo 📊 Test 2/6: Login
k6 run -e ENVIRONMENT=staging login.js

echo.
echo 📊 Test 3/6: Register
k6 run -e ENVIRONMENT=staging register.js

echo.
echo 📊 Test 4/6: Auth Flow
k6 run -e ENVIRONMENT=staging auth.js

echo.
echo 📊 Test 5/6: Feedback
k6 run -e ENVIRONMENT=staging feedback.js

echo.
echo 📊 Test 6/6: Board
k6 run -e ENVIRONMENT=staging board.js

echo.
echo 🎉 Tous les tests sont terminés !
echo 📁 Consultez le dossier results/ pour les rapports détaillés
pause
