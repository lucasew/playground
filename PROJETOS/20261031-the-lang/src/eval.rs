use std::collections::HashMap;

use crate::ast::*;
use crate::runtime::*;



#[derive(Clone, Debug)]
struct Method {
    /// `Type.method(self, ...)` is an instance method. `self` is the receiver, not a call argument.
    has_self: bool,
    params: Vec<String>,
    body: Vec<Stmt>,
}

#[derive(Clone, Debug)]
struct ActorDef {
    params: Vec<String>,
    fields: Vec<Field>,
}

#[derive(Clone, Debug)]
struct EnumDef {
    variants: Vec<Variant>,
}

enum Flow {
    Dead,
    Error(String),
}

#[derive(Debug)]
pub struct Outcome {
    pub status: i32,
    pub stderr: String,
}

struct Vm<'a> {
    program: &'a Program,
    methods: HashMap<(String, String), Method>,
    actors_def: HashMap<String, ActorDef>,
    enums: HashMap<String, EnumDef>,
    actors: HashMap<u64, Actor>,
    next_id: u64,
    promises: HashMap<u64, Promise>,
    next_promise: u64,
    testing: bool,
    stderr: String,
    current: u64,
}

impl<'a> Vm<'a> {
    fn new(program: &'a Program, testing: bool) -> Self {
        let mut methods = HashMap::new();
        let mut actors_def = HashMap::new();
        let mut enums = HashMap::new();
        for item in &program.items {
            match item {
                Item::Method { type_name, name, params, body } => {
                    let has_self = params.first().is_some_and(|p| p == "self");
                    let params = if has_self { params[1..].to_vec() } else { params.clone() };
                    methods.insert(
                        (type_name.clone(), name.clone()),
                        Method { has_self, params, body: body.clone() },
                    );
                }
                Item::Actor { name, params, fields } => {
                    actors_def.insert(name.clone(), ActorDef { params: params.clone(), fields: fields.clone() });
                }
                Item::Enum { name, variants } => {
                    enums.insert(name.clone(), EnumDef { variants: variants.clone() });
                }
                _ => {}
            }
        }
        Self {
            program,
            methods,
            actors_def,
            enums,
            actors: HashMap::new(),
            next_id: 1,
            promises: HashMap::new(),
            next_promise: 1,
            testing,
            stderr: String::new(),
            current: 0,
        }
    }

    fn spawn_bare(&mut self, type_name: &str) -> u64 {
        let id = self.next_id;
        self.next_id += 1;
        self.actors.insert(id, Actor::bare(id, type_name.into()));
        self.current = id;
        id
    }

    fn dead_error() -> Value {
        Value::Error(ErrorObj { tag: "dead".into(), text: "dead".into(), cause: None })
    }

    fn exec_block(&mut self, actor: u64, body: &[Stmt]) -> Result<Value, Flow> {
        let saved = self.current;
        self.current = actor;
        let mut last = Value::Unit;
        for stmt in body {
            last = self.exec_stmt(stmt)?;
        }
        self.current = saved;
        Ok(last)
    }

    fn exec_stmt(&mut self, stmt: &Stmt) -> Result<Value, Flow> {
        match stmt {
            Stmt::TestBlock(body) => {
                if self.testing {
                    self.exec_block(self.current, body)?;
                }
                Ok(Value::Unit)
            }
            Stmt::Expr(e) => self.eval(e),
            Stmt::Assign { target, op, value } => {
                let rhs = self.eval(value)?;
                let new_val = match op {
                    AssignOp::Eq => self.coerce_to_target(target, rhs)?,
                    AssignOp::PlusEq => {
                        let cur = self.eval(target)?;
                        self.add(&cur, &self.coerce_int_literal(&rhs, &cur)?)?
                    }
                };
                self.store(target, new_val.clone())?;
                Ok(new_val)
            }
        }
    }

    fn coerce_to_target(&mut self, target: &Expr, rhs: Value) -> Result<Value, Flow> {
        if let Value::Int(lit) = &rhs {
            if lit.kind == IntKind::Int && matches!(target, Expr::Name(_) | Expr::Field { .. }) {
                if let Ok(Value::Int(cur)) = self.eval(target) {
                    return Ok(Value::Int(IntVal::new(cur.kind, lit.bits)));
                }
            }
        }
        Ok(rhs)
    }

