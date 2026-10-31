#[derive(Clone, Debug)]
pub struct Program {
    pub items: Vec<Item>,
}

#[derive(Clone, Debug)]
pub enum Item {
    ImportNamed { names: Vec<String>, from: String },
    ImportModule { name: String, from: String },
    Actor {
        name: String,
        params: Vec<String>,
        fields: Vec<Field>,
    },
    Struct {
        name: String,
        fields: Vec<Field>,
    },
    Enum {
        name: String,
        variants: Vec<Variant>,
    },
    Method {
        type_name: String,
        name: String,
        params: Vec<String>,
        body: Vec<Stmt>,
    },
    Main(Vec<Stmt>),
    Test(Vec<Stmt>),
}

#[derive(Clone, Debug)]
pub struct Field {
    pub name: String,
    pub ty: TypeName,
    pub default: Option<Expr>,
}

#[derive(Clone, Debug)]
pub struct Variant {
    pub name: String,
    pub fields: Vec<VariantField>,
    pub named: bool,
}

#[derive(Clone, Debug)]
pub struct VariantField {
    pub name: String,
    pub ty: Option<TypeName>,
}

#[derive(Clone, Debug)]
pub struct TypeName {
    pub name: String,
    pub args: Vec<TypeName>,
}

#[derive(Clone, Debug)]
pub enum Stmt {
    Assign {
        target: Expr,
        op: AssignOp,
        value: Expr,
    },
    TestBlock(Vec<Stmt>),
    Expr(Expr),
}

#[derive(Clone, Copy, Debug)]
pub enum AssignOp {
    Eq,
    PlusEq,
}

#[derive(Clone, Debug)]
pub enum Expr {
    Name(String),
    Int(i64),
    Float(f64),
    Str(String),
    Bool(bool),
    Call {
        callee: Box<Expr>,
        args: Vec<Expr>,
    },
    Field {
        base: Box<Expr>,
        name: String,
    },
    TypeApp {
        base: Box<Expr>,
        args: Vec<TypeName>,
    },
    Construct {
        name: String,
        args: Vec<TypeName>,
        fields: Vec<(String, Expr)>,
    },
    Await(Box<Expr>),
    UnaryMinus(Box<Expr>),
    Binary {
        op: BinOp,
        left: Box<Expr>,
        right: Box<Expr>,
    },
    Match {
        expr: Box<Expr>,
        arms: Vec<Arm>,
    },
}

#[derive(Clone, Copy, Debug)]
pub enum BinOp {
    Add,
    Sub,
    Mul,
    Div,
    Eq,
}

#[derive(Clone, Debug)]
pub struct Arm {
    pub pattern: Pattern,
    pub body: Expr,
}

#[derive(Clone, Debug)]
pub enum Pattern {
    Wildcard,
    Bind(String),
    Variant {
        name: String,
        fields: Vec<PatternField>,
        named: bool,
    },
}

#[derive(Clone, Debug)]
pub struct PatternField {
    pub name: String,
    pub pat: Pattern,
}
