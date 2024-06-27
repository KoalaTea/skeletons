#[macro_use]
extern crate criterion;
use criterion::{black_box, Criterion};

use pprof::criterion::{Output, PProfProfiler};

use rust::expensive::fibonacci;

fn bench(c: &mut Criterion) {
    c.bench_function("Fibonacci", |b| b.iter(|| fibonacci(black_box(20))));
}

criterion_group! {
    name = benches;
    config = Criterion::default().with_profiler(PProfProfiler::new(100, Output::Flamegraph(None)));
    targets = bench
}
criterion_main!(benches);
