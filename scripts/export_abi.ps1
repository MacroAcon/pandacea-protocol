New-Item -ItemType Directory -Force -Path abi | Out-Null
forge inspect LeaseAgreement abi | Out-File -Encoding utf8 abi/LeaseAgreement.json
forge inspect Reputation abi | Out-File -Encoding utf8 abi/Reputation.json
forge inspect PGT abi | Out-File -Encoding utf8 abi/PGT.json
