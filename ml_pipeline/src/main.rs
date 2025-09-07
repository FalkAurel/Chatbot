use std::sync::Arc;

use axum::{extract::DefaultBodyLimit, routing::{delete, get, post}, Router};
use fastembed::{InitOptions, TextEmbedding, EmbeddingModel};
use rake::Rake;
use tokio::{
    net::TcpListener, 
    runtime::{Builder, Runtime}
};

use ml_pipeline::{
    db::{delete_user, Database, Init, DATABASE_CONNECTION}, documents::{delete_documents, process_pdf}, files::{Filesystem, Init as FSInit}, init_rake, message::{delete_message, history, inference, upload_message}, AppState
};

use tower_http::trace::TraceLayer;
use tracing_subscriber::EnvFilter;
use std::env;

#[cfg(not(feature = "build"))]
fn main() {
    let runtime: Runtime = Builder::new_multi_thread()
    .max_blocking_threads(512)
    .enable_all()
    .thread_stack_size(10 * 1024 * 1024)
    .global_queue_interval(40)
    .build()
    .unwrap();

    tracing_subscriber::fmt().with_env_filter(
        EnvFilter::try_from_default_env()
        .or_else(|_| EnvFilter::try_new("ml_pipeline=trace,tower_http=debug,"))
        .unwrap()
    ).init();

    runtime.block_on(entrypoint());
}

async fn entrypoint() {
    // Sketchy
    unsafe { env::set_var("HF_ENDPOINT", "https://hf-mirror.com") };
    unsafe { env::set_var("HF_HUB_ENABLE_TELEMETRY", "0")};

    let db: Database<Init> = Database::init_db(&*DATABASE_CONNECTION).await.unwrap();

    let embedder: TextEmbedding = TextEmbedding::try_new(
        InitOptions::new(
            EmbeddingModel::MultilingualE5Small
        ).with_cache_dir("./data/embedding_cache".into())
        .with_show_download_progress(true)
    ).unwrap();

    let filesystem: Filesystem<FSInit> = Filesystem::init().unwrap();
    let rake: Rake = init_rake();
    let app_state: Arc<AppState> = Arc::new(AppState { db, embedder, filesystem, rake });

    let router: Router = Router::new()
    .route("/api/document/upload", post(process_pdf))
    .layer(DefaultBodyLimit::max(200 << 20))
    .route("/api/message/upload", post(upload_message))
    .with_state(app_state.clone())
    .route("/api/message/delete", delete(delete_message))
    .with_state(app_state.clone())
    .route("/api/message/inference", get(inference))
    .with_state(app_state.clone())
    .route("/api/message/history", get(history))
    .with_state(app_state.clone())
    .route("/api/delete/user", delete(delete_user))
    .with_state(app_state.clone())
    .route("/api/delete/document", delete(delete_documents))
    .with_state(app_state.clone())
    .route("/api/delete/history", delete(delete_message))
    .with_state(app_state.clone())
    .route("/health", get(|| async {"OK"}))
    .layer(TraceLayer::new_for_http());

    axum::serve(TcpListener::bind("0.0.0.0:3030").await.unwrap(), router).await.unwrap();
}

#[cfg(feature = "build")]
fn main() {
    println!("Building");
    unsafe { env::set_var("HF_ENDPOINT", "https://hf-mirror.com") };
    unsafe { env::set_var("HF_HUB_ENABLE_TELEMETRY", "0")};
    let _: TextEmbedding = TextEmbedding::try_new(
        InitOptions::new(
            EmbeddingModel::MultilingualE5Small
        ).with_cache_dir("./data/embedding_cache".into())
        .with_show_download_progress(true)
    ).unwrap();
}