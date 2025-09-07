use std::sync::Arc;
use tokio::task::spawn_blocking;

use crate::{
    db::{related_messages, ChunkRecord, Database, Init, MessageRecordDB}, 
    message::Message, 
    reasoning::{prompt_llm, LLMResponse, ReasoningError}, AppState
};


fn deep_think_prompt<'a>(
    prompt: &'a mut String,
    message: &str,
    previous_reasoning: &[String],
    related_messages: &[MessageRecordDB],
    related_documents: &[ChunkRecord],
) -> &'a str {
    // Start with a directive for deep reasoning
    prompt.push_str("### Deep Thinking Mode Activated ###\n");
    prompt.push_str("## Instructions:\n");
    prompt.push_str("- Analyze the question rigorously, step-by-step.\n");
    prompt.push_str("- Cross-reference with prior reasoning and related messages.\n");
    prompt.push_str("- Use provided documents for factual accuracy.\n");
    prompt.push_str("- Avoid assumptions; verify when possible.\n\n");

    // Add the user's message
    prompt.push_str("## User's Query:\n");
    prompt.push_str(message);
    prompt.push_str("\n\n");

    // Include previous reasoning steps (if any)
    if !previous_reasoning.is_empty() {
        prompt.push_str("## Previous Reasoning Steps:\n");
        for (i, step) in previous_reasoning.iter().enumerate() {
            prompt.push_str(&format!("{}. {}\n", i + 1, step));
        }
        prompt.push_str("\n");
    }

    // Include related conversation history (if any)
    if !related_messages.is_empty() {
        prompt.push_str("## Conversation Context:\n");
        for msg in related_messages {
            prompt.push_str(&format!("[{}]: {}\n", msg.kind.to_string(), msg.message));
        }
        prompt.push_str("\n");
    }

    // Attach relevant document chunks (if any)
    if !related_documents.is_empty() {
        prompt.push_str("## Supporting Documents:\n");
        for (i, chunk) in related_documents.iter().enumerate() {
            prompt.push_str(
                &format!("### [Doc {} | Filename: {} | Similarity: {}]\n", i + 1, chunk.title, chunk.similarity)
            );
            prompt.push_str(&chunk.content);
            prompt.push_str("\n---\n");
        }
    }

    // Final instruction for structured reasoning
    prompt.push_str("\n## Required Output Format:\n");
    prompt.push_str("- **Thought Process:** [Detailed step-by-step reasoning]\n");
    prompt.push_str("- **Conclusion:** [Final synthesized answer]\n");
    prompt.push_str("- **Sources:** [Relevant document references, if applicable]\n");

    prompt.as_str()
}


/// Performs a deep thinking search
pub async fn deep_think(
    model: &str, 
    id: i64, 
    mut message: Message, 
    db: &Database<Init>,
    embedder: Arc<AppState>
) -> Result<LLMResponse, ReasoningError> {
    let cloned_version: Arc<AppState> = embedder.clone();
    let (embedding, ret_message) = spawn_blocking(move || {
        let embedding: Vec<f32> = cloned_version.embedder.embed(vec![&message.message], None)
        .map_err(|err| ReasoningError::Embedding(err))?
        .into_iter().next().ok_or(ReasoningError::EmptyEmbedding)?;
        Ok((embedding, message))
    }).await.map_err(|err| ReasoningError::TaskJoin(err))??;

    message = ret_message;
    let mut rc: Arc<Vec<f32>> = Arc::new(embedding);

    let mut prompt: String = String::with_capacity(2048);
    prompt.extend(message.message.chars());

    let mut thought_chain: Vec<String> = Vec::with_capacity(5);

    for _ in 0..5 {
        let related_messages: Vec<crate::db::MessageRecordDB> = related_messages(db, id, rc.clone())
        .await.map_err(|err| ReasoningError::DBError(err))?;

        let related_documents: Vec<crate::db::ChunkRecord> = ChunkRecord::most_related(db, id, rc.clone())
        .await.map_err(|err| ReasoningError::ChunkError(err))?;

        let reasoning: &str = deep_think_prompt(
            &mut prompt,
            &message.message,
            &thought_chain,
            &related_messages, 
            &related_documents, 
        );

        let response: String = prompt_llm(model, reasoning).await?.response;
        let cloned_version: Arc<AppState> = embedder.clone();

        // Generate new embeddings
        let (embedding, returned_response) = spawn_blocking(move || {
            let embedding: Vec<f32> = cloned_version.embedder.embed(vec![&response], None)
            .map_err(|err| ReasoningError::Embedding(err))?
            .into_iter().next().ok_or(ReasoningError::EmptyEmbedding)?;
            Ok((embedding, response))
        }).await.map_err(|err| ReasoningError::TaskJoin(err))??;


        // Update embeddings
        rc = Arc::new(embedding);

        // Preparing for next iteration
        thought_chain.push(returned_response);
        prompt.clear();
    }

    let related_messages: Vec<crate::db::MessageRecordDB> = related_messages(db, id, rc.clone())
    .await.map_err(|err| ReasoningError::DBError(err))?;

    let related_documents: Vec<crate::db::ChunkRecord> = ChunkRecord::most_related(db, id, rc.clone())
    .await.map_err(|err| ReasoningError::ChunkError(err))?;

    let reasoning: &str = deep_think_prompt(
        &mut prompt,
        &message.message,
        &thought_chain,
        &related_messages, 
        &related_documents, 
    );

    prompt_llm(model, reasoning).await
}