    fn coerce_int_literal(&self, rhs: &Value, cur: &Value) -> Result<Value, Flow> {
        match (rhs, cur) {
            (Value::Int(lit), Value::Int(cur)) if lit.kind == IntKind::Int => {
                Ok(Value::Int(IntVal::new(cur.kind, lit.bits)))
            }
            _ => Ok(rhs.clone()),
        }
    }

    fn store(&mut self, target: &Expr, value: Value) -> Result<(), Flow> {
        match target {
            Expr::Name(name) => {
                let actor = self.actors.get_mut(&self.current).ok_or_else(|| Flow::Error("no actor".into()))?;
                actor.set_local(name, value);
                Ok(())
            }
            Expr::Field { base, name } => {
                if let Expr::Name(b) = base.as_ref() {
                    if b == "self" {
                        return self.store_self_field(name, value);
                    }
                }
                Err(Flow::Error("assignment target is not a place".into()))
            }
            _ => Err(Flow::Error("assignment target is not a place".into())),
        }
    }

    fn eval(&mut self, expr: &Expr) -> Result<Value, Flow> {
        match expr {
            Expr::Name(n) => self.lookup(n),
            Expr::Int(n) => Ok(Value::Int(IntVal::new(IntKind::Int, *n as i128))),
            Expr::Float(n) => Ok(Value::Float(*n)),
            Expr::Str(s) => Ok(Value::Str(s.clone())),
            Expr::Bool(b) => Ok(Value::Bool(*b)),
            Expr::UnaryMinus(inner) => match self.eval(inner)? {
                Value::Int(i) => Ok(Value::Int(IntVal::new(i.kind, -i.bits))),
                Value::Float(f) => Ok(Value::Float(-f)),
                _ => Err(Flow::Error("unary minus needs a number".into())),
            },
            Expr::Binary { op, left, right } => {
                let l = self.eval(left)?;
                let r = self.eval(right)?;
                let r = self.coerce_int_literal(&r, &l)?;
                let l = self.coerce_int_literal(&l, &r)?;
                if let (Value::Float(a), Value::Float(b)) = (&l, &r) {
                    return Ok(match op {
                        BinOp::Add => Value::Float(a + b),
                        BinOp::Sub => Value::Float(a - b),
                        BinOp::Mul => Value::Float(a * b),
                        BinOp::Div => Value::Float(a / b),
                        BinOp::Eq => Value::Bool(a == b),
                    });
                }
                match op {
                    BinOp::Add => self.add(&l, &r),
                    BinOp::Sub => self.arith(&l, &r, |a, b| a - b),
                    BinOp::Mul => self.arith(&l, &r, |a, b| a * b),
                    BinOp::Div => self.arith(&l, &r, |a, b| if b == 0 { 0 } else { a / b }),
                    BinOp::Eq => Ok(Value::Bool(values_eq(&l, &r))),
                }
            }
            Expr::Field { base, name } => {
                let v = self.eval(base)?;
                self.field(v, name)
            }
            Expr::TypeApp { base, args } => {
                let v = self.eval(base)?;
                match v {
                    Value::Type(name) => Ok(Value::Enum {
                        type_name: name,
                        variant: format!("${}", args.iter().map(|a| a.name.as_str()).collect::<Vec<_>>().join(",")),
                        fields: Vec::new(),
                    }),
                    other => Ok(other),
                }
            }
            Expr::Construct { name, args, fields } => self.construct(name, args, fields),
            Expr::Call { callee, args } => self.call(callee, args),
            Expr::NamedConstruct { base, variant, fields } => {
                let Value::Type(type_name) = self.eval(base)? else {
                    return Err(Flow::Error("named construct needs a type".into()));
                };
                let declared = self.enums.get(&type_name).and_then(|def| {
                    def.variants.iter().find(|v| v.name == *variant).map(|v| v.fields.clone())
                });
                let mut built = Vec::new();
                for (name, expr) in fields {
                    let mut value = self.eval(expr)?;
                    if let Some(field) = declared.as_ref().and_then(|fs| fs.iter().find(|f| f.name == *name)) {
                        if let Some(ty) = &field.ty {
                            value = coerce_to_ty(value, ty);
                        }
                    }
                    built.push((name.clone(), value));
                }
                Ok(Value::Enum { type_name, variant: variant.clone(), fields: built })
            }
            Expr::Await(inner) => {
                let v = self.eval(inner)?;
                match v {
                    Value::Promise(id) => self.await_promise(id),
                    other => Ok(other),
                }
            }
            Expr::Block(stmts) => self.exec_block(self.current, stmts),
            Expr::Match { expr, arms } => {
                let v = self.eval(expr)?;
                for arm in arms {
                    if let Some(binds) = self.match_pat(&arm.pattern, &v) {
                        for (k, val) in binds {
                            self.actors.get_mut(&self.current).unwrap().set_local(&k, val);
                        }
                        return self.eval(&arm.body);
                    }
                }
                Err(Flow::Error("match was not exhaustive at runtime".into()))
            }
        }
    }

