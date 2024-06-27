// tests/fibonacci_tests.rs

extern crate rust;

use rust::expensive::fibonacci;

#[test]
fn test_fibonacci() {
    assert_eq!(fibonacci(0), 1);
    assert_eq!(fibonacci(1), 1);
    assert_eq!(fibonacci(2), 2);
    assert_eq!(fibonacci(3), 3);
    assert_eq!(fibonacci(4), 5);
    assert_eq!(fibonacci(5), 8);
    assert_eq!(fibonacci(10), 89);
    assert_eq!(fibonacci(20), 10946);
}
