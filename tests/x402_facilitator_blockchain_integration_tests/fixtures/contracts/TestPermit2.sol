// SPDX-License-Identifier: MIT
pragma solidity 0.8.28;

interface ITestPermit2Token {
    function transferFrom(address from, address to, uint256 value) external returns (bool);
}

// Minimal Permit2: permitWitnessTransferFrom with its unordered nonces, installed at Permit2's canonical address.
contract TestPermit2 {
    struct TokenPermissions {
        address token;
        uint256 amount;
    }

    struct PermitTransferFrom {
        TokenPermissions permitted;
        uint256 nonce;
        uint256 deadline;
    }

    struct SignatureTransferDetails {
        address to;
        uint256 requestedAmount;
    }

    bytes32 private constant TOKEN_PERMISSIONS_TYPEHASH = keccak256("TokenPermissions(address token,uint256 amount)");
    string private constant PERMIT_WITNESS_TRANSFER_FROM_TYPEHASH_STUB =
        "PermitWitnessTransferFrom(TokenPermissions permitted,address spender,uint256 nonce,uint256 deadline,";

    mapping(address => mapping(uint256 => uint256)) public nonceBitmap;

    // Computed per call: the code is installed at an address it was not deployed to.
    function DOMAIN_SEPARATOR() public view returns (bytes32) {
        return keccak256(
            abi.encode(
                keccak256("EIP712Domain(string name,uint256 chainId,address verifyingContract)"),
                keccak256("Permit2"),
                block.chainid,
                address(this)
            )
        );
    }

    function permitWitnessTransferFrom(
        PermitTransferFrom calldata permit,
        SignatureTransferDetails calldata transferDetails,
        address owner,
        bytes32 witness,
        string calldata witnessTypeString,
        bytes calldata signature
    ) external {
        require(block.timestamp <= permit.deadline, "signature expired");
        require(transferDetails.requestedAmount <= permit.permitted.amount, "invalid amount");
        require(recover(digest(permit, witness, witnessTypeString), signature) == owner, "invalid signature");
        useNonce(owner, permit.nonce);
        require(
            ITestPermit2Token(permit.permitted.token).transferFrom(
                owner,
                transferDetails.to,
                transferDetails.requestedAmount
            ),
            "transfer failed"
        );
    }

    function digest(
        PermitTransferFrom calldata permit,
        bytes32 witness,
        string calldata witnessTypeString
    ) private view returns (bytes32) {
        bytes32 typeHash = keccak256(abi.encodePacked(PERMIT_WITNESS_TRANSFER_FROM_TYPEHASH_STUB, witnessTypeString));
        bytes32 tokenPermissionsHash = keccak256(abi.encode(TOKEN_PERMISSIONS_TYPEHASH, permit.permitted));
        bytes32 dataHash = keccak256(
            abi.encode(typeHash, tokenPermissionsHash, msg.sender, permit.nonce, permit.deadline, witness)
        );
        return keccak256(abi.encodePacked("\x19\x01", DOMAIN_SEPARATOR(), dataHash));
    }

    function recover(bytes32 hash, bytes calldata signature) private pure returns (address) {
        require(signature.length == 65, "invalid signature length");
        return ecrecover(hash, uint8(signature[64]), bytes32(signature[0:32]), bytes32(signature[32:64]));
    }

    function useNonce(address owner, uint256 nonce) private {
        uint256 word = nonce >> 8;
        uint256 bit = 1 << (nonce & 0xff);
        require(nonceBitmap[owner][word] & bit == 0, "invalid nonce");
        nonceBitmap[owner][word] |= bit;
    }
}
