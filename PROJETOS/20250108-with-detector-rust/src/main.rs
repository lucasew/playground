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
    if let Some(invalid_with) = find_invalid_withs(ast.syntax()) {
        println!("invalid {:?} {:?}", invalid_with, invalid_with.to_string())
    }
}

fn find_invalid_withs(syntax: SyntaxNode<NixLanguage>) -> Option<SyntaxNode<NixLanguage>> {
    syntax
        .descendants()
        .filter(|node| node.kind() == rnix::SyntaxKind::NODE_WITH)
        .filter(|node| {
            node.descendants()
                .map(|child| {
                    if child == *node {
                        return None;
                    }
                    let node_if_invalid = match child.kind() {
                        SyntaxKind::NODE_WITH => Some(node),
                        SyntaxKind::NODE_LET_IN => Some(node),
                        SyntaxKind::NODE_ATTR_SET => Some(node),
                        _ => None,
                    };
                    println!(
                        "validate with={:?} subexpr={:?} invalid={:?}",
                        node.to_string(),
                        child.to_string(),
                        node_if_invalid
                    );
                    node_if_invalid
                })
                .any(|cond| cond != None)
        })
        .take(1)
        .last()
}
