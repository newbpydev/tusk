# Native Windows entry point; canonical checks still run through GNU Make/Git Bash.
[CmdletBinding()]
param(
 [Parameter(Mandatory=$true)][string]$Binary,
 [Parameter(Mandatory=$true)][string]$Manifest,
 [Parameter(Mandatory=$true)][string]$ManifestSHA256,
 [Parameter(Mandatory=$true)][string]$OutputDirectory,
 [string]$VerificationReceipt,
 [ValidateSet('trusted-candidate','local-fixture')][string]$EvidenceScope='trusted-candidate'
)
$ErrorActionPreference='Stop'
if ($env:OS -ne 'Windows_NT' -or $env:PROCESSOR_ARCHITECTURE -ne 'AMD64' -or -not [Environment]::Is64BitProcess) { throw 'Native Windows amd64 PowerShell is required; WSL/ARM emulation cannot close this gate.' }
if (Test-Path -LiteralPath $OutputDirectory) { throw 'Choose a new retained output directory.' }
$values=@{RELEASE_BINARY=(Resolve-Path -LiteralPath $Binary).Path.Replace('\','/');RELEASE_MANIFEST=(Resolve-Path -LiteralPath $Manifest).Path.Replace('\','/');CANDIDATE_MANIFEST_SHA256=$ManifestSHA256;RELEASE_SMOKE_OUTPUT=[IO.Path]::GetFullPath($OutputDirectory).Replace('\','/');CANDIDATE_VERIFICATION_RECEIPT=$VerificationReceipt.Replace('\','/');RELEASE_EVIDENCE_SCOPE=$EvidenceScope}
$previous=@{}
try {
 foreach ($key in $values.Keys) { $previous[$key]=[Environment]::GetEnvironmentVariable($key,'Process');[Environment]::SetEnvironmentVariable($key,$values[$key],'Process') }
 Push-Location (Split-Path -Parent $PSScriptRoot)
 try { & make release-smoke; if ($LASTEXITCODE -ne 0) {throw "Canonical native smoke failed ($LASTEXITCODE); retained output: $OutputDirectory"} }
 finally { Pop-Location }
 Write-Output 'Automated runtime fixtures passed. Windows Terminal interaction, console events and visual restoration remain separate acceptance checks.'
} finally { foreach ($key in $previous.Keys) { [Environment]::SetEnvironmentVariable($key,$previous[$key],'Process') } }
