use std::sync::Arc;

use axum::{extract::State, http::{HeaderMap, StatusCode}, response::IntoResponse};

use crate::{db::Chunk, extract_header, AppState, files::File};



///Expects header 
///     - X-Filename
///     - ID
pub async fn delete_documents(
    header: HeaderMap, 
    State(state): State<Arc<AppState>>
) -> impl IntoResponse {
    let id: i64 = extract_header(&header, "ID").unwrap_or(0);
    let file: File = File::from_request(&header, &[]).unwrap();


    Chunk::delete(&state.db, id, file.get_filename().to_string()).await.unwrap();
    state.filesystem.remove_file(file).await.unwrap();
    (StatusCode::OK, "Success")
}