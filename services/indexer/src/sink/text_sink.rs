use std::{borrow::Cow, cell::RefCell, collections::HashMap, rc::Rc};

use html5ever::{
    interface::{ElementFlags, NodeOrText, QuirksMode, TreeSink},
    local_name, ns,
    tendril::StrTendril,
    Attribute, QualName,
};

#[derive(Debug, Clone)]
pub enum Node {
    Element(QualName),
}

type Handle = Rc<Node>;

#[derive(Debug, Clone)]
pub struct TextSink {
    pub texts: RefCell<Vec<StrTendril>>,
    pub doc: Handle,
}

impl TextSink {
    pub fn new() -> Self {
        Self {
            texts: RefCell::new(Vec::new()),
            doc: Handle::new(Node::Element(QualName::new(None, ns!(), local_name!("")))),
        }
    }
}
impl TreeSink for TextSink {
    type Output = HashMap<String, u32>;
    type Handle = Handle;
    type ElemName<'a> = &'a QualName;

    fn finish(self) -> Self::Output {
        let mut out: HashMap<String, u32> = HashMap::new();
        for text in self.texts.into_inner() {
            for word in text.to_lowercase().split_whitespace().filter_map(|w| {
                let trimed_word = w.trim_matches(|c: char| {
                    c.is_whitespace() || matches!(c, '.' | ',' | ':' | '/' | ';' | '"')
                });
                if trimed_word.is_empty() || trimed_word.parse::<u32>().is_ok() {
                    return None;
                }
                Some(trimed_word.to_string())
            }) {
                *out.entry(word).or_insert(0) += 1;
            }
        }
        out
    }

    fn parse_error(&self, _msg: Cow<'static, str>) {
        // eprintln!("Parse error: {msg}");
    }

    fn get_document(&self) -> Self::Handle {
        self.doc.clone()
    }

    fn elem_name<'a>(&'a self, target: &'a Self::Handle) -> Self::ElemName<'a> {
        let Node::Element(name) = target.as_ref();
        name
    }

    fn create_comment(&self, _text: StrTendril) -> Self::Handle {
        Handle::new(Node::Element(QualName::new(None, ns!(), local_name!(""))))
    }

    fn create_element(&self, name: QualName, _: Vec<Attribute>, _: ElementFlags) -> Self::Handle {
        Handle::new(Node::Element(name))
    }

    fn create_pi(&self, _target: StrTendril, _data: StrTendril) -> Self::Handle {
        Handle::new(Node::Element(QualName::new(None, ns!(), local_name!("pi"))))
    }

    fn append(&self, parent: &Self::Handle, child: NodeOrText<Self::Handle>) {
        if let NodeOrText::AppendText(t) = child {
            let Node::Element(name) = parent.as_ref();
            let local = name.local.as_ref();
            if local == "script" || local == "style" {
                return; //ignore script and style elements
            }
            self.texts.borrow_mut().push(t);
        }
    }

    fn append_based_on_parent_node(
        &self,
        _element: &Self::Handle,
        _prev_element: &Self::Handle,
        _child: NodeOrText<Self::Handle>,
    ) {
    }

    fn append_doctype_to_document(
        &self,
        _name: StrTendril,
        _public_id: StrTendril,
        _system_id: StrTendril,
    ) {
    }

    fn mark_script_already_started(&self, _node: &Self::Handle) {}

    fn pop(&self, _node: &Self::Handle) {}

    fn get_template_contents(&self, _: &Self::Handle) -> Self::Handle {
        Handle::new(Node::Element(html5ever::QualName::new(
            None,
            ns!(),
            local_name!("template"),
        )))
    }

    fn same_node(&self, _: &Self::Handle, _: &Self::Handle) -> bool {
        false
    }

    fn set_quirks_mode(&self, _mode: QuirksMode) {}

    fn append_before_sibling(&self, _sibling: &Self::Handle, _new_node: NodeOrText<Self::Handle>) {}

    fn add_attrs_if_missing(&self, _target: &Self::Handle, _attrs: Vec<Attribute>) {}

    fn associate_with_form(
        &self,
        _target: &Self::Handle,
        _form: &Self::Handle,
        _nodes: (&Self::Handle, Option<&Self::Handle>),
    ) {
    }

    fn remove_from_parent(&self, _target: &Self::Handle) {}

    fn reparent_children(&self, _node: &Self::Handle, _new_parent: &Self::Handle) {}

    fn is_mathml_annotation_xml_integration_point(&self, _handle: &Self::Handle) -> bool {
        false
    }

    fn set_current_line(&self, _line_number: u64) {}

    fn allow_declarative_shadow_roots(&self, _intended_parent: &Self::Handle) -> bool {
        true
    }

    fn attach_declarative_shadow(
        &self,
        _location: &Self::Handle,
        _template: &Self::Handle,
        _attrs: &[Attribute],
    ) -> bool {
        false
    }
}
