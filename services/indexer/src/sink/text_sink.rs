use std::{borrow::Cow, cell::RefCell, rc::Rc};

use html5ever::{
    interface::{ElementFlags, NodeOrText, QuirksMode, TreeSink},
    local_name, ns,
    tendril::StrTendril,
    Attribute, QualName,
};

#[derive(Debug, Clone)]
pub enum Node {
    Element(&'static QualName),
    Text(String),
}

type Handle = Rc<Node>;

pub struct TextSink {
    pub elems: RefCell<Vec<Handle>>,
    pub doc: Handle,
}

impl TextSink {
    pub fn new() -> Self {
        Self {
            elems: RefCell::new(vec![]),
            doc: Rc::new(Node::Text("".to_string())),
        }
    }
}
impl TreeSink for TextSink {
    type Output = Vec<Handle>;
    type Handle = Handle;
    type ElemName<'a> = &'a QualName;

    fn finish(self) -> Self::Output {
        self.elems.into_inner()
    }

    fn parse_error(&self, msg: Cow<'static, str>) {
        eprintln!("Parse error: {msg}");
    }

    fn get_document(&self) -> Self::Handle {
        self.doc.clone()
    }

    fn elem_name<'a>(&'a self, target: &'a Self::Handle) -> Self::ElemName<'a> {
        if let Node::Element(name) = target.as_ref() {
            name
        } else {
            panic!("Not an element");
        }
    }

    fn create_comment(&self, text: StrTendril) -> Self::Handle {
        Rc::new(Node::Text(text.to_string()))
    }

    fn create_element(&self, name: QualName, _: Vec<Attribute>, _: ElementFlags) -> Self::Handle {
        Rc::new(Node::Element(Box::leak(Box::new(name))))
    }

    #[allow(unused_variables)]
    fn create_pi(&self, target: StrTendril, data: StrTendril) -> Self::Handle {
        unimplemented!()
    }

    fn append(&self, _: &Self::Handle, child: NodeOrText<Self::Handle>) {
        if let NodeOrText::AppendText(t) = child {
            if !t.trim().is_empty() {
                self.elems
                    .borrow_mut()
                    .push(Rc::new(Node::Text(t.to_string())));
            }
        }
    }

    #[allow(unused_variables)]
    fn append_based_on_parent_node(
        &self,
        element: &Self::Handle,
        prev_element: &Self::Handle,
        child: NodeOrText<Self::Handle>,
    ) {
    }

    #[allow(unused_variables)]
    fn append_doctype_to_document(
        &self,
        name: StrTendril,
        public_id: StrTendril,
        system_id: StrTendril,
    ) {
    }

    fn mark_script_already_started(&self, _node: &Self::Handle) {}

    fn pop(&self, _node: &Self::Handle) {}

    fn get_template_contents(&self, _: &Self::Handle) -> Self::Handle {
        Rc::new(Node::Element(Box::leak(Box::new(
            html5ever::QualName::new(None, ns!(), local_name!("template")),
        ))))
    }

    #[allow(unused_variables)]
    fn same_node(&self, x: &Self::Handle, y: &Self::Handle) -> bool {
        false
    }

    #[allow(unused_variables)]
    fn set_quirks_mode(&self, mode: QuirksMode) {}

    #[allow(unused_variables)]
    fn append_before_sibling(&self, sibling: &Self::Handle, new_node: NodeOrText<Self::Handle>) {}

    #[allow(unused_variables)]
    fn add_attrs_if_missing(&self, target: &Self::Handle, attrs: Vec<Attribute>) {}

    #[allow(unused_variables)]
    fn associate_with_form(
        &self,
        _target: &Self::Handle,
        _form: &Self::Handle,
        _nodes: (&Self::Handle, Option<&Self::Handle>),
    ) {
    }

    #[allow(unused_variables)]
    fn remove_from_parent(&self, target: &Self::Handle) {}

    #[allow(unused_variables)]
    fn reparent_children(&self, node: &Self::Handle, new_parent: &Self::Handle) {}

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
