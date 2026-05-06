#Requires -Version 5.1
<#
.SYNOPSIS
    Crea y sube un tag de release para disparar el workflow de GitHub Actions.
.EXAMPLE
    .\release.ps1
    .\release.ps1 v1.2.0
#>

param(
    [string]$Version = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Get-NextVersion {
    $latest = git tag --sort=-v:refname | Where-Object { $_ -match '^v\d+\.\d+\.\d+$' } | Select-Object -First 1
    if (-not $latest) {
        return "v0.1.0"
    }
    $parts = $latest.TrimStart('v') -split '\.'
    $patch = [int]$parts[2] + 1
    return "v$($parts[0]).$($parts[1]).$patch"
}

if (-not $Version) {
    $Version = Get-NextVersion
    Write-Host "No se especificó versión. Usando la siguiente: $Version"
}

if ($Version -notmatch '^v\d+\.\d+\.\d+$') {
    Write-Error "La versión debe tener el formato vX.Y.Z (recibido: $Version)"
    exit 1
}

$existingTag = git tag | Where-Object { $_ -eq $Version }
if ($existingTag) {
    Write-Error "El tag $Version ya existe localmente."
    exit 1
}

Write-Host "Creando y subiendo el tag $Version..."
git tag $Version
git push origin $Version

Write-Host ""
Write-Host "Workflow de release disparado para $Version"
$repoUrl = gh repo view --json url -q .url
Write-Host "Seguilo en: $repoUrl/actions"