    fn lookup(&self, name: &str) -> Result<Value, Flow> {
        if let Some(actor) = self.actors.get(&self.current) {
            if let Some(v) = actor.local(name) {
                return Ok(v);
            }
        }
        if name == "self" {
            return Ok(Value::Pid(self.current));
        }
        if name == "error" {
            return Ok(self.error_module());
        }
        if name == "ok" {
            return Ok(Value::Native(Native::Ok));
        }
        if name == "err" {
            return Ok(Value::Native(Native::Err));
        }
        if name == "some" {
            return Ok(Value::Native(Native::Some));
        }
        if name == "none" {
            return Ok(Value::Native(Native::None));
        }
        if name == "self" {
            return Err(Flow::Error("self is not bound".into()));
        }
        if self.actors_def.contains_key(name) || self.enums.contains_key(name) || is_builtin_type(name) {
            return Ok(Value::Type(name.into()));
        }
        Err(Flow::Error(format!("unknown name {name}")))
    }

    fn error_module(&self) -> Value {
        let mut map = HashMap::new();
        map.insert("dead".into(), Self::dead_error());
        map.insert("is".into(), Value::Native(Native::Is));
        Value::Module(map)
    }

    fn field(&self, base: Value, name: &str) -> Result<Value, Flow> {
        match base {
            Value::Module(map) => map.get(name).cloned().ok_or_else(|| Flow::Error(format!("no field {name}"))),
            Value::Type(type_name) => Ok(Value::Ctor { type_name, variant: name.into() }),
            Value::Pid(pid) => {
                if let Some(v) = self.actors.get(&pid).and_then(|a| a.field(name)) {
                    return Ok(v);
                }
                let type_name = self.actors.get(&pid).map(|a| a.type_name.clone()).unwrap_or_default();
                Ok(Value::Method { recv: Box::new(Value::Pid(pid)), type_name, name: name.into() })
            }
            Value::Enum { type_name, variant, fields } => {
                if let Some((_, v)) = fields.iter().find(|(n, _)| n == name) {
                    return Ok(v.clone());
                }
                Ok(Value::Method {
                    recv: Box::new(Value::Enum { type_name: type_name.clone(), variant, fields }),
                    type_name,
                    name: name.into(),
                })
            }
            Value::Result(_) => Ok(Value::Method {
                recv: Box::new(base),
                type_name: "Result".into(),
                name: name.into(),
            }),
            Value::Option(_) => Ok(Value::Method {
                recv: Box::new(base),
                type_name: "Option".into(),
                name: name.into(),
            }),
            Value::Int(_) if name == "unwrap" => Err(Flow::Error("unwrap is not a method of int".into())),
            _ => Err(Flow::Error(format!("no field {name}"))),
        }
    }

    fn call(&mut self, callee: &Expr, args: &[Expr]) -> Result<Value, Flow> {
        let mut values = Vec::new();
        for arg in args {
            values.push(self.eval(arg)?);
        }
        let fun = self.eval(callee)?;
        match fun {
            Value::Native(n) => self.native(n, values),
            Value::Func(f) => self.call_func(*f, values),
            Value::Method { recv, type_name, name } => self.invoke(*recv, &type_name, &name, values),
            Value::Ctor { type_name, variant } => self.make_enum(&type_name, &variant, values),
            Value::Type(type_name) if self.enums.contains_key(&type_name) && args.is_empty() => {
                Err(Flow::Error("enum needs a variant".into()))
            }
            other => Err(Flow::Error(format!("cannot call {other:?}"))),
        }
    }

