//! The objects a backend lowers to.
//!
//! A value is immediate data or a pointer-sized payload. An actor owns its field
//! slots and a mailbox. A frame owns the locals of one turn. Names are resolved
//! by the front end into these slots. The scheduler only sees [`ActorState`].

use std::collections::{HashMap, VecDeque};

use crate::ast::Stmt;

/// One value. Scalars are immediate. The rest are pointer-sized payloads.
#[derive(Clone, Debug)]
pub enum Value {
    Unit,
    Int(IntVal),
    Float(f64),
    Bool(bool),
    Str(String),
    Pid(u64),
    Result(Result<Box<Value>, Box<Value>>),
    Option(Option<Box<Value>>),
    Enum {
        type_name: String,
        variant: String,
        fields: Vec<(String, Value)>,
    },
    Error(ErrorObj),
    /// `Type.variant`, before the call supplies the fields.
    Ctor { type_name: String, variant: String },
    /// `recv.name`, before the call runs.
    Method {
        recv: Box<Value>,
        type_name: String,
        name: String,
    },
    /// A pid call that has not been awaited.
    Promise(u64),
    Module(HashMap<String, Value>),
    Type(String),
    Native(Native),
    /// Interpreter only. A backend replaces this with a code pointer.
    Func(Box<FuncObj>),
}

#[derive(Clone, Debug)]
pub(crate) struct FuncObj {
    pub(crate) params: Vec<String>,
    pub(crate) body: Vec<Stmt>,
    pub(crate) env: HashMap<String, Value>,
}

#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum IntKind {
    I8,
    I16,
    I32,
    I64,
    I128,
    U8,
    U16,
    U32,
    U64,
    U128,
    Int,
    Usize,
    Isize,
}

impl IntKind {
    pub(crate) fn from_name(name: &str) -> Option<Self> {
        Some(match name {
            "i8" => Self::I8,
            "i16" => Self::I16,
            "i32" => Self::I32,
            "i64" => Self::I64,
            "i128" => Self::I128,
            "u8" => Self::U8,
            "u16" => Self::U16,
            "u32" => Self::U32,
            "u64" => Self::U64,
            "u128" => Self::U128,
            "int" => Self::Int,
            "usize" => Self::Usize,
            "isize" => Self::Isize,
            _ => return None,
        })
    }

    fn bits(self) -> u32 {
        match self {
            Self::I8 | Self::U8 => 8,
            Self::I16 | Self::U16 => 16,
            Self::I32 | Self::U32 | Self::Int => 32,
            Self::I64 | Self::U64 | Self::Usize | Self::Isize => 64,
            Self::I128 | Self::U128 => 128,
        }
    }

    fn signed(self) -> bool {
        !matches!(self, Self::U8 | Self::U16 | Self::U32 | Self::U64 | Self::U128 | Self::Usize)
    }
}

#[derive(Clone, Copy, Debug)]
pub struct IntVal {
    pub kind: IntKind,
    pub bits: i128,
}

impl IntVal {
    pub(crate) fn new(kind: IntKind, raw: i128) -> Self {
        Self { kind, bits: wrap(kind, raw) }
    }

    pub(crate) fn as_i64(self) -> i64 {
        self.bits as i64
    }
}

fn wrap(kind: IntKind, raw: i128) -> i128 {
    let bits = kind.bits();
    if bits >= 128 {
        return raw;
    }
    let mask = (1i128 << bits) - 1;
    let masked = raw & mask;
    if kind.signed() {
        let sign = 1i128 << (bits - 1);
        if masked & sign != 0 { masked - (1i128 << bits) } else { masked }
    } else {
        masked
    }
}

#[derive(Clone, Debug)]
pub struct ErrorObj {
    pub tag: String,
    pub text: String,
    pub cause: Option<Box<Value>>,
}

#[derive(Clone, Copy, Debug)]
pub enum Native {
    Fprintf,
    Is,
    TestOnly,
    FailCheck,
    Ok,
    Err,
    Some,
    None,
}

/// What the scheduler is allowed to do with an actor.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum ActorState {
    /// Its mailbox may be run.
    Ready,
    /// Parked in `await`.
    Waiting,
    /// The body ended, or `unwrap` killed it.
    Dead,
}

/// Locals of one turn. The slot index is the backend name of the local.
#[derive(Clone, Debug)]
pub struct Frame {
    pub names: Vec<String>,
    pub slots: Vec<Value>,
}

impl Frame {
    fn get(&self, name: &str) -> Option<Value> {
        self.names.iter().position(|n| n == name).map(|i| self.slots[i].clone())
    }

    fn set(&mut self, name: &str, value: Value) {
        if let Some(i) = self.names.iter().position(|n| n == name) {
            self.slots[i] = value;
        } else {
            self.names.push(name.to_string());
            self.slots.push(value);
        }
    }
}

/// One actor. Field slots outlive a turn. Frames do not.
#[derive(Debug)]
pub struct Actor {
    pub id: u64,
    pub type_name: String,
    pub state: ActorState,
    pub field_names: Vec<String>,
    pub fields: Vec<Value>,
    pub frames: Vec<Frame>,
    pub mailbox: VecDeque<Message>,
    pub check_failed: bool,
}

impl Actor {
    pub(crate) fn bare(id: u64, type_name: String) -> Self {
        Self {
            id,
            type_name,
            state: ActorState::Ready,
            field_names: Vec::new(),
            fields: Vec::new(),
            frames: Vec::new(),
            mailbox: VecDeque::new(),
            check_failed: false,
        }
    }

    pub(crate) fn local(&self, name: &str) -> Option<Value> {
        self.frames.iter().rev().find_map(|frame| frame.get(name))
    }

    pub(crate) fn set_local(&mut self, name: &str, value: Value) {
        if self.frames.is_empty() {
            self.frames.push(Frame { names: Vec::new(), slots: Vec::new() });
        }
        self.frames.last_mut().unwrap().set(name, value);
    }

    pub(crate) fn field(&self, name: &str) -> Option<Value> {
        self.field_names.iter().position(|n| n == name).map(|i| self.fields[i].clone())
    }

    pub(crate) fn set_field(&mut self, name: &str, value: Value) -> bool {
        let Some(i) = self.field_names.iter().position(|n| n == name) else {
            return false;
        };
        self.fields[i] = value;
        true
    }

    pub(crate) fn push_frame(&mut self) {
        self.frames.push(Frame { names: Vec::new(), slots: Vec::new() });
    }

    pub(crate) fn pop_frame(&mut self) {
        self.frames.pop();
    }
}

/// One queued pid call.
#[derive(Debug)]
pub struct Message {
    pub method: String,
    pub args: Vec<Value>,
    pub promise: u64,
}

/// `None` while the handler has not returned.
#[derive(Debug)]
pub struct Promise {
    pub resolved: Option<Value>,
}
