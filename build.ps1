# Builds the self-contained release binary.
#
# All textures, models, the menu track and the default level live inside the exe
# (see assets.go), so this is the only file that needs shipping. settings.json is
# still written next to it at runtime - it is per-machine, not an asset.
#
#   -trimpath   drops absolute build paths
#   -s -w       drops the symbol table and DWARF, roughly halving the binary
#
# UPX, if installed, packs it further (~2.5x). It is optional because a packed
# binary occasionally trips heuristic antivirus scanners; drop it if that bites.
# Install with: scoop install upx

param([string]$Out = "resistor.exe")

$ErrorActionPreference = "Stop"

go build -trimpath -ldflags "-s -w" -o $Out ./cmd
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

$size = (Get-Item $Out).Length
Write-Host ("built {0}: {1:N1} MB" -f $Out, ($size / 1MB))

$upx = Get-Command upx -ErrorAction SilentlyContinue
if ($upx) {
    & $upx.Source --best --lzma -q $Out
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    Write-Host ("packed {0}: {1:N1} MB" -f $Out, ((Get-Item $Out).Length / 1MB))
} else {
    Write-Host "upx not found - skipping compression (scoop install upx)"
}