    fn make_enum(&self, type_name: &str, variant: &str, args: Vec<Value>) -> Result<Value, Flow> {
        let def = self.enums.get(type_name).ok_or_else(|| Flow::Error("no enum".into()))?.clone();
        let var = def.variants.iter().find(|v| v.name == variant).ok_or_else(|| Flow::Error(format!("no variant {variant}")))?;
        let mut fields = Vec::new();
        if var.named && args.is_empty() {
            return Err(Flow::Error("named variant needs fields".into()));
        }
        for (i, arg) in args.into_iter().enumerate() {
            let name = var.fields.get(i).map(|f| f.name.clone()).unwrap_or_else(|| format!("_{i}"));
            let arg = match var.fields.get(i).and_then(|f| f.ty.as_ref()) {
                Some(ty) => coerce_to_ty(arg, ty),
                None => arg,
            };
            fields.push((name, arg));
        }
        Ok(Value::Enum { type_name: type_name.into(), variant: variant.into(), fields })
    }

    fn invoke(&mut self, recv: Value, type_name: &str, method: &str, args: Vec<Value>) -> Result<Value, Flow> {
        if method == "unwrap" {
            return self.unwrap_value(recv);
        }
        match recv {
            Value::Pid(pid) => self.rpc(pid, method, args),
            Value::Enum { .. } => {
                let key = (type_name.to_string(), method.to_string());
                if self.methods.contains_key(&key) {
                    return self.run_method(type_name, method, Some(recv), args);
                }
                Err(Flow::Error(format!("no method {type_name}.{method}")))
            }
            _ => Err(Flow::Error(format!("no method {type_name}.{method}"))),
        }
    }

    fn unwrap_value(&mut self, recv: Value) -> Result<Value, Flow> {
        match recv {
            Value::Result(Ok(v)) => Ok(*v),
            Value::Result(Err(_)) => {
                self.actors.get_mut(&self.current).map(|a| a.state = ActorState::Dead);
                Err(Flow::Dead)
            }
            Value::Option(Some(v)) => Ok(*v),
            Value::Option(None) => {
                self.actors.get_mut(&self.current).map(|a| a.state = ActorState::Dead);
                Err(Flow::Dead)
            }
            _ => Err(Flow::Error("unwrap needs Result or Option".into())),
        }
    }

    fn rpc(&mut self, pid: u64, method: &str, args: Vec<Value>) -> Result<Value, Flow> {
        let alive = self.actors.get(&pid).is_some_and(|a| a.state != ActorState::Dead);
        let promise = self.next_promise;
        self.next_promise += 1;
        if !alive {
            self.promises.insert(promise, Promise { resolved: Some(err_dead()) });
        } else {
            self.promises.insert(promise, Promise { resolved: None });
            self.actors.get_mut(&pid).unwrap().mailbox.push_back(Message {
                method: method.into(),
                args,
                promise,
            });
        }
        Ok(Value::Promise(promise))
    }

    fn await_promise(&mut self, promise: u64) -> Result<Value, Flow> {
        let me = self.current;
        self.actors.get_mut(&me).map(|a| a.state = ActorState::Waiting);
        loop {
            if let Some(v) = self.promises.get(&promise).and_then(|p| p.resolved.clone()) {
                self.actors.get_mut(&me).map(|a| a.state = ActorState::Ready);
                return Ok(v);
            }
            let next = self.actors.values().find(|a| a.state == ActorState::Ready && !a.mailbox.is_empty()).map(|a| a.id);
            let Some(id) = next else {
                return Err(Flow::Error("deadlock: await never resolves".into()));
            };
            self.run_one(id)?;
        }
    }

    fn run_one(&mut self, id: u64) -> Result<(), Flow> {
        let Some(msg) = self.actors.get_mut(&id).and_then(|a| a.mailbox.pop_front()) else {
            return Ok(());
        };
        let type_name = self.actors.get(&id).map(|a| a.type_name.clone()).unwrap_or_default();
        let saved = self.current;
        self.current = id;
        let result = self.run_method(&type_name, &msg.method, None, msg.args);
        self.current = saved;
        match result {
            Ok(v) => {
                self.promises.get_mut(&msg.promise).map(|p| p.resolved = Some(ok(v)));
                Ok(())
            }
            Err(Flow::Dead) => {
                if let Some(actor) = self.actors.get_mut(&id) {
                    actor.state = ActorState::Dead;
                    actor.mailbox.clear();
                }
                self.promises.get_mut(&msg.promise).map(|p| p.resolved = Some(err_dead()));
                Ok(())
            }
            Err(e) => Err(e),
        }
    }

