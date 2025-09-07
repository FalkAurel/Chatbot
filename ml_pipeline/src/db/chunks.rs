use crate::db::{Database, Init, EMBEDDING_DIMENSION};
use serde::{Deserialize, Serialize};
use surrealdb::{Error as DBError, Response};
use std::sync::Arc;


/// Errors that can occur when working with document chunks.
///
/// This enum represents various failure cases including:
/// - Invalid embedding dimensions
/// - Database operation failures
/// - Failed write operations
#[derive(Debug)]
pub enum ChunkError {
    /// The provided embedding vector has incorrect dimensions
    InvalidDimensions,
    /// Wrapper for database-related errors
    /// 
    /// Note: Uses `Box` to keep the error enum small while allowing
    /// large error types from the database layer.
    DBError(Box<DBError>),
    /// The write operation completed but returned no data
    WriteFailed,
}

/// Represents a segmented portion of a document with embedded semantic information.
///
/// Documents are divided into chunks to enable:
/// - More precise vector similarity searches
/// - Better handling of large documents
/// - Focused retrieval of relevant content
///
/// # Fields
/// - `user_id`: Unique identifier of the user who owns this chunk
/// - `content`: The actual text content of this document segment
/// - `embedding`: Vector representation of the chunk's semantic meaning (dimension: EMBEDDING_DIMENSION)
/// - `filename`: Source document filename for traceability and context
///
/// # Tradeoffs
/// - Larger chunks provide more context but reduce search precision
/// - Smaller chunks enable precise matching but may lose broader context
/// - Optimal chunk size depends on your specific use case and document type
#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct Chunk {
    user_id: i64,
    title: String,
    content: String,
    embedding: Vec<f32>,
    filename: String
}

impl Chunk {
    /// Creates a new document chunk with validation.
    ///
    /// # Arguments
    /// - `user_id`: Owner identifier
    /// - `content`: Text content of the chunk
    /// - `embedding`: Semantic vector (must match EMBEDDING_DIMENSION)
    /// - `filename`: Source document name
    ///
    /// # Errors
    /// Returns `ChunkError::InvalidDimensions` if embedding length doesn't match expected size
    pub fn new(
        user_id: i64,
        title: String,
        content: String, 
        embedding: Vec<f32>, 
        filename: String
    ) -> Result<Self, ChunkError> {
        if embedding.len() != EMBEDDING_DIMENSION {
            return Err(ChunkError::InvalidDimensions)
        }
        Ok(Self { user_id, title, content, embedding, filename })
    }

    /// Persists the chunk to the database.
    ///
    /// # Arguments
    /// - `db`: Initialized database connection
    ///
    /// # Errors
    /// - `ChunkError::DBError` for database operation failures
    /// - `ChunkError::WriteFailed` if creation returns no data
    pub async fn write(self, db: &Database<Init>) -> Result<(), ChunkError> {
        let res: Option<Self> = db.db
        .create("chunks")
        .content(self).await
        .map_err(|err| ChunkError::DBError(Box::new(err)))?;

        res.ok_or(ChunkError::WriteFailed)?;
        Ok(())
    }

    pub async fn write_bulk(db: &Database<Init>, data: Vec<Self>) -> Result<(), ChunkError> {
        db.db.query(
            "
            INSERT INTO chunks $data;

            -- Rebuilding concurrently search indexes in the background to prevent performace degradation
            REBUILD INDEX IF EXISTS embedding_idx ON chunks;
            REBUILD INDEX IF EXISTS fts_chunk_content_idx ON chunks;
            REBUILD INDEX IF EXISTS fts_chunk_title_idx ON chunks;
            "
        ).bind(("data", data)).await.map_err(|err| ChunkError::DBError(Box::new(err)))?;
        Ok(())
    }


    pub async fn delete_all(db: &Database<Init>, id: i64) -> Result<(), ChunkError> {
        db.db.query("DELETE chunks WHERE user_id = $user_id;")
        .bind(("user_id", id))
        .await.map_err(|err| ChunkError::DBError(Box::new(err)))?;

        Ok(())
    }


    pub async fn delete(db: &Database<Init>, id: i64, storage_name: String) -> Result<(), ChunkError> {
        db.db.query("
        DELETE chunks WHERE user_id = $user_id AND filename = $storage_name
        ")
        .bind(("user_id", id))
        .bind(("storage_name", storage_name))
        .await.map_err(|err| ChunkError::DBError(Box::new(err)))?;

        Ok(())
    }

}

#[allow(dead_code)]
#[derive(Deserialize, Debug)]
pub struct ChunkRecord {
    pub(crate) title: String,
    pub(crate) content: String,
    pub(crate) similarity: f32,
}

impl ChunkRecord {
    const GEOMETRIC_SEARCH: &str = r#"
        SELECT *, vector::similarity::cosine(embedding, $embedding) AS similarity
        FROM chunks
        WHERE user_id = $user_id AND embedding <|100|> $embedding 
        ORDER BY similarity DESC
        LIMIT 20;
    "#;

    pub async fn most_related(
        db: &Database<Init>,
        user_id: i64,
        embedding: Arc<Vec<f32>>
    ) -> Result<Vec<Self>, ChunkError> {
        let mut response: Response = db.db.query(Self::GEOMETRIC_SEARCH)
            .bind(("embedding", embedding))
            .bind(("user_id", user_id))
            .await.map_err(|err| ChunkError::DBError(Box::new(err)))?;
        
        let records: Vec<Self> = response.take(0)
        .map_err(|err| ChunkError::DBError(Box::new(err)))?;

        Ok(records)
    }
}


#[cfg(test)]
mod chunking {
    use std::{sync::Arc, time::Instant};
    use crate::db::{Chunk, ChunkRecord, Database, Init, DATABASE_CONNECTION, EMBEDDING_DIMENSION};

    #[tokio::test]
    async fn test_insertion() {
        let db: Database<Init> = Database::init_db(&*DATABASE_CONNECTION).await.unwrap();

        let a = std::iter::repeat_with(|| Chunk::new(
            1, 
            "TEST".to_string(),
            "Hallo Welt".to_owned(), 
            vec![0.5; EMBEDDING_DIMENSION], 
            "hallo.txt".to_string()).unwrap()).take(57);

        let instant: Instant = Instant::now();
        Chunk::write_bulk(&db, a.collect()).await.unwrap();
        println!("10 Insertions took: {:?}", instant.elapsed());


        let response: Vec<ChunkRecord> = ChunkRecord::most_related(
            &db, 
            1,
            Arc::new(vec![0.5; EMBEDDING_DIMENSION])
        ).await.unwrap();

        assert_eq!(response.len(), 10);
    }
}