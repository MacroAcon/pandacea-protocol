Param(
  [string]$OutName = "pandacea-protocol-upload"
)

$src  = Resolve-Path .
$dest = Join-Path $src "..\$OutName"
if (Test-Path $dest) { rmdir /s /q $dest }

# Exclude dirs & files we don't want in the upload
$ExcludeDirs = @(
  ".git","node_modules",".venv","venv","__pycache__",".pytest_cache",
  ".mypy_cache",".ruff_cache",".idea",".vscode","dist","build",
  "coverage",".cache",".github/workflows" # keep actions? remove from upload by default
)
$ExcludeFiles = @("Thumbs.db",".DS_Store","desktop.ini","*.log","*.tmp","*.bak","*.old","*.orig")

# Copy everything else (fast, quiet)
robocopy $src $dest /E /NFL /NDL /NJH /NJS /NP /MT:8 /XD $ExcludeDirs /XF $ExcludeFiles | Out-Null

# Make a zip next to it
$zip = Join-Path (Split-Path $dest) "$OutName.zip"
if (Test-Path $zip) { Remove-Item $zip -Force }
Compress-Archive -Path "$dest\*" -DestinationPath $zip -Force

Write-Host "Upload folder: $dest"
Write-Host "Upload zip:    $zip"
