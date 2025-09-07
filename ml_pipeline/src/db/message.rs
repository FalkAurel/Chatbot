use serde::{Deserialize, Serialize};
use serde_repr::{Deserialize_repr, Serialize_repr};
use crate::db::{Database, Init, DBError};
use std::{fmt::Display, sync::Arc};


#[repr(u8)]
#[derive(Deserialize_repr, Serialize_repr, Debug)]
pub enum Kind {
    AI = 0,
    User = 1,
}

impl Display for Kind {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Self::AI => write!(f, "AI"),
            Self::User => write!(f, "User")
        }
    }
}

/// This struct can be written to the db
#[derive(Debug, Deserialize, Serialize)]
pub struct MessageRecord {
    pub user_id: i64,
    pub kind: Kind,
    pub message: String,
    pub embedding: Vec<f32>,
}

#[derive(Deserialize, Serialize)]
pub struct MessageHistoryRecord {
    pub kind: Kind,
    pub message: String
}

impl MessageRecord {
    pub const fn new(
        user_id: i64, 
        kind: Kind, 
        message: String, 
        embedding: Vec<f32>
    ) -> Self {
        Self { user_id, kind, message, embedding }
    }
}


// ToDo: Fix the res unwrap
pub async fn write_message(
    db: &Database<Init>,
    msg: MessageRecord
) -> Result<(), DBError> {
    let res: Option<MessageRecord> = db.db
    .create("message")
    .content(msg)
    .await?;

    res.unwrap();
    Ok(())
}

pub async fn delete_message(db: &Database<Init>, user_id: i64) -> Result<(), DBError> {
    db.db.query("DELETE FROM message WHERE user_id = $user_id;")
    .bind(("user_id", user_id))
    .await?;

    Ok(())
}

/// This struct ressembles a database entry.
#[derive(Debug, Serialize, Deserialize)]
pub struct MessageRecordDB {
    pub user_id: i64,
    pub kind: Kind,
    pub message: String,
    pub embedding: Vec<f32>,
    pub combined_score: f32,
}

pub async fn related_messages(
    db: &Database<Init>,
    user_id: i64,
    embedding: Arc<Vec<f32>>
) -> Result<Vec<MessageRecordDB>, DBError> {
    let query: &str = r#"
        SELECT 
            user_id, 
            kind, 
            message, 
            embedding,
            vector::similarity::cosine(embedding, $embedding) AS combined_score
        FROM message
        WHERE user_id = $user_id AND embedding <|10|> $embedding
        ORDER BY combined_score DESC
    "#;

    db.db.query(query)
    .bind(("embedding", embedding))
    .bind(("user_id", user_id))
    .await?.take(0)
}

pub async fn get_history(db: &Database<Init>, user_id: i64) -> Result<Vec<MessageHistoryRecord>, DBError> {
    const QUERY_HISTORY: &str = "
    SELECT kind, message, creation FROM message WHERE user_id = $user_id
    ORDER BY creation ASC
    ";

    db.db.query(QUERY_HISTORY)
    .bind(("user_id", user_id))
    .await?.take(0)
}

pub async fn get_last_10(db: &Database<Init>, user_id: i64) -> Result<Vec<MessageHistoryRecord>, DBError> {
    const QUERY_HISTORY: &str = "
    SELECT kind, message, creation FROM message WHERE user_id = $user_id
    ORDER BY creation ASC
    LIMIT 10;
    ";

    db.db.query(QUERY_HISTORY)
    .bind(("user_id", user_id))
    .await?.take(0)
}


#[cfg(test)]
mod message {
    use rand::random;
    use std::sync::Arc;

    use crate::db::{
        related_messages, write_message, Database, Init, Kind, MessageRecord, MessageRecordDB, DATABASE_CONNECTION, EMBEDDING_DIMENSION
    };

    #[tokio::test]
    async fn test_message_insertion() {
        let db: Database<Init> = Database::init_db(&*DATABASE_CONNECTION).await.unwrap();

        let res: Result<(), surrealdb::Error> = write_message(&db, MessageRecord { 
            kind: Kind::AI, 
            message: String::from("Hallo Welt"),
            embedding: vec![0.0; EMBEDDING_DIMENSION as usize],
            user_id: 1,
        }).await;

        assert!(res.is_ok());
    }

    #[tokio::test]
    async fn test_message_extraction() {
        let db: Database<Init> = Database::init_db(&*DATABASE_CONNECTION).await.unwrap();

        let cluster_centers: [Vec<f32>; 3] = [
            vec![0.9; EMBEDDING_DIMENSION],
            vec![0.7; EMBEDDING_DIMENSION], 
            vec![0.5; EMBEDDING_DIMENSION]
        ];

        for i in 0..10 {
            let cluster_idx: usize = i % cluster_centers.len();
            let mut embedding: Vec<f32> = cluster_centers[cluster_idx].clone();
            
            // Add some noise within cluster
            for j in 0..EMBEDDING_DIMENSION {
                embedding[j] += rand::random_range(-0.5..0.5);
            }
            
            write_message(&db, MessageRecord {
                kind: Kind::AI,
                message: format!("C{}M{}", cluster_idx, i),
                embedding,
                user_id: 1,
            }).await.unwrap();
        }

        let neighbours: Vec<MessageRecordDB> = related_messages(
            &db,
            1,
            Arc::new(std::iter::repeat_n(random::<f32>(), EMBEDDING_DIMENSION).collect())
        ).await.unwrap();

        assert_eq!(neighbours.len(), 10);
        assert!(
            neighbours.iter().all(|neighbour| {
                dbg!(&neighbour.message, &neighbour.combined_score);
                neighbour.combined_score > 0.0
            })
        );
    }
}