    fn run_method(&mut self, type_name: &str, method: &str, recv_enum: Option<Value>, args: Vec<Value>) -> Result<Value, Flow> {
        let key = (type_name.to_string(), method.to_string());
        let m = self.methods.get(&key).cloned().ok_or_else(|| Flow::Error(format!("no method {type_name}.{method}")))?;
        let me = self.current;
        self.actors.get_mut(&me).unwrap().push_frame();
        {
            let actor = self.actors.get_mut(&me).unwrap();
            for (i, p) in m.params.iter().enumerate() {
                if let Some(arg) = args.get(i) {
                    actor.set_local(p, arg.clone());
                }
            }
            if m.has_self {
                if let Some(v) = recv_enum {
                    actor.set_local("self", v);
                }
            }
        }
        let value = self.exec_block(me, &m.body);
        self.actors.get_mut(&me).unwrap().pop_frame();
        value
    }

    fn construct(&mut self, name: &str, args: &[TypeName], inits: &[(String, Expr)]) -> Result<Value, Flow> {
        if let Some(def) = self.actors_def.get(name).cloned() {
            let id = self.next_id;
            self.next_id += 1;
            let mut subst = HashMap::new();
            for (i, p) in def.params.iter().enumerate() {
                if let Some(a) = args.get(i) {
                    subst.insert(p.clone(), a.name.clone());
                }
            }
            let mut field_names = Vec::new();
            let mut fields = Vec::new();
            for field in &def.fields {
                let kind_name = subst.get(&field.ty.name).cloned().unwrap_or_else(|| field.ty.name.clone());
                let value = if let Some(expr) = &field.default {
                    let v = self.eval(expr)?;
                    coerce_default(v, &kind_name)
                } else {
                    Value::Unit
                };
                field_names.push(field.name.clone());
                fields.push(value);
            }
            for (fname, expr) in inits {
                let value = self.eval(expr)?;
                if let Some(i) = field_names.iter().position(|n| n == fname) {
                    fields[i] = value;
                } else {
                    field_names.push(fname.clone());
                    fields.push(value);
                }
            }
            let mut actor = Actor::bare(id, name.into());
            actor.field_names = field_names;
            actor.fields = fields;
            self.actors.insert(id, actor);
            return Ok(Value::Pid(id));
        }
        Err(Flow::Error(format!("cannot construct {name}")))
    }

    fn add(&self, l: &Value, r: &Value) -> Result<Value, Flow> {
        self.arith(l, r, |a, b| a + b)
    }

    fn arith(&self, l: &Value, r: &Value, f: impl Fn(i128, i128) -> i128) -> Result<Value, Flow> {
        match (l, r) {
            (Value::Int(a), Value::Int(b)) => {
                let kind = if a.kind == IntKind::Int { b.kind } else { a.kind };
                Ok(Value::Int(IntVal::new(kind, f(a.bits, b.bits))))
            }
            (Value::Float(a), Value::Float(b)) => Ok(Value::Float(f(*a as i128, *b as i128) as f64)),
            (Value::Float(a), Value::Int(b)) => Ok(Value::Float((*a) * (b.as_i64() as f64))),
            _ => Err(Flow::Error("arithmetic needs numbers".into())),
        }
    }

    fn native(&mut self, native: Native, args: Vec<Value>) -> Result<Value, Flow> {
        match native {
            Native::Fprintf => self.fprintf(&args),
            Native::Is => {
                let err = args.first().cloned().unwrap_or(Value::Unit);
                let target = args.get(1).cloned().unwrap_or(Value::Unit);
                Ok(Value::Bool(error_is(&err, &target)))
            }
            Native::TestOnly => self.test_only(),
            Native::FailCheck => {
                self.note_check();
                Ok(Value::Unit)
            }
            Native::Ok => Ok(ok(args.into_iter().next().unwrap_or(Value::Unit))),
            Native::Err => Ok(Value::Result(Err(Box::new(args.into_iter().next().unwrap_or(Value::Unit))))),
            Native::Some => Ok(Value::Option(Some(Box::new(args.into_iter().next().unwrap_or(Value::Unit))))),
            Native::None => Ok(Value::Option(None)),
        }
    }

