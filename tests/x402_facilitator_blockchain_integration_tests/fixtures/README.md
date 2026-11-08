# Contract fixtures

`artifacts/*.json` are trimmed solc artifacts (`{contractName, abi, bytecode}`) that these
integration tests deploy onto a throwaway anvil chain. They are committed on purpose, next to
their sources in `contracts/`.

- `TestEip3009Token`: the slice of USDC's EIP-3009 the facilitator settles against
  (`balanceOf`, `authorizationState`, `transferWithAuthorization(v, r, s)`), plus a free `mint`.
- `TestAgentCollectionsManager`: answers `findAgentByAddress` like the real manager, over two
  pages, so the registry lookup must follow `next`. `register(value, tokenId)` adds an address.

## Refreshing

Only after editing a source:

```sh
docker run --rm -v "$PWD/contracts:/sources" ethereum/solc:0.8.28 --optimize \
  --combined-json abi,bin /sources/TestEip3009Token.sol /sources/TestAgentCollectionsManager.sol > combined.json
python3 -c "
import json
for key, c in json.load(open('combined.json'))['contracts'].items():
    name = key.split(':')[1]
    json.dump({'contractName': name, 'abi': c['abi'], 'bytecode': '0x' + c['bin']}, open(f'artifacts/{name}.json', 'w'), indent=2)
"
rm combined.json
```
