import pathlib, hashlib, json, os, datetime

ROOT = pathlib.Path(__file__).resolve().parents[1]
DOCS = ROOT / "docs"
out = []

for p in sorted(DOCS.rglob("*")):
    if p.is_dir(): 
        continue
    rel = p.relative_to(ROOT).as_posix()
    size = p.stat().st_size
    sha = hashlib.sha256(p.read_bytes()).hexdigest()[:12]
    out.append({"path": rel, "kb": round(size/1024,1), "sha12": sha})

audit = {
    "generated_at": datetime.datetime.utcnow().isoformat() + "Z",
    "count": len(out),
    "files": out,
}

MD = ["# Docs Audit", "", f"_Generated: {audit['generated_at']}_", ""]
MD.append(f"Total files: **{audit['count']}**")
MD.append("")
MD.append("| Path | Size (KB) | SHA-256 (12) |")
MD.append("| --- | ---: | --- |")
for f in out:
    MD.append(f"| {f['path']} | {f['kb']} | `{f['sha12']}` |")

(ROOT/"docs"/"AUDIT.md").write_text("\n".join(MD), encoding="utf-8")
print(f"Wrote docs/AUDIT.md with {audit['count']} entries")
