Param([string]$Path = ".")
$full = Resolve-Path $Path
$files = Get-ChildItem -Recurse -Force $full -File | Measure-Object
Write-Host "$($files.Count) files in $full"
