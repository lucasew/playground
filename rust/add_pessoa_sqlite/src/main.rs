use rusqlite::{Connection};
use std::io::BufRead;

fn read_line(buf: &mut String) -> std::io::Result<usize> {
    let ret = std::io::stdin().lock().read_line(buf);
    buf.pop().unwrap(); // remove the newline in the end of the string
    ret
}

fn main() {
    let conn = Connection::open("database.db").unwrap();
    conn.execute_batch("
    begin;
    create table if not exists pessoa (
        id integer primary key,
        name text not null,
        email text not null
    );
    commit;
    ").unwrap();

    // alter table pessoa add column idade integer not null default 20;
    // alter table pessoa add column peso real not null default 100;
    let mut name = String::new();
    read_line(&mut name).unwrap();
    let mut email = String::new();
    read_line(&mut email).unwrap();
    conn.execute("insert into pessoa (name, email, idade) values (?1, ?2, ?3)", &[&name, &email, "douglas"]).unwrap();
}