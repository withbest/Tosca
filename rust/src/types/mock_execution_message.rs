use evmc_vm::{Address, ExecutionMessage, MessageKind, Uint256};

use crate::types::u256;

/// The same as ExecutionMessage but with `pub` fields for easier testing.
#[derive(Debug)]
pub struct MockExecutionMessage<'a> {
    pub kind: MessageKind,
    pub flags: u32,
    pub depth: i32,
    pub gas: i64,
    pub recipient: Address,
    pub sender: Address,
    pub input: Option<&'a [u8]>,
    pub value: Uint256,
    pub create2_salt: Uint256,
    pub code_address: Address,
    pub code: Option<&'a [u8]>,
}

impl<'a> MockExecutionMessage<'a> {
    pub const DEFAULT_INIT_GAS: u64 = 1_000_000;
}

impl<'a> Default for MockExecutionMessage<'a> {
    fn default() -> Self {
        MockExecutionMessage {
            kind: MessageKind::EVMC_CALL,
            flags: 0,
            depth: 1,
            gas: Self::DEFAULT_INIT_GAS as i64,
            recipient: u256::ZERO.into(),
            sender: u256::ZERO.into(),
            input: None,
            value: u256::ZERO.into(),
            create2_salt: u256::ZERO.into(),
            code_address: u256::ZERO.into(),
            code: None,
        }
    }
}

impl<'a> From<MockExecutionMessage<'a>> for ExecutionMessage {
    fn from(value: MockExecutionMessage) -> Self {
        Self::new(
            value.kind,
            value.flags,
            value.depth,
            value.gas,
            value.recipient,
            value.sender,
            value.input,
            value.value,
            value.create2_salt,
            value.code_address,
            value.code,
        )
    }
}