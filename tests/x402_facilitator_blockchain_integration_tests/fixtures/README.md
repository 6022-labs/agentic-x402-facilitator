# Contract fixtures

`artifacts/*.json` are trimmed solc artifacts (`{contractName, abi, bytecode}`) that these
integration tests deploy onto a throwaway anvil chain. They are committed on purpose, next to
their sources in `contracts/`.

- `TestEip3009Token`: the slice of USDC's EIP-3009 the facilitator settles against
  (`balanceOf`, `authorizationState`, `transferWithAuthorization(v, r, s)`), the ERC-20 allowance Permit2 pulls
  through, plus a free `mint` and `setAllowance`.
- `TestPermit2` and `TestX402ExactPermit2Proxy`: Permit2's `permitWitnessTransferFrom` with its unordered nonces,
  and the x402 proxy's `settle`. `InstallPermit2` puts their runtime code at the canonical addresses.
- `TestAgentCollectionsManager`: answers `findAgentByAddress` like the real manager, over two
  pages, so the registry lookup must follow `next`. `register(value, tokenId)` adds an address.

## Refreshing

Only after editing a source:

```sh
docker run --rm -v "$PWD/contracts:/sources" ethereum/solc:0.8.28 --optimize \
  --combined-json abi,bin,bin-runtime /sources/TestEip3009Token.sol /sources/TestAgentCollectionsManager.sol \
  /sources/TestPermit2.sol /sources/TestX402ExactPermit2Proxy.sol > combined.json
python3 -c "
import json
for key, c in json.load(open('combined.json'))['contracts'].items():
    name = key.split(':')[1]
    if name.startswith('ITest'):
        continue
    artifact = {'contractName': name, 'abi': c['abi'], 'bytecode': '0x' + c['bin']}
    if name in ('TestPermit2', 'TestX402ExactPermit2Proxy'):
        artifact['deployedBytecode'] = '0x' + c['bin-runtime']
    json.dump(artifact, open(f'artifacts/{name}.json', 'w'), indent=2)
"
rm combined.json
```
