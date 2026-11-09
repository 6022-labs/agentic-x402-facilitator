// SPDX-License-Identifier: MIT
pragma solidity 0.8.28;

import "./TestPermit2.sol";

// x402ExactPermit2Proxy's settle, installed at its canonical address: Permit2 sends the exact amount to witness.to.
contract TestX402ExactPermit2Proxy {
    struct Witness {
        address to;
        uint256 validAfter;
    }

    TestPermit2 private constant PERMIT2 = TestPermit2(0x000000000022D473030F116dDEE9F6B43aC78BA3);
    string private constant WITNESS_TYPE_STRING =
        "Witness witness)TokenPermissions(address token,uint256 amount)Witness(address to,uint256 validAfter)";
    bytes32 private constant WITNESS_TYPEHASH = keccak256("Witness(address to,uint256 validAfter)");

    function settle(
        TestPermit2.PermitTransferFrom calldata permit,
        address owner,
        Witness calldata witness,
        bytes calldata signature
    ) external {
        require(witness.to != address(0), "invalid destination");
        require(block.timestamp >= witness.validAfter, "payment too early");
        bytes32 witnessHash = keccak256(abi.encode(WITNESS_TYPEHASH, witness.to, witness.validAfter));
        PERMIT2.permitWitnessTransferFrom(
            permit,
            TestPermit2.SignatureTransferDetails({to: witness.to, requestedAmount: permit.permitted.amount}),
            owner,
            witnessHash,
            WITNESS_TYPE_STRING,
            signature
        );
    }
}
