"""Prepares what `build.ps1 -Linux` needs to cross-compile the game for
linux/amd64 with zig, into the directory given as the only argument:

  sysroot/         the X11, xkbcommon and GL headers and libraries raylib's
                   X11 build compiles and links against, from Debian bookworm
  enet/include/    enet's headers
  enet/lib/        enet built from source as a static library: on Linux
                   go-enet expects a system libenet, and this links it in so
                   the player does not need one
  pkg-config.cmd   a stand-in for pkg-config, which go-enet asks for libenet

Packages are resolved by name from the live Debian index and checked against
its SHA-256 sums, so nothing pins a version that the mirror has since dropped.
A directory that finished preparing is left alone; bump STAMP to redo it.
"""

import hashlib
import io
import lzma
import os
import shutil
import subprocess
import sys
import tarfile
import urllib.request

STAMP = "1"
MIRROR = "https://deb.debian.org/debian"
SUITE = "bookworm"
TARGET = "x86_64-linux-gnu.2.31"  # must match build.ps1

# The X11 build of GLFW links X11, xkbcommon and GL; the rest are what those
# libraries themselves link, so the linker can follow them.
PACKAGES = [
    "libx11-dev", "libx11-6", "x11proto-dev",
    "libxcursor-dev", "libxrandr-dev", "libxinerama-dev", "libxi-dev",
    "libxext-dev", "libxrender-dev", "libxfixes-dev",
    "libxkbcommon-dev", "libxkbcommon0",
    "libgl-dev", "libglx-dev", "libgl1", "libglvnd0", "libglx0",
    "libxcb1", "libxau6", "libxdmcp6", "libbsd0", "libmd0",
]

ENET_VERSION = "1.3.15"  # what go-enet's own headers and Windows enet.lib are
ENET_URL = f"https://github.com/lsalzman/enet/archive/refs/tags/v{ENET_VERSION}.tar.gz"
ENET_SOURCES = ["callbacks", "compress", "host", "list", "packet", "peer", "protocol", "unix"]
# what enet's configure script finds on any glibc system
ENET_DEFINES = [
    "-DHAS_FCNTL=1", "-DHAS_POLL=1", "-DHAS_GETADDRINFO=1", "-DHAS_GETNAMEINFO=1",
    "-DHAS_GETHOSTBYNAME_R=1", "-DHAS_GETHOSTBYADDR_R=1", "-DHAS_INET_PTON=1",
    "-DHAS_INET_NTOP=1", "-DHAS_MSGHDR_FLAGS=1", "-DHAS_SOCKLEN_T=1",
]


def fetch(url):
    with urllib.request.urlopen(url, timeout=120) as r:
        return r.read()


def resolve():
    data = lzma.decompress(fetch(f"{MIRROR}/dists/{SUITE}/main/binary-amd64/Packages.xz"))
    found = {}
    for block in data.decode("utf-8").split("\n\n"):
        fields = {}
        for line in block.splitlines():
            if ": " in line and not line.startswith(" "):
                k, v = line.split(": ", 1)
                fields[k] = v
        if fields.get("Package") in PACKAGES:
            found[fields["Package"]] = fields
    missing = [p for p in PACKAGES if p not in found]
    if missing:
        sys.exit(f"not in Debian {SUITE}: {', '.join(missing)}")
    return [found[p] for p in PACKAGES]


def ar_members(data):
    if data[:8] != b"!<arch>\n":
        sys.exit("not a .deb")
    pos = 8
    while pos + 60 <= len(data):
        header = data[pos:pos + 60]
        name = header[:16].decode().strip().rstrip("/")
        size = int(header[48:58])
        pos += 60
        yield name, data[pos:pos + size]
        pos += size + (size & 1)


def inside(root, path):
    dest = os.path.normpath(os.path.join(root, path.lstrip("./").lstrip("/")))
    if os.path.commonpath([root, dest]) != root:
        sys.exit(f"refusing a path outside the sysroot: {path}")
    return dest


def extract_deb(deb, root, links):
    for name, body in ar_members(deb):
        if not name.startswith("data.tar"):
            continue
        mode = {"data.tar.xz": "r:xz", "data.tar.gz": "r:gz", "data.tar": "r:"}.get(name)
        if mode is None:
            sys.exit(f"cannot read {name}")
        with tarfile.open(fileobj=io.BytesIO(body), mode=mode) as tf:
            for m in tf.getmembers():
                dest = inside(root, m.name)
                if m.isdir():
                    os.makedirs(dest, exist_ok=True)
                elif m.issym():
                    links.append((dest, m.linkname))
                elif m.isfile():
                    os.makedirs(os.path.dirname(dest), exist_ok=True)
                    with tf.extractfile(m) as src, open(dest, "wb") as dst:
                        shutil.copyfileobj(src, dst)
        return
    sys.exit("no data.tar in the package")


