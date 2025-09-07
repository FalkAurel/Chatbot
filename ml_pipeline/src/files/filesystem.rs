use std::{
    fs::{create_dir_all, exists, remove_file},
    marker::PhantomData,
    path::{Path, PathBuf},
    sync::LazyLock
};
use tokio::{fs::{remove_dir_all, write}, io};
use crate::files::file::File;

/// Base directory path for file storage (`./data/files`).
/// Uses `LazyLock` for thread-safe, one-time initialization.
static BASE_PATH: LazyLock<PathBuf> = LazyLock::new(|| {
    let mut base = PathBuf::from("./data");
    base.push("files");  // OS-agnostic path joining
    base
});

/// Marker type for uninitialized filesystem state.
pub struct Uninit;

/// Marker type for initialized filesystem state.
pub struct Init;

/// A filesystem abstraction with compile-time state enforcement.
///
/// # Type States
/// - `Filesystem<Uninit>`: Must be initialized via `init()` before use.
/// - `Filesystem<Init>`: Guaranteed to be ready for file operations.
///
/// # Safety
/// - `Send`/`Sync` are manually implemented for `Filesystem<Init>` because:
///   - It contains no mutable state (`PhantomData` is zero-sized).
///   - All operations are thread-safe (filesystem I/O uses OS locks).
#[derive(Debug, PartialEq)]
pub struct Filesystem<T> {
    _marker: PhantomData<T>,
}

impl Filesystem<Uninit> {
    /// Initializes the filesystem by creating the base directory if it doesn't exist.
    ///
    /// # Returns
    /// - `Ok(Filesystem<Init>)` on success.
    /// - `Err(std::io::Error)` if directory creation fails.
    pub fn init() -> std::io::Result<Filesystem<Init>> {
        if !exists(BASE_PATH.as_path())? {
            create_dir_all(BASE_PATH.as_path())?;
        }
        Ok(Filesystem::<Init> { _marker: PhantomData })
    }
}

impl Filesystem<Init> {
    /// Writes a file to the filesystem.
    ///
    /// # Behavior
    /// 1. Creates parent directories if needed (e.g., `./data/files/<id>/`).
    /// 2. Writes file contents atomically.
    ///
    /// # Arguments
    /// - `file`: A validated `File` with `<id>/<name>` structure.
    ///
    /// # Errors
    /// - Fails if directory creation or file write fails.
    ///
    /// # Safety
    /// - Assumes `file.get_filename()` follows `<id>/<name>` format (enforced by `File` type).
    pub async fn write(&self, file: &File<'_>) -> io::Result<()> {
        let file_path: PathBuf = BASE_PATH.join(file.get_filename());
        let parent_dir: &Path = file_path.parent()
        .ok_or_else(|| io::Error::new(
            io::ErrorKind::InvalidInput,
             "Filename must contain a parent directory"
            )
        )?;

        tokio::fs::create_dir_all(parent_dir).await?;

        println!("Saving to: {:?}", file_path);
        write(file_path, file.get_content()).await?;
        Ok(())
    }


    pub async fn remove_user(&self, user_id: i64) -> io::Result<()> {
        remove_dir_all(BASE_PATH.join(user_id.to_string())).await
    }

    pub async fn remove_file(&self, file: File<'_>) -> io::Result<()> {
        let file_path: PathBuf = BASE_PATH.join(file.get_filename());
        
        
        println!("Deleting: {:?}", file_path);
        remove_file(file_path)
    }
}

// SAFETY: `Filesystem<Init>` is thread-safe because:
// - `BASE_PATH` is immutable after initialization (`LazyLock`).
// - Filesystem ops use OS-level synchronization.
unsafe impl Send for Filesystem<Init> {}
unsafe impl Sync for Filesystem<Init> {}

#[cfg(test)]
mod tests {
    use super::*;
    use std::fs::{remove_dir, remove_file};

    #[test]
    fn init_fs() {
        assert!(Filesystem::init().is_ok());
    }

    #[tokio::test]
    async fn write_file() {
        let fs: Filesystem<Init> = Filesystem::init().unwrap();
        let file: File = unsafe { File::new("test/hello.txt", b"content") };
        
        fs.write(&file).await.unwrap();
        assert!(exists(BASE_PATH.join(file.get_filename())).unwrap());
        
        // Cleanup
        remove_file(BASE_PATH.join(file.get_filename())).unwrap();
        remove_dir(BASE_PATH.join("test")).unwrap();
    }
}