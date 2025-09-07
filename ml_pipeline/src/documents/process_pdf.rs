use std::{collections::BTreeMap, sync::Arc};
use axum::{
    body::Bytes, 
    extract::State, 
    http::{self, HeaderMap}, response::IntoResponse,
};

use tokio::task::spawn_blocking;
use tracing::{instrument, error};
use lopdf::{Document, ObjectId};

use crate::{
    db::Chunk, documents::{
        chunk_text::{Chunker, FixedSizeOverlap}, normalization::normalize,
    }, extract_header, files::File, AppState
};

#[instrument(skip_all)]
pub async fn process_pdf(
    headers: HeaderMap,
    State(state): State<Arc<AppState>>,
    body: Bytes
) -> impl IntoResponse {
    let id: i64 = unsafe { extract_header(&headers, "ID").unwrap_unchecked() };
    let title: String = unsafe { extract_header(&headers, "Title").unwrap_unchecked() };

    let file: File = File::from_request(&headers, &body).unwrap();
    let document: Document = Document::load_mem(file.get_content()).unwrap();
    let pages: BTreeMap<u32, ObjectId> = document.get_pages();
    let storage_name: String = file.get_filename().to_string();

    let cloned_state: Arc<AppState> = state.clone();

    let chunks: Vec<Chunk> = spawn_blocking(move || {
        let normalized_text: String = normalize(
            document.extract_text(
                &pages.keys().copied().collect::<Vec<u32>>()
            ).unwrap(), 
            &[]
        );

        let chunker: FixedSizeOverlap::<1000, 100> = FixedSizeOverlap::new(&normalized_text);
        let text: Vec<&str> = chunker.chunks().collect();

        let chunks: Vec<Chunk> = text.chunks(16).map(move |batch | {
            let embedding: Vec<Vec<f32>> = state.embedder.embed(batch.to_vec(), Some(16)).unwrap();

            let cloned_storagename: String = storage_name.clone();
            let title_cloned: String = title.clone();

            batch.iter().zip(embedding).map(move |(c, e)| {
                Chunk::new(id, title_cloned.clone(), c.to_string(), e, cloned_storagename.clone()).unwrap()
            })
        }).flatten().collect();

        chunks
    }).await.unwrap();

    if let Err(err) = cloned_state.filesystem.write(&file).await {
        error!("Failed to save file: {err:?}");
        return (http::StatusCode::INTERNAL_SERVER_ERROR, "Failed to save file")
    }

    for batch in chunks.chunks(40) {
        Chunk::write_bulk(&cloned_state.db, batch.to_vec()).await.unwrap();
    }

    (http::StatusCode::OK, "NICE")
}