use super::FrameComment;

pub enum FrameData {
	Text(String),
	MultiText(Vec<String>),
	Comment(FrameComment),
	Binary(Vec<u8>),
}
