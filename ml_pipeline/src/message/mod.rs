use std::error::Error;
use serde::{Deserialize, Serialize};
use axum::{
    http::{HeaderMap, StatusCode}, 
    response::IntoResponse
};

use crate::db::Kind;

mod inference;
mod history;
mod upload;
mod delete;
pub use upload::upload_message;
pub use inference::inference;
pub use history::history;
pub use delete::delete_message;

#[derive(Deserialize, Serialize, Debug)]
pub struct Message {
    pub kind: Kind,
    pub message: String,
}

pub enum MessageError {
    IDHeaderMissing,
    InvalidHeader(Box<dyn Error + Send + Sync>)
}

impl IntoResponse for MessageError {
    fn into_response(self) -> axum::response::Response {
        let response: (StatusCode, String) = match self {
            Self::IDHeaderMissing => (
                StatusCode::BAD_REQUEST, "Request header does not contain 'ID'".to_string()
            ),
            Self::InvalidHeader(err) => (
                StatusCode::BAD_REQUEST, err.to_string()
            )
        };

        response.into_response()
    }
}


pub fn extract_id(headers: &HeaderMap) -> Result<i64, MessageError> {
    match headers.get("ID") {
        Some(value) => match value.to_str() {
            Ok(value) => match value.parse::<i64>() {
                Ok(id) => Ok(id),
                Err(err) => Err(MessageError::InvalidHeader(Box::new(err)))
            },
            Err(err) => return Err(MessageError::InvalidHeader(Box::new(err)))
        },
        None => return Err(MessageError::IDHeaderMissing)
    }
}