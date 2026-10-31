use crate::ast::*;

#[derive(Clone, Debug)]
enum Tok {
    Ident(String),
    Int(i64),
    Float(f64),
    Str(String),
    LBrace,
    RBrace,
    LParen,
    RParen,
    LBrack,
    RBrack,
    Comma,
    Dot,
    Colon,
    Eq,
    EqEq,
    PlusEq,
    Plus,
    Minus,
    Star,
    Slash,
    FatArrow,
    Semi,
    Eof,
}

struct Lexer<'a> {
    src: &'a [u8],
    i: usize,
    line: usize,
    /// Previous token can end a statement. A newline then yields `Semi`.
    can_semi: bool,
}

impl<'a> Lexer<'a> {
    fn new(src: &'a str) -> Self {
        Self { src: src.as_bytes(), i: 0, line: 1, can_semi: false }
    }

    fn peek(&self) -> Option<u8> {
        self.src.get(self.i).copied()
    }

    fn bump(&mut self) -> Option<u8> {
        let c = self.peek()?;
        self.i += 1;
        if c == b'\n' {
            self.line += 1;
        }
        Some(c)
    }

    fn skip(&mut self) {
        loop {
            while self.peek().is_some_and(|c| c.is_ascii_whitespace()) {
                self.bump();
            }
            if self.peek() == Some(b'/') && self.src.get(self.i + 1) == Some(&b'/') {
                while self.peek().is_some_and(|c| c != b'\n') {
                    self.bump();
                }
                continue;
            }
            break;
        }
    }

    fn next(&mut self) -> Result<(Tok, usize), String> {
        let line_before = self.line;
        self.skip();
        if self.line > line_before && self.can_semi {
            self.can_semi = false;
            return Ok((Tok::Semi, line_before));
        }
        let line = self.line;
        let Some(c) = self.bump() else {
            return Ok((Tok::Eof, line));
        };
        let tok = match c {
            b'{' => Tok::LBrace,
            b'}' => Tok::RBrace,
            b'(' => Tok::LParen,
            b')' => Tok::RParen,
            b'[' => Tok::LBrack,
            b']' => Tok::RBrack,
            b',' => Tok::Comma,
            b'.' => Tok::Dot,
            b':' => Tok::Colon,
            b'+' => {
                if self.peek() == Some(b'=') {
                    self.bump();
                    Tok::PlusEq
                } else {
                    Tok::Plus
                }
            }
            b'-' => Tok::Minus,
            b'*' => Tok::Star,
            b'/' => Tok::Slash,
            b'=' => {
                if self.peek() == Some(b'=') {
                    self.bump();
                    Tok::EqEq
                } else if self.peek() == Some(b'>') {
                    self.bump();
                    Tok::FatArrow
                } else {
                    Tok::Eq
                }
            }
            b'"' => Tok::Str(self.string()?),
            b'0'..=b'9' => self.number(c)?,
            b'a'..=b'z' | b'A'..=b'Z' | b'_' => Tok::Ident(self.ident(c)),
            _ => return Err(format!("line {line}: unexpected character")),
        };
        self.can_semi = matches!(
            tok,
            Tok::Ident(_)
                | Tok::Int(_)
                | Tok::Float(_)
                | Tok::Str(_)
                | Tok::RParen
                | Tok::RBrack
                | Tok::RBrace
        );
        Ok((tok, line))
    }

    fn ident(&mut self, first: u8) -> String {
        let mut s = String::new();
        s.push(first as char);
        while self.peek().is_some_and(|c| c.is_ascii_alphanumeric() || c == b'_') {
            s.push(self.bump().unwrap() as char);
        }
        s
    }

