# ============================================================
# AI SSH/FTP Proxy - 一键项目部署脚本 (PowerShell)
# One-Click Project Deployment Script
# ============================================================
# 用法 / Usage:
#   .\deploy.ps1 -LocalDir .\dist -RemotePath /www/wwwroot/app/
#   .\deploy.ps1 -LocalDir .\src -RemotePath /home/user/project/ -ServerUrl http://myserver:48891
# ============================================================

param(
    [Parameter(Mandatory=$true)]
    [string]$LocalDir,
    
    [Parameter(Mandatory=$true)]
    [string]$RemotePath,
    
    [string]$ServerUrl = "http://127.0.0.1:48891"
)

# --- Functions ---
function Write-Success { param($msg) Write-Host "✓ $msg" -ForegroundColor Green }
function Write-Error2 { param($msg) Write-Host "✗ $msg" -ForegroundColor Red }
function Write-Warning2 { param($msg) Write-Host "⚠ $msg" -ForegroundColor Yellow }
function Write-Info { param($msg) Write-Host "ℹ $msg" -ForegroundColor Cyan }

# --- Validation ---
if (-not (Test-Path $LocalDir -PathType Container)) {
    Write-Error2 "本地目录不存在: $LocalDir"
    exit 1
}

$LocalDir = (Resolve-Path $LocalDir).Path
$DirName = Split-Path $LocalDir -Leaf
$TempArchive = Join-Path $env:TEMP "deploy_$(Get-Date -Format 'yyyyMMddHHmmss').tar.gz"

# --- Banner ---
Write-Host ""
Write-Host "╔════════════════════════════════════════════════════════════╗" -ForegroundColor Blue
Write-Host "║          AI SSH/FTP Proxy - 一键部署 / One-Click Deploy    ║" -ForegroundColor Blue
Write-Host "╚════════════════════════════════════════════════════════════╝" -ForegroundColor Blue
Write-Host ""
Write-Host "  本地目录 / Local:    $LocalDir"
Write-Host "  远程路径 / Remote:   $RemotePath"
Write-Host "  服务地址 / Server:   $ServerUrl"
Write-Host ""

# --- Step 1: Health Check ---
Write-Info "[1/4] 检查服务状态 / Checking server..."
try {
    $health = Invoke-RestMethod -Uri "$ServerUrl/api/health" -Method GET -TimeoutSec 5
    if ($health.status -ne "ok") { throw "Server not healthy" }
    Write-Success "服务正常 / Server OK"
} catch {
    Write-Error2 "无法连接到服务器 / Cannot connect to server: $ServerUrl"
    exit 1
}

# --- Step 2: Compress ---
Write-Info "[2/4] 压缩目录 / Compressing..."
try {
    Push-Location (Split-Path $LocalDir -Parent)
    tar -czf $TempArchive $DirName 2>$null
    Pop-Location
    $archiveSize = (Get-Item $TempArchive).Length / 1KB
    Write-Success "压缩完成 / Compressed: $([math]::Round($archiveSize, 1)) KB"
} catch {
    Write-Error2 "压缩失败 / Compression failed: $_"
    exit 1
}

# --- Step 3: Upload & Extract ---
Write-Info "[3/4] 上传并解压 / Uploading & extracting..."
try {
    $pathB64 = [Convert]::ToBase64String([System.Text.Encoding]::UTF8.GetBytes($RemotePath))
    
    $response = curl.exe -s --max-time 300 `
        -X POST "$ServerUrl/api/file/upload" `
        -F "file=@$TempArchive" `
        -F "path=$pathB64" `
        -F "extract=true"
    
    $result = $response | ConvertFrom-Json -ErrorAction SilentlyContinue
    if ($result.success) {
        Write-Success "上传成功 / Upload successful"
    } else {
        throw $response
    }
} catch {
    Write-Error2 "上传失败 / Upload failed: $_"
    Remove-Item $TempArchive -Force -ErrorAction SilentlyContinue
    exit 1
}

# --- Step 4: Cleanup ---
Remove-Item $TempArchive -Force -ErrorAction SilentlyContinue

# --- Step 5: Verify ---
Write-Info "[4/4] 验证部署 / Verifying..."
try {
    $listBody = @{ path = $pathB64 } | ConvertTo-Json
    $listResponse = Invoke-RestMethod -Uri "$ServerUrl/api/file/list" -Method POST -ContentType "application/json" -Body $listBody -TimeoutSec 10
    
    if ($listResponse.files) {
        Write-Success "部署完成 / Deployment complete"
        Write-Host ""
        Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Green
        Write-Host "   ✓ 部署成功 / Deployment Successful" -ForegroundColor Green
        Write-Host "     远程路径 / Remote: $RemotePath$DirName/" -ForegroundColor Green
        Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Green
    }
} catch {
    Write-Warning2 "无法验证，但上传成功 / Cannot verify, but upload succeeded"
}

Write-Host ""
