#!/usr/bin/env bash
set -euo pipefail
mkdir -p abi
# example with forge; adjust names:
forge inspect LeaseAgreement abi > abi/LeaseAgreement.json
forge inspect Reputation abi > abi/Reputation.json
forge inspect PGT abi > abi/PGT.json
