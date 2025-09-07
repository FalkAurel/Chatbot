use std::str::{self, Split};
use axum::http::HeaderMap;


#[derive(Debug, PartialEq)]
pub enum FileError {
    MalformedFilename,
    InvalidCharacters,
    HeaderMissing,
}

/// Represents a file with a validated filename and raw binary data.
///
/// The filename must follow the format `<numeric_id>/<identifier>` (e.g., `123/foo.txt`).
/// Validation is enforced in [`from_request`](#method.from_request).
///
/// # Safety
/// - [`new`](#method.new) is `unsafe` because it bypasses filename validation.
/// - Use [`from_request`](#method.from_request) for safe construction.
#[derive(Debug, PartialEq)]
pub struct File<'data> {
    filename: &'data str,
    data: &'data [u8],
}

impl <'data> File<'data> {
    pub const unsafe fn new(
        filename: &'data str, 
        data: &'data [u8]
    ) -> Self {
        Self { filename, data }
    }

    /// Constructs a `File` from HTTP headers and body.
    ///
    /// # Errors
    /// - Returns [`FileError::HeaderMissing`] if `X-Filename` is absent.
    /// - Returns [`FileError::InvalidCharacters`] if the filename is not UTF-8.
    /// - Returns [`FileError::MalformedFilename`] if the format is invalid.
    pub fn from_request(
        headers: &'data HeaderMap, 
        body: &'data [u8]
    ) -> Result<Self, FileError> {
        let filename: &'data [u8] = headers.get("X-Filename").ok_or(FileError::HeaderMissing)?.as_bytes();
        
        let filename: &'data str = Self::extract_filename(
            str::from_utf8(filename)
            .map_err(|_| FileError::InvalidCharacters)?
        )?;

        Ok(Self { filename, data: body })
    }

    /// Validates the filename format (`<numeric_id>/<identifier>`).
    fn extract_filename(filename: &str) -> Result<&str, FileError> {
        let mut split_filename: Split<&str> = filename.split("/");
        
        let id: &str = split_filename.next().ok_or(FileError::MalformedFilename)?;
        let is_numeric: bool = id.chars().all(|c| c.is_digit(10));

        if !is_numeric {
            return Err(FileError::MalformedFilename)
        }

        split_filename.next().ok_or(FileError::MalformedFilename)?;
        Ok(filename)
    }

    pub fn get_content(&self) -> &[u8] {
        self.data
    }

    pub fn get_filename(&self) -> &str {
        self.filename
    }
}

#[cfg(test)]
mod file {
    use axum::http::{HeaderMap, HeaderValue};
    use crate::files::file::{File, FileError};

    #[test]
    fn from_request_success() {
        let mut map = HeaderMap::new();
        map.append("X-Filename", HeaderValue::from_static("123/Es funktioniert"));
        let result = File::from_request(
            &map,
            &[]
        );

        let result = result.unwrap();
        assert_eq!(result.filename, "123/Es funktioniert")
    }

    #[test]
    fn from_request_failure_non_numeric_id() {
        let mut map = HeaderMap::new();
        map.append("X-Filename", HeaderValue::from_static("123A/Es funktioniert"));
        let result = File::from_request(
            &map,
            &[]
        );

        assert_eq!(result, Err(FileError::MalformedFilename))
    }

    #[test]
    fn from_request_failure_wrong_seperator() {
        let mut map = HeaderMap::new();
        map.append("X-Filename", HeaderValue::from_static("123A_Es funktioniert"));
        let result = File::from_request(
            &map,
            &[]
        );

        assert_eq!(result, Err(FileError::MalformedFilename))
    }

    #[test]
    fn from_request_failure_header_missing() {
        let map = HeaderMap::new();
        let result = File::from_request(
            &map,
            &[]
        );

        assert_eq!(result, Err(FileError::HeaderMissing))
    }
}