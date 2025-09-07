use std::sync::Arc;

use axum::{extract::State, http::{HeaderMap, StatusCode}, response::IntoResponse};
use crate::{
    db::{delete_message, Chunk, ChunkError}, 
    extract_header, AppState, HeaderError
};


pub enum DeleteUser {
    IDHeader(HeaderError<i64>),
    DBError(surrealdb::Error),
    ChunkError(ChunkError)
}

impl IntoResponse for DeleteUser {
    fn into_response(self) -> axum::response::Response {
        match self {
            Self::ChunkError(err) => (StatusCode::INTERNAL_SERVER_ERROR, format!("{err:?}")).into_response(),
            Self::DBError(err) => (StatusCode::INTERNAL_SERVER_ERROR, err.to_string()).into_response(),
            Self::IDHeader(err) => err.into_response()
        }
    }
}


pub async fn delete_user(
    header: HeaderMap, 
    State(app_state): State<Arc<AppState>>
) -> Result<(StatusCode, &'static str), DeleteUser> {
    let id: i64 = extract_header(&header, "ID")
    .map_err(|err| DeleteUser::IDHeader(err))?;


    delete_message(&app_state.db, id)
    .await.map_err(|err| DeleteUser::DBError(err))?;

    Chunk::delete_all(&app_state.db, id)
    .await.map_err(|err| DeleteUser::ChunkError(err))?;

    app_state.filesystem.remove_user(id).await.unwrap();


    Ok((StatusCode::OK, "Deletion Success"))
}