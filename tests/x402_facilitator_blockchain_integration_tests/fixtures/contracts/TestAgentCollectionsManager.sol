// SPDX-License-Identifier: MIT
pragma solidity 0.8.28;

// AgentCollectionsManager stand-in: findAgentByAddress spans two pages, so callers must follow `next`.
contract TestAgentCollectionsManager {
    address public constant COLLECTION = address(0x6022);

    mapping(string => uint256) private tokenIds;

    function register(string calldata value, uint256 tokenId) external {
        tokenIds[value] = tokenId;
    }

    function findAgentByAddress(
        string memory addressType,
        string memory value,
        uint256 offset,
        uint256
    ) external view returns (address collection, uint256 tokenId, string memory name, uint256 next) {
        if (offset < 2) {
            return (address(0), 0, "", 2);
        }
        uint256 id = tokenIds[value];
        if (id != 0 && keccak256(bytes(addressType)) == keccak256("evm")) {
            return (COLLECTION, id, "agent", 0);
        }
        return (address(0), 0, "", 0);
    }
}
