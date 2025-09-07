use std::sync::Arc;

use axum::{
    extract::State, http::{HeaderMap, StatusCode},
    response::IntoResponse, Json
};
use crate::{
    db::{get_history, MessageHistoryRecord},
    message::extract_id, AppState
};
use surrealdb::Error;
use super::MessageError;

pub enum MessageHistoryError {
    MessageError(MessageError),
    DbError(Error)
}

impl IntoResponse for MessageHistoryError {
    fn into_response(self) -> axum::response::Response {
        match self {
            Self::MessageError(err) => err.into_response(),
            Self::DbError(err) => (StatusCode::INTERNAL_SERVER_ERROR, err.to_string()).into_response()
        }
    }
}

#[axum::debug_handler]
pub async fn history(
    headers: HeaderMap, 
    State(app_state): State<Arc<AppState>>
) -> Result<(StatusCode, Json<Vec<MessageHistoryRecord>>), MessageHistoryError>
{
    let id: i64 = extract_id(&headers).map_err(|err| MessageHistoryError::MessageError(err))?;
    let history: Vec<MessageHistoryRecord> = get_history(&app_state.db, id)
    .await
    .map_err(|err| MessageHistoryError::DbError(err))?;

    Ok((StatusCode::OK, Json(history)))
}