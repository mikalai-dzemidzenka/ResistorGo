# Builds the self-contained release binary.
#
# All textures, models, the menu track and the default level live inside the exe
# (see assets.go), so this is the only file that needs shipping. settings.json is
# still written next to it at runtime - it is per-machine, not an asset.
#
#   -trimpath   drops absolute build paths
#   -s -w       drops the symbol table and DWARF, roughly halving the binary
#   -tags release  leaves the developer tools out: no level editor, debug overlay,
#                  network-simulation keys or level writes (devtools_release.go)
#
# UPX packs it further (~1.7x, 9.0 MB -> 5.2 MB) and is required: resistor.exe is
# tracked in the repo, so an unpacked build silently doubles what every clone
# carries. A missing upx fails the build rather than quietly shipping 9 MB.
# Install with: scoop install upx
#
# If a packed binary ever trips a heuristic antivirus scanner, pass -NoPack to
# get the unpacked one deliberately, rather than by accident.

param([string]$Out = "resistor.exe", [switch]$NoPack)

$ErrorActionPreference = "Stop"

go build -trimpath -tags release -ldflags "-s -w" -o $Out ./cmd
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

$size = (Get-Item $Out).Length
Write-Host ("built {0}: {1:N1} MB" -f $Out, ($size / 1MB))

if ($NoPack) {
    Write-Host "-NoPack given: leaving $Out unpacked"
    exit 0
}

$upx = Get-Command upx -ErrorAction SilentlyContinue
if (-not $upx) {
    Write-Error "upx not found. Install it (scoop install upx), or pass -NoPack to build without it."
    exit 1
}
& $upx.Source --best --lzma -q $Out
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
Write-Host ("packed {0}: {1:N1} MB" -f $Out, ((Get-Item $Out).Length / 1MB))
