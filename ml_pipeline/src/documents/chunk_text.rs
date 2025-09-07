use std::str;

/// A trait defining a chunking mechanism for string data.
/// Implementations must be thread-safe (Send + Sync) and provide
/// an iterator over string chunks.
pub trait Chunker: Send + Sync {
    /// Returns an iterator over chunks of the underlying data.
    /// Each item in the iterator represents a discrete chunk.
    fn chunks(&self) -> impl Iterator<Item = &str>;
}

/// A basic chunker implementation that splits data based on a separator character.
/// 
/// Characteristics:
///     - Low computational cost
///     - Simple implementation
///     - Chunk boundaries are strictly determined by the separator
pub struct SeperatorBased<'data> {
    /// The data to be chunked
    data: &'data str,
    /// The character used to separate chunks
    seperator: &'static str,
}

impl<'data> SeperatorBased<'data> {
    /// Creates a new SeperatorBased chunker with the given data and separator.
    #[allow(dead_code)]
    pub const fn new(data: &'data str, seperator: &'static str) -> Self {
        Self { data, seperator }
    }
}

impl<'data> Chunker for SeperatorBased<'data> {
    /// Returns an iterator that yields substrings of the original data,
    /// split by the configured separator character.
    fn chunks(&self) -> impl Iterator<Item = &str> {
        self.data.split(self.seperator)
    }
}

pub struct FixedSizeOverlap<'data, const STEP_SIZE: usize, const OVERLAP: usize> {
    data: &'data str
}

impl <'data, const STEP_SIZE: usize, const OVERLAP: usize> FixedSizeOverlap<'data, STEP_SIZE, OVERLAP> {
    pub fn new(data: &'data str) -> Self {
        Self { data }
    }
}

impl <'data, const STEP_SIZE: usize, const OVERLAP: usize> Chunker for FixedSizeOverlap<'data, STEP_SIZE, OVERLAP> {
    fn chunks(&self) -> impl Iterator<Item = &str> {
        FixedSizeOverlapIterator::<STEP_SIZE, OVERLAP> { data: self.data, index: 0 }
    }
}

pub struct FixedSizeOverlapIterator<'data, const STEP_SIZE: usize, const OVERLAP: usize> {
    data: &'data str,
    index: usize
}

impl <'data, const STEP_SIZE: usize, const OVERLAP: usize> Iterator for FixedSizeOverlapIterator<'data, STEP_SIZE, OVERLAP> {
    type Item = &'data str;
    fn next(&mut self) -> Option<Self::Item> {
        if self.index >= self.data.len() {
            return None;
        }

        let mut end: usize = self.index + STEP_SIZE;
        end = end.min(self.data.len() - 1);

        // is_char_boundary has as a side effect to return true if it is the end of the slice
        while !self.data.is_char_boundary(end) {
            end += 1
        }

        let chunk: &str = &self.data[self.index..end];
        self.index += STEP_SIZE.saturating_sub(OVERLAP);

        while !self.data.is_char_boundary(self.index) {
            self.index -= 1;
        }


        Some(chunk)
    }
}

#[cfg(test)]
mod chunking {
    use crate::documents::chunk_text::{Chunker, FixedSizeOverlap};

    #[test]
    fn fixed_size_overlap() {
        let data: &str = "Hallo wie geht's dir? Mir geht es gut!";

        let chunker: FixedSizeOverlap<10, 4> = FixedSizeOverlap::<10, 4>::new(data);
        println!("Testing Fixed size");

        let mut chunker = chunker.chunks();

        assert_eq!(Some("Hallo wie "), chunker.next());
        assert_eq!(Some("wie geht's"), chunker.next());
        assert_eq!(Some("ht's dir? "), chunker.next());
        assert_eq!(Some("ir? Mir ge"), chunker.next());
        assert_eq!(Some("r geht es "), chunker.next());
        assert_eq!(Some(" es gut!"), chunker.next());
        assert_eq!(Some("t!"), chunker.next());
        assert_eq!(None, chunker.next())
    }
}






