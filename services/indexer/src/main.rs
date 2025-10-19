mod sink;
use std::{
    fs::{self, File},
    io::Cursor,
};

use html5ever::parse_document;
use html5ever::tendril::TendrilSink;

use crate::sink::text_sink::TextSink;
use std::time::Instant;

fn timed<F: FnOnce() -> u32>(f: F) -> (u32, std::time::Duration) {
    let start = Instant::now();
    let c = f();
    (c, start.elapsed())
}

fn main() {
    let sink = TextSink::new();
    let f1 = fs::read("1.html").unwrap();
    let f2 = fs::read("2.html").unwrap();
    let f3 = fs::read("3.html").unwrap();
    let size1 = f1.len() as f64 / (1024.0 * 1024.0);
    let size2 = f2.len() as f64 / 1024.0;
    let size3 = f3.len() as f64 / 1024.0;

    let (c1, dt1) = timed(|| {
        parse_document(sink.clone(), Default::default())
            .from_utf8()
            .read_from(&mut Cursor::new(f1))
            .unwrap()
            .values()
            .sum()
    });
    let (c2, dt2) = timed(|| {
        parse_document(sink.clone(), Default::default())
            .from_utf8()
            .read_from(&mut Cursor::new(f2))
            .unwrap()
            .values()
            .sum()
    });
    let (c3, dt3) = timed(|| {
        parse_document(sink, Default::default())
            .from_utf8()
            .read_from(&mut Cursor::new(f3))
            .unwrap()
            .values()
            .sum()
    });

    println!(
        "1.html: {:.1} MB, {:.2} s, Words Count: {}",
        size1,
        dt1.as_secs() as f64,
        c1
    );
    println!(
        "2.html: {:.1} KiB, {:.2} ms, Words Count: {}",
        size2,
        dt2.as_millis() as f64,
        c2
    );
    println!(
        "3.html: {:.1} KiB, {:.2} ms, Words Count: {}",
        size3,
        dt3.as_millis() as f64,
        c3
    );
}
