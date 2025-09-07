use std::sync::Arc;
use tokio::task::spawn_blocking;

use crate::{
    db::{get_last_10, related_messages, ChunkRecord, Database, Init, MessageHistoryRecord, MessageRecordDB}, 
    message::Message, 
    reasoning::{build_dynamic_prompt, prompt_llm, LLMResponse, ReasoningError}, 
    AppState
};


// One shot prompting
// Much faster but less reliable
pub async fn normal_think(
    model: &str, 
    id: i64,
    pre_prompt: String,
    message: Message, 
    db: &Database<Init>,
    embedder: Arc<AppState>
) -> Result<LLMResponse, ReasoningError> {
    let (embedding, message) =  spawn_blocking(move || {
        let embedding: Vec<f32> = embedder.embedder.embed(vec![&message.message], None)
        .map_err(|err| ReasoningError::Embedding(err))?
        .into_iter().next().ok_or(ReasoningError::EmptyEmbedding)?;
        Ok((embedding, message))
    }).await.map_err(|err| ReasoningError::TaskJoin(err))??;

    let shared_embedding: Arc<Vec<f32>> = Arc::new(embedding);

    let related_messages: Vec<MessageRecordDB> = related_messages(db, id, shared_embedding.clone())
    .await.map_err(|err| ReasoningError::DBError(err))?;

    let related_documents: Vec<ChunkRecord> = ChunkRecord::most_related(db, id, shared_embedding)
    .await.map_err(|err| ReasoningError::ChunkError(err))?;


    let last_10_messages: Vec<MessageHistoryRecord> = get_last_10(db, id)
    .await.map_err(|err| ReasoningError::DBError(err))?;

    let prompt: String = build_dynamic_prompt(
        &pre_prompt, 
        related_messages, 
        related_documents,
        last_10_messages, 
        &message.message
    );

    println!("Prompt: {}", prompt);


    prompt_llm(model, &prompt).await
}