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
#
# -Linux cross-compiles for desktop linux/amd64 instead (default output
# resistor-linux-amd64), with zig as the C compiler: raylib and go-enet are both
# cgo. tools/linux_sysroot.py fetches the X11/xkbcommon/GL files raylib links
# against from Debian and builds enet into the binary, once, into
# %LOCALAPPDATA%\resistor-build\linux-amd64. Needs zig and python besides go and
# upx (scoop install zig python). raylib is built for X11 only (-tags x11), so
# the player needs libX11, libxkbcommon and libGL - any desktop has them, and a
# Wayland desktop runs it through XWayland. glibc 2.31 or newer.

param([string]$Out = "", [switch]$NoPack, [switch]$Linux)

$ErrorActionPreference = "Stop"

$tags = "release"
$buildEnv = @{}
if ($Linux) {
    if (-not $Out) { $Out = "resistor-linux-amd64" }
    foreach ($tool in "zig", "python") {
        if (-not (Get-Command $tool -ErrorAction SilentlyContinue)) {
            Write-Error "$tool not found. Install it (scoop install $tool) to build for Linux."
            exit 1
        }
    }
    $cache = Join-Path $env:LOCALAPPDATA "resistor-build\linux-amd64"
    & python (Join-Path $PSScriptRoot "tools\linux_sysroot.py") $cache
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

    $root = ($cache -replace '\\', '/') + "/sysroot"
    $tags = "release x11"
    $buildEnv = @{
        GOOS        = "linux"
        GOARCH      = "amd64"
        CGO_ENABLED = "1"
        CC          = "zig cc -target x86_64-linux-gnu.2.31" # glibc must match linux_sysroot.py
        # raylib-go passes a GCC-only warning flag that clang does not know
        CGO_CFLAGS  = "-O2 -Wno-unknown-warning-option -isystem $root/usr/include"
        # the Debian libraries were built against a newer glibc than we target;
        # on the player's machine they come with their own, so do not hold the
        # binary to theirs
        CGO_LDFLAGS = "-L$root/usr/lib/x86_64-linux-gnu -Wl,--allow-shlib-undefined"
        PKG_CONFIG  = Join-Path $cache "pkg-config.cmd"
    }
} elseif (-not $Out) {
    $Out = "resistor.exe"
}

# Build into a fresh file, and put it in place of $Out only once it is built and
# packed. go build leaves an output it judges up to date alone, and a packed
# binary can pass for the one it would link - so building over a packed $Out
# could write nothing, and upx would then refuse the already packed file. A
# failed build or pack also leaves the previous $Out as it was.
$tmp = "$Out.tmp"
Remove-Item $tmp -ErrorAction SilentlyContinue

# The cross-compile settings are for this one go build: the script runs in the
# caller's session, and a GOOS=linux left behind would break their next build.
$saved = @{}
foreach ($k in $buildEnv.Keys) {
    $saved[$k] = [Environment]::GetEnvironmentVariable($k, "Process")
    Set-Item "env:$k" $buildEnv[$k]
}
try {
    go build -trimpath -tags $tags -ldflags "-s -w" -o $tmp ./cmd
    $code = $LASTEXITCODE
} finally {
    foreach ($k in $saved.Keys) {
        if ($null -eq $saved[$k]) {
            Remove-Item "env:$k" -ErrorAction SilentlyContinue
        } else {
            Set-Item "env:$k" $saved[$k]
        }
    }
}
if ($code -ne 0) {
    Remove-Item $tmp -ErrorAction SilentlyContinue
    exit $code
}

$size = (Get-Item $tmp).Length
Write-Host ("built {0}: {1:N1} MB" -f $Out, ($size / 1MB))

if ($NoPack) {
    Move-Item -Force $tmp $Out
    Write-Host "-NoPack given: leaving $Out unpacked"
    exit 0
}

$upx = Get-Command upx -ErrorAction SilentlyContinue
if (-not $upx) {
    Remove-Item $tmp -ErrorAction SilentlyContinue
    Write-Error "upx not found. Install it (scoop install upx), or pass -NoPack to build without it."
    exit 1
}
& $upx.Source --best --lzma -q $tmp
if ($LASTEXITCODE -ne 0) {
    $code = $LASTEXITCODE
    Remove-Item $tmp -ErrorAction SilentlyContinue
    exit $code
}
Move-Item -Force $tmp $Out
Write-Host ("packed {0}: {1:N1} MB" -f $Out, ((Get-Item $Out).Length / 1MB))
