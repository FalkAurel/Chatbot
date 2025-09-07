use std::sync::Arc;

use axum::{extract::State, http::{HeaderMap, StatusCode}, response::IntoResponse};
use surrealdb::Error;

use crate::{
    db::delete_message as delete,
    message::{extract_id, MessageError}, AppState
};

pub enum MessageDeletionError {
    HeaderError(MessageError),
    DBError(Error)
}

impl IntoResponse for MessageDeletionError {
    fn into_response(self) -> axum::response::Response {
        match self {
            Self::DBError(err) => (
                StatusCode::INTERNAL_SERVER_ERROR, err.to_string()
            ).into_response(),
            Self::HeaderError(err) => err.into_response()
        }
    }
}



pub async fn delete_message(
    headers: HeaderMap,
    State(state): State<Arc<AppState>>
) -> Result<(StatusCode, &'static str), MessageDeletionError> {
    let id: i64 = extract_id(&headers)
    .map_err(|err| MessageDeletionError::HeaderError(err))?;

    delete(&state.db, id).await
    .map_err(|err| MessageDeletionError::DBError(err))?;
    Ok((StatusCode::OK, "Successfully deleted"))
}

