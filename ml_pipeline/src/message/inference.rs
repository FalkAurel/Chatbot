use axum::{extract::State, http::{HeaderMap, StatusCode}, response::IntoResponse, Json};
use serde::{Deserialize, Serialize};
use tracing::{instrument, info};
use std::sync::Arc;
use crate::{
    db::Kind, extract_header, message::{extract_id, Message, MessageError}, reasoning::{deep_think, normal_think, LLMResponse, ReasoningError}, AppState, HeaderError
};



pub enum InferenceError {
    MessageError(MessageError),
    ReasoningError(ReasoningError),
    HeaderError(HeaderError<String>)
}

impl IntoResponse for InferenceError {
    fn into_response(self) -> axum::response::Response {
        match self {
            Self::ReasoningError(err) => err.into_response(),
            Self::MessageError(err) => err.into_response(),
            Self::HeaderError(err) => err.into_response()
        }
    }
}


#[derive(Deserialize, Serialize, Debug)]
pub struct MessageInference {
    #[serde(rename="Kind")]
    pub kind: Kind,
    #[serde(rename="Message")]
    pub message: String,
    #[serde(rename="Model")]
    pub model: String,
    #[serde(rename="Preprompt")]
    pub pre_prompt: String
}

impl Into<Message> for MessageInference {
    fn into(self) -> Message {
        Message { kind: self.kind, message: self.message }
    }
}



#[axum::debug_handler]
#[instrument(skip(headers, app_state))]
pub async fn inference(
    headers: HeaderMap, 
    State(app_state): State<Arc<AppState>>,
    Json(message): Json<MessageInference>,
) -> Result<(StatusCode, Json<LLMResponse>), InferenceError> {
    let id: i64 = extract_id(&headers)
    .map_err(|err| InferenceError::MessageError(err))?;

    let deep_think_header: String = extract_header(&headers, "Deep_think")
    .map_err(|err| InferenceError::HeaderError(err))?;

    let model: String = message.model.clone();
    let pre_prompt: String = message.pre_prompt.clone();

    info!("Received pre_prompt: {}", pre_prompt);

    let response: Result<LLMResponse, ReasoningError> = match deep_think_header.as_str() {
        "True" => {
            deep_think(
                &model, 
                id, message.into(), 
                &app_state.db, 
                app_state.clone()
            ).await
        },
        "False" => {
            normal_think(
                &model,
                id,
                pre_prompt, 
                message.into(), 
                &app_state.db, 
                app_state.clone()
            ).await
        },
        _ => panic!("Fuck")
    };

    let json: Json<LLMResponse> = Json(
        response.map_err(|err| InferenceError::ReasoningError(err))?
    );

    Ok((StatusCode::OK, json))
}


