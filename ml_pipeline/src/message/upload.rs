use std::sync::Arc;
use tokio::task::{spawn_blocking, JoinError};
use tracing::instrument;

use axum::{
    http::{HeaderMap, StatusCode}, 
    Json, 
    extract::State, 
    response::IntoResponse
};

use crate::{
    AppState,
    message::{Message, extract_id},
    db::{MessageRecord, write_message},
};


pub enum UploadError {
    HeaderMissing,
    DBError(surrealdb::Error),
    EmbeddingError(fastembed::Error),
    EmbeddingMissing,
    JoinError(JoinError),
}

impl IntoResponse for UploadError {
    fn into_response(self) -> axum::response::Response {
        match self {
            Self::HeaderMissing => (
                StatusCode::BAD_REQUEST, "Header missing"
            ).into_response(),

            Self::DBError(err) => (
                StatusCode::INTERNAL_SERVER_ERROR, err.to_string()
            ).into_response(),

            Self::EmbeddingError(err) => (
                StatusCode::INTERNAL_SERVER_ERROR, err.to_string()
            ).into_response(),
            Self::EmbeddingMissing => (
                StatusCode::INTERNAL_SERVER_ERROR, "Embedding is missing"
            ).into_response(),
            Self::JoinError(err) => (
                StatusCode::INTERNAL_SERVER_ERROR, err.to_string()
            ).into_response()
        }
    }
}


#[instrument(skip(app))]
pub async fn upload_message(
    header: HeaderMap, 
    State(app): State<Arc<AppState>>, 
    Json(message): Json<Message>
) -> Result<impl IntoResponse, UploadError>
{
    let id: i64 = extract_id(&header).map_err(|_| UploadError::HeaderMissing)?;  
    
    let embedder: Arc<AppState> = app.clone();
    let (message, embedding) = spawn_blocking(move || {
        let embedding: Vec<f32> = embedder
        .embedder
        .embed(vec![&message.message], Some(16))
        .map_err(|err| UploadError::EmbeddingError(err))?
        .into_iter().next().ok_or(UploadError::EmbeddingMissing)?;

        Ok((message, embedding))
    }).await
    .map_err(|err| UploadError::JoinError(err))??;

    let msg: MessageRecord = MessageRecord::new(
        id, 
        message.kind, 
        message.message,
        embedding
    );

    write_message(&app.db, msg).await.map_err(|err| UploadError::DBError(err))?;
    Ok((StatusCode::OK, "Success".to_string()))
}