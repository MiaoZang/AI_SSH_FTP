<#
.SYNOPSIS
    AI SSH/FTP Proxy - SSH Command Execution Helper for PowerShell
.DESCRIPTION
    Wraps the SSH exec API call, handling Base64 encoding/decoding and JSON construction.
    Solves PowerShell JSON escaping issues when calling curl.exe directly.
.EXAMPLE
    .\ssh_exec.ps1 -Command "ls -la /" -Server "http://23.159.8.47:48891"
    .\ssh_exec.ps1 -Command "pm2 restart all" -Server "http://SERVER:48891"
    .\ssh_exec.ps1 -Command "cat /etc/os-release" -Server "http://SERVER:48891" -Raw
#>

param(
    [Parameter(Mandatory=$true)]
    [string]$Command,

    [Parameter(Mandatory=$true)]
    [string]$Server,

    [switch]$Raw,         # Output raw (not decoded from Base64)
    [int]$Timeout = 30    # Timeout in seconds
)

# Encode command to Base64
$cmdBase64 = [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes($Command))

Write-Host "[*] Executing: $Command" -ForegroundColor Cyan
Write-Host "[*] Server: $Server" -ForegroundColor Cyan

# Method 1: Try GET API (simplest, avoids JSON entirely)
$url = "$Server/api/ssh/exec?cmd=$cmdBase64"

try {
    $response = Invoke-RestMethod -Uri $url -Method GET -TimeoutSec $Timeout -ErrorAction Stop

    if ($response.error -and $response.error -ne "") {
        $errMsg = [Text.Encoding]::UTF8.GetString([Convert]::FromBase64String($response.error))
        Write-Host "[!] Error: $errMsg" -ForegroundColor Red
    }

    if ($response.stdout -and $response.stdout -ne "") {
        if ($Raw) {
            Write-Output $response.stdout
        } else {
            $output = [Text.Encoding]::UTF8.GetString([Convert]::FromBase64String($response.stdout))
            Write-Host ""
            Write-Host "--- stdout ---" -ForegroundColor Green
            Write-Output $output
        }
    }

    if ($response.stderr -and $response.stderr -ne "") {
        if ($Raw) {
            Write-Output $response.stderr
        } else {
            $errOutput = [Text.Encoding]::UTF8.GetString([Convert]::FromBase64String($response.stderr))
            Write-Host ""
            Write-Host "--- stderr ---" -ForegroundColor Yellow
            Write-Output $errOutput
        }
    }

    Write-Host ""
    Write-Host "[*] Exit code: $($response.exit_code)" -ForegroundColor $(if ($response.exit_code -eq 0) { "Green" } else { "Red" })

} catch {
    # Fallback: Try POST with JSON file method
    Write-Host "[!] GET method failed, trying POST with temp file..." -ForegroundColor Yellow

    $body = @{ command = $cmdBase64 } | ConvertTo-Json -Compress
    $tmpFile = [System.IO.Path]::GetTempFileName()
    [System.IO.File]::WriteAllText($tmpFile, $body, [System.Text.Encoding]::UTF8)

    try {
        $result = curl.exe -s -X POST "$Server/api/ssh/exec" -H "Content-Type: application/json" -d "@$tmpFile" 2>&1
        $response = $result | ConvertFrom-Json

        if ($response.stdout -and $response.stdout -ne "") {
            $output = [Text.Encoding]::UTF8.GetString([Convert]::FromBase64String($response.stdout))
            Write-Host ""
            Write-Host "--- stdout ---" -ForegroundColor Green
            Write-Output $output
        }

        if ($response.stderr -and $response.stderr -ne "") {
            $errOutput = [Text.Encoding]::UTF8.GetString([Convert]::FromBase64String($response.stderr))
            Write-Host ""
            Write-Host "--- stderr ---" -ForegroundColor Yellow
            Write-Output $errOutput
        }

        Write-Host ""
        Write-Host "[*] Exit code: $($response.exit_code)" -ForegroundColor $(if ($response.exit_code -eq 0) { "Green" } else { "Red" })
    } catch {
        Write-Host "[!] Failed: $_" -ForegroundColor Red
    } finally {
        Remove-Item $tmpFile -Force -ErrorAction SilentlyContinue
    }
}