    fn call_func(&mut self, func: FuncObj, args: Vec<Value>) -> Result<Value, Flow> {
        let me = self.current;
        self.actors.get_mut(&me).unwrap().push_frame();
        {
            let actor = self.actors.get_mut(&me).unwrap();
            for (k, v) in &func.env {
                actor.set_local(k, v.clone());
            }
            for (i, p) in func.params.iter().enumerate() {
                if let Some(arg) = args.get(i) {
                    actor.set_local(p, arg.clone());
                }
            }
        }
        let value = self.exec_block(me, &func.body);
        self.actors.get_mut(&me).unwrap().pop_frame();
        value
    }

    fn test_only(&self) -> Result<Value, Flow> {
        if self.testing {
            Ok(Value::Unit)
        } else {
            Err(Flow::Error("check and require are legal only inside a test block".into()))
        }
    }

    fn fprintf(&mut self, args: &[Value]) -> Result<Value, Flow> {
        if args.len() < 2 {
            return Err(Flow::Error("fprintf needs a file and a format".into()));
        }
        let Value::Str(fmt) = &args[1] else {
            return Err(Flow::Error("fprintf format must be a string".into()));
        };
        let mut out = String::new();
        let mut rest = args[2..].iter();
        let bytes = fmt.as_bytes();
        let mut i = 0;
        while i < bytes.len() {
            if bytes[i] == b'%' && i + 1 < bytes.len() {
                match bytes[i + 1] {
                    b'd' => {
                        let v = rest.next().cloned().unwrap_or(Value::Unit);
                        if let Value::Int(n) = v {
                            out.push_str(&n.as_i64().to_string());
                        }
                    }
                    b's' => {
                        if let Some(Value::Str(s)) = rest.next() {
                            out.push_str(s);
                        }
                    }
                    b'f' => match rest.next() {
                        Some(Value::Float(n)) => out.push_str(&format!("{n}")),
                        Some(Value::Int(n)) => out.push_str(&n.as_i64().to_string()),
                        _ => {}
                    },
                    b'%' => out.push('%'),
                    _ => out.push(bytes[i + 1] as char),
                }
                i += 2;
            } else {
                out.push(bytes[i] as char);
                i += 1;
            }
        }
        self.stderr.push_str(&out);
        Ok(Value::Unit)
    }

    fn note_check(&mut self) {
        if let Some(actor) = self.actors.get_mut(&self.current) {
            actor.check_failed = true;
        }
    }

    fn match_pat(&self, pat: &Pattern, value: &Value) -> Option<Vec<(String, Value)>> {
        match pat {
            Pattern::Wildcard => Some(Vec::new()),
            Pattern::Bool(b) => match value {
                Value::Bool(v) if v == b => Some(Vec::new()),
                _ => None,
            },
            Pattern::Bind(n) => Some(vec![(n.clone(), value.clone())]),
            Pattern::Variant { type_name, name, fields, named } => match value {
                Value::Result(Ok(v)) if type_name == "Result" && name == "ok" => bind_one(fields, v),
                Value::Result(Err(v)) if type_name == "Result" && name == "err" => bind_one(fields, v),
                Value::Option(Some(v)) if type_name == "Option" && name == "some" => bind_one(fields, v),
                Value::Option(None) if type_name == "Option" && name == "none" => Some(Vec::new()),
                Value::Enum { type_name: got, variant, fields: vfields, .. }
                    if got == type_name && variant == name =>
                {
                    let mut binds = Vec::new();
                    if *named {
                        for pf in fields {
                            let found = vfields.iter().find(|(n, _)| n == &pf.name)?.1.clone();
                            if let Pattern::Bind(bn) = &pf.pat {
                                binds.push((bn.clone(), found));
                            } else if let Pattern::Wildcard = &pf.pat {
                            } else if let Some(more) = self.match_pat(&pf.pat, &found) {
                                binds.extend(more);
                            } else {
                                return None;
                            }
                        }
                    } else {
                        for (i, pf) in fields.iter().enumerate() {
                            let found = vfields.get(i)?.1.clone();
                            if let Pattern::Bind(bn) = &pf.pat {
                                binds.push((bn.clone(), found));
                            } else if !matches!(pf.pat, Pattern::Wildcard) {
                                if let Some(more) = self.match_pat(&pf.pat, &found) {
                                    binds.extend(more);
                                } else {
                                    return None;
                                }
                            }
                        }
                    }
                    Some(binds)
                }
                _ => None,
            },
        }
    }

