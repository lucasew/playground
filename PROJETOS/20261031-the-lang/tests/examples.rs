use the_lang::run_source;

fn file(name: &str) -> String {
    std::fs::read_to_string(format!("examples/{name}")).unwrap()
}

#[test]
fn counter_prints_one() {
    let out = run_source(&file("01-counter.lang"), false).unwrap();
    assert_eq!(out.status, 0, "{}", out.stderr);
    assert_eq!(out.stderr, "count: 1\n");
}

#[test]
fn shape_area() {
    let src = r#"
        import { fprintf, stderr } from "fmt"
        type Shape enum {
          circle(r: f64)
          rect { w: f64, h: f64 }
        }
        Shape.area() f64 {
          match self {
            circle(r) => 3.14 * r * r
            rect { w, h } => w * h
          }
        }
        main {
          s = Shape.circle(2.0)
          fprintf(stderr, "%f\n", s.area())
          t = Shape.rect { w: 3.0, h: 4.0 }
          fprintf(stderr, "%f\n", t.area())
        }
    "#;
    let out = run_source(src, false).unwrap();
    assert_eq!(out.status, 0, "{}", out.stderr);
    assert_eq!(out.stderr, "12.56\n12\n");
}

#[test]
fn live_counter_is_ok() {
    let out = run_source(&file("03-dead.lang"), false).unwrap();
    assert_eq!(out.status, 0, "{}", out.stderr);
}

#[test]
fn dead_actor_is_error_dead() {
    let src = r#"
        import { is } from "error"
        actor Bomb[N] {
          n: N = 0
        }
        Bomb.boom() int {
          err(error.dead).unwrap()
        }
        Bomb.value() int {
          self.n
        }
        main {
          b = Bomb[int]{}
          await b.boom()
          match await b.value() {
            ok(_) => false
            err(e) => is(e, error.dead)
          }
        }
    "#;
    let out = run_source(src, false).unwrap();
    assert_eq!(out.status, 0, "{}", out.stderr);
}

#[test]
fn require_passes() {
    let out = run_source(&file("04-test.lang"), true).unwrap();
    assert_eq!(out.status, 0, "{}", out.stderr);
}

#[test]
fn require_equal_fails() {
    let src = r#"
        import require from "require"
        test {
          require.equal(1, 0)
        }
    "#;
    let out = run_source(src, true).unwrap();
    assert_eq!(out.status, 1);
}

#[test]
fn check_records_and_continues() {
    let src = r#"
        import check from "check"
        test {
          check.equal(1, 0)
          check.true(true)
        }
    "#;
    let out = run_source(src, true).unwrap();
    assert_eq!(out.status, 1, "{}", out.stderr);
}

#[test]
fn star_import_is_rejected() {
    let err = run_source("import * from \"fmt\"\nmain { 1 }\n", false).unwrap_err();
    assert!(err.contains("star"), "{err}");
}

#[test]
fn int_wraps() {
    let src = r#"
        import { fprintf, stderr } from "fmt"
        actor Wrap[N] {
          n: N = 2147483647
        }
        Wrap.next() N {
          self.n += 1
          self.n
        }
        main {
          w = Wrap[i32]{}
          fprintf(stderr, "%d\n", (await w.next()).unwrap())
        }
    "#;
    let out = run_source(src, false).unwrap();
    assert_eq!(out.status, 0, "{}", out.stderr);
    assert_eq!(out.stderr, "-2147483648\n");
}