def copy_links(root, links):
    """Windows cannot be relied on to create symlinks, and the linker needs
    libX11.so and friends, which are symlinks in the packages - so each one
    becomes a copy of what it points at. A link can point at another link, so
    this goes round until nothing more resolves. Links left pointing at
    nothing (documentation, mostly) are dropped."""
    pending = links
    while pending:
        left = []
        for dest, target in pending:
            if target.startswith("/"):
                src = inside(root, target)
            else:
                src = os.path.normpath(os.path.join(os.path.dirname(dest), target))
            if os.path.isfile(src):
                os.makedirs(os.path.dirname(dest), exist_ok=True)
                shutil.copyfile(src, dest)
            else:
                left.append((dest, target))
        if len(left) == len(pending):
            break
        pending = left


def build_enet(out, zig):
    src_root = os.path.join(out, "enet-src")
    with tarfile.open(fileobj=io.BytesIO(fetch(ENET_URL)), mode="r:gz") as tf:
        tf.extractall(src_root, filter="data")
    src = os.path.join(src_root, f"enet-{ENET_VERSION}")
    include = os.path.join(out, "enet", "include")
    shutil.copytree(os.path.join(src, "include"), include)
    lib_dir = os.path.join(out, "enet", "lib")
    obj_dir = os.path.join(out, "enet", "obj")
    os.makedirs(lib_dir)
    os.makedirs(obj_dir)
    objs = []
    for name in ENET_SOURCES:
        obj = os.path.join(obj_dir, name + ".o")
        subprocess.run([zig, "cc", "-target", TARGET, "-O2", "-fPIC", "-I", include, *ENET_DEFINES,
                        "-c", os.path.join(src, name + ".c"), "-o", obj], check=True)
        objs.append(obj)
    subprocess.run([zig, "ar", "rcs", os.path.join(lib_dir, "libenet.a"), *objs], check=True)
    shutil.rmtree(src_root)
    shutil.rmtree(obj_dir)
    return include, lib_dir


def write_pkg_config(out, include, lib_dir):
    fwd = lambda p: p.replace("\\", "/")
    with open(os.path.join(out, "pkg-config.cmd"), "w", newline="\r\n") as f:
        f.write("@echo off\n"
                "rem Stand-in for pkg-config: go-enet asks it for libenet, and nothing else does.\n"
                f'if "%~1"=="--cflags" echo -I{fwd(include)}\n'
                f'if "%~1"=="--libs" echo -L{fwd(lib_dir)} -lenet\n'
                "exit /b 0\n")


def main():
    if len(sys.argv) != 2:
        sys.exit("usage: linux_sysroot.py <directory>")
    out = os.path.abspath(sys.argv[1])
    stamp = os.path.join(out, ".complete")
    if os.path.isfile(stamp) and open(stamp).read().strip() == STAMP:
        return
    zig = shutil.which("zig")
    if not zig:
        sys.exit("zig not found. Install it (scoop install zig).")

    if os.path.isdir(out):
        shutil.rmtree(out)  # an unfinished or outdated directory
    root = os.path.join(out, "sysroot")
    os.makedirs(root)

    print(f"preparing the linux/amd64 sysroot in {out}")
    links = []
    for pkg in resolve():
        deb = fetch(f"{MIRROR}/{pkg['Filename']}")
        if hashlib.sha256(deb).hexdigest() != pkg["SHA256"]:
            sys.exit(f"{pkg['Package']}: checksum mismatch")
        extract_deb(deb, root, links)
    copy_links(root, links)
    for lib in ("libX11.so", "libxkbcommon.so", "libGL.so"):
        if not os.path.isfile(os.path.join(root, "usr", "lib", "x86_64-linux-gnu", lib)):
            sys.exit(f"sysroot is missing {lib}")

    print(f"building enet {ENET_VERSION}")
    include, lib_dir = build_enet(out, zig)
    write_pkg_config(out, include, lib_dir)

    with open(stamp, "w") as f:
        f.write(STAMP)


if __name__ == "__main__":
    main()