    fn number(&mut self, first: u8) -> Result<Tok, String> {
        let mut s = String::new();
        s.push(first as char);
        while self.peek().is_some_and(|c| c.is_ascii_digit() || c == b'_') {
            let c = self.bump().unwrap();
            if c != b'_' {
                s.push(c as char);
            }
        }
        if self.peek() == Some(b'.') && self.src.get(self.i + 1).is_some_and(|c| c.is_ascii_digit()) {
            s.push(self.bump().unwrap() as char);
            while self.peek().is_some_and(|c| c.is_ascii_digit() || c == b'_') {
                let c = self.bump().unwrap();
                if c != b'_' {
                    s.push(c as char);
                }
            }
            let n: f64 = s.parse().map_err(|_| format!("bad float {s}"))?;
            return Ok(Tok::Float(n));
        }
        let n: i64 = s.parse().map_err(|_| format!("bad integer {s}"))?;
        Ok(Tok::Int(n))
    }

    fn string(&mut self) -> Result<String, String> {
        let mut s = String::new();
        loop {
            match self.bump() {
                None => return Err("unterminated string".into()),
                Some(b'"') => return Ok(s),
                Some(b'\\') => {
                    let e = self.bump().ok_or("bad escape")?;
                    s.push(match e {
                        b'n' => '\n',
                        b't' => '\t',
                        b'\\' => '\\',
                        b'"' => '"',
                        other => other as char,
                    });
                }
                Some(c) => s.push(c as char),
            }
        }
    }
}

struct Parser {
    toks: Vec<(Tok, usize)>,
    i: usize,
    /// When set, `{` ends the expression. `match` uses this so the brace is the arm list.
    brace_ends: bool,
}

impl Parser {
    fn line(&self) -> usize {
        self.toks.get(self.i).map(|t| t.1).unwrap_or(1)
    }

    fn err(&self, msg: &str) -> String {
        format!("line {}: {msg}", self.line())
    }

    fn peek(&self) -> &Tok {
        self.toks.get(self.i).map(|t| &t.0).unwrap_or(&Tok::Eof)
    }

    fn bump(&mut self) -> Tok {
        let tok = self.toks.get(self.i).map(|t| t.0.clone()).unwrap_or(Tok::Eof);
        if self.i < self.toks.len() {
            self.i += 1;
        }
        tok
    }

    fn eat_ident(&mut self) -> Result<String, String> {
        match self.bump() {
            Tok::Ident(s) => Ok(s),
            _ => Err(self.err("expected a name")),
        }
    }

    fn want(&mut self, tok: &Tok) -> Result<(), String> {
        let got = self.bump();
        if std::mem::discriminant(&got) == std::mem::discriminant(tok) {
            Ok(())
        } else {
            Err(self.err("unexpected token"))
        }
    }

    fn skip_semi(&mut self) {
        while matches!(self.peek(), Tok::Semi) {
            self.bump();
        }
    }

    fn parse(mut self) -> Result<Program, String> {
        let mut items = Vec::new();
        while !matches!(self.peek(), Tok::Eof) {
            self.skip_semi();
            if matches!(self.peek(), Tok::Eof) {
                break;
            }
            items.push(self.item()?);
        }
        Ok(Program { items })
    }

    fn item(&mut self) -> Result<Item, String> {
        let Tok::Ident(k) = self.peek().clone() else {
            return Err(self.err("expected a declaration"));
        };
        match k.as_str() {
            "import" => self.import(),
            "actor" => self.actor(),
            "type" => self.type_decl(),
            "main" => {
                self.bump();
                Ok(Item::Main(self.block()?))
            }
            "test" => {
                self.bump();
                Ok(Item::Test(self.block()?))
            }
            _ => self.method(),
        }
    }

    fn import(&mut self) -> Result<Item, String> {
        self.bump();
        if matches!(self.peek(), Tok::LBrace) {
            self.bump();
            let mut names = vec![self.eat_ident()?];
            while matches!(self.peek(), Tok::Comma) {
                self.bump();
                if matches!(self.peek(), Tok::RBrace) {
                    break;
                }
                names.push(self.eat_ident()?);
            }
            self.want(&Tok::RBrace)?;
            self.kw("from")?;
            let from = self.string_lit()?;
            return Ok(Item::ImportNamed { names, from });
        }
        if matches!(self.peek(), Tok::Star) {
            return Err(self.err("star import is rejected"));
        }
        let name = self.eat_ident()?;
        self.kw("from")?;
        let from = self.string_lit()?;
        Ok(Item::ImportModule { name, from })
    }

