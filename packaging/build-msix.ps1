# build-msix.ps1
# Builds an MSIX package for Microsoft Store submission.
#
# Prerequisites:
#   - Windows 10/11 SDK (provides makeappx.exe and signtool.exe)
#   - The app must already be built: "Idu Mishmi Keyboard.exe"
#   - Store assets must be generated: .\generate-assets.ps1
#   - For local testing: a self-signed certificate (see below)
#
# Usage:
#   .\build-msix.ps1                          # Build unsigned MSIX (for Store upload)
#   .\build-msix.ps1 -Sign -CertThumbprint "ABC123..."  # Build and sign for local testing
#
# For Store submission:
#   Upload the unsigned .msix to Microsoft Partner Center.
#   Microsoft will sign it with their certificate.
#
# To create a self-signed certificate for local testing:
#   $cert = New-SelfSignedCertificate -Type Custom -Subject "CN=IduMishmi-Dev" `
#     -KeyUsage DigitalSignature -FriendlyName "Idu Mishmi Dev" `
#     -CertStoreLocation "Cert:\CurrentUser\My" `
#     -TextExtension @("2.5.29.37={text}1.3.6.1.5.5.7.3.3", "2.5.29.19={text}")
#   # Then update AppxManifest.xml Publisher to match: CN=IduMishmi-Dev

param(
    [switch]$Sign,
    [string]$CertThumbprint = ""
)

$ErrorActionPreference = "Stop"

$projectRoot = Split-Path $PSScriptRoot -Parent
$packagingDir = $PSScriptRoot
$exePath = Join-Path $projectRoot "Idu Mishmi Keyboard.exe"
$manifestPath = Join-Path $packagingDir "AppxManifest.xml"
$assetsDir = Join-Path $packagingDir "assets"
$stagingDir = Join-Path $packagingDir "staging"
$outputMsix = Join-Path $projectRoot "IduMishmiKeyboard.msix"

# Validate prerequisites
if (-not (Test-Path $exePath)) {
    Write-Error "App not built. Run 'make build' first. Expected: $exePath"
    exit 1
}

if (-not (Test-Path $manifestPath)) {
    Write-Error "AppxManifest.xml not found: $manifestPath"
    exit 1
}

if (-not (Test-Path (Join-Path $assetsDir "StoreLogo.png"))) {
    Write-Error "Store assets not generated. Run .\generate-assets.ps1 first."
    exit 1
}

# Check for makeappx
$makeappx = Get-Command "makeappx.exe" -ErrorAction SilentlyContinue
if (-not $makeappx) {
    # Try common SDK paths
    $sdkPaths = @(
        "${env:ProgramFiles(x86)}\Windows Kits\10\bin\*\x64\makeappx.exe"
    )
    foreach ($pattern in $sdkPaths) {
        $found = Get-Item $pattern -ErrorAction SilentlyContinue | Sort-Object -Descending | Select-Object -First 1
        if ($found) {
            $makeappx = $found.FullName
            break
        }
    }
    if (-not $makeappx) {
        Write-Error "makeappx.exe not found. Install the Windows 10/11 SDK."
        exit 1
    }
}

Write-Host "Using makeappx: $makeappx"

# Clean and create staging directory
if (Test-Path $stagingDir) {
    Remove-Item -Recurse -Force $stagingDir
}
New-Item -ItemType Directory -Path $stagingDir | Out-Null
New-Item -ItemType Directory -Path (Join-Path $stagingDir "assets") | Out-Null

# Copy files to staging
Write-Host "Staging files..."
Copy-Item $exePath $stagingDir
Copy-Item $manifestPath $stagingDir
Copy-Item (Join-Path $assetsDir "*.png") (Join-Path $stagingDir "assets")

Write-Host "Files staged in: $stagingDir"

# Build MSIX
Write-Host "Creating MSIX package..."
if ($makeappx -is [System.Management.Automation.ApplicationInfo]) {
    & $makeappx.Source pack /d $stagingDir /p $outputMsix /o
} else {
    & $makeappx pack /d $stagingDir /p $outputMsix /o
}

if ($LASTEXITCODE -ne 0) {
    Write-Error "makeappx failed with exit code $LASTEXITCODE"
    exit 1
}

Write-Host "MSIX created: $outputMsix"

# Optional signing for local testing
if ($Sign) {
    if ([string]::IsNullOrEmpty($CertThumbprint)) {
        Write-Error "Specify -CertThumbprint when using -Sign"
        exit 1
    }

    $signtool = Get-Command "signtool.exe" -ErrorAction SilentlyContinue
    if (-not $signtool) {
        $sdkPaths = @(
            "${env:ProgramFiles(x86)}\Windows Kits\10\bin\*\x64\signtool.exe"
        )
        foreach ($pattern in $sdkPaths) {
            $found = Get-Item $pattern -ErrorAction SilentlyContinue | Sort-Object -Descending | Select-Object -First 1
            if ($found) {
                $signtool = $found.FullName
                break
            }
        }
        if (-not $signtool) {
            Write-Error "signtool.exe not found. Install the Windows 10/11 SDK."
            exit 1
        }
    }

    Write-Host "Signing MSIX with certificate $CertThumbprint..."
    if ($signtool -is [System.Management.Automation.ApplicationInfo]) {
        & $signtool.Source sign /fd SHA256 /sha1 $CertThumbprint /td SHA256 $outputMsix
    } else {
        & $signtool sign /fd SHA256 /sha1 $CertThumbprint /td SHA256 $outputMsix
    }

    if ($LASTEXITCODE -ne 0) {
        Write-Error "signtool failed with exit code $LASTEXITCODE"
        exit 1
    }
    Write-Host "MSIX signed successfully."
}

# Cleanup staging
Remove-Item -Recurse -Force $stagingDir

Write-Host ""
Write-Host "Done! Next steps:"
if (-not $Sign) {
    Write-Host "  - For local testing: re-run with -Sign -CertThumbprint <thumbprint>"
}
Write-Host "  - For Store submission: upload $outputMsix to Microsoft Partner Center"
Write-Host "    https://partner.microsoft.com/dashboard/apps-and-games/overview"
