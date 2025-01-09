use std::{fs, path::PathBuf};

use rnix::{NixLanguage, SyntaxKind};
use rowan::SyntaxNode;
use structopt::StructOpt;

#[derive(Debug, StructOpt)]
struct Args {
    #[structopt(parse(from_os_str))]
    file: PathBuf,
}

fn main() {
    let args = Args::from_args();
    let content = fs::read_to_string(&args.file).unwrap();
    let ast = rnix::Root::parse(&content);
    for error in ast.errors() {
        println!("error: {}", error)
    }
    for w in find_invalid_withs(ast.syntax()) {
        println!("invalid {:?} {:?}", w, w.to_string())
    }
}

fn find_invalid_withs(
    syntax: SyntaxNode<NixLanguage>,
) -> impl Iterator<Item = SyntaxNode<NixLanguage>> {
    syntax
        .descendants()
        .filter(|node| node.kind() == rnix::SyntaxKind::NODE_WITH)
        .filter(|node| {
            node.descendants()
                .map(|child| {
                    if child == *node {
                        return false;
                    }
                    let is_invalid = match child.kind() {
                        SyntaxKind::NODE_WITH => true,
                        SyntaxKind::NODE_LET_IN => true,
                        SyntaxKind::NODE_ATTR_SET => true,
                        _ => false,
                    };
                    println!(
                        "validate with={:?} subexpr={:?} invalid={:?}",
                        node.to_string(),
                        child.to_string(),
                        is_invalid
                    );
                    is_invalid
                })
                .any(|cond| cond)
        })
}
