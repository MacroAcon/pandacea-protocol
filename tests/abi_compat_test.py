import json
import pathlib

REQUIRED = {
    "LeaseAgreement.json": {"createLease", "getLease"},
    "Reputation.json": {"scoreOf"},
    "PGT.json": {"transfer", "balanceOf"},
}

def test_abi_has_required_functions():
    abi_dir = pathlib.Path("abi")
    for name, req in REQUIRED.items():
        path = abi_dir / name
        data = json.loads(path.read_text(encoding="utf-8"))
        funcs = {x.get("name") for x in data if x.get("type") == "function"}
        missing = req - funcs
        assert not missing, f"{name} missing {missing}"