    fn kw(&mut self, word: &str) -> Result<(), String> {
        match self.bump() {
            Tok::Ident(s) if s == word => Ok(()),
            _ => Err(self.err(&format!("expected {word}"))),
        }
    }

    fn string_lit(&mut self) -> Result<String, String> {
        match self.bump() {
            Tok::Str(s) => Ok(s),
            _ => Err(self.err("expected a string")),
        }
    }

    fn actor(&mut self) -> Result<Item, String> {
        self.bump();
        let name = self.eat_ident()?;
        let params = self.type_params()?;
        self.want(&Tok::LBrace)?;
        let mut fields = Vec::new();
        while !matches!(self.peek(), Tok::RBrace | Tok::Eof) {
            self.skip_semi();
            if matches!(self.peek(), Tok::RBrace | Tok::Eof) {
                break;
            }
            fields.push(self.field()?);
        }
        self.want(&Tok::RBrace)?;
        Ok(Item::Actor { name, params, fields })
    }

    fn type_params(&mut self) -> Result<Vec<String>, String> {
        if !matches!(self.peek(), Tok::LBrack) {
            return Ok(Vec::new());
        }
        self.bump();
        let mut params = vec![self.eat_ident()?];
        while matches!(self.peek(), Tok::Comma) {
            self.bump();
            params.push(self.eat_ident()?);
        }
        self.want(&Tok::RBrack)?;
        Ok(params)
    }

    fn field(&mut self) -> Result<Field, String> {
        let name = self.eat_ident()?;
        self.want(&Tok::Colon)?;
        let ty = self.type_name()?;
        let default = if matches!(self.peek(), Tok::Eq) {
            self.bump();
            Some(self.expr()?)
        } else {
            None
        };
        Ok(Field { name, ty, default })
    }

    fn type_name(&mut self) -> Result<TypeName, String> {
        let name = self.eat_ident()?;
        let mut args = Vec::new();
        if matches!(self.peek(), Tok::LBrack) {
            self.bump();
            args.push(self.type_name()?);
            while matches!(self.peek(), Tok::Comma) {
                self.bump();
                args.push(self.type_name()?);
            }
            self.want(&Tok::RBrack)?;
        }
        Ok(TypeName { name, args })
    }

    fn type_decl(&mut self) -> Result<Item, String> {
        self.bump();
        let name = self.eat_ident()?;
        let Tok::Ident(kind) = self.bump() else {
            return Err(self.err("expected struct or enum"));
        };
        match kind.as_str() {
            "struct" => {
                self.want(&Tok::LBrace)?;
                let mut fields = Vec::new();
                while !matches!(self.peek(), Tok::RBrace | Tok::Eof) {
                    self.skip_semi();
                    if matches!(self.peek(), Tok::RBrace | Tok::Eof) {
                        break;
                    }
                    fields.push(self.field()?);
                }
                self.want(&Tok::RBrace)?;
                Ok(Item::Struct { name, fields })
            }
            "enum" => {
                self.want(&Tok::LBrace)?;
                let mut variants = Vec::new();
                while !matches!(self.peek(), Tok::RBrace | Tok::Eof) {
                    self.skip_semi();
                    if matches!(self.peek(), Tok::RBrace | Tok::Eof) {
                        break;
                    }
                    variants.push(self.variant()?);
                }
                self.want(&Tok::RBrace)?;
                Ok(Item::Enum { name, variants })
            }
            _ => Err(self.err("expected struct or enum")),
        }
    }