    fn install_imports(&mut self, actor: u64) -> Result<(), Flow> {
        for item in &self.program.items {
            match item {
                Item::ImportNamed { names, from } => {
                    let module = std_module(from).ok_or_else(|| Flow::Error(format!("unknown module {from}")))?;
                    for name in names {
                        let v = module_get(&module, name).ok_or_else(|| Flow::Error(format!("no export {name}")))?;
                        self.actors.get_mut(&actor).unwrap().set_local(name, v);
                    }
                }
                Item::ImportModule { name, from } => {
                    let module = std_module(from).ok_or_else(|| Flow::Error(format!("unknown module {from}")))?;
                    self.actors.get_mut(&actor).unwrap().set_local(name, module);
                }
                _ => {}
            }
        }
        Ok(())
    }
}

fn bind_one(fields: &[PatternField], v: &Value) -> Option<Vec<(String, Value)>> {
    let pf = fields.first()?;
    match &pf.pat {
        Pattern::Wildcard => Some(Vec::new()),
        Pattern::Bind(n) => Some(vec![(n.clone(), v.clone())]),
        _ => None,
    }
}

fn ok(v: Value) -> Value {
    Value::Result(Ok(Box::new(v)))
}

fn err_dead() -> Value {
    Value::Result(Err(Box::new(Vm::dead_error())))
}

fn error_is(err: &Value, target: &Value) -> bool {
    match err {
        Value::Error(e) => match target {
            Value::Error(t) if e.tag == t.tag && (e.tag == "dead" || e.text == t.text) => true,
            _ => e.cause.as_ref().is_some_and(|c| error_is(c, target)),
        },
        Value::Result(Err(inner)) => error_is(inner, target),
        _ => false,
    }
}

fn values_eq(a: &Value, b: &Value) -> bool {
    match (a, b) {
        (Value::Int(a), Value::Int(b)) => a.bits == b.bits,
        (Value::Float(a), Value::Float(b)) => a == b,
        (Value::Bool(a), Value::Bool(b)) => a == b,
        (Value::Str(a), Value::Str(b)) => a == b,
        (Value::Unit, Value::Unit) => true,
        _ => false,
    }
}

fn coerce_to_ty(v: Value, ty: &TypeName) -> Value {
    match v {
        Value::Int(n) if ty.name == "f32" || ty.name == "f64" => Value::Float(n.bits as f64),
        Value::Int(n) => {
            let kind = IntKind::from_name(&ty.name).unwrap_or(n.kind);
            Value::Int(IntVal::new(kind, n.bits))
        }
        other => other,
    }
}

fn coerce_default(v: Value, type_name: &str) -> Value {
    match v {
        Value::Int(n) => {
            let kind = IntKind::from_name(type_name).unwrap_or(IntKind::Int);
            Value::Int(IntVal::new(kind, n.bits))
        }
        other => other,
    }
}

fn is_builtin_type(name: &str) -> bool {
    IntKind::from_name(name).is_some() || matches!(name, "f32" | "f64" | "bool" | "string" | "Result" | "Option")
}

fn std_module(name: &str) -> Option<Value> {
    let mut map = HashMap::new();
    match name {
        "runtime" => {
            map.insert("test_only".into(), Value::Native(Native::TestOnly));
            map.insert("fail_check".into(), Value::Native(Native::FailCheck));
        }
        "fmt" => {
            map.insert("fprintf".into(), Value::Native(Native::Fprintf));
            map.insert("stderr".into(), Value::Int(IntVal::new(IntKind::I32, 2)));
        }
        "error" => {
            map.insert("is".into(), Value::Native(Native::Is));
            map.insert("dead".into(), Vm::dead_error());
        }
        "require" => return Some(lang_module(include_str!("../std/require.lang"))),
        "check" => return Some(lang_module(include_str!("../std/check.lang"))),
        _ => return None,
    }
    Some(Value::Module(map))
}

