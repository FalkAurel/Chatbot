use std::{marker::PhantomData, sync::LazyLock};
use surrealdb::{
    engine::remote::http::{Client, Http}, 
    opt::auth::Root, Error as DBError, 
    Surreal
};

mod message;
mod chunks;
mod delete_user;
pub use message::*;
pub use chunks::*;
pub use delete_user::*;

const EMBEDDING_DIMENSION: usize = 384;
pub const DATABASE_CONNECTION: LazyLock<String> = LazyLock::new(|| {
    let endpoint: String = std::env::var("SURREAL_URL").unwrap_or("localhost:8000".to_string());
    println!("Connecting to Endpoint {endpoint}");
    endpoint
});

pub struct Uninit {}
pub struct Init {}

pub struct Database<T> {
    _marker: PhantomData<T>,
    db: surrealdb::Surreal<Client>
}

impl Database<Uninit> {
    pub async fn init_db(resource: &str) -> Result<Database<Init>, DBError> {
        let db:Surreal<Client>  = Surreal::new::<Http>(resource).await?;
        db.signin(Root {
            username: "root",
            password: "root"
        }).await?;

        db.use_ns("Law-China").use_db("ML-Pipeline").await?;
        let query: String = format!("
            -- Define message Table
            DEFINE TABLE IF NOT EXISTS message SCHEMAFULL
            PERMISSIONS
                FOR delete WHERE user = $auth.id;

            DEFINE FIELD IF NOT EXISTS user_id ON TABLE message TYPE int;
            DEFINE FIELD IF NOT EXISTS kind ON TABLE message TYPE int;
            DEFINE FIELD IF NOT EXISTS message ON TABLE message TYPE string;
            DEFINE FIELD IF NOT EXISTS embedding ON TABLE message TYPE array<float, {EMBEDDING_DIMENSION}>;
            DEFINE FIELD IF NOT EXISTS creation ON TABLE message TYPE datetime DEFAULT time::now();
            DEFINE INDEX IF NOT EXISTS embedding_idx ON TABLE message COLUMNS embedding MTREE DIMENSION {EMBEDDING_DIMENSION} DIST COSINE CONCURRENTLY;

            -- Define Full-Text-Search Analyzer to search for similar or matching keywords
            DEFINE ANALYZER IF NOT EXISTS message_fts_global TOKENIZERS class FILTERS lowercase, ascii;

            DEFINE INDEX IF NOT EXISTS fts_message_idx ON TABLE message COLUMNS message SEARCH ANALYZER message_fts_global  BM25(1.2, 0.75) CONCURRENTLY;

            -- Define document table with chunking
            DEFINE TABLE IF NOT EXISTS chunks SCHEMAFULL
            PERMISSIONS
                FOR delete WHERE user = $auth.id;

            DEFINE FIELD IF NOT EXISTS user_id ON TABLE chunks TYPE int;
            DEFINE FIELD IF NOT EXISTS title ON TABLE chunks TYPE string;
            DEFINE FIELD IF NOT EXISTS content ON TABLE chunks TYPE string;
            DEFINE FIELD IF NOT EXISTS embedding ON TABLE chunks TYPE array<float, {EMBEDDING_DIMENSION}>;
            DEFINE FIELD IF NOT EXISTS filename ON TABLE chunks TYPE string;

            -- Vector similarity index
            DEFINE INDEX IF NOT EXISTS embedding_idx ON TABLE chunks COLUMNS embedding MTREE DIMENSION {EMBEDDING_DIMENSION} DIST COSINE CONCURRENTLY;

            -- Full-Text-Search setup
            DEFINE ANALYZER IF NOT EXISTS chunk_fts_global TOKENIZERS class FILTERS lowercase, ascii;
            DEFINE INDEX IF NOT EXISTS fts_chunk_content_idx ON TABLE chunks FIELDS content SEARCH ANALYZER chunk_fts_global BM25 CONCURRENTLY;
            DEFINE INDEX IF NOT EXISTS fts_chunk_title_idx ON TABLE chunks FIELDS title SEARCH ANALYZER chunk_fts_global BM25 CONCURRENTLY;
        ");

        db.query(query).await?;

        Ok(Database { _marker: PhantomData, db })
    }
}