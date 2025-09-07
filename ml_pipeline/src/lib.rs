pub mod db;
pub mod files;
pub mod message;
pub mod documents;
pub mod text_processing;
pub mod reasoning;

use axum::{http::{HeaderMap, HeaderValue, StatusCode}, response::IntoResponse};
use fastembed::TextEmbedding;
use rake::{Rake, StopWords};
use reqwest::header::ToStrError;
use crate::{db::{Database, Init}, files::{Filesystem, Init as FSInit}};
use std::{collections::HashSet, str::FromStr};

pub enum HeaderError<T: FromStr> {
    HeaderMissing,
    EncodingError(ToStrError),
    ParseError(<T as FromStr>::Err)
}

impl <T: FromStr> IntoResponse for HeaderError<T>
where
    <T as FromStr>::Err: ToString
{
    fn into_response(self) -> axum::response::Response {
        match self {
            Self::EncodingError(err) => (StatusCode::BAD_REQUEST, err.to_string()).into_response(),
            Self::HeaderMissing => (StatusCode::BAD_REQUEST, "Header missing").into_response(),
            Self::ParseError(err) => (StatusCode::BAD_REQUEST, err.to_string()).into_response()
        }
    }
}

pub (crate) fn extract_header<T: FromStr>(header: &HeaderMap, key: &str) -> Result<T, HeaderError<T>> {
    let header: &HeaderValue = header.get(key).ok_or(HeaderError::HeaderMissing)?;
    let content: &str = header.to_str().map_err(|err| HeaderError::EncodingError(err))?;
    content.parse::<T>().map_err(|err| HeaderError::ParseError(err))
}

fn get_multilingual_stopwords() -> HashSet<&'static str> {
    HashSet::from_iter([
        // English
        "a", "an", "the", "and", "or", "but", "of", "in", "on", "at", "to", "for", "with", "by", 
        "from", "as", "is", "are", "was", "were", "be", "been", "have", "has", "had", "do", "does", 
        "did", "this", "that", "these", "those", "it", "its", "they", "them", "their", "you", "your",
        
        // German
        "der", "die", "das", "ein", "eine", "einen", "einer", "und", "oder", "aber", "in", "auf", 
        "an", "von", "zu", "mit", "für", "als", "ist", "sind", "war", "waren", "habe", "hast", "hat",
        "dieser", "jener", "man", "mich", "mir", "mein",
        
        // Chinese (Simplified/Traditional)
        "的", "了", "和", "是", "在", "有", "我", "你", "他", "她", "我们", "你们", "他们", "这", "那",
        "这些", "那些", "一个", "可以", "要", "会", "也", "都", "与", "为", "之", "著", "們", "這", "那",
        
        // Japanese
        "の", "に", "は", "を", "た", "が", "で", "て", "と", "し", "れ", "さ", "ある", "いる", "も",
        "する", "から", "なる", "こと", "よう", "これ", "それ", "あれ", "この", "その", "私", "あなた",
        
        // Universal noise
        "http", "https", "www", "com", "org", "net", "click", "here", "page", "section"
    ])
}
pub fn init_rake() -> Rake {
    let mut stopwords: StopWords = StopWords::new();

    for word in  get_multilingual_stopwords().drain() {
        stopwords.insert(word.to_string());
    }

    Rake::new(stopwords)
}



pub struct AppState {
    pub db: Database<Init>,
    pub embedder: TextEmbedding,
    pub filesystem: Filesystem<FSInit>,
    pub rake: Rake
}