    fn variant(&mut self) -> Result<Variant, String> {
        let name = self.eat_ident()?;
        if matches!(self.peek(), Tok::LParen) {
            self.bump();
            let mut fields = Vec::new();
            while !matches!(self.peek(), Tok::RParen | Tok::Eof) {
                self.skip_semi();
                if matches!(self.peek(), Tok::Comma) {
                    self.bump();
                    continue;
                }
                if matches!(self.peek(), Tok::RParen | Tok::Eof) {
                    break;
                }
                fields.push(self.variant_field()?);
            }
            self.want(&Tok::RParen)?;
            return Ok(Variant { name, fields, named: false });
        }
        if matches!(self.peek(), Tok::LBrace) {
            self.bump();
            let mut fields = Vec::new();
            while !matches!(self.peek(), Tok::RBrace | Tok::Eof) {
                self.skip_semi();
                if matches!(self.peek(), Tok::Comma) {
                    self.bump();
                    continue;
                }
                if matches!(self.peek(), Tok::RBrace | Tok::Eof) {
                    break;
                }
                let fname = self.eat_ident()?;
                self.want(&Tok::Colon)?;
                let ty = self.type_name()?;
                fields.push(VariantField { name: fname, ty: Some(ty) });
            }
            self.want(&Tok::RBrace)?;
            return Ok(Variant { name, fields, named: true });
        }
        Ok(Variant { name, fields: Vec::new(), named: false })
    }

    fn variant_field(&mut self) -> Result<VariantField, String> {
        let name = self.eat_ident()?;
        let ty = if matches!(self.peek(), Tok::Colon) {
            self.bump();
            Some(self.type_name()?)
        } else {
            None
        };
        Ok(VariantField { name, ty })
    }

    fn method(&mut self) -> Result<Item, String> {
        let type_name = self.eat_ident()?;
        self.want(&Tok::Dot)?;
        let name = self.eat_ident()?;
        self.want(&Tok::LParen)?;
        let mut params = Vec::new();
        if !matches!(self.peek(), Tok::RParen) {
            params.push(self.eat_ident()?);
            while matches!(self.peek(), Tok::Comma) {
                self.bump();
                params.push(self.eat_ident()?);
            }
        }
        self.want(&Tok::RParen)?;
        if matches!(self.peek(), Tok::Ident(_)) && !matches!(self.peek(), Tok::LBrace) {
            let _ret = self.type_name()?;
        }
        let body = self.block()?;
        Ok(Item::Method { type_name, name, params, body })
    }

    fn block(&mut self) -> Result<Vec<Stmt>, String> {
        self.want(&Tok::LBrace)?;
        let mut stmts = Vec::new();
        while !matches!(self.peek(), Tok::RBrace | Tok::Eof) {
            self.skip_semi();
            if matches!(self.peek(), Tok::RBrace | Tok::Eof) {
                break;
            }
            stmts.push(self.stmt()?);
        }
        self.want(&Tok::RBrace)?;
        Ok(stmts)
    }

    fn stmt(&mut self) -> Result<Stmt, String> {
        if let Tok::Ident(k) = self.peek().clone() {
            if k == "test" {
                self.bump();
                return Ok(Stmt::TestBlock(self.block()?));
            }
        }
        let expr = self.expr()?;
        if matches!(self.peek(), Tok::Eq | Tok::PlusEq) {
            let op = if matches!(self.bump(), Tok::PlusEq) { AssignOp::PlusEq } else { AssignOp::Eq };
            let value = self.expr()?;
            return Ok(Stmt::Assign { target: expr, op, value });
        }
        Ok(Stmt::Expr(expr))
    }

    fn expr(&mut self) -> Result<Expr, String> {
        self.equality()
    }

    fn equality(&mut self) -> Result<Expr, String> {
        let mut left = self.sum()?;
        if matches!(self.peek(), Tok::EqEq) {
            self.bump();
            let right = self.sum()?;
            left = Expr::Binary { op: BinOp::Eq, left: Box::new(left), right: Box::new(right) };
        }
        Ok(left)
    }

    fn sum(&mut self) -> Result<Expr, String> {
        let mut left = self.product()?;
        while matches!(self.peek(), Tok::Plus | Tok::Minus) {
            let op = if matches!(self.bump(), Tok::Plus) { BinOp::Add } else { BinOp::Sub };
            let right = self.product()?;
            left = Expr::Binary { op, left: Box::new(left), right: Box::new(right) };
        }
        Ok(left)
    }

