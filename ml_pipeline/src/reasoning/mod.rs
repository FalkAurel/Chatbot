mod deep_think;
mod normal;


use crate::db::{ChunkError, ChunkRecord, MessageHistoryRecord, MessageRecordDB};
use std::{sync::LazyLock, fmt::Write};
use axum::response::IntoResponse;
use reqwest::{Client, StatusCode};
use serde::{Deserialize, Serialize};
use tokio::task::JoinError;


pub use deep_think::deep_think;
pub use normal::normal_think;



static HTTPCLIENT: LazyLock<Client> = LazyLock::new(|| Client::new());
const RESPONSE_GENERATION_URL: &str ="http://ollama:11434/api/generate";

pub enum ReasoningError {
    DBError(surrealdb::Error),
    ChunkError(ChunkError),
    RequestError(reqwest::Error),
    LLMError(String),
    TaskJoin(JoinError),
    Embedding(fastembed::Error),
    EmptyEmbedding,
}


impl IntoResponse for ReasoningError {
    fn into_response(self) -> axum::response::Response {
        match self {
            Self::ChunkError(err) => (StatusCode::INTERNAL_SERVER_ERROR, format!("{err:?}")).into_response(),
            Self::DBError(err) => (StatusCode::INTERNAL_SERVER_ERROR, err.to_string()).into_response(),
            Self::Embedding(err) => (StatusCode::INTERNAL_SERVER_ERROR, err.to_string()).into_response(),
            Self::LLMError(err) => (StatusCode::INTERNAL_SERVER_ERROR, err).into_response(),
            Self::RequestError(err) => (StatusCode::INTERNAL_SERVER_ERROR, err.to_string()).into_response(),
            Self::TaskJoin(err) => (StatusCode::INTERNAL_SERVER_ERROR, err.to_string()).into_response(),
            Self::EmptyEmbedding => (StatusCode::INTERNAL_SERVER_ERROR, "Failed to create embeddings".to_string()).into_response()
        }
    }
}

/// Constructs a dynamic LLM prompt from multiple contextual components.
///
/// Organizes information hierarchically with clear section demarcations:
/// 1. Task overview and preprompt
/// 2. Context integration (messages + documents)
/// 3. User query with response guidelines
///
/// Memory-efficient implementation with preallocated String capacity.
pub(super) fn build_dynamic_prompt(
    pre_prompt: &str,
    related_messages: Vec<MessageRecordDB>,
    related_documents: Vec<ChunkRecord>,
    previous_messages: Vec<MessageHistoryRecord>,
    user_input: &str
) -> String {
    // Preallocate to reduce reallocations (adjust capacity based on typical use)
    let mut prompt = String::with_capacity(4096);  // Increased for document content

    // Task Overview Section
    prompt.push_str("# Task Overview\n\n");
    prompt.push_str("To complete your task you will receive:\n");
    prompt.push_str("- Preprompt: Task background and instructions\n");
    prompt.push_str("- Context: Relevant messages/documents\n");
    prompt.push_str("- Conversation history: Last 10 messages\n");
    prompt.push_str("- User query: The request to address\n\n");

    // Preprompt Section
    prompt.push_str("## Preprompt\n");
    prompt.push_str(pre_prompt.trim());
    prompt.push_str("\n\n");

    // Context Integration Section
    prompt.push_str("## Context Integration\n");

    // Message Context
    prompt.push_str("### Relevant Messages\n");
    if !related_messages.is_empty() {
        for msg in related_messages {
            writeln!(
                prompt,
                "- [{} | Score: {:.2}]: {}",
                msg.kind,
                msg.combined_score,
                msg.message.trim()
            ).unwrap();
        }
    } else {
        prompt.push_str("(No relevant messages found)\n");
    }

    // Document Context
    prompt.push_str("\n### Relevant Documents\n");
    if !related_documents.is_empty() {
        for doc in related_documents {
            writeln!(
                prompt,
                "#### Title: {}\n[Similarity: {:.4}]\n{}\n---",
                doc.title,
                doc.similarity,
                doc.content.trim()
            ).unwrap();
        }
    } else {
        prompt.push_str("(No relevant documents found)\n");
    }

    // Conversation History
    prompt.push_str("\n### Conversation History\n");
    if !previous_messages.is_empty() {
        for msg in previous_messages.iter().take(10) {
            writeln!(
                prompt,
                "- [{}]: {}",
                msg.kind,
                msg.message.trim()
            ).unwrap();
        }
    } else {
        prompt.push_str("(No previous messages)\n");
    }

    // User Query Section
    prompt.push_str("\n## User Query\n");
    if !user_input.trim().is_empty() {
        writeln!(prompt, "> {}", user_input.trim().replace('\n', "\n> ")).unwrap();
        prompt.push_str("\n**Response Guidelines:**\n");
        prompt.push_str("- Provide clear, concise answers\n");
        prompt.push_str("- Flag legal/regulatory content\n");
        prompt.push_str("- Reference sources when available\n");
    } else {
        prompt.push_str("(No user input detected)\n");
    }

    prompt
}


#[derive(Serialize)]
pub struct LLMRequest<'a> {
    model: &'a str,
    prompt: &'a str,
    stream: bool,
    think: bool,
}

#[derive(Deserialize, Serialize)]
pub struct LLMResponse {
    response: String
}

#[derive(Serialize)]
pub struct Options {
    temperature: f32,  // 0.3 für strikte Einhaltung
    repeat_penalty: f32,
}


/// Prompts the LLM
pub async fn prompt_llm(model: &str, prompt: &str) -> Result<LLMResponse, ReasoningError> {
    let request: LLMRequest = LLMRequest { 
        model, 
        prompt, 
        stream: false, 
        think: false,
    };

    let response: reqwest::Response = HTTPCLIENT.post(RESPONSE_GENERATION_URL)
    .json(&request).send().await.map_err(|err| ReasoningError::RequestError(err))?;


    if response.status() != StatusCode::OK {
        let error: String = str::from_utf8(&response.bytes().await.unwrap()).unwrap().to_string();
        return Err(ReasoningError::LLMError(error))
    }


    response.json().await.map_err(|err| ReasoningError::RequestError(err))
}