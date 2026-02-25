@echo off
setlocal

echo [1/4] Creando carpeta de release...
if not exist "release" mkdir release

echo [2/4] Cerrando instancia de JDSA si esta abierta...
taskkill /f /im JDSA.exe >nul 2>&1
timeout /t 1 /nobreak >nul

echo [3/4] Compilando aplicacion con Wails...
wails build -o JDSA.exe

if %errorlevel% neq 0 (
    echo.
    echo [ERROR] La compilacion fallo.
    pause
    exit /b %errorlevel%
)

echo [4/4] Moviendo ejecutable a la carpeta de release...
if exist "release\JDSA.exe" del /f "release\JDSA.exe"
move "build\bin\JDSA.exe" "release\JDSA.exe"

echo.
echo [EXITO] Build completado: .\release\JDSA.exe
pause
