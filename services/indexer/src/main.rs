mod sink;
use std::io::Cursor;

use html5ever::parse_document;
use html5ever::tendril::TendrilSink;

use crate::sink::text_sink::{Node, TextSink};

fn main() {
    let html = r#"
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Example</title>
</head>
<body>
    <header>
        <h1>Heading</h1>
    </header>
    <main>
        <p>Paragraph.</p>
        <ul>
            <li>Item 1</li>
            <li>Item 2</li>
        </ul>
    </main>
    <footer>
        <p>Footer.</p>
    </footer>
</body>
</html>
"#;

    let sink = TextSink::new();
    let texts = parse_document(sink, Default::default())
        .from_utf8()
        .read_from(&mut Cursor::new(html.as_bytes()))
        .unwrap(); // This already returns Vec<String>

    let mut txt: Vec<String> = vec![];
    for t in texts {
        match t.as_ref() {
            Node::Text(x) => {
                txt.push(x.clone());
            }
            Node::Element(_) => {}
        }
    }

    // Example check
    assert_eq!(
        txt,
        vec!["Heading", "Paragraph.", "Item 1", "Item 2", "Footer."]
    );
}
