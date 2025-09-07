//! A secure, type-driven abstraction for filesystem operations.
//!
//! This module provides two core types:
//! - [`File`]: A validated representation of a file and its metadata, enforcing invariants at compile time.
//! - [`Filesystem`]: A platform-agnostic interface for filesystem interactions (e.g., writing files).
//!
//! # Design Philosophy
//! - **Type-Safety**: Invariants (e.g., filename structure) are enforced by the type system.
//! - **Zero-Cost Abstractions**: No runtime overhead for validation when invariants are statically known.
//! - **Platform Agnosticism**: Uses Rust's `std::path` for cross-platform paths (Windows/Unix).
//!
//! # Performance Notes
//! - **Reuse `Filesystem`**: Initialization requires syscalls (checking/creating directories).
//! - **Stateless**: Safe to create in blocking threads (implements `Send + Sync`).
//!
//! # Safety
//! - `unsafe` APIs (e.g., `File::new`) require callers to uphold invariants. Prefer [`File::from_request`]
//!   for untrusted input.

mod filesystem;
mod file;

pub use file::*;
pub use filesystem::{Filesystem, Init};