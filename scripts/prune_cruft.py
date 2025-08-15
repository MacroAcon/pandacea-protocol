import os, pathlib, hashlib, shutil

ROOT = pathlib.Path(__file__).resolve().parents[1]
ASSETS = ROOT / "docs" / "_assets"
ASSETS.mkdir(parents=True, exist_ok=True)

CRUFT_PATTERNS = ("Thumbs.db",".DS_Store",".Spotlight-V100",".Trashes",".TemporaryItems")
DELETE_EXT = (".log",".tmp",".bak",".old",".orig","~")

def sha256(p: pathlib.Path, buf=bytearray(1024*1024)):
    h = hashlib.sha256()
    with p.open("rb") as f:
        while True:
            n = f.readinto(buf)
            if not n: break
            h.update(memoryview(buf)[:n])
    return h.hexdigest()

seen = {}
for p in ROOT.rglob("*"):
    if p.is_dir(): continue
    name = p.name
    lower = name.lower()
    if name in CRUFT_PATTERNS or lower.endswith(DELETE_EXT):
        try: p.unlink()
        except: pass
        continue
    # Move duplicate images/pdfs to docs/_assets (keep one canonical)
    if p.suffix.lower() in {".png",".jpg",".jpeg",".svg",".gif",".pdf"}:
        h = sha256(p)
        if h in seen:
            # duplicate; delete this one
            try: p.unlink()
            except: pass
        else:
            seen[h] = p
            # if not already under docs/_assets, move it there
            if ASSETS not in p.parents:
                dest = ASSETS / name.replace(" ", "-")
                # avoid clobber
                i=1
                stem, suf = dest.stem, dest.suffix
                while dest.exists():
                    dest = dest.with_name(f"{stem}-{i}{suf}")
                    i+=1
                p.replace(dest)
