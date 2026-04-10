# generate-assets.ps1
# Generates Microsoft Store icon assets from the source icon (winres/icon.png).
# Requires ImageMagick (magick) or can use PowerShell System.Drawing.
#
# Usage: .\generate-assets.ps1

$ErrorActionPreference = "Stop"
$sourceIcon = Join-Path $PSScriptRoot "..\winres\icon.png"
$assetsDir = Join-Path $PSScriptRoot "assets"

if (-not (Test-Path $sourceIcon)) {
    Write-Error "Source icon not found: $sourceIcon"
    exit 1
}

if (-not (Test-Path $assetsDir)) {
    New-Item -ItemType Directory -Path $assetsDir | Out-Null
}

# Required Store assets: name -> [width, height]
$assets = @{
    "StoreLogo"            = @(50, 50)
    "Square44x44Logo"      = @(44, 44)
    "Square71x71Logo"      = @(71, 71)
    "Square150x150Logo"    = @(150, 150)
    "Square310x310Logo"    = @(310, 310)
    "Wide310x150Logo"      = @(310, 150)
}

# Scaled variants (Microsoft Store recommends these)
$scales = @{
    "StoreLogo"         = @(100, 125, 150, 200, 400)
    "Square44x44Logo"   = @(100, 125, 150, 200, 400)
    "Square71x71Logo"   = @(100, 125, 150, 200, 400)
    "Square150x150Logo" = @(100, 125, 150, 200, 400)
}

function Resize-WithMagick {
    param($src, $dst, $w, $h)
    & magick $src -resize "${w}x${h}!" -background none -gravity center -extent "${w}x${h}" $dst
}

function Resize-WithDrawing {
    param($src, $dst, $w, $h)
    Add-Type -AssemblyName System.Drawing
    $srcImg = [System.Drawing.Image]::FromFile($src)
    $dstImg = New-Object System.Drawing.Bitmap($w, $h)
    $graphics = [System.Drawing.Graphics]::FromImage($dstImg)
    $graphics.InterpolationMode = [System.Drawing.Drawing2D.InterpolationMode]::HighQualityBicubic
    $graphics.SmoothingMode = [System.Drawing.Drawing2D.SmoothingMode]::HighQuality
    $graphics.DrawImage($srcImg, 0, 0, $w, $h)
    $dstImg.Save($dst, [System.Drawing.Imaging.ImageFormat]::Png)
    $graphics.Dispose()
    $dstImg.Dispose()
    $srcImg.Dispose()
}

# Detect resize method
$useMagick = $null -ne (Get-Command "magick" -ErrorAction SilentlyContinue)

if ($useMagick) {
    Write-Host "Using ImageMagick for resizing"
    $resizeFn = { param($s, $d, $w, $h) Resize-WithMagick $s $d $w $h }
} else {
    Write-Host "Using System.Drawing for resizing (install ImageMagick for better quality)"
    $resizeFn = { param($s, $d, $w, $h) Resize-WithDrawing $s $d $w $h }
}

# Generate base assets
foreach ($name in $assets.Keys) {
    $dims = $assets[$name]
    $w = $dims[0]; $h = $dims[1]
    $outPath = Join-Path $assetsDir "$name.png"
    Write-Host "  $name.png (${w}x${h})"
    & $resizeFn $sourceIcon $outPath $w $h
}

# Generate scaled variants
foreach ($name in $scales.Keys) {
    $baseDims = $assets[$name]
    $baseW = $baseDims[0]; $baseH = $baseDims[1]
    foreach ($scale in $scales[$name]) {
        $w = [math]::Round($baseW * $scale / 100)
        $h = [math]::Round($baseH * $scale / 100)
        $outPath = Join-Path $assetsDir "${name}.scale-${scale}.png"
        Write-Host "  ${name}.scale-${scale}.png (${w}x${h})"
        & $resizeFn $sourceIcon $outPath $w $h
    }
}

Write-Host "`nAssets generated in: $assetsDir"
Write-Host "Review the images and replace with custom designs if desired."