    fn product(&mut self) -> Result<Expr, String> {
        let mut left = self.unary()?;
        while matches!(self.peek(), Tok::Star | Tok::Slash) {
            let op = if matches!(self.bump(), Tok::Star) { BinOp::Mul } else { BinOp::Div };
            let right = self.unary()?;
            left = Expr::Binary { op, left: Box::new(left), right: Box::new(right) };
        }
        Ok(left)
    }

    fn unary(&mut self) -> Result<Expr, String> {
        if let Tok::Ident(k) = self.peek().clone() {
            if k == "await" {
                self.bump();
                return Ok(Expr::Await(Box::new(self.unary()?)));
            }
        }
        if matches!(self.peek(), Tok::Minus) {
            self.bump();
            return Ok(Expr::UnaryMinus(Box::new(self.unary()?)));
        }
        if let Tok::Ident(k) = self.peek().clone() {
            if k == "match" {
                return self.match_expr();
            }
        }
        self.postfix()
    }

    fn match_expr(&mut self) -> Result<Expr, String> {
        self.bump();
        let saved = self.brace_ends;
        self.brace_ends = true;
        let expr = self.expr()?;
        self.brace_ends = saved;
        self.want(&Tok::LBrace)?;
        let mut arms = Vec::new();
        while !matches!(self.peek(), Tok::RBrace | Tok::Eof) {
            self.skip_semi();
            if matches!(self.peek(), Tok::RBrace | Tok::Eof) {
                break;
            }
            let pattern = self.pattern()?;
            self.want(&Tok::FatArrow)?;
            let body = self.expr()?;
            arms.push(Arm { pattern, body });
        }
        self.want(&Tok::RBrace)?;
        Ok(Expr::Match { expr: Box::new(expr), arms })
    }

    fn pattern(&mut self) -> Result<Pattern, String> {
        if let Tok::Ident(name) = self.peek().clone() {
            if name == "_" {
                self.bump();
                return Ok(Pattern::Wildcard);
            }
            self.bump();
            if matches!(self.peek(), Tok::LParen) {
                self.bump();
                let mut fields = Vec::new();
                if !matches!(self.peek(), Tok::RParen) {
                    fields.push(self.pat_field_pos(0)?);
                    let mut i = 1;
                    while matches!(self.peek(), Tok::Comma) {
                        self.bump();
                        fields.push(self.pat_field_pos(i)?);
                        i += 1;
                    }
                }
                self.want(&Tok::RParen)?;
                return Ok(Pattern::Variant { name, fields, named: false });
            }
            if matches!(self.peek(), Tok::LBrace) {
                self.bump();
                let mut fields = Vec::new();
                while !matches!(self.peek(), Tok::RBrace | Tok::Eof) {
                    self.skip_semi();
                    if matches!(self.peek(), Tok::Comma) {
                        self.bump();
                        continue;
                    }
                    if matches!(self.peek(), Tok::RBrace | Tok::Eof) {
                        break;
                    }
                    let fname = self.eat_ident()?;
                    let pat = if matches!(self.peek(), Tok::Colon) {
                        self.bump();
                        self.pattern()?
                    } else {
                        Pattern::Bind(fname.clone())
                    };
                    fields.push(PatternField { name: fname, pat });
                }
                self.want(&Tok::RBrace)?;
                return Ok(Pattern::Variant { name, fields, named: true });
            }
            return Ok(Pattern::Bind(name));
        }
        Err(self.err("expected a pattern"))
    }

    fn pat_field_pos(&mut self, index: usize) -> Result<PatternField, String> {
        let pat = self.pattern()?;
        let name = match &pat {
            Pattern::Bind(n) => n.clone(),
            Pattern::Wildcard => format!("_{index}"),
            Pattern::Variant { name, .. } => name.clone(),
        };
        Ok(PatternField { name, pat })
    }

