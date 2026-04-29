@echo off

REM Garante que o bat roda a partir do seu proprio diretorio
cd /d "%~dp0"

REM Adiciona o diretorio de binarios do Go ao PATH desta sessao
set "PATH=%USERPROFILE%\go\bin;%PATH%"

echo ================================
echo   godocker - inicializador
echo ================================
echo.

REM --- Cria .env a partir do exemplo se ainda nao existe ---
if not exist ".env" (
    echo [1/5] Criando .env a partir de .env.example...
    copy ".env.example" ".env" >nul
    echo       .env criado. Edite com os IPs das suas maquinas Docker.
    echo.
) else (
    echo [1/5] .env ja existe - mantendo configuracao atual.
    echo.
)

REM --- Instala o swag CLI se ainda nao estiver disponivel ---
echo [2/5] Verificando swag CLI...
where swag >nul 2>&1
if errorlevel 1 (
    echo       swag nao encontrado - instalando...
    go install github.com/swaggo/swag/cmd/swag@latest
    if errorlevel 1 (
        echo ERRO: falha ao instalar swag. Verifique sua conexao.
        goto fim
    )
    echo       swag instalado.
) else (
    echo       swag OK.
)
echo.

REM --- Baixa e organiza as dependencias Go ---
echo [3/5] Verificando dependencias (go mod tidy)...
go mod tidy
if errorlevel 1 (
    echo ERRO: falha ao resolver dependencias.
    goto fim
)
echo       Dependencias OK.
echo.

REM --- Gera docs/swagger.json a partir das anotacoes nos handlers ---
echo [4/5] Gerando documentacao Swagger...
swag init -g main.go -o docs --parseDependency
if errorlevel 1 (
    echo ERRO: falha ao gerar documentacao Swagger.
    goto fim
)

REM docs.go adiciona dependencia de runtime do swag - removemos pois
REM o servidor embute os arquivos via go:embed
if exist "docs\docs.go" del "docs\docs.go"

REM swag gera .yaml - renomeia para .yml (padrao adotado neste projeto)
if exist "docs\swagger.yaml" (
    if exist "docs\swagger.yml" del "docs\swagger.yml"
    ren "docs\swagger.yaml" "swagger.yml"
)

go mod tidy >nul 2>&1

echo       Documentacao gerada: docs\swagger.json + docs\swagger.yml
echo.

REM --- Inicia o servidor ---
echo [5/5] Iniciando servidor...
echo.
echo   API:     http://localhost:8080/
echo   Swagger: http://localhost:8080/swagger
echo   Hosts:   http://localhost:8080/hosts
echo.
echo   Pressione Ctrl+C para parar.
echo ================================
echo.

go run .

:fim
echo.
pause