fn lang_module(src: &str) -> Value {
    let program = crate::parse::parse(src).expect("stdlib");
    let mut env = HashMap::new();
    for item in &program.items {
        match item {
            Item::ImportNamed { names, from } => {
                let module = std_module(from).expect("stdlib import");
                for name in names {
                    env.insert(name.clone(), module_get(&module, name).expect("stdlib export"));
                }
            }
            Item::ImportModule { name, from } => {
                env.insert(name.clone(), std_module(from).expect("stdlib import"));
            }
            _ => {}
        }
    }
    let mut map = HashMap::new();
    for item in &program.items {
        if let Item::Func { name, params, body } = item {
            map.insert(
                name.clone(),
                Value::Func(Box::new(FuncObj {
                    params: params.clone(),
                    body: body.clone(),
                    env: env.clone(),
                })),
            );
        }
    }
    Value::Module(map)
}

fn module_get(module: &Value, name: &str) -> Option<Value> {
    match module {
        Value::Module(map) => map.get(name).cloned(),
        _ => None,
    }
}

impl<'a> Vm<'a> {
    fn boot(&mut self, id: u64) -> Result<(), Flow> {
        self.actors.get_mut(&id).unwrap().push_frame();
        self.install_imports(id)
    }

    fn store_self_field(&mut self, name: &str, value: Value) -> Result<(), Flow> {
        let me = self.current;
        let actor = self.actors.get_mut(&me).ok_or_else(|| Flow::Error("no actor".into()))?;
        if let Some(Value::Enum { fields, .. }) = actor.local("self") {
            if fields.iter().any(|(n, _)| n == name) {
                let Value::Enum { type_name, variant, mut fields } = actor.local("self").unwrap() else {
                    unreachable!();
                };
                if let Some(slot) = fields.iter_mut().find(|(n, _)| n == name) {
                    slot.1 = value;
                }
                actor.set_local("self", Value::Enum { type_name, variant, fields });
                return Ok(());
            }
        }
        if actor.set_field(name, value) {
            Ok(())
        } else {
            Err(Flow::Error(format!("no field {name}")))
        }
    }
}

pub fn run_source(src: &str, testing: bool) -> Result<Outcome, String> {
    let program = crate::parse::parse(src)?;
    if testing {
        let tests: Vec<Vec<Stmt>> = program
            .items
            .iter()
            .filter_map(|i| match i {
                Item::Test(b) => Some(b.clone()),
                _ => None,
            })
            .collect();
        let mut stderr = String::new();
        let mut failed = tests.is_empty() && false;
        if tests.is_empty() {
            return Ok(Outcome { status: 0, stderr });
        }
        for body in tests {
            let mut vm = Vm::new(&program, true);
            let id = vm.spawn_bare("test");
            vm.boot(id).map_err(|e| flow_err(e))?;
            match vm.exec_block(id, &body) {
                Ok(_) => {
                    if vm.actors[&id].check_failed {
                        failed = true;
                    }
                }
                Err(Flow::Dead) => failed = true,
                Err(Flow::Error(e)) => return Err(e),
            }
            stderr.push_str(&vm.stderr);
        }
        return Ok(Outcome { status: if failed { 1 } else { 0 }, stderr });
    }
    let body = program
        .items
        .iter()
        .find_map(|i| match i {
            Item::Main(b) => Some(b.clone()),
            _ => None,
        })
        .ok_or_else(|| "program has no main".to_string())?;
    let mut vm = Vm::new(&program, false);
    let id = vm.spawn_bare("main");
    vm.boot(id).map_err(|e| flow_err(e))?;
    let status = match vm.exec_block(id, &body) {
        Ok(_) => 0,
        Err(Flow::Dead) => 1,
        Err(Flow::Error(e)) => return Err(e),
    };
    Ok(Outcome { status, stderr: vm.stderr })
}

fn flow_err(flow: Flow) -> String {
    match flow {
        Flow::Dead => "actor died during import".into(),
        Flow::Error(e) => e,
    }
}
