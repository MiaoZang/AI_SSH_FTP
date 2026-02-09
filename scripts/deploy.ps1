<# 
.SYNOPSIS
AI SSH/FTP Proxy - One-Click Project Deployment Script

.DESCRIPTION
Deploys local directory to remote server via AI SSH/FTP Proxy.
Automatically compresses, uploads, and extracts the directory.

.PARAMETER LocalDir
Local directory to deploy

.PARAMETER RemotePath  
Remote destination path (must end with /)

.PARAMETER ServerUrl
Proxy server URL (default: http://127.0.0.1:48891)

.EXAMPLE
.\deploy.ps1 -LocalDir .\dist -RemotePath /www/wwwroot/app/

.EXAMPLE
.\deploy.ps1 -LocalDir .\src -RemotePath /home/user/project/ -ServerUrl http://myserver:48891
#>

param(
    [Parameter(Mandatory=$true)]
    [string]$LocalDir,
    
    [Parameter(Mandatory=$true)]
    [string]$RemotePath,
    
    [string]$ServerUrl = "http://127.0.0.1:48891"
)

# --- Functions ---
function Write-Success { param($msg) Write-Host "[OK] $msg" -ForegroundColor Green }
function Write-Fail { param($msg) Write-Host "[FAIL] $msg" -ForegroundColor Red }
function Write-Warn { param($msg) Write-Host "[WARN] $msg" -ForegroundColor Yellow }
function Write-Step { param($msg) Write-Host "[INFO] $msg" -ForegroundColor Cyan }

# --- Validation ---
if (-not (Test-Path $LocalDir -PathType Container)) {
    Write-Fail "Local directory not found: $LocalDir"
    exit 1
}

$LocalDir = (Resolve-Path $LocalDir).Path
$DirName = Split-Path $LocalDir -Leaf
$TempArchive = Join-Path $env:TEMP "deploy_$(Get-Date -Format 'yyyyMMddHHmmss').tar.gz"

# --- Banner ---
Write-Host ""
Write-Host "================================================================" -ForegroundColor Blue
Write-Host "   AI SSH/FTP Proxy - One-Click Deploy" -ForegroundColor Blue
Write-Host "================================================================" -ForegroundColor Blue
Write-Host ""
Write-Host "  Local:   $LocalDir"
Write-Host "  Remote:  $RemotePath"
Write-Host "  Server:  $ServerUrl"
Write-Host ""

# --- Step 1: Health Check ---
Write-Step "[1/4] Checking server..."
try {
    $health = Invoke-RestMethod -Uri "$ServerUrl/api/health" -Method GET -TimeoutSec 5
    if ($health.status -ne "ok") { throw "Server not healthy" }
    Write-Success "Server OK"
} catch {
    Write-Fail "Cannot connect to server: $ServerUrl"
    exit 1
}

# --- Step 2: Compress ---
Write-Step "[2/4] Compressing directory..."
try {
    Push-Location (Split-Path $LocalDir -Parent)
    tar -czf $TempArchive $DirName 2>$null
    Pop-Location
    $archiveSize = (Get-Item $TempArchive).Length / 1KB
    Write-Success "Compressed: $([math]::Round($archiveSize, 1)) KB"
} catch {
    Write-Fail "Compression failed: $_"
    exit 1
}

# --- Step 3: Upload & Extract ---
Write-Step "[3/4] Uploading and extracting..."
try {
    $pathB64 = [Convert]::ToBase64String([System.Text.Encoding]::UTF8.GetBytes($RemotePath))
    
    $response = curl.exe -s --max-time 300 `
        -X POST "$ServerUrl/api/file/upload" `
        -F "file=@$TempArchive" `
        -F "path=$pathB64" `
        -F "extract=true"
    
    $result = $response | ConvertFrom-Json -ErrorAction SilentlyContinue
    if ($result.success) {
        Write-Success "Upload successful"
    } else {
        throw $response
    }
} catch {
    Write-Fail "Upload failed: $_"
    Remove-Item $TempArchive -Force -ErrorAction SilentlyContinue
    exit 1
}

# --- Step 4: Cleanup ---
Remove-Item $TempArchive -Force -ErrorAction SilentlyContinue

# --- Step 5: Verify ---
Write-Step "[4/4] Verifying deployment..."
try {
    $listBody = @{ path = $pathB64 } | ConvertTo-Json
    $listResponse = Invoke-RestMethod -Uri "$ServerUrl/api/file/list" -Method POST -ContentType "application/json" -Body $listBody -TimeoutSec 10
    
    if ($listResponse.files) {
        Write-Success "Deployment complete"
        Write-Host ""
        Write-Host "================================================================" -ForegroundColor Green
        Write-Host "   DEPLOYMENT SUCCESSFUL" -ForegroundColor Green
        Write-Host "   Remote: $RemotePath$DirName/" -ForegroundColor Green
        Write-Host "================================================================" -ForegroundColor Green
    }
} catch {
    Write-Warn "Cannot verify, but upload succeeded"
}

Write-Host ""