    fn postfix(&mut self) -> Result<Expr, String> {
        let mut expr = self.primary()?;
        loop {
            if matches!(self.peek(), Tok::LBrack) {
                self.bump();
                let mut args = vec![self.type_name()?];
                while matches!(self.peek(), Tok::Comma) {
                    self.bump();
                    args.push(self.type_name()?);
                }
                self.want(&Tok::RBrack)?;
                if matches!(self.peek(), Tok::LBrace) && !self.brace_ends {
                    let Expr::Name(name) = expr else {
                        return Err(self.err("construct needs a type name"));
                    };
                    let fields = self.field_inits()?;
                    expr = Expr::Construct { name, args, fields };
                    continue;
                }
                expr = Expr::TypeApp { base: Box::new(expr), args };
                continue;
            }
            if matches!(self.peek(), Tok::LParen) {
                self.bump();
                let saved = self.brace_ends;
                self.brace_ends = false;
                let mut args = Vec::new();
                loop {
                    self.skip_semi();
                    if matches!(self.peek(), Tok::RParen | Tok::Eof) {
                        break;
                    }
                    if matches!(self.peek(), Tok::Comma) {
                        self.bump();
                        continue;
                    }
                    args.push(self.expr()?);
                }
                self.brace_ends = saved;
                self.want(&Tok::RParen)?;
                expr = Expr::Call { callee: Box::new(expr), args };
                continue;
            }
            if matches!(self.peek(), Tok::LBrace) && !self.brace_ends {
                if let Expr::Name(name) = &expr {
                    let name = name.clone();
                    let fields = self.field_inits()?;
                    expr = Expr::Construct { name, args: Vec::new(), fields };
                    continue;
                }
                if let Expr::Field { base, name } = expr {
                    let fields = self.field_inits()?;
                    expr = Expr::NamedConstruct { base, variant: name, fields };
                    continue;
                }
            }
            if matches!(self.peek(), Tok::Dot) {
                self.bump();
                let name = self.eat_ident()?;
                expr = Expr::Field { base: Box::new(expr), name };
                continue;
            }
            break;
        }
        Ok(expr)
    }

    fn field_inits(&mut self) -> Result<Vec<(String, Expr)>, String> {
        self.want(&Tok::LBrace)?;
        let mut fields = Vec::new();
        while !matches!(self.peek(), Tok::RBrace | Tok::Eof) {
            self.skip_semi();
            if matches!(self.peek(), Tok::Comma) {
                self.bump();
                continue;
            }
            if matches!(self.peek(), Tok::RBrace | Tok::Eof) {
                break;
            }
            let name = self.eat_ident()?;
            self.want(&Tok::Colon)?;
            let value = self.expr()?;
            fields.push((name, value));
        }
        self.want(&Tok::RBrace)?;
        Ok(fields)
    }

    fn primary(&mut self) -> Result<Expr, String> {
        match self.bump() {
            Tok::Ident(s) if s == "true" => Ok(Expr::Bool(true)),
            Tok::Ident(s) if s == "false" => Ok(Expr::Bool(false)),
            Tok::Ident(s) => Ok(Expr::Name(s)),
            Tok::Int(n) => Ok(Expr::Int(n)),
            Tok::Float(n) => Ok(Expr::Float(n)),
            Tok::Str(s) => Ok(Expr::Str(s)),
            Tok::LParen => {
                let saved = self.brace_ends;
                self.brace_ends = false;
                let e = self.expr()?;
                self.brace_ends = saved;
                self.want(&Tok::RParen)?;
                Ok(e)
            }
            _ => Err(self.err("expected an expression")),
        }
    }
}

pub fn parse(src: &str) -> Result<Program, String> {
    let mut lexer = Lexer::new(src);
    let mut toks = Vec::new();
    loop {
        let (tok, line) = lexer.next()?;
        let done = matches!(tok, Tok::Eof);
        toks.push((tok, line));
        if done {
            break;
        }
    }
    Parser { toks, i: 0, brace_ends: false }.parse()
}
