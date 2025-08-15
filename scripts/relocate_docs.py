import os, re, pathlib, sys, hashlib

ROOT = pathlib.Path(__file__).resolve().parents[1]
MD = list(ROOT.rglob("*.md"))

# Mapping of known old filenames to new paths (extend as needed)
MAPPING = {
  "AGGRESSIVE_REPUTATION_DECAY_IMPLEMENTATION.md": "docs/economics/aggressive-reputation-decay.md",
  "DIFFERENTIATED_DISPUTE_STAKES_IMPLEMENTATION.md": "docs/economics/differentiated-dispute-stakes.md",
  "STAKE_BASED_DISPUTES_IMPLEMENTATION.md": "docs/economics/differentiated-dispute-stakes.md",
  "COMPLETE_ECONOMIC_MODEL_IMPLEMENTATION.md": "docs/economics/complete-economic-model.md",
  "DISPUTE_RESOLUTION_IMPLEMENTATION.md": "docs/protocol/dispute-resolution.md",
  "IPFS_INTEGRATION.md": "docs/architecture/ipfs-integration.md",
  "IMPLEMENTATION_SUMMARY.md": "docs/overview/implementation-summary.md",
  "ASYNC_REFACTOR_SUMMARY.md": "docs/agent/async-refactor-summary.md",
  "SECURITY_IMPLEMENTATION.md": "docs/security/implementation.md",
  "SECURITY_SCANNING_IMPLEMENTATION.md": "docs/security/scanning.md",
  "VERIFICATION.md": "docs/security/verification.md",
  "REPRODUCIBLE_BUILDS_SUMMARY.md": "docs/operations/reproducible-builds.md",
}

# Auto-learn moved assets under docs/_assets by basename
for p in ROOT.joinpath("docs/_assets").glob("*"):
    MAPPING[p.name] = p.as_posix()

LINK = re.compile(r'\[([^\]]+)\]\(([^)]+)\)')
def rewrite(md_path: pathlib.Path):
    txt = md_path.read_text(encoding="utf-8", errors="ignore")
    def repl(m):
        label, target = m.group(1), m.group(2)
        # Skip external links and anchors
        if target.startswith(("http://","https://","#","mailto:")):
            return m.group(0)
        base = target.split('#',1)[0]
        anchor = '' if '#' not in target else '#' + target.split('#',1)[1]
        name = pathlib.Path(base).name
        if name in MAPPING:
            new_abs = ROOT / MAPPING[name]
            rel = os.path.relpath(new_abs, start=md_path.parent)
            return f'[{label}]({rel.replace(os.sep,"/")}{anchor})'
        return m.group(0)
    new = LINK.sub(repl, txt)
    if new != txt:
        md_path.write_text(new, encoding="utf-8")

for f in MD:
    # Skip module READMEs
    parts = {p.lower() for p in f.parts}
    if any(x in parts for x in {".git","node_modules",".venv","venv"}):
        continue
    rewrite(f)
